// Package broadcast provides functionality for real-time message broadcasting to connected clients.
// It defines interfaces and implementations for managing client connections, sending messages,
// and handling topic-based subscriptions. The package includes heartbeat mechanisms and
// connection monitoring to ensure reliable real-time communication.
package broadcast

import (
	"reflect"
	"testing"
	"time"
)

// MockBroadcaster is a mock implementation of the Broadcaster interface for testing
type MockBroadcaster struct {
	connected     bool
	lastActivity  time.Time
	topics        []string
	receivedMsgs  [][]byte
	connectErr    error
	disconnectErr error
	sendErr       error
}

func NewMockBroadcaster(connected bool, topics []string) *MockBroadcaster {
	return &MockBroadcaster{
		connected:    connected,
		lastActivity: time.Now(),
		topics:       topics,
		receivedMsgs: [][]byte{},
	}
}

func (m *MockBroadcaster) Connect() error {
	m.connected = true
	return m.connectErr
}

func (m *MockBroadcaster) Disconnect() error {
	m.connected = false
	return m.disconnectErr
}

func (m *MockBroadcaster) Send(message []byte) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.receivedMsgs = append(m.receivedMsgs, message)
	m.lastActivity = time.Now()
	return nil
}

func (m *MockBroadcaster) IsConnected() bool {
	return m.connected
}

func (m *MockBroadcaster) LastActivity() time.Time {
	return m.lastActivity
}

func (m *MockBroadcaster) Subscribe(topic string) error {
	for _, t := range m.topics {
		if t == topic {
			return nil // Already subscribed
		}
	}
	m.topics = append(m.topics, topic)
	return nil
}

func (m *MockBroadcaster) Unsubscribe(topic string) error {
	for i, t := range m.topics {
		if t == topic {
			m.topics = append(m.topics[:i], m.topics[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockBroadcaster) GetTopics() []string {
	topics := make([]string, len(m.topics))
	copy(topics, m.topics)
	return topics
}

func TestConnectionManager_BroadcastToTopic(t *testing.T) {
	tests := []struct {
		name            string
		clients         map[string]*MockBroadcaster
		topic           string
		message         []byte
		expectedClients []string // IDs of clients that should receive the message
	}{
		{
			name: "Broadcast to single client subscribed to topic",
			clients: map[string]*MockBroadcaster{
				"client1": NewMockBroadcaster(true, []string{"topic1"}),
				"client2": NewMockBroadcaster(true, []string{"topic2"}),
			},
			topic:           "topic1",
			message:         []byte("test message"),
			expectedClients: []string{"client1"},
		},
		{
			name: "Broadcast to multiple clients subscribed to topic",
			clients: map[string]*MockBroadcaster{
				"client1": NewMockBroadcaster(true, []string{"topic1", "topic2"}),
				"client2": NewMockBroadcaster(true, []string{"topic1"}),
				"client3": NewMockBroadcaster(true, []string{"topic3"}),
			},
			topic:           "topic1",
			message:         []byte("test message"),
			expectedClients: []string{"client1", "client2"},
		},
		{
			name: "Broadcast to topic with no subscribers",
			clients: map[string]*MockBroadcaster{
				"client1": NewMockBroadcaster(true, []string{"topic1"}),
				"client2": NewMockBroadcaster(true, []string{"topic2"}),
			},
			topic:           "topic3",
			message:         []byte("test message"),
			expectedClients: []string{},
		},
		{
			name: "Broadcast to disconnected clients",
			clients: map[string]*MockBroadcaster{
				"client1": NewMockBroadcaster(false, []string{"topic1"}),
				"client2": NewMockBroadcaster(true, []string{"topic1"}),
			},
			topic:           "topic1",
			message:         []byte("test message"),
			expectedClients: []string{"client2"},
		},
		{
			name: "Broadcast to event_updates topic",
			clients: map[string]*MockBroadcaster{
				"client1": NewMockBroadcaster(true, []string{"event_updates"}),
				"client2": NewMockBroadcaster(true, []string{"central_ABC123"}),
			},
			topic:           "event_updates",
			message:         []byte("test message"),
			expectedClients: []string{"client1"},
		},
		{
			name: "Broadcast to central_ABC123 topic",
			clients: map[string]*MockBroadcaster{
				"client1": NewMockBroadcaster(true, []string{"event_updates"}),
				"client2": NewMockBroadcaster(true, []string{"central_ABC123"}),
			},
			topic:           "central_ABC123",
			message:         []byte("test message"),
			expectedClients: []string{"client2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new connection manager
			cm := NewConnectionManager()

			// Add mock clients to the connection manager
			for id, client := range tt.clients {
				cm.Clients[id] = client
			}

			// Broadcast the message to the topic
			cm.BroadcastToTopic(tt.topic, tt.message)

			// Check that the expected clients received the message
			for id, client := range tt.clients {
				shouldReceive := false
				for _, expectedID := range tt.expectedClients {
					if id == expectedID {
						shouldReceive = true
						break
					}
				}

				if shouldReceive {
					if len(client.receivedMsgs) == 0 {
						t.Errorf("Client %s should have received the message but didn't", id)
					} else if !reflect.DeepEqual(client.receivedMsgs[0], tt.message) {
						t.Errorf("Client %s received message %v, want %v", id, client.receivedMsgs[0], tt.message)
					}
				} else {
					if len(client.receivedMsgs) > 0 {
						t.Errorf("Client %s should not have received the message but did", id)
					}
				}
			}
		})
	}
}

func TestConnectionManager_BroadcastToTopic_WithSendError(t *testing.T) {
	// Create a new connection manager
	cm := NewConnectionManager()

	// Create a client that will return an error when Send is called
	errorClient := &MockBroadcaster{
		connected: true,
		topics:    []string{"topic1"},
		sendErr:   &mockError{message: "send error"},
	}

	// Create a normal client
	normalClient := NewMockBroadcaster(true, []string{"topic1"})

	// Add clients to the connection manager
	cm.Clients["error_client"] = errorClient
	cm.Clients["normal_client"] = normalClient

	// Broadcast a message to the topic
	message := []byte("test message")
	cm.BroadcastToTopic("topic1", message)

	// Check that the normal client received the message despite the error from the other client
	if len(normalClient.receivedMsgs) == 0 {
		t.Errorf("Normal client should have received the message but didn't")
	} else if !reflect.DeepEqual(normalClient.receivedMsgs[0], message) {
		t.Errorf("Normal client received message %v, want %v", normalClient.receivedMsgs[0], message)
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestConnectionManager_ClientExists(t *testing.T) {
	// Create a new connection manager
	cm := NewConnectionManager()

	// Add a client
	cm.Clients["client1"] = NewMockBroadcaster(true, []string{})

	// Test that ClientExists returns true for existing client
	if !cm.ClientExists("client1") {
		t.Errorf("ClientExists() returned false for existing client")
	}

	// Test that ClientExists returns false for non-existing client
	if cm.ClientExists("client2") {
		t.Errorf("ClientExists() returned true for non-existing client")
	}
}

func TestConnectionManager_GetClient(t *testing.T) {
	// Create a new connection manager
	cm := NewConnectionManager()

	// Create a client
	client := NewMockBroadcaster(true, []string{})

	// Add the client
	cm.Clients["client1"] = client

	// Test that GetClient returns the client and true for existing client
	gotClient, exists := cm.GetClient("client1")
	if !exists {
		t.Errorf("GetClient() returned exists=false for existing client")
	}
	if gotClient != client {
		t.Errorf("GetClient() returned wrong client")
	}

	// Test that GetClient returns nil and false for non-existing client
	gotClient, exists = cm.GetClient("client2")
	if exists {
		t.Errorf("GetClient() returned exists=true for non-existing client")
	}
	if gotClient != nil {
		t.Errorf("GetClient() returned non-nil client for non-existing client")
	}
}
