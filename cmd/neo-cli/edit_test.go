package main

import (
	"bytes"
	"image"
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/neo-mcp/internal/cli"
)

func TestEditCommandRendersTheRequestedCanvas(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)

	tests := []struct {
		name string
		args []string
		side int
	}{
		{"square with the default padding", []string{"--square"}, 68},
		{"circle with explicit padding", []string{"--circle", "--padding", "5"}, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			args := append([]string{"edit", input.URL + "/image.png", "--out", dir}, tt.args...)

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

func TestEditCommandRejectsAnUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "http://example.test/image.png", "--format", "bmp", "--out", dir,
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
		"edit", server.URL+"/image.png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "edit: fetch image") {
		t.Fatalf("Execute() error = %v, want a fetch failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestEditCommandSurfacesAnUndecodableImage(t *testing.T) {
	server, _ := captureServer(t, http.StatusOK, "not an image")
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", server.URL+"/image.png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "decode image") {
		t.Fatalf("Execute() error = %v, want a decode failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}
