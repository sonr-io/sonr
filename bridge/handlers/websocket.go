package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// TaskStatusMessage represents a WebSocket message for task status updates
type TaskStatusMessage struct {
	TaskID   string    `json:"task_id"`
	Status   string    `json:"status"`
	Progress int       `json:"progress,omitempty"`
	Data     any       `json:"data,omitempty"`
	Error    string    `json:"error,omitempty"`
	Time     time.Time `json:"timestamp"`
}

// ConnectionManager manages WebSocket connections for task status broadcasting
type ConnectionManager struct {
	connections map[string]map[*websocket.Conn]bool // taskID -> connections
	mutex       sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]map[*websocket.Conn]bool),
		mutex:       sync.RWMutex{},
	}
}

// AddConnection adds a WebSocket connection for a specific task ID
func (cm *ConnectionManager) AddConnection(taskID string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.connections[taskID] == nil {
		cm.connections[taskID] = make(map[*websocket.Conn]bool)
	}
	cm.connections[taskID][conn] = true
	log.Printf("WebSocket connection added for task: %s", taskID)
}

// RemoveConnection removes a WebSocket connection
func (cm *ConnectionManager) RemoveConnection(taskID string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if connections, exists := cm.connections[taskID]; exists {
		delete(connections, conn)
		if len(connections) == 0 {
			delete(cm.connections, taskID)
		}
	}
	log.Printf("WebSocket connection removed for task: %s", taskID)
}

// BroadcastToTask broadcasts a message to all connections listening to a specific task
func (cm *ConnectionManager) BroadcastToTask(taskID string, message TaskStatusMessage) {
	cm.mutex.RLock()
	connections, exists := cm.connections[taskID]
	cm.mutex.RUnlock()

	if !exists {
		return
	}

	for conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting to WebSocket: %v", err)
			cm.RemoveConnection(taskID, conn)
			conn.Close()
		}
	}
}

// SSEManager manages Server-Sent Event connections for task status streaming
type SSEManager struct {
	connections map[string]map[chan TaskStatusMessage]bool // taskID -> channels
	mutex       sync.RWMutex
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	return &SSEManager{
		connections: make(map[string]map[chan TaskStatusMessage]bool),
		mutex:       sync.RWMutex{},
	}
}

// AddSSEConnection adds an SSE channel for a specific task ID
func (sm *SSEManager) AddSSEConnection(taskID string, ch chan TaskStatusMessage) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.connections[taskID] == nil {
		sm.connections[taskID] = make(map[chan TaskStatusMessage]bool)
	}
	sm.connections[taskID][ch] = true
	log.Printf("SSE connection added for task: %s", taskID)
}

// RemoveSSEConnection removes an SSE channel
func (sm *SSEManager) RemoveSSEConnection(taskID string, ch chan TaskStatusMessage) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if channels, exists := sm.connections[taskID]; exists {
		delete(channels, ch)
		if len(channels) == 0 {
			delete(sm.connections, taskID)
		}
	}
	close(ch)
	log.Printf("SSE connection removed for task: %s", taskID)
}

// BroadcastToSSE broadcasts a message to all SSE connections listening to a specific task
func (sm *SSEManager) BroadcastToSSE(taskID string, message TaskStatusMessage) {
	sm.mutex.RLock()
	channels, exists := sm.connections[taskID]
	sm.mutex.RUnlock()

	if !exists {
		return
	}

	for ch := range channels {
		select {
		case ch <- message:
			// Message sent successfully
		default:
			// Channel is blocked, remove it
			log.Printf("SSE channel blocked, removing for task: %s", taskID)
			sm.RemoveSSEConnection(taskID, ch)
		}
	}
}

// WebSocketHandler handles WebSocket connections for real-time task status updates
func WebSocketHandler(
	upgrader *websocket.Upgrader,
	connectionManager *ConnectionManager,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		taskID := c.Param("task_id")
		if taskID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID is required"})
		}

		// Upgrade HTTP connection to WebSocket
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return err
		}
		defer ws.Close()

		// Add connection to manager
		connectionManager.AddConnection(taskID, ws)
		defer connectionManager.RemoveConnection(taskID, ws)

		// Send initial connection confirmation
		initialMessage := TaskStatusMessage{
			TaskID: taskID,
			Status: "connected",
			Time:   time.Now(),
		}
		if err := ws.WriteJSON(initialMessage); err != nil {
			log.Printf("Error sending initial message: %v", err)
			return err
		}

		// Keep connection alive and handle incoming messages
		for {
			// Read message from client (ping/pong for keepalive)
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}
			// Echo back a pong message to keep connection alive
			pongMessage := TaskStatusMessage{
				TaskID: taskID,
				Status: "pong",
				Time:   time.Now(),
			}
			if err := ws.WriteJSON(pongMessage); err != nil {
				log.Printf("Error sending pong: %v", err)
				break
			}
		}

		return nil
	}
}

// SSEHandler handles Server-Sent Events for streaming task status updates
func SSEHandler(sseManager *SSEManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		taskID := c.Param("task_id")
		if taskID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task ID is required"})
		}

		// Set SSE headers
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

		// Create a channel for this SSE connection
		messageCh := make(chan TaskStatusMessage, 10)
		sseManager.AddSSEConnection(taskID, messageCh)
		defer sseManager.RemoveSSEConnection(taskID, messageCh)

		// Send initial connection message
		initialMessage := TaskStatusMessage{
			TaskID: taskID,
			Status: "connected",
			Time:   time.Now(),
		}
		fmt.Fprintf(
			c.Response().Writer,
			"data: {\"task_id\":\"%s\",\"status\":\"connected\",\"timestamp\":\"%s\"}\n\n",
			taskID,
			initialMessage.Time.Format(time.RFC3339),
		)
		c.Response().Flush()

		// Keep connection alive and send messages
		for {
			select {
			case message, ok := <-messageCh:
				if !ok {
					return nil
				}

				// Format message as SSE data
				sseData := fmt.Sprintf(
					"data: {\"task_id\":\"%s\",\"status\":\"%s\",\"progress\":%d,\"timestamp\":\"%s\"}",
					message.TaskID,
					message.Status,
					message.Progress,
					message.Time.Format(time.RFC3339),
				)

				if message.Error != "" {
					sseData = fmt.Sprintf(
						"data: {\"task_id\":\"%s\",\"status\":\"%s\",\"error\":\"%s\",\"timestamp\":\"%s\"}",
						message.TaskID,
						message.Status,
						message.Error,
						message.Time.Format(time.RFC3339),
					)
				}

				fmt.Fprintf(c.Response().Writer, "%s\n\n", sseData)
				c.Response().Flush()

			case <-c.Request().Context().Done():
				// Client disconnected
				return nil

			case <-time.After(30 * time.Second):
				// Send keepalive message every 30 seconds
				fmt.Fprintf(
					c.Response().Writer,
					"data: {\"task_id\":\"%s\",\"status\":\"keepalive\",\"timestamp\":\"%s\"}\n\n",
					taskID,
					time.Now().Format(time.RFC3339),
				)
				c.Response().Flush()
			}
		}
	}
}
