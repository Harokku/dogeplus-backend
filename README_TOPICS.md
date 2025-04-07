# Topic-Based WebSocket Subscription System

This document describes the topic-based subscription system implemented for WebSocket connections in the DogePlus Backend. This system allows clients to subscribe to specific topics and receive only the messages they're interested in, reducing bandwidth and CPU usage.

## Overview

The topic-based subscription system allows WebSocket clients to:

1. Subscribe to specific topics
2. Unsubscribe from topics
3. Retrieve a list of subscribed topics
4. Receive messages only for topics they're subscribed to

The system maintains subscriptions across reconnections, so if a client disconnects and reconnects, it will automatically receive messages for the topics it was previously subscribed to.

## Client-Side Usage

### Subscribing to a Topic

To subscribe to a topic, send a JSON message with the following format:

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

To unsubscribe from a topic, send a JSON message with the following format:

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

To get a list of topics you're subscribed to, send a JSON message with the following format:

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

## Available Topics

The following topics are currently available:

### `event_updates`

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

### `central_[ID]`

Subscribe to this topic to receive updates about events for a specific central ID. Replace `[ID]` with the actual central ID you're interested in.

Example: `central_ABC123`

## Implementation Details

The topic-based subscription system is implemented using the following components:

1. The `Broadcaster` interface in `broadcast/broadcaster.go` defines methods for subscribing and unsubscribing to topics.
2. The `Service` struct in `ws/service.go` implements the `Broadcaster` interface and stores the topics a client is subscribed to.
3. The `ConnectionManager` in `broadcast/manager.go` provides a method for broadcasting messages to clients subscribed to a specific topic.
4. The `WsHandler` in `handlers/realtime.go` handles WebSocket messages for topic subscription and unsubscription.

## Example: Subscribing to Event Updates for a Specific Central

If you're only interested in event updates for a specific central (e.g., "ABC123"), you can subscribe to the `central_ABC123` topic:

```javascript
// Connect to the WebSocket server
const socket = new WebSocket('ws://your-server/ws');

socket.onopen = function() {
  // Subscribe to updates for central ABC123
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'central_ABC123'
  }));
};

socket.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  // Handle different message types
  if (message.type === 'subscribe_ack') {
    console.log(`Subscribed to ${message.topic}`);
  } else if (message.Result === 'Event Task Updated') {
    console.log('Received event update:', message.Events);
  }
};
```

This way, you'll only receive updates for events related to central ABC123, reducing bandwidth and CPU usage.