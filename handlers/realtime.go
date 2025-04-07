package handlers

import (
	"dogeplus-backend/broadcast"
	"dogeplus-backend/ws"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"time"
)

// WsUpgrader is a middleware that upgrades HTTP connections to WebSocket connections.
// It generates a unique ID for each WebSocket connection by combining the IP address with a UUID.
// This ensures that each browser tab gets a unique identifier, even if they're in the same browser.
func WsUpgrader(manager *broadcast.ConnectionManager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Always generate a new unique ID for each connection
		// This ensures each tab gets a unique identifier, even in the same browser
		tabUUID := uuid.New().String()
		clientID := ctx.IP() + "-" + tabUUID

		// Set the cookie with SameSite=Strict to prevent it from being sent in cross-site requests
		// Path=/ ensures the cookie is available for the entire domain
		ctx.Cookie(&fiber.Cookie{
			Name:     "client_id",
			Value:    clientID,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			SameSite: "strict",
			Path:     "/",
		})

		// Check if is an update request, set the local with unique id and pass to next handler
		if websocket.IsWebSocketUpgrade(ctx) {
			ctx.Locals("client_id", clientID)
			return ctx.Next()
		}

		return fiber.ErrUpgradeRequired
	}
}

// WsHandler handles WebSocket connections.
// It creates a new Service or updates an existing one, adds it to the ConnectionManager,
// and handles WebSocket messages including heartbeats.
func WsHandler(cm *broadcast.ConnectionManager) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		clientID := c.Locals("client_id").(string)

		// Check if client already exists (reconnection case)
		var wsBroadcaster *ws.Service

		if cm.ClientExists(clientID) {
			// Get existing client
			broadcaster, exists := cm.GetClient(clientID)
			if exists {
				// Try to cast to Service
				service, ok := broadcaster.(*ws.Service)
				if ok {
					wsBroadcaster = service
					wsBroadcaster.UpdateConnection(c)
					log.Infof("Client %s reconnected", clientID)
				} else {
					// Unexpected type, create new
					wsBroadcaster = ws.NewService(clientID)
					wsBroadcaster.UpdateConnection(c)
					cm.AddClient(clientID, wsBroadcaster)
					defer cm.RemoveClient(clientID)
					log.Infof("Client %s connected (new - type mismatch)", clientID)
				}
			} else {
				// Client exists but couldn't be retrieved, create new
				wsBroadcaster = ws.NewService(clientID)
				wsBroadcaster.UpdateConnection(c)
				cm.AddClient(clientID, wsBroadcaster)
				defer cm.RemoveClient(clientID)
				log.Infof("Client %s connected (new - retrieval failed)", clientID)
			}
		} else {
			// New client
			wsBroadcaster = ws.NewService(clientID)
			wsBroadcaster.UpdateConnection(c)
			cm.AddClient(clientID, wsBroadcaster)
			defer cm.RemoveClient(clientID)
			log.Infof("Client %s connected (new)", clientID)
		}

		// Send initial connection confirmation
		initialMsg := map[string]interface{}{
			"type":      "connected",
			"client_id": clientID,
			"timestamp": time.Now().Unix(),
		}

		initialMsgJSON, _ := json.Marshal(initialMsg)
		_ = wsBroadcaster.Send(initialMsgJSON)

		var (
			messageType int
			messageData []byte
			readErr     error
		)

		// Set read deadline for initial message
		_ = c.SetReadDeadline(time.Now().Add(60 * time.Second))

		for {
			if messageType, messageData, readErr = c.ReadMessage(); readErr != nil {
				log.Infof("Read error for client %s: %v", clientID, readErr)
				_ = wsBroadcaster.Disconnect()
				return
			}

			// Reset read deadline after successful read
			_ = c.SetReadDeadline(time.Now().Add(60 * time.Second))

			// Handle ping/pong for heartbeat
			if messageType == websocket.TextMessage {
				if string(messageData) == "ping" {
					_ = wsBroadcaster.Send([]byte("pong"))
				} else {
					// Try to parse as JSON to see if it's a structured message
					var msg map[string]interface{}
					if err := json.Unmarshal(messageData, &msg); err == nil {
						// Handle structured messages if needed
						if msgType, ok := msg["type"].(string); ok {
							switch msgType {
							case "heartbeat":
								// Respond to heartbeat
								heartbeatResponse := map[string]interface{}{
									"type":      "heartbeat_ack",
									"timestamp": time.Now().Unix(),
								}
								heartbeatJSON, _ := json.Marshal(heartbeatResponse)
								_ = wsBroadcaster.Send(heartbeatJSON)
							case "subscribe":
								// Handle topic subscription
								if topic, ok := msg["topic"].(string); ok && topic != "" {
									err := wsBroadcaster.Subscribe(topic)
									response := map[string]interface{}{
										"type":    "subscribe_ack",
										"topic":   topic,
										"success": err == nil,
									}
									if err != nil {
										response["error"] = err.Error()
									}
									responseJSON, _ := json.Marshal(response)
									_ = wsBroadcaster.Send(responseJSON)
									log.Infof("Client %s subscribed to topic: %s", clientID, topic)
								}
							case "unsubscribe":
								// Handle topic unsubscription
								if topic, ok := msg["topic"].(string); ok && topic != "" {
									err := wsBroadcaster.Unsubscribe(topic)
									response := map[string]interface{}{
										"type":    "unsubscribe_ack",
										"topic":   topic,
										"success": err == nil,
									}
									if err != nil {
										response["error"] = err.Error()
									}
									responseJSON, _ := json.Marshal(response)
									_ = wsBroadcaster.Send(responseJSON)
									log.Infof("Client %s unsubscribed from topic: %s", clientID, topic)
								}
							case "get_topics":
								// Return the list of subscribed topics
								topics := wsBroadcaster.GetTopics()
								response := map[string]interface{}{
									"type":   "topics",
									"topics": topics,
								}
								responseJSON, _ := json.Marshal(response)
								_ = wsBroadcaster.Send(responseJSON)
							}
						}
					}
				}
			}
		}
	})
}
