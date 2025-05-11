// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

import (
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMergeTasks(t *testing.T) {
	tests := []struct {
		name     string
		original []Task
		update   []Task
		want     []Task
	}{
		{
			name:     "Empty Original and Update",
			original: []Task{},
			update:   []Task{},
			want:     []Task{},
		},
		{
			name: "No Update, Original Unchanged",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: "verde"},
			},
			update: []Task{},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: "verde"},
			},
		},
		{
			name: "Update with new tasks",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: "verde"},
			},
			update: []Task{
				{ID: 3, Title: "title3", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: "verde"},
				{ID: 3, Title: "title3", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
		},
		{
			name: "Update Existing Task",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: "verde"},
			},
			update: []Task{
				{ID: 2, Title: "title2", Category: "category2", Priority: 3, Description: "DescUpdated", Role: "RoleUpdated", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 3, Description: "DescUpdated", Role: "RoleUpdated", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
		},
		{
			name: "Update Causes Task Deletion",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: "verde"},
			},
			update: []Task{
				{ID: 2, Title: "title2", Category: "", Priority: 0, Description: "", Role: "", EscalationLevel: "", IncidentLevel: ""},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: "bianca"},
			},
		},
		// Rule 1: Multiple tasks with same title in original slice - keep only the one with higher escalation level
		{
			name: "Rule 1: Keep Task with Higher Escalation Level",
			original: []Task{
				{ID: 1, Title: "sameTitle", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "sameTitle", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "uniqueTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "allarme", IncidentLevel: ""},
			},
			update: []Task{},
			want: []Task{
				{ID: 2, Title: "sameTitle", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "uniqueTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "allarme", IncidentLevel: ""},
			},
		},
		// Rule 1 (edge case): Multiple tasks with same title and "incidente" escalation level - keep only the one with higher incident level
		{
			name: "Rule 1 Edge Case: Keep Task with Higher Incident Level",
			original: []Task{
				{ID: 1, Title: "sameTitle", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 2, Title: "sameTitle", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 3, Title: "uniqueTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			update: []Task{},
			want: []Task{
				{ID: 2, Title: "sameTitle", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 3, Title: "uniqueTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
		},
		// Rule 3 (edge case): Multiple tasks with same title in update slice - apply Rule 1 before appending
		{
			name: "Rule 3 Edge Case: Filter Update Tasks Before Appending",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
			},
			update: []Task{
				{ID: 2, Title: "newTitle", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 3, Title: "newTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 3, Title: "newTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
		},
		// Complex scenario: Combining multiple rules
		{
			name: "Complex Scenario: Combining Multiple Rules",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title1", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title2", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "incidente", IncidentLevel: "bianca"},
			},
			update: []Task{
				{ID: 4, Title: "title1", Category: "category4", Priority: 4, Description: "Desc4", Role: "Role4", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 5, Title: "title2", Category: "category5", Priority: 5, Description: "Desc5", Role: "Role5", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 6, Title: "title3", Category: "category6", Priority: 6, Description: "Desc6", Role: "Role6", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 7, Title: "title3", Category: "category7", Priority: 7, Description: "Desc7", Role: "Role7", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
			want: []Task{
				{ID: 4, Title: "title1", Category: "category4", Priority: 4, Description: "Desc4", Role: "Role4", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 5, Title: "title2", Category: "category5", Priority: 5, Description: "Desc5", Role: "Role5", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 7, Title: "title3", Category: "category7", Priority: 7, Description: "Desc7", Role: "Role7", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
		},
		// Rule 3 (edge case): Multiple tasks with same title and "incidente" escalation level in update slice
		{
			name: "Rule 3 Edge Case: Incident Level Comparison in Update Slice",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
			},
			update: []Task{
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 3, Title: "title2", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 4, Title: "title2", Category: "category4", Priority: 4, Description: "Desc4", Role: "Role4", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 4, Title: "title2", Category: "category4", Priority: 4, Description: "Desc4", Role: "Role4", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
		},
		// Deletion scenario with multiple tasks with the same title in the original slice
		{
			name: "Delete Task with Multiple Instances in Original",
			original: []Task{
				{ID: 1, Title: "sameTitle", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "sameTitle", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "uniqueTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "allarme", IncidentLevel: ""},
			},
			update: []Task{
				{ID: 4, Title: "sameTitle", Category: "", Priority: 0, Description: "", Role: "", EscalationLevel: "", IncidentLevel: ""},
			},
			want: []Task{
				{ID: 3, Title: "uniqueTitle", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "allarme", IncidentLevel: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeTasks(tt.original, tt.update)
			if err != nil {
				t.Errorf("MergeTasks() error = %v", err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("MergeTasks() length got = %v, want = %v", len(got), len(tt.want))
				return
			}

			// Create maps to check for task presence regardless of order
			gotMap := make(map[string]Task)
			wantMap := make(map[string]Task)

			for _, task := range got {
				gotMap[task.Title] = task
			}

			for _, task := range tt.want {
				wantMap[task.Title] = task
			}

			// Check that each expected task is in the result
			for title, wantTask := range wantMap {
				if gotTask, ok := gotMap[title]; !ok {
					t.Errorf("MergeTasks() missing expected task with title %s", title)
				} else if gotTask != wantTask {
					t.Errorf("MergeTasks() task with title %s got = %v, want = %v", title, gotTask, wantTask)
				}
			}
		})
	}
}

// Helper function to get the project root directory from the current directory
func getProjectRootDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Assuming the project folder name is "dogeplus-backend"
	for !strings.HasSuffix(dir, "dogeplus-backend") {
		dir = filepath.Dir(dir)
	}
	return dir, nil
}

func TestParseXLSXToTasks(t *testing.T) {
	projectRoot, err := getProjectRootDir()
	if err != nil {
		t.Fatalf("Failed to get the project root directory: %v", err)
	}

	tests := []struct {
		name    string
		file    string
		want    []Task
		wantErr bool
	}{
		{
			name:    "EmptyFile",
			file:    filepath.Join(projectRoot, "testdata", "empty.xlsx"),
			want:    nil,
			wantErr: true,
		},
		{
			name: "SingleTask",
			file: filepath.Join(projectRoot, "testdata", "single_task.xlsx"),
			want: []Task{
				{Category: "Sheet1", Role: "Medico", Priority: 1, Title: "Title1", Description: "Desc1", EscalationLevel: "el1", IncidentLevel: "il1"},
			},
			wantErr: false,
		},
		{
			name: "MultipleTasks",
			file: filepath.Join(projectRoot, "testdata", "multiple_tasks.xlsx"),
			want: []Task{
				{Category: "Sheet1", Role: "Medico", Priority: 1, Title: "Title1", Description: "Desc1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{Category: "Sheet1", Role: "RTT", Priority: 2, Title: "Title2", Description: "Desc2", EscalationLevel: "el2", IncidentLevel: "il2"},
			},
			wantErr: false,
		},
		{
			name: "MultipleTasks1stEmpty",
			file: filepath.Join(projectRoot, "testdata", "multiple_task_1st_empty.xlsx"),
			want: []Task{
				{Category: "Sheet1", Role: "Medico", Priority: 1, Title: "Title1", Description: "Desc1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{Category: "Sheet1", Role: "RTT", Priority: 2, Title: "Title2", Description: "Desc2", EscalationLevel: "el2", IncidentLevel: "il2"},
				{Category: "Sheet1", Role: "RTT", Priority: 3, Title: "Title3", Description: "Desc3", EscalationLevel: "el3", IncidentLevel: "il3"},
			},
			wantErr: false,
		},
		{
			name: "MultipleTasks2ndEmpty",
			file: filepath.Join(projectRoot, "testdata", "multiple_task_2nd_empty.xlsx"),
			want: []Task{
				{Category: "Sheet1", Role: "Medico", Priority: 1, Title: "Title1", Description: "Desc1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{Category: "Sheet1", Role: "RTT", Priority: 2, Title: "Title2", Description: "Desc2", EscalationLevel: "el2", IncidentLevel: "il2"},
				{Category: "Sheet1", Role: "Medico", Priority: 3, Title: "Title3", Description: "Desc3", EscalationLevel: "el3", IncidentLevel: "il3"},
			},
			wantErr: false,
		},
		{
			name: "MultipleTasksLastEmpty",
			file: filepath.Join(projectRoot, "testdata", "multiple_tasks_last_empty.xlsx"),
			want: []Task{
				{Category: "Sheet1", Role: "Medico", Priority: 1, Title: "Title1", Description: "Desc1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{Category: "Sheet1", Role: "RTT", Priority: 2, Title: "Title2", Description: "Desc2", EscalationLevel: "el2", IncidentLevel: ""},
			},
			wantErr: false,
		},
		{
			name:    "NonExistentFile",
			file:    filepath.Join(projectRoot, "testdata", "non_existent.xlsx"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := excelize.OpenFile(tt.file)
			if os.IsNotExist(err) && tt.wantErr {
				// if file doesn't exist as expected, skip the test
				t.Skip("Skipping test due to missing file.")
			} else if err != nil && !tt.wantErr {
				t.Errorf("OpenFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := ParseXLSXToTasks(f)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseXLSXToTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseXLSXToTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterTasksForEscalation(t *testing.T) {
	tests := []struct {
		name            string
		tasks           []Task
		category        string
		startingEscal   string
		finalEscalation string
		incidentLevel   string
		want            []Task
		wantErr         bool
	}{
		{
			name: "ValidFilter",
			tasks: []Task{
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: ""},
			},
			category:        "cat1",
			startingEscal:   "allarme",
			finalEscalation: "emergenza",
			incidentLevel:   "",
			want: []Task{
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
			wantErr: false,
		},
		{
			name: "ValidFilterIncidenteCaseToBianca",
			tasks: []Task{
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			category:        "cat1",
			startingEscal:   "allarme",
			finalEscalation: "incidente",
			incidentLevel:   "bianca",
			want: []Task{
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
			},
			wantErr: false,
		},
		{
			name: "ValidFilterIncidenteCaseToBiancaWithDefaultCategory",
			tasks: []Task{
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{Category: "pro22", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			category:        "cat1",
			startingEscal:   "allarme",
			finalEscalation: "incidente",
			incidentLevel:   "bianca",
			want: []Task{
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{Category: "pro22", EscalationLevel: "incidente", IncidentLevel: "bianca"},
			},
			wantErr: false,
		},
		{
			name: "ValidFilterIncidenteCaseToVerde",
			tasks: []Task{
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
			category:        "cat1",
			startingEscal:   "allarme",
			finalEscalation: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
			wantErr: false,
		},
		{
			name: "InvalidStartingAndFinalLevels",
			tasks: []Task{
				{Category: "cat1", EscalationLevel: "level1", IncidentLevel: "green"},
				{Category: "cat1", EscalationLevel: "level2", IncidentLevel: "green"},
				{Category: "cat1", EscalationLevel: "level3", IncidentLevel: "green"},
			},
			category:        "cat1",
			startingEscal:   "level4",
			finalEscalation: "level5",
			incidentLevel:   "",
			want:            nil,
			wantErr:         true,
		},
		// ... other test cases ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterTasksForEscalation(tt.tasks, tt.category, tt.startingEscal, tt.finalEscalation, tt.incidentLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterTasksForEscalation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterTasksForEscalation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterTasks(t *testing.T) {
	tests := []struct {
		name            string
		tasks           []Task
		category        string
		escalationLevel string
		incidentLevel   string
		want            []Task
	}{
		{
			name: "All Tasks Matched",
			tasks: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
			},
			category:        "cat1",
			escalationLevel: "allarme",
			incidentLevel:   "",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
			},
		},
		{
			name: "Some Tasks Matched",
			tasks: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat2", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
			category:        "cat1",
			escalationLevel: "allarme",
			incidentLevel:   "",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
			},
		},
		{
			name: "Higher escalation level",
			tasks: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
			category:        "cat1",
			escalationLevel: "emergenza",
			incidentLevel:   "",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
			},
		},
		{
			name: "Up to incidente, single IncidentLevel",
			tasks: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "bianca",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
			},
		},
		{
			name: "Up to incidente, multiple IncidentLevel",
			tasks: []Task{
				{ID: 1, Priority: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Priority: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Priority: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Priority: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{ID: 1, Priority: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Priority: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Priority: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Priority: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
		},
		{
			name: "Up to incidente, multiple IncidentLevel, some IncidentLevel not matched",
			tasks: []Task{
				{ID: 1, Priority: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Priority: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Priority: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Priority: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 5, Priority: 5, Title: "title5", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{ID: 1, Priority: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Priority: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Priority: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Priority: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
		},
		{
			name: "Up to incidente, multiple IncidentLevel, some IncidentLevel not matched, some Category not matched",
			tasks: []Task{
				{ID: 1, Priority: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Priority: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Priority: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Priority: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 6, Priority: 6, Title: "title6", Category: "cat2", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 5, Priority: 5, Title: "title5", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{ID: 1, Priority: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Priority: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Priority: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Priority: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
		},
		// ... other test cases ...
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterTasks(tt.tasks, tt.category, tt.escalationLevel, tt.incidentLevel)
			if reflect.DeepEqual(got, tt.want) == false {
				t.Errorf("FilterTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}
