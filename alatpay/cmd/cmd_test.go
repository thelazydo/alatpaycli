package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTriggerCommand(t *testing.T) {
	// 1. Create a dynamic test server to intercept the trigger webhook
	var receivedPayload map[string]interface{}
	var receivedHeaders http.Header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedPayload)
		receivedHeaders = r.Header

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// 2. Point trigger function to our test server
	triggerTarget = ts.URL

	// 3. Execute the command programmatically
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"trigger", "payment.successful"})
	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("trigger execution failed: %v", err)
	}

	// 4. Validate output
	if receivedPayload == nil {
		t.Fatal("Expected test server to receive a payload, got nil")
	}

	if receivedPayload["event"] != "payment.successful" {
		t.Errorf("Expected event payload to be 'payment.successful', got %v", receivedPayload["event"])
	}

	// Check headers
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %v", receivedHeaders.Get("Content-Type"))
	}
}

func TestSamplesCreateCommand(t *testing.T) {
	// Switch to a temporary directory for safe testing
	tempDir, err := os.MkdirTemp("", "alatpay_samples_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // clean up

	// Change working directory to temp dir
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	// Execute samples command
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"samples", "create", "node-checkout"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("samples execution failed: %v", err)
	}

	// Verify the folder and files were generated
	expectedFolder := "alatpay-sample-node-checkout"

	if _, err := os.Stat(expectedFolder); os.IsNotExist(err) {
		t.Fatalf("Expected directory %s to be created, but it was not", expectedFolder)
	}

	expectedPackageJSON := filepath.Join(expectedFolder, "package.json")
	if _, err := os.Stat(expectedPackageJSON); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPackageJSON)
	}

	// Verify content
	content, _ := os.ReadFile(expectedPackageJSON)
	if !strings.Contains(string(content), "alatpay-node-checkout") {
		t.Errorf("package.json did not contain expected content")
	}
}

func TestCustomersCreateCommand(t *testing.T) {
	// Execute customers create command directly
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"customers", "create", "--email", "test@wema.com", "--name", "Wema Test"})

	// Reset the global variables to avoid test pollution
	customerEmail = ""
	customerName = ""
	customerPhone = ""

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("customers create execution failed: %v", err)
	}

	// Check output
	output := b.String()

	if !strings.Contains(output, "test@wema.com") {
		t.Errorf("Expected output to contain email 'test@wema.com', got: %s", output)
	}
	if !strings.Contains(output, "Wema Test") {
		t.Errorf("Expected output to contain name 'Wema Test', got: %s", output)
	}
}
