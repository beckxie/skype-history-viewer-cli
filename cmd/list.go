package cmd

import (
	"fmt"

	"github.com/beckxie/skype-history-viewer-cli/pkg/utils"
	"github.com/beckxie/skype-history-viewer-cli/pkg/viewer"
	"github.com/spf13/cobra"
)

var (
	showSystem bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all conversations",
	Long:  `Display a list of all conversations in your Skype history with participant counts and message statistics.`,
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

		// Create viewer with options
		viewerOptions := viewer.ViewerOptions{
			ShowSystemMessages: showSystem,
		}
		messageViewer := viewer.NewMessageViewer(viewerOptions)

		// Display conversation list
		messageViewer.DisplayConversationList(history.Conversations)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Local flags
	listCmd.Flags().BoolVar(&showSystem, "show-system", false, "Include system messages in counts")
}
