package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/bitwarden/sdk-go/v2"
    "github.com/joho/godotenv"
)

// FingerprintResponse represents the output structure for the fingerprint command
type FingerprintResponse struct {
    Type    string `json:"type"`
    Version string `json:"version"`
}

// FetchResult represents the output structure for the fetch command
type FetchResult struct {
    Result map[string]string `json:"result"`
}

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Error: Not enough arguments")
        fmt.Fprintln(os.Stderr, "Usage: bws-nomad <command> [arg]")
        os.Exit(1)
    }

    command := os.Args[1]

    switch command {
    case "fingerprint":
        handleFingerprint()
    case "fetch":
        handleFetch()
    default:
        fmt.Fprintln(os.Stderr, "Error: Unknown command")
        fmt.Fprintln(os.Stderr, "Usage: bws-nomad fingerprint | fetch <secret_id>")
        os.Exit(1)
    }
}

func handleFingerprint() {
    response := FingerprintResponse{
        Type:    "secrets",
        Version: "0.1.0",
    }

    jsonData, err := json.Marshal(response)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(jsonData))
}

func handleFetch() {
    if len(os.Args) < 3 {
        fmt.Fprintln(os.Stderr, "Error: secret_id required for fetch command")
        fmt.Fprintln(os.Stderr, "Usage: bws-nomad fetch <secret_id>")
        os.Exit(1)
    }

    // Load .env file
    if err := godotenv.Load(".env"); err != nil {
        fmt.Fprintln(os.Stderr, "Warning: No .env file found, using system environment variables")
    }

    // Retrieve environment variables
    accessToken := os.Getenv("BITWARDEN_MACHINE_ACCESS_TOKEN") // Updated variable name
    stateFile := os.Getenv("STATE_FILE")

    if accessToken == "" {
        fmt.Fprintln(os.Stderr, "Error: BITWARDEN_MACHINE_ACCESS_TOKEN is not set") // Updated error message
        os.Exit(1)
    }

    secretID := os.Args[2]

    // Initialize Client
    bitwardenClient, err := sdk.NewBitwardenClient(nil, nil)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
        os.Exit(1)
    }
    defer bitwardenClient.Close()

    // Login using AccessToken
    if err := bitwardenClient.AccessTokenLogin(accessToken, &stateFile); err != nil {
        fmt.Fprintf(os.Stderr, "Error logging in: %v\n", err)
        os.Exit(1)
    }

    // Get Secret
    secret, err := bitwardenClient.Secrets().Get(secretID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error retrieving secret: %v\n", err)
        os.Exit(1)
    }

    // Build result in the specified format using secretID as the key
    result := FetchResult{
        Result: map[string]string{
            secretID: secret.Value,
        },
    }

    // Output the result in the specified format
    outputJSON, err := json.Marshal(result)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(outputJSON))
}
