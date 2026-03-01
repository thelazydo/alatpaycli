package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"alatpay/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var triggerTarget string

// triggerCmd represents the trigger command
var triggerCmd = &cobra.Command{
	Use:   "trigger [event_type]",
	Short: "Send a mock webhook event",
	Long: `Triggers a mock webhook (e.g., payment.successful, payment.failed).
By default, it targets the local listener at http://localhost:8080, but you
can specify a different URL with the --target flag.
It automatically generates the correct AlatPay HMAC signature if a secret is configured.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eventType := "payment.successful"
		if len(args) > 0 {
			eventType = args[0]
		}

		payload := fmt.Sprintf(`{"event":"%s","data":{"amount":5000,"currency":"NGN","reference":"mock-%d","status":"%s"}}`, eventType, time.Now().Unix(), strings.Split(eventType, ".")[1])

		req, err := http.NewRequest("POST", triggerTarget, bytes.NewBuffer([]byte(payload)))
		if err != nil {
			fmt.Printf(color.RedString("[!] Failed to create request: %v\n"), err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// Apply Signature if we have a secret
		cfg, _ := config.Load()
		if cfg != nil && cfg.WebhookSecret != "" {
			mac := hmac.New(sha512.New, []byte(cfg.WebhookSecret))
			mac.Write([]byte(payload))
			signature := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("x-alatpay-signature", signature)
		}

		fmt.Printf("Triggering '%s' event to %s...\n", eventType, triggerTarget)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf(color.RedString("[✗] Failed to deliver event: %v\n"), err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf(color.GreenString("[✓] Event delivered successfully (HTTP %d)\n"), resp.StatusCode)
		} else {
			fmt.Printf(color.YellowString("[!] Event delivered but returned HTTP %d\n"), resp.StatusCode)
		}
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)
	triggerCmd.Flags().StringVarP(&triggerTarget, "target", "t", "http://localhost:8081/webhook", "Target URL to send the mock webhook to")
}
