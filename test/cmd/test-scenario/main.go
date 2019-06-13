package main

import (
	"fmt"
	"os"

	"github.com/maistra/istio-workspace/cmd/ike/config"
)

var (
	testImageName = ""
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("required arg 'scenario name' missing")
		os.Exit(-100)
	}

	testImageName = getTestImageName()

	scenarios := map[string]func(){
		"scenario-1": TestScenario1ThreeServicesInSequence,
		"scenario-2": TestScenario2ThreeServicesInSequenceDeploymentConfig,
	}
	scenario := os.Args[1]
	if f, ok := scenarios[scenario]; ok {
		f()

	} else {
		fmt.Println("Scenario not found", scenario)
		os.Exit(-101)
	}
}

func getTestImageName() string {
	reg := ""
	repo := "aslakknutsen"
	image := "istio-workspace-test"
	tag := "latest"

	if s, f := os.LookupEnv(config.EnvDockerRegistry); f {
		reg = s + "/"
	}
	if s, f := os.LookupEnv(config.EnvDockerRepository); f {
		repo = s
	}
	if s, f := os.LookupEnv("IKE_TEST_IMAGE_NAME"); f {
		image = s
	}
	if s, f := os.LookupEnv("COMMIT"); f {
		tag = s
	}

	return fmt.Sprintf("%v%v/%v:%v", reg, repo, image, tag)
}
