package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// Task represents a task with its properties.
type Task struct {
	ID              int    `json:"ID,omitempty"`
	Priority        int    `json:"priority,omitempty"`
	Title           string `json:"title,omitempty"`
	Description     string `json:"description,omitempty"`
	Role            string `json:"role,omitempty"`
	Category        string `json:"category,omitempty"`
	EscalationLevel string `json:"escalation_level,omitempty"`
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

// GetCategories retrieves distinct categories from the "tasks" table in the database.
// It does not take any input and returns an array of category strings and an error.
// A query is executed to select distinct categories from the "tasks" table.
// The retrieved rows are scanned and each category is appended to the categories array.
// If any error occurs during the process, it is returned along with the categories array.
// Finally, the categories array and the error are returned.
func (t *TaskRepository) GetCategories() ([]string, error) {
	var categories []string

	rows, err := t.db.Query("SELECT DISTINCT category FROM tasks")
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
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
	placeholders := make([]string, len(categories))
	args := make([]interface{}, len(categories))

	for i, category := range categories {
		placeholders[i] = "?"
		args[i] = category
	}

	query := fmt.Sprintf(`SELECT id, priority, title, description, role, category, escalation_level FROM tasks WHERE category IN (%s) ORDER BY priority`,
		strings.Join(placeholders, ","))

	return t.executeAndScanResults(query, args)
}

// GetByCategoriesAndEscalationLevels retrieves tasks based on the provided categories and escalation levels.
// It takes two arrays of category and escalation level strings as input and returns an array of Task objects and an error.
// For each category, a placeholder is created and an argument is assigned.
// For each escalation level, a placeholder is created and an argument is assigned.
// A query is constructed using the placeholders and executed against the database connection.
// The retrieved rows are scanned and transformed into Task objects,
// which are then appended to the tasks array.
// If any error occurs during the process, it is returned along with the tasks array.
// Finally, the tasks array and the error are returned.
func (t *TaskRepository) GetByCategoriesAndEscalationLevels(categories []string, escalationLevels []string) ([]Task, error) {
	catPlaceholders := make([]string, len(categories))
	escPlaceholders := make([]string, len(escalationLevels))
	args := make([]interface{}, len(categories)+len(escalationLevels))
	for i, category := range categories {
		catPlaceholders[i] = "?"
		args[i] = category
	}
	for i, escalation := range escalationLevels {
		escPlaceholders[i] = "?"
		args[i+len(categories)] = escalation
	}
	query := fmt.Sprintf(`SELECT id, priority, title, description, role, category, escalation_level FROM tasks WHERE category IN (%s) AND escalation_level IN (%s) ORDER BY priority`,
		strings.Join(catPlaceholders, ","), strings.Join(escPlaceholders, ","))

	return t.executeAndScanResults(query, args)
}

// executeAndScanResults executes the given SQL query with the provided arguments,
// scans the resulting rows and returns an array of Task objects and an error.
// It takes a query string and an array of interface{} for the arguments as input.
// A query is executed against the database connection with the given query and arguments.
// The retrieved rows are scanned and transformed into Task objects,
// which are then appended to the tasks array.
// If any error occurs during the process, it is returned along with the tasks array.
// Finally, the tasks array and the error are returned.
func (t *TaskRepository) executeAndScanResults(query string, args []interface{}) ([]Task, error) {
	var tasks []Task
	rows, err := t.db.Query(query, args...)
	if err != nil {
		return tasks, err
	}

	defer rows.Close()
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Priority, &task.Title, &task.Description, &task.Role, &task.Category, &task.EscalationLevel); err != nil {
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

	stmt, err := tx.Prepare("INSERT INTO tasks (priority, title, description, role, category, escalation_level) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for _, task := range tasks {
		_, err = stmt.Exec(task.Priority, task.Title, task.Description, task.Role, task.Category, task.EscalationLevel)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
