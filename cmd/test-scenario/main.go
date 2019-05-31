package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("required arg 'scenario name' missing")
		os.Exit(-100)
	}
	scenarios := map[string]func(){
		"scenario-1": TestScenario1,
	}
	scenario := os.Args[1]
	if f, ok := scenarios[scenario]; ok {
		f()

	} else {
		fmt.Println("Scenario not found", scenario)
		os.Exit(-101)
	}
}
