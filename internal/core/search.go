package core

import (
	"sort"
	"strings"
)

// SearchOption configures the behavior of Registry.Search.
type SearchOption func(*searchConfig)

type searchConfig struct {
	maxResults int
}

// MaxResults limits the number of search results returned.
// A value of 0 or less means no limit.
func MaxResults(n int) SearchOption {
	return func(cfg *searchConfig) {
		cfg.maxResults = n
	}
}

const (
	scoreExactTitle    = 100.0
	scoreTitlePrefix   = 60.0
	scoreTitleWord     = 40.0
	scoreTitleContains = 20.0
	scoreDescContains  = 5.0
	scoreLevelBonus    = 0.5 // higher-level codes get a small boost
)

// Search finds industries whose title or description matches the query string.
// Results are ranked by relevance (title matches score higher than description
// matches) and sorted by score descending, then by code ascending.
//
// Search is case-insensitive. An empty query returns nil.
func (reg *Registry) Search(query string, opts ...SearchOption) []SearchResult {
	if query == "" {
		return nil
	}

	cfg := searchConfig{maxResults: 0}
	for _, opt := range opts {
		opt(&cfg)
	}

	q := strings.ToLower(strings.TrimSpace(query))
	qWords := strings.Fields(q)

	var results []SearchResult

	for _, ind := range reg.list {
		score := computeScore(ind, q, qWords)
		if score <= 0 {
			continue
		}
		results = append(results, SearchResult{Industry: ind, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Industry.Code < results[j].Industry.Code
	})

	if cfg.maxResults > 0 && len(results) > cfg.maxResults {
		results = results[:cfg.maxResults]
	}

	return results
}

func computeScore(ind *Industry, query string, queryWords []string) float64 {
	title := strings.ToLower(ind.Title)
	desc := strings.ToLower(ind.Description)

	var score float64

	if title == query {
		score = scoreExactTitle
	} else if strings.HasPrefix(title, query) {
		score = scoreTitlePrefix
	} else if containsAllWords(title, queryWords) {
		score = scoreTitleWord
	} else if strings.Contains(title, query) {
		score = scoreTitleContains
	}

	if score == 0 && desc != "" {
		if strings.Contains(desc, query) {
			score = scoreDescContains
		} else if containsAllWords(desc, queryWords) {
			score = scoreDescContains * 0.5
		}
	}

	if score > 0 {
		levelBoost := float64(7-ind.Level) * scoreLevelBonus
		score += levelBoost
	}

	return score
}

func containsAllWords(text string, words []string) bool {
	for _, w := range words {
		if !strings.Contains(text, w) {
			return false
		}
	}
	return true
}
