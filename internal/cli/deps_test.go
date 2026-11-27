package cli

import (
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/wishmatic/neo-mcp/internal/config"
	"github.com/wishmatic/neo-mcp/internal/forge"
	"github.com/wishmatic/neo-mcp/internal/format"
	"github.com/wishmatic/neo-mcp/internal/resolve"
)

func TestLoadConfigUsesTheInjectedConfig(t *testing.T) {
	cfg := config.Config{SDURL: "http://forge.test:7860"}

	loaded, err := LoadConfig(Deps{Config: cfg})
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loaded != cfg {
		t.Errorf("config = %+v, want the injected %+v", loaded, cfg)
	}
}

func TestLoadConfigReadsTheEnvironment(t *testing.T) {
	t.Setenv("SD_URL", "http://env.test:7860")
	t.Setenv("OUTPUT_FORMAT", "jpeg")

	loaded, err := LoadConfig(Deps{})
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loaded.SDURL != "http://env.test:7860" {
		t.Errorf("SD_URL = %q, want the environment value", loaded.SDURL)
	}
	if loaded.DefaultOutput != "jpeg" {
		t.Errorf("OUTPUT_FORMAT = %q, want jpeg", loaded.DefaultOutput)
	}
}

func TestOutputFormatFallsBackToTheDefault(t *testing.T) {
	target, err := OutputFormat(config.Config{}, "")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if target != format.Default {
		t.Errorf("format = %q, want %q", target, format.Default)
	}
}

func TestOutputFormatFollowsTheConfig(t *testing.T) {
	target, err := OutputFormat(config.Config{DefaultOutput: "png"}, "")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if target != format.PNG {
		t.Errorf("format = %q, want png", target)
	}
}

func TestOutputFormatPrefersTheRequest(t *testing.T) {
	target, err := OutputFormat(config.Config{DefaultOutput: "png"}, "jpeg")
	if err != nil {
		t.Fatalf("OutputFormat() error: %v", err)
	}

	if target != format.JPEG {
		t.Errorf("format = %q, want jpeg", target)
	}
}

func TestOutputFormatRejectsAnUnknownFormat(t *testing.T) {
	if _, err := OutputFormat(config.Config{}, "bmp"); err == nil {
		t.Fatal("OutputFormat() error = nil, want an unknown format rejection")
	}
}

func TestInjectedClientsAreReturned(t *testing.T) {
	forgeClient := forge.New("http://forge.test:7860", zap.NewNop())

	resolver, err := resolve.New(nil, "", nil)
	if err != nil {
		t.Fatalf("resolve.New() error: %v", err)
	}

	deps := Deps{Forge: forgeClient, Resolver: resolver}

	if got := ForgeClient(deps, config.Config{}); got != forgeClient {
		t.Error("ForgeClient() did not return the injected client")
	}

	if got, err := ImageResolver(deps, config.Config{}); err != nil || got != resolver {
		t.Errorf("ImageResolver() = %v, %v; want the injected resolver", got, err)
	}
}

func TestImageResolverRejectsABadSourceMap(t *testing.T) {
	_, err := ImageResolver(Deps{}, config.Config{ImageURLMap: "nonsense"})
	if err == nil || !strings.Contains(err.Error(), "IMAGE_URL_MAP") {
		t.Errorf("ImageResolver() error = %v, want the source map named", err)
	}
}
