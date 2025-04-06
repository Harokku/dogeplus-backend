package database

import (
	"database/sql"
	"github.com/google/uuid"
)

type EscalationLevelsDefinition struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type EscalationLevelsDefinitionRepository struct {
	db *sql.DB
}

func NewEscalationLevelsDefinitionRepository(db *sql.DB) *EscalationLevelsDefinitionRepository {
	return &EscalationLevelsDefinitionRepository{db: db}
}

// Add inserts a new EscalationLevelsDefinition record into the database.
// It uses a SQL query to add the UUID, name, and description to the escalation_levels table.
// Returns an error if the execution fails.
func (eld *EscalationLevelsDefinitionRepository) Add(escalationLevel EscalationLevelsDefinition) error {
	_, err := eld.db.Exec(`INSERT INTO escalation_levels (uuid, name, description) VALUES (?, ?, ?)`, escalationLevel.UUID, escalationLevel.Name, escalationLevel.Description)
	return err
}

// GetAll retrieves all escalation levels from the database.
// It executes a SQL query to fetch the UUID, name, and description of all entries in the escalation_levels table.
// Returns a slice of EscalationLevelsDefinition or an error if the query fails or a row scan encounters an issue.
func (eld *EscalationLevelsDefinitionRepository) GetAll() ([]EscalationLevelsDefinition, error) {
	rows, err := eld.db.Query(`SELECT uuid, name, description FROM escalation_levels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var escalationLevels []EscalationLevelsDefinition
	for rows.Next() {
		var escalationLevel EscalationLevelsDefinition
		if err := rows.Scan(&escalationLevel.UUID, &escalationLevel.Name, &escalationLevel.Description); err != nil {
			return escalationLevels, err
		}
		escalationLevels = append(escalationLevels, escalationLevel)
	}

	return escalationLevels, nil
}

// GetByName retrieves an EscalationLevelsDefinition from the database by its name.
// It executes a SQL query to fetch the UUID, name, and description where the name matches the provided value.
// If no row is found or an error occurs during the query, it returns an error.
func (eld *EscalationLevelsDefinitionRepository) GetByName(name string) (EscalationLevelsDefinition, error) {
	query := `SELECT uuid, name, description FROM escalation_levels WHERE name = ?`

	row := eld.db.QueryRow(query, name)

	var escalationLevel EscalationLevelsDefinition

	// Scan row to return variable
	err := row.Scan(&escalationLevel.UUID, &escalationLevel.Name, &escalationLevel.Description)
	if err != nil {
		return escalationLevel, err
	}

	return escalationLevel, nil
}
