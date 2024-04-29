package ws

import (
	"github.com/gofiber/websocket/v2"
)

type Service struct {
	Conn *websocket.Conn
}

func (s *Service) Connect() error {
	return nil
}

func (s *Service) Disconnect() error {
	return s.Conn.Close()
}

func (s *Service) Send(message []byte) error {
	return s.Conn.WriteMessage(websocket.TextMessage, message)
}
