
# WebSocket Functionality Documentation for DogePlus Backend

This documentation provides a comprehensive guide on how to use the WebSocket functionality in the DogePlus Backend project. The WebSocket implementation enables real-time communication between the server and clients, allowing for instant updates and notifications without requiring page refreshes.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Establishing a WebSocket Connection](#establishing-a-websocket-connection)
4. [Topic-Based Subscription System](#topic-based-subscription-system)
5. [Heartbeat and Ping Response Management](#heartbeat-and-ping-response-management)
6. [Available Topics](#available-topics)
7. [Client-Side Implementation](#client-side-implementation)
8. [Server-Side Implementation](#server-side-implementation)
9. [Best Practices and Considerations](#best-practices-and-considerations)
10. [Troubleshooting](#troubleshooting)

## Overview

The WebSocket functionality in DogePlus Backend provides:

- Real-time bidirectional communication between server and clients
- Topic-based subscription system for targeted message delivery
- Automatic reconnection handling
- Heartbeat mechanism to maintain connection health
- Connection monitoring and cleanup of stale connections

This system is designed to deliver updates about events, tasks, and other important information to clients in real-time, improving the user experience and reducing the need for polling or page refreshes.

## Architecture

The WebSocket implementation consists of several key components:

1. **WebSocket Service (`ws.Service`)**: Represents a WebSocket connection and implements the `Broadcaster` interface.
2. **Connection Manager (`broadcast.ConnectionManager`)**: Manages WebSocket connections and provides methods for broadcasting messages.
3. **WebSocket Handlers (`handlers.WsUpgrader` and `handlers.WsHandler`)**: Handle WebSocket connection upgrades and message processing.
4. **Topic-Based Subscription System**: Allows clients to subscribe to specific topics and receive only relevant messages.

### Component Relationships

```
Client <---> WsHandler <---> Service <---> ConnectionManager
                                |
                                v
                          Topic Subscriptions
```

## Establishing a WebSocket Connection

### Endpoint

The WebSocket endpoint is available at:

```
ws://your-server/api/v1/ws/
```

### Connection Process

1. The client initiates a WebSocket connection to the endpoint.
2. The `WsUpgrader` middleware generates a unique client ID and sets it as a cookie.
3. The connection is upgraded from HTTP to WebSocket.
4. The `WsHandler` creates a new `Service` instance or updates an existing one.
5. The client receives a connection confirmation message.

### Connection Confirmation Message

Upon successful connection, the client receives a JSON message:

```json
{
  "type": "connected",
  "client_id": "127.0.0.1-550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1616161616
}
```

## Topic-Based Subscription System

The topic-based subscription system allows clients to subscribe to specific topics and receive only the messages they're interested in, reducing bandwidth and CPU usage.

### Subscribing to a Topic

To subscribe to a topic, send a JSON message:

```json
{
  "type": "subscribe",
  "topic": "topic_name"
}
```

The server will respond with an acknowledgment:

```json
{
  "type": "subscribe_ack",
  "topic": "topic_name",
  "success": true
}
```

### Unsubscribing from a Topic

To unsubscribe from a topic, send a JSON message:

```json
{
  "type": "unsubscribe",
  "topic": "topic_name"
}
```

The server will respond with an acknowledgment:

```json
{
  "type": "unsubscribe_ack",
  "topic": "topic_name",
  "success": true
}
```

### Getting Subscribed Topics

To get a list of topics you're subscribed to, send a JSON message:

```json
{
  "type": "get_topics"
}
```

The server will respond with a list of topics:

```json
{
  "type": "topics",
  "topics": ["topic1", "topic2", "topic3"]
}
```

## Heartbeat and Ping Response Management

The WebSocket protocol doesn't provide built-in mechanisms to detect if a connection is still alive or has silently failed. To address this, DogePlus Backend implements a heartbeat mechanism that ensures connections remain healthy and allows for the cleanup of stale connections.

### How Heartbeats Work

1. **Server-Initiated Heartbeats**: The server sends periodic heartbeat messages to all connected clients.
2. **Client Responses**: Clients must respond to these heartbeats to indicate they're still active.
3. **Connection Monitoring**: The server monitors these responses to detect and clean up stale connections.

### Heartbeat Message Format

The server sends heartbeat messages in the following format:

```json
{
  "type": "heartbeat",
  "timestamp": 1616161616
}
```

Where:
- `type`: Always "heartbeat" to identify the message type
- `timestamp`: The server's current Unix timestamp when the heartbeat was sent

### Client Response Format

Clients should respond to heartbeat messages with a similar format:

```json
{
  "type": "heartbeat",
  "timestamp": 1616161616
}
```

Where:
- `type`: Always "heartbeat" to identify the message type
- `timestamp`: The client's current timestamp (typically `Date.now()` in JavaScript)

### Implementation Details

#### Server-Side

The server implements heartbeats through the `ConnectionManager`:

1. **Heartbeat Loop**: A background goroutine (`heartbeatLoop`) runs continuously, sending heartbeats at regular intervals (defined by `heartbeatTick`, typically 30 seconds).
2. **Sending Heartbeats**: The `sendHeartbeats` method sends a heartbeat message to all connected clients.
3. **Connection Monitoring**: Another background goroutine (`monitorConnections`) periodically checks for stale connections.
4. **Stale Connection Cleanup**: The `cleanStaleConnections` method removes connections that haven't had activity in the last 10 minutes.

#### Client-Side

Clients should implement heartbeat handling in their message processing logic:

```javascript
socket.onmessage = function(event) {
  const message = JSON.parse(event.data);

  if (message.type === 'heartbeat') {
    // Respond to heartbeat
    socket.send(JSON.stringify({
      type: 'heartbeat',
      timestamp: Date.now()
    }));
  }

  // Handle other message types...
};
```

### Benefits of Heartbeat Mechanism

1. **Connection Health Monitoring**: Detects connections that have silently failed.
2. **Resource Management**: Allows the server to clean up resources associated with stale connections.
3. **Network Issue Detection**: Helps identify network issues that might not trigger WebSocket close events.

### Best Practices for Heartbeat Handling

1. **Always Respond Promptly**: Clients should respond to heartbeats as soon as they receive them.
2. **Handle Reconnection**: If a client doesn't receive heartbeats for an extended period, it should attempt to reconnect.
3. **Implement Timeouts**: Clients can implement their own timeout logic to detect server unavailability.

### Example: Client-Side Heartbeat Monitoring

```javascript
function connectWebSocket() {
  const socket = new WebSocket('ws://your-server/api/v1/ws/');
  let lastHeartbeat = Date.now();

  // Check for missing heartbeats
  const heartbeatMonitor = setInterval(() => {
    const now = Date.now();
    // If no heartbeat received in 2 minutes, reconnect
    if (now - lastHeartbeat > 120000) {
      console.log('No heartbeat received, reconnecting...');
      clearInterval(heartbeatMonitor);
      socket.close();
      // Reconnect will be triggered by onclose handler
    }
  }, 30000);

  socket.onmessage = function(event) {
    const message = JSON.parse(event.data);

    if (message.type === 'heartbeat') {
      lastHeartbeat = Date.now();
      socket.send(JSON.stringify({
        type: 'heartbeat',
        timestamp: lastHeartbeat
      }));
    }

    // Handle other message types...
  };

  socket.onclose = function() {
    clearInterval(heartbeatMonitor);
    setTimeout(connectWebSocket, 3000);
  };

  return socket;
}
```

## Available Topics

The following topics are available for subscription:

### `task_completion_update`

Subscribe to this topic to receive updates about event tasks. This includes when an event task is updated, such as when its status changes.

Example message:

```json
{
  "Result": "Event Task Updated",
  "Events": {
    "uuid": "123e4567-e89b-12d3-a456-426614174000",
    "event_number": 123,
    "event_date": "2023-01-01T12:00:00Z",
    "central_id": "ABC123",
    "priority": 1,
    "title": "Task Title",
    "description": "Task Description",
    "role": "Role",
    "status": "done",
    "modified_by": "User",
    "ip_address": "127.0.0.1",
    "timestamp": "2023-01-01T12:30:00Z",
    "escalation_level": "allarme"
  }
}
```

### `event_updates`

Subscribe to this topic to receive updates about events, including when new overviews are added.

Example message:

```json
{
  "message": "Overview added successfully",
  "data": {
    "event_number": 123,
    "central_id": "ABC123",
    "type": "fire",
    "level": "allarme",
    "incident_level": "high",
    "timestamp": "2023-01-01T12:00:00Z"
  }
}
```

### `central_[ID]`

Subscribe to this topic to receive updates about events for a specific central ID. Replace `[ID]` with the actual central ID you're interested in (e.g., `central_ABC123`).

The message format is the same as for the `event_updates` topic.

### `task_completion_map_update`

Subscribe to this topic to receive real-time updates about task completion progress. This includes when an event's tasks are updated, added, or deleted.

For a specific event update:
```json
{
  "type": "task_completion_update",
  "data": {
    "event_number": 123,
    "info": {
      "Completed": 5,
      "Total": 10
    }
  }
}
```

For a full map update (e.g., when an event is deleted):
```json
{
  "type": "task_completion_update",
  "data": {
    "123": {
      "Completed": 5,
      "Total": 10
    },
    "456": {
      "Completed": 2,
      "Total": 8
    }
  }
}
```

## Client-Side Implementation

### Basic Connection Example

```javascript
// Connect to the WebSocket server
const socket = new WebSocket('ws://your-server/api/v1/ws/');

// Connection opened
socket.onopen = function(event) {
  console.log('Connected to WebSocket server');
};

// Listen for messages
socket.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('Message from server:', message);
};

// Connection closed
socket.onclose = function(event) {
  console.log('Disconnected from WebSocket server');
};

// Connection error
socket.onerror = function(error) {
  console.error('WebSocket error:', error);
};
```

### Subscribing to Topics Example

```javascript
// Connect to the WebSocket server
const socket = new WebSocket('ws://your-server/api/v1/ws/');

// Connection opened
socket.onopen = function(event) {
  console.log('Connected to WebSocket server');

  // Subscribe to task completion updates
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'task_completion_update'
  }));

  // Subscribe to event updates
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'event_updates'
  }));

  // Subscribe to updates for a specific central
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'central_ABC123'
  }));
};

// Listen for messages
socket.onmessage = function(event) {
  const message = JSON.parse(event.data);

  // Handle different message types
  if (message.type === 'subscribe_ack') {
    console.log(`Subscribed to ${message.topic}`);
  } else if (message.type === 'unsubscribe_ack') {
    console.log(`Unsubscribed from ${message.topic}`);
  } else if (message.type === 'topics') {
    console.log('Subscribed topics:', message.topics);
  } else if (message.type === 'heartbeat') {
    // Respond to heartbeat
    socket.send(JSON.stringify({
      type: 'heartbeat',
      timestamp: Date.now()
    }));
  } else if (message.Result === 'Event Task Updated') {
    console.log('Task update:', message.Events);
    // Update UI with task information
  } else if (message.message === 'Overview added successfully') {
    console.log('Event update:', message.data);
    // Update UI with event information
  } else if (message.type === 'task_completion_update') {
    console.log('Task completion update:', message.data);
    // Update UI with task completion information
  }
};
```

### Reconnection Handling Example

```javascript
function connectWebSocket() {
  const socket = new WebSocket('ws://your-server/api/v1/ws/');

  socket.onopen = function(event) {
    console.log('Connected to WebSocket server');
    // Subscribe to topics
    // ...
  };

  socket.onclose = function(event) {
    console.log('Disconnected from WebSocket server');
    // Reconnect after a delay
    setTimeout(connectWebSocket, 3000);
  };

  socket.onerror = function(error) {
    console.error('WebSocket error:', error);
  };

  socket.onmessage = function(event) {
    // Handle messages
    // ...
  };

  return socket;
}

// Initial connection
let socket = connectWebSocket();
```

## Server-Side Implementation

The server-side implementation consists of several key components:

### WebSocket Service (`ws.Service`)

The `Service` struct represents a WebSocket connection and implements the `Broadcaster` interface. It provides methods for:

- Connecting and disconnecting
- Sending messages
- Checking connection status
- Managing topic subscriptions

```go
// Service is a type that represents a WebSocket connection service.
// It implements the broadcast.Broadcaster interface.
type Service struct {
    Conn         *websocket.Conn // WebSocket connection
    Id           string          // Client identifier
    connected    bool            // Connection state
    lastActivity time.Time       // Time of last activity
    topics       []string        // Topics the client is subscribed to
    mu           sync.Mutex      // Mutex for thread safety
}
```

### Connection Manager (`broadcast.ConnectionManager`)

The `ConnectionManager` struct manages WebSocket connections and provides methods for:

- Adding and removing clients
- Broadcasting messages to all clients
- Broadcasting messages to clients subscribed to a specific topic
- Sending heartbeats
- Monitoring connections and cleaning up stale ones

```go
// ConnectionManager handles the management of client connections for real-time broadcasting.
// It maintains a map of connected broadcasters and provides methods to add, remove, and broadcast messages to clients.
type ConnectionManager struct {
    Clients       map[string]Broadcaster // Map of client ID to broadcaster
    lock          sync.RWMutex
    heartbeatTick time.Duration // Interval for sending heartbeats
    done          chan struct{} // Channel to signal shutdown
}
```

### WebSocket Handlers

The `WsUpgrader` and `WsHandler` functions handle WebSocket connection upgrades and message processing:

- `WsUpgrader`: Generates a unique client ID and upgrades the connection
- `WsHandler`: Creates a new `Service` instance or updates an existing one, and handles WebSocket messages

```go
// WsUpgrader is a middleware that upgrades HTTP connections to WebSocket connections.
func WsUpgrader(manager *broadcast.ConnectionManager) fiber.Handler {
    // Implementation...
}

// WsHandler handles WebSocket connections.
func WsHandler(cm *broadcast.ConnectionManager) fiber.Handler {
    // Implementation...
}
```

### Broadcasting Messages

To broadcast a message to all clients:

```go
// Broadcast sends a message to all connected clients.
func (cm *ConnectionManager) Broadcast(message []byte) {
    // Implementation...
}
```

To broadcast a message to clients subscribed to a specific topic:

```go
// BroadcastToTopic sends a message to all connected clients subscribed to the specified topic.
func (cm *ConnectionManager) BroadcastToTopic(topic string, message []byte) {
    // Implementation...
}
```

## Best Practices and Considerations

### Connection Management

1. **Unique Client IDs**: Each client should have a unique ID to ensure proper connection management.
2. **Reconnection Handling**: Implement reconnection logic on the client side to handle disconnections.
3. **Heartbeats**: Use heartbeats to detect and clean up stale connections.

### Message Handling

1. **Structured Messages**: Use structured JSON messages for communication.
2. **Error Handling**: Handle errors gracefully and provide meaningful error messages.
3. **Message Validation**: Validate incoming messages to prevent security issues.

### Performance Considerations

1. **Topic-Based Subscriptions**: Use topic-based subscriptions to reduce unnecessary message delivery.
2. **Message Size**: Keep messages small to reduce bandwidth usage.
3. **Connection Limits**: Be aware of connection limits and implement proper scaling strategies.

### Security Considerations

1. **Authentication**: Implement proper authentication for WebSocket connections.
2. **Authorization**: Ensure clients can only subscribe to topics they have access to.
3. **Input Validation**: Validate all input to prevent injection attacks.

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if the WebSocket server is running and accessible.
2. **Connection Closed Unexpectedly**: Check for network issues or server-side errors.
3. **Messages Not Received**: Verify that the client is subscribed to the correct topics.

### Debugging Tips

1. **Enable Logging**: Enable detailed logging to track WebSocket connections and messages.
2. **Check Network Traffic**: Use browser developer tools to inspect WebSocket traffic.
3. **Test with Simple Clients**: Use simple WebSocket clients (like wscat) to test the server.

### Example Debugging Code

```javascript
// Enable detailed logging
const socket = new WebSocket('ws://your-server/api/v1/ws/');

socket.onopen = function(event) {
  console.log('Connected to WebSocket server', event);
};

socket.onmessage = function(event) {
  console.log('Message received:', event.data);
  try {
    const message = JSON.parse(event.data);
    console.log('Parsed message:', message);
  } catch (error) {
    console.error('Error parsing message:', error);
  }
};

socket.onclose = function(event) {
  console.log('Connection closed:', event.code, event.reason);
};

socket.onerror = function(error) {
  console.error('WebSocket error:', error);
};
```

---

This documentation provides a comprehensive guide to using the WebSocket functionality in the DogePlus Backend project. By following these guidelines, you can implement real-time communication between your server and clients, enhancing the user experience and reducing the need for polling or page refreshes.
