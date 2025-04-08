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

### `task_completion_update`

Subscribe to this topic to receive updates about event tasks. This includes when an event task is updated, such as when its status changes. This topic is used by the `UpdateEventTask` function.

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

Subscribe to this topic to receive updates about events, including when new overviews are added. This topic is used by the `PostNewOverview` function.

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

Subscribe to this topic to receive updates about events for a specific central ID. Replace `[ID]` with the actual central ID you're interested in (e.g., `central_ABC123`). This topic is used by the `PostNewOverview` function.

The message format is the same as for the `event_updates` topic.

### `task_completion_map_update`

Subscribe to this topic to receive real-time updates about task completion progress. This includes when an event's tasks are updated, added, or deleted. This topic is used by the TaskCompletionMap methods: `UpdateEventStatus`, `AddMultipleNotDoneTasks`, `AddNewEvent`, and `DeleteEvent`.

The message format depends on whether a specific event is updated or the entire map is broadcast:

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

## Implementation Details

The topic-based subscription system is implemented using the following components:

1. The `Broadcaster` interface in `broadcast/broadcaster.go` defines methods for subscribing and unsubscribing to topics.
2. The `Service` struct in `ws/service.go` implements the `Broadcaster` interface and stores the topics a client is subscribed to.
3. The `ConnectionManager` in `broadcast/manager.go` provides a method for broadcasting messages to clients subscribed to a specific topic.
4. The `WsHandler` in `handlers/realtime.go` handles WebSocket messages for topic subscription and unsubscription.

## Examples

### Example: Subscribing to Task Completion Updates

To receive updates about task completions, you can subscribe to the `task_completion_update` topic:

```javascript
// Connect to the WebSocket server
const socket = new WebSocket('ws://your-server/ws');

socket.onopen = function() {
  // Subscribe to task completion updates
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'task_completion_update'
  }));
};

socket.onmessage = function(event) {
  const message = JSON.parse(event.data);

  // Handle different message types
  if (message.type === 'subscribe_ack') {
    console.log(`Subscribed to ${message.topic}`);
  } else if (message.Result === 'Event Task Updated') {
    console.log('Received task completion update:', message.Events);
  }
};
```

This way, you'll receive updates for all task completions in the system.

### Example: Subscribing to Event Updates

To receive updates about all events, including when new overviews are added, you can subscribe to the `event_updates` topic:

```javascript
// Connect to the WebSocket server
const socket = new WebSocket('ws://your-server/ws');

socket.onopen = function() {
  // Subscribe to event updates
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'event_updates'
  }));
};

socket.onmessage = function(event) {
  const message = JSON.parse(event.data);

  // Handle different message types
  if (message.type === 'subscribe_ack') {
    console.log(`Subscribed to ${message.topic}`);
  } else if (message.message === 'Overview added successfully') {
    console.log('Received event update:', message.data);
  }
};
```

This way, you'll receive updates for all events in the system.

### Example: Subscribing to Updates for a Specific Central

If you're only interested in updates for a specific central (e.g., "ABC123"), you can subscribe to the `central_ABC123` topic:

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
  } else if (message.message === 'Overview added successfully') {
    console.log('Received update for central ABC123:', message.data);
  }
};
```

This way, you'll only receive updates for events related to central ABC123, reducing bandwidth and CPU usage.

### Example: Subscribing to Task Completion Map Updates

To receive real-time updates about task completion progress, you can subscribe to the `task_completion_map_update` topic:

```javascript
// Connect to the WebSocket server
const socket = new WebSocket('ws://your-server/ws');

socket.onopen = function() {
  // Subscribe to task completion map updates
  socket.send(JSON.stringify({
    type: 'subscribe',
    topic: 'task_completion_map_update'
  }));
};

socket.onmessage = function(event) {
  const message = JSON.parse(event.data);

  // Handle different message types
  if (message.type === 'subscribe_ack') {
    console.log(`Subscribed to ${message.topic}`);
  } else if (message.type === 'task_completion_update') {
    // Check if this is a specific event update or a full map update
    if (message.data.event_number) {
      // This is a specific event update
      const eventNumber = message.data.event_number;
      const completed = message.data.info.Completed;
      const total = message.data.info.Total;
      const percentage = (completed / total) * 100;

      console.log(`Event ${eventNumber}: ${completed}/${total} tasks completed (${percentage.toFixed(2)}%)`);
      // Update UI for this specific event
      updateEventProgress(eventNumber, completed, total);
    } else {
      // This is a full map update
      console.log('Received full task completion map update:', message.data);
      // Update UI for all events
      updateAllEventsProgress(message.data);
    }
  }
};

// Function to update the UI for a specific event
function updateEventProgress(eventNumber, completed, total) {
  // Update progress bars, counters, etc. for this event
  const progressBar = document.getElementById(`progress-${eventNumber}`);
  if (progressBar) {
    const percentage = (completed / total) * 100;
    progressBar.style.width = `${percentage}%`;
    progressBar.setAttribute('aria-valuenow', percentage);
    document.getElementById(`progress-text-${eventNumber}`).textContent = `${completed}/${total}`;
  }
}

// Function to update the UI for all events
function updateAllEventsProgress(data) {
  // Iterate through all events in the data and update their UI
  for (const [eventNumber, info] of Object.entries(data)) {
    updateEventProgress(parseInt(eventNumber), info.Completed, info.Total);
  }
}
```

This way, you'll receive real-time updates about task completion progress for all events in the system, allowing you to display accurate progress information to users without requiring page refreshes.
