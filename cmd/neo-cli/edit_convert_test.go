package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/format"
)

func TestConvertCommandReEncodesTheImage(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)
	dir := t.TempDir()

	out, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "convert", input.URL+"/image.png", "--format", "webp", "--out", dir,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	files := writtenFiles(t, dir)
	if len(files) != 1 || !strings.HasSuffix(files[0], ".webp") {
		t.Fatalf("files = %v, want one webp", files)
	}

	converted := readFile(t, files[0])
	if string(converted) == string(png) {
		t.Error("converted bytes are the input, want a re-encode")
	}

	img, err := format.Decode(converted)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 8 {
		t.Errorf("size = %v, want the source's 8x8", img.Bounds().Size())
	}

	if !strings.Contains(out, files[0]) {
		t.Errorf("stdout does not contain the saved path:\n%s", out)
	}
}

func TestConvertCommandRefusesTheSourceFormat(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "convert", input.URL+"/image.png", "--format", "png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "already png") {
		t.Fatalf("Execute() error = %v, want the already-in-format rejection", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestConvertCommandRequiresAFormat(t *testing.T) {
	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "convert", "http://example.test/image.png")
	if err == nil || !strings.Contains(err.Error(), "--format is required") {
		t.Errorf("Execute() error = %v, want a missing-format rejection", err)
	}
}

func TestConvertCommandRejectsAnUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "convert", "http://example.test/image.png", "--format", "bmp", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("Execute() error = %v, want an unknown format rejection", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestConvertCommandSurfacesAFetchFailure(t *testing.T) {
	server, _ := captureServer(t, http.StatusNotFound, "missing")
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "convert", server.URL+"/image.png", "--format", "webp", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "convert: read image") {
		t.Fatalf("Execute() error = %v, want a read failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestConvertCommandSurfacesAnUndecodableImage(t *testing.T) {
	server, _ := captureServer(t, http.StatusOK, "\x89PNG\r\n\x1a\n not really a png")
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig()},
		"edit", "convert", server.URL+"/image.png", "--format", "webp", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "decode source image") {
		t.Fatalf("Execute() error = %v, want a decode failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}
