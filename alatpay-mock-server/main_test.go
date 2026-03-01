package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHandleWebhook(t *testing.T) {
	// Create a test payload
	payload := `{"event": "payment.successful", "data": {"amount": 5000}}`
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-alatpay-signature", "dummy-signature")

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Call the handler directly
	handleWebhook(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	expected := `{"status": "forwarded to CLI clients"}`
	if !strings.Contains(rr.Body.String(), "forwarded") {
		t.Errorf("handler returned unexpected body: got %v want to contain %v", rr.Body.String(), expected)
	}
}

func TestWebSocketRelay(t *testing.T) {
	// Start a test server with our WS handler
	server := httptest.NewServer(http.HandlerFunc(handleWS))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect a test client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
	}
	defer ws.Close()

	// Allow connection to register
	time.Sleep(100 * time.Millisecond)

	// Verify client is registered in the map
	clientsMu.Lock()
	if len(clients) != 1 {
		t.Errorf("expected 1 connected client, got %d", len(clients))
	}
	clientsMu.Unlock()

	// Broadcast an event to test the fan-out
	testEvent := Event{
		Timestamp: time.Now().Unix(),
		Path:      "/webhook",
		Method:    "POST",
		Payload:   json.RawMessage(`{"test":true}`),
		Headers:   map[string]string{"X-Test": "true"},
	}

	go broadcast(testEvent)

	// Read the broadcasted message from the connected client
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read broadcast message: %v", err)
	}

	var receivedEvent Event
	if err := json.Unmarshal(msg, &receivedEvent); err != nil {
		t.Fatalf("failed to unmarshal received event: %v", err)
	}

	if receivedEvent.Path != "/webhook" || receivedEvent.Method != "POST" {
		t.Errorf("received incorrect event data. Path: %s, Method: %s", receivedEvent.Path, receivedEvent.Method)
	}
}
