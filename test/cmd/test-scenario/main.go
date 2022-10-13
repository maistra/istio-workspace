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
		printAvailableScenarios()
		os.Exit(0)
	}

	if h, f := os.LookupEnv("TEST_NAMESPACE"); f {
		Namespace = h
	}

	if generateScenario, ok := scenarios.TestScenarios[os.Args[1]]; ok {
		generateScenario(Namespace, getTestImageName(), generator.WrapInYamlPrinter(os.Stdout))
	} else {
		fmt.Printf("Scenario [%s] not found!\n", os.Args[1])
		printAvailableScenarios()
		os.Exit(1)
	}
}

func printAvailableScenarios() {
	fmt.Println("Available scenarios:")
	for s := range scenarios.TestScenarios {
		fmt.Printf(" * %s\n", s)
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
