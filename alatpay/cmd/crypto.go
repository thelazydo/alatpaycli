package cmd

import (
	"alatpay/config"
	"alatpay/internal/crypto"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cryptoKey string
	cryptoIV  string
)

// cryptoCmd represents the utility wrapper for testing payload encryption
var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "Utility commands for AlatPay payload encryption, decryption, and signature verification",
	Long:  `Developer utilities to verify your own backend implementations of AlatPay's AES/CBC/PKCS5Padding encryption and HMAC SHA512 signatures.`,
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt [plaintext json]",
	Short: "Encrypt a JSON payload using AlatPay's AES algorithm",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		payload := strings.Join(args, " ")
		res, err := crypto.Encrypt(payload, cryptoKey, cryptoIV)
		if err != nil {
			color.Red("[!] Encryption failed: %v\n", err)
			return
		}

		fmt.Printf("Base64 Encrypted Ciphertext:\n")
		color.Green(res)
	},
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt [base64_ciphertext]",
	Short: "Decrypt an AlatPay payload using AES",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		res, err := crypto.Decrypt(args[0], cryptoKey, cryptoIV)
		if err != nil {
			color.Red("[!] Decryption failed: %v\n", err)
			return
		}

		fmt.Printf("Decrypted Plaintext:\n")
		color.Green(res)
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify-signature [payload] [signature]",
	Short: "Verify an AlatPay HMAC SHA512 signature against a raw JSON payload",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.WebhookSecret == "" {
			color.Red("Error: Please run 'alatpay auth' to set your webhook secret first before verifying signatures.")
			return
		}

		payload := args[0]
		providedSig := args[1]

		mac := hmac.New(sha512.New, []byte(cfg.WebhookSecret))
		mac.Write([]byte(payload))
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if hmac.Equal([]byte(providedSig), []byte(expectedMAC)) {
			color.Green("\n[✓] Signature VERIFIED locally!")
		} else {
			color.Red("\n[✗] Signature MISMATCH.")
			fmt.Printf("Expected: %s\nProvided: %s\n", expectedMAC, providedSig)
		}
	},
}

func init() {
	rootCmd.AddCommand(cryptoCmd)
	cryptoCmd.AddCommand(encryptCmd)
	cryptoCmd.AddCommand(decryptCmd)
	cryptoCmd.AddCommand(verifyCmd)

	// Add persistent flags for key/iv defaulting to AlatPay sandbox keys
	cryptoCmd.PersistentFlags().StringVarP(&cryptoKey, "key", "k", ")KCSWITHC%^$$%@H", "AES Encryption Key (16 bytes)")
	cryptoCmd.PersistentFlags().StringVarP(&cryptoIV, "iv", "i", "#$%#^%KCSWITC945", "AES Initialization Vector (16 bytes)")
}
