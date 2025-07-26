package cmd

import (
	"fmt"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/search"
	"github.com/beckxie/skype-history-viewer-cli/pkg/utils"
	"github.com/beckxie/skype-history-viewer-cli/pkg/viewer"
	"github.com/spf13/cobra"
)

var (
	searchQuery        string
	searchInContent    bool
	searchInSender     bool
	caseSensitive      bool
	conversationFilter string
	searchLimit        int
	searchDateFrom     string
	searchDateTo       string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search through messages",
	Long:  `Search for specific text in your Skype chat history with various filters and options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if JSON path is provided
		if err := checkJSONPath(); err != nil {
			return err
		}

		// Check if search query is provided
		if searchQuery == "" {
			return fmt.Errorf("please provide a search query using -q or --query flag")
		}

		// Load Skype history
		history, err := utils.LoadSkypeHistory(jsonPath)
		if err != nil {
			return fmt.Errorf("failed to load Skype history: %w", err)
		}

		// Parse date filters
		var dateFromTime, dateToTime *time.Time
		if searchDateFrom != "" {
			t, err := utils.ParseDateString(searchDateFrom)
			if err != nil {
				return fmt.Errorf("invalid date-from: %w", err)
			}
			dateFromTime = t
		}
		if searchDateTo != "" {
			t, err := utils.ParseDateString(searchDateTo)
			if err != nil {
				return fmt.Errorf("invalid date-to: %w", err)
			}
			dateToTime = t
		}

		// Create search manager
		searchManager := search.NewSearchManager(history)

		// Prepare search options
		searchOptions := search.SearchOptions{
			Query:              searchQuery,
			SearchInContent:    searchInContent,
			SearchInSender:     searchInSender,
			CaseSensitive:      caseSensitive,
			ConversationFilter: conversationFilter,
			DateFrom:           dateFromTime,
			DateTo:             dateToTime,
			Limit:              searchLimit,
		}

		// Perform search
		results := searchManager.Search(searchOptions)

		// Create viewer and display results
		viewerOptions := viewer.ViewerOptions{
			ShowSystemMessages: false,
		}
		messageViewer := viewer.NewMessageViewer(viewerOptions)
		messageViewer.DisplaySearchResults(results)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Required flags
	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Search query text (required)")
	searchCmd.MarkFlagRequired("query")

	// Optional flags
	searchCmd.Flags().BoolVar(&searchInContent, "content", true, "Search in message content")
	searchCmd.Flags().BoolVar(&searchInSender, "sender", true, "Search in sender names")
	searchCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case-sensitive search")
	searchCmd.Flags().StringVar(&conversationFilter, "conversation", "", "Filter by conversation name")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 50, "Maximum number of results (0 for unlimited)")
	searchCmd.Flags().StringVar(&searchDateFrom, "date-from", "", "Search from this date (YYYY-MM-DD)")
	searchCmd.Flags().StringVar(&searchDateTo, "date-to", "", "Search to this date (YYYY-MM-DD)")
}
