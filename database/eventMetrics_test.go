package database

import (
	"reflect"
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

func TestNewEscalationLevelsFromData(t *testing.T) {
	tests := []struct {
		name string
		data map[int][]Level
		want *EscalationLevels
	}{
		{
			name: "Test with empty data",
			data: map[int][]Level{},
			want: NewEscalationLevels(),
		},
		{
			name: "Test with non-empty data",
			data: map[int][]Level{1: {Allarme, Emergenza}, 2: {Incidente}},
			want: &EscalationLevels{Levels: map[int]Level{1: Emergenza, 2: Incidente}},
		},
		{
			name: "Test with single event, multiple levels",
			data: map[int][]Level{5: {Emergenza, Incidente, Allarme}},
			want: &EscalationLevels{Levels: map[int]Level{5: Incidente}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEscalationLevelsFromData(tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEscalationLevelsFromData() = %v, want %v", got, tt.want)
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

func TestAdd(t *testing.T) {
	tests := []struct {
		name       string
		initData   map[int][]Level
		addEvent   int
		addLevel   Level
		outputData map[int][]Level
	}{
		{
			name:     "Initial Empty",
			initData: map[int][]Level{},
			addEvent: 1,
			addLevel: Allarme,
			outputData: map[int][]Level{
				1: {Allarme},
			},
		},
		{
			name: "Add To Existing",
			initData: map[int][]Level{
				1: {Allarme},
			},
			addEvent: 1,
			addLevel: Emergenza,
			outputData: map[int][]Level{
				1: {Emergenza},
			},
		},
		{
			name: "Add Lower",
			initData: map[int][]Level{
				1: {Incidente},
			},
			addEvent: 1,
			addLevel: Emergenza,
			outputData: map[int][]Level{
				1: {Incidente},
			},
		},
		{
			name: "Add New",
			initData: map[int][]Level{
				1: {Emergenza},
			},
			addEvent: 2,
			addLevel: Allarme,
			outputData: map[int][]Level{
				1: {Emergenza},
				2: {Allarme},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := NewEscalationLevelsFromData(tt.initData)
			el.Add(tt.addEvent, tt.addLevel)

			// Loop through all levels in tt.outputData
			for e, levels := range tt.outputData {
				for idx, l := range levels {
					if got, exist := el.Levels[e]; !exist || got != l {
						t.Errorf("Event %d: expected level %q at index %d, got level %q", e, l, idx, got)
					}
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
			// NewEscalationLevelsFromData() is assumed to be like NewEscalationLevels() only
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
			// NewEscalationLevelsFromData() is assumed to be like NewEscalationLevels() only
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
