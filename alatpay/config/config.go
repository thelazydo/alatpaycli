package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the core settings for AlatPay CLI
type Config struct {
	APIKey        string `json:"api_key" mapstructure:"api_key"`
	BusinessID    string `json:"business_id" mapstructure:"business_id"`
	VendorId      string `json:"vendor_id" mapstructure:"vendor_id"`
	WebhookSecret string `json:"webhook_secret" mapstructure:"webhook_secret"`
}

// Load loads the configuration from Viper
func Load() (*Config, error) {
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable directly to decode into struct, %v", err)
	}
	return &cfg, nil
}

// Save securely saves the configuration to the user's home directory under ~/.alatpay/config.json
func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home dir: %v", err)
	}

	configDir := filepath.Join(home, ".alatpay")
	if err := os.MkdirAll(configDir, 0700); err != nil { // Secure permissions
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600) // Read/write only by user
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
