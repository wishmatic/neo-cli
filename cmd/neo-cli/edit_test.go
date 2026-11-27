package main

import (
	"bytes"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wishmatic/neo-mcp/internal/cli"
)

func TestEditCropCommandsRenderTheRequestedCanvas(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)

	tests := []struct {
		name string
		args []string
		side int
	}{
		{"trim without padding", []string{"trim"}, 4},
		{"square with the default padding", []string{"square"}, 68},
		{"square with explicit padding", []string{"square", "--padding", "4"}, 12},
		{"circle with explicit padding", []string{"circle", "--padding", "3"}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			args := append([]string{"edit"}, tt.args...)
			args = append(args, input.URL+"/image.png", "--out", dir)

			out, _, err := runCommand(t, cli.Deps{Config: testConfig()}, args...)
			if err != nil {
				t.Fatalf("Execute() error: %v", err)
			}

			files := writtenFiles(t, dir)
			if len(files) != 1 || !strings.HasSuffix(files[0], ".png") {
				t.Fatalf("files = %v, want one png", files)
			}

			config, _, err := image.DecodeConfig(bytes.NewReader(readFile(t, files[0])))
			if err != nil {
				t.Fatalf("DecodeConfig() error: %v", err)
			}

			if config.Width != tt.side || config.Height != tt.side {
				t.Errorf("size = %dx%d, want %dx%d", config.Width, config.Height, tt.side, tt.side)
			}

			if !strings.Contains(out, files[0]) {
				t.Errorf("stdout does not contain the saved path:\n%s", out)
			}
		})
	}
}

func TestEditCommandsReadALocalFile(t *testing.T) {
	png := testPNG(t)

	path := filepath.Join(t.TempDir(), "input.png")
	if err := os.WriteFile(path, png, 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	dir := t.TempDir()

	if _, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "square", path, "--out", dir,
	); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	files := writtenFiles(t, dir)
	if len(files) != 1 {
		t.Fatalf("files = %v, want one image", files)
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(readFile(t, files[0])))
	if err != nil {
		t.Fatalf("DecodeConfig() error: %v", err)
	}

	if config.Width != 68 || config.Height != 68 {
		t.Errorf("size = %dx%d, want the local file's 68x68", config.Width, config.Height)
	}
}

func TestEditCommandRejectsAnUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "trim", "http://example.test/image.png", "--format", "bmp", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("Execute() error = %v, want an unknown format rejection", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestEditCommandSurfacesAFetchFailure(t *testing.T) {
	server, _ := captureServer(t, http.StatusNotFound, "missing")
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "trim", server.URL+"/image.png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "read image") {
		t.Fatalf("Execute() error = %v, want a read failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestEditCommandRejectsADirectory(t *testing.T) {
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "trim", dir, "--out", t.TempDir(),
	)
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("Execute() error = %v, want a directory rejection", err)
	}
}

func TestEditCommandHelpListsEverySubcommand(t *testing.T) {
	out, _, err := runCommand(t, cli.Deps{Config: testConfig()}, "edit")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	for _, name := range []string{"trim", "square", "circle", "convert", "bgkill"} {
		if !strings.Contains(out, name) {
			t.Errorf("edit help does not mention %q:\n%s", name, out)
		}
	}
}
