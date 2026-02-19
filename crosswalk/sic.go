// Package crosswalk provides mappings between NAICS and other industry
// classification systems (SIC codes, NAICS version changes).
//
// All types are safe for concurrent reads after construction.
package crosswalk

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/faisalmushtaq007/naics/internal/data"
)

// SICMapping represents a single mapping between a SIC code and a NAICS code.
type SICMapping struct {
	NAICSCode  string
	NAICSTitle string
	SICCode    string
	SICTitle   string
}

// SIC holds bidirectional SIC-to-NAICS mappings.
type SIC struct {
	sicToNAICS map[string][]SICMapping
	naicsToSIC map[string][]SICMapping
	mappings   []SICMapping
}

// NewSIC creates a SIC crosswalk loaded with the embedded SIC-NAICS data.
func NewSIC() (*SIC, error) {
	f, err := data.SICCrosswalk.Open("sic_naics.csv")
	if err != nil {
		return nil, fmt.Errorf("crosswalk: open embedded SIC data: %w", err)
	}
	defer f.Close()

	return newSICFromReader(f)
}

// NewSICFromCSV creates a SIC crosswalk from CSV data read from r.
// The CSV must have a header row and columns: naics_code, naics_title, sic_code, sic_title.
func NewSICFromCSV(r io.Reader) (*SIC, error) {
	return newSICFromReader(r)
}

func newSICFromReader(r io.Reader) (*SIC, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("crosswalk: read SIC CSV header: %w", err)
	}
	if len(header) < 4 {
		return nil, fmt.Errorf("crosswalk: SIC CSV must have 4 columns, got %d", len(header))
	}

	sicToNAICS := make(map[string][]SICMapping)
	naicsToSIC := make(map[string][]SICMapping)
	var mappings []SICMapping

	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("crosswalk: read SIC CSV row: %w", err)
		}
		if len(row) < 4 {
			continue
		}

		m := SICMapping{
			NAICSCode:  strings.TrimSpace(row[0]),
			NAICSTitle: strings.TrimSpace(row[1]),
			SICCode:    strings.TrimSpace(row[2]),
			SICTitle:   strings.TrimSpace(row[3]),
		}
		if m.NAICSCode == "" || m.SICCode == "" {
			continue
		}

		mappings = append(mappings, m)
		sicToNAICS[m.SICCode] = append(sicToNAICS[m.SICCode], m)
		naicsToSIC[m.NAICSCode] = append(naicsToSIC[m.NAICSCode], m)
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("crosswalk: no valid SIC records found")
	}

	return &SIC{
		sicToNAICS: sicToNAICS,
		naicsToSIC: naicsToSIC,
		mappings:   mappings,
	}, nil
}

// SICToNAICS returns all NAICS mappings for a given SIC code.
// One SIC code may map to multiple NAICS codes (splits).
// Returns false if the SIC code has no mappings.
func (s *SIC) SICToNAICS(sicCode string) ([]SICMapping, bool) {
	m, ok := s.sicToNAICS[sicCode]
	if !ok || len(m) == 0 {
		return nil, false
	}
	out := make([]SICMapping, len(m))
	copy(out, m)
	return out, true
}

// NAICSToSIC returns all SIC mappings for a given NAICS code.
// One NAICS code may map to multiple SIC codes (merges).
// Returns false if the NAICS code has no mappings.
func (s *SIC) NAICSToSIC(naicsCode string) ([]SICMapping, bool) {
	m, ok := s.naicsToSIC[naicsCode]
	if !ok || len(m) == 0 {
		return nil, false
	}
	out := make([]SICMapping, len(m))
	copy(out, m)
	return out, true
}

// AllMappings returns all SIC-NAICS mappings.
func (s *SIC) AllMappings() []SICMapping {
	out := make([]SICMapping, len(s.mappings))
	copy(out, s.mappings)
	return out
}

// Len returns the total number of crosswalk mapping records.
func (s *SIC) Len() int {
	return len(s.mappings)
}
