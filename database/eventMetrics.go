package database

import (
	"fmt"
	"sync"
)

var (
	taskCompletionInstance *TaskCompletionMap
	taskCompletionOnce     sync.Once
)

// TaskCompletionInfo is a struct type that represents the completion information of a task.
//
// It has two fields 'Completed' and 'Total' which store the number of completed tasks and
// the total number of tasks respectively.
type TaskCompletionInfo struct {
	Completed int
	Total     int
}

// Ratio calculates the completion ratio of a task.
// It returns 0 if the total number of tasks is 0.
func (tci *TaskCompletionInfo) Ratio() float32 {
	if tci.Total == 0 {
		return 0
	}
	return float32(tci.Completed) / float32(tci.Total)
}

// TaskCompletionMap is a struct type that represents a map of task completion information
//
// It has two fields 'sync.RWMutex' for concurrent-safe access to the map and 'Data' which is the actual map
// that associates task number (string) with their completion information (TaskCompletionInfo)
type TaskCompletionMap struct {
	mu   sync.RWMutex
	Data map[int]TaskCompletionInfo
}

// UpdateEventStatus updates the status of a specific event in the TaskCompletionMap.
// It takes the event number and the new status as parameters.
// If the event number exists in the map, the function updates the completion information
// based on the new status. If the status is "done", the number of completed tasks is
// incremented by 1. If the status is "working" or "notdone", the number of completed tasks
// is decremented by 1. The updated completion information is then stored back in the map.
// If the event number does not exist in the map, no action is taken.
// This method uses a lock to ensure concurrent-safe access to the map.
func (tcm *TaskCompletionMap) UpdateEventStatus(eventNumber int, status string) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	if data, ok := tcm.Data[eventNumber]; ok {
		if status == "done" {
			data.Completed++
		} else if status == "working" || status == "notdone" {
			data.Completed--
		}
		tcm.Data[eventNumber] = data
	}
}

// AddMultipleNotDoneTasks is a method of the TaskCompletionMap type. It adds the specified number
// of tasks to the total number of tasks for the given event number. If the event number does not
// exist in the map, no action is taken. This method uses a lock to ensure concurrent-safe access
// to the map.
func (tcm *TaskCompletionMap) AddMultipleNotDoneTasks(eventNumber int, numberOfTasks int) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	if data, ok := tcm.Data[eventNumber]; ok {
		data.Total += numberOfTasks
		tcm.Data[eventNumber] = data
	}
}

// AddNewEvent adds a new event to the TaskCompletionMap with the specified
// event number and the number of tasks. If the event number already exists
// in the map, no action is taken. This method uses a lock to ensure
// concurrent-safe access to the map.
func (tcm *TaskCompletionMap) AddNewEvent(eventNumber int, numberOfTasks int) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	if _, ok := tcm.Data[eventNumber]; ok {
		return
	}

	tcm.Data[eventNumber] = TaskCompletionInfo{
		Completed: 0,
		Total:     numberOfTasks,
	}
}

// DeleteEvent removes an event from the TaskCompletionMap with the specified
// event ID. If the event ID does not exist in the map, no action is taken.
// This method uses a lock to ensure concurrent-safe access to the map.
func (tcm *TaskCompletionMap) DeleteEvent(eventId int) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()
	delete(tcm.Data, eventId)
}

// GetTaskCompletionMapInstance retrieves the singleton instance of TaskCompletionMap.
//
// It constructs a new TaskCompletionMap instance if it hasn't been created yet.
// The function initializes the Data map of the TaskCompletionMap instance
// with the task completion information from the input AggregatedActiveEvents slice.
//
// This function is thread-safe due to the use of sync.Once.
// It returns a pointer to the TaskCompletionMap instance.
func GetTaskCompletionMapInstance(events []AggregatedActiveEvents) *TaskCompletionMap {
	taskCompletionOnce.Do(func() {
		taskCompletionInstance = &TaskCompletionMap{
			Data: make(map[int]TaskCompletionInfo),
			mu:   sync.RWMutex{},
		}
		for _, event := range events {
			taskCompletionInstance.Data[event.EventNumber] = TaskCompletionInfo{
				Completed: event.Done,
				Total:     event.Total,
			}
		}
	})
	return taskCompletionInstance
}

// Level is a string type used to represent different levels of allowed escalation or incident severity.
type Level string

const (
	Allarme   Level = "allarme"
	Emergenza Level = "emergenza"
	Incidente Level = "incidente"
)

// rankedLevels is used to determine Level priority
var rankedLevels = map[Level]int{
	Allarme:   1,
	Emergenza: 2,
	Incidente: 3,
}

// EscalationLevels is a struct type that represents a set of escalation levels for different event numbers.
// It has one field 'Levels' which is a map that stores the event numbers as keys and their respective escalation levels as values.
// Level is a string type used to represent different levels of allowed escalation or incident severity.
// Add adds a new escalation level for a specific event number to the EscalationLevels struct.
// If the event number is not already present in the levels map or the new level is higher than the existing level,
// the new level is added to the levels map.
type EscalationLevels struct {
	Levels map[int]Level
	mu     sync.RWMutex
}

// NewEscalationLevels constructs a new EscalationLevels struct.
// It returns a pointer to the new EscalationLevels struct with an empty Levels map.
func NewEscalationLevels() *EscalationLevels {
	return &EscalationLevels{Levels: make(map[int]Level)}
}

// NewEscalationLevelsFromData constructs a new EscalationLevels struct from the given data map.
// It iterates over the data map, retrieves the levels for each event number, and adds them to the new EscalationLevels struct.
// The function returns the constructed EscalationLevels struct.
func NewEscalationLevelsFromData(data map[int][]Level) *EscalationLevels {
	el := NewEscalationLevels()
	for eventNumber, levels := range data {
		for _, level := range levels {
			el.Add(eventNumber, level)
		}
	}
	return el
}

// convertDbResultToData converts the provided DB data into a map of event numbers and levels.
// It iterates over the dbData slice and retrieves the escalation level for each event.
// If the escalation level is one of the allowed levels (Allarme, Emergenza, Incidente), it adds it to the map under the respective event number.
// If the escalation level is not recognized, it returns an error with a message indicating the unknown event number with associated wrong level.
// The function returns the resulting map of event numbers and levels, along with any potential error.
func convertDbResultToData(dbData []ActiveEvents) (map[int][]Level, error) {
	var result = make(map[int][]Level)
	for _, event := range dbData {
		level := Level(event.EscalationLevel)

		switch level {
		case Allarme, Emergenza, Incidente:
			result[event.EventNumber] = append(result[event.EventNumber], level)
		default:
			return nil, fmt.Errorf("unknown level: %s for event number: %d", level, event.EventNumber)
		}
	}
	return result, nil
}

// Add adds a new escalation level for a specific event number to the EscalationLevels struct.
// If the event number is not already present in the levels map or the new level is higher than the existing level,
// the new level is added to the levels map.
func (el *EscalationLevels) Add(eventNumber int, level Level) {
	el.mu.Lock()
	defer el.mu.Unlock()

	// Only add if it does not exist or level is higher
	if existingLevel, ok := el.Levels[eventNumber]; !ok || rankedLevels[level] > rankedLevels[existingLevel] {
		el.Levels[eventNumber] = level
	}
}

// Remove deletes the escalation level for a specific event number from the Levels map.
// If the event number is not present in the Levels map, nothing happens.
func (el *EscalationLevels) Remove(eventNumber int) {
	el.mu.Lock()
	defer el.mu.Unlock()

	delete(el.Levels, eventNumber)
}

// Escalate escalates the level of a specific event number in the EscalationLevels struct.
// It checks if the newLevel is higher than the existing level for the given event number.
// If it is, the newLevel is updated in the Levels map.
// If the newLevel is not one of the allowed levels (Allarme, Emergenza, Incidente),
// an error is returned with a message indicating the invalid level.
//
// Parameters:
// - eventNumber: the number of the event for which the level is being escalated
// - newLevel: the new level to be escalated to
//
// Returns:
//   - error: an error if the newLevel is not one of the allowed levels
//     or if the newLevel is not higher than the existing level for the event number
func (el *EscalationLevels) Escalate(eventNumber int, newLevel Level) error {
	el.mu.Lock()
	defer el.mu.Unlock()

	switch newLevel {
	case Allarme, Emergenza, Incidente:
		// Only escalate if newLevel is higher
		if existingLevel, ok := el.Levels[eventNumber]; ok && rankedLevels[newLevel] > rankedLevels[existingLevel] {
			el.Levels[eventNumber] = newLevel
		}
	default:
		return fmt.Errorf("invalid level provided: %s", newLevel)
	}

	return nil
}

// Deescalate deescalates the level of a specific event number in the EscalationLevels struct.
// It checks if the newLevel is lower than the existing level for the given event number.
// If it is, the newLevel is updated in the Levels map.
// If the newLevel is not one of the allowed levels (Allarme, Emergenza, Incidente),
// an error is returned with a message indicating the invalid level.
//
// Parameters:
// - eventNumber: the number of the event for which the level is being deescalated
// - newLevel: the new level to be deescalated to
//
// Returns:
//   - error: an error if the newLevel is not one of the allowed levels
//     or if the newLevel is not lower than the existing level for the event number
func (el *EscalationLevels) Deescalate(eventNumber int, newLevel Level) error {
	el.mu.Lock()
	defer el.mu.Unlock()

	switch newLevel {
	case Allarme, Emergenza, Incidente:
		// Only deescalate if newLevel is lower
		if existingLevel, ok := el.Levels[eventNumber]; ok && rankedLevels[newLevel] < rankedLevels[existingLevel] {
			el.Levels[eventNumber] = newLevel
		}
	default:
		return fmt.Errorf("invalid level provided: %s", newLevel)
	}

	return nil
}
