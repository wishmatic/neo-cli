package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/format"
)

type convertFlags struct {
	format string
	out    string
}

func newEditConvertCommand(deps cli.Deps) *cobra.Command {
	var flags convertFlags

	cmd := &cobra.Command{
		Use:   "convert <image>",
		Short: "Re-encode an image as png, jpeg, jxl, or webp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return convertRun(cmd, deps, flags, args[0])
		},
	}

	cmd.Flags().StringVar(&flags.format, "format", "", "format to convert to: png, jpeg, jxl, or webp")
	cmd.Flags().StringVar(&flags.out, "out", ".", "directory for the converted image")

	return cmd
}

func convertRun(cmd *cobra.Command, deps cli.Deps, flags convertFlags, image string) error {
	target, err := convertTarget(flags.format)
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	cfg, err := cli.LoadConfig(deps)
	if err != nil {
		return err
	}

	resolver, err := cli.ImageResolver(deps, cfg)
	if err != nil {
		return err
	}

	source, err := cli.ReadImage(cmd.Context(), resolver, image)
	if err != nil {
		return fmt.Errorf("convert: read image: %w", err)
	}

	out, err := format.Reformat(source, target)
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	path, err := cli.WriteImage(flags.out, out, target.MediaType())
	if err != nil {
		return fmt.Errorf("convert: %w", err)
	}

	return cli.PrintPaths(deps.Out, []string{path})
}

func convertTarget(value string) (format.Format, error) {
	if value == "" {
		return "", fmt.Errorf("--format is required")
	}

	return format.Parse(value)
}
