package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func AddAuthCmds(rootCmd *cobra.Command) {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "User authentication with Passkeys",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// Add auth commands
	authCmd.AddCommand(
		authLoginCmd(),
		authRegisterCmd(),
	)

	// Add to root command
	rootCmd.AddCommand(authCmd)
}

func authLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with WebAuthn authentication using email or phone",
		Long: `Login to your existing identity using WebAuthn/Passkey authentication.
This command will:
1. Start a local auth server
2. Open your browser for WebAuthn credential authentication
3. Verify your existing WebAuthn credential
4. Unlock your DWN vault for data access

You must provide the same email or phone number used during registration.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewLogger(os.Stderr)

			// Get email flag
			email, err := cmd.Flags().GetString("email")
			if err != nil {
				return fmt.Errorf("failed to get email flag: %w", err)
			}

			// Get tel flag
			tel, err := cmd.Flags().GetString("tel")
			if err != nil {
				return fmt.Errorf("failed to get tel flag: %w", err)
			}

			// Validate that exactly one assertion method is provided
			if email == "" && tel == "" {
				return fmt.Errorf("you must provide either --email or --tel")
			}

			if email != "" && tel != "" {
				return fmt.Errorf("please provide only one assertion method (--email or --tel)")
			}

			// Validate email format if provided
			if email != "" && !isValidEmail(email) {
				return fmt.Errorf("invalid email format: %s", email)
			}

			// Validate phone format if provided
			if tel != "" && !isValidPhone(tel) {
				return fmt.Errorf("invalid phone format: %s (must be E.164 format like +1234567890)", tel)
			}

			// Use assertion value as identifier
			identifier := email
			if tel != "" {
				identifier = tel
			}

			logger.Info("Starting WebAuthn login", "identifier", identifier)

			// Execute WebAuthn login
			if err := LoginUserWithWebAuthn(identifier); err != nil {
				logger.Error("WebAuthn login failed", "error", err)
				return fmt.Errorf("WebAuthn login failed: %w", err)
			}

			logger.Info("WebAuthn login completed successfully", "identifier", identifier)
			fmt.Printf("✅ Successfully logged in with: %s\n", identifier)
			return nil
		},
	}

	// Add assertion method flags (one is required)
	cmd.Flags().StringP("email", "e", "", "Email address used during registration")
	cmd.Flags().StringP("tel", "t", "", "Phone number used during registration (E.164 format)")

	return cmd
}

func authRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new identity using WebAuthn with email or phone",
		Long: `Register a new decentralized identity using WebAuthn/Passkey authentication.
This command will:
1. Start a local auth server
2. Open your browser for WebAuthn credential creation  
3. Create a DID document using your email or phone as the assertion method
4. Auto-create a DWN vault for data storage
5. Initialize UCAN delegation chain for authorization

You must provide either an email address or phone number as your primary identifier.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewLogger(os.Stderr)

			// Get client context for transaction broadcasting
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to get client context: %w", err)
			}

			// Get auto-vault flag
			autoCreateVault, err := cmd.Flags().GetBool("auto-vault")
			if err != nil {
				return fmt.Errorf("failed to get auto-vault flag: %w", err)
			}

			// Get email flag for assertion method
			email, err := cmd.Flags().GetString("email")
			if err != nil {
				return fmt.Errorf("failed to get email flag: %w", err)
			}

			// Get tel flag for assertion method
			tel, err := cmd.Flags().GetString("tel")
			if err != nil {
				return fmt.Errorf("failed to get tel flag: %w", err)
			}

			// Validate that exactly one assertion method is provided
			if email == "" && tel == "" {
				return fmt.Errorf("you must provide either --email or --tel")
			}

			if email != "" && tel != "" {
				return fmt.Errorf("please provide only one assertion method (--email or --tel)")
			}

			// Validate email format if provided
			if email != "" && !isValidEmail(email) {
				return fmt.Errorf("invalid email format: %s", email)
			}

			// Validate phone format if provided
			if tel != "" && !isValidPhone(tel) {
				return fmt.Errorf("invalid phone format: %s (must be E.164 format like +1234567890)", tel)
			}

			// Determine assertion type and value
			var assertionType, assertionValue string
			if email != "" {
				assertionType = "email"
				assertionValue = email
				logger.Info("Starting WebAuthn registration with email assertion", "email", email)
			} else {
				assertionType = "tel"
				assertionValue = tel
				logger.Info("Starting WebAuthn registration with phone assertion", "tel", tel)
			}

			// Execute WebAuthn registration and broadcast to blockchain
			if err := RegisterUserWithWebAuthnAndBroadcastWithAssertion(
				clientCtx, "", autoCreateVault, assertionType, assertionValue,
			); err != nil {
				logger.Error("WebAuthn registration failed", "error", err)
				return fmt.Errorf("WebAuthn registration failed: %w", err)
			}

			logger.Info("WebAuthn registration completed successfully",
				"assertionType", assertionType,
				"assertionValue", assertionValue)
			fmt.Printf("✅ Successfully registered identity\n")
			fmt.Printf("   Assertion method: %s (%s)\n", assertionType, assertionValue)
			if autoCreateVault {
				fmt.Printf("   Vault: Auto-created\n")
			}
			return nil
		},
	}

	// Add assertion method flags (one is required)
	cmd.Flags().StringP("email", "e", "", "Email address for identity (e.g., alice@example.com)")
	cmd.Flags().StringP("tel", "t", "", "Phone number for identity (E.164 format, e.g., +1234567890)")

	// Add auto-vault flag
	cmd.Flags().Bool("auto-vault", true, "Automatically create vault for DID (default: true)")

	return cmd
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidPhone validates phone number in E.164 format
func isValidPhone(phone string) bool {
	// E.164 format: + followed by 1-15 digits
	if !strings.HasPrefix(phone, "+") {
		return false
	}

	// Remove the + and check if the rest are digits
	digits := phone[1:]
	if len(digits) < 1 || len(digits) > 15 {
		return false
	}

	for _, ch := range digits {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}
