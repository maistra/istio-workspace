package flag

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateOptions will return a slice of pflag.Value with name and associated abbreviation.
// This limited set of values can then be bound to a cobra flag to limit choices for a given
// input. On top of that custom, completion can be defined.
//
// Example:
//    testCmd := &cobra.Command{...}
//    beerStyles := flag.CreateOptions("stout", "s", "ale", "a", "kolsch", "k")
//    beerStyle := beerStyles[0]
//    testCmd.Flags().Var(&beerStyle, "style", "beer styles")
//     _ = testCmd.RegisterFlagCompletionFunc("type", flag.CompletionFor(beerStyles))
func CreateOptions(namesAndAbbrevs ...string) []NameAndAbbrev {
	var values = []NameAndAbbrev{}
	var availableNames = func() []NameAndAbbrev {
		return values
	}

	for i := 0; i < len(namesAndAbbrevs); i += 2 {
		values = append(values, NameAndAbbrev{name: namesAndAbbrevs[i], abbrev: namesAndAbbrevs[i+1], avail: availableNames})
	}

	return values
}

// CompletionFor registers custom autocompletion for limited set of NameAndAbbrev values.
// This can be used to generate autocompletion for shell of your choice.
func CompletionFor(values []NameAndAbbrev) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	completions := []string{}
	for _, value := range values {
		completions = append(completions, value.String())
	}

	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completions, cobra.ShellCompDirectiveDefault
	}
}

// NameAndAbbrev holds flags name and its abbreviations.
// It also holds reference to all possible values, so it can validate itself
// and provide autocompletion.
type NameAndAbbrev struct {
	name,
	abbrev string
	avail func() []NameAndAbbrev
}

var _ pflag.Value = (*NameAndAbbrev)(nil)

// String is used both by fmt.Print and by Cobra in help text.
func (e *NameAndAbbrev) String() string {
	return e.name
}

// Set must have pointer receiver, so it doesn't change the value of a copy.
func (e *NameAndAbbrev) Set(v string) error {
	availableOpts := e.avail()
	for _, value := range availableOpts {
		if v == value.name || v == value.abbrev {
			*e = value

			return nil
		}
	}

	return fmt.Errorf("must be one of %s", e.Hint()) //nolint:goerr113 //reason it's dynamically constructed based on available options
}

// Type is only used in help text.
func (e *NameAndAbbrev) Type() string {
	return fmt.Sprintf("%s (or %s)", e.name, e.abbrev)
}

// Hint provides list of possible values and their abbreviations in the slice.
func (e *NameAndAbbrev) Hint() string {
	availableOpts := e.avail()
	hints := []string{}
	for _, opt := range availableOpts {
		hints = append(hints, fmt.Sprintf("%s (%s)", opt.name, opt.abbrev))
	}

	return strings.Join(hints, ", ")
}
