package models

import (
	"testing"
	"time"
)

func TestSkypeMessage_GetDisplayText(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "Plain text",
			content: "Hello world",
			want:    "Hello world",
		},
		{
			name:    "Text with HTML tags",
			content: "<b>Bold</b> and <i>Italic</i>",
			want:    "Bold and Italic",
		},
		{
			name:    "Nested HTML tags",
			content: "<div><p>Paragraph with <a href='#'>link</a></p></div>",
			want:    "Paragraph with link",
		},
		{
			name:    "HTML entities",
			content: "Fish &amp; Chips &gt; Burger &lt; Salad",
			want:    "Fish & Chips > Burger < Salad",
		},
		{
			name:    "Tags with attributes",
			content: "<span style=\"color:red\">Red text</span>",
			want:    "Red text",
		},
		{
			name:    "Empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SkypeMessage{Content: tt.content}
			if got := m.GetDisplayText(); got != tt.want {
				t.Errorf("GetDisplayText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkypeMessage_GetTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		want      time.Time
		wantErr   bool
	}{
		{
			name:      "Standard RFC3339",
			timestamp: "2024-01-01T10:00:00Z",
			want:      time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			wantErr:   false,
		},
		{
			name:      "With nanoseconds",
			timestamp: "2024-01-01T10:00:00.123456789Z",
			want:      time.Date(2024, 1, 1, 10, 0, 0, 123456789, time.UTC),
			wantErr:   false,
		},
		{
			name:      "Custom format 1",
			timestamp: "2024-01-01T10:00:00.123Z",
			want:      time.Date(2024, 1, 1, 10, 0, 0, 123000000, time.UTC),
			wantErr:   false,
		},
		{
			name:      "Invalid format",
			timestamp: "2024/01/01 10:00:00",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SkypeMessage{Timestamp: tt.timestamp}
			got, err := m.GetTimestamp()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("GetTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkypeMessage_IsSystemMessage(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		want        bool
	}{
		{
			name:        "RichText is not system",
			messageType: "RichText",
			want:        false,
		},
		{
			name:        "ThreadActivity is system",
			messageType: "Control/ThreadActivity",
			want:        true,
		},
		{
			name:        "ThreadActivity/AddMember is system",
			messageType: "ThreadActivity/AddMember",
			want:        true,
		},
		{
			name:        "Event/Call is NOT system",
			messageType: "Event/Call",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SkypeMessage{MessageType: tt.messageType}
			if got := m.IsSystemMessage(); got != tt.want {
				t.Errorf("IsSystemMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkypeConversation_GetParticipantCount(t *testing.T) {
	count5 := 5
	tests := []struct {
		name             string
		threadProperties *ThreadProperties
		messageList      []SkypeMessage
		want             int
	}{
		{
			name: "From ThreadProperties",
			threadProperties: &ThreadProperties{
				MemberCount: &count5,
			},
			want: 5,
		},
		{
			name: "From MessageList (unique senders)",
			messageList: []SkypeMessage{
				{From: "Alice"},
				{From: "Bob"},
				{From: "Alice"},
				{From: "Charlie"},
			},
			want: 3,
		},
		{
			name: "Empty everything",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SkypeConversation{
				ThreadProperties: tt.threadProperties,
				MessageList:      tt.messageList,
			}
			if got := c.GetParticipantCount(); got != tt.want {
				t.Errorf("GetParticipantCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkypeConversation_GetConversationDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		displayName *string
		topic       *string
		want        string
	}{
		{
			name:        "Use DisplayName",
			id:          "id1",
			displayName: stringPtr("Display"),
			topic:       stringPtr("Topic"),
			want:        "Display",
		},
		{
			name:        "Use Topic if DisplayName empty",
			id:          "id2",
			displayName: stringPtr(""),
			topic:       stringPtr("Topic"),
			want:        "Topic",
		},
		{
			name:        "Use ID as fallback",
			id:          "id3",
			displayName: nil,
			topic:       nil,
			want:        "id3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SkypeConversation{
				Id:          tt.id,
				DisplayName: tt.displayName,
				ThreadProperties: &ThreadProperties{
					Topic: tt.topic,
				},
			}
			if got := c.GetConversationDisplayName(); got != tt.want {
				t.Errorf("GetConversationDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkypeMessage_GetSenderDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		displayName *string
		want        string
	}{
		{
			name:        "Use DisplayName",
			from:        "user1",
			displayName: stringPtr("Alice"),
			want:        "Alice",
		},
		{
			name:        "Use From as fallback",
			from:        "user2",
			displayName: nil,
			want:        "user2",
		},
		{
			name:        "Use From if empty DisplayName",
			from:        "user3",
			displayName: stringPtr(""),
			want:        "user3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SkypeMessage{
				From:        tt.from,
				DisplayName: tt.displayName,
			}
			if got := m.GetSenderDisplayName(); got != tt.want {
				t.Errorf("GetSenderDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
