package session

import (
	"os/user"
	"regexp"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/naming"
)

var (
	logger = func() logr.Logger {
		return log.Log.WithValues("type", "session")
	}
	errorWrongRouteFormat = errors.Sentinel("route in wrong format. expected type:name=value")
)

// Options holds the variables used by the Session Handler.
type Options struct {
	NamespaceName  string            // name of the namespace for target resource
	DeploymentName string            // name of the initial resource to target
	SessionName    string            // name of the session create or join if exist
	RouteExp       string            // expression of how to route the traffic to the target resource
	Strategy       string            // name of the strategy to use for the target resource
	StrategyArgs   map[string]string // additional arguments for the strategy
	Revert         bool              // Revert back to previous known value if join/leave a existing session with a known ref
	Duration       *time.Duration    // Duration defines the interval used to check for changes to the session object
}

// State holds the new variables as presented by the creation of the session.
type State struct {
	DeploymentName string              // name of the resource to target within the cloned route.
	Hosts          []string            // currently exposed hosts
	Route          istiov1alpha1.Route // the current route configuration
}

// Handler is a function to setup a server session before attempting to connect. Returns a 'cleanup' function.
type Handler func(opts Options, client *Client) (State, func(), error)

// Offline is a empty Handler doing nothing. Used for testing.
func Offline(opts Options, client *Client) (State, func(), error) {
	return State{DeploymentName: opts.DeploymentName}, func() {}, nil
}

// handler wraps the session client and required metadata used to manipulate the resources.
type handler struct {
	c             *Client
	opts          Options
	previousState *istiov1alpha1.Ref // holds the previous Ref if replaced. Used to Revert back to old state on remove.
}

// RemoveHandler provides the option to delete an existing sessions if found.
// Rely on the following flags:
//  * namespace - the name of the target namespace where deployment is defined
//  * session - the name of the session.
func RemoveHandler(opts Options, client *Client) (State, func()) { //nolint:gocritic //reason too simple to use named results
	if client == nil {
		return State{}, func() {}
	}
	h := &handler{c: client,
		opts: opts}

	return State{}, func() {
		h.removeOrLeaveSession()
	}
}

// CreateOrJoinHandler provides the option to either create a new session if non exist or join an existing.
// Rely on the following flags:
//  * namespace - the name of the target namespace where deployment is defined
//  * deployment - the name of the target deployment and will update the flag with the new deployment name
//  * session - the name of the session
//  * route - the definition of traffic routing.
func CreateOrJoinHandler(opts Options, client *Client) (State, func(), error) {
	sessionName, err := getOrCreateSessionName(opts.SessionName)
	if err != nil {
		return State{}, func() {}, err
	}
	opts.SessionName = sessionName

	h := &handler{c: client, opts: opts}

	session, serviceName, err := h.createOrJoinSession()
	if err != nil {
		return State{}, func() {}, err
	}
	route := session.Status.Route //nolint:ifshort // route used in multiple locations
	if route == nil {
		route = &istiov1alpha1.Route{}
	}

	return State{
			DeploymentName: serviceName,
			Hosts:          session.Status.Hosts,
			Route:          *route,
		}, func() {
			h.removeOrLeaveSession()
		}, nil
}

// createOrJoinSession calls oc cli and creates a Session CD waiting for the 'success' status and return the new name.
func (h *handler) createOrJoinSession() (*istiov1alpha1.Session, string, error) {
	session, err := h.c.Get(h.opts.SessionName)
	if err != nil {
		session, err = h.createSession()
		if err != nil {
			return session, "", err
		}

		return h.removeSessionIfDeploymentNotFound()
	}
	ref := istiov1alpha1.Ref{Name: h.opts.DeploymentName, Strategy: h.opts.Strategy, Args: h.opts.StrategyArgs}
	// update ref in session
	for i, r := range session.Spec.Refs {
		if r.Name != h.opts.DeploymentName {
			continue
		}
		prev := session.Spec.Refs[i]
		h.previousState = &prev // point to a variable, not a array index
		session.Spec.Refs[i] = ref
		err = h.c.Update(session)
		if err != nil {
			return session, "", err
		}

		return h.removeSessionIfDeploymentNotFound()
	}
	// join session
	session.Spec.Refs = append(session.Spec.Refs, ref)
	err = h.c.Update(session)
	if err != nil {
		return session, "", err
	}

	return h.removeSessionIfDeploymentNotFound()
}

func (h *handler) removeSessionIfDeploymentNotFound() (*istiov1alpha1.Session, string, error) {
	session, result, err := h.waitForRefToComplete()
	if errors.As(err, &DeploymentNotFoundError{}) {
		h.removeOrLeaveSession()
	}

	return session, result, err
}

func (h *handler) createSession() (*istiov1alpha1.Session, error) {
	r, err := ParseRoute(h.opts.RouteExp)
	if err != nil {
		return nil, err
	}
	session := istiov1alpha1.Session{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "maistra.io/v1alpha1",
			Kind:       "Session",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: h.opts.SessionName,
		},
		Spec: istiov1alpha1.SessionSpec{
			Refs: []istiov1alpha1.Ref{
				{Name: h.opts.DeploymentName, Strategy: h.opts.Strategy, Args: h.opts.StrategyArgs},
			},
		},
	}

	if r != nil {
		session.Spec.Route = *r
	}

	return &session, h.c.Create(&session)
}

func (h *handler) waitForRefToComplete() (*istiov1alpha1.Session, string, error) {
	var sessionStatus *istiov1alpha1.Session
	duration := 1 * time.Minute
	if h.opts.Duration != nil {
		duration = *h.opts.Duration
	}
	err := wait.Poll(2*time.Second, duration, func() (bool, error) {
		sessionStatus, err := h.c.Get(h.opts.SessionName)
		if err != nil {
			return false, err
		}

		if sessionStatus.Status.State != nil && *sessionStatus.Status.State == istiov1alpha1.StateSuccess {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return sessionStatus, "", errors.Wrap(err, "timed out waiting for success")
	}
	for _, condition := range sessionStatus.Status.Conditions {
		if condition.Source.Ref == h.opts.DeploymentName {
			if condition.Source.Kind == "DeploymentConfig" || condition.Source.Kind == "Deployment" {
				if condition.Target != nil {
					return sessionStatus, condition.Target.Name, nil
				}

			}
		}
	}

	return sessionStatus, "", DeploymentNotFoundError{name: h.opts.DeploymentName}
}

func (h *handler) removeOrLeaveSession() {
	session, err := h.c.Get(h.opts.SessionName)
	if err != nil {
		logger().Error(err, "failed removing or leaving session")

		return // assume missing, nothing to clean?
	}
	// more than one participant, update session
	for i, r := range session.Spec.Refs {
		if r.Name == h.opts.DeploymentName {
			if h.opts.Revert && h.previousState != nil {
				session.Spec.Refs[i] = *h.previousState
			} else {
				session.Spec.Refs = append(session.Spec.Refs[:i], session.Spec.Refs[i+1:]...)
			}
		}
	}
	if len(session.Spec.Refs) == 0 {
		_ = h.c.Delete(session)
	} else {
		_ = h.c.Update(session)
	}
}

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

func getOrCreateSessionName(sessionName string) (string, error) {
	if sessionName != "" {
		namingErrors := validation.IsDNS1123Label(sessionName)
		if len(namingErrors) != 0 {
			var errs []error
			for _, namingError := range namingErrors {
				errs = append(errs, errors.New(namingError))
			}

			return "", errors.WrapIfWithDetails(errors.Combine(errs...), "your suggested session name is not a valid k8s value", "name", sessionName)
		}
	}

	if sessionName == "" {
		random := naming.RandName(5)
		u, err := user.Current()
		if err != nil {
			sessionName = random
		} else {
			sessionName = naming.ConcatToMax(63, "user", u.Username, random)
		}
	}

	return nonAlphaNumeric.ReplaceAllString(sessionName, "-"), nil
}

// ParseRoute maps string route representation into a Route struct by unwrapping its type, name and value.
func ParseRoute(route string) (*istiov1alpha1.Route, error) {
	if route == "" {
		return nil, nil
	}
	var t, n, v string

	typed := strings.Split(route, ":")
	if len(typed) != 2 {
		return nil, errorWrongRouteFormat
	}
	t = typed[0]

	pair := strings.Split(typed[1], "=")
	if len(pair) != 2 {
		return nil, errorWrongRouteFormat
	}
	n, v = pair[0], pair[1]

	return &istiov1alpha1.Route{
		Type:  t,
		Name:  n,
		Value: v,
	}, nil
}
