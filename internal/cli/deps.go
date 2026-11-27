package cli

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/wishmatic/neo-mcp/internal/config"
	"github.com/wishmatic/neo-mcp/internal/forge"
	"github.com/wishmatic/neo-mcp/internal/format"
	"github.com/wishmatic/neo-mcp/internal/resolve"
	"github.com/wishmatic/neo-mcp/internal/sourcemap"
)

type Deps struct {
	In       io.Reader
	Out      io.Writer
	Err      io.Writer
	Config   config.Config
	Forge    *forge.Client
	Resolver *resolve.Client
}

func DefaultDeps() Deps {
	return Deps{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}

func WithDefaults(deps Deps) Deps {
	if deps.In == nil {
		deps.In = os.Stdin
	}
	if deps.Out == nil {
		deps.Out = os.Stdout
	}
	if deps.Err == nil {
		deps.Err = os.Stderr
	}

	return deps
}

func LoadConfig(deps Deps) (config.Config, error) {
	if deps.Config != (config.Config{}) {
		return deps.Config, nil
	}

	return config.Load()
}

func ForgeClient(deps Deps, cfg config.Config) *forge.Client {
	if deps.Forge != nil {
		return deps.Forge
	}

	return forge.New(cfg.SDURL, zap.NewNop())
}

func ImageResolver(deps Deps, cfg config.Config) (*resolve.Client, error) {
	if deps.Resolver != nil {
		return deps.Resolver, nil
	}

	sources, err := sourcemap.Parse(cfg.ImageURLMap)
	if err != nil {
		return nil, fmt.Errorf("IMAGE_URL_MAP: %w", err)
	}

	return resolve.New(nil, "", sources)
}

func OutputFormat(cfg config.Config, requested string) (format.Format, error) {
	if requested != "" {
		return format.Parse(requested)
	}

	if cfg.DefaultOutput == "" {
		return format.Default, nil
	}

	return format.Parse(cfg.DefaultOutput)
}
