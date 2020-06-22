package session

import (
	"fmt"
	"os/user"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	istiov1alpha1 "github.com/maistra/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/naming"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	logger = log.CreateOperatorAwareLogger("session").WithValues("type", "controller")
)

// Options holds the variables used by the Session Handler.
type Options struct {
	NamespaceName  string            // name of the namespace for target resource
	DeploymentName string            // name of the initial resource to target
	SessionName    string            // name of the session create or join if exist
	RouteExp       string            // expression of how to route the traffic to the target resource
	Strategy       string            // name of the strategy to use for the target resource
	StrategyArgs   map[string]string // additional arguments for the strategy
}

// State holds the new variables as presented by the creation of the session.
type State struct {
	DeploymentName string // name of the resource to target within the cloned route.
}

// Handler is a function to setup a server session before attempting to connect. Returns a 'cleanup' function.
type Handler func(opts Options) (State, func(), error)

// Offline is a empty Handler doing nothing. Used for testing.
func Offline(opts Options) (State, func(), error) {
	return State{DeploymentName: opts.DeploymentName}, func() {}, nil
}

// handler wraps the session client and required metadata used to manipulate the resources.
type handler struct {
	c    *client
	opts Options
}

// RemoveHandler provides the option to delete an existing sessions if found.
// Rely on the following flags:
//  * namespace - the name of the target namespace where deployment is defined
//  * session - the name of the session.
func RemoveHandler(opts Options) (State, func(), error) {
	client, err := DefaultClient(opts.NamespaceName)

	if err != nil {
		return State{}, func() {}, err
	}

	h := &handler{c: client,
		opts: opts}

	return State{}, func() {
		h.removeOrLeaveSession()
	}, nil
}

// CreateOrJoinHandler provides the option to either create a new session if non exist or join an existing.
// Rely on the following flags:
//  * namespace - the name of the target namespace where deployment is defined
//  * deployment - the name of the target deployment and will update the flag with the new deployment name
//  * session - the name of the session
//  * route - the definition of traffic routing.
func CreateOrJoinHandler(opts Options) (State, func(), error) {
	sessionName := getOrCreateSessionName(opts.SessionName)
	opts.SessionName = sessionName

	client, err := DefaultClient(opts.NamespaceName)

	if err != nil {
		return State{}, func() {}, err
	}

	h := &handler{c: client,
		opts: opts}

	serviceName, err := h.createOrJoinSession()
	if err != nil {
		return State{}, func() {}, err
	}
	return State{
			DeploymentName: serviceName,
		}, func() {
			h.removeOrLeaveSession()
		}, nil
}

// createOrJoinSession calls oc cli and creates a Session CD waiting for the 'success' status and return the new name.
func (h *handler) createOrJoinSession() (string, error) {
	session, err := h.c.Get(h.opts.SessionName)
	if err != nil {
		err = h.createSession()
		if err != nil {
			return "", err
		}
		return h.removeSessionIfDeploymentNotFound()
	}
	ref := istiov1alpha1.Ref{Name: h.opts.DeploymentName, Strategy: h.opts.Strategy, Args: h.opts.StrategyArgs}
	// update ref in session
	for i, r := range session.Spec.Refs {
		if r.Name == h.opts.DeploymentName {
			session.Spec.Refs[i] = ref
			err = h.c.Update(session)
			if err != nil {
				return "", err
			}
			return h.removeSessionIfDeploymentNotFound()
		}
	}
	// join session
	session.Spec.Refs = append(session.Spec.Refs, ref)
	err = h.c.Update(session)
	if err != nil {
		return "", err
	}
	return h.removeSessionIfDeploymentNotFound()
}

func (h *handler) removeSessionIfDeploymentNotFound() (string, error) {
	result, err := h.waitForRefToComplete()
	if _, deploymentNotFound := err.(DeploymentNotFoundError); deploymentNotFound {
		h.removeOrLeaveSession()
	}
	return result, err
}

func (h *handler) createSession() error {
	r, err := ParseRoute(h.opts.RouteExp)
	if err != nil {
		return err
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
	return h.c.Create(&session)
}

func (h *handler) waitForRefToComplete() (string, error) {
	var name string
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		sessionStatus, err := h.c.Get(h.opts.SessionName)
		if err != nil {
			return false, err
		}
		for _, refs := range sessionStatus.Status.Refs {
			if refs.Name == h.opts.DeploymentName {
				for _, res := range refs.Resources {
					if *res.Kind == "Deployment" || *res.Kind == "DeploymentConfig" {
						name = *res.Name
						fmt.Printf("found %s\n", name)
						return true, nil
					}
				}
			}
		}
		return false, nil
	})
	if err != nil {
		logger.Error(err, "failed waiting for deployment to create")
		return name, DeploymentNotFoundError{name: h.opts.DeploymentName}
	}
	return name, nil
}

func (h *handler) removeOrLeaveSession() {
	session, err := h.c.Get(h.opts.SessionName)
	if err != nil {
		logger.Error(err, "failed removing or leaving session")
		return // assume missing, nothing to clean?
	}
	// more than one participant, update session
	for i, r := range session.Spec.Refs {
		if r.Name == h.opts.DeploymentName {
			session.Spec.Refs = append(session.Spec.Refs[:i], session.Spec.Refs[i+1:]...)
		}
	}
	if len(session.Spec.Refs) == 0 {
		_ = h.c.Delete(session)
	} else {
		_ = h.c.Update(session)
	}
}

func getOrCreateSessionName(sessionName string) string {
	if sessionName != "" {
		return sessionName
	}
	random := naming.RandName(5)
	u, err := user.Current()
	if err != nil {
		return random
	}
	return u.Username + "-" + random
}

// ParseRoute maps string route representation into a Route struct by unwrapping its type, name and value.
func ParseRoute(route string) (*istiov1alpha1.Route, error) {
	if route == "" {
		return nil, nil
	}
	var t, n, v string

	typed := strings.Split(route, ":")
	if len(typed) != 2 {
		return nil, fmt.Errorf("route in wrong format. expected type:name=value")
	}
	t = typed[0]

	pair := strings.Split(typed[1], "=")
	if len(pair) != 2 {
		return nil, fmt.Errorf("route in wrong format. expected type:name=value")
	}
	n, v = pair[0], pair[1]
	return &istiov1alpha1.Route{
		Type:  t,
		Name:  n,
		Value: v,
	}, nil
}
