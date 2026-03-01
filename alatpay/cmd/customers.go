package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	customerID    string
	customerEmail string
	customerName  string
	customerPhone string
)

func buildCustomerPayload() []byte {
	payload := map[string]string{}
	if customerEmail != "" {
		payload["email"] = customerEmail
	}
	if customerName != "" {
		payload["name"] = customerName
	}
	if customerPhone != "" {
		payload["phone"] = customerPhone
	}
	data, _ := json.Marshal(payload)
	return data
}

// customersCmd represents the customers command
var customersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage AlatPay customers",
	Long:  `Create, retrieve, update, and delete customers on the AlatPay platform.`,
}

// customersCreateCmd represents creating a customer
var customersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new customer",
	Run: func(cmd *cobra.Command, args []string) {
		payload := buildCustomerPayload()
		cmd.Printf("Creating customer with payload: %s\n", string(payload))

		// Typically you'd use the auth token and make a real HTTP request here
		// Example framework for standard REST payload
		req, _ := http.NewRequest("POST", "https://api.alatpay.ng/v1/customers", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/json")
		// req.Header.Add("Authorization", "Bearer " + token)

		// For now we mock the functionality to match the structure requirements
		cmd.Println(color.GreenString("[✓] Customer created successfully (mock)"))
		cmd.Println(color.WhiteString(string(payload)))
	},
}

// customersGetCmd represents retrieving a customer
var customersGetCmd = &cobra.Command{
	Use:   "get [customer_id]",
	Short: "Retrieve a customer by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		fmt.Printf("Retrieving customer %s...\n", id)

		// Example GET request logic
		// req, _ := http.NewRequest("GET", "https://api.alatpay.ng/v1/customers/"+id, nil)

		fmt.Println(color.GreenString("[✓] Customer retrieved successfully (mock)"))
		fmt.Printf(`{
  "id": "%s",
  "email": "user@example.com",
  "name": "Jane Doe",
  "created_at": "2026-02-28T12:00:00Z"
}`+"\n", id)
	},
}

// customersUpdateCmd represents updating a customer
var customersUpdateCmd = &cobra.Command{
	Use:   "update [customer_id]",
	Short: "Update a customer by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		payload := buildCustomerPayload()
		fmt.Printf("Updating customer %s with payload: %s\n", id, string(payload))

		// Example POST/PUT request logic
		// req, _ := http.NewRequest("PUT", "https://api.alatpay.ng/v1/customers/"+id, bytes.NewBuffer(payload))

		fmt.Println(color.GreenString("[✓] Customer updated successfully (mock)"))
	},
}

// customersListCmd represents listing customers
var customersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing customers...")

		// Example GET request logic
		// req, _ := http.NewRequest("GET", "https://api.alatpay.ng/v1/customers", nil)

		fmt.Println(color.GreenString("[✓] Customers retrieved successfully (mock)"))
		fmt.Println(`[
  {
    "id": "cus_12345",
    "email": "jane@example.com",
    "name": "Jane Doe"
  },
  {
    "id": "cus_67890",
    "email": "john@example.com",
    "name": "John Doe"
  }
]`)
	},
}

func init() {
	rootCmd.AddCommand(customersCmd)

	customersCmd.AddCommand(customersCreateCmd)
	customersCmd.AddCommand(customersGetCmd)
	customersCmd.AddCommand(customersUpdateCmd)
	customersCmd.AddCommand(customersListCmd)

	// Flags for create & update
	customersCreateCmd.Flags().StringVar(&customerEmail, "email", "", "The customer's email address")
	customersCreateCmd.Flags().StringVar(&customerName, "name", "", "The customer's full name")
	customersCreateCmd.Flags().StringVar(&customerPhone, "phone", "", "The customer's phone number")

	customersUpdateCmd.Flags().StringVar(&customerEmail, "email", "", "The customer's email address")
	customersUpdateCmd.Flags().StringVar(&customerName, "name", "", "The customer's full name")
	customersUpdateCmd.Flags().StringVar(&customerPhone, "phone", "", "The customer's phone number")
}
