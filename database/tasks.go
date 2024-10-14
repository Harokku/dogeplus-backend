package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"strconv"
	"strings"
)

const PRO22 = "pro22"

// Task represents a task with its properties.
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

// GetByCategories retrieves tasks based on the provided category.
// It constructs a query to fetch tasks from the "tasks" table where the category matches the provided input.
// The query is executed, and the resulting rows are scanned into a slice of Task objects.
// Returns a slice of Task objects and an error, if any occur during query execution.
func (t *TaskRepository) GetByCategories(category string) ([]Task, error) {
	// Prepare the placeholder and arguments for the query
	placeholders := "?, ?"
	args := []interface{}{category, "PRO22"}

	// Construct the query using the placeholders
	query := fmt.Sprintf(`SELECT id, priority, title, description, role, category, escalation_level, incident_level 
	                      FROM tasks 
	                      WHERE category IN (%s) 
	                      ORDER BY priority`, placeholders)

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
	query := fmt.Sprintf(`SELECT id, priority, title, description, role, category, escalation_level FROM tasks WHERE category IN (%s) AND escalation_level IN (%s) ORDER BY priority`,
		strings.Join(catPlaceholders, ","), strings.Join(escPlaceholders, ","))

	return t.executeAndScanResults(query, args)
}

// TODO: Delete after testing of new implementation
// GetGyCategoryAndEscalationLevel retrieves tasks by category and escalation level range.
// It takes a category, starting escalation, and final escalation as input and returns a list of Task objects and an error.
// The method validates the escalation levels and constructs an SQL query based on the input parameters.
//func (t *TaskRepository) GetGyCategoryAndEscalationLevel(category, startingEscalation, finalEscalation string) ([]Task, error) {
//	var tasks []Task
//
//	// The escalation levels ranked in order
//	rankedLevels := GetEscalationLevels()
//	levelMap := make(map[string]int)
//	for i, level := range rankedLevels {
//		levelMap[level] = i
//	}
//
//	startIdx, startOk := levelMap[strings.ToLower(startingEscalation)]
//	endIdx, endOk := levelMap[strings.ToLower(finalEscalation)]
//
//	if !startOk || !endOk {
//		return nil, fmt.Errorf("invalid escalation levels: %s or %s", startingEscalation, finalEscalation)
//	}
//
//	if startIdx == endIdx {
//		return nil, fmt.Errorf("starting and final escalation levels cannot be the same")
//	}
//
//	// Adjust the indices for correct slicing
//	startIdx++
//	if startIdx >= len(rankedLevels) || startIdx > endIdx {
//		return tasks, nil
//	}
//
//	// Construct the query
//	query := `
//		SELECT id, priority, title, description, role, category, escalation_level
//		FROM tasks
//		WHERE (LOWER(category) = ? OR LOWER(category) = LOWER(?))
//		AND LOWER(escalation_level) IN (?);
//	`
//
//	// Prepare the escalation levels for the query
//	escLevels := rankedLevels[startIdx : endIdx+1]
//	escPlaceholders := make([]string, len(escLevels))
//	args := make([]interface{}, len(escLevels)+2)
//	for i, level := range escLevels {
//		escPlaceholders[i] = "?"
//		args[i+2] = level
//	}
//
//	args[0] = PRO22
//	args[1] = category
//
//	query = fmt.Sprintf(strings.Replace(query, "IN (?)", "IN ("+strings.Join(escPlaceholders, ",")+")", 1))
//
//	return t.executeAndScanResults(query, args)
//}

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

	// Construct the base query
	query := `
        SELECT id, priority, title, description, role, category, escalation_level, incident_level 
        FROM tasks 
        WHERE (LOWER(category) = ? OR LOWER(category) = LOWER(?))
        AND (LOWER(escalation_level) IN (?)`

	// Prepare the escalation levels for the query
	escLevels := rankedLevels[startIdx : endIdx+1]
	escPlaceholders := make([]string, len(escLevels))
	args := make([]interface{}, len(escLevels)+2)
	for i, level := range escLevels {
		escPlaceholders[i] = "?"
		args[i+2] = level
	}

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
		query += ` OR (LOWER(escalation_level) = 'incidente' AND LOWER(incident_level) IN (` + strings.Join(incidentLevelPlaceholders, ",") + `))`
	}

	query += `);`

	args[0] = category
	args[1] = category
	query = fmt.Sprintf(strings.Replace(query, "IN (?)", "IN ("+strings.Join(escPlaceholders, ",")+")", 1))

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

// parsePriority converts a priority string to an int, returns 0 if invalid.
// It logs a warning if the conversion fails but does not halt execution.
func parsePriority(priority string) int {
	priorityInt, err := strconv.Atoi(priority)
	if err != nil {
		log.Printf("failed to parse priority: %v", err)
		return 0
	}
	return priorityInt
}

// isBlockEmpty checks if a block of 5 columns is empty
func isBlockEmpty(block []string) bool {
	for _, cell := range block {
		if cell != "" {
			return false
		}
	}
	return true
}

// padBlock ensures a block has exactly 'blockSize' columns by appending empty strings if necessary.
func padBlock(block []string, blockSize int) []string {
	if len(block) >= blockSize {
		return block
	}
	paddedBlock := make([]string, blockSize)
	copy(paddedBlock, block)
	return paddedBlock
}

// ParseXLSXToTasks converts an Excel file into a slice of Task instances, parsing data from each sheet and handling errors.
func ParseXLSXToTasks(f *excelize.File) ([]Task, error) {
	var tasks []Task

	// Check if the file has any sheets
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("the file does not contain any sheets")
	}

	// Iterate over each sheet in the file
	for _, sheetName := range sheetList {
		// Get all the rows from the current sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to get rows from sheet %s: %v", sheetName, err)
		}

		// Fetch the header row roles
		if len(rows) < 3 {
			return nil, fmt.Errorf("the sheet %s does not have the required structure (at least 3 rows needed)", sheetName)
		}

		headerRow := rows[0]

		// Iterate over each row in the sheet, skipping the first 2 rows
		for i, row := range rows {
			if i < 2 {
				continue // Skip the first 2 rows
			}

			// Iterate in blocks of 5 columns starting from the first block
			for j := 0; j < len(row); j += 5 {
				// Ensure there's a block to process
				block := row[j:min(j+5, len(row))]

				// Pad the block to ensure it has exactly 5 columns
				block = padBlock(block, 5)

				// Skip if the block is empty
				if isBlockEmpty(block) {
					continue
				}

				// Fetch the role from the header row based on the block's starting column
				role := ""
				if j < len(headerRow) {
					role = headerRow[j]
				}

				// Create a new Task struct with mapped fields
				task := Task{
					Category:        sheetName,
					Role:            role,
					Priority:        parsePriority(block[0]),   // Convert and map the priority
					Title:           block[1],                  // Map the title field
					Description:     block[2],                  // Map the description field
					EscalationLevel: strings.ToLower(block[3]), // Map the escalation level field
					IncidentLevel:   strings.ToLower(block[4]), // Map the incident level field
				}

				// Append the new task to the tasks slice
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

var escalationLevels = map[string]int{
	"allarme":   1,
	"emergenza": 2,
	"incidente": 3,
}

var incidentLevels = map[string]int{
	"bianca": 1,
	"verde":  2,
	"gialla": 3,
	"rossa":  4,
}

// FilterTasks filters the provided tasks based on category, escalation level, and incident level.
//
// Parameters:
// - tasks: A slice of Task structs to filter.
// - category: The category to filter by. Tasks must match this category or be "pro22" (case insensitive).
// - escalationLevel: The maximum escalation level to allow, incrementally ranked as "allarme", "emergenza", "incidente".
// - incidentLevel: The maximum incident level to allow when the escalation level is "incidente", incrementally ranked as "bianca", "verde", "gialla", "rossa".
//
// Returns:
// - A slice of Task structs that meet the filter criteria.
func FilterTasks(tasks []Task, category, escalationLevel, incidentLevel string) []Task {
	var filteredTasks []Task

	for _, task := range tasks {
		// Check if the task's category matches the input category or "pro22" (case insensitive).
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
	if startIdx == endIdx {
		return nil, fmt.Errorf("starting and final escalation levels cannot be the same")
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
		if strings.ToLower(task.Category) == strings.ToLower(category) && escLevelSet[strings.ToLower(task.EscalationLevel)] {
			filteredTasks = append(filteredTasks, task)
		} else if strings.ToLower(task.EscalationLevel) == "incidente" && strings.ToLower(finalEscalation) == "incidente" {
			if incidentOk && incidentLevels[strings.ToLower(task.IncidentLevel)] <= incidentIdx {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}
	return filteredTasks, nil
}

// MergeTasks merges two slices of Tasks by updating or removing existing tasks and adding new ones based on their Title and Category.
// If a task in the update slice has only Title and Category populated, it will remove the corresponding task from the original slice.
func MergeTasks(original, update []Task) ([]Task, error) {
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
