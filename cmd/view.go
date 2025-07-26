package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beckxie/SkypeHistoryViewer-go/pkg/models"
	"github.com/beckxie/SkypeHistoryViewer-go/pkg/utils"
	"github.com/beckxie/SkypeHistoryViewer-go/pkg/viewer"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	pageSize        int
	sortNewest      bool
	dateFrom        string
	dateTo          string
	conversationNum int
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view [conversation-number]",
	Short: "View messages from a conversation",
	Long:  `Display messages from a specific conversation with pagination support.`,
	Args:  cobra.MaximumNArgs(1),
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

		// Parse conversation number if provided
		if len(args) > 0 {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid conversation number: %s", args[0])
			}
			conversationNum = num
		}

		// Interactive mode if no conversation specified
		if conversationNum == 0 {
			return interactiveView(history)
		}

		// Validate conversation number
		if conversationNum < 1 || conversationNum > len(history.Conversations) {
			return fmt.Errorf("invalid conversation number: %d (valid range: 1-%d)", 
				conversationNum, len(history.Conversations))
		}

		// Get selected conversation
		conv := &history.Conversations[conversationNum-1]

		// Parse date filters
		var dateFromTime, dateToTime *time.Time
		if dateFrom != "" {
			t, err := utils.ParseDateString(dateFrom)
			if err != nil {
				return fmt.Errorf("invalid date-from: %w", err)
			}
			dateFromTime = t
		}
		if dateTo != "" {
			t, err := utils.ParseDateString(dateTo)
			if err != nil {
				return fmt.Errorf("invalid date-to: %w", err)
			}
			dateToTime = t
		}

		// Create viewer with options
		viewerOptions := viewer.ViewerOptions{
			ShowSystemMessages: showSystem,
			PageSize:          pageSize,
			SortNewestFirst:   sortNewest,
			DateFrom:          dateFromTime,
			DateTo:            dateToTime,
		}
		messageViewer := viewer.NewMessageViewer(viewerOptions)

		// Display conversation with pagination
		return viewConversationWithPagination(conv, messageViewer)
	},
}

// interactiveView provides an interactive conversation selection
func interactiveView(history *models.SkypeHistoryRoot) error {
	// Create viewer for listing
	viewerOptions := viewer.ViewerOptions{
		ShowSystemMessages: showSystem,
	}
	messageViewer := viewer.NewMessageViewer(viewerOptions)

	// Display conversation list
	messageViewer.DisplayConversationList(history.Conversations)

	// Prompt for selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter conversation number to view (or 'q' to quit): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "quit" {
		return nil
	}

	num, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid input: %s", input)
	}

	if num < 1 || num > len(history.Conversations) {
		return fmt.Errorf("invalid conversation number: %d", num)
	}

	// Get selected conversation
	conv := &history.Conversations[num-1]

	// View conversation
	viewerOptions.PageSize = pageSize
	viewerOptions.SortNewestFirst = sortNewest
	messageViewer = viewer.NewMessageViewer(viewerOptions)
	
	return viewConversationWithPagination(conv, messageViewer)
}

// viewConversationWithPagination handles paginated viewing
func viewConversationWithPagination(conv *models.SkypeConversation, v *viewer.MessageViewer) error {
	page := 1
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display current page
		v.DisplayConversation(conv, page)

		// Get navigation input
		fmt.Print("\nNavigation (n=next, p=prev, [number]=page, q=quit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit":
			return nil
		case "n", "next":
			page++
		case "p", "prev":
			if page > 1 {
				page--
			}
		default:
			// Try to parse as page number
			if num, err := strconv.Atoi(input); err == nil && num > 0 {
				page = num
			} else {
				color.New(color.FgRed).Println("Invalid input. Please try again.")
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(viewCmd)

	// Local flags
	viewCmd.Flags().IntVar(&pageSize, "page-size", 20, "Number of messages per page")
	viewCmd.Flags().BoolVar(&sortNewest, "newest-first", false, "Sort messages newest first")
	viewCmd.Flags().BoolVar(&showSystem, "show-system", false, "Show system messages")
	viewCmd.Flags().StringVar(&dateFrom, "date-from", "", "Filter messages from this date (YYYY-MM-DD)")
	viewCmd.Flags().StringVar(&dateTo, "date-to", "", "Filter messages to this date (YYYY-MM-DD)")
}
