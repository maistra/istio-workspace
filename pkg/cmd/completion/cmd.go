package completion

import (
	"bytes"
	"io"
	"os"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
)

const (
	eg = `
  ### generate completion code for bash
  source <(ike completion bash)

  ### generate completion code for zsh
  source <(ike completion zsh)
`
)

// NewCmd creates completion cmd provides autocomplete functionality for different shell environments.
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "completion [SHELL]",
		Short:        "Prints shell completion scripts",
		Long:         "This command provides shell completion code for bash and zsh",
		Example:      eg,
		SilenceUsage: true,
		ValidArgs:    []string{"bash", "zsh"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return errors.WrapIf(cmd.Root().GenBashCompletion(os.Stdout), "failed configuring autocompletion for bash")
			case "zsh":
				return errors.WrapIf(runCompletionZsh(os.Stdout, cmd.Root()), "failed configuring autocompletion for zsh")
			}

			return nil
		},
	}
}

// runCompletionZsh generate the zsh completion the same way kubectl is doing it
// https://git.io/fjRRc
// We are not using the builtin zsh completion that comes from cobra but instead doing it
// via bashcompinit and use the GenBashCompletion then
// This allows us the user to simply do a `source <(ike completion zsh)` to get
// zsh completion.
func runCompletionZsh(out io.Writer, ike *cobra.Command) error {
	zshHead := "#compdef ike\n"

	if _, err := out.Write([]byte(zshHead)); err != nil {
		return errors.Wrap(err, "failed to write to out stream")
	}

	if _, err := out.Write([]byte(zshInitialization)); err != nil {
		return errors.Wrap(err, "failed to write to out stream")
	}

	buf := new(bytes.Buffer)
	if err := ike.GenBashCompletion(buf); err != nil {
		return errors.Wrap(err, "failed to generate bash completion")
	}
	if _, err := out.Write(buf.Bytes()); err != nil {
		return errors.Wrap(err, "failed to write to out stream")
	}

	zshTail := `
BASH_COMPLETION_EOF
}
__ike_bash_source <(__ike_convert_bash_to_zsh)
_complete ike 2>/dev/null
`
	if _, err := out.Write([]byte(zshTail)); err != nil {
		return errors.Wrap(err, "failed to write to out stream")
	}

	return nil
}
