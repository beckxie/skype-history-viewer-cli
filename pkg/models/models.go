package models

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

// SkypeMessage represents a single message from Skype chat history
type SkypeMessage struct {
	OriginalId     string             `json:"id"`
	DisplayName    *string            `json:"displayName"`
	Content        string             `json:"content"`
	Timestamp      string             `json:"originalarrivaltime"`
	MessageType    string             `json:"messagetype"`
	From           string             `json:"from"`
	ConversationId string             `json:"conversationid"`
	Version        int64              `json:"version"`
	Properties     *MessageProperties `json:"properties"`
	AmsReferences  []string           `json:"amsreferences"`
}

// MessageProperties contains additional message properties
type MessageProperties struct {
	UrlPreviews *string `json:"urlpreviews"`
}

// SkypeConversation represents a conversation in Skype
type SkypeConversation struct {
	Id               string                  `json:"id"`
	DisplayName      *string                 `json:"displayName"`
	Version          int64                   `json:"version"`
	Properties       *ConversationProperties `json:"properties"`
	ThreadProperties *ThreadProperties       `json:"threadProperties"`
	MessageList      []SkypeMessage          `json:"MessageList"`
}

// ConversationProperties contains conversation-specific properties
type ConversationProperties struct {
	ConversationBlocked *bool   `json:"conversationblocked"`
	LastImReceivedTime  *string `json:"lastimreceivedtime"`
	ConsumptionHorizon  *string `json:"consumptionhorizon"`
	ConversationStatus  *string `json:"conversationstatus"`
}

// ThreadProperties contains thread-specific properties
type ThreadProperties struct {
	MemberCount *int    `json:"membercount"`
	Members     *string `json:"members"`
	Topic       *string `json:"topic"`
	Picture     *string `json:"picture"`
	Description *string `json:"description"`
}

// SkypeHistoryRoot represents the root structure of Skype export
type SkypeHistoryRoot struct {
	UserId        string              `json:"userId"`
	ExportDate    string              `json:"exportDate"`
	Conversations []SkypeConversation `json:"conversations"`
}

// GetDisplayText returns clean text without HTML/XML tags
func (m *SkypeMessage) GetDisplayText() string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]+>`)
	cleanContent := re.ReplaceAllString(m.Content, "")

	// Unescape HTML entities
	cleanContent = html.UnescapeString(cleanContent)

	return strings.TrimSpace(cleanContent)
}

// GetSenderDisplayName returns the display name or falls back to 'from' field
func (m *SkypeMessage) GetSenderDisplayName() string {
	if m.DisplayName != nil && *m.DisplayName != "" {
		return *m.DisplayName
	}
	return m.From
}

// IsSystemMessage determines if this is a system message
func (m *SkypeMessage) IsSystemMessage() bool {
	return strings.Contains(m.MessageType, "ThreadActivity") ||
		strings.Contains(m.MessageType, "Control")
}

// GetTimestamp parses and returns the message timestamp
func (m *SkypeMessage) GetTimestamp() (time.Time, error) {
	// Try different time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, m.Timestamp); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", m.Timestamp)
}

// GetConversationDisplayName returns the display name for the conversation
func (c *SkypeConversation) GetConversationDisplayName() string {
	if c.DisplayName != nil && *c.DisplayName != "" {
		return *c.DisplayName
	}

	// Try to get topic from thread properties
	if c.ThreadProperties != nil && c.ThreadProperties.Topic != nil {
		return *c.ThreadProperties.Topic
	}

	// Fallback to ID
	return c.Id
}

// GetParticipantCount returns the number of participants in the conversation
func (c *SkypeConversation) GetParticipantCount() int {
	if c.ThreadProperties != nil && c.ThreadProperties.MemberCount != nil {
		return *c.ThreadProperties.MemberCount
	}

	// Count unique participants from messages
	participants := make(map[string]bool)
	for _, msg := range c.MessageList {
		participants[msg.From] = true
	}

	return len(participants)
}

// FilterSystemMessages returns only non-system messages
func (c *SkypeConversation) FilterSystemMessages() []SkypeMessage {
	var filtered []SkypeMessage
	for _, msg := range c.MessageList {
		if !msg.IsSystemMessage() {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}
