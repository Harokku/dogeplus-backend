package database

import (
	"database/sql"
	"github.com/google/uuid"
)

type Overview struct {
	UUID           uuid.UUID `json:"uuid"`
	CentralId      string    `json:"central_id"`
	EventNumber    int       `json:"event_number"`
	Location       string    `json:"location"`
	LocationDetail string    `json:"location_detail"`
	Type           string    `json:"type"`
	Level          string    `json:"level"`
	IncidentLevel  string    `json:"incident_level"`
}

type OverviewRepository struct {
	db *sql.DB
}

func NewOverviewRepository(db *sql.DB) *OverviewRepository {
	return &OverviewRepository{db: db}
}

// Add inserts a new overview record into the database and returns an error if any operation fails.
// It also updates the overview struct with the generated UUID.
func (ov *OverviewRepository) Add(overview *Overview) error {
	query := `INSERT INTO overview (uuid, central_id, event_number, location, location_detail, type, level, incident_level) VALUES (?,?,?,?,?,?,?,?)`

	// Generate a new UUID
	newUUID := uuid.New()

	// Update the overview struct with the generated UUID
	overview.UUID = newUUID

	_, err := ov.db.Exec(query, newUUID, overview.CentralId, overview.EventNumber, overview.Location, overview.LocationDetail, overview.Type, overview.Level, overview.IncidentLevel)

	return err
}

// GetAllOverview retrieves all overview records from the database and returns a slice of Overview or an error if it fails.
func (ov *OverviewRepository) GetAllOverview() ([]Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview`

	rows, err := ov.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overviews []Overview

	// Scan rows to return slice
	for rows.Next() {
		var overview Overview
		if err := rows.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
			return nil, err
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetOverviewById retrieves an overview record by its event ID from the database.
func (ov *OverviewRepository) GetOverviewById(eventId int) (Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview WHERE event_number = ?`

	row := ov.db.QueryRow(query, eventId)

	var overview Overview

	// Scan row to return variable
	if err := row.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
		return overview, err
	}

	return overview, nil
}

// GetOverviewByCentralId retrieves an overview by the provided central ID from the database.
func (ov *OverviewRepository) GetOverviewByCentralId(centralId string) ([]Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview WHERE central_id = ?`

	rows, err := ov.db.Query(query, centralId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overviews []Overview

	// Scan rows to return slice
	for rows.Next() {
		var overview Overview
		if err := rows.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
			return nil, err
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// UpdateLevelByEventNumber updates the Level column of a specified EventNumber with the passed-in value.
func (ov *OverviewRepository) UpdateLevelByEventNumber(eventNumber int, newLevel Level) error {
	query := `UPDATE overview SET level = ? WHERE event_number = ?`

	_, err := ov.db.Exec(query, newLevel, eventNumber)
	if err != nil {
		return err
	}

	return nil
}

// GetOverviewByCentralIdAndEventNumber retrieves an overview by the provided central ID and event number from the database.
func (ov *OverviewRepository) GetOverviewByCentralIdAndEventNumber(centralId string, eventNumber int) ([]Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview WHERE central_id = ? AND event_number = ?`

	rows, err := ov.db.Query(query, centralId, eventNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overviews []Overview

	// Scan rows to return slice
	for rows.Next() {
		var overview Overview
		if err := rows.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
			return nil, err
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}
