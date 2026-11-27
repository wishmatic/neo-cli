package cli

import (
	"fmt"
	"io"
	"strconv"
)

func FormatUSD(amount float64) string {
	return "$" + strconv.FormatFloat(amount, 'f', -1, 64)
}

func PrintPaths(w io.Writer, paths []string) error {
	for _, path := range paths {
		if _, err := fmt.Fprintln(w, path); err != nil {
			return fmt.Errorf("write path: %w", err)
		}
	}

	return nil
}
