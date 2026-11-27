package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wishmatic/neo-mcp/internal/cli"
)

func main() {
	if err := newRoot(cli.DefaultDeps()).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRoot(deps cli.Deps) *cobra.Command {
	deps = cli.WithDefaults(deps)

	root := &cobra.Command{
		Use:   "neo-cli",
		Short: "Use the Neo image tools from the command line",
		Long: "Generate, edit, convert, and remove backgrounds from images without an MCP " +
			"client.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetIn(deps.In)
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)

	root.AddCommand(
		newServeCommand(),
		newGenCommand(deps),
		newEditCommand(deps),
		newConvertCommand(deps),
		newBgkillCommand(deps),
	)

	return root
}
