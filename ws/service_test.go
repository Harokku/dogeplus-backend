package ws

import (
	"reflect"
	"testing"
)

func TestService_Subscribe(t *testing.T) {
	tests := []struct {
		name          string
		initialTopics []string
		topicToAdd    string
		wantTopics    []string
		wantErr       bool
	}{
		{
			name:          "Subscribe to new topic",
			initialTopics: []string{},
			topicToAdd:    "event_updates",
			wantTopics:    []string{"event_updates"},
			wantErr:       false,
		},
		{
			name:          "Subscribe to existing topic",
			initialTopics: []string{"event_updates"},
			topicToAdd:    "event_updates",
			wantTopics:    []string{"event_updates"},
			wantErr:       false,
		},
		{
			name:          "Subscribe to second topic",
			initialTopics: []string{"event_updates"},
			topicToAdd:    "central_ABC123",
			wantTopics:    []string{"event_updates", "central_ABC123"},
			wantErr:       false,
		},
		{
			name:          "Subscribe with empty topic name",
			initialTopics: []string{"event_updates"},
			topicToAdd:    "",
			wantTopics:    []string{"event_updates", ""},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService("test-client")

			// Set initial topics
			s.topics = make([]string, len(tt.initialTopics))
			copy(s.topics, tt.initialTopics)

			// Call the method being tested
			err := s.Subscribe(tt.topicToAdd)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check topics
			if !reflect.DeepEqual(s.GetTopics(), tt.wantTopics) {
				t.Errorf("After Subscribe() topics = %v, want %v", s.GetTopics(), tt.wantTopics)
			}
		})
	}
}

func TestService_Unsubscribe(t *testing.T) {
	tests := []struct {
		name          string
		initialTopics []string
		topicToRemove string
		wantTopics    []string
		wantErr       bool
	}{
		{
			name:          "Unsubscribe from existing topic",
			initialTopics: []string{"event_updates", "central_ABC123"},
			topicToRemove: "event_updates",
			wantTopics:    []string{"central_ABC123"},
			wantErr:       false,
		},
		{
			name:          "Unsubscribe from non-existing topic",
			initialTopics: []string{"event_updates"},
			topicToRemove: "central_ABC123",
			wantTopics:    []string{"event_updates"},
			wantErr:       false,
		},
		{
			name:          "Unsubscribe from empty list",
			initialTopics: []string{},
			topicToRemove: "event_updates",
			wantTopics:    []string{},
			wantErr:       false,
		},
		{
			name:          "Unsubscribe with empty topic name",
			initialTopics: []string{"event_updates", ""},
			topicToRemove: "",
			wantTopics:    []string{"event_updates"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService("test-client")

			// Set initial topics
			s.topics = make([]string, len(tt.initialTopics))
			copy(s.topics, tt.initialTopics)

			// Call the method being tested
			err := s.Unsubscribe(tt.topicToRemove)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Unsubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check topics
			if !reflect.DeepEqual(s.GetTopics(), tt.wantTopics) {
				t.Errorf("After Unsubscribe() topics = %v, want %v", s.GetTopics(), tt.wantTopics)
			}
		})
	}
}

func TestService_GetTopics(t *testing.T) {
	tests := []struct {
		name          string
		initialTopics []string
		wantTopics    []string
	}{
		{
			name:          "Get empty topics list",
			initialTopics: []string{},
			wantTopics:    []string{},
		},
		{
			name:          "Get single topic",
			initialTopics: []string{"event_updates"},
			wantTopics:    []string{"event_updates"},
		},
		{
			name:          "Get multiple topics",
			initialTopics: []string{"event_updates", "central_ABC123", "central_DEF456"},
			wantTopics:    []string{"event_updates", "central_ABC123", "central_DEF456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService("test-client")

			// Set initial topics
			s.topics = make([]string, len(tt.initialTopics))
			copy(s.topics, tt.initialTopics)

			// Call the method being tested
			gotTopics := s.GetTopics()

			// Check topics
			if !reflect.DeepEqual(gotTopics, tt.wantTopics) {
				t.Errorf("GetTopics() = %v, want %v", gotTopics, tt.wantTopics)
			}

			// Verify that GetTopics returns a copy, not the original slice
			if len(gotTopics) > 0 {
				// Modify the returned slice
				gotTopics[0] = "modified_topic"

				// Check that the original slice is unchanged
				if s.topics[0] == "modified_topic" {
					t.Errorf("GetTopics() returned a reference to the original slice, not a copy")
				}
			}
		})
	}
}

func TestService_TopicOperationsConcurrency(t *testing.T) {
	// This test ensures that the topic operations are thread-safe
	s := NewService("test-client")

	// Add some initial topics
	s.Subscribe("topic1")
	s.Subscribe("topic2")

	// Run multiple goroutines to test concurrency
	done := make(chan bool)

	// Goroutine 1: Subscribe to topics
	go func() {
		for i := 0; i < 100; i++ {
			s.Subscribe("concurrent_topic_" + string(rune(i)))
		}
		done <- true
	}()

	// Goroutine 2: Unsubscribe from topics
	go func() {
		s.Unsubscribe("topic1")
		for i := 0; i < 50; i++ {
			s.Unsubscribe("concurrent_topic_" + string(rune(i)))
		}
		done <- true
	}()

	// Goroutine 3: Get topics
	go func() {
		for i := 0; i < 100; i++ {
			_ = s.GetTopics()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we got here without panicking, the test passes
}

func TestService_UpdateConnection(t *testing.T) {
	// Create a service with some topics
	s := NewService("test-client")
	s.Subscribe("topic1")
	s.Subscribe("topic2")

	// Store the topics for later comparison
	initialTopics := s.GetTopics()

	// Update the connection
	s.UpdateConnection(nil) // Using nil for testing as we don't need a real connection

	// Verify that topics are preserved after connection update
	updatedTopics := s.GetTopics()
	if !reflect.DeepEqual(initialTopics, updatedTopics) {
		t.Errorf("Topics not preserved after UpdateConnection(): got %v, want %v", updatedTopics, initialTopics)
	}

	// Verify that the connection state is updated
	if !s.IsConnected() {
		t.Errorf("Connection state not updated after UpdateConnection()")
	}
}
