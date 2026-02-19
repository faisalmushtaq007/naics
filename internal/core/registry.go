package core

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/faisalmushtaq007/naics/internal/data"
)

// Registry holds a collection of NAICS codes and provides lookup, search,
// and hierarchy navigation. A Registry is safe for concurrent use after
// construction; all mutating operations happen during New/NewFromCSV.
type Registry struct {
	codes map[string]*Industry
	list  []*Industry // sorted by code
}

// New creates a Registry loaded with the embedded NAICS 2022 data.
func New() (*Registry, error) {
	f, err := data.CSV.Open("naics2022.csv")
	if err != nil {
		return nil, fmt.Errorf("naics: open embedded data: %w", err)
	}
	defer f.Close()

	return newFromReader(f)
}

// NewFromCSV creates a Registry from CSV data read from r.
// The CSV must have a header row and at least two columns: code, title.
// An optional third column provides the description.
func NewFromCSV(r io.Reader) (*Registry, error) {
	return newFromReader(r)
}

func newFromReader(r io.Reader) (*Registry, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("naics: read CSV header: %w", err)
	}
	if len(header) < 2 {
		return nil, fmt.Errorf("naics: CSV must have at least 2 columns (code, title), got %d", len(header))
	}

	codes := make(map[string]*Industry, 2200)
	var list []*Industry

	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("naics: read CSV row: %w", err)
		}
		if len(row) < 2 {
			continue
		}

		code := strings.TrimSpace(row[0])
		title := strings.TrimSpace(row[1])
		if code == "" || title == "" {
			continue
		}

		desc := ""
		if len(row) >= 3 {
			desc = strings.TrimSpace(row[2])
		}

		lvl := codeLevel(code)
		if lvl < 0 {
			continue
		}

		ind := &Industry{
			Code:        code,
			Title:       title,
			Description: desc,
			Level:       lvl,
		}
		codes[code] = ind
		list = append(list, ind)
	}

	if len(codes) == 0 {
		return nil, fmt.Errorf("naics: no valid records found in CSV")
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Code < list[j].Code
	})

	return &Registry{codes: codes, list: list}, nil
}

// Lookup returns the Industry for the given NAICS code.
// The second return value indicates whether the code was found.
func (reg *Registry) Lookup(code string) (*Industry, bool) {
	ind, ok := reg.codes[code]
	return ind, ok
}

// Valid reports whether code is a known NAICS code in this registry.
func (reg *Registry) Valid(code string) bool {
	_, ok := reg.codes[code]
	return ok
}

// All returns a slice of all industries sorted by code.
// The returned slice is a copy; callers may modify it freely.
func (reg *Registry) All() []*Industry {
	out := make([]*Industry, len(reg.list))
	copy(out, reg.list)
	return out
}

// Len returns the total number of industry codes in the registry.
func (reg *Registry) Len() int {
	return len(reg.codes)
}
