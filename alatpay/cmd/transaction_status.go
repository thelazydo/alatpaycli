package cmd

import (
	"alatpay/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// statusCmd represents the transaction status command
var statusCmd = &cobra.Command{
	Use:   "status [transactionId]",
	Short: "Check the status of an AlatPay transaction",
	Long: `Queries the AlatPay API to retrieve the real-time status 
and details of a specific transaction by its ID or reference.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		cfg, err := config.Load()
		if err != nil || cfg.APIKey == "" {
			log.Fatalf("Authentication required. Please run 'alatpay auth' first.")
		}

		// Example endpoint to get a single transaction (assumed from research)
		url := fmt.Sprintf("%s/api/v1/transactions/%s", alatPayBaseURL, transactionID)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))

		client := &http.Client{Timeout: 10 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error making request: %v", err)
		}
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)

		fmt.Println("\n" + color.CyanString("--- Transaction Status Response ---"))
		fmt.Printf("Status: %s\n\n", res.Status)

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			fmt.Println(color.GreenString(prettyJSON.String()))
		} else {
			fmt.Println(color.WhiteString(string(body)))
		}
	},
}

func init() {
	transactionCmd.AddCommand(statusCmd)
}
