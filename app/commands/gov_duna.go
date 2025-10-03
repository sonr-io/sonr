package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func GovCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance utilities",
		Long:  `Utilities for governance operations and compliance.`,
	}

	cmd.AddCommand(DunaAmendmentCmd())
	return cmd
}

func DunaAmendmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duna-amendment",
		Short: "Generate Wyoming DAO compliance filing content",
		Long: `Extract governance identifiers and generate Wyoming DAO LC amendment filing content
for compliance with Wyoming Statute 17-31-106(b).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Use the hardcoded ChainID from app.go
			chainID := "sonrtest_1-1"
			// Allow override from client context if set
			if clientCtx.ChainID != "" {
				chainID = clientCtx.ChainID
			}

			saveToFile, _ := cmd.Flags().GetBool("output-files")

			fmt.Printf(
				"🏛️  Wyoming DAO compliance check for Decentralized Identity DAO LC (diDAO)\n\n",
			)

			// Get governance address
			govAddress, err := getGovernanceAddress(clientCtx)
			if err != nil {
				return fmt.Errorf("failed to get governance address: %w", err)
			}
			fmt.Printf("✓ Governance module address: %s\n", govAddress)

			// Get genesis hash
			genesisHash, err := getGenesisHash(clientCtx.HomeDir)
			if err != nil {
				fmt.Printf("⚠️  Warning: Could not calculate genesis hash: %v\n", err)
				genesisHash = fmt.Sprintf(
					"Genesis file location: %s/config/genesis.json",
					clientCtx.HomeDir,
				)
			} else {
				fmt.Printf("✓ Genesis hash: %s\n", genesisHash)
			}

			// Generate and output filing content
			filingContent, amendmentText := generateWyomingContent(govAddress, chainID, genesisHash)

			if saveToFile {
				err = saveWyomingFiles(filingContent, amendmentText)
				if err != nil {
					return fmt.Errorf("failed to save Wyoming filing content: %w", err)
				}
				fmt.Printf("✓ Wyoming filing content generated: wyoming_amendment_content.txt\n")
				fmt.Printf("✓ Article VI amendment text: article_vi_amendment.txt\n")
			} else {
				fmt.Printf("%s", "\n"+filingContent+"\n")
			}

			fmt.Printf("\n✅ Wyoming compliance identifiers ready!\n\n")
			fmt.Printf("Key Information:\n")
			fmt.Printf("- Governance Address: %s\n", govAddress)
			fmt.Printf("- Chain ID: %s\n", chainID)
			fmt.Printf("- Genesis Hash: %s\n", genesisHash)

			if saveToFile {
				fmt.Printf("\nNext Steps:\n")
				fmt.Printf("1. Review: wyoming_amendment_content.txt\n")
				fmt.Printf("2. File online at: https://wyobiz.wy.gov\n")
				fmt.Printf("3. Pay $60 fee and submit\n")
			} else {
				fmt.Printf("\nUse --output-files flag to save content to files.\n")
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().
		Bool("output-files", false, "Save filing content to files instead of printing to stdout")
	return cmd
}

func getGovernanceAddress(clientCtx client.Context) (string, error) {
	queryClient := authtypes.NewQueryClient(clientCtx)

	res, err := queryClient.ModuleAccounts(
		clientCtx.CmdContext,
		&authtypes.QueryModuleAccountsRequest{},
	)
	if err != nil {
		return "", err
	}

	// Find the governance module account
	for _, account := range res.Accounts {
		var acc sdk.AccountI
		if err := clientCtx.InterfaceRegistry.UnpackAny(account, &acc); err != nil {
			continue
		}

		if modAcc, ok := acc.(sdk.ModuleAccountI); ok {
			if modAcc.GetName() == "gov" {
				return acc.GetAddress().String(), nil
			}
		}
	}

	return "", fmt.Errorf("governance module account not found")
}

func getGenesisHash(homeDir string) (string, error) {
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")

	file, err := os.Open(genesisPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func generateWyomingContent(govAddress, chainID, genesisHash string) (string, string) {
	now := time.Now()

	// Generate full filing content
	filingContent := fmt.Sprintf(`WYOMING DAO LC AMENDMENT - ARTICLE VI

Entity Name: Decentralized Identity DAO LC
Amendment Type: Articles of Organization Amendment

ARTICLE VI - DAO PUBLIC IDENTIFIER:

The publicly available identifier for smart contracts used to manage, facilitate, and operate the decentralized autonomous organization is: %s (Cosmos SDK governance module address on chain %s). This identifier provides access to all governance functions including proposal submission, voting, parameter changes, and treasury management as required by Wyoming Statute 17-31-106(b).

VERIFICATION:
- Governance Address: %s
- Chain ID: %s
- Verification Command: snrd query auth module-account gov

Generated on: %s`, govAddress, chainID, govAddress, chainID, now.Format("2006-01-02 15:04:05"))

	// Generate Article VI amendment text only
	amendmentText := fmt.Sprintf(
		`The publicly available identifier for smart contracts used to manage, facilitate, and operate the decentralized autonomous organization is: %s (Cosmos SDK governance module address on chain %s). This identifier provides access to all governance functions including proposal submission, voting, parameter changes, and treasury management as required by Wyoming Statute 17-31-106(b).`,
		govAddress,
		chainID,
	)

	return filingContent, amendmentText
}

func saveWyomingFiles(filingContent, amendmentText string) error {
	// Write full filing content
	err := os.WriteFile("wyoming_amendment_content.txt", []byte(filingContent), 0o644)
	if err != nil {
		return err
	}

	// Write Article VI amendment text
	err = os.WriteFile("article_vi_amendment.txt", []byte(amendmentText), 0o644)
	if err != nil {
		return err
	}

	return nil
}
