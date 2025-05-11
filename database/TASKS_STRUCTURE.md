# Tasks Module Structure

This document provides a visual representation of the code structure in `tasks.go` to help developers understand the relationships between different components.

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          tasks.go                               │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Data Structures                           │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────┐  ┌───────────────────┐  ┌─────────────────────────┐ │
│ │  Task   │  │ escalationLevels  │  │    incidentLevels       │ │
│ └─────────┘  └───────────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Repository Pattern                          │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────────┐       ┌───────────────────────────────────┐ │
│ │ TaskRepository  │ ────▶ │    TaskRepositoryTransaction     │ │
│ └─────────────────┘       └───────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Database Operations                         │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐  ┌────────────────────┐  ┌──────────────────┐  │
│ │ GetCategories│  │ GetByCategories    │  │ BulkAdd          │  │
│ └─────────────┘  └────────────────────┘  └──────────────────┘  │
│                                                                 │
│ ┌───────────────────────────────────┐  ┌──────────────────┐    │
│ │ GetByCategoriesAndEscalationLevels│  │ ClearTasksTable  │    │
│ └───────────────────────────────────┘  └──────────────────┘    │
│                                                                 │
│ ┌───────────────────────────────┐  ┌──────────────────────┐    │
│ │ GetGyCategoryAndEscalationLevel│  │ executeAndScanResults│    │
│ └───────────────────────────────┘  └──────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Excel Parsing                             │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────────┐  ┌────────────┐  ┌──────────┐              │
│ │ ParseXLSXToTasks│  │ parsePriority│  │ padBlock │              │
│ └─────────────────┘  └────────────┘  └──────────┘              │
│                                                                 │
│                      ┌────────────┐                             │
│                      │ isBlockEmpty│                             │
│                      └────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Task Filtering & Merging                     │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐  ┌────────────────────────┐                     │
│ │ FilterTasks │  │ FilterTasksForEscalation│                     │
│ └─────────────┘  └────────────────────────┘                     │
│                                                                 │
│ ┌─────────────┐  ┌────────────────────────┐                     │
│ │ MergeTasks  │  │ MergeTasksFixCategory  │                     │
│ └─────────────┘  └────────────────────────┘                     │
└─────────────────────────────────────────────────────────────────┘
```

## Function Call Hierarchy

```
TaskRepository
├── BeginTrans
├── WithTransaction
│   └── TaskRepositoryTransaction
│       ├── DropTasksTable
│       │   └── ClearTasksTable
│       └── BulkAdd
│           └── BulkAdd
├── GetCategories
├── GetByCategories
│   └── executeAndScanResults
├── GetByCategoriesAndEscalationLevels
│   └── executeAndScanResults
└── GetGyCategoryAndEscalationLevel
    └── executeAndScanResults

ParseXLSXToTasks
├── parsePriority
├── isBlockEmpty
└── padBlock

FilterTasks

FilterTasksForEscalation

MergeTasks

MergeTasksFixCategory
```

## Data Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Excel File │────▶│    Tasks    │────▶│  Database   │
└─────────────┘     └─────────────┘     └─────────────┘
                         │  ▲
                         │  │
                         ▼  │
                    ┌─────────────┐
                    │   Filters   │
                    └─────────────┘
```

## Logical Groupings for Future Refactoring

If the file were to be refactored into smaller files, here's how the components could be grouped:

### task_model.go
- Task struct
- Constants (PRO22)
- escalationLevels map
- incidentLevels map
- GetEscalationLevels function

### task_repository.go
- TaskRepository struct
- TaskRepositoryTransaction struct
- Database operations (GetCategories, GetByCategories, etc.)
- Transaction handling (BeginTrans, WithTransaction, etc.)

### task_excel.go
- ParseXLSXToTasks function
- Helper functions (parsePriority, isBlockEmpty, padBlock)

### task_filter.go
- FilterTasks function
- FilterTasksForEscalation function
- MergeTasks function
- MergeTasksFixCategory function

This structure would improve code organization while maintaining the same functionality.