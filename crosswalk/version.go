package crosswalk

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/faisalmushtaq007/naics/internal/data"
)

// VersionMapping represents a single mapping between a NAICS 2017 code
// and a NAICS 2022 code.
type VersionMapping struct {
	Code2017  string
	Title2017 string
	Code2022  string
	Title2022 string
}

// Version holds bidirectional NAICS 2017-to-2022 mappings.
type Version struct {
	to2022   map[string][]VersionMapping
	to2017   map[string][]VersionMapping
	mappings []VersionMapping
}

// NewVersion creates a Version crosswalk loaded with the embedded
// NAICS 2017-to-2022 concordance data.
func NewVersion() (*Version, error) {
	f, err := data.NAICSVersionCrosswalk.Open("naics_2017_to_2022.csv")
	if err != nil {
		return nil, fmt.Errorf("crosswalk: open embedded version data: %w", err)
	}
	defer f.Close()

	return newVersionFromReader(f)
}

// NewVersionFromCSV creates a Version crosswalk from CSV data read from r.
// The CSV must have a header row and columns: naics_2017, title_2017, naics_2022, title_2022.
func NewVersionFromCSV(r io.Reader) (*Version, error) {
	return newVersionFromReader(r)
}

func newVersionFromReader(r io.Reader) (*Version, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("crosswalk: read version CSV header: %w", err)
	}
	if len(header) < 4 {
		return nil, fmt.Errorf("crosswalk: version CSV must have 4 columns, got %d", len(header))
	}

	to2022 := make(map[string][]VersionMapping)
	to2017 := make(map[string][]VersionMapping)
	var mappings []VersionMapping

	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("crosswalk: read version CSV row: %w", err)
		}
		if len(row) < 4 {
			continue
		}

		m := VersionMapping{
			Code2017:  strings.TrimSpace(row[0]),
			Title2017: strings.TrimSpace(row[1]),
			Code2022:  strings.TrimSpace(row[2]),
			Title2022: strings.TrimSpace(row[3]),
		}
		if m.Code2017 == "" || m.Code2022 == "" {
			continue
		}

		mappings = append(mappings, m)
		to2022[m.Code2017] = append(to2022[m.Code2017], m)
		to2017[m.Code2022] = append(to2017[m.Code2022], m)
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("crosswalk: no valid version records found")
	}

	return &Version{
		to2022:   to2022,
		to2017:   to2017,
		mappings: mappings,
	}, nil
}

// MapTo2022 returns the NAICS 2022 codes that a NAICS 2017 code maps to.
// One 2017 code may map to multiple 2022 codes when an industry was split.
// Returns false if the 2017 code has no mappings.
func (v *Version) MapTo2022(code2017 string) ([]VersionMapping, bool) {
	m, ok := v.to2022[code2017]
	if !ok || len(m) == 0 {
		return nil, false
	}
	out := make([]VersionMapping, len(m))
	copy(out, m)
	return out, true
}

// MapTo2017 returns the NAICS 2017 codes that a NAICS 2022 code was derived from.
// One 2022 code may map to multiple 2017 codes when industries were merged.
// Returns false if the 2022 code has no mappings.
func (v *Version) MapTo2017(code2022 string) ([]VersionMapping, bool) {
	m, ok := v.to2017[code2022]
	if !ok || len(m) == 0 {
		return nil, false
	}
	out := make([]VersionMapping, len(m))
	copy(out, m)
	return out, true
}

// AllMappings returns all version crosswalk mappings.
func (v *Version) AllMappings() []VersionMapping {
	out := make([]VersionMapping, len(v.mappings))
	copy(out, v.mappings)
	return out
}

// Len returns the total number of version crosswalk mapping records.
func (v *Version) Len() int {
	return len(v.mappings)
}
