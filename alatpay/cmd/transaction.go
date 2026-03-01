package cmd

import (
	"github.com/spf13/cobra"
)

// transactionCmd represents the transaction command
var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Manage AlatPay transactions",
	Long: `Create new transactions or check the status of existing ones 
using the AlatPay API. Requires authentication via 'alatpay auth'.`,
}

func init() {
	rootCmd.AddCommand(transactionCmd)
}
