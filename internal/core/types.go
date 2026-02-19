// Package core implements NAICS code lookup, search, and hierarchy navigation.
// External consumers should use the top-level naics package, which re-exports
// all public types and functions via type aliases.
package core

import "fmt"

// Level represents the hierarchical level of a NAICS code.
type Level int

const (
	// Sector is a 2-digit top-level classification (e.g., "51").
	Sector Level = 2
	// Subsector is a 3-digit classification (e.g., "511").
	Subsector Level = 3
	// IndustryGroup is a 4-digit classification (e.g., "5112").
	IndustryGroup Level = 4
	// NAICSIndustry is a 5-digit classification (e.g., "51121").
	NAICSIndustry Level = 5
	// NationalIndustry is a 6-digit classification (e.g., "511210").
	NationalIndustry Level = 6
)

// String returns the human-readable name of the level.
func (l Level) String() string {
	switch l {
	case Sector:
		return "Sector"
	case Subsector:
		return "Subsector"
	case IndustryGroup:
		return "Industry Group"
	case NAICSIndustry:
		return "NAICS Industry"
	case NationalIndustry:
		return "National Industry"
	default:
		return fmt.Sprintf("Level(%d)", int(l))
	}
}

// Industry represents a NAICS industry classification.
type Industry struct {
	// Code is the 2-6 digit NAICS code (e.g., "511210").
	// Range-based sector codes are stored as-is (e.g., "31-33").
	Code string

	// Title is the official name of the industry (e.g., "Software Publishers").
	Title string

	// Description is the detailed description of the industry.
	// May be empty for codes where only the title is available.
	Description string

	// Level indicates the hierarchical level of this code.
	Level Level
}

// String returns a formatted representation like "511210 - Software Publishers".
func (ind *Industry) String() string {
	if ind == nil {
		return "<nil>"
	}
	return ind.Code + " - " + ind.Title
}

// SearchResult represents a single result from a keyword search,
// including a relevance score for ranking.
type SearchResult struct {
	Industry *Industry
	Score    float64
}

// rangeSectors maps the first two digits of range-based sector codes
// to the canonical sector code. For example, codes starting with "32"
// belong to sector "31-33" (Manufacturing).
var rangeSectors = map[string]string{
	"31": "31-33",
	"32": "31-33",
	"33": "31-33",
	"44": "44-45",
	"45": "44-45",
	"48": "48-49",
	"49": "48-49",
}

// sectorCode returns the canonical sector code for the given NAICS code prefix.
// For range-based sectors (31-33, 44-45, 48-49) it returns the range form.
func sectorCode(codePrefix string) string {
	if len(codePrefix) < 2 {
		return codePrefix
	}
	prefix := codePrefix[:2]
	if ranged, ok := rangeSectors[prefix]; ok {
		return ranged
	}
	return prefix
}

// codeLevel returns the NAICS Level for a given code based on its length.
// Range-based sector codes (e.g., "31-33") are detected and return Sector.
// Returns -1 for invalid code lengths.
func codeLevel(code string) Level {
	if len(code) == 5 && code[2] == '-' {
		return Sector
	}
	n := len(code)
	if n >= 2 && n <= 6 {
		return Level(n)
	}
	return -1
}
