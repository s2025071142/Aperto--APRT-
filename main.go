package main

import (
	"fmt"
	"os"

	"://github.com"
	svccmd "://github.com/cmd"
	"://github.com"
)

var (
	// Global merge-mining endpoint configured via CLI flag for pow.go
	MoneroMergeMineRPC string
)

func main() {
	// Construct the root command loop for your daemon interface
	rootCmd, _ := server.NewRootCmd("mycustoml1d", "apto", DefaultNodeHome, "aperto-mainnet-1", ModuleBasics)

	// Automatically detect default storage mode ("utxo" for mobile, "pruned" for PC)
	defaultMode := DetectDefaultStorageMode()

	// Add persistent CLI flags so node operators can override defaults at launch
	rootCmd.PersistentFlags().String(
		"storage-mode",
		string(defaultMode),
		"Node storage mode: 'utxo' (default on mobile), 'pruned' (default on PC), or 'archival'",
	)
	rootCmd.PersistentFlags().Uint64(
		"prune-keep-recent",
		1000,
		"Number of recent blocks to retain when running in 'pruned' mode",
	)
	rootCmd.PersistentFlags().String(
		"xmr-merge-rpc",
		"http://127.0.0",
		"Monero daemon RPC endpoint for RandomX merge-mining (AuxPoW)",
	)

	// Hook into PreRunE to initialize and validate the chosen storage engine & merge-mining config
	originalPreRunE := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if originalPreRunE != nil {
			if err := originalPreRunE(cmd, args); err != nil {
				return err
			}
		}

		modeStr, _ := cmd.Flags().GetString("storage-mode")
		keepRecent, _ := cmd.Flags().GetUint64("prune-keep-recent")
		MoneroMergeMineRPC, _ = cmd.Flags().GetString("xmr-merge-rpc")

		cfg, err := NewNodeStorageConfig(modeStr, keepRecent)
		if err != nil {
			return fmt.Errorf("invalid storage configuration: %w", err)
		}

		// Store active storage config globally so app.go and abci.go can access it
		ActiveStorageManager = NewStorageManager(cfg)
		fmt.Printf("[Aperto Node] Storage Mode: %s | Block Target: %v | Lock-In Limit: <%v\n",
			cfg.Mode,
			TargetBlockTime,
			PreSignLockTimeout,
		)
		return nil
	}

	// Execute the lifecycle command engine using variables inherited from package scope
	if err := svccmd.Execute(rootCmd, "APERTO", DefaultNodeHome); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing blockchain daemon: %v\n", err)
		os.Exit(1)
	}
}
