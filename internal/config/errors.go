package config

import "fmt"

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

// NewConfigError creates a new configuration error
func NewConfigError(field, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error for %s: %s", e.Field, e.Message)
}
