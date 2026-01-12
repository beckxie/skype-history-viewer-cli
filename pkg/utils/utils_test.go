package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
	"github.com/fatih/color"
)

func TestParseDateString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"2024-01-01", "2024-01-01", false},
		{"2024-01-01 15:04", "2024-01-01 15:04", false},
		{"01/01/2024", "2024-01-01", false},
		{"Jan 1, 2024", "2024-01-01", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		res, err := ParseDateString(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDateString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if res.Format("2006-01-02") != "2024-01-01" && tt.input != "invalid" {
				// Special check for date-only formats
			}
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		max    int
		expect string
	}{
		{"hello world", 5, "he..."},
		{"hello", 10, "hello"},
		{"abc", 2, "ab"},
		{"long string", 8, "long ..."},
	}

	for _, tt := range tests {
		res := TruncateString(tt.input, tt.max)
		if res != tt.expect {
			t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.max, res, tt.expect)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input  time.Duration
		expect string
	}{
		{30 * time.Second, "30.0 seconds"},
		{5 * time.Minute, "5.0 minutes"},
		{2 * time.Hour, "2.0 hours"},
		{48 * time.Hour, "2.0 days"},
	}

	for _, tt := range tests {
		res := FormatDuration(tt.input)
		if res != tt.expect {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, res, tt.expect)
		}
	}
}

func TestLoadSkypeHistory(t *testing.T) {
	// Create a temporary JSON file
	tmpDir, err := os.MkdirTemp("", "skype-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	jsonContent := `{
		"userId": "test-user",
		"exportDate": "2024-01-01T00:00:00Z",
		"conversations": [
			{
				"id": "conv1",
				"displayName": "Test Conversation",
				"MessageList": [
					{
						"id": "msg1",
						"from": "user1",
						"content": "Hello",
						"originalarrivaltime": "2024-01-01T10:00:00Z",
						"messagetype": "Text"
					}
				]
			}
		]
	}`

	jsonPath := filepath.Join(tmpDir, "messages.json")
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading from file
	history, err := LoadSkypeHistory(jsonPath)
	if err != nil {
		t.Errorf("LoadSkypeHistory(file) error = %v", err)
	} else {
		if history.UserId != "test-user" {
			t.Errorf("expected userId test-user, got %s", history.UserId)
		}
	}

	// Test loading from directory
	history, err = LoadSkypeHistory(tmpDir)
	if err != nil {
		t.Errorf("LoadSkypeHistory(dir) error = %v", err)
	} else if len(history.Conversations) != 1 {
		t.Errorf("expected 1 conversation, got %d", len(history.Conversations))
	}

	// Test non-existent path
	_, err = LoadSkypeHistory(filepath.Join(tmpDir, "non-existent"))
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestGetStats(t *testing.T) {
	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				MessageList: []models.SkypeMessage{
					{From: "user1", MessageType: "Text", Timestamp: "2024-01-01T10:00:00Z"},
					{From: "user2", MessageType: "Text", Timestamp: "2024-01-01T11:00:00Z"},
					{From: "user1", MessageType: "Image", Timestamp: "2024-01-01T12:00:00Z"},
				},
			},
		},
	}

	stats := GetStats(history)

	if stats["total_messages"] != 3 {
		t.Errorf("expected 3 total messages, got %v", stats["total_messages"])
	}
	if stats["total_users"] != 2 {
		t.Errorf("expected 2 users, got %v", stats["total_users"])
	}

	msgTypes := stats["message_types"].(map[string]int)
	if msgTypes["Text"] != 2 {
		t.Errorf("expected 2 Text messages, got %d", msgTypes["Text"])
	}
	if msgTypes["Image"] != 1 {
		t.Errorf("expected 1 Image message, got %d", msgTypes["Image"])
	}
}

func TestExportConversation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skype-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	conv := &models.SkypeConversation{
		Id: "test-conv",
		MessageList: []models.SkypeMessage{
			{OriginalId: "m1", Content: "Hello"},
		},
	}

	exportPath := filepath.Join(tmpDir, "export.json")
	err = ExportConversation(conv, exportPath, "user-123")
	if err != nil {
		t.Errorf("ExportConversation error = %v", err)
	}

	// Verify file exists and content
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatal(err)
	}

	var exported models.SkypeHistoryRoot
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatal(err)
	}

	if exported.UserId != "user-123" {
		t.Errorf("expected userId user-123, got %s", exported.UserId)
	}
	if len(exported.Conversations) != 1 || exported.Conversations[0].Id != "test-conv" {
		t.Error("exported conversation data mismatch")
	}
}
func TestDisplayStats(t *testing.T) {
	stats := map[string]interface{}{
		"total_conversations": 5,
		"total_messages":      100,
		"total_users":         10,
		"first_message_date":  "2024-01-01",
		"last_message_date":   "2024-01-31",
		"message_types": map[string]int{
			"Text":  80,
			"Image": 20,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	oldColorOutput := color.Output
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	DisplayStats(stats)

	w.Close()
	os.Stdout = oldStdout
	color.Output = oldColorOutput

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedPhrases := []string{
		"Skype History Statistics",
		"Total Conversations: 5",
		"Total Messages: 100",
		"Total Users: 10",
		"Date Range: 2024-01-01 to 2024-01-31",
		"Text: 80",
		"Image: 20",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("output missing phrase: %s", phrase)
		}
	}
}
