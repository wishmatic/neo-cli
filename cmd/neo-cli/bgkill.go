package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/forge"
)

type bgkillFlags struct {
	model        string
	fullMode     bool
	outputFormat string
	out          string
}

func newBgkillCommand(deps cli.Deps) *cobra.Command {
	var flags bgkillFlags

	cmd := &cobra.Command{
		Use:   "bgkill <image-url>",
		Short: "Remove the background from an image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return bgkillRun(cmd, deps, flags, args[0])
		},
	}

	cmd.Flags().StringVar(&flags.model, "model", "General", "BiRefNet model to load")
	cmd.Flags().BoolVar(&flags.fullMode, "full-mode", false, "run the model in fp32 instead of fp16")
	cmd.Flags().StringVar(&flags.outputFormat, "format", "",
		"output image format: png, jpeg, jxl, or webp")
	cmd.Flags().StringVar(&flags.out, "out", ".", "directory for the foreground image")

	return cmd
}

func bgkillRun(cmd *cobra.Command, deps cli.Deps, flags bgkillFlags, imageURL string) error {
	cfg, err := cli.LoadConfig(deps)
	if err != nil {
		return err
	}

	target, err := cli.OutputFormat(cfg, flags.outputFormat)
	if err != nil {
		return fmt.Errorf("bgkill: %w", err)
	}

	resolver, err := cli.ImageResolver(deps, cfg)
	if err != nil {
		return err
	}

	source, err := resolver.Fetch(cmd.Context(), imageURL)
	if err != nil {
		return fmt.Errorf("bgkill: fetch image: %w", err)
	}

	out, err := cli.ForgeClient(deps, cfg).Bgkill(cmd.Context(), forge.BgkillRequest{
		ModelName:  flags.model,
		ImageData:  source,
		IsFullMode: flags.fullMode,
	})
	if err != nil {
		return fmt.Errorf("bgkill: %w", err)
	}

	paths, err := cli.WriteImages(flags.out, [][]byte{out}, target)
	if err != nil {
		return fmt.Errorf("bgkill: %w", err)
	}

	return cli.PrintPaths(deps.Out, paths)
}
