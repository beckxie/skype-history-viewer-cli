package viewer

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/beckxie/SkypeHistoryViewer-go/pkg/models"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// ViewerOptions contains options for displaying messages
type ViewerOptions struct {
	ShowSystemMessages bool
	PageSize          int
	SortNewestFirst   bool
	DateFrom          *time.Time
	DateTo            *time.Time
}

// MessageViewer handles the display of messages
type MessageViewer struct {
	options ViewerOptions
}

// NewMessageViewer creates a new message viewer
func NewMessageViewer(options ViewerOptions) *MessageViewer {
	if options.PageSize <= 0 {
		options.PageSize = 20
	}
	return &MessageViewer{options: options}
}

// DisplayConversationList shows all conversations in a table
func (v *MessageViewer) DisplayConversationList(conversations []models.SkypeConversation) {
	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Conversation", "Participants", "Messages", "Last Message"})
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	
	// Add data
	for i, conv := range conversations {
		messageCount := len(conv.MessageList)
		if !v.options.ShowSystemMessages {
			messageCount = len(conv.FilterSystemMessages())
		}
		
		lastMessage := ""
		if len(conv.MessageList) > 0 {
			sort.Slice(conv.MessageList, func(i, j int) bool {
				ti, _ := conv.MessageList[i].GetTimestamp()
				tj, _ := conv.MessageList[j].GetTimestamp()
				return ti.Before(tj)
			})
			lastMsg := conv.MessageList[len(conv.MessageList)-1]
			if t, err := lastMsg.GetTimestamp(); err == nil {
				lastMessage = t.Format("2006-01-02 15:04")
			}
		}
		
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			conv.GetConversationDisplayName(),
			fmt.Sprintf("%d", conv.GetParticipantCount()),
			fmt.Sprintf("%d", messageCount),
			lastMessage,
		})
	}
	
	table.Render()
}

// DisplayConversation shows messages from a specific conversation
func (v *MessageViewer) DisplayConversation(conv *models.SkypeConversation, page int) {
	// Filter messages
	messages := conv.MessageList
	if !v.options.ShowSystemMessages {
		messages = conv.FilterSystemMessages()
	}
	
	// Apply date filters
	if v.options.DateFrom != nil || v.options.DateTo != nil {
		filtered := []models.SkypeMessage{}
		for _, msg := range messages {
			t, err := msg.GetTimestamp()
			if err != nil {
				continue
			}
			
			if v.options.DateFrom != nil && t.Before(*v.options.DateFrom) {
				continue
			}
			if v.options.DateTo != nil && t.After(*v.options.DateTo) {
				continue
			}
			
			filtered = append(filtered, msg)
		}
		messages = filtered
	}
	
	// Sort messages
	if v.options.SortNewestFirst {
		sort.Slice(messages, func(i, j int) bool {
			ti, _ := messages[i].GetTimestamp()
			tj, _ := messages[j].GetTimestamp()
			return ti.After(tj)
		})
	}
	
	// Calculate pagination
	totalPages := (len(messages) + v.options.PageSize - 1) / v.options.PageSize
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	
	start := (page - 1) * v.options.PageSize
	end := start + v.options.PageSize
	if end > len(messages) {
		end = len(messages)
	}
	
	// Display header
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("=== %s ===\n", conv.GetConversationDisplayName())
	color.New(color.FgYellow).Printf("Page %d/%d (Messages %d-%d of %d)\n", page, totalPages, start+1, end, len(messages))
	fmt.Println(strings.Repeat("-", 80))
	
	// Display messages
	for _, msg := range messages[start:end] {
		v.DisplayMessage(&msg)
		fmt.Println(strings.Repeat("-", 80))
	}
	
	// Display navigation help
	if totalPages > 1 {
		fmt.Println()
		color.New(color.FgGreen).Println("Navigation: Use 'next'/'prev' commands or specify page number")
	}
}

// DisplayMessage shows a single message
func (v *MessageViewer) DisplayMessage(msg *models.SkypeMessage) {
	// Parse timestamp
	timestamp := "Unknown time"
	if t, err := msg.GetTimestamp(); err == nil {
		timestamp = t.Format("2006-01-02 15:04:05")
	}
	
	// Display sender and timestamp
	color.New(color.FgBlue, color.Bold).Printf("%s", msg.GetSenderDisplayName())
	color.New(color.FgWhite).Printf(" at ")
	color.New(color.FgGreen).Printf("%s", timestamp)
	
	// Display message type if not standard
	if msg.MessageType != "" && msg.MessageType != "Text" && msg.MessageType != "RichText" {
		color.New(color.FgMagenta).Printf(" [%s]", msg.MessageType)
	}
	fmt.Println()
	
	// Display content
	content := msg.GetDisplayText()
	if content != "" {
		fmt.Printf("  %s\n", content)
	}
	
	// Display attachments if any
	if len(msg.AmsReferences) > 0 {
		color.New(color.FgYellow).Printf("  📎 %d attachment(s)\n", len(msg.AmsReferences))
	}
	
	// Display URL previews if any
	if msg.Properties != nil && msg.Properties.UrlPreviews != nil && *msg.Properties.UrlPreviews != "" {
		color.New(color.FgCyan).Printf("  🔗 Contains URL preview\n")
	}
}

// DisplaySearchResults shows search results
func (v *MessageViewer) DisplaySearchResults(results []SearchResult) {
	if len(results) == 0 {
		color.New(color.FgRed).Println("No results found.")
		return
	}
	
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Printf("=== Search Results (%d) ===\n", len(results))
	fmt.Println(strings.Repeat("-", 80))
	
	for i, result := range results {
		// Display result number and conversation
		color.New(color.FgYellow).Printf("[%d] ", i+1)
		color.New(color.FgMagenta).Printf("In: %s\n", result.ConversationName)
		
		// Display message with highlighted match
		v.DisplayMessage(&result.Message)
		
		// Display match context
		if result.MatchContext != "" {
			color.New(color.FgGreen).Printf("  Match: %s\n", result.MatchContext)
		}
		
		fmt.Println(strings.Repeat("-", 80))
	}
}

// SearchResult represents a search match
type SearchResult struct {
	ConversationName string
	Message         models.SkypeMessage
	MatchContext    string
	MatchType       string // "content", "sender", or "both"
}
