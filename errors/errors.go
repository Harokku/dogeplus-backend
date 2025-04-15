// Package errors provides standardized error handling and logging functions for the DogePlus Backend.
// It offers consistent ways to wrap, log, and handle errors throughout the application.
package errors

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
)

// Wrap wraps an error with additional context using fmt.Errorf with %w verb.
// This preserves the original error type for unwrapping later.
// If err is nil, Wrap returns nil.
func Wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// WrapWithMessage wraps an error with a simple message.
// This is a convenience function for when you don't need format specifiers.
// If err is nil, WrapWithMessage returns nil.
func WrapWithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// LogError logs an error with the provided message and returns the original error.
// This is useful when you want to log an error but still return it up the call stack.
// If err is nil, LogError does nothing and returns nil.
func LogError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	log.Errorf(format+": %v", append(args, err)...)
	return err
}

// LogAndWrapError logs an error with the provided message and returns a wrapped error.
// This combines the functionality of LogError and Wrap.
// If err is nil, LogAndWrapError does nothing and returns nil.
func LogAndWrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	log.Errorf(format+": %v", append(args, err)...)
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// HandleCloser handles errors from Close() methods, typically used with defer statements.
// It logs the error but does not return it, making it suitable for defer statements.
// Example usage: defer HandleCloser(rows.Close(), "Error closing rows")
func HandleCloser(err error, message string) {
	if err != nil {
		log.Warnf("%s: %v", message, err)
	}
}

// HandleCloserWithContext handles errors from Close() methods with additional context.
// It logs the error with the provided format and arguments but does not return it.
// Example usage: defer HandleCloserWithContext(rows.Close(), "Error closing rows for query %s", queryName)
func HandleCloserWithContext(err error, format string, args ...interface{}) {
	if err != nil {
		log.Warnf(format+": %v", append(args, err)...)
	}
}

// Fatal logs a fatal error and panics.
// This should be used only for errors that should terminate the program.
func Fatal(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	log.Fatalf(format+": %v", append(args, err)...)
}

// Info logs an informational message.
// This is a convenience function for logging non-error information.
func Info(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warning logs a warning message.
// This is a convenience function for logging warnings.
func Warning(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Debug logs a debug message.
// This is a convenience function for logging debug information.
func Debug(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
