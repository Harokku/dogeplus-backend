// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"strconv"
	"strings"
)

// parsePriority converts a priority string to an int, returns 0 if invalid.
// It logs a warning if the conversion fails but does not halt execution.
func parsePriority(priority string) int {
	priorityInt, err := strconv.Atoi(priority)
	if err != nil {
		log.Printf("failed to parse priority: %v", err)
		return 0
	}
	return priorityInt
}

// isBlockEmpty checks if a block of 5 columns is empty
func isBlockEmpty(block []string) bool {
	for _, cell := range block {
		if cell != "" {
			return false
		}
	}
	return true
}

// padBlock ensures a block has exactly 'blockSize' columns by appending empty strings if necessary.
func padBlock(block []string, blockSize int) []string {
	if len(block) >= blockSize {
		return block
	}
	paddedBlock := make([]string, blockSize)
	copy(paddedBlock, block)
	return paddedBlock
}

// ParseXLSXToTasks converts an Excel file into a slice of Task instances, parsing data from each sheet and handling errors.
func ParseXLSXToTasks(f *excelize.File) ([]Task, error) {
	var tasks []Task

	// Check if the file has any sheets
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("the file does not contain any sheets")
	}

	// Iterate over each sheet in the file
	for _, sheetName := range sheetList {
		// Get all the rows from the current sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to get rows from sheet %s: %v", sheetName, err)
		}

		// Fetch the header row roles
		if len(rows) < 3 {
			return nil, fmt.Errorf("the sheet %s does not have the required structure (at least 3 rows needed)", sheetName)
		}

		headerRow := rows[0]

		// Iterate over each row in the sheet, skipping the first 2 rows
		for i, row := range rows {
			if i < 2 {
				continue // Skip the first 2 rows
			}

			// Iterate in blocks of 5 columns starting from the first block
			for j := 0; j < len(row); j += 5 {
				// Ensure there's a block to process
				block := row[j:min(j+5, len(row))]

				// Pad the block to ensure it has exactly 5 columns
				block = padBlock(block, 5)

				// Skip if the block is empty
				if isBlockEmpty(block) {
					continue
				}

				// Fetch the role from the header row based on the block's starting column
				role := ""
				if j < len(headerRow) {
					role = headerRow[j]
				}

				// Create a new Task struct with mapped fields
				task := Task{
					Category:        sheetName,
					Role:            role,
					Priority:        parsePriority(block[0]),   // Convert and map the priority
					Title:           block[1],                  // Map the title field
					Description:     block[2],                  // Map the description field
					EscalationLevel: strings.ToLower(block[3]), // Map the escalation level field
					IncidentLevel:   strings.ToLower(block[4]), // Map the incident level field
				}

				// Append the new task to the tasks slice
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}
