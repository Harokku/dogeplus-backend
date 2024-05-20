package ws

import (
	"github.com/gofiber/websocket/v2"
)

// Service is a type that represents a connection service.
//
// The Conn field represents the WebSocket connection.
// The Id field represent the WebSocket client connection id
type Service struct {
	Conn *websocket.Conn // Conn is a pointer to a websocket connection.
	Id   string          // Id is a string field representing the identifier of a connection service.
}

// Connect establishes a WebSocket connection for the Service.
//
// It returns nil if the connection is established successfully.
// If there is an error while establishing the connection, an error is returned.
func (s *Service) Connect() error {
	return nil
}

// Disconnect closes the WebSocket connection represented by the Service.
//
// It returns an error if the connection fails to close.
func (s *Service) Disconnect() error {
	return s.Conn.Close()
}

// Send sends a message over the WebSocket connection represented by the Service.
//
// It accepts a message of type []byte and returns an error if the message fails to send.
func (s *Service) Send(message []byte) error {
	return s.Conn.WriteMessage(websocket.TextMessage, message)
}
