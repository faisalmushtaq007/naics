// Package naics provides lookup, validation, search, and hierarchy navigation
// for North American Industry Classification System (NAICS) codes.
//
// # NAICS Hierarchy
//
// The NAICS hierarchy has five levels:
//
//   - 2-digit: Sector (e.g., 51 "Information")
//   - 3-digit: Subsector (e.g., 513 "Publishing Industries")
//   - 4-digit: Industry Group (e.g., 5132 "Software Publishers")
//   - 5-digit: NAICS Industry (e.g., 51321 "Software Publishers")
//   - 6-digit: National Industry (e.g., 513210 "Software Publishers")
//
// Three sectors use range-based codes: Manufacturing (31-33), Retail Trade (44-45),
// and Transportation and Warehousing (48-49).
//
// # Getting Started
//
// Create a Registry to look up, search, and navigate NAICS codes:
//
//	reg, err := naics.New()  // loads embedded NAICS 2022 data
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ind, ok := reg.Lookup("513210")
//	results := reg.Search("software", naics.MaxResults(10))
//	parent, _ := reg.Parent("513210")
//
// # Crosswalks
//
// The crosswalk subpackage provides SIC-to-NAICS and NAICS version mappings:
//
//	import "github.com/faisalmushtaq007/naics/crosswalk"
//
//	sic, _ := crosswalk.NewSIC()
//	mappings, ok := sic.SICToNAICS("7372")
//
//	ver, _ := crosswalk.NewVersion()
//	mappings, ok := ver.MapTo2022("511210")
//
// # Concurrency
//
// All types (Registry, crosswalk.SIC, crosswalk.Version) are safe for
// concurrent reads after construction. No synchronization is required.
package naics

import (
	"io"

	"github.com/faisalmushtaq007/naics/internal/core"
)

// Type aliases — all methods on the underlying core types are preserved.
type (
	Industry       = core.Industry
	Registry       = core.Registry
	Level          = core.Level
	SearchResult   = core.SearchResult
	SearchOption   = core.SearchOption
	SearchResponse = core.SearchResponse
)

const (
	Sector           = core.Sector
	Subsector        = core.Subsector
	IndustryGroup    = core.IndustryGroup
	NAICSIndustry    = core.NAICSIndustry
	NationalIndustry = core.NationalIndustry
)

// New creates a Registry loaded with the embedded NAICS 2022 data.
func New() (*Registry, error) {
	return core.New()
}

// NewFromCSV creates a Registry from CSV data read from r.
// The CSV must have a header row and at least two columns: code, title.
// An optional third column provides the description.
func NewFromCSV(r io.Reader) (*Registry, error) {
	return core.NewFromCSV(r)
}

// MaxResults limits the number of search results returned.
// A value of 0 or less means no limit.
func MaxResults(n int) SearchOption {
	return core.MaxResults(n)
}

// Offset skips the first n results after sorting, enabling pagination.
// Use together with MaxResults to implement paged search:
//
//	page1 := reg.Search("software", naics.MaxResults(10))
//	page2 := reg.Search("software", naics.MaxResults(10), naics.Offset(10))
func Offset(n int) SearchOption {
	return core.Offset(n)
}

// MinScore filters out results below the given relevance score.
func MinScore(score float64) SearchOption {
	return core.MinScore(score)
}

// AtLevel restricts search results to a specific NAICS hierarchy level.
// For example, AtLevel(naics.NationalIndustry) returns only 6-digit codes.
func AtLevel(l Level) SearchOption {
	return core.AtLevel(l)
}
