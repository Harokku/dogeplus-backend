package broadcast

import "sync"

type ConnectionManager struct {
	Clients map[Broadcaster]bool
	lock    sync.RWMutex
}

func (cm *ConnectionManager) AddClient(client Broadcaster) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.Clients[client] = true
}

func (cm *ConnectionManager) RemoveClient(client Broadcaster) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	delete(cm.Clients, client)
}

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
