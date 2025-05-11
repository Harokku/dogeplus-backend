# Tasks Module Documentation

This document provides an overview of the tasks module in the DogePlus Backend, which is implemented in `tasks.go`. The file has grown large and complex, so this documentation aims to help developers understand its structure and functionality.

## Table of Contents
1. [Overview](#overview)
2. [Data Structures](#data-structures)
3. [Repository Pattern](#repository-pattern)
4. [Database Operations](#database-operations)
5. [Excel Parsing](#excel-parsing)
6. [Task Filtering](#task-filtering)
7. [Task Merging](#task-merging)
8. [Best Practices for Future Development](#best-practices-for-future-development)

## Overview

The tasks module provides functionality for managing tasks in the DogePlus Backend. It includes:
- Task data structure definition
- Repository pattern implementation for database operations
- Excel file parsing for importing tasks
- Task filtering and merging utilities

## Data Structures

### Task
The core data structure is the `Task` struct, which represents a task with its properties:
```go
type Task struct {
    ID              int    `json:"ID,omitempty"`
    Priority        int    `json:"priority,omitempty"`
    Title           string `json:"title,omitempty"`
    Description     string `json:"description,omitempty"`
    Role            string `json:"role,omitempty"`
    Category        string `json:"category,omitempty"`
    EscalationLevel string `json:"escalation_level,omitempty"`
    IncidentLevel   string `json:"incident_level,omitempty"`
}
```

### Constants and Maps
The module defines several constants and maps for working with tasks:
- `PRO22`: A constant representing a special category
- `escalationLevels`: A map of escalation levels to their priority values
- `incidentLevels`: A map of incident levels to their priority values

## Repository Pattern

The module implements the repository pattern for database operations:

### TaskRepository
```go
type TaskRepository struct {
    db *sql.DB
}
```

### TaskRepositoryTransaction
```go
type TaskRepositoryTransaction struct {
    repo *TaskRepository
    *sql.Tx
}
```

## Database Operations

The module provides several methods for database operations:

### Query Operations
- `GetCategories`: Retrieves distinct categories from the tasks table
- `GetByCategories`: Retrieves tasks based on the provided category
- `GetByCategoriesAndEscalationLevels`: Retrieves tasks based on categories and escalation levels
- `GetGyCategoryAndEscalationLevel`: Retrieves tasks based on category, escalation levels, and incident level

### Transaction Operations
- `BeginTrans`: Starts a new transaction
- `WithTransaction`: Runs queries wrapped in a transaction
- `BulkAdd`: Adds a slice of tasks to the database
- `ClearTasksTable`: Clears the tasks table

## Excel Parsing

The module provides functionality for parsing Excel files into Task objects:

### Helper Functions
- `parsePriority`: Converts a priority string to an int
- `isBlockEmpty`: Checks if a block of columns is empty
- `padBlock`: Ensures a block has exactly the specified number of columns

### Main Parsing Function
- `ParseXLSXToTasks`: Converts an Excel file into a slice of Task instances

## Task Filtering

The module provides functions for filtering tasks based on various criteria:

- `FilterTasks`: Filters and sorts tasks based on category, escalation, and incident levels
- `FilterTasksForEscalation`: Filters tasks based on category, escalation levels, and incident level

## Task Merging

The module provides functions for merging tasks:

- `MergeTasks`: Merges two slices of Tasks by updating or removing existing tasks and adding new ones
- `MergeTasksFixCategory`: Merges two slices of Tasks based on Title and Category

## Best Practices for Future Development

To maintain and improve the tasks module, consider the following best practices:

1. **Separation of Concerns**: Consider refactoring the file into smaller, more focused files:
   - `task_model.go`: Task struct and related constants
   - `task_repository.go`: Repository pattern implementation
   - `task_excel.go`: Excel parsing functionality
   - `task_filter.go`: Task filtering and merging utilities

2. **Improved Documentation**: Add more detailed documentation to functions and methods

3. **Error Handling**: Ensure consistent error handling throughout the module

4. **Testing**: Add more unit tests to cover edge cases

5. **Code Organization**: Group related functions together and add section comments

By following these best practices, the tasks module can be maintained and extended more easily in the future.