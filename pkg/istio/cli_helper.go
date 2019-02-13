package istio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	istionetwork "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/networking/v1alpha3"
)

// TODO: Temp workaround for non functional istio API

func getDestinationRuleMapped(namespace, name string) (*istionetwork.DestinationRule, error) {
	body, err := getDestinationRule(namespace, name)
	if err != nil {
		return nil, err
	}
	dr := &istionetwork.DestinationRule{}
	err = json.Unmarshal([]byte(body), &dr)

	return dr, err
}

func setDestinationRule(namespace string, dr *istionetwork.DestinationRule) error {
	body, err := json.Marshal(dr)
	if err != nil {
		return err
	}
	sbody := string(body)
	_, err = executeCMD(&sbody, []string{"-c", fmt.Sprintf("oc apply -f - --namespace=%v -o json", namespace)})
	return err
}

func getVirtualServiceMapped(namespace, name string) (*istionetwork.VirtualService, error) {
	body, err := getVirtualService(namespace, name)
	if err != nil {
		return nil, err
	}
	dr := &istionetwork.VirtualService{}
	err = json.Unmarshal([]byte(body), &dr)
	return dr, err
}

func setVirtualService(namespace string, vs *istionetwork.VirtualService) error {
	body, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	sbody := string(body)
	_, err = executeCMD(&sbody, []string{"-c", fmt.Sprintf("oc apply -f - --namespace=%v -o json", namespace)})
	return err
}

func getDestinationRule(namespace, name string) (string, error) {
	return executeCMD(nil, []string{"-c", fmt.Sprintf("oc get destinationrule %v --namespace=%v -o json", name, namespace)})
}

func getVirtualService(namespace, name string) (string, error) {
	return executeCMD(nil, []string{"-c", fmt.Sprintf("oc get virtualservice %v --namespace=%v -o json", name, namespace)})
}

func executeCMD(input *string, cmdArgs []string) (string, error) {
	cmdName := "/usr/bin/sh"

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
			io.WriteString(stdin, *input)

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
