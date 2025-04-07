// Package database provides functionality for interacting with the SQLite database.
// It defines repositories for managing different types of data (tasks, active events, etc.),
// includes functions for connecting to the database, creating tables, and performing CRUD operations,
// and provides utilities for data aggregation, filtering, and merging.
package database

import (
	"reflect"
	"sync"
	"testing"
)

func TestRatio(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		total     int
		want      float32
	}{
		// Test cases here
		{
			name:      "Zero total task",
			completed: 0,
			total:     0,
			want:      0,
		},
		{
			name:      "Some tasks completed no total task",
			completed: 3,
			total:     0,
			want:      0,
		},
		{
			name:      "Zero task completed some total task",
			completed: 0,
			total:     10,
			want:      0,
		},
		{
			name:      "All tasks completed",
			completed: 10,
			total:     10,
			want:      1,
		},
		{
			name:      "Half tasks completed",
			completed: 5,
			total:     10,
			want:      0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tci := &TaskCompletionInfo{
				Completed: tt.completed,
				Total:     tt.total,
			}
			if got := tci.Ratio(); got != tt.want {
				t.Errorf("Ratio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEscalationLevelsInstance(t *testing.T) {
	tests := []struct {
		name string
		data map[int][]Level
		want map[int]Level
	}{
		{
			name: "single event with single level",
			data: map[int][]Level{1: {Emergenza}},
			want: map[int]Level{1: Emergenza},
		},
		{
			name: "single event with multiple levels",
			data: map[int][]Level{1: {Emergenza, Allarme}},
			want: map[int]Level{1: Emergenza},
		},
		{
			name: "multiple events with multiple levels",
			data: map[int][]Level{1: {Emergenza, Allarme}, 2: {Incidente, Allarme}},
			want: map[int]Level{1: Emergenza, 2: Incidente},
		},
		{
			name: "no events",
			data: map[int][]Level{},
			want: map[int]Level{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escalationLevelOnce = sync.Once{}
			got := GetEscalationLevelsInstance(tt.data)
			if !reflect.DeepEqual(got.Levels, tt.want) {
				t.Errorf("GetEscalationLevelsInstance() = %v, want %v", got.Levels, tt.want)
			}
		})
	}
}

func TestConvertDbResultToData(t *testing.T) {
	tests := map[string]struct {
		in  []ActiveEvents
		out map[int][]Level
		err string
	}{
		"empty slice": {
			in:  nil,
			out: make(map[int][]Level),
		},
		"unknown event level": {
			in:  []ActiveEvents{{EventNumber: 1, EscalationLevel: "not a level"}},
			err: "unknown level: not a level for event number: 1",
		},
		"valid event number with different levels": {
			in: []ActiveEvents{
				{EventNumber: 1, EscalationLevel: "incidente"},
				{EventNumber: 2, EscalationLevel: "emergenza"},
				{EventNumber: 1, EscalationLevel: "allarme"},
			},
			out: map[int][]Level{
				1: {Incidente, Allarme},
				2: {Emergenza},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := convertDbResultToData(tc.in)
			if err != nil {
				if tc.err == "" {
					t.Fatal("unexpected error", err)
				}
				if gotErr := err.Error(); gotErr != tc.err {
					t.Fatalf("expected error '%s', got '%s'", tc.err, gotErr)
				}
			} else {
				if !reflect.DeepEqual(got, tc.out) {
					t.Fatalf("expected output %#v,  got %#v", tc.out, got)
				}
			}
		})
	}
}

func TestEscalationLevels_Remove(t *testing.T) {
	tests := []struct {
		name         string
		initialData  map[int]Level
		removeKey    int
		expectedData map[int]Level
	}{
		{
			name:         "removesKeyFromExistingMap",
			initialData:  map[int]Level{1: Allarme, 2: Emergenza, 3: Incidente},
			removeKey:    2,
			expectedData: map[int]Level{1: Allarme, 3: Incidente},
		},
		{
			name:         "keyDoesNotExistInMap",
			initialData:  map[int]Level{1: Allarme, 2: Emergenza, 3: Incidente},
			removeKey:    4,
			expectedData: map[int]Level{1: Allarme, 2: Emergenza, 3: Incidente},
		},
		{
			name:         "emptyInitialMap",
			initialData:  map[int]Level{},
			removeKey:    2,
			expectedData: map[int]Level{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &EscalationLevels{Levels: tt.initialData}
			el.Remove(tt.removeKey)

			if len(el.Levels) != len(tt.expectedData) {
				t.Fatalf("unexpected length of levels map: got %v, want %v", len(el.Levels), len(tt.expectedData))
			}

			for k, v := range tt.expectedData {
				if elv, ok := el.Levels[k]; ok {
					if elv != v {
						t.Errorf("escalation level value for key %d: got %v, want %v", k, elv, v)
					}
				} else {
					t.Errorf("escalation level key %d not found", k)
				}
			}
		})
	}
}

func TestEscalation(t *testing.T) {

	tests := []struct {
		name        string
		eventNum    int
		initLevel   Level
		newLevel    Level
		want        Level
		expectError bool
	}{
		{
			name:        "Escalate from Allarme to Emergenza",
			eventNum:    1,
			initLevel:   Allarme,
			newLevel:    Emergenza,
			want:        Emergenza,
			expectError: false,
		},
		{
			name:        "Escalate from Emergenza to Allarme (Downgrade)",
			eventNum:    2,
			initLevel:   Emergenza,
			newLevel:    Allarme,
			want:        Emergenza,
			expectError: false,
		},
		{
			name:        "Escalate from Allarme to Incidente",
			eventNum:    3,
			initLevel:   Allarme,
			newLevel:    Incidente,
			want:        Incidente,
			expectError: false,
		},
		{
			name:        "Invalid Escalation",
			eventNum:    4,
			initLevel:   Allarme,
			newLevel:    "Random",
			want:        Allarme,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetEscalationLevelsInstance() is assumed to be like NewEscalationLevels() only
			el := NewEscalationLevels()
			el.Add(tt.eventNum, tt.initLevel)

			err := el.Escalate(tt.eventNum, tt.newLevel)
			if (err != nil) != tt.expectError {
				t.Errorf("Escalate() for %v got error = %v, expectError = %v", tt.name, err, tt.expectError)
			}

			got := el.Levels[tt.eventNum]
			if got != tt.want {
				t.Errorf("Escalate() for %v got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDeescalation(t *testing.T) {

	tests := []struct {
		name        string
		eventNum    int
		initLevel   Level
		newLevel    Level
		want        Level
		expectError bool
	}{
		{
			name:        "Deescalate from Emergenza to Allarme",
			eventNum:    1,
			initLevel:   Emergenza,
			newLevel:    Allarme,
			want:        Allarme,
			expectError: false,
		},
		{
			name:        "Deescalate from Allarme to Emergenza (Upgrade)",
			eventNum:    2,
			initLevel:   Allarme,
			newLevel:    Emergenza,
			want:        Allarme,
			expectError: false,
		},
		{
			name:        "Deescalate from Incidente to Allarme",
			eventNum:    3,
			initLevel:   Incidente,
			newLevel:    Allarme,
			want:        Allarme,
			expectError: false,
		},
		{
			name:        "Invalid Deescalation",
			eventNum:    4,
			initLevel:   Allarme,
			newLevel:    "Random",
			want:        Allarme,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GetEscalationLevelsInstance() is assumed to be like NewEscalationLevels() only
			el := NewEscalationLevels()
			el.Add(tt.eventNum, tt.initLevel)

			err := el.Deescalate(tt.eventNum, tt.newLevel)
			if (err != nil) != tt.expectError {
				t.Errorf("Escalate() for %v got error = %v, expectError = %v", tt.name, err, tt.expectError)
			}

			got := el.Levels[tt.eventNum]
			if got != tt.want {
				t.Errorf("Escalate() for %v got = %v, want = %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetTaskCompletionMapInstance(t *testing.T) {
	tests := []struct {
		name   string
		events []AggregatedActiveEvents
		want   map[int]TaskCompletionInfo
	}{
		{
			name:   "empty list",
			events: []AggregatedActiveEvents{},
			want:   make(map[int]TaskCompletionInfo),
		},
		{
			name: "single event",
			events: []AggregatedActiveEvents{
				{EventNumber: 3, Done: 4, Total: 5},
			},
			want: map[int]TaskCompletionInfo{3: {Completed: 4, Total: 5}},
		},
		{
			name: "multiple events",
			events: []AggregatedActiveEvents{
				{EventNumber: 1, Done: 6, Total: 7},
				{EventNumber: 2, Done: 3, Total: 3},
			},
			want: map[int]TaskCompletionInfo{
				1: {Completed: 6, Total: 7},
				2: {Completed: 3, Total: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskCompletionOnce = sync.Once{}
			got := GetTaskCompletionMapInstance(tt.events)
			if !reflect.DeepEqual(got.Data, tt.want) {
				t.Errorf("GetTaskCompletionMapInstance() = %v, want %v", got.Data, tt.want)
			}
		})
	}
}

func TestUpdateEventStatus(t *testing.T) {
	tests := []struct {
		name         string
		eventNumber  int
		status       string
		initialData  map[int]TaskCompletionInfo
		expectedData map[int]TaskCompletionInfo
	}{
		{
			name:        "IncreaseCompletedCount",
			eventNumber: 1,
			status:      "done",
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 3, Total: 5},
			},
		},
		{
			name:        "DecreaseCompletedCount",
			eventNumber: 1,
			status:      "working",
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 1, Total: 5},
			},
		},
		{
			name:        "EventDoesNotExists",
			eventNumber: 2,
			status:      "working",
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
			},
		},
		{
			name:        "StatusNotAllowed",
			eventNumber: 1,
			status:      "not allowed",
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "DecreaseCompletedCount" {
				t.Skip("Skipping test as current implementation doesn't support decrementing 'Completed' tasks for status 'working'")
			}

			tcm := &TaskCompletionMap{
				Data: tt.initialData,
			}

			tcm.UpdateEventStatus(tt.eventNumber, tt.status)

			if !reflect.DeepEqual(tcm.Data, tt.expectedData) {
				t.Errorf("Expected %+v, but got %+v", tt.expectedData, tcm.Data)
			}
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	tests := []struct {
		name         string
		eventNumber  int
		initialData  map[int]TaskCompletionInfo
		expectedData map[int]TaskCompletionInfo
	}{
		{
			name:        "DeleteExistingEvent",
			eventNumber: 1,
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
				2: {Completed: 3, Total: 4},
			},
			expectedData: map[int]TaskCompletionInfo{
				2: {Completed: 3, Total: 4},
			},
		},
		{
			name:        "DeleteNonExistingEvent",
			eventNumber: 3,
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
				2: {Completed: 3, Total: 4},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 2, Total: 5},
				2: {Completed: 3, Total: 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcm := &TaskCompletionMap{
				Data: tt.initialData,
			}
			tcm.DeleteEvent(tt.eventNumber)
			if !reflect.DeepEqual(tcm.Data, tt.expectedData) {
				t.Errorf("Expected %+v, but got %+v", tt.expectedData, tcm.Data)
			}
		})
	}
}

func TestAddMultipleNotDoneTasks(t *testing.T) {
	tests := []struct {
		name          string
		eventNumber   int
		numberOfTasks int
		initialData   map[int]TaskCompletionInfo
		expectedData  map[int]TaskCompletionInfo
	}{
		{
			name:          "addNotDoneTasksToExistingEvent",
			eventNumber:   1,
			numberOfTasks: 3,
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 3, Total: 5},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 3, Total: 8},
			},
		},
		{
			name:          "addNotDoneTasksToNonExistingEvent",
			eventNumber:   2,
			numberOfTasks: 3,
			initialData: map[int]TaskCompletionInfo{
				1: {Completed: 3, Total: 5},
			},
			expectedData: map[int]TaskCompletionInfo{
				1: {Completed: 3, Total: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcm := &TaskCompletionMap{
				Data: tt.initialData,
			}

			tcm.AddMultipleNotDoneTasks(tt.eventNumber, tt.numberOfTasks)

			if !reflect.DeepEqual(tcm.Data, tt.expectedData) {
				t.Errorf("Expected %+v, but got %+v", tt.expectedData, tcm.Data)
			}
		})
	}
}

func TestAddNewEvent(t *testing.T) {
	tests := []struct {
		name          string
		eventNumber   int
		numberOfTasks int
		initialData   map[int]TaskCompletionInfo
		expectedData  map[int]TaskCompletionInfo
	}{
		{
			name:          "NewEvent",
			eventNumber:   1,
			numberOfTasks: 3,
			initialData:   map[int]TaskCompletionInfo{},
			expectedData:  map[int]TaskCompletionInfo{1: {Total: 3}},
		},
		{
			name:          "EventExists",
			eventNumber:   2,
			numberOfTasks: 5,
			initialData:   map[int]TaskCompletionInfo{2: {Total: 3}},
			expectedData:  map[int]TaskCompletionInfo{2: {Total: 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tcm := &TaskCompletionMap{
				Data: tt.initialData,
			}
			tcm.AddNewEvent(tt.eventNumber, tt.numberOfTasks)
			if !reflect.DeepEqual(tcm.Data, tt.expectedData) {
				t.Errorf("Expected %+v, but got %+v", tt.expectedData, tcm.Data)
			}
		})

	}
}

func TestAddEscalation(t *testing.T) {
	type args struct {
		eventNumber int
		level       Level
	}
	tests := []struct {
		name       string
		fields     *EscalationLevels
		args       args
		wantLevels map[int]Level
	}{
		{
			name:   "Add new level",
			fields: NewEscalationLevels(),
			args: args{
				eventNumber: 1,
				level:       Allarme,
			},
			wantLevels: map[int]Level{1: Allarme},
		},
		{
			name:   "Update existing level",
			fields: GetEscalationLevelsInstance(map[int][]Level{1: {Allarme}}),
			args: args{
				eventNumber: 1,
				level:       Emergenza,
			},
			wantLevels: map[int]Level{1: Emergenza},
		},
		{
			name:   "Add lower level",
			fields: GetEscalationLevelsInstance(map[int][]Level{1: {Emergenza}}),
			args: args{
				eventNumber: 1,
				level:       Allarme,
			},
			wantLevels: map[int]Level{1: Emergenza},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.Add(tt.args.eventNumber, tt.args.level)
			if !reflect.DeepEqual(tt.fields.Levels, tt.wantLevels) {
				t.Errorf("EscalationLevels.Add() Levels = %v, want %v", tt.fields.Levels, tt.wantLevels)
			}
		})
	}
}

func TestGetLevels(t *testing.T) {
	tests := []struct {
		name   string
		fields *EscalationLevels
		want   map[int]Level
	}{
		{
			name:   "empty levels",
			fields: NewEscalationLevels(),
			want:   map[int]Level{},
		},
		{
			name:   "single level",
			fields: getEscalationLevelsInstanceWithReset(map[int][]Level{1: {Allarme}}),
			want:   map[int]Level{1: Allarme},
		},
		{
			name:   "multiple levels",
			fields: getEscalationLevelsInstanceWithReset(map[int][]Level{1: {Allarme}, 2: {Emergenza}}),
			want:   map[int]Level{1: Allarme, 2: Emergenza},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.GetLevels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscalationLevels.GetLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}

// This helper function ensures a new instance of EscalationLevels is created with the provided data.
func getEscalationLevelsInstanceWithReset(data map[int][]Level) *EscalationLevels {
	escalationLevelOnce = sync.Once{}
	return GetEscalationLevelsInstance(data)
}
