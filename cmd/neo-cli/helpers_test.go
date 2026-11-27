package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wishmatic/neo-mcp/internal/cli"
	"github.com/wishmatic/neo-mcp/internal/config"
	"github.com/wishmatic/neo-mcp/internal/format"
)

func runCommand(t *testing.T, deps cli.Deps, args ...string) (string, string, error) {
	t.Helper()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	deps.Out = out
	deps.Err = errOut

	command := newRoot(deps)
	command.SetArgs(args)

	err := command.Execute()

	return out.String(), errOut.String(), err
}

func testConfig() config.Config {
	return config.Config{DefaultOutput: "png"}
}

func testPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))

	for y := 2; y < 6; y++ {
		for x := 2; x < 6; x++ {
			img.Set(x, y, color.NRGBA{R: 0xff, A: 0xff})
		}
	}

	data, err := format.Encode(img, format.PNG)
	if err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	return data
}

func imageServer(t *testing.T, data []byte) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)

	return server
}

func writtenFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}

	return paths
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error: %v", path, err)
	}

	return data
}

type capturedRequest struct {
	Method string
	Path   string
	Header http.Header
	Body   []byte
}

func captureServer(
	t *testing.T, status int, response string,
) (*httptest.Server, <-chan capturedRequest) {
	t.Helper()

	return captureServerFunc(t, func(*http.Request) (int, string) {
		return status, response
	})
}

func captureServerFunc(
	t *testing.T, respond func(*http.Request) (int, string),
) (*httptest.Server, <-chan capturedRequest) {
	t.Helper()

	requests := make(chan capturedRequest, 16)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			http.Error(w, "read body", http.StatusInternalServerError)

			return
		}

		requests <- capturedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Header: r.Header.Clone(),
			Body:   body,
		}

		status, response := respond(r)

		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	return server, requests
}

func nextRequest(t *testing.T, requests <-chan capturedRequest) capturedRequest {
	t.Helper()

	select {
	case request := <-requests:
		return request
	default:
		t.Fatal("upstream received no request")

		return capturedRequest{}
	}
}

func nextRequestPath(t *testing.T, requests <-chan capturedRequest, path string) capturedRequest {
	t.Helper()

	for range 16 {
		request := nextRequest(t, requests)
		if request.Path == path {
			return request
		}
	}

	t.Fatalf("upstream received no request for %s", path)

	return capturedRequest{}
}

func assertNoRequest(t *testing.T, requests <-chan capturedRequest) {
	t.Helper()

	select {
	case request := <-requests:
		t.Fatalf("upstream received %s %s, want no request", request.Method, request.Path)
	default:
	}
}

func (r capturedRequest) jsonBody(t *testing.T) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(r.Body, &body); err != nil {
		t.Fatalf("decode request body %q: %v", r.Body, err)
	}

	return body
}

func assertField(t *testing.T, body map[string]any, name string, want any) {
	t.Helper()

	if body[name] != want {
		t.Errorf("%s = %v, want %v", name, body[name], want)
	}
}
