// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

import (
	"database/sql"
	"dogeplus-backend/errors"
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
	if err != nil {
		return errors.Wrap(err, "failed to add overview")
	}

	return nil
}

// GetAllOverview retrieves all overview records from the database and returns a slice of Overview or an error if it fails.
func (ov *OverviewRepository) GetAllOverview() ([]Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview`

	rows, err := ov.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query overviews")
	}
	defer func() {
		errors.HandleCloser(rows.Close(), "error closing rows in GetAllOverview")
	}()

	var overviews []Overview

	// Scan rows to return slice
	for rows.Next() {
		var overview Overview
		if err := rows.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
			return nil, errors.Wrap(err, "failed to scan overview row")
		}
		overviews = append(overviews, overview)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error during row iteration")
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
		return overview, errors.Wrap(err, "failed to scan overview row by ID")
	}

	return overview, nil
}

// GetOverviewByCentralId retrieves an overview by the provided central ID from the database.
func (ov *OverviewRepository) GetOverviewByCentralId(centralId string) ([]Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview WHERE central_id = ?`

	rows, err := ov.db.Query(query, centralId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query overviews by central ID")
	}
	defer func() {
		errors.HandleCloser(rows.Close(), "error closing rows in GetOverviewByCentralId")
	}()

	var overviews []Overview

	// Scan rows to return slice
	for rows.Next() {
		var overview Overview
		if err := rows.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
			return nil, errors.Wrap(err, "failed to scan overview row")
		}
		overviews = append(overviews, overview)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error during row iteration")
	}

	return overviews, nil
}

// UpdateLevelByEventNumber updates the level and incident level of an overview record based on the provided event number.
func (ov *OverviewRepository) UpdateLevelByEventNumber(eventNumber int, newLevel Level, incidentLevel string) error {
	query := `UPDATE overview SET level = ?, incident_level = ? WHERE event_number = ?`

	_, err := ov.db.Exec(query, newLevel, incidentLevel, eventNumber)
	if err != nil {
		return errors.Wrap(err, "failed to update level by event number")
	}

	return nil
}

// GetOverviewByCentralIdAndEventNumber retrieves an overview by the provided central ID and event number from the database.
func (ov *OverviewRepository) GetOverviewByCentralIdAndEventNumber(centralId string, eventNumber int) ([]Overview, error) {
	query := `SELECT uuid, central_id, event_number, location, location_detail, type, level, incident_level FROM overview WHERE central_id = ? AND event_number = ?`

	rows, err := ov.db.Query(query, centralId, eventNumber)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query overviews by central ID and event number")
	}
	defer func() {
		errors.HandleCloser(rows.Close(), "error closing rows in GetOverviewByCentralIdAndEventNumber")
	}()

	var overviews []Overview

	// Scan rows to return slice
	for rows.Next() {
		var overview Overview
		if err := rows.Scan(&overview.UUID, &overview.CentralId, &overview.EventNumber, &overview.Location, &overview.LocationDetail, &overview.Type, &overview.Level, &overview.IncidentLevel); err != nil {
			return nil, errors.Wrap(err, "failed to scan overview row")
		}
		overviews = append(overviews, overview)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error during row iteration")
	}

	return overviews, nil
}
