package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type AlatPayEvent string

type AlatPayData struct {
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
}
type AlatPayPayload struct {
	Event AlatPayEvent `json:"event"`
	Data  AlatPayData  `json:"data"`
}

const (
	Completed  AlatPayEvent = "payment.completed"
	Pending    AlatPayEvent = "payment.pending"
	Processing AlatPayEvent = "payment.processing"
	Failed     AlatPayEvent = "payment.failed"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	http.HandleFunc("/webhook/alatpay", handleAlatPay)
	fmt.Println("Application listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleAlatPay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body AlatPayPayload

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	log.Printf("Received webhook payload from alatpay cli")
	slog.Info("body", "body", body)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "webhook processed"}`))
}
