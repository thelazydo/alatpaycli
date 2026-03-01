package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"alatpay/config"

	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with your AlatPay credentials",
	Long: `Interactive command to set up your AlatPay CLI profile securely.
It will ask for your API Key, Business ID, and Webhook Secret Hash 
and save them to ~/.alatpay/config.json for global CLI usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter AlatPay API Key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		fmt.Print("Enter AlatPay Business ID: ")
		businessID, _ := reader.ReadString('\n')
		businessID = strings.TrimSpace(businessID)

		fmt.Print("Enter AlatPay Vendor ID: ")
		vendorID, _ := reader.ReadString('\n')
		vendorID = strings.TrimSpace(vendorID)

		fmt.Print("Enter AlatPay Webhook Secret (optional, for signature verification): ")
		webhookSecret, _ := reader.ReadString('\n')
		webhookSecret = strings.TrimSpace(webhookSecret)

		cfg := &config.Config{
			APIKey:        apiKey,
			BusinessID:    businessID,
			VendorId:      vendorID,
			WebhookSecret: webhookSecret,
		}

		err := config.Save(cfg)
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully saved AlatPay configuration.")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
