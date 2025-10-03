package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient provides WebSocket client for CometBFT event subscription
type WebSocketClient struct {
	baseURL string
	conn    *websocket.Conn
}

// EventSubscription represents an event subscription
type EventSubscription struct {
	Query  string
	Events chan *SubscriptionEvent
	Errors chan error
	done   chan struct{}
}

// SubscriptionEvent represents an event received via subscription
type SubscriptionEvent struct {
	Query  string          `json:"query"`
	Data   EventResultData `json:"data"`
	Events []any           `json:"events,omitempty"`
}

// EventResultData represents the data part of a subscription event
type EventResultData struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      int           `json:"id"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// SubscribeParams represents subscription parameters
type SubscribeParams struct {
	Query string `json:"query"`
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(baseURL string) *WebSocketClient {
	return &WebSocketClient{
		baseURL: baseURL,
	}
}

// Connect establishes a WebSocket connection to CometBFT
func (ws *WebSocketClient) Connect(ctx context.Context) error {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(ws.baseURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL += "/websocket"

	u, err := url.Parse(wsURL)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.conn = conn
	return nil
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() error {
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

// Subscribe subscribes to events matching the given query
func (ws *WebSocketClient) Subscribe(ctx context.Context, query string) (*EventSubscription, error) {
	if ws.conn == nil {
		return nil, fmt.Errorf("WebSocket connection not established")
	}

	// Send subscription request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "subscribe",
		Params: SubscribeParams{
			Query: query,
		},
		ID: 1,
	}

	if err := ws.conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("failed to send subscription request: %w", err)
	}

	// Read subscription response
	var resp JSONRPCResponse
	if err := ws.conn.ReadJSON(&resp); err != nil {
		return nil, fmt.Errorf("failed to read subscription response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("subscription error: %s", resp.Error.Message)
	}

	// Create subscription
	subscription := &EventSubscription{
		Query:  query,
		Events: make(chan *SubscriptionEvent, 100),
		Errors: make(chan error, 10),
		done:   make(chan struct{}),
	}

	// Start listening for events
	go ws.listenForEvents(ctx, subscription)

	return subscription, nil
}

// SubscribeToNewBlocks subscribes to new block events
func (ws *WebSocketClient) SubscribeToNewBlocks(ctx context.Context) (*EventSubscription, error) {
	return ws.Subscribe(ctx, "tm.event = 'NewBlock'")
}

// SubscribeToNewBlockHeaders subscribes to new block header events
func (ws *WebSocketClient) SubscribeToNewBlockHeaders(ctx context.Context) (*EventSubscription, error) {
	return ws.Subscribe(ctx, "tm.event = 'NewBlockHeader'")
}

// SubscribeToTxEvents subscribes to transaction events
func (ws *WebSocketClient) SubscribeToTxEvents(ctx context.Context) (*EventSubscription, error) {
	return ws.Subscribe(ctx, "tm.event = 'Tx'")
}

// SubscribeToDIDEvents subscribes to DID module events
func (ws *WebSocketClient) SubscribeToDIDEvents(ctx context.Context) (*EventSubscription, error) {
	return ws.Subscribe(ctx, "did.v1.EventDIDCreated EXISTS OR did.v1.EventDIDUpdated EXISTS OR did.v1.EventDIDDeactivated EXISTS")
}

// SubscribeToDWNEvents subscribes to DWN module events
func (ws *WebSocketClient) SubscribeToDWNEvents(ctx context.Context) (*EventSubscription, error) {
	return ws.Subscribe(ctx, "dwn.v1.EventRecordWritten EXISTS OR dwn.v1.EventRecordDeleted EXISTS")
}

// SubscribeToCustomEvents subscribes to custom events with specific attributes
func (ws *WebSocketClient) SubscribeToCustomEvents(ctx context.Context, eventType, attributeKey, attributeValue string) (*EventSubscription, error) {
	query := fmt.Sprintf("%s EXISTS", eventType)
	if attributeKey != "" && attributeValue != "" {
		query += fmt.Sprintf(" AND %s.%s = '%s'", eventType, attributeKey, attributeValue)
	}
	return ws.Subscribe(ctx, query)
}

// listenForEvents listens for incoming events on the WebSocket connection
func (ws *WebSocketClient) listenForEvents(ctx context.Context, subscription *EventSubscription) {
	defer close(subscription.Events)
	defer close(subscription.Errors)

	for {
		select {
		case <-ctx.Done():
			return
		case <-subscription.done:
			return
		default:
			// Set read deadline
			if err := ws.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				subscription.Errors <- fmt.Errorf("failed to set read deadline: %w", err)
				return
			}

			var message json.RawMessage
			if err := ws.conn.ReadJSON(&message); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					return
				}
				subscription.Errors <- fmt.Errorf("failed to read WebSocket message: %w", err)
				continue
			}

			// Try to parse as JSON-RPC response first
			var resp JSONRPCResponse
			if err := json.Unmarshal(message, &resp); err == nil && resp.Result != nil {
				// This is likely an event notification
				var event SubscriptionEvent
				if eventBytes, err := json.Marshal(resp.Result); err == nil {
					if err := json.Unmarshal(eventBytes, &event); err == nil {
						event.Query = subscription.Query
						select {
						case subscription.Events <- &event:
						case <-ctx.Done():
							return
						case <-subscription.done:
							return
						}
					}
				}
			}
		}
	}
}

// Unsubscribe unsubscribes from the event subscription
func (ws *WebSocketClient) Unsubscribe(ctx context.Context, subscription *EventSubscription) error {
	if ws.conn == nil {
		return fmt.Errorf("WebSocket connection not established")
	}

	// Send unsubscribe request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "unsubscribe",
		Params: SubscribeParams{
			Query: subscription.Query,
		},
		ID: 2,
	}

	if err := ws.conn.WriteJSON(req); err != nil {
		return fmt.Errorf("failed to send unsubscribe request: %w", err)
	}

	// Signal the listening goroutine to stop
	close(subscription.done)

	return nil
}

// Close closes the event subscription
func (sub *EventSubscription) Close() {
	if sub.done != nil {
		select {
		case <-sub.done:
			// Already closed
		default:
			close(sub.done)
		}
	}
}

// WaitForEvent waits for a specific event with timeout
func (sub *EventSubscription) WaitForEvent(ctx context.Context, timeout time.Duration, eventFilter func(*SubscriptionEvent) bool) (*SubscriptionEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for event")
		case err := <-sub.Errors:
			return nil, fmt.Errorf("subscription error: %w", err)
		case event := <-sub.Events:
			if event == nil {
				return nil, fmt.Errorf("event channel closed")
			}
			if eventFilter == nil || eventFilter(event) {
				return event, nil
			}
		}
	}
}

// WaitForEventByType waits for an event of a specific type
func (sub *EventSubscription) WaitForEventByType(ctx context.Context, timeout time.Duration, eventType string) (*SubscriptionEvent, error) {
	return sub.WaitForEvent(ctx, timeout, func(event *SubscriptionEvent) bool {
		// This is a simplified check - in practice, you'd parse the event data more carefully
		eventStr := fmt.Sprintf("%v", event.Data.Value)
		return strings.Contains(eventStr, eventType)
	})
}

// GetAllEvents returns all events received so far (non-blocking)
func (sub *EventSubscription) GetAllEvents() []*SubscriptionEvent {
	var events []*SubscriptionEvent

	for {
		select {
		case event := <-sub.Events:
			if event == nil {
				return events
			}
			events = append(events, event)
		default:
			return events
		}
	}
}
