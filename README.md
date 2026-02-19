# naics

A zero-dependency Go library for working with [NAICS](https://www.census.gov/naics/) (North American Industry Classification System) codes.

Includes embedded **NAICS 2022** data (2,125 industry codes) compiled into the binary via `go:embed` -- no external files or network access needed at runtime.

## Features

- **Lookup & validation** -- O(1) code lookup, instant validity checks
- **Keyword search** -- case-insensitive search across titles and descriptions with relevance scoring
- **Hierarchy navigation** -- parent, children, descendants, ancestors, siblings, sectors
- **Range-sector support** -- correctly handles Manufacturing (31-33), Retail Trade (44-45), Transportation (48-49)
- **SIC-to-NAICS crosswalk** -- bidirectional mapping between legacy SIC codes and NAICS codes
- **NAICS version crosswalk** -- map between NAICS 2017 and 2022 codes (handles splits and merges)
- **CLI tool** -- `naics` command for terminal lookups, search, validation, and hierarchy browsing
- **External data loading** -- load custom NAICS data from any CSV `io.Reader`
- **Concurrent safe** -- all registries are immutable after construction
- **Zero dependencies** -- standard library only

## Installation

```bash
go get github.com/faisalmushtaq007/naics
```

Requires Go 1.22 or later.

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/faisalmushtaq007/naics"
)

func main() {
    reg, err := naics.New()
    if err != nil {
        log.Fatal(err)
    }

    // Lookup by code
    ind, ok := reg.Lookup("513210")
    if ok {
        fmt.Printf("%s - %s (%s)\n", ind.Code, ind.Title, ind.Level)
        // 513210 - Software Publishers (National Industry)
    }

    // Validate a code
    fmt.Println(reg.Valid("513210")) // true
    fmt.Println(reg.Valid("999999")) // false

    // Search by keyword
    results := reg.Search("software", naics.MaxResults(5))
    for _, r := range results {
        fmt.Printf("  %-6s %s\n", r.Industry.Code, r.Industry.Title)
    }

    // Navigate hierarchy
    parent, _ := reg.Parent("513210")
    fmt.Printf("Parent: %s\n", parent) // 51321 - Software Publishers

    ancestors := reg.Ancestors("513210")
    for _, a := range ancestors {
        fmt.Printf("  %s (%s)\n", a, a.Level)
    }

    children := reg.Children("51")
    fmt.Printf("Sector 51 has %d subsectors\n", len(children))

    sectors := reg.Sectors()
    fmt.Printf("Total sectors: %d\n", len(sectors))
}
```

## API Reference

### Creating a Registry

| Function | Description |
|----------|-------------|
| `New() (*Registry, error)` | Load embedded NAICS 2022 data |
| `NewFromCSV(r io.Reader) (*Registry, error)` | Load from custom CSV (columns: code, title, description) |

### Lookup & Validation

| Method | Description |
|--------|-------------|
| `Lookup(code string) (*Industry, bool)` | O(1) lookup by NAICS code |
| `Valid(code string) bool` | Check if a code exists |
| `All() []*Industry` | All industries sorted by code |
| `Len() int` | Total number of codes |

### Search

| Method | Description |
|--------|-------------|
| `Search(query string, opts ...SearchOption) []SearchResult` | Keyword search with relevance ranking |
| `MaxResults(n int) SearchOption` | Limit number of results |

### Hierarchy

| Method | Description |
|--------|-------------|
| `Parent(code string) (*Industry, bool)` | Immediate parent (one level up) |
| `Children(code string) []*Industry` | Direct children (one level down) |
| `Descendants(code string) []*Industry` | All codes below in hierarchy |
| `Ancestors(code string) []*Industry` | Chain from parent up to sector |
| `Siblings(code string) []*Industry` | Codes sharing the same parent |
| `Sectors() []*Industry` | All 20 top-level sectors |

### SIC Crosswalk (`crosswalk` subpackage)

| Function/Method | Description |
|-----------------|-------------|
| `crosswalk.NewSIC() (*SIC, error)` | Load embedded SIC-NAICS mapping |
| `crosswalk.NewSICFromCSV(r io.Reader) (*SIC, error)` | Load from custom CSV |
| `SICToNAICS(sicCode string) ([]SICMapping, bool)` | SIC to NAICS (may return multiple) |
| `NAICSToSIC(naicsCode string) ([]SICMapping, bool)` | NAICS to SIC (reverse) |

### Version Crosswalk (`crosswalk` subpackage)

| Function/Method | Description |
|-----------------|-------------|
| `crosswalk.NewVersion() (*Version, error)` | Load embedded 2017-2022 concordance |
| `crosswalk.NewVersionFromCSV(r io.Reader) (*Version, error)` | Load from custom CSV |
| `MapTo2022(code2017 string) ([]VersionMapping, bool)` | Map 2017 code to 2022 |
| `MapTo2017(code2022 string) ([]VersionMapping, bool)` | Map 2022 code to 2017 |

### Types

```go
type Industry struct {
    Code        string  // 2-6 digit NAICS code (e.g., "513210")
    Title       string  // Official name (e.g., "Software Publishers")
    Description string  // Detailed description
    Level       Level   // Sector, Subsector, IndustryGroup, NAICSIndustry, NationalIndustry
}

type SearchResult struct {
    Industry *Industry
    Score    float64   // Relevance score (higher = better match)
}

type Level int // Sector=2, Subsector=3, IndustryGroup=4, NAICSIndustry=5, NationalIndustry=6
```

## NAICS Hierarchy

```
Sector (2-digit)         e.g., 51 - Information
  Subsector (3-digit)    e.g., 513 - Publishing Industries
    Industry Group (4)   e.g., 5132 - Software Publishers
      NAICS Industry (5) e.g., 51321 - Software Publishers
        National (6)     e.g., 513210 - Software Publishers
```

Three sectors use range codes: **Manufacturing (31-33)**, **Retail Trade (44-45)**, **Transportation (48-49)**.

## Loading External Data

```go
f, _ := os.Open("custom_naics.csv")
defer f.Close()

reg, err := naics.NewFromCSV(f)
```

The CSV must have a header row with at least `code` and `title` columns. An optional `description` column is supported.

## SIC-to-NAICS Crosswalk

Map between legacy SIC (Standard Industrial Classification) codes and NAICS codes:

```go
import "github.com/faisalmushtaq007/naics/crosswalk"

sic, err := crosswalk.NewSIC()
if err != nil {
    log.Fatal(err)
}

// SIC -> NAICS (one SIC code may map to multiple NAICS codes)
mappings, ok := sic.SICToNAICS("7372")
for _, m := range mappings {
    fmt.Printf("SIC %s -> NAICS %s (%s)\n", m.SICCode, m.NAICSCode, m.NAICSTitle)
}

// NAICS -> SIC (reverse lookup)
mappings, ok = sic.NAICSToSIC("513210")
```

## NAICS Version Crosswalk (2017 to 2022)

Map codes between NAICS 2017 and 2022 revisions:

```go
import "github.com/faisalmushtaq007/naics/crosswalk"

ver, err := crosswalk.NewVersion()
if err != nil {
    log.Fatal(err)
}

// 2017 -> 2022 (handles splits: one 2017 code may map to multiple 2022 codes)
mappings, ok := ver.MapTo2022("511210")
// Returns: 513210 "Software Publishers"

// 2022 -> 2017 (reverse: handles merges)
mappings, ok = ver.MapTo2017("513210")
```

## CLI Tool

Install the command-line tool:

```bash
go install github.com/faisalmushtaq007/naics/cmd/naics@latest
```

Usage:

```bash
naics lookup 513210              # Look up an industry by code
naics search "software"          # Search by keyword
naics validate 513210 999999     # Validate codes (exits 1 if any invalid)
naics tree 51                    # Display hierarchy tree
naics parent 513210              # Show ancestor chain to sector
naics sectors                    # List all 20 sectors
naics sic 7372                   # SIC-to-NAICS crosswalk
naics version 511210             # Map 2017 NAICS code to 2022
```

## Refreshing Embedded Data

A utility is included to regenerate the embedded CSV from upstream sources:

```bash
go run ./internal/cmd/fetchdata
```

This downloads the latest NAICS data and writes it to `internal/data/naics2022.csv`.

## Benchmarks

```
BenchmarkLookup           ~9 ns/op     0 allocs/op
BenchmarkValid            ~9 ns/op     0 allocs/op
BenchmarkParent          ~27 ns/op     0 allocs/op
BenchmarkAncestors      ~570 ns/op     3 allocs/op
BenchmarkChildren        ~4 µs/op      4 allocs/op
BenchmarkSectors         ~7 µs/op      6 allocs/op
BenchmarkSearch_Short    ~3 ms/op
BenchmarkNew             ~1.5 ms/op
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Run tests (`go test -race ./...`)
4. Commit your changes
5. Open a pull request

## License

[MIT](LICENSE)
