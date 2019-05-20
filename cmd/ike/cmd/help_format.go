package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// EnhanceHelper wraps helper function with alternative formatting and templates (e.g. asciidoc) based on --helper-format flag
// Applies to all subcommands.
// This can be useful when automatically generating documentation for CLI
func EnhanceHelper(command *cobra.Command) {
	originalHelpFunc := command.HelpFunc()
	command.SetHelpFunc(func(cmd *cobra.Command, i []string) {
		helpFormat, err := cmd.Flags().GetString("help-format")
		if err != nil {
			return
		}

		if helpFormat == "standard" {
			return
		}

		cobra.AddTemplateFunc("localFlagsSlice", func(set *pflag.FlagSet) []pflag.Flag {
			flags := make([]pflag.Flag, 0)
			set.VisitAll(func(flag *pflag.Flag) {
				flags = append(flags, *flag)
			})
			return flags
		})
		cobra.AddTemplateFunc("type", func(flag *pflag.Flag) string {
			flagType := flag.Value.Type()
			if strings.Contains(flagType, "Slice") {
				return "takes comma-separated value as arguments and split them accordingly (`" + strings.Replace(flagType, "Slice", "", 1) + "`)"
			}
			return "`" + flagType + "`"
		})

		cmd.SetHelpTemplate(OnlyUsageString)

		if helpFormat == "adoc" {
			cmd.SetUsageTemplate(ADocHelpTable)
		} else {
			fmt.Printf("unknown help format: [%s]. using standard one\n", helpFormat)
		}

		originalHelpFunc(cmd, i)
	})
}

const (
	OnlyUsageString = "{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}"
	ADocHelpTable   = `{{if .HasAvailableLocalFlags}}{{ $tick := "` + "`" + `" }}
[cols="2,4,1,1"]
|===
|Option|Purpose|Format|Default
{{range localFlagsSlice .LocalFlags}}{{ if not .Hidden }}
|{{$tick}}--{{.Name}}{{$tick}} {{if .Shorthand}}({{$tick}}-{{.Shorthand}}{{$tick}}){{end}}
|{{.Usage}}
|{{type .}}
|{{.DefValue | trimTrailingWhitespaces}}{{end}}
{{end}}
|===
{{end}}`
)
