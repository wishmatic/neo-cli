package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/wishmatic/neo-mcp/internal/resolve"
)

// ReadImage reads an image argument: the file at that path when one exists, and otherwise whatever
// the resolver accepts, such as an http(s) URL, a data URI, or raw base64.
func ReadImage(ctx context.Context, resolver *resolve.Client, address string) ([]byte, error) {
	data, isFile, err := readLocalImage(address)
	if isFile || err != nil {
		return data, err
	}

	image, err := resolver.Resolve(ctx, address)
	if err != nil {
		return nil, err
	}

	return image.Data, nil
}

func readLocalImage(path string) ([]byte, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false, nil
	}

	if info.IsDir() {
		return nil, true, fmt.Errorf("%q is a directory, not an image file", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, true, fmt.Errorf("read %s: %w", path, err)
	}

	return data, true, nil
}
