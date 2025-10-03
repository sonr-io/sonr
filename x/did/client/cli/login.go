// Package cli contains the implementation of the CLI commands
package cli

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"cosmossdk.io/log"
	"github.com/sonr-io/sonr/x/did/client/server"
)

// LoginUserWithWebAuthn authenticates a user using WebAuthn through browser interaction
func LoginUserWithWebAuthn(username string) error {
	logger := log.NewLogger(os.Stderr)

	// If no username provided, prompt for it using standard input
	if strings.TrimSpace(username) == "" {
		var err error
		username, err = promptForUsername()
		if err != nil {
			return fmt.Errorf("failed to get username: %w", err)
		}
	}

	// Initialize database and check if username exists
	if err := server.InitDB(); err != nil {
		logger.Warn("Failed to initialize database", "error", err)
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Check if username exists with WebAuthn credentials
	service := server.NewWebAuthnCredentialService()
	existingCredentials, err := service.GetByUsername(username)
	if err != nil || len(existingCredentials) == 0 {
		return fmt.Errorf(
			"username '%s' not found or has no WebAuthn credentials. Please register first.",
			username,
		)
	}

	logger.Info(
		"Found WebAuthn credentials for user",
		"username",
		username,
		"credentialCount",
		len(existingCredentials),
	)

	// Find available port for auth server
	port, err := findAvailablePortForLogin()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	// Create channel to signal completion
	done := make(chan error, 1)

	// Setup server with WebAuthn login context
	err = server.StartAuthServerForLogin(port, username, done)
	if err != nil {
		return fmt.Errorf("failed to start auth server: %w", err)
	}

	defer func() {
		if stopErr := server.StopAuthServer(); stopErr != nil {
			logger.Error("Failed to stop auth server", "error", stopErr)
		}
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// Open browser to WebAuthn login page
	url := fmt.Sprintf("http://localhost:%d/login?username=%s", port, username)
	logger.Info("Opening browser for WebAuthn login", "url", url)

	if err := openBrowserForLogin(url); err != nil {
		logger.Warn("Failed to open browser automatically", "error", err)
		logger.Info("Please navigate manually to the URL", "url", url)
	}

	logger.Info("Waiting for WebAuthn login to complete...")

	// Wait for login to complete or timeout (30 seconds for login vs 10 for registration)
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("WebAuthn login failed: %w", err)
		}
		logger.Info("WebAuthn login completed successfully")
		return nil
	case <-time.After(30 * time.Second):
		logger.Warn("WebAuthn login timed out after 30 seconds")
		return fmt.Errorf("WebAuthn login timed out after 30 seconds - please try again")
	}
}

// findAvailablePortForLogin finds an available port starting from 8090 to avoid conflicts with registration
func findAvailablePortForLogin() (int, error) {
	for port := 8090; port < 8100; port++ {
		conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = conn.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range 8090-8100")
}

// openBrowserForLogin opens the default browser with the given login URL
func openBrowserForLogin(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// #nosec G204 - cmd is hardcoded based on OS, not user input
	return exec.Command(cmd, args...).Start()
}
