package selectx

import (
	"fmt"
	"strings"
)

// FilterItem represents an item with a score for filtering
type FilterItem struct {
	Index int
	Text  string
	Score int
}

// FilterByQuery filters items by a query string using simple substring matching
func FilterByQuery(items []string, query string) ([]FilterItem, error) {
	if query == "" {
		// Return all items when query is empty
		result := make([]FilterItem, len(items))
		for i, item := range items {
			result[i] = FilterItem{Index: i, Text: item, Score: 0}
		}
		return result, nil
	}

	query = strings.ToLower(query)
	var matches []FilterItem

	for i, item := range items {
		lowerItem := strings.ToLower(item)

		// Exact match
		if lowerItem == query {
			matches = append(matches, FilterItem{Index: i, Text: item, Score: 100})
			continue
		}

		// Prefix match
		if strings.HasPrefix(lowerItem, query) {
			matches = append(matches, FilterItem{Index: i, Text: item, Score: 80})
			continue
		}

		// Substring match
		if strings.Contains(lowerItem, query) {
			matches = append(matches, FilterItem{Index: i, Text: item, Score: 50})
			continue
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for query: %s", query)
	}

	return matches, nil
}
