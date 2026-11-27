package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/edit"
	"github.com/wishmatic/neo-mcp/internal/format"
)

type editFlags struct {
	square       bool
	circle       bool
	outputFormat string
	out          string
}

func newEditCommand(deps cli.Deps) *cobra.Command {
	var flags editFlags

	cmd := &cobra.Command{
		Use:   "edit <image-url>",
		Short: "Trim, pad, square, or circularise an image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editRun(cmd, deps, flags, args[0])
		},
	}

	cmd.Flags().BoolVar(&flags.square, "square", false,
		"center the trimmed content on a transparent square")
	cmd.Flags().BoolVar(&flags.circle, "circle", false,
		"cut the output into the inscribed circle; implies --square")
	cmd.Flags().Int("padding", 0, "transparent padding in pixels around the trimmed content")
	cmd.Flags().StringVar(&flags.outputFormat, "format", "",
		"output image format: png, jpeg, jxl, or webp")
	cmd.Flags().StringVar(&flags.out, "out", ".", "directory for the edited image")

	return cmd
}

func editRun(cmd *cobra.Command, deps cli.Deps, flags editFlags, imageURL string) error {
	cfg, err := cli.LoadConfig(deps)
	if err != nil {
		return err
	}

	target, err := cli.OutputFormat(cfg, flags.outputFormat)
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	resolver, err := cli.ImageResolver(deps, cfg)
	if err != nil {
		return err
	}

	source, err := resolver.Fetch(cmd.Context(), imageURL)
	if err != nil {
		return fmt.Errorf("edit: fetch image: %w", err)
	}

	out, err := editImage(source, editOptions(cmd, flags), target)
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	path, err := cli.WriteImage(flags.out, out, target.MediaType())
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	return cli.PrintPaths(deps.Out, []string{path})
}

func editOptions(cmd *cobra.Command, flags editFlags) edit.Options {
	opts, _ := edit.OptionsFor(edit.Request{
		IsCrop:   true,
		IsSquare: flags.square,
		IsCircle: flags.circle,
		Padding:  cli.OptionalInt(cmd, "padding"),
	})

	return opts
}

func editImage(data []byte, opts edit.Options, target format.Format) ([]byte, error) {
	src, err := format.Decode(data)
	if err != nil {
		return nil, err
	}

	out, err := edit.Apply(src, opts)
	if err != nil {
		return nil, err
	}

	return format.Encode(out, target)
}
