package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// samplesCmd represents the samples command
var samplesCmd = &cobra.Command{
	Use:   "samples",
	Short: "Create sample integrations for AlatPay",
	Long: `Quickly bootstrap a new integration by generating sample AlatPay projects. 
This lets you see a functional end-to-end integration in minutes.`,
}

// samplesListCmd represents listing available samples
var samplesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available sample integrations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available AlatPay Sample Integrations:")
		fmt.Println(color.CyanString("  node-checkout") + "\t- A complete check out flow in Node.js/Express")
		fmt.Println(color.CyanString("  python-checkout") + "\t- A complete check out flow in Python/Flask")
		fmt.Println(color.CyanString("  go-webhook") + "\t- A webhook processing application in Go")
		fmt.Println("\nRun " + color.YellowString("alatpay samples create <sample_name>") + " to scaffold a project.")
	},
}

// samplesCreateCmd represents generating a sample
var samplesCreateCmd = &cobra.Command{
	Use:   "create [sample_name]",
	Short: "Create a sample integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sampleName := args[0]
		dirName := fmt.Sprintf("alatpay-sample-%s", sampleName)

		fmt.Printf("Scaffolding %s sample into ./%s...\n", color.CyanString(sampleName), dirName)

		if err := os.Mkdir(dirName, 0755); err != nil {
			fmt.Printf(color.RedString("[!] Failed to create directory: %v\n"), err)
			return
		}

		// Very basic dynamic scaffolding based on name
		switch sampleName {
		case "node-checkout":
			createFile(dirName, "package.json", `{
  "name": "alatpay-node-checkout",
  "version": "1.0.0",
  "main": "server.js",
  "dependencies": {
    "express": "^4.18.2"
  }
}`)
			createFile(dirName, "server.js", `const express = require('express');
const app = express();
app.use(express.json());

app.post('/webhook', (req, res) => {
  console.log('Webhook received!', req.body);
  res.send({status: 'ok'});
});

app.listen(3000, () => console.log('Node sample running on port 3000'));
`)

		case "python-checkout":
			createFile(dirName, "requirements.txt", `flask==3.0.0
requests==2.31.0
`)
			createFile(dirName, "app.py", `from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/webhook', methods=['POST'])
def webhook():
    print(f"Webhook received! {request.json}")
    return jsonify(status='ok')

if __name__ == '__main__':
    app.run(port=3000)
`)

		default:
			createFile(dirName, "README.md", fmt.Sprintf("# %s Sample\n\nThis is an auto-generated sample project.", sampleName))
		}

		fmt.Println(color.GreenString("[✓] Sample created successfully!"))
		fmt.Printf("\nNext steps:\n  cd %s\n", dirName)
	},
}

func createFile(dir, filename, content string) {
	path := filepath.Join(dir, filename)
	os.WriteFile(path, []byte(content), 0644)
}

func init() {
	rootCmd.AddCommand(samplesCmd)
	samplesCmd.AddCommand(samplesListCmd)
	samplesCmd.AddCommand(samplesCreateCmd)
}
