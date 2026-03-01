package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn *websocket.Conn
}

var (
	clients   = make(map[*Client]bool)
	clientsMu sync.Mutex
)

// Event simulates a webhook wrapper emitted by the mock server
type Event struct {
	Timestamp int64             `json:"timestamp"`
	Payload   json.RawMessage   `json:"payload"`
	Headers   map[string]string `json:"headers"`
	Path      string            `json:"path"`
	Method    string            `json:"method"`
}

func main() {
	// Initialize global logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	http.HandleFunc("/ws", handleWS)
	http.HandleFunc("/webhook", handleWebhook)

	fmt.Println("Mock server started on :8081")
	fmt.Println(" - Webhooks expected at: POST http://localhost:8081/webhook")
	fmt.Println(" - WebSocket relay at: ws://localhost:8081/ws")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()
	log.Println("New CLI client connected")

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("Client disconnected")
			break
		}
	}

	clientsMu.Lock()
	delete(clients, client)
	clientsMu.Unlock()
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slog.Info("Hello: ", "headerrrs", r.Header.Get("x-alatpay-signature"))
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	header_signature := r.Header.Get("x-alatpay-signature")

	fmt.Printf("webhookSecret: %s\n", header_signature)

	mac := hmac.New(sha512.New, []byte(header_signature))
	mac.Write([]byte(bodyBytes))
	signature := hex.EncodeToString(mac.Sum(nil))
	headers["X-Alatpay-Signature"] = signature

	event := Event{
		Timestamp: time.Now().Unix(),
		Payload:   bodyBytes,
		Headers:   headers,
		Path:      r.URL.Path,
		Method:    r.Method,
	}
	fmt.Printf("Hello: %s\n", event.Payload)
	log.Printf("Received webhook payload, forwarding to %d clients...", len(clients))
	broadcast(event)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "forwarded to CLI clients"}`))
}

func broadcast(event Event) {
	eventBytes, _ := json.Marshal(event)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		err := client.conn.WriteMessage(websocket.TextMessage, eventBytes)
		if err != nil {
			log.Println("Error writing to client:", err)
			client.conn.Close()
			delete(clients, client)
		}
	}
}
