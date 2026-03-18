package main

import (
    "encoding/json"
    "fmt"
    "log/syslog"
    "os"

    "github.com/bitwarden/sdk-go/v2"
    "github.com/joho/godotenv"
)

// Exit codes
const (
    ExitSuccess            = 0
    ExitInvalidArgs        = 1
    ExitEnvVarNotSet       = 2
    ExitClientInitError    = 3
    ExitLoginError         = 4
    ExitSecretRetrievalErr = 5
    ExitJSONMarshalError   = 6
    ExitUnknownCommand     = 7
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

var syslogInfo *syslog.Writer
var syslogErr *syslog.Writer

func initSyslog() error {
    var err error
    syslogInfo, err = syslog.New(syslog.LOG_INFO, "bws-nomad")
    if err != nil {
        return fmt.Errorf("failed to initialize syslog info: %w", err)
    }
    
    syslogErr, err = syslog.New(syslog.LOG_ERR, "bws-nomad")
    if err != nil {
        return fmt.Errorf("failed to initialize syslog err: %w", err)
    }
    return nil
}

func closeSyslog() {
    if syslogInfo != nil {
        syslogInfo.Close()
    }
    if syslogErr != nil {
        syslogErr.Close()
    }
}

func logInfo(message string) {
    if syslogInfo != nil {
        _, _ = syslogInfo.Write([]byte(message))
    }
}

func logError(message string) {
    if syslogErr != nil {
        _, _ = syslogErr.Write([]byte(message))
    }
}

func logInput(command string, args ...string) {
    logInfo(fmt.Sprintf("INPUT: command=%s, args=%v", command, args))
}

func logSuccess(command string, secretID string) {
    logInfo(fmt.Sprintf("SUCCESS: command=%s, secret_id=%s", command, secretID))
}

func logErrorWithDetails(command string, secretID string, err error) {
    logError(fmt.Sprintf("ERROR: command=%s, secret_id=%s, error=%v", command, secretID, err))
}

func main() {
    // Initialize syslog
    if err := initSyslog(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
        os.Exit(ExitUnknownCommand)
    }
    defer closeSyslog()

    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Error: Not enough arguments")
        fmt.Fprintln(os.Stderr, "Usage: bws-nomad <command> [arg]")
        os.Exit(ExitInvalidArgs)
    }

    command := os.Args[1]
    args := os.Args[2:]

    logInput(command, args...)

    switch command {
    case "fingerprint":
        handleFingerprint()
    case "fetch":
        handleFetch(args)
    default:
        logErrorWithDetails(command, "", fmt.Errorf("unknown command: %s", command))
        fmt.Fprintln(os.Stderr, "Error: Unknown command")
        fmt.Fprintln(os.Stderr, "Usage: bws-nomad fingerprint | fetch <secret_id>")
        os.Exit(ExitUnknownCommand)
    }
}

func handleFingerprint() {
    response := FingerprintResponse{
        Type:    "secrets",
        Version: "0.1.0",
    }

    jsonData, err := json.Marshal(response)
    if err != nil {
        logErrorWithDetails("fingerprint", "", err)
        fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
        os.Exit(ExitJSONMarshalError)
    }

    logSuccess("fingerprint", "")
    fmt.Println(string(jsonData))
    os.Exit(ExitSuccess)
}

func handleFetch(args []string) {
    if len(args) < 1 {
        logErrorWithDetails("fetch", "", fmt.Errorf("secret_id required"))
        fmt.Fprintln(os.Stderr, "Error: secret_id required for fetch command")
        fmt.Fprintln(os.Stderr, "Usage: bws-nomad fetch <secret_id>")
        os.Exit(ExitInvalidArgs)
    }
    
    // Load .env file
    if err := godotenv.Load(".env"); err != nil {
        logInfo("Warning: No .env file found, using system environment variables")
    }
    
    // Retrieve environment variables
    accessToken := os.Getenv("BITWARDEN_MACHINE_ACCESS_TOKEN")
    // stateFile := os.Getenv("STATE_FILE")
    if accessToken == "" {
        logErrorWithDetails("fetch", "", fmt.Errorf("BITWARDEN_MACHINE_ACCESS_TOKEN is not set"))
        fmt.Fprintln(os.Stderr, "Error: BITWARDEN_MACHINE_ACCESS_TOKEN is not set")
        os.Exit(ExitEnvVarNotSet)
    }
    
    // Add partial token logging for debugging (show first 8 chars only)
    logInfo(fmt.Sprintf("DEBUG: Machine Access Token starts with: %s...", accessToken[:min(len(accessToken), 5)]))
    
    secretID := args[0]
    
    // Initialize Client
    bitwardenClient, err := sdk.NewBitwardenClient(nil, nil)
    if err != nil {
        logErrorWithDetails("fetch", secretID, err)
        fmt.Fprintf(os.Stderr, "Error initializing client: %v\n", err)
        os.Exit(ExitClientInitError)
    }
    defer bitwardenClient.Close()
    
    // Login using AccessToken
    if err := bitwardenClient.AccessTokenLogin(accessToken, nil); err != nil {
        logErrorWithDetails("fetch", secretID, err)
        fmt.Fprintf(os.Stderr, "Error logging in: %v\n", err)
        os.Exit(ExitLoginError)
    }
    
    // Get Secret
    secret, err := bitwardenClient.Secrets().Get(secretID)
    if err != nil {
        logErrorWithDetails("fetch", secretID, err)
        fmt.Fprintf(os.Stderr, "Error retrieving secret: %v\n", err)
        os.Exit(ExitSecretRetrievalErr)
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
        logErrorWithDetails("fetch", secretID, err)
        fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
        os.Exit(ExitJSONMarshalError)
    }
    logSuccess("fetch", secretID)
    fmt.Println(string(outputJSON))
    os.Exit(ExitSuccess)
}

// Helper function for safe string slicing
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
