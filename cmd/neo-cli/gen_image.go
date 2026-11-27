package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/forge"
	"github.com/wishmatic/neo-mcp/internal/resolve"
)

const (
	forgeDefaultSampler   = "DPM++ 2M"
	forgeDefaultScheduler = "Automatic"
)

type forgeFlags struct {
	model             string
	preset            string
	vaeAndTextModels  []string
	negativePrompt    string
	sampler           string
	scheduler         string
	steps             int
	width             int
	height            int
	cfgScale          float64
	seed              int
	initImageURL      string
	denoisingStrength float64
	enableHR          bool
	hrScale           float64
	hrUpscaler        string
	hrSecondPassSteps int
	hrCFG             float64
	outputFormat      string
	out               string
}

func newGenImageCommand(deps cli.Deps) *cobra.Command {
	var flags forgeFlags

	cmd := &cobra.Command{
		Use:   "image <prompt...>",
		Short: "Generate an image with the local Forge Neo instance",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return forgeRun(cmd.Context(), deps, flags, args)
		},
	}

	cmd.Flags().StringVar(&flags.model, "model", "", "checkpoint (model) filename to load")
	cmd.Flags().StringVar(&flags.preset, "forge-preset", "", "Forge UI preset to apply")
	cmd.Flags().StringArrayVar(&flags.vaeAndTextModels, "vae-and-text-model", nil,
		"VAE or text encoder filename to load; repeat for more")
	cmd.Flags().StringVar(&flags.negativePrompt, "negative-prompt", "",
		"things to avoid in the generated image")
	cmd.Flags().StringVar(&flags.sampler, "sampler-name", forgeDefaultSampler, "sampler to use")
	cmd.Flags().StringVar(&flags.scheduler, "scheduler", forgeDefaultScheduler, "scheduler to use")
	cmd.Flags().IntVar(&flags.steps, "steps", 20, "number of sampling steps")
	cmd.Flags().IntVar(&flags.width, "width", 0, "output width in pixels")
	cmd.Flags().IntVar(&flags.height, "height", 0, "output height in pixels")
	cmd.Flags().Float64Var(&flags.cfgScale, "cfg-scale", 7, "classifier-free guidance scale")
	cmd.Flags().IntVar(&flags.seed, "seed", -1, "random seed; -1 is random")
	cmd.Flags().StringVar(&flags.initImageURL, "init-image-url", "",
		"image to transform instead of generating from scratch")
	cmd.Flags().Float64Var(&flags.denoisingStrength, "denoising-strength", 0,
		"how much of the init image to change")
	cmd.Flags().BoolVar(&flags.enableHR, "enable-hr", false, "enable hi-res second-pass upscaling")
	cmd.Flags().Float64Var(&flags.hrScale, "hr-scale", 0, "hi-res upscaling factor")
	cmd.Flags().StringVar(&flags.hrUpscaler, "hr-upscaler", "", "hi-res upscaler to use")
	cmd.Flags().IntVar(&flags.hrSecondPassSteps, "hr-second-pass-steps", 0,
		"steps for the hi-res second pass")
	cmd.Flags().Float64Var(&flags.hrCFG, "hr-cfg", 0, "CFG scale for the hi-res second pass")
	cmd.Flags().StringVar(&flags.outputFormat, "format", "",
		"output image format: png, jpeg, jxl, or webp")
	cmd.Flags().StringVar(&flags.out, "out", ".", "directory for the generated image")

	return cmd
}

func forgeRun(ctx context.Context, deps cli.Deps, flags forgeFlags, args []string) error {
	prompt, err := promptFromArgs(args)
	if err != nil {
		return err
	}

	if flags.model == "" {
		return fmt.Errorf("forge: --model is required")
	}
	if flags.preset == "" {
		return fmt.Errorf("forge: --forge-preset is required")
	}

	cfg, err := cli.LoadConfig(deps)
	if err != nil {
		return err
	}

	target, err := cli.OutputFormat(cfg, flags.outputFormat)
	if err != nil {
		return fmt.Errorf("forge: %w", err)
	}

	resolver, err := cli.ImageResolver(deps, cfg)
	if err != nil {
		return err
	}

	initImage, err := fetchInitImage(ctx, resolver, flags.initImageURL)
	if err != nil {
		return fmt.Errorf("forge: fetch init image: %w", err)
	}

	images, err := cli.ForgeClient(deps, cfg).Generate(ctx, flags.request(prompt, initImage))
	if err != nil {
		return fmt.Errorf("forge: %w", err)
	}

	paths, err := cli.WriteImages(flags.out, images, target)
	if err != nil {
		return fmt.Errorf("forge: %w", err)
	}

	return cli.PrintPaths(deps.Out, paths)
}

func (f forgeFlags) request(prompt string, initImage []byte) forge.GenerateRequest {
	return forge.GenerateRequest{
		InitImage: initImage,

		Checkpoint:             f.model,
		ForgePreset:            f.preset,
		ForgeAdditionalModules: f.vaeAndTextModels,

		Prompt:         prompt,
		NegativePrompt: f.negativePrompt,

		Sampler:   f.sampler,
		Scheduler: f.scheduler,
		Steps:     f.steps,

		Width:    f.width,
		Height:   f.height,
		CFGScale: f.cfgScale,

		DenoisingStrength: f.denoisingStrength,

		Seed: f.seed,

		EnableHR:          f.enableHR,
		HRScale:           f.hrScale,
		HRUpscaler:        f.hrUpscaler,
		HRSecondPassSteps: f.hrSecondPassSteps,
		HRCFGScale:        f.hrCFG,
	}
}

func fetchInitImage(ctx context.Context, resolver *resolve.Client, address string) ([]byte, error) {
	if address == "" {
		return nil, nil
	}

	return resolver.Fetch(ctx, address)
}
