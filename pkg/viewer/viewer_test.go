package viewer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
)

func TestNewMessageViewer(t *testing.T) {
	options := ViewerOptions{PageSize: 0}
	v := NewMessageViewer(options)
	if v.options.PageSize != 20 {
		t.Errorf("expected default page size 20, got %d", v.options.PageSize)
	}

	options = ViewerOptions{PageSize: 50}
	v = NewMessageViewer(options)
	if v.options.PageSize != 50 {
		t.Errorf("expected page size 50, got %d", v.options.PageSize)
	}
}

func TestDisplayConversation(t *testing.T) {
	// Mock stdout to avoid cluttering test output
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	os.Stdout = os.NewFile(0, os.DevNull)

	conv := &models.SkypeConversation{
		DisplayName: stringPtr("Test Conv"),
		MessageList: []models.SkypeMessage{
			{OriginalId: "1", Timestamp: "2024-01-01T10:00:00Z", Content: "Msg 1"},
			{OriginalId: "2", Timestamp: "2024-01-01T11:00:00Z", Content: "Msg 2"},
			{OriginalId: "3", Timestamp: "2024-01-01T12:00:00Z", Content: "Msg 3"},
		},
	}

	options := ViewerOptions{PageSize: 2}
	v := NewMessageViewer(options)

	// Test page 1
	v.DisplayConversation(conv, 1)

	// Test page 2
	v.DisplayConversation(conv, 2)

	// Test out of bounds page
	v.DisplayConversation(conv, 10)
}

func TestDisplayConversationFiltering(t *testing.T) {
	// Mock stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	os.Stdout = os.NewFile(0, os.DevNull)

	from, _ := time.Parse(time.RFC3339, "2024-01-01T10:30:00Z")
	to, _ := time.Parse(time.RFC3339, "2024-01-01T11:30:00Z")

	conv := &models.SkypeConversation{
		MessageList: []models.SkypeMessage{
			{OriginalId: "1", Timestamp: "2024-01-01T10:00:00Z", Content: "Before"},
			{OriginalId: "2", Timestamp: "2024-01-01T11:00:00Z", Content: "Inside"},
			{OriginalId: "3", Timestamp: "2024-01-01T12:00:00Z", Content: "After"},
		},
	}

	options := ViewerOptions{
		DateFrom: &from,
		DateTo:   &to,
		PageSize: 10,
	}
	v := NewMessageViewer(options)
	v.DisplayConversation(conv, 1)
}

func TestDisplaySearchResults(t *testing.T) {
	// Mock stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	v := NewMessageViewer(ViewerOptions{})

	// Test empty results
	v.DisplaySearchResults([]SearchResult{})

	// Test results
	results := []SearchResult{
		{
			ConversationName: "Conv 1",
			Message: models.SkypeMessage{
				OriginalId: "m1",
				Timestamp:  "2024-01-01T10:00:00Z",
				Content:    "Hello",
			},
			MatchContext: "Hello world",
		},
	}
	v.DisplaySearchResults(results)

	w.Close()
	out, _ := io.ReadAll(r)
	if len(out) == 0 {
		t.Error("expected output from DisplaySearchResults, got empty")
	}
}

func TestDisplayConversationList(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				Id:          "c1",
				DisplayName: stringPtr("General Chat"),
				MessageList: []models.SkypeMessage{{}, {}},
			},
			{
				Id:          "c2",
				DisplayName: stringPtr("Project X"),
				MessageList: []models.SkypeMessage{{}},
			},
		},
	}

	v := NewMessageViewer(ViewerOptions{})
	v.DisplayConversationList(history.Conversations)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedPhrases := []string{
		"#", "CONVERSATION", "PARTICIPANTS", "MESSAGES", // Table headers (uppercase by default)
		"1", "General Chat", "2",
		"2", "Project X", "1",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("output missing phrase: %s", phrase)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
