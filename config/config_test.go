// Package config provides functionality for loading and managing application configuration.
// It handles reading configuration from both TOML files and environment variables,
// providing fallback mechanisms and validation. The package also includes utilities
// for file path sanitization and Excel file loading based on configuration settings.
package config

import (
	"log"
	"os"
	"testing"
)

func TestSanitizeFilePath(t *testing.T) {
	// Create a temporary file for testing purposes
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatalf("Failed to create temporary file for testing: %s", err)
	}
	defer os.Remove(tempFile.Name())

	testCases := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "Absolute Path",
			path:      tempFile.Name(),
			expectErr: false,
		},
		{
			name:      "Relative Path",
			path:      "relative/path",
			expectErr: true,
		},
		{
			name:      "Path With Double Dot",
			path:      "../path",
			expectErr: true,
		},
		{
			name:      "Nonexistent File",
			path:      "/nonexistent/file",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SanitizeFilePath(tc.path)
			if (err != nil) != tc.expectErr {
				t.Fatalf("SanitizeFilePath() error = %v, expectErr = %v", err, tc.expectErr)
			}
		})
	}
}
