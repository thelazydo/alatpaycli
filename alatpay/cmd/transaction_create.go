package cmd

import (
	"alatpay/config"
	"alatpay/internal/api"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// alatPayBaseURL is the base URL for AlatPay APIs (Sandbox by default)
const alatPayBaseURL = "https://api.alatpay.ng/rest-api"

// createCmd represents the transaction create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new bank transfer transaction (Virtual Account)",
	Long: `Initiates a new payment by generating a virtual account
using AlatPay's Bank Transfer API.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.APIKey == "" || cfg.BusinessID == "" {
			log.Fatalf("Authentication required. Please run 'alatpay auth' first.")
		}

		amount, _ := cmd.Flags().GetFloat64("amount")
		email, _ := cmd.Flags().GetString("email")
		orderID := fmt.Sprintf("CLI-ORD-%d", time.Now().Unix())

		payload := map[string]interface{}{
			"businessId": cfg.BusinessID,
			"amount":     amount,
			"currency":   "NGN",
			"orderId":    orderID,
			"customer": map[string]interface{}{
				"email":     email,
				"firstName": "Alat",
				"lastName":  "CLI",
				"phone":     "08000000000",
			},
		}

		apiClient := api.NewClient(cfg, alatPayBaseURL)

		fmt.Println(color.YellowString("Initiating transaction request via encrypted client..."))
		resString, err := apiClient.Post("/bank-transfer/api/v1/bankTransfer/virtualAccount", payload)
		if err != nil {
			log.Fatalf("Error making request: %v", err)
		}

		fmt.Println("\n" + color.CyanString("--- Transaction Create Response ---"))

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(resString), "", "  "); err == nil {
			fmt.Println(color.GreenString(prettyJSON.String()))
		} else {
			fmt.Println(color.WhiteString(resString))
		}
	},
}

func init() {
	transactionCmd.AddCommand(createCmd)

	createCmd.Flags().Float64P("amount", "a", 100.0, "Amount to charge")
	createCmd.Flags().StringP("email", "e", "test@example.com", "Customer email")
}
