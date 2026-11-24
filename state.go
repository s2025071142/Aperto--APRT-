package main

import (
	"fmt"
	"os"

	// Replace with your exact Cosmos SDK / server imports
	"://github.com"
	svccmd "://github.com/cmd"
)

func main() {
	// Construct the root command loop for your daemon interface
	rootCmd, _ := server.NewRootCmd("mycustoml1d", "l1", "/.mycustoml1", "custom-chain-id", ModuleBasics)

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

	// Hook into PreRunE to initialize and validate the chosen storage engine
	originalPreRunE := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if originalPreRunE != nil {
			if err := originalPreRunE(cmd, args); err != nil {
				return err
			}
		}

		modeStr, _ := cmd.Flags().GetString("storage-mode")
		keepRecent, _ := cmd.Flags().GetUint64("prune-keep-recent")

		cfg, err := NewNodeStorageConfig(modeStr, keepRecent)
		if err != nil {
			return fmt.Errorf("invalid storage configuration: %w", err)
		}

		// Store active storage config globally so app.go and abci.go can access it
		ActiveStorageManager = NewStorageManager(cfg)
		fmt.Printf("[Aperto Node] Initialized storage mode: %s\n", cfg.Mode)
		return nil
	}

	// Execute the lifecycle command engine using variables inherited from package scope
	if err := svccmd.Execute(rootCmd, "MYCUSTOML1", DefaultNodeHome); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing blockchain daemon: %v\n", err)
		os.Exit(1)
	}
}
