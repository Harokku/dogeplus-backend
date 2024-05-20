package broadcast

import "sync"

// ConnectionManager handles the management of client connections for real-time broadcasting.
// It maintains a map of connected broadcasters and provides methods to add, remove, and broadcast messages to clients.
type ConnectionManager struct {
	Clients map[Broadcaster]bool
	lock    sync.RWMutex
}

// AddClient adds a client to the ConnectionManager's list of clients.
// It takes a Broadcaster client as a parameter.
// It acquires a lock on the ConnectionManager to ensure thread safety.
// It unlocks the ConnectionManager when the function exits using a deferred statement.
// It adds the client to the Clients map with a value of true.
//
// Example usage:
//
//	cm := &ConnectionManager{}
//	client := &ws.Service{}
//	cm.AddClient(client)
func (cm *ConnectionManager) AddClient(client Broadcaster) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.Clients[client] = true
}

// RemoveClient removes a client from the ConnectionManager's list of clients.
// It takes a Broadcaster client as a parameter.
// It acquires a lock on the ConnectionManager to ensure thread safety.
// It unlocks the ConnectionManager when the function exits using a deferred statement.
// It deletes the client from the Clients map.
//
// Example usage:
//
//	cm := &ConnectionManager{}
//	client := &ws.Service{}
//	cm.RemoveClient(client)
func (cm *ConnectionManager) RemoveClient(client Broadcaster) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	delete(cm.Clients, client)
}

// Broadcast sends a message to all connected clients.
// It acquires a read lock on the ConnectionManager to ensure thread safety.
// It releases the read lock when the function exits using a deferred statement.
// It iterates over each client in the Clients map and calls the Send method to send the message.
// If there is an error while sending the message, the function returns.
func (cm *ConnectionManager) Broadcast(message []byte) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	for client := range cm.Clients {
		err := client.Send(message)
		if err != nil {
			return
		}
	}
}
