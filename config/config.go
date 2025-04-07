// Package config provides functionality for loading and managing application configuration.
// It handles reading configuration from both TOML files and environment variables,
// providing fallback mechanisms and validation. The package also includes utilities
// for file path sanitization and Excel file loading based on configuration settings.
package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"path/filepath"
	"strings"
)

// Config is a struct that defines the configuration for the application.
// It contains a map of string key-value pairs, with the key being the config variable name
// and the value being the corresponding value for that variable.
// This struct is used to store the configuration values that are read from a configuration file.
type Config struct {
	Variable map[string]interface{} `toml:"variables"`
}

type EnvVars string

// Port is a constant representing the environment variable name .
const (
	Port     = "PORT"
	DbFile   = "DBFILE"
	TaskRoot = "TASKROOT"
)

// EnvVarsSlice is a slice of the EnvVars type, representing a collection of environment variables.
var EnvVarsSlice = []EnvVars{
	Port,
	DbFile,
	TaskRoot,
}

// checkAllEnvVars checks if all environment variables specified in EnvVarsSlice are set.
// It takes a Config instance as input and uses GetEnvWithFallback to retrieve the value for each
// environment variable. If any of the variables are not set, an error is returned.
// The returned error contains a formatted string indicating which environment variable is not set.
// See EnvVarsSlice and GetEnvWithFallback for more information.
func checkAllEnvVars(config Config) []error {
	var errs []error
	for _, key := range EnvVarsSlice {
		value := GetEnvWithFallback(config, key)
		if value == "" {
			errs = append(errs, fmt.Errorf("environment variable %s is empty", key))
		}
	}
	return errs
}

// LoadConfig loads the configuration from the "config.toml" file.
// It decodes the file and returns a Config struct.
// If there is an error loading the config.toml file, a warning is logged.
// It checks if all variables are set and halt with fatal error if not
func LoadConfig() Config {
	var config Config

	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Warn("Error loading config.toml: ", err)
	}

	errs := checkAllEnvVars(config)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Warn(err)
		}
		log.Fatal("Not all environment variables are set.")
	}

	return config
}

// GetEnvWithFallback retrieves the value of a config variable from the provided Config instance.
// If the value is not found in the Config, it falls back to the corresponding environment variable.
// The function takes a Config instance and an EnvVars key as input.
// The Config instance is used to look for the value in the config file, and the EnvVars key is used
// to retrieve the value from the environment variable if it is not found in the Config.
// If the value is not found in either the Config or the environment variable, an empty string is returned.
// See Config and EnvVars for more information.
func GetEnvWithFallback(config Config, key EnvVars) string {
	// Check if value has been parsed from config file
	value, exist := config.Variable[string(key)]
	if !exist {
		// Fallback to environment variable
		value = os.Getenv(string(key))
	}

	return fmt.Sprintf("%v", value)
}

// SanitizeFilePath checks if the given path is an absolute path and does not contain any parent directory traversals.
// It returns an error if the path is not absolute or contains "..".
// Additionally, it checks if the file at the given path exists. If the file does not exist, an error is returned.
// The error message contains the path that failed the check.
// This function relies on the filepath.IsAbs and os.Stat functions.
func SanitizeFilePath(path string) error {
	// Clean the path to normalize it
	cleanPath := filepath.Clean(path)

	// check if path is absolute to prevent path traversal and ensure it doesn't traverse root
	if !filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf(`"%s" is not an absolute path or contains traversal`, path)
	}

	// Detect and reject any null bytes in the file path
	if strings.ContainsRune(cleanPath, '\x00') {
		return fmt.Errorf(`"%s" contains null byte`, path)
	}

	// check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return fmt.Errorf(`"%s" does not exist`, path)
	}

	// Check if the program has sufficient permission to access the file
	file, err := os.Open(cleanPath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf(`"%s" cannot be opened: %v`, path, err)
	}

	return nil
}
