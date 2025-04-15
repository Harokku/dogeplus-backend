package errors

import (
	"fmt"
	"testing"
)

// TestWrap tests the Wrap function
func TestWrap(t *testing.T) {
	// Create a simple error
	originalErr := fmt.Errorf("original error")

	// Wrap the error
	wrappedErr := Wrap(originalErr, "context message")

	// Check that the wrapped error contains both the context and the original error
	expected := "context message: original error"
	if wrappedErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, wrappedErr.Error())
	}
}

// TestWrapWithMessage tests the WrapWithMessage function
func TestWrapWithMessage(t *testing.T) {
	// Create a simple error
	originalErr := fmt.Errorf("original error")

	// Wrap the error with a message
	wrappedErr := WrapWithMessage(originalErr, "context message")

	// Check that the wrapped error contains both the context and the original error
	expected := "context message: original error"
	if wrappedErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, wrappedErr.Error())
	}
}

// TestWrapNil tests that Wrap returns nil when given a nil error
func TestWrapNil(t *testing.T) {
	// Wrap a nil error
	wrappedErr := Wrap(nil, "context message")

	// Check that the result is nil
	if wrappedErr != nil {
		t.Errorf("Expected nil, got '%v'", wrappedErr)
	}
}

// TestHandleCloser tests the HandleCloser function
// This is a simple test that just ensures the function doesn't panic
func TestHandleCloser(t *testing.T) {
	// Create a simple error
	err := fmt.Errorf("close error")

	// Call HandleCloser
	HandleCloser(err, "error closing resource")

	// No assertion needed, we're just making sure it doesn't panic
}

// ExampleWrap demonstrates how to use the Wrap function
func ExampleWrap() {
	// Simulate a database operation that fails
	err := fmt.Errorf("connection refused")

	// Wrap the error with context
	wrappedErr := Wrap(err, "failed to query database")

	// Print the error
	fmt.Println(wrappedErr)
	// Output: failed to query database: connection refused
}

// ExampleHandleCloser demonstrates how to use the HandleCloser function
func ExampleHandleCloser() {
	// Simulate a Close operation that returns an error
	closeErr := fmt.Errorf("resource already closed")

	// Use HandleCloser in a defer statement
	defer func() {
		HandleCloser(closeErr, "error closing resource")
	}()

	// Function body would go here
	fmt.Println("Function body")
	// Output: Function body
}
