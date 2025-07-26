package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beckxie/SkypeHistoryViewer-go/pkg/models"
	"github.com/fatih/color"
)

// LoadSkypeHistory loads Skype history from a JSON file
func LoadSkypeHistory(path string) (*models.SkypeHistoryRoot, error) {
	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access path: %w", err)
	}

	var jsonPath string
	if info.IsDir() {
		// Look for messages.json in the directory
		jsonPath = filepath.Join(path, "messages.json")
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("messages.json not found in directory: %s", path)
		}
	} else {
		jsonPath = path
	}

	// Open file
	file, err := os.Open(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size for progress
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Show loading progress
	fmt.Println()
	color.New(color.FgCyan).Printf("Loading Skype history from: %s\n", jsonPath)
	color.New(color.FgYellow).Printf("File size: %.2f MB\n", float64(fileSize)/(1024*1024))

	// For large files, use streaming decoder
	if fileSize > 100*1024*1024 { // If file is larger than 100MB
		fmt.Println("\nLarge file detected, using streaming decoder...")
		return loadLargeSkypeHistory(file)
	}

	// For smaller files, read normally
	data, err := readFileWithProgress(file, fileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	fmt.Print("\nParsing JSON data...")
	var history models.SkypeHistoryRoot
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Display summary
	fmt.Println(" Done!")
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("✓ Successfully loaded Skype history")
	fmt.Printf("  User ID: %s\n", history.UserId)
	fmt.Printf("  Export Date: %s\n", history.ExportDate)
	fmt.Printf("  Conversations: %d\n", len(history.Conversations))
	
	totalMessages := 0
	for _, conv := range history.Conversations {
		totalMessages += len(conv.MessageList)
	}
	fmt.Printf("  Total Messages: %d\n", totalMessages)
	fmt.Println()

	return &history, nil
}

// loadLargeSkypeHistory loads large Skype history files using streaming
func loadLargeSkypeHistory(file *os.File) (*models.SkypeHistoryRoot, error) {
	fmt.Print("Parsing JSON data (this may take a while)...")
	
	// Use JSON decoder for streaming
	decoder := json.NewDecoder(file)
	
	// Create progress ticker
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-ticker.C:
				dots = (dots + 1) % 4
				fmt.Printf("\rParsing JSON data%s   ", strings.Repeat(".", dots))
			case <-done:
				return
			}
		}
	}()
	
	// Decode JSON
	var history models.SkypeHistoryRoot
	if err := decoder.Decode(&history); err != nil {
		done <- true
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	done <- true
	
	// Display summary
	fmt.Println(" Done!")
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("✓ Successfully loaded Skype history")
	fmt.Printf("  User ID: %s\n", history.UserId)
	fmt.Printf("  Export Date: %s\n", history.ExportDate)
	fmt.Printf("  Conversations: %d\n", len(history.Conversations))
	
	totalMessages := 0
	for _, conv := range history.Conversations {
		totalMessages += len(conv.MessageList)
	}
	fmt.Printf("  Total Messages: %d\n", totalMessages)
	fmt.Println()
	
	return &history, nil
}

// readFileWithProgress reads a file and shows progress
func readFileWithProgress(file *os.File, totalSize int64) ([]byte, error) {
	// Use chunked reading for better memory efficiency
	chunkSize := int64(1024 * 1024) // 1MB chunks
	if chunkSize > totalSize {
		chunkSize = totalSize
	}
	
	var result []byte
	buffer := make([]byte, chunkSize)
	bytesRead := int64(0)
	lastUpdate := time.Now()

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			result = append(result, buffer[:n]...)
			bytesRead += int64(n)
		}

		// Update progress every 100ms
		if time.Since(lastUpdate) > 100*time.Millisecond || err == io.EOF {
			percentage := float64(bytesRead) / float64(totalSize) * 100
			fmt.Printf("\rReading file... %.1f%%", percentage)
			lastUpdate = time.Now()
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	fmt.Print("\r" + strings.Repeat(" ", 30) + "\r")
	return result, nil
}

// ExportConversation exports a conversation to JSON
func ExportConversation(conv *models.SkypeConversation, outputPath string, userId string) error {
	// Create export structure matching the expected format
	export := models.SkypeHistoryRoot{
		UserId:        userId,
		ExportDate:    time.Now().Format(time.RFC3339),
		Conversations: []models.SkypeConversation{*conv},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Display success message
	color.New(color.FgGreen).Printf("✓ Exported conversation to: %s\n", outputPath)
	color.New(color.FgYellow).Printf("  Size: %.2f KB\n", float64(len(data))/1024)
	
	return nil
}

// ParseDateString parses a date string in various formats
func ParseDateString(dateStr string) (*time.Time, error) {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"02/01/2006",
		"02/01/2006 15:04",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date: %s", dateStr)
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1f seconds", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1f minutes", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	} else {
		return fmt.Sprintf("%.1f days", d.Hours()/24)
	}
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	if maxLen <= 3 {
		return s[:maxLen]
	}
	
	return s[:maxLen-3] + "..."
}

// GetStats generates statistics for Skype history
func GetStats(history *models.SkypeHistoryRoot) map[string]interface{} {
	stats := make(map[string]interface{})
	
	// Basic counts
	stats["total_conversations"] = len(history.Conversations)
	
	totalMessages := 0
	totalUsers := make(map[string]bool)
	messageTypes := make(map[string]int)
	
	var firstMessageTime, lastMessageTime *time.Time
	
	for _, conv := range history.Conversations {
		for _, msg := range conv.MessageList {
			totalMessages++
			totalUsers[msg.From] = true
			messageTypes[msg.MessageType]++
			
			// Track first and last messages
			if msgTime, err := msg.GetTimestamp(); err == nil {
				if firstMessageTime == nil || msgTime.Before(*firstMessageTime) {
					t := msgTime
					firstMessageTime = &t
				}
				if lastMessageTime == nil || msgTime.After(*lastMessageTime) {
					t := msgTime
					lastMessageTime = &t
				}
			}
		}
	}
	
	stats["total_messages"] = totalMessages
	stats["total_users"] = len(totalUsers)
	stats["message_types"] = messageTypes
	
	// Date range
	if firstMessageTime != nil {
		stats["first_message_date"] = firstMessageTime.Format("2006-01-02")
	}
	if lastMessageTime != nil {
		stats["last_message_date"] = lastMessageTime.Format("2006-01-02")
	}
	
	return stats
}

// DisplayStats shows statistics in a formatted way
func DisplayStats(stats map[string]interface{}) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("=== Skype History Statistics ===")
	fmt.Println()
	
	if val, ok := stats["total_conversations"]; ok {
		fmt.Printf("Total Conversations: %v\n", val)
	}
	
	if val, ok := stats["total_messages"]; ok {
		fmt.Printf("Total Messages: %v\n", val)
	}
	
	if val, ok := stats["total_users"]; ok {
		fmt.Printf("Total Users: %v\n", val)
	}
	
	if first, ok := stats["first_message_date"]; ok {
		if last, ok2 := stats["last_message_date"]; ok2 {
			fmt.Printf("Date Range: %s to %s\n", first, last)
		}
	}
	
	if types, ok := stats["message_types"].(map[string]int); ok {
		fmt.Println("\nMessage Types:")
		for msgType, count := range types {
			if msgType == "" {
				msgType = "Unknown"
			}
			fmt.Printf("  %s: %d\n", msgType, count)
		}
	}
	
	fmt.Println()
}
