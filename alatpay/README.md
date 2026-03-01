# AlatPay CLI (`alat`)

A powerful Command Line Interface built in Go for testing, managing, and interacting with the AlatPay payment gateway APIs. Designed for developers integrating AlatPay, this CLI allows you to spoof webhooks, spin up local inspection web servers, tail logs, and securely test transactions.

## Installation

Assuming you have [Go](https://go.dev/) installed:

```bash
git clone <this-repo>
cd alatpay
go build -o alatpay .
# Move to your bin path, for example:
# mv alat /usr/local/bin/
```

## Global Configuration & Authentication

To interact securely with the AlatPay Sandbox or Production APIs without hardcoding secrets, use the `auth` command.

### `alatpay auth`
Prompts you for your API Key, Business ID, Vendor ID, and Webhook Secret. This saves securely to `~/.alatpay/config.json`.
The internal API Client automatically handles injecting these values (like `Ocp-Apim-Subscription-Key` and `VendorId`) into all outgoing requests.

```bash
> alatpay auth
Enter AlatPay API Key: 
Enter AlatPay Business ID: 
Enter AlatPay Vendor ID: 
Enter AlatPay Webhook Secret (optional, for signature verification): 
```

## Core Commands

### 1. Webhook Listener & Dashboard (`alat listen`)
Intersects webhooks sent from AlatPay or fired via the CLI trigger command.

```bash
alatpay listen [flags]
```

**Flags:**
- `-p, --port <int>` : The port to listen on for POST webhooks (default 8080).
- `--ui` : Spin up the embedded Web UI Dashboard in the background.
- `--ui-port <int>` : The port to bind the Web UI to (default 8081).
- `--forward-to <url>` : Act as a reverse proxy, forwarding intercepted payloads to your own local development server (e.g. `http://localhost:3000/api/webhook`).

**Example: Local UI Server**
```bash
> alatpay listen --port 8080 --ui --ui-port 8081
```
*Navigating to `http://localhost:8081` will present a rich, real-time dashboard of captured webhooks!*

### 2. Event Simulation (`alatpay trigger`)
Want to simulate an AlatPay transaction succeeding or failing without actually making a test card payment? Use the `trigger` command. It correctly calculates the HMAC SHA512 signature automatically using your `auth` config.

```bash
alatpay trigger [event_type] [flags]
```

**Supported Event Types (Examples):**
- `payment.successful`
- `payment.failed`
- `payment.pending`

**Flags:**
- `-t, --target <url>` : Target URL to send the mock webhook to (default `http://localhost:8080`).

**Example:**
```bash
> alatpay trigger payment.successful
Triggering 'payment.successful' event to http://localhost:8080...
[✓] Event delivered successfully (HTTP 200)
```

### 3. Log Management (`alatpay logs tail`)
Streams a centralized log file (`~/.alatpay/events.log`) to provide real-time HTTP interaction insights across multiple terminal tabs. Note: The current MVP utilizes primarily in-memory dashboards, but this hooks into global file logging setups seamlessly.

```bash
> alatpay logs tail
```

### 4. Cryptography & Payload Verifiers (`alatpay crypto`)
AlatPay mandates that all JSON payloads are encrypted using AES/CBC/PKCS5Padding before transmission, and that responses are decrypted. The CLI handles this transparently for all `transaction` commands using its internal API Client.

However, if you are building your own backend and want to verify your encryption/decryption logic matches AlatPay's expectations, use the `crypto` command space:

#### `alatpay crypto encrypt <json>`
Encrypts raw JSON into AlatPay's expected Base64 padded ciphertext.
```bash
> alatpay crypto encrypt '{"amount": 100, "currency": "NGN"}'
Base64 Encrypted Ciphertext:
[Base64 Output]
```

#### `alatpay crypto decrypt <base64>`
Decrypts AlatPay's Base64 padded ciphertext back into readable JSON.
```bash
> alatpay crypto decrypt "[Base64 Ciphertext]"
Decrypted Plaintext:
{"amount": 100, "currency": "NGN"}
```

#### `alatpay crypto verify-signature <payload> <signature>`
Verifies an AlatPay HMAC SHA512 signature hash against a raw JSON payload using your configured `WebhookSecret`. Perfect for debugging local webhook handlers!
```bash
> alatpay crypto verify-signature '{"event":"payment.successful"}' 'a8c9b...'
[✓] Signature VERIFIED locally!
```

### 5. Transactions (`alatpay transaction`)
Direct API interactors to create or verify physical transactions against AlatPay. Requires `alatpay auth` to be setup.

#### `alatpay transaction create`
Initiates a mock virtual account bank transfer.

**Flags:**
- `-a, --amount <float>` : Amount to charge (default 100.0)
- `-e, --email <string>` : Customer email (default test@example.com)

```bash
> alatpay transaction create --amount 5000 --email dev@example.com
```

#### `alatpay transaction status <id>`
Queries the API to retrieve the real-time status details of a specific transaction ID.

```bash
> alatpay transaction status CLI-ORD-1234567890
```

## Built With
- [Go](https://go.dev)
- [Cobra](https://cobra.dev/) (CLI Framework)
- [Viper](https://github.com/spf13/viper) (Configuration)
- [Fatih/Color](https://github.com/fatih/color) (Terminal Colors)
- Embedded Vue.js + Tailwind CSS (for Web UI)
