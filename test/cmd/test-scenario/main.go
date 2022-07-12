package main

import (
	"fmt"
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/diagram"
	"github.com/maistra/istio-workspace/test/cmd/test-scenario/generator"
)

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
		generator.Namespace = h
	}

	scenarios := map[string]func(generator.Printer){
		"scenario-1":   generator.TestScenario1HTTPThreeServicesInSequence,
		"scenario-1.1": generator.TestScenario1GRPCThreeServicesInSequence,
		"scenario-2":   generator.TestScenario2ThreeServicesInSequenceDeploymentConfig,
		"demo":         generator.DemoScenario,
	}
	scenario := os.Args[1] //nolint:ifshort // scenario used in multiple locations
	if f, ok := scenarios[scenario]; ok {
		f(generator.NewSysOutPrinter(os.Stdout))
		f(diagram.NewPrinter(scenario, os.Stdout))
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
