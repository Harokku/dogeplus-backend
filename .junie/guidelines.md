# Go Style Guide for DogePlus Backend

This style guide outlines the coding conventions, architectural patterns, and best practices for the DogePlus Backend project. Following these guidelines will ensure code consistency, maintainability, and adherence to Go's idiomatic practices.

## Table of Contents
1. [Code Organization](#code-organization)
2. [Naming Conventions](#naming-conventions)
3. [Architectural Patterns](#architectural-patterns)
4. [Error Handling](#error-handling)
5. [Comments and Documentation](#comments-and-documentation)
6. [Testing](#testing)
7. [Common Pitfalls to Avoid](#common-pitfalls-to-avoid)

## Code Organization

### Package Structure
- Organize code by domain functionality, not by technical role
- Package names should be short, concise, and lowercase (e.g., `handlers`, `database`, `router`)
- Avoid package names like `util`, `common`, or `misc` that don't convey clear purpose
- Keep the number of exported symbols in a package to a minimum

### File Organization
- Group related functionality in the same file
- Keep files to a reasonable size (generally under 500 lines)
- Place interfaces near where they are used, not in separate files
- Use consistent file naming: lowercase with underscores for multi-word names (e.g., `active_events.go`)

## Naming Conventions

### General
- Use MixedCaps (camelCase) for multi-word names; avoid underscores
- Use short, descriptive names for variables and functions
- Acronyms should be consistently cased (e.g., `HTTPServer` or `httpServer`, not `HttpServer`)

### Exported vs Unexported
- Exported (public) names start with uppercase (e.g., `CreateNewEvent`)
- Unexported (private) names start with lowercase (e.g., `createTables`)
- Provide meaningful comments for all exported names

### Interface Names
- One-method interfaces should be named by the method name plus 'er' (e.g., `Reader` for `Read()`)
- Multi-method interfaces should describe their purpose (e.g., `Repository`)

### Constants
- Use camelCase for constant names, not SCREAMING_SNAKE_CASE
- Group related constants in a const block

## Architectural Patterns

### Web Application Structure
- Follow clean architecture principles with clear separation of concerns
- Use the following layers:
  - **Handlers**: HTTP request handling and response formatting
  - **Services**: Business logic (if complex enough to warrant separation)
  - **Repositories**: Data access and persistence
  - **Models**: Data structures and domain objects

### Dependency Injection
- Pass dependencies explicitly to functions and structs that need them
- Avoid global state and singletons except for truly global resources (like database connections)
- Use functional middleware pattern for HTTP handlers:
  ```go
  func HandlerName(dependencies) func(ctx *fiber.Ctx) error {
      return func(ctx *fiber.Ctx) error {
          // Handler implementation
      }
  }
  ```

### Repository Pattern
- Use the repository pattern for data access
- Each domain entity should have its own repository
- Repositories should be interfaces to allow for mocking in tests
- Group repositories in a `Repositories` struct for convenience

## Error Handling

### Error Creation
- Use meaningful error messages that help diagnose issues
- For simple errors, use `errors.New()` or `fmt.Errorf()`
- For complex errors, consider custom error types that implement the `error` interface

### Error Propagation
- Return errors rather than using panic
- Wrap errors with context when crossing package boundaries using `fmt.Errorf("doing X: %w", err)`
- Don't ignore errors; either handle them or return them

### Error Logging
- Log errors at the appropriate level (debug, info, warn, error)
- Include relevant context in error logs
- Avoid logging the same error multiple times

## Comments and Documentation

### Package Comments
- Every package should have a package comment
- Package comments should describe the purpose of the package

### Function Comments
- All exported functions should have a comment
- Comments should start with the function name and describe what the function does
- Document parameters and return values for complex functions

### Code Comments
- Write comments for complex or non-obvious code
- Focus on why, not what (the code shows what, comments explain why)
- Keep comments up-to-date with code changes

## Testing

### Test Organization
- Place tests in the same package as the code being tested with a `_test.go` suffix
- Use table-driven tests for testing multiple cases
- Organize tests by function: `TestFunctionName` or `TestFunctionName_Scenario`

### Test Coverage
- Aim for high test coverage, especially for critical paths
- Test both happy paths and error cases
- Use mocks or stubs for external dependencies

### Test Readability
- Make test failures easy to diagnose
- Use descriptive test names
- Include expected vs. actual values in failure messages

## Common Pitfalls to Avoid

### Concurrency Issues
- Be careful with goroutines and shared state
- Use proper synchronization (mutexes, channels) when accessing shared resources
- Avoid goroutine leaks by ensuring all goroutines can terminate

### Memory Management
- Be mindful of memory usage, especially for long-lived applications
- Avoid unnecessary allocations in hot paths
- Release resources (files, connections) using defer statements

### API Design
- Design APIs for future extensibility
- Use versioning for APIs (e.g., `/api/v1/`)
- Be consistent with error responses and status codes

### Database Access
- Use prepared statements for repeated queries
- Implement retry logic for transient database errors
- Close database resources properly
- Use transactions for operations that must be atomic

### Configuration
- Use environment variables or configuration files for settings
- Validate configuration at startup
- Provide sensible defaults

---

This style guide is a living document and may be updated as the project evolves. When in doubt, follow the principle: "Clear is better than clever."