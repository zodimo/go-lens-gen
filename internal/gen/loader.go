package gen

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// NewLocalReferenceLoader creates a custom loader that intercepts requests to
// https://example.com/ and resolves them locally relative to baseDir.
func NewLocalReferenceLoader(baseDir string) func(s string) (io.ReadCloser, error) {
	return func(s string) (io.ReadCloser, error) {
		u, err := url.Parse(s)
		if err != nil {
			return jsonschema.LoadURL(s) // fallback
		}

		// Intercept example.com
		if u.Scheme == "https" && u.Host == "example.com" {
			// Extract filename from the path
			filename := filepath.Base(u.Path)
			localPath := filepath.Join(baseDir, filename)
			
			file, err := os.Open(localPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load referenced file '%s' from '%s': %w", s, localPath, err)
			}
			return file, nil
		}

		// Fallback to the default package loader
		return jsonschema.LoadURL(s)
	}
}
