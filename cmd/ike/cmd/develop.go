package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os/user"
	"strings"
	"time"

	"github.com/aslakknutsen/istio-workspace/cmd/ike/config"

	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	helper "github.com/aslakknutsen/istio-workspace/pkg/istio"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const telepresenceBin = "telepresence"

var excludeLogs = []string{"*.log"}

// NewDevelopCmd creates instance of "develop" Cobra Command with flags and execution logic defined
func NewDevelopCmd() *cobra.Command {

	developCmd := &cobra.Command{
		Use:   "develop",
		Short: "Starts the development flow",

		PreRunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]
			if !BinaryExists(telepresenceBin, "Head over to https://www.telepresence.io/reference/install for installation instructions.\n") {
				return fmt.Errorf("unable to find %s on your $PATH", telepresenceBin)
			}

			return config.SyncFlags(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint[:unparam]

			closeSession, err := prepareSession(cmd)
			if err != nil {
				return err
			}
			defer closeSession()

			if err := build(cmd); err != nil {
				return err
			}

			done := make(chan gocmd.Status, 1)
			defer close(done)

			arguments := parseArguments(cmd)

			go func() {
				tp := gocmd.NewCmdOptions(StreamOutput, telepresenceBin, arguments...)
				RedirectStreams(tp, cmd.OutOrStdout(), cmd.OutOrStderr(), done)
				ShutdownHook(tp, done)
				Start(tp, done)
			}()

			finalStatus := <-done

			return finalStatus.Error
		},
	}

	developCmd.Flags().StringP("deployment", "d", "", "name of the deployment or deployment config")
	developCmd.Flags().StringP("port", "p", "8000", "port to be exposed in format local[:remote]")
	developCmd.Flags().StringP(runFlagName, "r", "", "command to run your application")
	developCmd.Flags().StringP(buildFlagName, "b", "", "command to build your application before run")
	developCmd.Flags().Bool(noBuildFlagName, false, "always skips build")
	developCmd.Flags().Bool("watch", false, "enables watch")
	developCmd.Flags().StringSliceP("watch-include", "w", []string{CurrentDir()}, "list of directories to watch")
	developCmd.Flags().StringSlice("watch-exclude", excludeLogs, "list of patterns to exclude (defaults to telepresence.log which is always excluded)")
	developCmd.Flags().Int64("watch-interval", 500, "watch interval (in ms)")
	if err := developCmd.Flags().MarkHidden("watch-interval"); err != nil {
		log.Error(err, "failed while trying to hide a flag")
	}
	developCmd.Flags().StringP("method", "m", "inject-tcp", "telepresence proxying mode - see https://www.telepresence.io/reference/methods")
	developCmd.Flags().StringP("session", "s", "", "create or join an existing session")

	developCmd.Flags().VisitAll(config.BindFullyQualifiedFlag(developCmd))

	_ = developCmd.MarkFlagRequired("deployment")
	_ = developCmd.MarkFlagRequired(runFlagName)

	return developCmd
}

func parseArguments(cmd *cobra.Command) []string {
	run := cmd.Flag(runFlagName).Value.String()
	watch, _ := cmd.Flags().GetBool("watch")
	runArgs := strings.Split(run, " ") // default value

	if watch {
		runArgs = []string{
			"ike", "watch",
			"--dir", flag(cmd.Flags(), "watch-include"),
			"--exclude", flag(cmd.Flags(), "watch-exclude"),
			"--interval", cmd.Flag("watch-interval").Value.String(),
			"--" + runFlagName, run,
		}
		if cmd.Flag(buildFlagName).Changed {
			runArgs = append(runArgs, "--"+buildFlagName, cmd.Flag(buildFlagName).Value.String())
		}
	}

	return append([]string{
		"--deployment", cmd.Flag("deployment").Value.String(),
		"--expose", cmd.Flag("port").Value.String(),
		"--method", cmd.Flag("method").Value.String(),
		"--run"}, runArgs...)
}

func flag(flags *pflag.FlagSet, name string) string {
	slice, _ := flags.GetStringSlice(name)
	return fmt.Sprintf(`"%s"`, strings.Join(slice, ","))
}

// Session handling

func prepareSession(cmd *cobra.Command) (func(), error) {
	sessionName := getSessionName(cmd)
	deploymentName, _ := cmd.Flags().GetString("deployment")
	serviceName, err := createOrJoinSession(sessionName, deploymentName)
	if err != nil {
		return func() {}, err
	}
	cmd.Flags().Set("deployment", serviceName) // HACK: pass arguments around, not flags?

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
	applySession(session)
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

func applySession(session istiov1alpha1.Session) error {
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sessionData := string(b)
	fmt.Println(sessionData)

	resp, err := helper.ExecuteOCCMD(&sessionData, fmt.Sprintf("oc apply -f -"))
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func waitForRefToComplete(sessionName, ref string) (string, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second) // wait 1 s to let the server attempt to prepare first
		sessionStatus, err := getSession(sessionName)
		if err != nil {
			return "", nil
		}

		for _, refs := range sessionStatus.Status.Refs {
			if refs.Name == ref {
				for _, res := range refs.Resources {
					if *res.Kind == "Deployment" || *res.Kind == "DeploymentConfig" {
						return *res.Name, nil
					}
				}
			}
		}
	}
	return "", errors.New("no Deployment or DeploymentConfig found to target")
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
	applySession(session)
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
