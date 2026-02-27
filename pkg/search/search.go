package search

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beckxie/skype-history-viewer-cli/pkg/models"
	"github.com/beckxie/skype-history-viewer-cli/pkg/viewer"
	"github.com/fatih/color"
)

// SearchManager handles searching through Skype history
type SearchManager struct {
	history     *models.SkypeHistoryRoot
	searchCache map[string][]viewer.SearchResult
	cacheMutex  sync.RWMutex
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

type compiledSearchPattern struct {
	content *regexp.Regexp
	sender  *regexp.Regexp
}

// Search performs a search across all conversations
func (sm *SearchManager) Search(ctx context.Context, options SearchOptions) ([]viewer.SearchResult, error) {
	// Check cache
	cacheKey := sm.buildCacheKey(options)
	sm.cacheMutex.RLock()
	if cached, ok := sm.searchCache[cacheKey]; ok {
		sm.cacheMutex.RUnlock()
		return cached, nil
	}
	sm.cacheMutex.RUnlock()

	// Perform search
	results := []viewer.SearchResult{}
	totalMessages := 0
	searchedMessages := 0

	pattern, err := sm.compileSearchPattern(options)
	if err != nil {
		return nil, err
	}

	// Count total messages for progress
	for _, conv := range sm.history.Conversations {
		totalMessages += len(conv.MessageList)
	}

	// Progress indicator
	progressChan := make(chan float64)
	go sm.showProgress(ctx, progressChan, totalMessages)

	// Search through conversations
	for _, conv := range sm.history.Conversations {
		// Filter by conversation if specified
		if options.ConversationFilter != "" {
			convName := conv.GetConversationDisplayName()
			if !strings.Contains(strings.ToLower(convName), strings.ToLower(options.ConversationFilter)) {
				searchedMessages += len(conv.MessageList)

				select {
				case <-ctx.Done():
					close(progressChan)
					return results, ctx.Err()
				case progressChan <- float64(searchedMessages):
				default:
				}
				continue
			}
		}

		// Search in messages
		for _, msg := range conv.MessageList {
			searchedMessages++

			select {
			case <-ctx.Done():
				close(progressChan)
				return results, ctx.Err()
			case progressChan <- float64(searchedMessages):
			default:
			}

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
			matchResult := sm.checkMatch(&msg, options, pattern)
			if matchResult != nil {
				matchResult.ConversationName = conv.GetConversationDisplayName()
				results = append(results, *matchResult)

				// Check limit
				if options.Limit > 0 && len(results) >= options.Limit {
					close(progressChan)
					sm.cacheResults(cacheKey, results)
					return results, nil
				}
			}
		}
	}

	close(progressChan)

	// Cache results
	sm.cacheResults(cacheKey, results)

	return results, nil
}

// checkMatch checks if a message matches search criteria
func (sm *SearchManager) checkMatch(msg *models.SkypeMessage, options SearchOptions, pattern *compiledSearchPattern) *viewer.SearchResult {
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
		contentToSearch := content
		if !options.CaseSensitive {
			contentToSearch = strings.ToLower(contentToSearch)
		}

		if sm.matchesContent(contentToSearch, query, options, pattern) {
			contentMatch = true
			// Extract context around match
			matchContext = sm.extractContext(content, options, pattern, 50)
		}
	}

	// Search in sender
	if options.SearchInSender {
		sender := msg.GetSenderDisplayName()
		if !options.CaseSensitive {
			sender = strings.ToLower(sender)
		}

		if sm.matchesSender(sender, query, options, pattern) {
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

func (sm *SearchManager) compileSearchPattern(options SearchOptions) (*compiledSearchPattern, error) {
	if !options.RegexSearch {
		return nil, nil
	}

	prefix := ""
	if !options.CaseSensitive {
		prefix = "(?i)"
	}

	compiled, err := regexp.Compile(prefix + options.Query)
	if err != nil {
		return nil, fmt.Errorf("invalid regex query: %w", err)
	}

	return &compiledSearchPattern{content: compiled, sender: compiled}, nil
}

func (sm *SearchManager) matchesContent(contentToSearch, query string, options SearchOptions, pattern *compiledSearchPattern) bool {
	if options.RegexSearch {
		return pattern != nil && pattern.content != nil && pattern.content.MatchString(contentToSearch)
	}

	return strings.Contains(contentToSearch, query)
}

func (sm *SearchManager) matchesSender(sender, query string, options SearchOptions, pattern *compiledSearchPattern) bool {
	if options.RegexSearch {
		return pattern != nil && pattern.sender != nil && pattern.sender.MatchString(sender)
	}

	return strings.Contains(sender, query)
}

// extractContext extracts text around the match
func (sm *SearchManager) extractContext(text string, options SearchOptions, pattern *compiledSearchPattern, contextSize int) string {
	searchText := text
	searchQuery := options.Query
	if !options.CaseSensitive {
		searchText = strings.ToLower(searchText)
		searchQuery = strings.ToLower(searchQuery)
	}

	index := strings.Index(searchText, searchQuery)
	matchLength := len(options.Query)
	if options.RegexSearch && pattern != nil && pattern.content != nil {
		loc := pattern.content.FindStringIndex(text)
		if loc == nil {
			return ""
		}
		index = loc[0]
		matchLength = loc[1] - loc[0]
	}

	if index == -1 {
		return ""
	}

	start := index - contextSize
	if start < 0 {
		start = 0
	}

	end := index + matchLength + contextSize
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

	matchStart := index - start
	matchEnd := matchStart + matchLength
	if matchStart < 0 || matchEnd > len(context) {
		return context
	}
	highlighted := context[:matchStart] +
		color.New(color.FgYellow, color.Bold).Sprint(context[matchStart:matchEnd]) +
		context[matchEnd:]

	return highlighted
}

// buildCacheKey creates a unique key for caching
func (sm *SearchManager) buildCacheKey(options SearchOptions) string {
	parts := []string{
		options.Query,
		fmt.Sprintf("%v", options.SearchInContent),
		fmt.Sprintf("%v", options.SearchInSender),
		fmt.Sprintf("%v", options.CaseSensitive),
		fmt.Sprintf("%v", options.RegexSearch),
		options.ConversationFilter,
		fmt.Sprintf("%d", options.Limit),
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
func (sm *SearchManager) showProgress(ctx context.Context, progressChan <-chan float64, total int) {
	startTime := time.Now()
	lastUpdate := time.Now()

	for {
		select {
		case <-ctx.Done():
			// Clear progress line on cancellation
			fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
			return
		case progress, ok := <-progressChan:
			if !ok {
				// Channel closed, clear line and return
				fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
				return
			}

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
	}
}

// ClearCache clears the search cache
func (sm *SearchManager) ClearCache() {
	sm.cacheMutex.Lock()
	defer sm.cacheMutex.Unlock()

	sm.searchCache = make(map[string][]viewer.SearchResult)
}
