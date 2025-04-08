// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

import (
	"database/sql"
	"dogeplus-backend/config"
	"github.com/gofiber/fiber/v2/log"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
)

// Singleton instance
var (
	instance *sql.DB
	once     sync.Once
)

// retry function tries to execute the provided function fn up to maxRetries times, waiting delay between attempts.
func retry(maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}

// GetInstance returns a singleton instance of *sql.DB and an error.
// If the instance has already been created, it returns the existing one.
// If the instance hasn't been created yet, it creates a new one using the specified database driver and connection string.
func GetInstance(configFile config.Config) (*sql.DB, error) {
	var initErr error
	db := config.GetEnvWithFallback(configFile, config.DbFile)

	once.Do(func() {
		// Check if db file is sanitized
		if err := config.SanitizeFilePath(db); err != nil {
			initErr = err
			return
		}

		// Retry logic for db connection
		initErr = retry(5, 10*time.Second, func() error {
			var err error
			instance, err = sql.Open("sqlite3", db)
			if err != nil {
				return err
			}

			// Ping db to check connection
			if err = instance.Ping(); err != nil {
				return err
			}

			log.Info("Db connection established")

			// Create table structure if not already exist
			if err = createTables(instance); err != nil {
				return err
			}

			log.Info("Db table structure created")
			return nil
		})
	})

	return instance, initErr
}

// createTables creates the necessary tables in the provided *sql.DB instance.
// It takes a transactional *sql.DB instance as an input parameter.
// The function begins a transaction, executes the table creation commands,
// and commits the transaction if all commands were successful.
//
// The table creation commands are defined in the `commands` slice and are executed one by one.
// If any error occurs while executing a command, the transaction is rolled back and the error is returned.
//
// Upon successful execution of all commands, the transaction is committed and nil is returned.
//
// Note: This function assumes that the provided *sql.DB instance is already connected to the database.
// The function does not check the connection status of the *sql.DB instance.
func createTables(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Define table creation commands
	commands := []string{
		`CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, data TEXT)`,

		// Tasks table
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			priority INTEGER, 
			title TEXT, 
			description TEXT, 
			role TEXT, 
			category TEXT,
			escalation_level TEXT CHECK ( escalation_level IN ('allarme','emergenza','incidente')),
			incident_level TEXT CHECK ( incident_level IN ('','bianca', 'verde', 'gialla', 'rossa')))`,
		// No trigger for task table

		// Active events table
		`CREATE TABLE IF NOT EXISTS active_events (
			uuid TEXT PRIMARY KEY,
			event_number INTEGER, 
			event_date TEXT NOT NULL,
			central_id TEXT CHECK ( central_id IN ('HQ','SRA','SRL','SRM','SRP')),
			priority INTEGER, 
			title TEXT,
			description TEXT,
			role TEXT,
			status TEXT CHECK ( status IN ('notdone','working','done')), 
			modified_by TEXT, 
			ip_address TEXT DEFAULT '0.0.0.0',
			timestamp TEXT,
			escalation_level TEXT CHECK (escalation_level in ('allarme', 'emergenza', 'incidente')))`,

		// Overview table
		`create table IF NOT EXISTS overview(
			uuid            text    not null
				constraint overview_pk
				primary key,
			central_id      text    not null,
			event_number    integer not null
				constraint event_number_unique_ck
				unique,
			location        text    not null,
			location_detail text,
			type            text    not null,
			level			text 	not null,
			incident_level	text)`,

		// Escalation levels definition table
		`create table IF NOT EXISTS escalation_levels(
			uuid        TEXT not null,
			name        text not null,
			description text,
				constraint escalation_levels_pk
				primary key (uuid));`,
	}

	// Execute each command within the transaction
	for _, command := range commands {
		_, err := tx.Exec(command)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Repositories represents a collection of different repositories for managing tasks and active events.
type Repositories struct {
	Tasks                       *TaskRepository
	ActiveEvents                *ActiveEventsRepository
	Overview                    *OverviewRepository
	TaskCompletionAggregation   *TaskCompletionMap
	EscalationLevelsAggregation *EscalationLevels
	EscalationLevelsDefinition  *EscalationLevelsDefinitionRepository
}

// NewRepositories initializes a new instance of Repositories with the provided *sql.DB object.
// It returns a pointer to the created Repositories.
func NewRepositories(db *sql.DB) *Repositories {
	repos := &Repositories{
		Tasks:                      NewTaskRepository(db),
		ActiveEvents:               NewActiveEventRepository(db),
		Overview:                   NewOverviewRepository(db),
		EscalationLevelsDefinition: NewEscalationLevelsDefinitionRepository(db),
	}

	// initialize aggregation map using data from db trough repos
	initialTaskAggregation, err := repos.ActiveEvents.GetAggregatedEventStatus()
	if err != nil {
		log.Fatal(err)
	}

	// Get raw escalation levels from db
	initialRawEscalationLevel, err := repos.ActiveEvents.GetRawEscalationLevels()
	if err != nil {
		log.Fatal(err)
	}

	// Convert db data to aggregation raw format
	convertedEscalationData, err := convertDbResultToData(initialRawEscalationLevel)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize actual aggregation keeping only higher escalation level
	initialEscalationLevelsAggregation := GetEscalationLevelsInstance(convertedEscalationData)

	// initialize task aggregation repo
	repos.TaskCompletionAggregation = GetTaskCompletionMapInstance(initialTaskAggregation, nil)
	repos.EscalationLevelsAggregation = initialEscalationLevelsAggregation

	return repos
}
