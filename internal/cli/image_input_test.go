package cli

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wishmatic/neo-mcp/internal/resolve"
)

var testImageBytes = []byte("\x89PNG\r\n\x1a\n test image")

func newTestResolver(t *testing.T) *resolve.Client {
	t.Helper()

	resolver, err := resolve.New(nil, "", nil)
	if err != nil {
		t.Fatalf("resolve.New() error: %v", err)
	}

	return resolver
}

func TestReadImageReadsALocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.png")
	if err := os.WriteFile(path, testImageBytes, 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	data, err := ReadImage(context.Background(), newTestResolver(t), path)
	if err != nil {
		t.Fatalf("ReadImage() error: %v", err)
	}

	if string(data) != string(testImageBytes) {
		t.Error("ReadImage() did not return the file's contents")
	}
}

func TestReadImageRejectsADirectory(t *testing.T) {
	_, err := ReadImage(context.Background(), newTestResolver(t), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Errorf("ReadImage() error = %v, want a directory rejection", err)
	}
}

func TestReadImageResolvesAURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(testImageBytes)
	}))
	t.Cleanup(server.Close)

	data, err := ReadImage(context.Background(), newTestResolver(t), server.URL+"/image.png")
	if err != nil {
		t.Fatalf("ReadImage() error: %v", err)
	}

	if string(data) != string(testImageBytes) {
		t.Error("ReadImage() did not return the served image")
	}
}

func TestReadImageResolvesRawBase64(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(testImageBytes)

	data, err := ReadImage(context.Background(), newTestResolver(t), encoded)
	if err != nil {
		t.Fatalf("ReadImage() error: %v", err)
	}

	if string(data) != string(testImageBytes) {
		t.Error("ReadImage() did not return the decoded image")
	}
}
