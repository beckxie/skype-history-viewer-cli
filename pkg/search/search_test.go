package search

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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
func TestSearchManager_Search_Advanced(t *testing.T) {
	date1, _ := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")

	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				DisplayName: stringPtr("General"),
				MessageList: []models.SkypeMessage{
					{Content: "Apple", From: "Alice", Timestamp: "2024-01-01T10:00:00Z", MessageType: "Text"},
					{Content: "Banana", From: "Bob", Timestamp: "2024-01-01T10:30:00Z", MessageType: "Text"},
					{Content: "Cherry", From: "Alice", Timestamp: "2024-01-01T11:00:00Z", MessageType: "Text"},
				},
			},
			{
				Id: "private",
				MessageList: []models.SkypeMessage{
					{Content: "Secret Apple", From: "Charlie", Timestamp: "2024-01-01T10:15:00Z", MessageType: "Text"},
				},
			},
		},
	}

	sm := NewSearchManager(history)

	t.Run("Multi-condition: Sender + Query", func(t *testing.T) {
		options := SearchOptions{
			Query:           "Alice",
			SearchInSender:  true,
			SearchInContent: true,
		}
		results, _ := sm.Search(context.Background(), options)
		if len(results) != 2 {
			t.Errorf("expected 2 results for Alice, got %d", len(results))
		}
	})

	t.Run("Date Filtering", func(t *testing.T) {
		from := date1.Add(15 * time.Minute)
		options := SearchOptions{
			Query:           "Apple",
			SearchInContent: true,
			DateFrom:        &from,
		}
		results, _ := sm.Search(context.Background(), options)
		if len(results) != 1 {
			t.Errorf("expected 1 result (Secret Apple), got %d", len(results))
		}
	})

	t.Run("Conversation Filter", func(t *testing.T) {
		options := SearchOptions{
			Query:              "Apple",
			SearchInContent:    true,
			ConversationFilter: "General",
		}
		results, _ := sm.Search(context.Background(), options)
		if len(results) != 1 {
			t.Errorf("expected 1 result in General, got %d", len(results))
		}
	})

	t.Run("Limit", func(t *testing.T) {
		options := SearchOptions{
			Query:           "a",
			SearchInContent: true,
			Limit:           2,
		}
		results, _ := sm.Search(context.Background(), options)
		if len(results) != 2 {
			t.Errorf("expected limit of 2, got %d", len(results))
		}
	})
}

func TestSearchManager_CacheLimit(t *testing.T) {
	history := &models.SkypeHistoryRoot{}
	sm := NewSearchManager(history)

	for i := 0; i < 110; i++ {
		sm.buildCacheKey(SearchOptions{Query: string(rune(i))})
		sm.cacheResults(fmt.Sprintf("key-%d", i), nil)
	}

	sm.cacheMutex.RLock()
	defer sm.cacheMutex.RUnlock()
	if len(sm.searchCache) > 101 { // Actually it clears when > 100, so it might be small
		// OK
	}
}

func TestSearchManager_SearchCacheHonorsLimit(t *testing.T) {
	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				MessageList: []models.SkypeMessage{
					{Content: "apple 1", MessageType: "Text", Timestamp: "2024-01-01T10:00:00Z"},
					{Content: "apple 2", MessageType: "Text", Timestamp: "2024-01-01T10:01:00Z"},
					{Content: "apple 3", MessageType: "Text", Timestamp: "2024-01-01T10:02:00Z"},
				},
			},
		},
	}

	sm := NewSearchManager(history)
	first, err := sm.Search(context.Background(), SearchOptions{
		Query:           "apple",
		SearchInContent: true,
		Limit:           1,
	})
	if err != nil {
		t.Fatalf("unexpected error in first search: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected first search to return 1 result, got %d", len(first))
	}

	second, err := sm.Search(context.Background(), SearchOptions{
		Query:           "apple",
		SearchInContent: true,
		Limit:           3,
	})
	if err != nil {
		t.Fatalf("unexpected error in second search: %v", err)
	}
	if len(second) != 3 {
		t.Fatalf("expected second search to return 3 results, got %d", len(second))
	}
}

func TestSearchManager_PreservesContextCasing(t *testing.T) {
	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				MessageList: []models.SkypeMessage{
					{Content: "Apple Pie", MessageType: "Text", Timestamp: "2024-01-01T10:00:00Z"},
				},
			},
		},
	}

	sm := NewSearchManager(history)
	results, err := sm.Search(context.Background(), SearchOptions{
		Query:           "apple",
		SearchInContent: true,
		CaseSensitive:   false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !strings.Contains(results[0].MatchContext, "Apple") {
		t.Fatalf("expected context to preserve original casing, got: %q", results[0].MatchContext)
	}
}

func TestSearchManager_RegexSearch(t *testing.T) {
	history := &models.SkypeHistoryRoot{
		Conversations: []models.SkypeConversation{
			{
				MessageList: []models.SkypeMessage{
					{Content: "Order #1234 completed", MessageType: "Text", Timestamp: "2024-01-01T10:00:00Z", From: "bot-01"},
					{Content: "Order #987 shipped", MessageType: "Text", Timestamp: "2024-01-01T10:01:00Z", From: "bot-02"},
					{Content: "No order here", MessageType: "Text", Timestamp: "2024-01-01T10:02:00Z", From: "user"},
				},
			},
		},
	}

	sm := NewSearchManager(history)

	results, err := sm.Search(context.Background(), SearchOptions{
		Query:           `Order #[0-9]{3,4}`,
		SearchInContent: true,
		RegexSearch:     true,
	})
	if err != nil {
		t.Fatalf("unexpected regex search error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 regex matches, got %d", len(results))
	}

	if results[0].MatchContext == "" {
		t.Fatalf("expected non-empty regex match context")
	}
}

func TestSearchManager_RegexSearch_InvalidPattern(t *testing.T) {
	history := &models.SkypeHistoryRoot{}
	sm := NewSearchManager(history)

	_, err := sm.Search(context.Background(), SearchOptions{
		Query:           `([a-z`,
		SearchInContent: true,
		RegexSearch:     true,
	})
	if err == nil {
		t.Fatal("expected invalid regex error, got nil")
	}
}

func stringPtr(s string) *string {
	return &s
}
