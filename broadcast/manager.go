// Package broadcast provides functionality for real-time message broadcasting to connected clients.
// It defines interfaces and implementations for managing client connections, sending messages,
// and handling topic-based subscriptions. The package includes heartbeat mechanisms and
// connection monitoring to ensure reliable real-time communication.
package broadcast

import (
	"github.com/gofiber/fiber/v2/log"
	"strconv"
	"sync"
	"time"
)

// ConnectionManager handles the management of client connections for real-time broadcasting.
// It maintains a map of connected broadcasters and provides methods to add, remove, and broadcast messages to clients.
type ConnectionManager struct {
	Clients       map[string]Broadcaster // Map of client ID to broadcaster
	lock          sync.RWMutex
	heartbeatTick time.Duration // Interval for sending heartbeats
	done          chan struct{} // Channel to signal shutdown
}

// NewConnectionManager creates a new connection manager with default settings
// and starts the heartbeat and monitoring goroutines.
func NewConnectionManager() *ConnectionManager {
	cm := &ConnectionManager{
		Clients:       make(map[string]Broadcaster),
		heartbeatTick: 30 * time.Second,
		done:          make(chan struct{}),
	}

	// Start the heartbeat goroutine
	go cm.heartbeatLoop()

	// Start the connection monitoring goroutine
	go cm.monitorConnections()

	return cm
}

// AddClient adds a client to the ConnectionManager's list of clients.
// It takes a client ID and a Broadcaster as parameters.
// It acquires a lock on the ConnectionManager to ensure thread safety.
// It unlocks the ConnectionManager when the function exits using a deferred statement.
// It adds the client to the Clients map with the client ID as the key.
//
// Example usage:
//
//	cm := NewConnectionManager()
//	client := &ws.Service{Id: "client1"}
//	cm.AddClient("client1", client)
func (cm *ConnectionManager) AddClient(clientID string, client Broadcaster) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.Clients[clientID] = client
}

// RemoveClient removes a client from the ConnectionManager's list of clients.
// It takes a client ID as a parameter.
// It acquires a lock on the ConnectionManager to ensure thread safety.
// It unlocks the ConnectionManager when the function exits using a deferred statement.
// It deletes the client from the Clients map.
//
// Example usage:
//
//	cm := NewConnectionManager()
//	cm.RemoveClient("client1")
func (cm *ConnectionManager) RemoveClient(clientID string) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	delete(cm.Clients, clientID)
}

// Broadcast sends a message to all connected clients.
// It acquires a read lock on the ConnectionManager to ensure thread safety.
// It releases the read lock when the function exits using a deferred statement.
// It iterates over each client in the Clients map and calls the Send method to send the message.
// If there is an error while sending the message to a client, it logs the error and continues
// broadcasting to other clients.
func (cm *ConnectionManager) Broadcast(message []byte) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	for id, client := range cm.Clients {
		if !client.IsConnected() {
			continue
		}

		err := client.Send(message)
		if err != nil {
			log.Errorf("Error broadcasting to client %s: %v", id, err)
			// Continue broadcasting to other clients
		}
	}
}

// BroadcastToTopic sends a message to all connected clients subscribed to the specified topic.
// It acquires a read lock on the ConnectionManager to ensure thread safety.
// It releases the read lock when the function exits using a deferred statement.
// It iterates over each client in the Clients map, checks if the client is subscribed to the topic,
// and calls the Send method to send the message if it is.
// If there is an error while sending the message to a client, it logs the error and continues
// broadcasting to other clients.
func (cm *ConnectionManager) BroadcastToTopic(topic string, message []byte) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	for id, client := range cm.Clients {
		if !client.IsConnected() {
			continue
		}

		// Check if the client is subscribed to the topic
		topics := client.GetTopics()
		subscribed := false
		for _, t := range topics {
			if t == topic {
				subscribed = true
				break
			}
		}

		if !subscribed {
			continue
		}

		err := client.Send(message)
		if err != nil {
			log.Errorf("Error broadcasting to client %s on topic %s: %v", id, topic, err)
			// Continue broadcasting to other clients
		}
	}
}

// heartbeatLoop sends periodic heartbeats to all clients
func (cm *ConnectionManager) heartbeatLoop() {
	// Recover from panics to prevent the goroutine from crashing the application
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic in heartbeatLoop: %v", r)
			// Restart the goroutine after a short delay
			time.Sleep(time.Second)
			go cm.heartbeatLoop()
		}
	}()

	ticker := time.NewTicker(cm.heartbeatTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.sendHeartbeats()
		case <-cm.done:
			return
		}
	}
}

// sendHeartbeats sends a heartbeat to all connected clients
func (cm *ConnectionManager) sendHeartbeats() {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	heartbeatMsg := []byte(`{"type":"heartbeat","timestamp":` + strconv.FormatInt(time.Now().Unix(), 10) + `}`)

	for id, client := range cm.Clients {
		if !client.IsConnected() {
			continue
		}

		err := client.Send(heartbeatMsg)
		if err != nil {
			// Log the error but don't remove the client yet
			log.Errorf("Error sending heartbeat to client %s: %v", id, err)
		}
	}
}

// monitorConnections periodically checks for stale connections
func (cm *ConnectionManager) monitorConnections() {
	// Recover from panics to prevent the goroutine from crashing the application
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic in monitorConnections: %v", r)
			// Restart the goroutine after a short delay
			time.Sleep(time.Second)
			go cm.monitorConnections()
		}
	}()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.cleanStaleConnections()
		case <-cm.done:
			return
		}
	}
}

// cleanStaleConnections removes connections that haven't had activity in too long
func (cm *ConnectionManager) cleanStaleConnections() {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	staleThreshold := time.Now().Add(-10 * time.Minute)

	for id, client := range cm.Clients {
		if client.LastActivity().Before(staleThreshold) {
			log.Infof("Removing stale connection for client %s", id)
			if err := client.Disconnect(); err != nil {
				log.Warnf("Error disconnecting stale client %s: %v", id, err)
				// Continue with removal despite the error
			}
			delete(cm.Clients, id)
		}
	}
}

// Shutdown gracefully stops the connection manager
func (cm *ConnectionManager) Shutdown() {
	close(cm.done)

	cm.lock.Lock()
	defer cm.lock.Unlock()

	// Disconnect all clients
	for id, client := range cm.Clients {
		if err := client.Disconnect(); err != nil {
			log.Warnf("Error disconnecting client %s during shutdown: %v", id, err)
			// Continue with shutdown despite errors
		}
	}

	// Clear the clients map
	cm.Clients = make(map[string]Broadcaster)
}

// ClientExists checks if a client with the given ID exists
func (cm *ConnectionManager) ClientExists(clientID string) bool {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	_, exists := cm.Clients[clientID]
	return exists
}

// GetClient returns the client with the given ID and a boolean indicating if it exists
func (cm *ConnectionManager) GetClient(clientID string) (Broadcaster, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	client, exists := cm.Clients[clientID]
	return client, exists
}
