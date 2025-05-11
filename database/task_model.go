// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

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

const PRO22 = "pro22"

// GetEscalationLevels returns a slice of escalation levels in order of severity.
func GetEscalationLevels() []string {
	return []string{"allarme", "emergenza", "incidente"}
}

// Escalation and incident level maps for comparison
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
