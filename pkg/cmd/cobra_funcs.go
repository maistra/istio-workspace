package cmd

import "github.com/spf13/cobra"

func VisitAll(command *cobra.Command, apply func(command *cobra.Command)) {
	for _, child := range command.Commands() {
		VisitAll(child, apply)
	}
	apply(command)
}
