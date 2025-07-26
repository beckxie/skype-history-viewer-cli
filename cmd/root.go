package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	jsonPath string
	verbose  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "skype-viewer",
	Short: "A CLI tool to view and search Skype chat history",
	Long: `SkypeHistoryViewer-go is a command-line tool that allows you to:
- View your exported Skype chat history
- Search through messages
- Export conversations
- View statistics

To use this tool, first export your Skype data from:
https://support.microsoft.com/en-us/skype/how-do-i-export-or-delete-my-skype-data-84546e00-2fef-4c45-8ef6-3a27f83242cc`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand is provided
		cmd.Help()
	},
}

// Execute adds all child commands to the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&jsonPath, "file", "f", "", "Path to Skype export JSON file or directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

// Helper function to check if JSON path is provided
func checkJSONPath() error {
	if jsonPath == "" {
		return fmt.Errorf("please provide a JSON file path using -f or --file flag")
	}
	
	// Check if file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return fmt.Errorf("file or directory not found: %s", jsonPath)
	}
	
	return nil
}
