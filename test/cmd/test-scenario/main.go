package main

import (
	"fmt"
	"io"
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/generator"
	"github.com/maistra/istio-workspace/test/scenarios"
)

var Namespace = "default"

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("required arg 'scenario name' missing")
		os.Exit(-100)
	}

	generator.TestImageName = getTestImageName()
	if h, f := os.LookupEnv("IKE_SCENARIO_GATEWAY"); f {
		generator.GatewayHost = h
	}

	if h, f := os.LookupEnv("TEST_NAMESPACE"); f {
		Namespace = h
	}

	// FIX give better names
	testScenarios := map[string]func(io.Writer, string){
		"scenario-1":   scenarios.TestScenario1HTTPThreeServicesInSequence,
		"scenario-1.1": scenarios.TestScenario1GRPCThreeServicesInSequence,
		"scenario-2":   scenarios.TestScenario2ThreeServicesInSequenceDeploymentConfig,
		"demo":         scenarios.DemoScenario,
	}
	scenario := os.Args[1] //nolint:ifshort // scenario used in multiple locations
	if f, ok := testScenarios[scenario]; ok {
		f(os.Stdout, Namespace)
	} else {
		fmt.Println("Scenario not found", scenario)
		os.Exit(-101)
	}
}

func getTestImageName() string {
	reg := ""
	repo := "maistra"
	image := "istio-workspace-test"
	tag := "latest"

	if s, f := os.LookupEnv(config.EnvImageRegistry); f {
		reg = s + "/"
	}
	if s, f := os.LookupEnv(config.EnvImageDevRepository); f {
		repo = s
	}
	if s, f := os.LookupEnv(config.EnvTestImage); f {
		image = s
	}
	if s, f := os.LookupEnv(config.EnvImageTag); f {
		tag = s
	}

	return fmt.Sprintf("%v%v/%v:%v", reg, repo, image, tag)
}
