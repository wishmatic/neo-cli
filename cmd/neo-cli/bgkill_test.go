package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/forge"
)

func bgkillDeps(serverURL string) cli.Deps {
	return cli.Deps{Config: testConfig(), Forge: forge.New(serverURL, zap.NewNop())}
}

func TestBgkillCommandPostsTheImageAndWritesTheForeground(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)

	server, requests := captureServer(t, http.StatusOK,
		fmt.Sprintf(`{"output_image":%q}`, base64.StdEncoding.EncodeToString(png)))
	dir := t.TempDir()

	_, _, err := runCommand(t, bgkillDeps(server.URL),
		"bgkill", input.URL+"/image.png",
		"--model", "Portrait",
		"--full-mode",
		"--out", dir,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)
	if request.Path != "/birefnet/single" {
		t.Errorf("path = %s, want /birefnet/single", request.Path)
	}

	payload := request.jsonBody(t)

	assertField(t, payload, "model_name", "Portrait")
	assertField(t, payload, "image", base64.StdEncoding.EncodeToString(png))
	assertField(t, payload, "return_foreground", true)
	assertField(t, payload, "send_output", true)
	assertField(t, payload, "use_fp16", false)

	files := writtenFiles(t, dir)
	if len(files) != 1 || !strings.HasSuffix(files[0], ".png") {
		t.Fatalf("files = %v, want one png", files)
	}

	if string(readFile(t, files[0])) != string(png) {
		t.Errorf("%s content does not match the foreground", files[0])
	}
}

func TestBgkillCommandDefaultsToGeneralAndFP16(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)

	server, requests := captureServer(t, http.StatusOK,
		fmt.Sprintf(`{"output_image":%q}`, base64.StdEncoding.EncodeToString(png)))

	_, _, err := runCommand(t, bgkillDeps(server.URL),
		"bgkill", input.URL+"/image.png", "--out", t.TempDir(),
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	payload := nextRequest(t, requests).jsonBody(t)

	assertField(t, payload, "model_name", "General")
	assertField(t, payload, "use_fp16", true)
}

func TestBgkillCommandSurfacesAFetchFailure(t *testing.T) {
	input, _ := captureServer(t, http.StatusNotFound, "missing")
	dir := t.TempDir()

	_, _, err := runCommand(t, bgkillDeps("http://forge.test"),
		"bgkill", input.URL+"/image.png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "bgkill: fetch image") {
		t.Fatalf("Execute() error = %v, want a fetch failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestBgkillCommandSurfacesAnUndecodableForeground(t *testing.T) {
	input := imageServer(t, testPNG(t))

	server, _ := captureServer(t, http.StatusOK, `{"output_image":"!!!"}`)
	dir := t.TempDir()

	_, _, err := runCommand(t, bgkillDeps(server.URL),
		"bgkill", input.URL+"/image.png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "failed to decode foreground") {
		t.Fatalf("Execute() error = %v, want a decode failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestBgkillCommandRejectsAnUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()

	_, _, err := runCommand(t, bgkillDeps("http://forge.test"),
		"bgkill", "http://example.test/image.png", "--format", "bmp", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("Execute() error = %v, want an unknown format rejection", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}

func TestBgkillCommandSurfacesAnUpstreamError(t *testing.T) {
	input := imageServer(t, testPNG(t))

	server, _ := captureServer(t, http.StatusInternalServerError, "boom")
	dir := t.TempDir()

	_, _, err := runCommand(t, bgkillDeps(server.URL),
		"bgkill", input.URL+"/image.png", "--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "bgkill") {
		t.Fatalf("Execute() error = %v, want a bgkill failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}
