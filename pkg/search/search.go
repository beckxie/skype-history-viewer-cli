package search

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
	"github.com/beckxie/skype-history-viewer-cli/pkg/viewer"
	"github.com/fatih/color"
)

// SearchManager handles searching through Skype history
type SearchManager struct {
	history      *models.SkypeHistoryRoot
	searchCache  map[string][]viewer.SearchResult
	cacheMutex   sync.RWMutex
	lastSearchTime time.Time
}

// NewSearchManager creates a new search manager
func NewSearchManager(history *models.SkypeHistoryRoot) *SearchManager {
	return &SearchManager{
		history:     history,
		searchCache: make(map[string][]viewer.SearchResult),
	}
}

// SearchOptions contains search parameters
type SearchOptions struct {
	Query              string
	SearchInContent    bool
	SearchInSender     bool
	CaseSensitive      bool
	RegexSearch        bool
	ConversationFilter string
	DateFrom           *time.Time
	DateTo             *time.Time
	Limit              int
}

// Search performs a search across all conversations
func (sm *SearchManager) Search(options SearchOptions) []viewer.SearchResult {
	// Check cache
	cacheKey := sm.buildCacheKey(options)
	sm.cacheMutex.RLock()
	if cached, ok := sm.searchCache[cacheKey]; ok {
		sm.cacheMutex.RUnlock()
		return cached
	}
	sm.cacheMutex.RUnlock()

	// Perform search
	results := []viewer.SearchResult{}
	totalMessages := 0
	searchedMessages := 0

	// Count total messages for progress
	for _, conv := range sm.history.Conversations {
		totalMessages += len(conv.MessageList)
	}

	// Progress indicator
	progressChan := make(chan float64)
	go sm.showProgress(progressChan, totalMessages)

	// Search through conversations
	for _, conv := range sm.history.Conversations {
		// Filter by conversation if specified
		if options.ConversationFilter != "" {
			convName := conv.GetConversationDisplayName()
			if !strings.Contains(strings.ToLower(convName), strings.ToLower(options.ConversationFilter)) {
				searchedMessages += len(conv.MessageList)
				progressChan <- float64(searchedMessages)
				continue
			}
		}

		// Search in messages
		for _, msg := range conv.MessageList {
			searchedMessages++
			progressChan <- float64(searchedMessages)

			// Skip system messages
			if msg.IsSystemMessage() {
				continue
			}

			// Apply date filters
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

			// Check for match
			matchResult := sm.checkMatch(&msg, options)
			if matchResult != nil {
				matchResult.ConversationName = conv.GetConversationDisplayName()
				results = append(results, *matchResult)

				// Check limit
				if options.Limit > 0 && len(results) >= options.Limit {
					close(progressChan)
					sm.cacheResults(cacheKey, results)
					return results
				}
			}
		}
	}

	close(progressChan)
	
	// Cache results
	sm.cacheResults(cacheKey, results)
	
	return results
}

// checkMatch checks if a message matches search criteria
func (sm *SearchManager) checkMatch(msg *models.SkypeMessage, options SearchOptions) *viewer.SearchResult {
	query := options.Query
	if !options.CaseSensitive {
		query = strings.ToLower(query)
	}

	contentMatch := false
	senderMatch := false
	matchContext := ""

	// Search in content
	if options.SearchInContent {
		content := msg.GetDisplayText()
		if !options.CaseSensitive {
			content = strings.ToLower(content)
		}

		if strings.Contains(content, query) {
			contentMatch = true
			// Extract context around match
			matchContext = sm.extractContext(content, query, 50)
		}
	}

	// Search in sender
	if options.SearchInSender {
		sender := msg.GetSenderDisplayName()
		if !options.CaseSensitive {
			sender = strings.ToLower(sender)
		}

		if strings.Contains(sender, query) {
			senderMatch = true
		}
	}

	// Return result if match found
	if contentMatch || senderMatch {
		matchType := ""
		if contentMatch && senderMatch {
			matchType = "both"
		} else if contentMatch {
			matchType = "content"
		} else {
			matchType = "sender"
		}

		return &viewer.SearchResult{
			Message:      *msg,
			MatchContext: matchContext,
			MatchType:    matchType,
		}
	}

	return nil
}

// extractContext extracts text around the match
func (sm *SearchManager) extractContext(text, query string, contextSize int) string {
	index := strings.Index(strings.ToLower(text), strings.ToLower(query))
	if index == -1 {
		return ""
	}

	start := index - contextSize
	if start < 0 {
		start = 0
	}

	end := index + len(query) + contextSize
	if end > len(text) {
		end = len(text)
	}

	context := text[start:end]
	
	// Add ellipsis if truncated
	if start > 0 {
		context = "..." + context
	}
	if end < len(text) {
		context = context + "..."
	}

	// Highlight match
	highlighted := strings.ReplaceAll(
		context,
		text[index:index+len(query)],
		color.New(color.FgYellow, color.Bold).Sprint(text[index:index+len(query)]),
	)

	return highlighted
}

// buildCacheKey creates a unique key for caching
func (sm *SearchManager) buildCacheKey(options SearchOptions) string {
	parts := []string{
		options.Query,
		fmt.Sprintf("%v", options.SearchInContent),
		fmt.Sprintf("%v", options.SearchInSender),
		fmt.Sprintf("%v", options.CaseSensitive),
		options.ConversationFilter,
	}

	if options.DateFrom != nil {
		parts = append(parts, options.DateFrom.Format(time.RFC3339))
	}
	if options.DateTo != nil {
		parts = append(parts, options.DateTo.Format(time.RFC3339))
	}

	return strings.Join(parts, "|")
}

// cacheResults stores search results in cache
func (sm *SearchManager) cacheResults(key string, results []viewer.SearchResult) {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	// Limit cache size
	if len(sm.searchCache) > 100 {
		// Remove oldest entries
		for k := range sm.searchCache {
			delete(sm.searchCache, k)
			break
		}
	}

	sm.searchCache[key] = results
}

// showProgress displays search progress
func (sm *SearchManager) showProgress(progressChan <-chan float64, total int) {
	startTime := time.Now()
	lastUpdate := time.Now()

	for progress := range progressChan {
		// Update every 100ms
		if time.Since(lastUpdate) < 100*time.Millisecond {
			continue
		}

		percentage := (progress / float64(total)) * 100
		elapsed := time.Since(startTime)
		
		// Clear line and show progress
		fmt.Printf("\r")
		color.New(color.FgYellow).Printf("Searching... %.1f%% (%d/%d messages) - %.1fs", 
			percentage, int(progress), total, elapsed.Seconds())
		
		lastUpdate = time.Now()
	}

	// Clear progress line
	fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
}

// ClearCache clears the search cache
func (sm *SearchManager) ClearCache() {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()
	
	sm.searchCache = make(map[string][]viewer.SearchResult)
}
