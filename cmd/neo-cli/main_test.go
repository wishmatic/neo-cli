package main

import (
	"strings"
	"testing"

	"github.com/wishmatic/neo-mcp/internal/cli"
)

func TestHelpListsEveryCommandWithoutConfig(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "help")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	for _, command := range []string{"serve", "gen", "edit"} {
		if !strings.Contains(out, command) {
			t.Errorf("help output does not mention %q:\n%s", command, out)
		}
	}
}

func TestHelpForASubcommand(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "help", "gen", "image")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !strings.Contains(out, "neo-cli gen image <prompt...>") {
		t.Errorf("help output = %q, want the gen image usage", out)
	}
}

func TestServeHelpDoesNotLoadConfig(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{}, "serve", "--help")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if !strings.Contains(out, "neo-cli serve") {
		t.Errorf("help output = %q, want the serve usage", out)
	}
}
