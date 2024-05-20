package handlers

import (
	"dogeplus-backend/broadcast"
	"dogeplus-backend/ws"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// TODO: implement unique id reading from client
func WsUpgrader(manager *broadcast.ConnectionManager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Read the id from http request
		clientID := ctx.Cookies("client_id")

		// Check if is an update request, set che local with unique id and pass to next handler
		if websocket.IsWebSocketUpgrade(ctx) {
			ctx.Locals("client_id", clientID)
			return ctx.Next()
		}

		return fiber.ErrUpgradeRequired
	}
}

// TODO: add conn id saving logic
func WsHandler(cm *broadcast.ConnectionManager) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		clientID := c.Locals("client_id").(string)

		wsBroadcaster := &ws.Service{Conn: c, Id: clientID}
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

			// TODO: Debug only, remove in prod
			if messageType == websocket.TextMessage {
				cm.Broadcast(messageData)
			}
		}
	})
}
