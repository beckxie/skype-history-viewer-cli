package cmd

import (
	"fmt"

	"github.com/beckxie/skype-history-viewer-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display statistics about your Skype history",
	Long:  `Show detailed statistics about your Skype chat history including message counts, date ranges, and user information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if JSON path is provided
		if err := checkJSONPath(); err != nil {
			return err
		}

		// Load Skype history
		history, err := utils.LoadSkypeHistory(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to load Skype history: %w", err)
		}

		// Generate and display statistics
		stats := utils.GetStats(history)
		utils.DisplayStats(stats)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
