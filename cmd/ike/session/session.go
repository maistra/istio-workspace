package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os/user"
	"time"

	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	helper "github.com/aslakknutsen/istio-workspace/pkg/istio"
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
func CreateOrJoinHandler(cmd *cobra.Command) (func(), error) {
	sessionName := getSessionName(cmd)
	deploymentName, _ := cmd.Flags().GetString("deployment")
	serviceName, err := createOrJoinSession(sessionName, deploymentName)
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
func createOrJoinSession(sessionName, ref string) (string, error) {

	session, err := getSession(sessionName)
	if err != nil {
		err = createSession(sessionName, ref)
		if err != nil {
			return "", err
		}
		return waitForRefToComplete(sessionName, ref)
	}
	// join session

	session.Spec.Refs = append(session.Spec.Refs, ref)
	err = applySession(session)
	if err != nil {
		return "", err
	}
	return waitForRefToComplete(sessionName, ref)
}

func createSession(sessionName, ref string) error {
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
	return applySession(session)
}

func applySession(session istiov1alpha1.Session) error { //nolint[:hugeParam]
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sessionData := string(b)
	resp, err := helper.ExecuteOCCMD(&sessionData, fmt.Sprintf("oc apply -f -"))
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func waitForRefToComplete(sessionName, ref string) (string, error) {
	var name string
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		sessionStatus, err := getSession(sessionName)
		if err != nil {
			return false, nil
		}
		for _, refs := range sessionStatus.Status.Refs {
			if refs.Name == ref {
				for _, res := range refs.Resources {
					if *res.Kind == "Deployment" || *res.Kind == "DeploymentConfig" {
						name = *res.Name
						fmt.Println("Found")
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

func getSession(sessionName string) (istiov1alpha1.Session, error) {
	status, err := helper.ExecuteOCCMD(nil, fmt.Sprintf("oc get session %v -o json", sessionName))
	if err != nil {
		return istiov1alpha1.Session{}, err
	}
	sessionStatus := istiov1alpha1.Session{}
	err = json.Unmarshal([]byte(status), &sessionStatus)
	if err != nil {
		return istiov1alpha1.Session{}, err
	}
	return sessionStatus, nil
}

func removeOrLeaveSession(sessionName, ref string) {
	session, err := getSession(sessionName)
	if err != nil {
		return // assume missing, nothing to clean?
	}
	// TODO: Check if our ref name is the same as Spec. If not, ignore and move on
	if len(session.Spec.Refs) == 1 {
		removeSession(sessionName)
		return
	}
	// more then one participant, update session
	for i, r := range session.Spec.Refs {
		if r == ref {
			session.Spec.Refs = append(session.Spec.Refs[:i], session.Spec.Refs[i+1:]...)
		}
	}
	_ = applySession(session)
}

func removeSession(sessionName string) {
	resp, _ := helper.ExecuteOCCMD(nil, fmt.Sprintf("oc delete session %v", sessionName))
	fmt.Println(resp)
}

func getSessionName(cmd *cobra.Command) string {
	sessionName, err := cmd.Flags().GetString("session")
	if err == nil && sessionName != "" {
		return sessionName
	}
	random := randStringRunes(5)
	u, err := user.Current()
	if err != nil {
		return random
	}
	return u.Username + "-" + random
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
