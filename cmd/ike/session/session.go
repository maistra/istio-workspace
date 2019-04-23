package session

import (
	"errors"
	"fmt"
	"os/user"
	"strings"
	"time"

	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/naming"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Handler is a function to setup a server sessionName before attempting to connect. Returns a 'cleanup' function.
type Handler func(cmd *cobra.Command) (func(), error)

// Offline is a empty Handler doing nothing. Used for testing
func Offline(cmd *cobra.Command) (func(), error) {
	return func() {}, nil
}

type handler struct {
	c           *client
	ref,
	namespace,
	sessionName,
	route       string
}

// CreateOrJoinHandler provides the option to either create a new sessionName if non exist or join an existing.
// Rely on the following flags:
//  * namespace - the name of the target namespace where deployment is defined
//  * deployment - the name of the target deployment and will update the flag with the new deployment name
//  * sessionName - the name of the sessionName
//  * route - the definition of traffic routing
func CreateOrJoinHandler(cmd *cobra.Command) (func(), error) {
	deploymentName, _ := cmd.Flags().GetString("deployment")
	namespace, _ := cmd.Flags().GetString("namespace")
	route, _ := cmd.Flags().GetString("route")
	sessionName := getSessionName(cmd)

	h := &handler{c: DefaultClient(namespace),
		ref:         deploymentName,
		namespace:   namespace,
		route:       route,
		sessionName: sessionName}

	serviceName, err := h.createOrJoinSession()
	if err != nil {
		return func() {}, err
	}
	err = cmd.Flags().Set("deployment", serviceName) // HACK: pass arguments around, not flags?
	if err != nil {
		return func() {}, err
	}

	return func() {
		h.removeOrLeaveSession()
	}, nil
}

// createOrJoinSession calls oc cli and creates a Session CD waiting for the 'success' status and return the new name
func (h *handler) createOrJoinSession() (string, error) {

	session, err := h.c.Get(h.sessionName)
	if err != nil {
		err = h.createSession()
		if err != nil {
			return "", err
		}
		return h.waitForRefToComplete()
	}
	// join sessionName
	session.Spec.Refs = append(session.Spec.Refs, h.ref)
	err = h.c.Create(session)
	if err != nil {
		return "", err
	}
	return h.waitForRefToComplete()
}

func (h *handler) createSession() error {
	r, err := ParseRoute(h.route)
	if err != nil {
		return err
	}
	session := istiov1alpha1.Session{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "istio.openshift.com/v1alpha1",
			Kind:       "Session",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: h.sessionName,
		},
		Spec: istiov1alpha1.SessionSpec{
			Refs: []string{
				h.ref,
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
		sessionStatus, err := h.c.Get(h.sessionName)
		if err != nil {
			return false, nil
		}
		for _, refs := range sessionStatus.Status.Refs {
			if refs.Name == h.ref {
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
		return "", errors.New("no Deployment or DeploymentConfig found for target")
	}
	return name, nil
}

func (h *handler) removeOrLeaveSession() {
	session, err := h.c.Get(h.sessionName)
	if err != nil {
		return // assume missing, nothing to clean?
	}
	// more then one participant, update sessionName
	for i, r := range session.Spec.Refs {
		if r == h.ref {
			session.Spec.Refs = append(session.Spec.Refs[:i], session.Spec.Refs[i+1:]...)
		}
	}
	if len(session.Spec.Refs) == 0 {
		_ = h.c.Delete(session)
	} else {
		_ = h.c.Create(session)
	}
}

func getSessionName(cmd *cobra.Command) string {
	sessionName, err := cmd.Flags().GetString("sessionName")
	if err == nil && sessionName != "" {
		return sessionName
	}
	random := naming.RandName(5)
	u, err := user.Current()
	if err != nil {
		return random
	}
	return u.Username + "-" + random
}

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
