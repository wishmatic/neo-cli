package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/edit"
	"github.com/wishmatic/neo-mcp/internal/format"
)

func newEditCommand(deps cli.Deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Trim, pad, square, circularise, re-encode, or remove a background",
		Long: "Operate on an image given as a local path or an address the resolver can read. Every " +
			"subcommand writes its result to --out and prints the path.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newEditCropCommand(deps, cropCommand{
			name:  "trim",
			short: "Trim an image to the bounding box of its visible content",
		}),
		newEditCropCommand(deps, cropCommand{
			name:           "square",
			short:          "Trim an image and center it on a transparent square",
			isSquare:       true,
			defaultPadding: edit.DefaultSquarePadding,
		}),
		newEditCropCommand(deps, cropCommand{
			name:     "circle",
			short:    "Trim an image and cut it into the inscribed circle",
			isCircle: true,
		}),
		newEditConvertCommand(deps),
		newEditBgkillCommand(deps),
	)

	return cmd
}

type cropCommand struct {
	name           string
	short          string
	isSquare       bool
	isCircle       bool
	defaultPadding int
}

func newEditCropCommand(deps cli.Deps, spec cropCommand) *cobra.Command {
	var (
		padding      int
		outputFormat string
		out          string
	)

	cmd := &cobra.Command{
		Use:   spec.name + " <image>",
		Short: spec.short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, _ := edit.OptionsFor(edit.Request{
				IsCrop:   true,
				IsSquare: spec.isSquare,
				IsCircle: spec.isCircle,
				Padding:  &padding,
			})

			return editCropRun(cmd, deps, opts, outputFormat, out, args[0])
		},
	}

	cmd.Flags().IntVar(&padding, "padding", spec.defaultPadding,
		"transparent padding in pixels around the trimmed content")
	cmd.Flags().StringVar(&outputFormat, "format", "",
		"output image format: png, jpeg, jxl, or webp")
	cmd.Flags().StringVar(&out, "out", ".", "directory for the edited image")

	return cmd
}

func editCropRun(
	cmd *cobra.Command, deps cli.Deps, opts edit.Options, outputFormat, out, image string,
) error {
	cfg, err := cli.LoadConfig(deps)
	if err != nil {
		return err
	}

	target, err := cli.OutputFormat(cfg, outputFormat)
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	resolver, err := cli.ImageResolver(deps, cfg)
	if err != nil {
		return err
	}

	source, err := cli.ReadImage(cmd.Context(), resolver, image)
	if err != nil {
		return fmt.Errorf("edit: read image: %w", err)
	}

	edited, err := editImage(source, opts, target)
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	path, err := cli.WriteImage(out, edited, target.MediaType())
	if err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	return cli.PrintPaths(deps.Out, []string{path})
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
