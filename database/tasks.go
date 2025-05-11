// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
//
// This file contains functionality related to tasks management:
// - Task data structure
// - TaskRepository for database operations
// - Excel file parsing
// - Task filtering and merging utilities
package database

import (
	"database/sql"
	"dogeplus-backend/config"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ==================================================================================
// CONSTANTS AND DATA STRUCTURES
// ==================================================================================

// TaskRepository represents a repository for managing tasks.
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository initializes a new instance of TaskRepository.
// It takes a *sql.DB object as input and returns a pointer to TaskRepository.
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// BeginTrans starts a new transaction from the given DB connection.
func (t *TaskRepository) BeginTrans() (*sql.Tx, error) {
	tx, err := t.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	return tx, nil
}

type TaskRepositoryTransaction struct {
	repo *TaskRepository
	*sql.Tx
}

// WithTransaction runs the queries wrapped in a transaction.
func (t *TaskRepository) WithTransaction(fn func(*TaskRepositoryTransaction) error) error {
	tx, err := t.BeginTrans()
	if err != nil {
		return err
	}

	trx := &TaskRepositoryTransaction{t, tx}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err with commit err
		}
	}()

	if err := fn(trx); err != nil {
		return err
	}
	return nil
}

// GetCategories retrieves distinct categories from the "tasks" table in the database.
// It does not take any input and returns an array of category strings and an error.
// A query is executed to select distinct categories from the "tasks" table.
// The retrieved rows are scanned and each category is appended to the categories array.
// If any error occurs during the process, it is returned along with the categories array.
// Finally, the categories array and the error are returned.
func (t *TaskRepository) GetCategories(configFile config.Config) ([]string, error) {
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

	//TODO: implement local task merge

	return categories, nil
}

// GetByCategories retrieves tasks based on the provided category.
// It constructs a query to fetch tasks from the "tasks" table where the category matches the provided input.
// The query is executed, and the resulting rows are scanned into a slice of Task objects.
// Returns a slice of Task objects and an error, if any occur during query execution.
func (t *TaskRepository) GetByCategories(category string) ([]Task, error) {
	// Prepare the placeholder and arguments for the query
	placeholders := "?, ?"
	args := []interface{}{category, "PRO22"}

	// Construct the query using the placeholders
	query := "SELECT id, priority, title, description, role, category, escalation_level, incident_level FROM tasks WHERE category IN (" + placeholders + ") ORDER BY priority"

	// Execute the query and scan the results
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
	query := fmt.Sprintf("SELECT id, priority, title, description, role, category, escalation_level FROM tasks WHERE category IN (%s) AND escalation_level IN (%s) ORDER BY priority",
		strings.Join(catPlaceholders, ","), strings.Join(escPlaceholders, ","))

	return t.executeAndScanResults(query, args)
}

// GetGyCategoryAndEscalationLevel retrieves tasks based on specified category, escalation levels, and incident level conditions.
// category: the category to filter tasks.
// startingEscalation: the initial escalation level.
// finalEscalation: the target escalation level.
// incidentLevel: the level of incident to filter if the final escalation is 'incidente'.
// Returns a slice of Task and an error if any occur during query execution.
func (t *TaskRepository) GetGyCategoryAndEscalationLevel(category, startingEscalation, finalEscalation, incidentLevel string) ([]Task, error) {
	var tasks []Task

	// The escalation levels ranked in order
	rankedLevels := GetEscalationLevels()
	levelMap := make(map[string]int)
	for i, level := range rankedLevels {
		levelMap[level] = i
	}
	startIdx, startOk := levelMap[strings.ToLower(startingEscalation)]
	endIdx, endOk := levelMap[strings.ToLower(finalEscalation)]
	if !startOk || !endOk {
		return nil, fmt.Errorf("invalid escalation levels: %s or %s", startingEscalation, finalEscalation)
	}
	if startIdx == endIdx {
		return nil, fmt.Errorf("starting and final escalation levels cannot be the same")
	}

	// Check if incidentLevel is required and provided
	if strings.ToLower(finalEscalation) == "incidente" && incidentLevel == "" {
		return nil, fmt.Errorf("incident level must be provided when final escalation is 'incidente'")
	}

	// Adjust the indices for correct slicing
	startIdx++
	if startIdx >= len(rankedLevels) || startIdx > endIdx {
		return tasks, nil
	}

	// Construct the base query with category filtering
	query := `
        SELECT id, priority, title, description, role, category, escalation_level, incident_level 
        FROM tasks 
  		WHERE (LOWER(category) = LOWER(?) OR LOWER(category) = 'pro22')`

	args := make([]interface{}, 0)
	args = append(args, category)

	// Prepare the escalation levels for the query
	escLevels := rankedLevels[startIdx : endIdx+1]
	escPlaceholders := make([]string, len(escLevels))
	for i, level := range escLevels {
		escPlaceholders[i] = "?"
		args = append(args, level)
	}

	// Append the escalation level placeholders into the query
	query += ` AND (LOWER(escalation_level) IN (` + strings.Join(escPlaceholders, ", ") + `)`

	// Handle 'incidente' specific case if incidentLevel is provided
	if strings.ToLower(finalEscalation) == "incidente" {
		incidentLevels := map[string]int{
			"bianca": 0,
			"verde":  1,
			"gialla": 2,
			"rossa":  3,
		}
		incidentIdx, incidentOk := incidentLevels[strings.ToLower(incidentLevel)]
		if !incidentOk {
			return nil, fmt.Errorf("invalid incident level: %s", incidentLevel)
		}

		incidentLevelPlaceholders := []string{}
		for level, idx := range incidentLevels {
			if idx <= incidentIdx {
				incidentLevelPlaceholders = append(incidentLevelPlaceholders, "?")
				args = append(args, level)
			}
		}

		// Enclose the entire condition for incidents within parentheses
		query += ` OR (LOWER(escalation_level) = 'incidente' AND LOWER(incident_level) IN (` + strings.Join(incidentLevelPlaceholders, ", ") + `)))`
	} else {
		// End the escalation level condition, if no incident level is provided or escalation level is not incidente
		query += `)`
	}

	return t.executeAndScanResults(query, args)
}

// executeAndScanResults executes the given SQL query with the provided arguments,
// scans the resulting rows, and returns an array of Task objects and an error.
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
		if err := rows.Scan(&task.ID, &task.Priority, &task.Title, &task.Description, &task.Role, &task.Category, &task.EscalationLevel, &task.IncidentLevel); err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return tasks, err
	}

	return tasks, nil
}

// BulkAdd adds a slice of tasks using the given transaction or connection.
func (t *TaskRepository) BulkAdd(tx *sql.Tx, tasks []Task) error {
	var err error
	for _, task := range tasks {
		if tx != nil {
			_, err = tx.Exec(
				"INSERT INTO tasks (category, role, priority, title, description,escalation_level,incident_level) VALUES (?, ?, ?, ?, ?, ?, ?)",
				task.Category, task.Role, task.Priority, task.Title, task.Description, task.EscalationLevel, task.IncidentLevel,
			)
		} else {
			_, err = t.db.Exec(
				"INSERT INTO tasks (category, role, priority, title, description,escalation_level,incident_level) VALUES (?, ?, ?, ?, ?, ?, ?)",
				task.Category, task.Role, task.Priority, task.Title, task.Description, task.EscalationLevel, task.IncidentLevel,
			)
		}
		if err != nil {
			return fmt.Errorf("failed to insert task: %v", err)
		}
	}
	return nil
}

// ClearTasksTable drops the tasks table using the given transaction or connection.
func (t *TaskRepository) ClearTasksTable(tx *sql.Tx) error {
	var err error
	if tx != nil {
		_, err = tx.Exec("DELETE FROM tasks")
		if err != nil {
			return fmt.Errorf("failed to clear tasks table: %v", err)
		}
		_, err = tx.Exec("DELETE FROM sqlite_sequence WHERE name='tasks'")
	} else {
		_, err = t.db.Exec("DELETE FROM tasks")
		if err != nil {
			return fmt.Errorf("failed to clear tasks table: %v", err)
		}
		_, err = t.db.Exec("DELETE FROM sqlite_sequence WHERE name='tasks'")
	}
	if err != nil {
		return fmt.Errorf("failed to reset auto-increment counter: %v", err)
	}
	return nil
}

func (trx *TaskRepositoryTransaction) DropTasksTable() error {
	return trx.repo.ClearTasksTable(trx.Tx)
}

func (trx *TaskRepositoryTransaction) BulkAdd(tasks []Task) error {
	return trx.repo.BulkAdd(trx.Tx, tasks)
}

// FilterTasks filters and sorts tasks based on category, escalation, and incident levels criteria.
// It removes duplicates by keeping tasks with higher escalation/incident levels for the same title.
func FilterTasks(tasks []Task, category, escalationLevel, incidentLevel string) []Task {
	var filteredTasks []Task

	for _, task := range tasks {
		// Check if the task's category matches the input category or "pro22" (case-insensitive).
		if strings.EqualFold(task.Category, category) || strings.EqualFold(task.Category, "pro22") {
			// Ensure the task's escalation level is less than or equal to the input escalation level.
			if escalationLevels[task.EscalationLevel] <= escalationLevels[escalationLevel] {
				// If the input escalation level is "incidente", filter based on the incident level.
				if strings.EqualFold(escalationLevel, "incidente") {
					// Include the task if its incident level is less than or equal to the input incident level.
					if incidentLevels[task.IncidentLevel] <= incidentLevels[incidentLevel] {
						filteredTasks = append(filteredTasks, task)
					}
				} else {
					// Include the task if the escalation level criteria is met and it's not "incidente".
					filteredTasks = append(filteredTasks, task)
				}
			}
		}
	}

	// Create a map to store tasks by title, keeping only the one with higher escalation/incident level
	tasksByTitle := make(map[string]Task)
	for _, task := range filteredTasks {
		if existingTask, exists := tasksByTitle[task.Title]; exists {
			// Compare escalation levels
			if escalationLevels[task.EscalationLevel] > escalationLevels[existingTask.EscalationLevel] {
				tasksByTitle[task.Title] = task
			} else if escalationLevels[task.EscalationLevel] == escalationLevels[existingTask.EscalationLevel] {
				// If same escalation level and it's "incidente", compare incident levels
				if strings.EqualFold(task.EscalationLevel, "incidente") {
					if incidentLevels[task.IncidentLevel] > incidentLevels[existingTask.IncidentLevel] {
						tasksByTitle[task.Title] = task
					}
				}
			}
		} else {
			tasksByTitle[task.Title] = task
		}
	}

	// Convert map back to slice
	filteredTasks = make([]Task, 0, len(tasksByTitle))
	for _, task := range tasksByTitle {
		filteredTasks = append(filteredTasks, task)
	}

	// Sort by Priority
	sort.Slice(filteredTasks, func(i, j int) bool {
		return filteredTasks[i].Priority < filteredTasks[j].Priority
	})

	return filteredTasks
}

// FilterTasksForEscalation filters tasks based on category, escalation levels, and incident level conditions.
func FilterTasksForEscalation(tasks []Task, category, startingEscalation, finalEscalation, incidentLevel string) ([]Task, error) {
	var filteredTasks []Task
	// The escalation levels ranked in order
	rankedLevels := GetEscalationLevels()
	levelMap := make(map[string]int)
	for i, level := range rankedLevels {
		levelMap[level] = i
	}
	startIdx, startOk := levelMap[strings.ToLower(startingEscalation)]
	endIdx, endOk := levelMap[strings.ToLower(finalEscalation)]
	if !startOk || !endOk {
		return nil, fmt.Errorf("invalid escalation levels: %s or %s", startingEscalation, finalEscalation)
	}
	// New condition to handle same starting and final escalation levels
	if startIdx == endIdx {
		// Allow processing if both are "incidente"
		if !(strings.EqualFold(startingEscalation, "incidente") && strings.EqualFold(finalEscalation, "incidente")) {
			return nil, fmt.Errorf("starting and final escalation levels cannot be the same")
		}
	}
	// Check if incidentLevel is required and provided
	if strings.ToLower(finalEscalation) == "incidente" && incidentLevel == "" {
		return nil, errors.New("incident level must be provided when final escalation is 'incidente'")
	}
	// Adjust the indices for correct slicing
	startIdx++
	if startIdx >= len(rankedLevels) || startIdx > endIdx {
		return filteredTasks, nil
	}
	// Prepare the escalation levels for filtering
	escLevels := rankedLevels[startIdx : endIdx+1]
	escLevelSet := make(map[string]bool, len(escLevels))
	for _, level := range escLevels {
		escLevelSet[strings.ToLower(level)] = true
	}
	// Handle 'incidente' specific case if incidentLevel is provided
	incidentLevels := map[string]int{
		"bianca": 0,
		"verde":  1,
		"gialla": 2,
		"rossa":  3,
	}
	var incidentIdx int
	var incidentOk bool
	if strings.ToLower(finalEscalation) == "incidente" {
		incidentIdx, incidentOk = incidentLevels[strings.ToLower(incidentLevel)]
		if !incidentOk {
			return nil, fmt.Errorf("invalid incident level: %s", incidentLevel)
		}
	}
	// Filter tasks
	for _, task := range tasks {
		taskEscLevel := strings.ToLower(task.EscalationLevel)
		taskIncidentLevel := strings.ToLower(task.IncidentLevel)
		taskCategory := strings.ToLower(task.Category)

		if taskCategory == strings.ToLower(category) || taskCategory == strings.ToLower("pro22") {
			if strings.ToLower(finalEscalation) == "incidente" && taskEscLevel == "incidente" {
				taskIncidentLevelIdx, ok := incidentLevels[taskIncidentLevel]
				if ok && taskIncidentLevelIdx <= incidentIdx {
					filteredTasks = append(filteredTasks, task)
				}
			} else if escLevelSet[taskEscLevel] {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	return filteredTasks, nil
}

// MergeTasks merges two slices of Tasks by updating or removing existing tasks and adding new ones based on their Title.
// It applies the following rules:
//  1. If multiple tasks with the same title exist in original slice: keep only the one with higher Escalation level,
//     also check for the same task if escalationLevel is "incidente" keep only the one with higher incidentLevel.
//  2. If task exists in both original and update with the same title: update the original task with update data.
//  3. If task exists only in update slice append it to the original, follow same rules as 1 if multiple tasks
//     with same title exist in update slice (for escalationLevel and incidentLevel).
func MergeTasks(original, update []Task) ([]Task, error) {
	// Helper function to check if a Task in the update slice has only Title populated.
	isOnlyTitleAndCategoryPopulated := func(task Task) bool {
		return task.Priority == 0 &&
			task.Description == "" &&
			task.Category == "" &&
			task.Role == "" &&
			task.EscalationLevel == "" &&
			task.IncidentLevel == ""
	}

	// Helper function to compare tasks and return the one with higher escalation level
	// If both have the same escalation level "incidente", return the one with higher incident level
	compareTaskLevels := func(task1, task2 Task) Task {
		// Get escalation levels for comparison
		esc1 := strings.ToLower(task1.EscalationLevel)
		esc2 := strings.ToLower(task2.EscalationLevel)

		// If escalation levels are different, return the task with higher level
		if esc1 != esc2 {
			if escalationLevels[esc1] > escalationLevels[esc2] {
				return task1
			}
			return task2
		}

		// If both have escalation level "incidente", compare incident levels
		if esc1 == "incidente" && esc2 == "incidente" {
			inc1 := strings.ToLower(task1.IncidentLevel)
			inc2 := strings.ToLower(task2.IncidentLevel)

			if incidentLevels[inc1] > incidentLevels[inc2] {
				return task1
			}
			return task2
		}

		// If escalation levels are the same but not "incidente", return either one (task1 in this case)
		return task1
	}

	// Rule 1: Process original slice to keep only tasks with highest escalation/incident level for each title
	originalByTitle := make(map[string]Task)
	for _, task := range original {
		title := task.Title
		if existingTask, exists := originalByTitle[title]; exists {
			// Compare and keep the task with higher level
			originalByTitle[title] = compareTaskLevels(existingTask, task)
		} else {
			originalByTitle[title] = task
		}
	}

	// Rule 3: Process update slice to keep only tasks with highest escalation/incident level for each title
	updateByTitle := make(map[string]Task)
	for _, task := range update {
		// Skip tasks that only have title populated (used for deletion)
		if !isOnlyTitleAndCategoryPopulated(task) {
			title := task.Title
			if existingTask, exists := updateByTitle[title]; exists {
				// Compare and keep the task with higher level
				updateByTitle[title] = compareTaskLevels(existingTask, task)
			} else {
				updateByTitle[title] = task
			}
		} else {
			// For deletion tasks, always keep them
			updateByTitle[task.Title] = task
		}
	}

	// Convert the filtered original tasks back to a slice
	var filteredOriginal []Task
	for _, task := range originalByTitle {
		filteredOriginal = append(filteredOriginal, task)
	}

	// Rule 2: Merge update tasks into original
	// Create a map for O(1) lookups of filtered original tasks
	originalTaskMap := make(map[string]int)
	for i, task := range filteredOriginal {
		originalTaskMap[task.Title] = i
	}

	// Process the filtered update tasks
	for _, updatedTask := range updateByTitle {
		key := updatedTask.Title
		if idx, exists := originalTaskMap[key]; exists {
			// If the task exists in the original slice
			if isOnlyTitleAndCategoryPopulated(updatedTask) {
				// If only Title and Category are populated in the updated Task, delete the task from original
				filteredOriginal = append(filteredOriginal[:idx], filteredOriginal[idx+1:]...)
				delete(originalTaskMap, key) // Also remove it from the map
			} else {
				// If other fields are populated, update the task in the original slice
				filteredOriginal[idx] = updatedTask
			}
		} else {
			// If the task does not exist in the original slice, append it to original
			if !isOnlyTitleAndCategoryPopulated(updatedTask) {
				filteredOriginal = append(filteredOriginal, updatedTask)
			}
		}
	}

	return filteredOriginal, nil
}

// MergeTasksFixCategory merges two slices of Tasks by updating or removing existing tasks and adding new ones based on their Title and Category.
// If a task in the update slice has only Title and Category populated, it will remove the corresponding task from the original slice.
func MergeTasksFixCategory(original, update []Task) ([]Task, error) {
	// Helper function to check if a Task in the update slice has only Title and Category populated.
	isOnlyTitleAndCategoryPopulated := func(task Task) bool {
		return task.Priority == 0 &&
			task.Description == "" &&
			task.Role == "" &&
			task.EscalationLevel == "" &&
			task.IncidentLevel == ""
	}

	// Create a map to store the index of each original task keyed by "Title|Category".
	// This allows for O(1) lookups to check if a task exists and to find its index quickly.
	originalTaskMap := make(map[string]int)
	for i, task := range original {
		key := task.Title + "|" + task.Category
		originalTaskMap[key] = i
	}

	// Iterate over each Task in the update slice
	for _, updatedTask := range update {
		// Generate a key using "Title|Category"
		key := updatedTask.Title + "|" + updatedTask.Category
		if idx, exists := originalTaskMap[key]; exists {
			// If the task exists in the original slice
			if isOnlyTitleAndCategoryPopulated(updatedTask) {
				// If only Title and Category are populated in the updated Task, delete the task from original
				original = append(original[:idx], original[idx+1:]...)
				delete(originalTaskMap, key) // Also remove it from the map
			} else {
				// If other fields are populated, update the task in the original slice
				original[idx] = updatedTask
			}
		} else {
			// If the task does not exist in the original slice, append it to original
			original = append(original, updatedTask)
		}
	}

	return original, nil
}
