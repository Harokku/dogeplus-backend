package broadcast

import "time"

// Broadcaster defines the interface for a client that can connect, disconnect, and send messages.
// It also provides methods to check connection status and last activity time.
type Broadcaster interface {
	Connect() error
	Disconnect() error
	Send(message []byte) error
	IsConnected() bool
	LastActivity() time.Time
}
