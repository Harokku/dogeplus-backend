package ws

import (
	"dogeplus-backend/broadcast"
	"reflect"
	"sync"
	"testing"
	"time"
)

// TestBroadcaster is a custom implementation of the Broadcaster interface for testing
type TestBroadcaster struct {
	id           string
	connected    bool
	lastActivity time.Time
	topics       []string
	receivedMsgs [][]byte
	mu           sync.Mutex
}

func NewTestBroadcaster(id string) *TestBroadcaster {
	return &TestBroadcaster{
		id:           id,
		connected:    false,
		lastActivity: time.Now(),
		topics:       []string{},
		receivedMsgs: [][]byte{},
	}
}

func (tb *TestBroadcaster) Connect() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.connected = true
	tb.lastActivity = time.Now()
	return nil
}

func (tb *TestBroadcaster) Disconnect() error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.connected = false
	return nil
}

func (tb *TestBroadcaster) Send(message []byte) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.receivedMsgs = append(tb.receivedMsgs, message)
	tb.lastActivity = time.Now()
	return nil
}

func (tb *TestBroadcaster) IsConnected() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.connected
}

func (tb *TestBroadcaster) LastActivity() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.lastActivity
}

func (tb *TestBroadcaster) Subscribe(topic string) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for _, t := range tb.topics {
		if t == topic {
			return nil // Already subscribed
		}
	}
	tb.topics = append(tb.topics, topic)
	return nil
}

func (tb *TestBroadcaster) Unsubscribe(topic string) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for i, t := range tb.topics {
		if t == topic {
			tb.topics = append(tb.topics[:i], tb.topics[i+1:]...)
			return nil
		}
	}
	return nil
}

func (tb *TestBroadcaster) GetTopics() []string {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	topics := make([]string, len(tb.topics))
	copy(topics, tb.topics)
	return topics
}

func (tb *TestBroadcaster) GetReceivedMessages() [][]byte {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	msgs := make([][]byte, len(tb.receivedMsgs))
	copy(msgs, tb.receivedMsgs)
	return msgs
}

// TestIntegration_SubscriptionAndBroadcast tests the full flow from subscription to receiving topic-specific messages
func TestIntegration_SubscriptionAndBroadcast(t *testing.T) {
	// Create a connection manager
	cm := broadcast.NewConnectionManager()

	// Create two test broadcasters
	broadcaster1 := NewTestBroadcaster("client1")
	broadcaster2 := NewTestBroadcaster("client2")

	// Add the broadcasters to the connection manager
	cm.AddClient("client1", broadcaster1)
	cm.AddClient("client2", broadcaster2)

	// Connect the broadcasters
	err := broadcaster1.Connect()
	if err != nil {
		t.Fatalf("Failed to connect broadcaster1: %v", err)
	}

	err = broadcaster2.Connect()
	if err != nil {
		t.Fatalf("Failed to connect broadcaster2: %v", err)
	}

	// Subscribe broadcaster1 to topic1
	err = broadcaster1.Subscribe("topic1")
	if err != nil {
		t.Fatalf("Failed to subscribe broadcaster1 to topic1: %v", err)
	}

	// Subscribe broadcaster2 to topic2
	err = broadcaster2.Subscribe("topic2")
	if err != nil {
		t.Fatalf("Failed to subscribe broadcaster2 to topic2: %v", err)
	}

	// Broadcast a message to topic1
	message1 := []byte("message to topic1")
	cm.BroadcastToTopic("topic1", message1)

	// Broadcast a message to topic2
	message2 := []byte("message to topic2")
	cm.BroadcastToTopic("topic2", message2)

	// Give some time for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Check that broadcaster1 received the message for topic1 but not topic2
	messages1 := broadcaster1.GetReceivedMessages()
	if len(messages1) != 1 {
		t.Errorf("broadcaster1 received %d messages, want 1", len(messages1))
	} else if !reflect.DeepEqual(messages1[0], message1) {
		t.Errorf("broadcaster1 received message %v, want %v", messages1[0], message1)
	}

	// Check that broadcaster2 received the message for topic2 but not topic1
	messages2 := broadcaster2.GetReceivedMessages()
	if len(messages2) != 1 {
		t.Errorf("broadcaster2 received %d messages, want 1", len(messages2))
	} else if !reflect.DeepEqual(messages2[0], message2) {
		t.Errorf("broadcaster2 received message %v, want %v", messages2[0], message2)
	}
}

// TestIntegration_ReconnectionWithTopicPersistence tests that subscriptions are preserved across reconnections
func TestIntegration_ReconnectionWithTopicPersistence(t *testing.T) {
	// Create a connection manager
	cm := broadcast.NewConnectionManager()

	// Create a test broadcaster
	broadcaster := NewTestBroadcaster("client1")

	// Add the broadcaster to the connection manager
	cm.AddClient("client1", broadcaster)

	// Connect the broadcaster
	err := broadcaster.Connect()
	if err != nil {
		t.Fatalf("Failed to connect broadcaster: %v", err)
	}

	// Subscribe to topics
	topics := []string{"topic1", "topic2", "topic3"}
	for _, topic := range topics {
		err = broadcaster.Subscribe(topic)
		if err != nil {
			t.Fatalf("Failed to subscribe to %s: %v", topic, err)
		}
	}

	// Verify that the broadcaster is subscribed to the topics
	subscribedTopics := broadcaster.GetTopics()
	if !reflect.DeepEqual(subscribedTopics, topics) {
		t.Errorf("broadcaster is subscribed to %v, want %v", subscribedTopics, topics)
	}

	// Disconnect the broadcaster
	err = broadcaster.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect broadcaster: %v", err)
	}

	// Verify that the broadcaster is disconnected
	if broadcaster.IsConnected() {
		t.Errorf("broadcaster is still connected after Disconnect()")
	}

	// Reconnect the broadcaster
	err = broadcaster.Connect()
	if err != nil {
		t.Fatalf("Failed to reconnect broadcaster: %v", err)
	}

	// Verify that the broadcaster is connected
	if !broadcaster.IsConnected() {
		t.Errorf("broadcaster is not connected after Connect()")
	}

	// Verify that the broadcaster is still subscribed to the same topics
	subscribedTopics = broadcaster.GetTopics()
	if !reflect.DeepEqual(subscribedTopics, topics) {
		t.Errorf("After reconnection, broadcaster is subscribed to %v, want %v", subscribedTopics, topics)
	}

	// Clear any messages that might have been received
	broadcaster.receivedMsgs = [][]byte{}

	// Broadcast a message to one of the topics
	message := []byte("message after reconnection")
	cm.BroadcastToTopic("topic2", message)

	// Give some time for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Check that the broadcaster received the message
	messages := broadcaster.GetReceivedMessages()
	if len(messages) != 1 {
		t.Errorf("After reconnection, broadcaster received %d messages, want 1", len(messages))
	} else if !reflect.DeepEqual(messages[0], message) {
		t.Errorf("After reconnection, broadcaster received message %v, want %v", messages[0], message)
	}
}
