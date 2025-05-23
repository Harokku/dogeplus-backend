package database

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create the active_events table
	_, err = db.Exec(`CREATE TABLE active_events (
		uuid TEXT PRIMARY KEY,
		event_number INTEGER,
		event_date TEXT,
		central_id TEXT,
		priority INTEGER,
		title TEXT,
		description TEXT,
		role TEXT,
		status TEXT,
		modified_by TEXT,
		ip_address TEXT,
		timestamp TEXT,
		escalation_level TEXT
	)`)
	require.NoError(t, err)

	return db
}

// TestActiveEventsRepository_Add tests the Add method of ActiveEventsRepository
func TestActiveEventsRepository_Add(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewActiveEventRepository(db)

	// Create a test event
	testEvent := ActiveEvents{
		UUID:            uuid.New(),
		EventNumber:     1,
		EventDate:       time.Now(),
		CentralID:       "test-central",
		Priority:        1,
		Title:           "Test Event",
		Description:     "This is a test event",
		Role:            "tester",
		Status:          "pending",
		ModifiedBy:      "test-user",
		IpAddress:       "127.0.0.1",
		Timestamp:       time.Now(),
		EscalationLevel: "allarme",
	}

	// Begin a transaction
	tx, err := db.Begin()
	require.NoError(t, err)

	// Add the event
	err = repo.Add(tx, testEvent)
	require.NoError(t, err)

	// Commit the transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify the event was added
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM active_events WHERE uuid = ?", testEvent.UUID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestActiveEventsRepository_TaskToActiveEvent tests the TaskToActiveEvent method
func TestActiveEventsRepository_TaskToActiveEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewActiveEventRepository(db)

	// Create a test task
	testTask := Task{
		Title:       "Test Task",
		Description: "This is a test task",
		Priority:    1,
		Role:        "tester",
	}

	// Convert task to active event
	eventNumber := 1
	centralID := "test-central"
	activeEvent := repo.TaskToActiveEvent(testTask, eventNumber, centralID)

	// Verify the conversion
	assert.Equal(t, eventNumber, activeEvent.EventNumber)
	assert.Equal(t, centralID, activeEvent.CentralID)
	assert.Equal(t, testTask.Title, activeEvent.Title)
	assert.Equal(t, testTask.Description, activeEvent.Description)
	assert.Equal(t, testTask.Role, activeEvent.Role)
	assert.Equal(t, testTask.Priority, activeEvent.Priority) // Priority should match the task's priority
	assert.Equal(t, TaskNotdone, activeEvent.Status)
}

// TestActiveEventsRepository_GetByCentralID tests the GetByCentralID method
func TestActiveEventsRepository_GetByCentralID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewActiveEventRepository(db)

	// Create and add test events
	centralID := "test-central"
	events := []ActiveEvents{
		{
			UUID:            uuid.New(),
			EventNumber:     1,
			EventDate:       time.Now(),
			CentralID:       centralID,
			Priority:        1,
			Title:           "Test Event 1",
			Description:     "This is test event 1",
			Role:            "tester",
			Status:          "pending",
			ModifiedBy:      "test-user",
			IpAddress:       "127.0.0.1",
			Timestamp:       time.Now(),
			EscalationLevel: "allarme",
		},
		{
			UUID:            uuid.New(),
			EventNumber:     2,
			EventDate:       time.Now(),
			CentralID:       centralID,
			Priority:        2,
			Title:           "Test Event 2",
			Description:     "This is test event 2",
			Role:            "tester",
			Status:          "pending",
			ModifiedBy:      "test-user",
			IpAddress:       "127.0.0.1",
			Timestamp:       time.Now(),
			EscalationLevel: "emergenza",
		},
	}

	// Add events to the database
	tx, err := db.Begin()
	require.NoError(t, err)

	for _, event := range events {
		err = repo.Add(tx, event)
		require.NoError(t, err)
	}

	err = tx.Commit()
	require.NoError(t, err)

	// Get events by central ID
	_, eventNumbers, err := repo.GetByCentralID(centralID)

	// Verify the error is of the expected type
	require.Error(t, err)
	multipleEventsErr, ok := err.(*MultipleEventsIdError)
	require.True(t, ok, "Expected MultipleEventsIdError, got %T", err)
	assert.Contains(t, multipleEventsErr.Error(), "Multiple events found for specified centralId")

	// Verify the event numbers
	assert.Len(t, eventNumbers, 2)
	assert.Contains(t, eventNumbers, 1)
	assert.Contains(t, eventNumbers, 2)
}

// TestActiveEventsRepository_UpdateStatus tests the UpdateStatus method
func TestActiveEventsRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewActiveEventRepository(db)

	// Create and add a test event
	eventUUID := uuid.New()
	testEvent := ActiveEvents{
		UUID:            eventUUID,
		EventNumber:     1,
		EventDate:       time.Now(),
		CentralID:       "test-central",
		Priority:        1,
		Title:           "Test Event",
		Description:     "This is a test event",
		Role:            "tester",
		Status:          "pending",
		ModifiedBy:      "test-user",
		IpAddress:       "127.0.0.1",
		Timestamp:       time.Now(),
		EscalationLevel: "allarme",
	}

	// Add the event to the database
	tx, err := db.Begin()
	require.NoError(t, err)

	err = repo.Add(tx, testEvent)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Update the status
	newStatus := "done"
	modifiedBy := "test-updater"
	ipAddress := "192.168.1.1"

	updatedEvent, err := repo.UpdateStatus(eventUUID, newStatus, modifiedBy, ipAddress)
	require.NoError(t, err)

	// Verify the update
	assert.Equal(t, newStatus, updatedEvent.Status)
	assert.Equal(t, modifiedBy, updatedEvent.ModifiedBy)
	assert.Equal(t, ipAddress, updatedEvent.IpAddress)
}

// TestActiveEventsRepository_FilterAndUpdateExistingTasks tests the FilterAndUpdateExistingTasks method
func TestActiveEventsRepository_FilterAndUpdateExistingTasks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewActiveEventRepository(db)
	eventNumber := 1
	centralID := "test-central"

	// Create test tasks
	tasks := []Task{
		{
			ID:          1,
			Priority:    1,
			Title:       "Task 1",
			Description: "Description 1",
			Role:        "Role 1",
		},
		{
			ID:          2,
			Priority:    2,
			Title:       "Task 2",
			Description: "Description 2",
			Role:        "Role 2",
		},
		{
			ID:          3,
			Priority:    3,
			Title:       "Task 3",
			Description: "Description 3",
			Role:        "Role 3",
		},
	}

	// Test scenario 1: No existing events (all tasks should be returned)
	filteredTasks, err := repo.FilterAndUpdateExistingTasks(tasks, eventNumber, centralID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(filteredTasks), "All tasks should be returned when no existing events")

	// Add some events to the database
	tx, err := db.Begin()
	require.NoError(t, err)

	// Add Task 1 with status "done"
	event1 := repo.TaskToActiveEvent(tasks[0], eventNumber, centralID)
	event1.Status = TaskDone
	err = repo.Add(tx, event1)
	require.NoError(t, err)

	// Add Task 2 with status "notdone"
	event2 := repo.TaskToActiveEvent(tasks[1], eventNumber, centralID)
	event2.Status = TaskNotdone
	err = repo.Add(tx, event2)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Test scenario 2 & 3: Existing events with different statuses
	filteredTasks, err = repo.FilterAndUpdateExistingTasks(tasks, eventNumber, centralID)
	require.NoError(t, err)

	// Only Task 3 should be in the filtered list (Task 1 is done, Task 2 is notdone but updated)
	assert.Equal(t, 1, len(filteredTasks), "Only tasks that don't exist should be returned")
	assert.Equal(t, "Task 3", filteredTasks[0].Title, "Task 3 should be in the filtered list")

	// Verify Task 2 was updated in the database
	var priority int
	var description string
	err = db.QueryRow("SELECT priority, description FROM active_events WHERE title = ?", "Task 2").Scan(&priority, &description)
	require.NoError(t, err)
	assert.Equal(t, 2, priority, "Priority should be updated")
	assert.Equal(t, "Description 2", description, "Description should be updated")
}

func TestActiveEventsRepository_DeleteEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewActiveEventRepository(db)

	// Create and add test events
	centralID := "test-central"
	eventNumber := 1
	testEvent := ActiveEvents{
		UUID:            uuid.New(),
		EventNumber:     eventNumber,
		EventDate:       time.Now(),
		CentralID:       centralID,
		Priority:        1,
		Title:           "Test Event",
		Description:     "This is a test event",
		Role:            "tester",
		Status:          "pending",
		ModifiedBy:      "test-user",
		IpAddress:       "127.0.0.1",
		Timestamp:       time.Now(),
		EscalationLevel: "allarme",
	}

	// Add the event to the database
	tx, err := db.Begin()
	require.NoError(t, err)

	err = repo.Add(tx, testEvent)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// Delete the event
	err = repo.DeleteEvent(eventNumber, centralID)
	require.NoError(t, err)

	// Verify the event was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM active_events WHERE event_number = ? AND central_id = ?", eventNumber, centralID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
