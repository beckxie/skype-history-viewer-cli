package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
	"github.com/beckxie/skype-history-viewer-cli/pkg/utils"
	"github.com/beckxie/skype-history-viewer-cli/pkg/viewer"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
			PageSize:           pageSize,
			SortNewestFirst:    sortNewest,
			DateFrom:           dateFromTime,
			DateTo:             dateToTime,
		}
		messageViewer := viewer.NewMessageViewer(viewerOptions)

		// Display conversation with pagination
		return viewConversationWithPagination(conv, messageViewer, viewerOptions)
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

	return viewConversationWithPagination(conv, messageViewer, viewerOptions)
}

// viewConversationWithPagination handles paginated viewing
func viewConversationWithPagination(conv *models.SkypeConversation, v *viewer.MessageViewer, options viewer.ViewerOptions) error {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if err := viewConversationWithKeyNavigation(conv, v, options); err == nil {
			return nil
		}
		color.New(color.FgYellow).Println("Key navigation unavailable, fallback to line mode.")
	}
	return viewConversationWithLineNavigation(conv, v)
}

func viewConversationWithLineNavigation(conv *models.SkypeConversation, v *viewer.MessageViewer) error {
	page := 1
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display current page
		v.DisplayConversation(conv, page)

		// Get navigation input
		fmt.Print("\nNavigation (n=next, p=prev, q=quit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit":
			return nil
		case "n":
			page++
		case "p":
			if page > 1 {
				page--
			}
		default:
			color.New(color.FgRed).Println("Invalid input. Please use n, p, or q.")
		}
	}
}

func viewConversationWithKeyNavigation(conv *models.SkypeConversation, v *viewer.MessageViewer, options viewer.ViewerOptions) error {
	fd := int(os.Stdin.Fd())
	reader := bufio.NewReader(os.Stdin)
	page := 1

	for {
		totalPages := calculateTotalPages(conv, options)
		switch {
		case totalPages == 0:
			page = 0
		case page < 1:
			page = 1
		case page > totalPages:
			page = totalPages
		}

		clearTerminalScreen()
		v.DisplayConversation(conv, page)
		fmt.Print(navigationPrompt(totalPages))

		action, err := readNavigationActionRaw(fd, reader)
		fmt.Println()
		if err != nil {
			return err
		}

		switch action {
		case "quit":
			return nil
		case "next":
			if totalPages > 0 && page < totalPages {
				page++
			}
		case "prev":
			if page > 1 {
				page--
			}
		}
	}
}

func readNavigationActionRaw(fd int, reader *bufio.Reader) (string, error) {
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)
	return readNavigationAction(reader)
}

func clearTerminalScreen() {
	fmt.Print("\033[H\033[2J")
}

func navigationPrompt(totalPages int) string {
	if totalPages <= 1 {
		return "\nNavigation (q to quit): "
	}
	return "\nNavigation (n/p, q): "
}

func readNavigationAction(reader *bufio.Reader) (string, error) {
	key, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	switch key {
	case 'q', 'Q', 3:
		return "quit", nil
	case 'n', 'N':
		return "next", nil
	case 'p', 'P':
		return "prev", nil
	case '\r', '\n':
		return "", nil
	}

	return "", nil
}

func calculateTotalPages(conv *models.SkypeConversation, options viewer.ViewerOptions) int {
	messageCount := 0
	for _, msg := range conv.MessageList {
		if !options.ShowSystemMessages && msg.IsSystemMessage() {
			continue
		}

		if options.DateFrom != nil || options.DateTo != nil {
			t, err := msg.GetTimestamp()
			if err != nil {
				continue
			}
			if options.DateFrom != nil && t.Before(*options.DateFrom) {
				continue
			}
			if options.DateTo != nil && t.After(*options.DateTo) {
				continue
			}
		}

		messageCount++
	}

	if messageCount == 0 {
		return 0
	}

	return (messageCount + options.PageSize - 1) / options.PageSize
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
