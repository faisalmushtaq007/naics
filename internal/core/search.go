package core

import (
	"sort"
	"strings"
)

// SearchOption configures the behavior of Registry.Search and Registry.SearchPage.
type SearchOption func(*searchConfig)

type searchConfig struct {
	maxResults int
	offset     int
	minScore   float64
	level      Level
}

// MaxResults limits the number of search results returned.
// A value of 0 or less means no limit.
func MaxResults(n int) SearchOption {
	return func(cfg *searchConfig) {
		cfg.maxResults = n
	}
}

// Offset skips the first n results after sorting, enabling pagination.
// Use together with MaxResults to implement paged search:
//
//	page1 := reg.Search("software", naics.MaxResults(10))
//	page2 := reg.Search("software", naics.MaxResults(10), naics.Offset(10))
func Offset(n int) SearchOption {
	return func(cfg *searchConfig) {
		if n > 0 {
			cfg.offset = n
		}
	}
}

// MinScore filters out results below the given relevance score.
func MinScore(score float64) SearchOption {
	return func(cfg *searchConfig) {
		cfg.minScore = score
	}
}

// AtLevel restricts search results to a specific NAICS hierarchy level.
// For example, AtLevel(naics.NationalIndustry) returns only 6-digit codes.
func AtLevel(l Level) SearchOption {
	return func(cfg *searchConfig) {
		cfg.level = l
	}
}

// SearchResponse contains paginated search results with metadata.
type SearchResponse struct {
	Results []SearchResult
	Total   int // total matches (before offset/limit)
	Offset  int // number of results skipped
	Limit   int // max results requested (0 = unlimited)
}

// HasMore reports whether there are additional results beyond this page.
func (r SearchResponse) HasMore() bool {
	return r.Offset+len(r.Results) < r.Total
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
//
// Supports pagination via Offset and MaxResults options:
//
//	page1 := reg.Search("software", naics.MaxResults(10))
//	page2 := reg.Search("software", naics.MaxResults(10), naics.Offset(10))
func (reg *Registry) Search(query string, opts ...SearchOption) []SearchResult {
	resp := reg.SearchPage(query, opts...)
	return resp.Results
}

// SearchPage works like Search but returns a SearchResponse with pagination
// metadata including the total match count, offset, and limit.
func (reg *Registry) SearchPage(query string, opts ...SearchOption) SearchResponse {
	empty := SearchResponse{}
	if query == "" {
		return empty
	}

	cfg := searchConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	results := reg.collectMatches(query, cfg)

	total := len(results)

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Industry.Code < results[j].Industry.Code
	})

	offset := cfg.offset
	if offset > len(results) {
		offset = len(results)
	}
	results = results[offset:]

	limit := cfg.maxResults
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return SearchResponse{
		Results: results,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	}
}

// Count returns the number of industries matching the query without
// computing full results. More efficient than len(Search(...)) when
// you only need the count.
func (reg *Registry) Count(query string, opts ...SearchOption) int {
	if query == "" {
		return 0
	}

	cfg := searchConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	count := 0
	q := strings.ToLower(strings.TrimSpace(query))
	qWords := strings.Fields(q)

	for _, ind := range reg.list {
		if cfg.level > 0 && ind.Level != cfg.level {
			continue
		}
		score := computeScore(ind, q, qWords)
		if score <= 0 || score < cfg.minScore {
			continue
		}
		count++
	}
	return count
}

func (reg *Registry) collectMatches(query string, cfg searchConfig) []SearchResult {
	q := strings.ToLower(strings.TrimSpace(query))
	qWords := strings.Fields(q)

	var results []SearchResult
	for _, ind := range reg.list {
		if cfg.level > 0 && ind.Level != cfg.level {
			continue
		}
		score := computeScore(ind, q, qWords)
		if score <= 0 || score < cfg.minScore {
			continue
		}
		results = append(results, SearchResult{Industry: ind, Score: score})
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
