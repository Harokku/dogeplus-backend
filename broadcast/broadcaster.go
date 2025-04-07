// Package broadcast provides functionality for real-time message broadcasting to connected clients.
// It defines interfaces and implementations for managing client connections, sending messages,
// and handling topic-based subscriptions. The package includes heartbeat mechanisms and
// connection monitoring to ensure reliable real-time communication.
package broadcast

import "time"

// Broadcaster defines the interface for a client that can connect, disconnect, and send messages.
// It also provides methods to check connection status and last activity time.
// It includes methods for subscribing and unsubscribing to topics.
type Broadcaster interface {
	Connect() error
	Disconnect() error
	Send(message []byte) error
	IsConnected() bool
	LastActivity() time.Time
	Subscribe(topic string) error
	Unsubscribe(topic string) error
	GetTopics() []string
}
