package ws

import (
	"errors"
	"github.com/gofiber/websocket/v2"
	"sync"
	"time"
)

// Service is a type that represents a WebSocket connection service.
// It implements the broadcast.Broadcaster interface.
type Service struct {
	Conn         *websocket.Conn // WebSocket connection
	Id           string          // Client identifier
	connected    bool            // Connection state
	lastActivity time.Time       // Time of last activity
	mu           sync.Mutex      // Mutex for thread safety
}

// NewService creates a new Service with the given client ID.
// The service is initialized as disconnected with the current time as last activity.
func NewService(id string) *Service {
	return &Service{
		Id:           id,
		connected:    false,
		lastActivity: time.Now(),
	}
}

// Connect establishes a WebSocket connection for the Service.
// It sets the connected state to true and updates the last activity time.
// It returns nil if the connection is established successfully.
func (s *Service) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = true
	s.lastActivity = time.Now()
	return nil
}

// Disconnect closes the WebSocket connection represented by the Service.
// It sets the connected state to false and returns any error from closing the connection.
func (s *Service) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}

// Send sends a message over the WebSocket connection represented by the Service.
// It checks if the service is connected and has a valid connection before sending.
// It updates the last activity time if the message is sent successfully.
// It returns an error if the service is not connected or if the message fails to send.
func (s *Service) Send(message []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected || s.Conn == nil {
		return errors.New("not connected")
	}

	err := s.Conn.WriteMessage(websocket.TextMessage, message)
	if err == nil {
		s.lastActivity = time.Now()
	}
	return err
}

// IsConnected returns whether the service is currently connected.
func (s *Service) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

// LastActivity returns the time of the last activity for this service.
func (s *Service) LastActivity() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastActivity
}

// UpdateConnection updates the connection with a new WebSocket connection.
// It closes any existing connection, sets the new connection, updates the
// connected state to true, and updates the last activity time.
func (s *Service) UpdateConnection(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Conn != nil {
		_ = s.Conn.Close()
	}

	s.Conn = conn
	s.connected = true
	s.lastActivity = time.Now()
}
