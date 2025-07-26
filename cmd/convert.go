package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beckxie/SkypeHistoryViewer-go/pkg/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert [old-export-file]",
	Short: "Convert old export format to new format",
	Long: `Convert JSON files exported with the old format (with 'conversation' field) 
to the new format (with 'conversations' array) that can be read by all commands.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]

		// Check if file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", inputFile)
		}

		// Read the old format file
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Try to parse as old format
		var oldFormat struct {
			ExportDate   string                     `json:"exportDate"`
			Conversation models.SkypeConversation   `json:"conversation"`
		}

		if err := json.Unmarshal(data, &oldFormat); err != nil {
			// Maybe it's already in new format?
			var newFormat models.SkypeHistoryRoot
			if err2 := json.Unmarshal(data, &newFormat); err2 == nil {
				color.New(color.FgYellow).Println("File is already in the correct format!")
				return nil
			}
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Convert to new format
		newFormat := models.SkypeHistoryRoot{
			UserId:        "converted_user",
			ExportDate:    oldFormat.ExportDate,
			Conversations: []models.SkypeConversation{oldFormat.Conversation},
		}

		// Generate output filename
		outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + "_converted.json"
		
		// Marshal to JSON
		outputData, err := json.MarshalIndent(newFormat, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Write to file
		if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		// Display success
		color.New(color.FgGreen).Printf("✓ Successfully converted to: %s\n", outputFile)
		fmt.Printf("  Original size: %.2f KB\n", float64(len(data))/1024)
		fmt.Printf("  New size: %.2f KB\n", float64(len(outputData))/1024)
		fmt.Println("\nYou can now use this file with all commands:")
		color.New(color.FgCyan).Printf("  ./skype-viewer list -f %s\n", outputFile)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
