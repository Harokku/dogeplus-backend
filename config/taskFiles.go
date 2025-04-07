// Package config provides functionality for loading and managing application configuration.
// It handles reading configuration from both TOML files and environment variables,
// providing fallback mechanisms and validation. The package also includes utilities
// for file path sanitization and Excel file loading based on configuration settings.
package config

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"path/filepath"
)

// LoadExcelFile loads an Excel file based on a given config and filename.
// It validates the task root path, builds the full file path, and sanitizes it before opening the file.
func LoadExcelFile(config Config, filename string) (*excelize.File, error) {
	taskRoot := GetEnvWithFallback(config, TaskRoot)
	if taskRoot == "" {
		return nil, fmt.Errorf("TASKROOT is not set in config.toml or environment variables")
	}

	// Default extension handling
	defaultExt := ".xlsx"
	if filepath.Ext(filename) == "" {
		filename += defaultExt
	}

	filePath := filepath.Join(taskRoot, filename)
	if err := SanitizeFilePath(filePath); err != nil {
		return nil, fmt.Errorf("file path validation failed: %v", err)
	}

	fmt.Printf("Loading file: %s\n", filePath)

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file at %s: %v", filePath, err)
	}

	return f, nil
}
