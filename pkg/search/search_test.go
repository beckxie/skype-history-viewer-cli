package search

import (
	"context"
	"testing"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
)

func TestSearchManager_Search_Cancellation(t *testing.T) {
	// Create a dummy history with a few messages
	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				MessageList: []models.SkypeMessage{
					{Content: "test message 1", MessageType: "Text", From: "user1", Timestamp: "2024-01-01T10:00:00Z"},
					{Content: "test message 2", MessageType: "Text", From: "user2", Timestamp: "2024-01-01T10:01:00Z"},
					{Content: "another one", MessageType: "Text", From: "user1", Timestamp: "2024-01-01T10:02:00Z"},
				},
			},
		},
	}

	sm := NewSearchManager(history)

	t.Run("Immediate Cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		options := SearchOptions{
			Query:           "test",
			SearchInContent: true,
		}

		results, err := sm.Search(ctx, options)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("Normal Search", func(t *testing.T) {
		ctx := context.Background()
		options := SearchOptions{
			Query:           "test",
			SearchInContent: true,
		}

		results, err := sm.Search(ctx, options)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("Cache Independence from Cancellation", func(t *testing.T) {
		sm.ClearCache()

		// First search: immediate cancel
		ctx1, cancel1 := context.WithCancel(context.Background())
		cancel1()
		sm.Search(ctx1, SearchOptions{Query: "cache-test", SearchInContent: true})

		// Second search: normal
		results, err := sm.Search(context.Background(), SearchOptions{Query: "cache-test", SearchInContent: true})
		if err != nil {
			t.Errorf("unexpected error on second search: %v", err)
		}
		// Since history doesn't have "cache-test", results should be 0, but no error should persist from previous cancel
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}
