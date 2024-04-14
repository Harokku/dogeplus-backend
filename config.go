package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2/log"
	"os"
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
	Port = "PORT"
)

// EnvVarsSlice is a slice of the EnvVars type, representing a collection of environment variables.
var EnvVarsSlice = []EnvVars{
	Port,
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
