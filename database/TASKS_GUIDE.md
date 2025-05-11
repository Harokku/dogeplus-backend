# Tasks Module Usage Guide

This guide provides practical examples and best practices for working with the tasks module in the DogePlus Backend. It complements the `TASKS_README.md` file, which provides an overview of the module's structure.

## Table of Contents
1. [Common Operations](#common-operations)
2. [Working with the Repository](#working-with-the-repository)
3. [Excel File Processing](#excel-file-processing)
4. [Task Filtering and Merging](#task-filtering-and-merging)
5. [Error Handling](#error-handling)
6. [Testing](#testing)
7. [Future Development](#future-development)

## Common Operations

### Retrieving Tasks by Category

```go
// Get tasks for a specific category
tasks, err := taskRepo.GetByCategories("SomeCategory")
if err != nil {
    // Handle error
}

// Process tasks
for _, task := range tasks {
    // Do something with each task
}
```

### Retrieving Tasks by Category and Escalation Level

```go
// Get tasks for a specific category and escalation level
tasks, err := taskRepo.GetGyCategoryAndEscalationLevel("SomeCategory", "allarme", "emergenza", "")
if err != nil {
    // Handle error
}

// Process tasks
for _, task := range tasks {
    // Do something with each task
}
```

### Adding Tasks to the Database

```go
// Create some tasks
tasks := []database.Task{
    {
        Category:        "SomeCategory",
        Role:            "SomeRole",
        Priority:        1,
        Title:           "Task Title",
        Description:     "Task Description",
        EscalationLevel: "allarme",
    },
    // More tasks...
}

// Add tasks to the database using a transaction
err := taskRepo.WithTransaction(func(tx *database.TaskRepositoryTransaction) error {
    return tx.BulkAdd(tasks)
})
if err != nil {
    // Handle error
}
```

## Working with the Repository

### Initializing the Repository

The `TaskRepository` is typically initialized as part of the `Repositories` struct in `sqLiteConnect.go`:

```go
// This is done in the NewRepositories function
repos := &Repositories{
    Tasks: NewTaskRepository(db),
    // Other repositories...
}
```

### Using Transactions

Transactions are important for ensuring data consistency when performing multiple operations:

```go
// Example of using a transaction to clear the tasks table and add new tasks
err := taskRepo.WithTransaction(func(tx *database.TaskRepositoryTransaction) error {
    // Clear the tasks table
    if err := tx.DropTasksTable(); err != nil {
        return err
    }
    
    // Add new tasks
    if err := tx.BulkAdd(tasks); err != nil {
        return err
    }
    
    return nil
})
if err != nil {
    // Handle error
}
```

## Excel File Processing

### Parsing an Excel File

```go
// Open the Excel file
file, err := excelize.OpenFile("path/to/file.xlsx")
if err != nil {
    // Handle error
}
defer file.Close()

// Parse the Excel file into tasks
tasks, err := database.ParseXLSXToTasks(file)
if err != nil {
    // Handle error
}

// Process the tasks
for _, task := range tasks {
    // Do something with each task
}
```

### Excel File Structure Requirements

The Excel file should follow this structure:
- Each sheet represents a category
- The first row contains role names
- The second row is skipped
- Starting from the third row, data is organized in blocks of 5 columns:
  1. Priority (integer)
  2. Title (string)
  3. Description (string)
  4. Escalation Level (string: "allarme", "emergenza", or "incidente")
  5. Incident Level (string: "bianca", "verde", "gialla", or "rossa")

## Task Filtering and Merging

### Filtering Tasks

```go
// Filter tasks based on category, escalation level, and incident level
filteredTasks := database.FilterTasks(tasks, "SomeCategory", "incidente", "verde")

// Process the filtered tasks
for _, task := range filteredTasks {
    // Do something with each filtered task
}
```

### Merging Tasks

```go
// Merge two sets of tasks
originalTasks := []database.Task{/* ... */}
updateTasks := []database.Task{/* ... */}

mergedTasks, err := database.MergeTasks(originalTasks, updateTasks)
if err != nil {
    // Handle error
}

// Process the merged tasks
for _, task := range mergedTasks {
    // Do something with each merged task
}
```

## Error Handling

The tasks module uses standard Go error handling patterns. Always check for errors when calling functions that return them:

```go
tasks, err := taskRepo.GetCategories(configFile)
if err != nil {
    // Log the error
    log.Printf("Failed to get categories: %v", err)
    
    // Return an appropriate error to the caller
    return nil, fmt.Errorf("failed to get categories: %w", err)
}
```

## Testing

The tasks module has tests in `tasks_test.go`. When making changes to the module, ensure that all tests pass:

```bash
go test -v ./database -run "^TestMergeTasks|TestParseXLSXToTasks|TestFilterTasksForEscalation|TestFilterTasks$"
```

When adding new functionality, also add corresponding tests.

## Future Development

When working on the tasks module in the future, consider the following:

1. **Maintain Backward Compatibility**: Ensure that changes don't break existing code that uses the module.

2. **Follow the Existing Patterns**: Use the same patterns and conventions that are already in place.

3. **Consider Refactoring**: If making significant changes, consider refactoring the module into smaller, more focused files as suggested in `TASKS_README.md`.

4. **Update Documentation**: Keep this guide and the README up to date with any changes.

By following these guidelines, you can help maintain and improve the tasks module while minimizing the risk of introducing bugs or breaking existing functionality.