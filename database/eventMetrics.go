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
//
// Example usage:
//
//	var tci TaskCompletionInfo
//	tci.Completed = 5
//	tci.Total = 10
//
//	ratio := tci.Ratio() // Calculate the completion ratio of the task
//	fmt.Println(ratio)
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
// that associates task names (string) with their completion information (TaskCompletionInfo)
//
// Example usage:
//
//	var tcm TaskCompletionMap
//	tcm.Data = make(map[string]TaskCompletionInfo)
//	tcm.Lock()    // Lock access to the map
//	tcm.Data["task1"] = TaskCompletionInfo{Completed: 2, Total: 5}
//	tcm.Data["task2"] = TaskCompletionInfo{Completed: 3, Total: 8}
//	tcm.Unlock()  // Unlock access to the map
//
//	tcm.RLock()   // Read lock access to the map
//	info, exists := tcm.Data["task1"]
//	tcm.RUnlock() // Unlock read access to the map
//	if exists {
//	  ratio := info.Ratio() // Calculate the completion ratio of the task
//	  fmt.Println(ratio)
//	} else {
//	  fmt.Println("Task not found")
//	}
type TaskCompletionMap struct {
	mu   sync.RWMutex
	Data map[string]TaskCompletionInfo
}

// GetTaskCompletionMapInstance returns a singleton instance of TaskCompletionMap
func GetTaskCompletionMapInstance() *TaskCompletionMap {
	taskCompletionOnce.Do(func() {
		taskCompletionInstance = &TaskCompletionMap{
			Data: make(map[string]TaskCompletionInfo),
			mu:   sync.RWMutex{},
		}
		//TODO: Implement initial state fetch from db
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

func (el *EscalationLevels) RemoveEvent(event int) {
	el.mu.Lock()
	defer el.mu.Unlock()

	delete(el.Levels, event)
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
