package handlers

import (
	"dogeplus-backend/broadcast"
	"dogeplus-backend/ws"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func WsUpgrade(cm *broadcast.ConnectionManager) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		wsBroadcaster := &ws.Service{Conn: c}
		cm.AddClient(wsBroadcaster)

		defer cm.RemoveClient(wsBroadcaster)

		var (
			messageType int
			messageData []byte
			readErr     error
		)

		for {
			if messageType, messageData, readErr = c.ReadMessage(); readErr != nil {
				_ = wsBroadcaster.Disconnect()
				return
			}

			if messageType == websocket.TextMessage {
				cm.Broadcast(messageData)
			}
		}
	})
}
