package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ##### Constants ###########################################################

const APP_TITLE string = "filesender"
const APP_VERSION string = "v0.0.1"

// ##### Variables ###########################################################

var cmdRoot = &cobra.Command{
	Use:   "filesender",
	Short: "filesender sends files",
	Long:  `filesender sends files using google drive, and uses mnemonic codes for simplisity`,
	Run:   root,
}

// ##### Functions ###########################################################

// Root/base/default command e.g. just display the app info
func Execute() {

	fmt.Printf("\n%s %s\n\n", APP_TITLE, APP_VERSION)

	if err := cmdRoot.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

// Core functionality for the root/base/default command
func root(cmd *cobra.Command, args []string) {

	// If no verb is supplied then just display the "help"
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}
