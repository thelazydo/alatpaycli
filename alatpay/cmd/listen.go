package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"alatpay/config"
	"alatpay/internal/store"
	"alatpay/ui"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	uiPort    int
	forwardTo string
	enableUI  bool
)

// Global memory store for webhooks
var webhookStore = store.NewMemoryStore()

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("evt_%x%d", b, time.Now().Unix())
}

// Event structure representing incoming WebSocket messages from the mock server
type MockEvent struct {
	Timestamp int64             `json:"timestamp"`
	Payload   json.RawMessage   `json:"payload"`
	Headers   map[string]string `json:"headers"`
	Path      string            `json:"path"`
	Method    string            `json:"method"`
}

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for AlatPay webhooks via WebSocket relay",
	Long: `Connects to an AlatPay WebSocket relay (mock server) to receive real-time events.
It pretty-prints the payloads, optionally verifies HMAC SHA512 signatures, 
can forward webhooks to another local service, and spin up a Web UI.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}

		home, _ := os.UserHomeDir()
		logPath := home + "/.alatpay/events.log"
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
		if err == nil {
			logFile.WriteString(fmt.Sprintf("[%s] ==== Alat CLI Listener Started ====\n", time.Now().Format("2006-01-02 15:04:05")))
		} else {
			fmt.Printf("Warning: Could not open %s for logging\n", logPath)
		}

		relayURL := "ws://localhost:8081/ws"

		// Channel to signal application shutdown
		stopFunc := make(chan os.Signal, 1)
		signal.Notify(stopFunc, os.Interrupt, syscall.SIGTERM)

		// Start WebSocket connection with auto-reconnect
		go func() {
			retryCount := 0
			maxBackoff := 30 * time.Second

			for {
				fmt.Printf(color.CyanString("=> Connecting to WebSocket relay at %s...\n"), relayURL)
				wsConn, _, err := websocket.DefaultDialer.Dial(relayURL, nil)
				if err != nil {
					retryCount++
					backoff := time.Duration(1<<retryCount) * time.Second
					if backoff > maxBackoff {
						backoff = maxBackoff
					}

					log.Printf(color.RedString("[!] Failed to connect: %v"), err)
					fmt.Printf(color.YellowString("[-] Retrying in %v (Attempt %d)...\n"), backoff, retryCount)

					select {
					case <-time.After(backoff):
						continue
					case <-stopFunc:
						return
					}
				}

				fmt.Println(color.GreenString("[✓] Connected! Waiting for events..."))
				retryCount = 0 // Reset on successful connection

				// Read message loop
				for {
					_, message, err := wsConn.ReadMessage()
					if err != nil {
						log.Printf(color.RedString("\n[!] Disconnected from relay server: %v\n"), err)
						wsConn.Close()
						break // Break out of read loop to trigger reconnect
					}

					var netEvent MockEvent
					if err := json.Unmarshal(message, &netEvent); err != nil {
						log.Printf("Error parsing event from relay: %v", err)
						continue
					}

					fmt.Println("\n" + color.CyanString(strings.Repeat("-", 60)))
					fmt.Printf(color.YellowString("[WEBHOOK RECEIVED] -- Path: %s\n"), netEvent.Path)

					event := store.WebhookEvent{
						ID:        generateID(),
						Method:    netEvent.Method,
						Path:      netEvent.Path,
						Headers:   netEvent.Headers,
						Body:      netEvent.Payload, // The raw byte array of just the payload portion
						Timestamp: time.Unix(netEvent.Timestamp, 0),
						Status:    "Captured",
					}

					// Signature Verification
					// The netEvent.Payload is exactly the raw JSON bytes sent in the trigger.
					// We must hash these exact bytes.
					signature := event.Headers["X-Alatpay-Signature"]
					if signature == "" {
						signature = event.Headers["x-alatpay-signature"]
					}
					fmt.Printf("webhookSecret: %s\n", cfg.WebhookSecret)
					fmt.Printf("Hello: %s\n", event.Body)

					if signature != "" && cfg.WebhookSecret != "" {
						mac := hmac.New(sha512.New, []byte(cfg.WebhookSecret))
						mac.Write([]byte(event.Body))
						expectedMAC := hex.EncodeToString(mac.Sum(nil))

						if hmac.Equal([]byte(signature), []byte(expectedMAC)) {
							event.Verified = true
							fmt.Println(color.GreenString("[✓] Signature VERIFIED successfully"))
						} else {
							fmt.Printf(color.RedString("[!] Signature MISMATCH. Expected: %s\n"), expectedMAC)
						}
					} else {
						fmt.Println(color.MagentaString("[i] Skipping signature verification\n"))
					}

					// Format & Parse JSON
					var parsedPayload map[string]interface{}
					if err := json.Unmarshal(event.Body, &parsedPayload); err == nil {
						event.Payload = parsedPayload
					}

					fmt.Printf(color.WhiteString("Headers: %v\n"), event.Headers)
					fmt.Println(color.WhiteString("Payload:\n"))

					var prettyJSON bytes.Buffer
					if json.Indent(&prettyJSON, event.Body, "", "\t") == nil {
						fmt.Println(color.GreenString(prettyJSON.String()))
					} else {
						fmt.Println(color.WhiteString(string(event.Body)))
					}

					// Save to Memory
					webhookStore.Add(event)

					if logFile != nil {
						logFile.WriteString(fmt.Sprintf("[%s] WEBHOOK CAPTURED: %s %s | Verified: %v\n", time.Now().Format("2006-01-02 15:04:05"), event.Method, event.Path, event.Verified))
					}

					// Forwarding Logic
					if forwardTo != "" {
						fmt.Printf(color.MagentaString("=> Forwarding to: %s\n"), forwardTo)
						proxyReq, _ := http.NewRequest("POST", forwardTo, bytes.NewBuffer(event.Body))
						for k, v := range event.Headers {
							proxyReq.Header.Set(k, v)
						}
						client := &http.Client{Timeout: 5 * time.Second}
						if resp, err := client.Do(proxyReq); err != nil {
							fmt.Printf(color.RedString("[✗] Forward failed: %v\n"), err)
						} else {
							fmt.Printf(color.GreenString("[✓] Forward succeeded. HTTP %d\n"), resp.StatusCode)
						}
					}

					fmt.Println(color.CyanString(strings.Repeat("-", 60)) + "\n")
				}

				// Small delay before reconnecting if we broke out of read loop
				time.Sleep(1 * time.Second)
			}
		}()

		// Web UI Server Logic (if enabled)
		var uiServer *http.Server
		if enableUI {
			uiMux := http.NewServeMux()

			// Static Vue UI
			uiMux.Handle("/", ui.Handler())

			// JSON API for UI Dashboard
			uiMux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				events := webhookStore.GetAll()
				json.NewEncoder(w).Encode(events)
			})

			uiAddr := fmt.Sprintf(":%d", uiPort)
			uiServer = &http.Server{Addr: uiAddr, Handler: uiMux}
			go func() {
				fmt.Printf(color.HiBlueString("=> UI Dashboard available at: http://localhost%s/\n"), uiAddr)
				if err := uiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("UI Server error: %s\n", err)
				}
			}()
		}

		// Wait for interrupt
		<-stopFunc

		fmt.Println("\nShutting down safely...")
		if uiServer != nil {
			uiServer.Close()
		}
		if logFile != nil {
			logFile.WriteString(fmt.Sprintf("[%s] ==== Alat CLI Listener Stopped ====\n\n", time.Now().Format("2006-01-02 15:04:05")))
			logFile.Close()
		}
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.Flags().BoolVar(&enableUI, "ui", false, "Enable the Web UI dashboard")
	listenCmd.Flags().IntVar(&uiPort, "ui-port", 8181, "Port for the Web UI dashboard")
	listenCmd.Flags().StringVar(&forwardTo, "forward-to", "", "Forward webhooks to a local server URL (e.g. http://localhost:3000/api/webhook)")
}
