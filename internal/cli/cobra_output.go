package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// WireOutput routes all command output to w (for tests and tooling).
func WireOutput(cmd *cobra.Command, w io.Writer) {
	cmd.SetOut(w)
	cmd.SetErr(w)
	for _, c := range cmd.Commands() {
		WireOutput(c, w)
	}
}
