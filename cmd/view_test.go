package cmd

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
	"github.com/beckxie/skype-history-viewer-cli/pkg/viewer"
)

func TestReadNavigationAction(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{name: "next n", input: []byte{'n'}, want: "next"},
		{name: "prev p", input: []byte{'p'}, want: "prev"},
		{name: "first g", input: []byte{'g'}, want: "first"},
		{name: "last G", input: []byte{'G'}, want: "last"},
		{name: "quit q", input: []byte{'q'}, want: "quit"},
		{name: "arrow up", input: []byte{27, '[', 'A'}, want: "prev"},
		{name: "arrow down", input: []byte{27, '[', 'B'}, want: "next"},
		{name: "unknown", input: []byte{'x'}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(tt.input))
			got, err := readNavigationAction(reader)
			if err != nil {
				t.Fatalf("readNavigationAction returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("readNavigationAction() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	from, _ := time.Parse(time.RFC3339, "2024-01-01T10:30:00Z")
	to, _ := time.Parse(time.RFC3339, "2024-01-01T11:30:00Z")

	conv := &models.SkypeConversation{
		MessageList: []models.SkypeMessage{
			{MessageType: "Text", Timestamp: "2024-01-01T10:00:00Z"},
			{MessageType: "Text", Timestamp: "2024-01-01T11:00:00Z"},
			{MessageType: "Control/ThreadActivity", Timestamp: "2024-01-01T11:10:00Z"},
			{MessageType: "Text", Timestamp: "2024-01-01T12:00:00Z"},
		},
	}

	options := viewer.ViewerOptions{
		PageSize:           2,
		ShowSystemMessages: false,
		DateFrom:           &from,
		DateTo:             &to,
	}

	got := calculateTotalPages(conv, options)
	if got != 1 {
		t.Fatalf("calculateTotalPages() = %d, want 1", got)
	}

	options.DateFrom = nil
	options.DateTo = nil
	got = calculateTotalPages(conv, options)
	if got != 2 {
		t.Fatalf("calculateTotalPages() = %d, want 2", got)
	}
}
