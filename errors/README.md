# Error Handling Package

This package provides standardized error handling and logging functions for the DogePlus Backend. It offers consistent ways to wrap, log, and handle errors throughout the application.

## Overview

The error handling package provides several key functions:

1. **Error wrapping** - Add context to errors while preserving the original error
2. **Error logging** - Log errors with appropriate context and severity
3. **Special-purpose error handling** - Handle specific error cases like `Close()` errors
4. **Logging convenience functions** - Standardized logging for different severity levels

## Usage Examples

### Basic Error Wrapping

```go
import "dogeplus-backend/errors"

func DoSomething() error {
    result, err := someOperation()
    if err != nil {
        return errors.Wrap(err, "failed to perform operation")
    }
    return nil
}
```

### Error Logging

```go
func ProcessData() error {
    data, err := fetchData()
    if err != nil {
        // Log the error and return it
        return errors.LogError(err, "failed to fetch data")
    }
    return nil
}
```

### Handling Close Errors

```go
func QueryDatabase() ([]Record, error) {
    rows, err := db.Query("SELECT * FROM records")
    if err != nil {
        return nil, errors.Wrap(err, "failed to query records")
    }
    defer func() {
        errors.HandleCloser(rows.Close(), "error closing rows")
    }()
    
    // Process rows...
    
    return records, nil
}
```

### Combined Logging and Wrapping

```go
func UpdateRecord(id string, data Record) error {
    _, err := db.Exec("UPDATE records SET value = ? WHERE id = ?", data.Value, id)
    if err != nil {
        return errors.LogAndWrapError(err, "failed to update record %s", id)
    }
    return nil
}
```

## Function Reference

### Error Wrapping

- `Wrap(err error, format string, args ...interface{}) error` - Wraps an error with a formatted message
- `WrapWithMessage(err error, message string) error` - Wraps an error with a simple message

### Error Logging

- `LogError(err error, format string, args ...interface{}) error` - Logs an error and returns it
- `LogAndWrapError(err error, format string, args ...interface{}) error` - Logs and wraps an error

### Special-Purpose Error Handling

- `HandleCloser(err error, message string)` - Handles errors from Close() methods
- `HandleCloserWithContext(err error, format string, args ...interface{})` - Handles Close() errors with context

### Logging Convenience Functions

- `Fatal(err error, format string, args ...interface{})` - Logs a fatal error and panics
- `Info(format string, args ...interface{})` - Logs an informational message
- `Warning(format string, args ...interface{})` - Logs a warning message
- `Debug(format string, args ...interface{})` - Logs a debug message

## Best Practices

1. **Be descriptive** - Error messages should clearly indicate what operation failed
2. **Include context** - Include relevant parameters in error messages (e.g., IDs, names)
3. **Use appropriate severity** - Use the right logging level for different types of errors
4. **Handle all errors** - Don't ignore errors, especially from Close() operations
5. **Wrap errors at package boundaries** - Add context when errors cross package boundaries