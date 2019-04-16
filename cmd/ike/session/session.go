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

// Handler is a function to setup a server session before attempting to connect. Returns a 'cleanup' function.
type Handler func(cmd *cobra.Command) (func(), error)

// Offline is a empty Handler doing nothing. Used for testing
func Offline(cmd *cobra.Command) (func(), error) {
	return func() {}, nil
}

// CreateOrJoinHandler provides the option to either create a new session if non exist or join an existing.
// Rely on the following flags:
//  * deployment - the name of the target deployment and will update the flag with the new deployment name
//  * session - the name of the session
//  * route - the definition of traffic routing
func CreateOrJoinHandler(cmd *cobra.Command) (func(), error) {
	sessionName := getSessionName(cmd)
	deploymentName, _ := cmd.Flags().GetString("deployment")
	route, _ := cmd.Flags().GetString("route")

	serviceName, err := createOrJoinSession(sessionName, route, deploymentName)
	if err != nil {
		return func() {}, err
	}
	err = cmd.Flags().Set("deployment", serviceName) // HACK: pass arguments around, not flags?
	if err != nil {
		return func() {}, err
	}

	return func() {
		removeOrLeaveSession(sessionName, deploymentName)
	}, nil
}

// createOrJoinSession calls oc cli and creates a Session CD waiting for the 'success' status and return the new name
func createOrJoinSession(sessionName, route, ref string) (string, error) {
	session, err := DefaultClient().Get(sessionName)
	if err != nil {
		err = createSession(sessionName, route, ref)
		if err != nil {
			return "", err
		}
		return waitForRefToComplete(sessionName, ref)
	}
	// join session

	session.Spec.Refs = append(session.Spec.Refs, ref)
	err = DefaultClient().Create(session)
	if err != nil {
		return "", err
	}
	return waitForRefToComplete(sessionName, ref)
}

func createSession(sessionName, route, ref string) error {
	r, err := parseRoute(route)
	if err != nil {
		return err
	}
	session := istiov1alpha1.Session{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "istio.openshift.com/v1alpha1",
			Kind:       "Session",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: sessionName,
		},
		Spec: istiov1alpha1.SessionSpec{
			Refs: []string{
				ref,
			},
		},
	}

	if r != nil {
		session.Spec.Route = *r
	}
	return DefaultClient().Create(&session)
}

func parseRoute(route string) (*istiov1alpha1.Route, error) {
	if route == "" {
		return nil, nil
	}
	var t, n, v string

	typed := strings.Split(route, ":")
	if len(typed) != 2 {
		return nil, fmt.Errorf("route in wrong format type:name=value")
	}
	t = typed[0]

	pair := strings.Split(typed[1], "=")
	if len(typed) != 2 {
		return nil, fmt.Errorf("route in wrong format type:name=value")
	}
	n, v = pair[0], pair[1]
	return &istiov1alpha1.Route{
		Type:  t,
		Name:  n,
		Value: v,
	}, nil
}

func waitForRefToComplete(sessionName, ref string) (string, error) {
	var name string
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		sessionStatus, err := DefaultClient().Get(sessionName)
		if err != nil {
			return false, nil
		}
		for _, refs := range sessionStatus.Status.Refs {
			if refs.Name == ref {
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

func removeOrLeaveSession(sessionName, ref string) {
	session, err := DefaultClient().Get(sessionName)
	if err != nil {
		return // assume missing, nothing to clean?
	}
	// more then one participant, update session
	for i, r := range session.Spec.Refs {
		if r == ref {
			session.Spec.Refs = append(session.Spec.Refs[:i], session.Spec.Refs[i+1:]...)
		}
	}
	if len(session.Spec.Refs) == 0 {
		_ = DefaultClient().Delete(session)
	} else {
		_ = DefaultClient().Create(session)
	}
}

func getSessionName(cmd *cobra.Command) string {
	sessionName, err := cmd.Flags().GetString("session")
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
