package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// Task represents a task with its properties.
type Task struct {
	ID          int    `json:"ID,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Role        string `json:"role,omitempty"`
	Category    string `json:"category,omitempty"`
}

// TaskRepository represents a repository for managing tasks.
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository initializes a new instance of TaskRepository.
// It takes a *sql.DB object as input and returns a pointer to TaskRepository.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetByCategories retrieves tasks based on the provided categories.
// It takes an array of category strings as input and returns an array of Task objects and an error.
// For each category, a placeholder is created and an argument is assigned.
// A query is constructed using the placeholders and executed against the database connection.
// The retrieved rows are scanned and transformed into Task objects,
// which are then appended to the tasks array.
// If any error occurs during the process, it is returned along with the tasks array.
// Finally, the tasks array and the error are returned.
func (t *TaskRepository) GetByCategories(categories []string) ([]Task, error) {
	var tasks []Task
	placeholders := make([]string, len(categories))
	args := make([]interface{}, len(categories))

	for i, category := range categories {
		placeholders[i] = "?"
		args[i] = category
	}

	query := fmt.Sprintf(`SELECT id, title, description, role, category FROM tasks WHERE category IN (%s) ORDER BY priority`,
		strings.Join(placeholders, ","))

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return tasks, err
	}

	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Role, &task.Category); err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return tasks, err
	}

	return tasks, nil
}

// BulkAdd inserts multiple tasks into the database.
// It takes an array of Task objects as input and returns an error.
// A transaction is started using the database connection.
// A prepared statement is created to insert a task into the "tasks" table.
// For each task in the input array, the statement is executed with the task properties as arguments.
// If any error occurs during the insertion process, the transaction is rolled back and the error is returned.
// Otherwise, the transaction is committed and the method returns nil.
func (t *TaskRepository) BulkAdd(tasks []Task) error {

	tx, err := t.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO tasks (priority, title, description, role, category) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for _, task := range tasks {
		_, err = stmt.Exec(task.Priority, task.Title, task.Description, task.Role, task.Category)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}