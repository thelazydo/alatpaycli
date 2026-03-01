package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"alatpay/config"
	"alatpay/internal/crypto"
)

// Client is a wrapper around http.Client that handles AlatPay's specific
// cryptographic and authentication requirements.
type Client struct {
	BaseURL    string
	Config     *config.Config
	HTTPClient *http.Client
}

// NewClient initializes a new AlatPay API Client
func NewClient(cfg *config.Config, baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Config:  cfg,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Post sends an encrypted POST request to the AlatPay API
func (c *Client) Post(endpoint string, payload interface{}) (string, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	// Marshal payload to JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Constants for testing AlatPay decryption. In production, load from config
	// assuming static sandbox keys per RESEARCH.md
	testKey := ")KCSWITHC%^$$%@H"
	testIV := "#$%#^%KCSWITC945"

	encryptedPayload, err := crypto.Encrypt(string(jsonBytes), testKey, testIV)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(encryptedPayload)))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Accept", "application/json")

	if c.Config.APIKey != "" {
		req.Header.Add("Ocp-Apim-Subscription-Key", c.Config.APIKey)
		// Assuming we also use this as bearer for now until token refresh logic is fleshed out
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Config.APIKey))
	}
	if c.Config.VendorId != "" {
		req.Header.Add("VendorId", c.Config.VendorId)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	respStr := string(bodyBytes)

	// Intercept and handle potential 401s for token refresh
	if resp.StatusCode == 401 {
		return "", fmt.Errorf("401 Unauthorized: Bearer Token may have expired")
	}

	// AlatPay encrypts responses. Attempt to decrypt assuming it's an AES Base64 string.
	decrypted, err := crypto.Decrypt(respStr, testKey, testIV)
	if err == nil {
		return decrypted, nil
	}

	// If decrypt fails (e.g. response is plaintext JSON or an HTML error page), return raw string
	return respStr, nil
}
