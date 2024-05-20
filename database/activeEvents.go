package database

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type MultipleEventsIdError struct {
	Detail string
}

func (e MultipleEventsIdError) Error() string {
	return fmt.Sprintf("multiple events id: %s", e.Detail)
}

type NoEventsFoundError struct {
	Detail string
}

func (e NoEventsFoundError) Error() string {
	return fmt.Sprintf("no events found: %s", e.Detail)
}

// ActiveEvents represents active events with relative properties
type ActiveEvents struct {
	UUID        uuid.UUID `json:"uuid"`
	EventNumber int       `json:"event_number"`
	EventDate   time.Time `json:"event_date"`
	CentralID   string    `json:"central_id"`
	Priority    int       `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	ModifiedBy  string    `json:"modified_by"`
	IpAddress   string    `json:"ip_address"`
	Timestamp   time.Time `json:"timestamp"`
}

const (
	TaskNotdone = "notdone"
	TaskWorking = "working"
	TaskDone    = "done"
)

// ActiveEventsRepository represents a repository for managing active events
type ActiveEventsRepository struct {
	db *sql.DB
}

// NewActiveEventRepository creates a new instance of ActiveEventsRepository with the provided database connection.
// It returns a pointer to the created ActiveEventsRepository.
func NewActiveEventRepository(db *sql.DB) *ActiveEventsRepository {
	return &ActiveEventsRepository{
		db: db,
	}
}

// Add inserts a new active event record into the database.
// The task parameter represents the active event object to be added.
// The tx parameter is a transaction object that encapsulates the database transaction.
// This method executes a database query to insert the provided active event data into the active_events table.
// It returns an error if the database operation fails.
func (e *ActiveEventsRepository) Add(tx *sql.Tx, task ActiveEvents) error {
	query := `INSERT INTO active_events (UUID, event_number , event_date, central_id, Priority, Title, Description, 
				Role, Status,modified_by,ip_address, Timestamp)
			   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?)`

	_, err := tx.Exec(query, task.UUID, task.EventNumber, task.EventDate, task.CentralID, task.Priority, task.Title,
		task.Description, task.Role, task.Status, task.ModifiedBy, task.IpAddress, task.Timestamp)

	return err
}

// TaskToActiveEvent converts a Task object into an ActiveEvents object.
// It creates a new ActiveEvents object with the UUID generated by uuid.New(),
// the eventNumber and centralId properties from the ActiveEventsRepository,
// and the priority, title, description, role, and status properties from the Task object.
// It returns the converted ActiveEvents object.
func (e *ActiveEventsRepository) TaskToActiveEvent(task Task, eventNumber int, centralId string) ActiveEvents {
	return ActiveEvents{
		UUID:        uuid.New(),
		EventNumber: eventNumber,
		EventDate:   time.Now(),
		CentralID:   centralId,
		Priority:    task.Priority,
		Title:       task.Title,
		Description: task.Description,
		Role:        task.Role,
		Status:      TaskNotdone,
		Timestamp:   time.Now(),
	}
}

// CreateFromTaskList takes a list of tasks and creates corresponding active event records in the database.
// The tasks parameter is a slice of Task objects representing the tasks to be converted into active events.
// This method begins a transaction on the database, converts each task into an ActiveEvents object,
// and inserts it into the active_events table using the Add method. If any error occurs during this process,
// the transaction is rolled back and the error is returned. Otherwise, the transaction is committed.
// It returns an error if the transaction fails to begin, any Add operation fails, or the transaction fails to commit.
func (e *ActiveEventsRepository) CreateFromTaskList(tasks []Task, eventNumber int, centralId string) (err error) {

	// Begin transaction
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}

	// Ensure the transaction will be closed before returning
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	for _, task := range tasks {
		t := e.TaskToActiveEvent(task, eventNumber, centralId)
		err = e.Add(tx, t)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetByCentralID retrieves active events from the database based on the provided central ID.
// The db parameter is a reference to the database connection.
// The centralID parameter is the central ID used to filter the events.
// This method executes a database query to select active events from the active_events table
// with the matching central ID.
// It returns a slice of ActiveEvents representing the retrieved events,
// a slice of int representing the unique event numbers found,
// and an error if the database operation fails.
func (e *ActiveEventsRepository) GetByCentralID(centralId string) ([]ActiveEvents, []int, error) {
	rows, err := e.db.Query(`SELECT uuid, event_number, event_date, central_id, priority, title, 
    description, role, status, modified_by, ip_address, timestamp FROM active_events WHERE central_id = ?`, centralId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	events := []ActiveEvents{}
	eventNumbers := []int{}
	layout := "2006-01-02 15:04:05.999999-07:00"

	// Scan row to return slice and count unique event numbers
	for rows.Next() {
		var tmpEventDate string // event date as string to be scanned to before parsing
		var tmpTimestamp string // timestamp as string to be scanned to before parsing
		var event ActiveEvents
		if err := rows.Scan(&event.UUID, &event.EventNumber, &tmpEventDate, &event.CentralID, &event.Priority, &event.Title,
			&event.Description, &event.Role, &event.Status, &event.ModifiedBy, &event.IpAddress, &tmpTimestamp); err != nil {
			return nil, nil, err
		}
		// parse time to actual type
		parsedEventDate, err := time.Parse(layout, tmpEventDate)
		if err != nil {
			return nil, nil, err
		}
		parsedTimestamp, err := time.Parse(layout, tmpTimestamp)
		if err != nil {
			return nil, nil, err
		}
		event.EventDate = parsedEventDate
		event.Timestamp = parsedTimestamp

		// Append event to slice
		events = append(events, event)

		// Check if event number already exist in slice
		exist := false
		for _, a := range eventNumbers {
			if a == event.EventNumber {
				exist = true
				break
			}
		}

		// add event number if now exist
		if !exist {
			eventNumbers = append(eventNumbers, event.EventNumber)
		}
	}

	// Check number of events found and respond accordingly
	switch {
	case len(eventNumbers) == 0:
		return nil, nil, &NoEventsFoundError{Detail: "No events found for specified centralId"}
	case len(eventNumbers) > 1:
		return nil, eventNumbers, &MultipleEventsIdError{Detail: "Multiple events found for specified centralId"}
	default:
		return events, eventNumbers, nil
	}
}

// GetByCentralAndNumber retrieves active events from the database based on the provided central ID and event number.
// This method executes a database query to select active events from the active_events table
// with the matching central ID and event number.
// It returns a slice of ActiveEvents representing the retrieved events
// and an error if the database operation fails.
func (e *ActiveEventsRepository) GetByCentralAndNumber(eventNumber int, centralId string) ([]ActiveEvents, error) {
	rows, err := e.db.Query(`SELECT uuid, event_number, event_date, central_id, priority, title, description, role, status, modified_by, ip_address, timestamp
								FROM active_events WHERE central_id = ? AND event_number = ?`, centralId, eventNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []ActiveEvents{}
	layout := "2006-01-02 15:04:05.999999-07:00"

	for rows.Next() {
		var tmpEventDate string // event date as string to be scanned to before parsing
		var tmpTimestamp string // timestamp as string to be scanned to before parsing
		var event ActiveEvents
		if err := rows.Scan(&event.UUID, &event.EventNumber, &tmpEventDate, &event.CentralID, &event.Priority,
			&event.Title, &event.Description, &event.Role, &event.Status, &event.ModifiedBy, &event.IpAddress, &tmpTimestamp); err != nil {
			return nil, err
		}
		// parse time to actual type
		parsedEventDate, err := time.Parse(layout, tmpEventDate)
		if err != nil {
			return nil, err
		}
		parsedTimestamp, err := time.Parse(layout, tmpTimestamp)
		if err != nil {
			return nil, err
		}
		event.EventDate = parsedEventDate
		event.Timestamp = parsedTimestamp

		// Append event to slice
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Check number of events fount and respond accordingly
	switch {
	case len(events) == 0:
		return nil, &NoEventsFoundError{Detail: "No events found for specified centralId and event number"}
	default:
		return events, nil
	}
}

// UpdateStatus updates the status of an active event record in the database.
// The uuid parameter is the UUID of the active event record to update.
// The status parameter is the new status value to set.
// The modifiedBy parameter is the username of the user performing the update.
// This method begins a transaction, executes an UPDATE query to update the status and modified_by columns
// of the active event record with the matching UUID, and fetches the updated row.
// If any error occurs during the transaction, the transaction is rolled back and an error is returned.
// Otherwise, the transaction is committed and the updated active event record is returned.
// It returns an error if the database transaction fails to begin, the UPDATE query fails,
// the row fetch fails, or the transaction fails to commit.
func (e *ActiveEventsRepository) UpdateStatus(uuid uuid.UUID, status string, modifiedBy string, ipAddress string) (ActiveEvents, error) {
	// Begin a transaction
	tx, err := e.db.Begin()
	if err != nil {
		return ActiveEvents{}, err
	}

	// Update the status
	_, err = tx.Exec("UPDATE active_events SET status = ?, modified_by = ?, ip_address=? WHERE uuid = ?", status, modifiedBy, ipAddress, uuid)
	if err != nil {
		tx.Rollback()
		return ActiveEvents{}, err
	}

	// Fetch the updated row
	row := tx.QueryRow("SELECT uuid, event_number, event_date, central_id, priority, title, description, role, status, modified_by, ip_address, timestamp FROM active_events WHERE uuid = ?", uuid)

	var event ActiveEvents
	err = row.Scan(&event.UUID, &event.EventNumber, &event.EventDate, &event.CentralID, &event.Priority, &event.Title,
		&event.Description, &event.Role, &event.Status, &event.ModifiedBy, &event.IpAddress, &event.Timestamp)
	if err != nil {
		tx.Rollback()
		return ActiveEvents{}, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return ActiveEvents{}, err
	}

	return event, nil
}

// DeleteEvent deletes an active event record from the database based on the provided central ID and event number.
// This method executes a database query to delete the active event record from the active_events table
// with the matching central ID and event number.
// It returns an error if the database operation fails.
func (e *ActiveEventsRepository) DeleteEvent(eventNumber int, centralId string) error {
	stmt, err := e.db.Prepare("DELETE FROM active_events where central_id = ? AND event_number = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(centralId, eventNumber)
	if err != nil {
		return err
	}

	return nil
}
