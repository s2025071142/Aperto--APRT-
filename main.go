package main

import (
	"fmt"
	"os"

	"://github.com"
	svccmd "://github.com/cmd"
)

func main() {
	// Construct the root command loop for your daemon interface
	rootCmd, _ := server.NewRootCmd("mycustoml1d", "l1", "/.mycustoml1", "custom-chain-id", ModuleBasics)
	
	// Execute the lifecycle command engine using variables inherited from package scope
	if err := svccmd.Execute(rootCmd, "MYCUSTOML1", DefaultNodeHome); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing blockchain daemon: %v\n", err)
		os.Exit(1)
	}
}
