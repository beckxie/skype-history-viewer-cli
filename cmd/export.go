package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/beckxie/SkypeHistoryViewer-go/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	outputPath    string
	exportConvNum int
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [conversation-number]",
	Short: "Export a conversation to JSON",
	Long:  `Export a specific conversation from your Skype history to a JSON file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if JSON path is provided
		if err := checkJSONPath(); err != nil {
			return err
		}

		// Parse conversation number
		num := 0
		fmt.Sscanf(args[0], "%d", &num)
		if num < 1 {
			return fmt.Errorf("invalid conversation number: %s", args[0])
		}

		// Load Skype history
		history, err := utils.LoadSkypeHistory(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to load Skype history: %w", err)
		}

		// Validate conversation number
		if num > len(history.Conversations) {
			return fmt.Errorf("conversation number %d not found (valid range: 1-%d)", 
				num, len(history.Conversations))
		}

		// Get conversation
		conv := &history.Conversations[num-1]

		// Generate output filename if not specified
		if outputPath == "" {
			// Clean conversation name for filename
			convName := conv.GetConversationDisplayName()
			convName = strings.ReplaceAll(convName, "/", "_")
			convName = strings.ReplaceAll(convName, ":", "_")
			convName = strings.ReplaceAll(convName, " ", "_")
			
			outputPath = fmt.Sprintf("conversation_%s.json", convName)
		}

		// Ensure .json extension
		if !strings.HasSuffix(outputPath, ".json") {
			outputPath += ".json"
		}

		// Make path absolute
		absPath, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}

		// Export conversation with original userId
		if err := utils.ExportConversation(conv, absPath, history.UserId); err != nil {
			return fmt.Errorf("failed to export conversation: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Local flags
	exportCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: auto-generated)")
}
