package istio

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func executeCMD(input *string, cmdArgs []string) (string, error) { //nolint[:unused]
	cmdName := "sh"

	var buf bytes.Buffer
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	if input != nil {
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, *input) //nolint[:errcheck]

		}()
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return buf.String(), err
	}

	return buf.String(), nil
}

func addOCAuth(cmd string) string { //nolint[:unused]
	config, err := k8sConfig.GetConfig()
	if err != nil {
		return cmd
	}

	newCmd := cmd

	if config.CAFile != "" {
		newCmd = fmt.Sprintf(newCmd+" --certificate-authority='%v'", config.CAFile)
	}
	if config.CertFile != "" {
		newCmd = fmt.Sprintf(newCmd+" --client-key='%v'", config.CertFile)
	}
	if config.BearerToken != "" {
		newCmd = fmt.Sprintf(newCmd+" --token='%v'", config.BearerToken)
	}

	if config.Host != "" {
		newCmd = fmt.Sprintf(newCmd+" --server='%v'", config.Host)
	}
	newCmd = fmt.Sprintf(newCmd+" --insecure-skip-tls-verify=%v", config.Insecure)

	return newCmd
}

func removeToken(cmd string) string { //nolint[:unused]
	r := regexp.MustCompile(`\-\-token='(.+)'`)
	return r.ReplaceAllString(cmd, "--token='xxxx'")
}

// ExecuteOCCMD executes OC commands and add auth info found in env
func ExecuteOCCMD(input *string, cmdArg string) (string, error) {
	cmd := addOCAuth(cmdArg)
	output, err := executeCMD(input, []string{"-c", cmd})
	if err != nil { // TODO: Handle error else where
		fmt.Println("Failed to execute", removeToken(cmd))
		fmt.Println("Output:", output)
	}
	return output, err
}
