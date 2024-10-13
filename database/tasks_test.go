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
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "el2", IncidentLevel: "il2"},
			},
			update: []Task{},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "el2", IncidentLevel: "il2"},
			},
		},
		{
			name: "Update with new tasks",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "el2", IncidentLevel: "il2"},
			},
			update: []Task{
				{ID: 3, Title: "title3", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "el3", IncidentLevel: "il3"},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "el2", IncidentLevel: "il2"},
				{ID: 3, Title: "title3", Category: "category3", Priority: 3, Description: "Desc3", Role: "Role3", EscalationLevel: "el3", IncidentLevel: "il3"},
			},
		},
		{
			name: "Update Existing Task",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "el2", IncidentLevel: "il2"},
			},
			update: []Task{
				{ID: 2, Title: "title2", Category: "category2", Priority: 3, Description: "DescUpdated", Role: "RoleUpdated", EscalationLevel: "elUpdated", IncidentLevel: "ilUpdated"},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 3, Description: "DescUpdated", Role: "RoleUpdated", EscalationLevel: "elUpdated", IncidentLevel: "ilUpdated"},
			},
		},
		{
			name: "Update Causes Task Deletion",
			original: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
				{ID: 2, Title: "title2", Category: "category2", Priority: 2, Description: "Desc2", Role: "Role2", EscalationLevel: "el2", IncidentLevel: "il2"},
			},
			update: []Task{
				{ID: 2, Title: "title2", Category: "category2", Priority: 0, Description: "", Role: "", EscalationLevel: "", IncidentLevel: ""},
			},
			want: []Task{
				{ID: 1, Title: "title1", Category: "category1", Priority: 1, Description: "Desc1", Role: "Role1", EscalationLevel: "el1", IncidentLevel: "il1"},
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
			for i, task := range got {
				if task != tt.want[i] {
					t.Errorf("MergeTasks() task %d got = %v, want = %v", i, task, tt.want[i])
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
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
		},
		{
			name: "Up to incidente, multiple IncidentLevel, some IncidentLevel not matched",
			tasks: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 5, Title: "title5", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
			},
		},
		{
			name: "Up to incidente, multiple IncidentLevel, some IncidentLevel not matched, some Category not matched",
			tasks: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 6, Title: "title6", Category: "cat2", EscalationLevel: "incidente", IncidentLevel: "verde"},
				{ID: 5, Title: "title5", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "gialla"},
			},
			category:        "cat1",
			escalationLevel: "incidente",
			incidentLevel:   "verde",
			want: []Task{
				{ID: 1, Title: "title1", Category: "cat1", EscalationLevel: "allarme", IncidentLevel: ""},
				{ID: 2, Title: "title2", Category: "cat1", EscalationLevel: "emergenza", IncidentLevel: ""},
				{ID: 3, Title: "title3", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "bianca"},
				{ID: 4, Title: "title4", Category: "cat1", EscalationLevel: "incidente", IncidentLevel: "verde"},
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
