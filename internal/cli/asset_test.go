package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteImageNamesTheFileAfterItsMediaType(t *testing.T) {
	tests := map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/webp": ".webp",
		"image/jxl":  ".jxl",
	}

	for mediaType, ext := range tests {
		t.Run(mediaType, func(t *testing.T) {
			dir := t.TempDir()

			path, err := WriteImage(dir, []byte("image bytes"), mediaType)
			if err != nil {
				t.Fatalf("WriteImage() error: %v", err)
			}

			if filepath.Dir(path) != dir {
				t.Errorf("path = %q, want it under %s", path, dir)
			}
			if filepath.Ext(path) != ext {
				t.Errorf("extension = %q, want %q", filepath.Ext(path), ext)
			}
			if !strings.HasPrefix(filepath.Base(path), "neo-image-") {
				t.Errorf("name = %q, want the neo-image prefix", filepath.Base(path))
			}

			assertFileContent(t, path, "image bytes")
		})
	}
}

func TestWriteImageLeavesNoTemporaryFile(t *testing.T) {
	dir := t.TempDir()

	if _, err := WriteImage(dir, []byte("image bytes"), "image/png"); err != nil {
		t.Fatalf("WriteImage() error: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("directory has %d entries, want only the renamed file", len(entries))
	}

	if strings.HasSuffix(entries[0].Name(), ".tmp") {
		t.Errorf("file %q kept its temporary name", entries[0].Name())
	}
}

func TestWriteImageCreatesAMissingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "output")

	path, err := WriteImage(dir, []byte("image bytes"), "image/png")
	if err != nil {
		t.Fatalf("WriteImage() error: %v", err)
	}

	assertFileContent(t, path, "image bytes")
}

func TestWriteImageUsesMode0644(t *testing.T) {
	path, err := WriteImage(t.TempDir(), []byte("image bytes"), "image/png")
	if err != nil {
		t.Fatalf("WriteImage() error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Errorf("mode = %o, want 644", perm)
	}
}

func TestWriteImageRejectsAnUnknownMediaType(t *testing.T) {
	dir := t.TempDir()

	_, err := WriteImage(dir, []byte("image bytes"), "image/gif")
	if err == nil || !strings.Contains(err.Error(), "image/gif") {
		t.Fatalf("WriteImage() error = %v, want the media type named", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("directory has %d entries, want none", len(entries))
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error: %v", path, err)
	}

	if string(data) != want {
		t.Errorf("%s content = %q, want %q", path, data, want)
	}
}
