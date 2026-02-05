// Package config provides centralized environment variable management.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetEnv retrieves an environment variable or returns a default value if not defined.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

// GetEnvRequired retrieves a required environment variable.
// Returns an error if the variable is not defined.
func GetEnvRequired(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("required environment variable %s is not defined", key)
	}

	return value, nil
}

// GetEnvRequiredNonEmpty retrieves a required non-empty environment variable.
// Returns an error if the variable is not defined or is empty.
func GetEnvRequiredNonEmpty(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("required environment variable %s is not defined", key)
	}

	if value == "" {
		return "", fmt.Errorf("required environment variable %s is defined but empty", key)
	}

	return value, nil
}

// Exists checks if an environment variable is defined (even if empty).
func Exists(key string) bool {
	_, exists := os.LookupEnv(key)

	return exists
}

// GetEnvInt retrieves an environment variable as an integer.
// Returns defaultValue if not defined or if parsing fails.
func GetEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intVal
}

// GetEnvIntRequired retrieves a required environment variable as an integer.
// Returns an error if not defined or if parsing fails.
func GetEnvIntRequired(key string) (int, error) {
	value, err := GetEnvRequiredNonEmpty(key)
	if err != nil {
		return 0, err
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s must be a valid integer: %w", key, err)
	}

	return intVal, nil
}

// GetEnvBool retrieves an environment variable as a boolean.
// Returns defaultValue if not defined.
// Accepts: true, 1, yes (case insensitive) as true values.
func GetEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	lower := strings.ToLower(value)

	return lower == "true" || lower == "1" || lower == "yes"
}

// GetEnvDuration retrieves an environment variable as a time.Duration.
// Returns defaultValue if not defined or if parsing fails.
// Accepts formats like "30s", "5m", "1h".
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return duration
}

// GetEnvSlice retrieves an environment variable as a string slice.
// Values should be comma-separated. Returns defaultValue if not defined.
func GetEnvSlice(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// MustGetEnv retrieves a required environment variable.
// Panics if the variable is not defined (use only during initialization).
func MustGetEnv(key string) string {
	value, err := GetEnvRequired(key)
	if err != nil {
		panic(err)
	}

	return value
}

// MustGetEnvNonEmpty retrieves a required non-empty environment variable.
// Panics if the variable is not defined or empty (use only during initialization).
func MustGetEnvNonEmpty(key string) string {
	value, err := GetEnvRequiredNonEmpty(key)
	if err != nil {
		panic(err)
	}

	return value
}

// MustGetEnvInt retrieves a required environment variable as an integer.
// Panics if not defined or invalid (use only during initialization).
func MustGetEnvInt(key string) int {
	value, err := GetEnvIntRequired(key)
	if err != nil {
		panic(err)
	}

	return value
}
