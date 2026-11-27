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

func forgeClientAt(serverURL string) *forge.Client {
	return forge.New(serverURL, zap.NewNop())
}

func TestForgeCommandSendsTheGenerateRequest(t *testing.T) {
	png := testPNG(t)
	body := fmt.Sprintf(`{"images":[%q]}`, base64.StdEncoding.EncodeToString(png))

	server, requests := captureServer(t, http.StatusOK, body)
	dir := t.TempDir()

	out, _, err := runCommand(t, cli.Deps{Config: testConfig(), Forge: forgeClientAt(server.URL)},
		"gen", "image", "a", "cat",
		"--model", "model",
		"--forge-preset", "preset",
		"--vae-and-text-model", "vae",
		"--vae-and-text-model", "text",
		"--negative-prompt", "ugly",
		"--sampler-name", "Euler",
		"--scheduler", "Karras",
		"--steps", "12",
		"--width", "640",
		"--height", "384",
		"--cfg-scale", "6.5",
		"--seed", "7",
		"--enable-hr",
		"--hr-scale", "2",
		"--hr-upscaler", "Latent",
		"--hr-second-pass-steps", "8",
		"--hr-cfg", "5.5",
		"--out", dir,
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)
	if request.Path != "/sdapi/v1/txt2img" {
		t.Errorf("path = %s, want /sdapi/v1/txt2img", request.Path)
	}

	payload := request.jsonBody(t)

	checks := map[string]any{
		"prompt":               "a cat",
		"negative_prompt":      "ugly",
		"sampler_name":         "Euler",
		"scheduler":            "Karras",
		"steps":                float64(12),
		"width":                float64(640),
		"height":               float64(384),
		"cfg_scale":            6.5,
		"seed":                 float64(7),
		"enable_hr":            true,
		"hr_scale":             float64(2),
		"hr_upscaler":          "Latent",
		"hr_second_pass_steps": float64(8),
		"hr_cfg":               5.5,
	}
	for field, want := range checks {
		assertField(t, payload, field, want)
	}

	overrides, ok := payload["override_settings"].(map[string]any)
	if !ok {
		t.Fatalf("override_settings = %v, want an object", payload["override_settings"])
	}

	assertField(t, overrides, "sd_model_checkpoint", "model.safetensors")
	assertField(t, overrides, "forge_preset", "preset")

	modules, ok := overrides["forge_additional_modules"].([]any)
	if !ok || len(modules) != 2 || modules[0] != "vae.safetensors" ||
		modules[1] != "text.safetensors" {
		t.Errorf("forge_additional_modules = %v, want the two names with extensions", modules)
	}

	files := writtenFiles(t, dir)
	if len(files) != 1 || !strings.HasSuffix(files[0], ".png") {
		t.Fatalf("files = %v, want one png", files)
	}

	if string(readFile(t, files[0])) != string(png) {
		t.Errorf("%s content does not match the generated image", files[0])
	}

	if !strings.Contains(out, files[0]) {
		t.Errorf("stdout does not contain the saved path:\n%s", out)
	}
}

func TestForgeCommandUsesTheInitImage(t *testing.T) {
	png := testPNG(t)
	input := imageServer(t, png)

	server, requests := captureServer(t, http.StatusOK,
		fmt.Sprintf(`{"images":[%q]}`, base64.StdEncoding.EncodeToString(png)))

	_, _, err := runCommand(t, cli.Deps{Config: testConfig(), Forge: forgeClientAt(server.URL)},
		"gen", "image", "a cat",
		"--model", "model",
		"--forge-preset", "preset",
		"--init-image-url", input.URL+"/init.png",
		"--out", t.TempDir(),
	)
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)
	if request.Path != "/sdapi/v1/img2img" {
		t.Errorf("path = %s, want /sdapi/v1/img2img", request.Path)
	}

	payload := request.jsonBody(t)

	images, ok := payload["init_images"].([]any)
	if !ok || len(images) != 1 {
		t.Fatalf("init_images = %v, want the init image base64 encoded", payload["init_images"])
	}

	if images[0] != base64.StdEncoding.EncodeToString(png) {
		t.Error("init_images does not carry the fetched image")
	}

	assertField(t, payload, "denoising_strength", 0.75)
	assertField(t, payload, "width", float64(8))
	assertField(t, payload, "height", float64(8))
}

func TestForgeCommandRejectsInvalidInput(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK, `{"images":[]}`)
	deps := cli.Deps{Config: testConfig(), Forge: forgeClientAt(server.URL)}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing model", []string{"gen", "image", "a cat", "--forge-preset", "preset"}, "--model is required"},
		{"missing preset", []string{"gen", "image", "a cat", "--model", "model"}, "--forge-preset is required"},
		{
			"empty prompt",
			[]string{"gen", "image", "   ", "--model", "model", "--forge-preset", "preset"},
			"prompt is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := runCommand(t, deps, tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Execute() error = %v, want %q", err, tt.want)
			}

			assertNoRequest(t, requests)
		})
	}
}

func TestForgeCommandSurfacesAnUpstreamError(t *testing.T) {
	server, _ := captureServer(t, http.StatusInternalServerError, "boom")
	dir := t.TempDir()

	_, _, err := runCommand(t, cli.Deps{Config: testConfig(), Forge: forgeClientAt(server.URL)},
		"gen", "image", "a cat",
		"--model", "model",
		"--forge-preset", "preset",
		"--out", dir,
	)
	if err == nil || !strings.Contains(err.Error(), "forge") {
		t.Fatalf("Execute() error = %v, want a forge failure", err)
	}

	if files := writtenFiles(t, dir); len(files) != 0 {
		t.Errorf("files = %v, want none", files)
	}
}
