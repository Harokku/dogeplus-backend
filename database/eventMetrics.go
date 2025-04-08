// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

import (
	"dogeplus-backend/broadcast"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"sync"
)

const (
	// TopicTaskCompletionMapUpdate is the topic for broadcasting TaskCompletionMap updates
	TopicTaskCompletionMapUpdate = "task_completion_map_update"
)

var (
	taskCompletionInstance  *TaskCompletionMap
	taskCompletionOnce      sync.Once
	escalationLevelInstance *EscalationLevels
	escalationLevelOnce     sync.Once
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
// It has three fields:
// - 'sync.RWMutex' for concurrent-safe access to the map
// - 'Data' which is the actual map that associates task number (int) with their completion information (TaskCompletionInfo)
// - 'cm' which is a reference to the ConnectionManager for broadcasting updates
type TaskCompletionMap struct {
	mu   sync.RWMutex
	Data map[int]TaskCompletionInfo
	cm   *broadcast.ConnectionManager
}

// UpdateEventStatus updates the completion count of a specific event based on the provided status (e.g., "done").
func (tcm *TaskCompletionMap) UpdateEventStatus(eventNumber int, status string) {
	tcm.mu.Lock()

	if data, ok := tcm.Data[eventNumber]; ok {
		if status == "done" {
			data.Completed++
		}
		tcm.Data[eventNumber] = data
	}

	tcm.mu.Unlock()

	// Broadcast the update
	tcm.broadcastUpdate(eventNumber)
}

// AddMultipleNotDoneTasks is a method of the TaskCompletionMap type. It adds the specified number
// of tasks to the total number of tasks for the given event number. If the event number does not
// exist in the map, no action is taken. This method uses a lock to ensure concurrent-safe access
// to the map.
func (tcm *TaskCompletionMap) AddMultipleNotDoneTasks(eventNumber int, numberOfTasks int) {
	tcm.mu.Lock()

	if data, ok := tcm.Data[eventNumber]; ok {
		data.Total += numberOfTasks
		tcm.Data[eventNumber] = data
	}

	tcm.mu.Unlock()

	// Broadcast the update
	tcm.broadcastUpdate(eventNumber)
}

// AddNewEvent adds a new event to the TaskCompletionMap with the specified
// event number and the number of tasks. If the event number already exists
// in the map, no action is taken. This method uses a lock to ensure
// concurrent-safe access to the map.
func (tcm *TaskCompletionMap) AddNewEvent(eventNumber int, numberOfTasks int) {
	tcm.mu.Lock()

	if _, ok := tcm.Data[eventNumber]; ok {
		tcm.mu.Unlock()
		return
	}

	tcm.Data[eventNumber] = TaskCompletionInfo{
		Completed: 0,
		Total:     numberOfTasks,
	}

	tcm.mu.Unlock()

	// Broadcast the update
	tcm.broadcastUpdate(eventNumber)
}

// DeleteEvent removes an event from the TaskCompletionMap with the specified
// event ID. If the event ID does not exist in the map, no action is taken.
// This method uses a lock to ensure concurrent-safe access to the map.
func (tcm *TaskCompletionMap) DeleteEvent(eventId int) {
	tcm.mu.Lock()
	delete(tcm.Data, eventId)
	tcm.mu.Unlock()

	// Broadcast the update - send full map since an event was deleted
	tcm.broadcastUpdate(0)
}

// broadcastUpdate sends the current state of the TaskCompletionMap to all subscribers.
// It marshals the task completion data to JSON and broadcasts it to the TopicTaskCompletionMapUpdate topic.
//
// Parameters:
//   - eventNumber: the number of the event to broadcast. If eventNumber is 0, the entire map is broadcast.
//     If eventNumber is greater than 0, only the data for that specific event is broadcast.
//
// The method does nothing if the ConnectionManager is nil or if the specified event number
// does not exist in the map.
func (tcm *TaskCompletionMap) broadcastUpdate(eventNumber int) {
	if tcm.cm == nil {
		return // No ConnectionManager, can't broadcast
	}

	tcm.mu.RLock()
	var data interface{}

	if eventNumber > 0 {
		// If a specific event number is provided, only broadcast that event's data
		if info, ok := tcm.Data[eventNumber]; ok {
			data = map[string]interface{}{
				"event_number": eventNumber,
				"info":         info,
			}
		} else {
			tcm.mu.RUnlock()
			return // Event not found, nothing to broadcast
		}
	} else {
		// Otherwise broadcast the entire map
		data = tcm.Data
	}
	tcm.mu.RUnlock()

	message := map[string]interface{}{
		"type": "task_completion_update",
		"data": data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		// Log error but continue
		log.Errorf("Error marshalling task completion data: %v", err)
		return
	}

	tcm.cm.BroadcastToTopic(TopicTaskCompletionMapUpdate, jsonData)
}

// GetTaskCompletionMapInstance retrieves the singleton instance of TaskCompletionMap.
//
// It constructs a new TaskCompletionMap instance if it hasn't been created yet.
// The function initializes the Data map of the TaskCompletionMap instance
// with the task completion information from the input AggregatedActiveEvents slice.
//
// This function is thread-safe due to the use of sync.Once.
// It returns a pointer to the TaskCompletionMap instance.
//
// The cm parameter is a pointer to a ConnectionManager instance that will be used
// to broadcast updates to the TaskCompletionMap. If nil, no broadcasting will occur.
func GetTaskCompletionMapInstance(events []AggregatedActiveEvents, cm *broadcast.ConnectionManager) *TaskCompletionMap {
	taskCompletionOnce.Do(func() {
		taskCompletionInstance = &TaskCompletionMap{
			Data: make(map[int]TaskCompletionInfo),
			mu:   sync.RWMutex{},
			cm:   cm,
		}
		for _, event := range events {
			taskCompletionInstance.Data[event.EventNumber] = TaskCompletionInfo{
				Completed: event.Done,
				Total:     event.Total,
			}
		}
	})

	// Update the ConnectionManager if it's provided and different from the current one
	if cm != nil && taskCompletionInstance.cm != cm {
		taskCompletionInstance.cm = cm
	}

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

// GetEscalationLevels returns a slice of strings representing different levels of escalation.
func GetEscalationLevels() []string {
	return []string{"allarme", "emergenza", "incidente"}
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

// GetEscalationLevelsInstance constructs a new EscalationLevels struct from the given data map.
// It iterates over the data map, retrieves the levels for each event number, and adds them to the new EscalationLevels struct.
// The function returns the constructed EscalationLevels struct.
func GetEscalationLevelsInstance(data map[int][]Level) *EscalationLevels {
	escalationLevelOnce.Do(func() {
		escalationLevelInstance = NewEscalationLevels()

		for eventNumber, levels := range data {
			for _, level := range levels {
				escalationLevelInstance.Add(eventNumber, level)
			}
		}
	})
	return escalationLevelInstance
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

// GetLevels returns a thread-safe copy of the Levels map, ensuring the original map cannot be altered by the caller.
func (el *EscalationLevels) GetLevels() map[int]Level {
	el.mu.RLock()
	defer el.mu.RUnlock()

	// Creating a copy of the map to ensure thread-safety and prevent modification by caller.
	copyLevels := make(map[int]Level, len(el.Levels))
	for k, v := range el.Levels {
		copyLevels[k] = v
	}
	return copyLevels
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
