package main

import (
	"fmt"
	"os"

	"github.com/maistra/istio-workspace/pkg/cmd/config"
	"github.com/maistra/istio-workspace/pkg/generator"
	"github.com/maistra/istio-workspace/test/scenarios"
)

var Namespace = "default"

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Available scenarios:")
		for s := range scenarios.TestScenarios {
			fmt.Printf(" * %s\n", s)
		}

		os.Exit(0)
	}

	if h, f := os.LookupEnv("IKE_SCENARIO_GATEWAY"); f {
		generator.GatewayHost = h
	}

	if h, f := os.LookupEnv("TEST_NAMESPACE"); f {
		Namespace = h
	}

	scenario := os.Args[1] //nolint:ifshort // scenario used in multiple locations
	if generateScenario, ok := scenarios.TestScenarios[scenario]; ok {
		generateScenario(Namespace, getTestImageName(), generator.WrapInYamlPrinter(os.Stdout))
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
