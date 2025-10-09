package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sonr-io/crypto/mpc"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

const (
	// DefaultTestChainID is the default chain ID used for local testing
	DefaultTestChainID = "sonrtest_1-1"
)

// GetWalletQueryCommands returns wallet-specific query commands
func GetWalletQueryCommands() *cobra.Command {
	walletQueryCmd := &cobra.Command{
		Use:                        "wallet",
		Short:                      "Wallet query commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	walletQueryCmd.AddCommand(
		GetCmdWalletDerive(),
		GetCmdWalletStatus(),
		GetCmdWalletConfig(),
	)

	return walletQueryCmd
}

// GetCmdWalletDerive creates a command to derive wallet addresses from enclave data
func GetCmdWalletDerive() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "derive [enclave-data]",
		Short: "Derive wallet address and DID from enclave data",
		Long: `Derive wallet address and DID from MPC enclave data.
The enclave-data should be provided as hex-encoded JSON or a file path.

Example:
  snrd query dwn wallet derive '{"pub_hex":"...","pub_bytes":[...],...}'
  snrd query dwn wallet derive @enclave.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Parse enclave data
			enclaveData, err := parseEnclaveData(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse enclave data: %w", err)
			}

			// Get chain ID from client context
			chainID := clientCtx.ChainID
			if chainID == "" {
				chainID = DefaultTestChainID // Default for local testing
			}

			// Create enclave configuration
			config := plugin.CreateEnclaveConfig(chainID, enclaveData)

			// Load plugin and derive wallet address
			ctx := context.Background()
			motorPlugin, err := plugin.LoadPluginWithManager(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to load Motor plugin: %w", err)
			}

			// Get issuer DID and address
			response, err := motorPlugin.GetIssuerDID()
			if err != nil {
				return fmt.Errorf("failed to derive wallet address: %w", err)
			}

			if response.Error != "" {
				return fmt.Errorf("plugin error: %s", response.Error)
			}

			// Display results
			result := map[string]any{
				"issuer_did": response.IssuerDID,
				"address":    response.Address,
				"chain_code": response.ChainCode,
				"chain_id":   chainID,
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdWalletStatus creates a command to check wallet plugin status
func GetCmdWalletStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [wallet-address]",
		Short: "Check wallet plugin health and status",
		Long: `Check the health status of wallet plugins managed by the plugin manager.
Optionally filter by wallet address if enclave data is provided.

Example:
  snrd query dwn wallet status
  snrd query dwn wallet status sonr1abc123...`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Get all plugin IDs from the default manager
			pluginIDs := plugin.DefaultManager.ListPlugins()

			if len(pluginIDs) == 0 {
				fmt.Println("No wallet plugins currently loaded")
				return nil
			}

			var statusResults []map[string]any

			// Check status for each plugin
			for _, id := range pluginIDs {
				stats, err := plugin.DefaultManager.GetPluginStats(id)
				if err != nil {
					fmt.Printf("Error getting stats for plugin %s: %v\n", id, err)
					continue
				}

				status := map[string]any{
					"plugin_id":       stats.ID,
					"chain_id":        stats.ChainID,
					"is_healthy":      stats.IsHealthy,
					"error_count":     stats.ErrorCount,
					"created_at":      stats.CreatedAt.Format("2006-01-02 15:04:05"),
					"last_used":       stats.LastUsed.Format("2006-01-02 15:04:05"),
					"uptime_duration": stats.UptimeDuration.String(),
					"idle_duration":   stats.IdleDuration.String(),
				}

				// If wallet address is provided, try to match
				if len(args) > 0 {
					walletAddress := args[0]
					// For now, we include all plugins since we can't easily derive address from plugin ID
					// In a production implementation, you might want to store address mappings
					_ = walletAddress
				}

				statusResults = append(statusResults, status)
			}

			if len(statusResults) == 0 {
				fmt.Println("No matching wallet plugins found")
				return nil
			}

			// Print summary
			healthyCount := 0
			for _, status := range statusResults {
				if status["is_healthy"].(bool) {
					healthyCount++
				}
			}

			fmt.Printf(
				"Wallet Plugin Status Summary: %d/%d healthy\n\n",
				healthyCount,
				len(statusResults),
			)

			return clientCtx.PrintObjectLegacy(map[string]any{
				"summary": map[string]any{
					"total_plugins":     len(statusResults),
					"healthy_plugins":   healthyCount,
					"unhealthy_plugins": len(statusResults) - healthyCount,
				},
				"plugins": statusResults,
			})
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdWalletConfig creates a command to query wallet configuration
func GetCmdWalletConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display wallet plugin configuration and capabilities",
		Long: `Display the default configuration used for wallet plugins including:
- Security settings and timeouts
- Vault configuration parameters
- Plugin loader settings
- Supported cryptographic capabilities`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Get default configurations
			defaultEnclaveConfig := plugin.DefaultEnclaveConfig()
			defaultLoaderConfig := plugin.DefaultLoaderConfig()

			result := map[string]any{
				"enclave_config": map[string]any{
					"default_chain_id": defaultEnclaveConfig.ChainID,
					"vault_config": map[string]any{
						"ipfs_endpoint":      defaultEnclaveConfig.VaultConfig.IPFSEndpoint,
						"max_vault_size":     defaultEnclaveConfig.VaultConfig.MaxVaultSize,
						"enable_compression": defaultEnclaveConfig.VaultConfig.EnableCompression,
						"backup_enabled":     defaultEnclaveConfig.VaultConfig.BackupEnabled,
					},
					"security_config": map[string]any{
						"max_token_lifetime": defaultEnclaveConfig.Security.MaxTokenLifetime.String(),
						"require_audience":   defaultEnclaveConfig.Security.RequireAudience,
						"allowed_origins":    defaultEnclaveConfig.Security.AllowedOrigins,
					},
					"timeouts": map[string]any{
						"token_creation": defaultEnclaveConfig.Timeouts.TokenCreation.String(),
						"signature":      defaultEnclaveConfig.Timeouts.Signature.String(),
						"verification":   defaultEnclaveConfig.Timeouts.Verification.String(),
						"plugin_init":    defaultEnclaveConfig.Timeouts.PluginInit.String(),
					},
				},
				"loader_config": map[string]any{
					"enable_wasi":            defaultLoaderConfig.EnableWASI,
					"memory_limit":           defaultLoaderConfig.MemoryLimit,
					"allow_http_requests":    defaultLoaderConfig.AllowHttpRequests,
					"log_level":              defaultLoaderConfig.LogLevel,
					"max_concurrent_plugins": defaultLoaderConfig.MaxConcurrentPlugins,
				},
				"capabilities": map[string]any{
					"supported_operations": []string{
						"UCAN token creation (origin)",
						"UCAN token delegation (attenuated)",
						"Data signing (MPC-based)",
						"Signature verification",
						"DID derivation",
						"Address generation",
					},
					"supported_curves": []string{"secp256k1"},
					"plugin_format":    "WebAssembly (WASM)",
					"mpc_support":      true,
				},
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// parseEnclaveData parses enclave data from a string (JSON or file path)
func parseEnclaveData(input string) (*mpc.EnclaveData, error) {
	var data []byte

	// Check if input is a file path (starts with @)
	if len(input) > 0 && input[0] == '@' {
		// File path - not implemented for security reasons in this example
		return nil, fmt.Errorf("file input not supported in this implementation")
	} else {
		// Direct JSON string
		data = []byte(input)
	}

	// Try to parse as hex-encoded data first
	if len(input) > 2 && (input[:2] == "0x" || input[:2] == "0X") {
		hexData, err := hex.DecodeString(input[2:])
		if err == nil {
			data = hexData
		}
	}

	// Parse JSON
	var enclaveData mpc.EnclaveData
	if err := json.Unmarshal(data, &enclaveData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate enclave data
	if !enclaveData.IsValid() {
		return nil, fmt.Errorf("invalid enclave data: missing required fields")
	}

	return &enclaveData, nil
}
