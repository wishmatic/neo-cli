package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wishmatic/neo-mcp/internal/format"
)

const imagePrefix = "neo-image"

func WriteImage(dir string, data []byte, mediaType string) (string, error) {
	ext, err := imageExtension(mediaType)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	stamp := time.Now().UTC().Format("20060102-150405")

	tmp, err := os.CreateTemp(dir, imagePrefix+"-"+stamp+"-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}

	name := tmp.Name()
	defer func() {
		_ = os.Remove(name)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()

		return "", fmt.Errorf("write output file: %w", err)
	}

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()

		return "", fmt.Errorf("set output permissions: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close output file: %w", err)
	}

	final := filepath.Join(dir, strings.TrimSuffix(filepath.Base(name), ".tmp")+"."+ext)

	if err := os.Rename(name, final); err != nil {
		return "", fmt.Errorf("store output file: %w", err)
	}

	name = ""

	return final, nil
}

func WriteImages(dir string, images [][]byte, target format.Format) ([]string, error) {
	paths := make([]string, 0, len(images))

	for index, data := range images {
		converted, err := format.Convert(data, target)
		if err != nil {
			return nil, fmt.Errorf("image %d: %w", index+1, err)
		}

		path, err := WriteImage(dir, converted, target.MediaType())
		if err != nil {
			return nil, err
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func imageExtension(mediaType string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mediaType))

	for _, candidate := range []format.Format{format.PNG, format.JPEG, format.JXL, format.WebP} {
		if candidate.MediaType() == normalized {
			return candidate.Extension(), nil
		}
	}

	return "", fmt.Errorf("cannot store media type %q", mediaType)
}
