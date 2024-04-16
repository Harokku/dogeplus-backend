package database

import "database/sql"

// ActiveEvents represents active events with relative properties
type ActiveEvents struct {
	UUID        string `json:"uuid"`
	EventNumber int    `json:"event_number"`
	EventDate   string `json:"event_date"`
	CentralID   string `json:"central_id"`
	Priority    int    `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	ModifiedBy  string `json:"modified_by"`
	Timestamp   string `json:"timestamp"`
}

const (
	taskNotdone = "notdone"
	taskWorking = "working"
	taskDone    = "done"
)

// ActiveEventsRepository represents a repository for managing active events
type ActiveEventsRepository struct {
	db *sql.DB
}

// NewActiveEventRepository creates a new instance of ActiveEventsRepository with the provided database connection.
// It returns a pointer to the created ActiveEventsRepository.
func NewActiveEventRepository(db *sql.DB) *ActiveEventsRepository {
	return &ActiveEventsRepository{db: db}
}

// Add inserts a new active event record into the database.
// The task parameter represents the active event object to be added.
// The tx parameter is a transaction object that encapsulates the database transaction.
// This method executes a database query to insert the provided active event data into the active_events table.
// It returns an error if the database operation fails.
func (e *ActiveEventsRepository) Add(tx *sql.Tx, task ActiveEvents) error {
	query := `INSERT INTO active_events (UUID, event_number , event_date, central_id, Priority, Title, Description, Role, Status,modified_by, Timestamp)
			   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := tx.Exec(query, task.UUID, task.EventNumber, task.EventDate, task.CentralID, task.Priority, task.Title, task.Description, task.Role, task.Status, task.ModifiedBy, task.Timestamp)

	return err
}

// TODO: Implement builder
func (e *ActiveEventsRepository) TaskToActiveEvent(task Task) ActiveEvents {
	return ActiveEvents{
		UUID:        "",
		EventNumber: 0,
		EventDate:   "",
		CentralID:   "",
		Priority:    task.Priority,
		Title:       task.Title,
		Description: task.Description,
		Role:        task.Role,
		Status:      taskNotdone,
		ModifiedBy:  "",
		Timestamp:   "",
	}
}

func (e *ActiveEventsRepository) CreateFromTaskList(tasks []Task) error {

	// Begin transaction
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		t := e.TaskToActiveEvent(task)
		err = e.Add(tx, t)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
