# naics

[![Go Reference](https://pkg.go.dev/badge/github.com/faisalmushtaq007/naics.svg)](https://pkg.go.dev/github.com/faisalmushtaq007/naics)
[![CI](https://github.com/faisalmushtaq007/naics/actions/workflows/ci.yml/badge.svg)](https://github.com/faisalmushtaq007/naics/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A zero-dependency Go library for working with [NAICS](https://www.census.gov/naics/) (North American Industry Classification System) industry codes.

Ships with embedded **NAICS 2022** data -- 2,125 industry codes baked into your binary via `go:embed`. No external files, no network calls, no setup.

## Features

- **O(1) lookup & validation** of any NAICS code
- **Keyword search** with relevance scoring, pagination, and level filtering
- **Full hierarchy** -- parent, children, descendants, ancestors, siblings, sectors
- **SIC crosswalk** -- bidirectional SIC-to-NAICS mapping
- **Version crosswalk** -- NAICS 2017 to 2022 (handles splits and merges)
- **CLI tool** for quick terminal lookups
- **Concurrent safe** -- immutable after construction, no locks needed
- **Zero dependencies** -- standard library only

## Install

```bash
go get github.com/faisalmushtaq007/naics
```

Requires **Go 1.24+**.

## Quick Start

```go
reg, _ := naics.New()

ind, ok := reg.Lookup("513210")
// ind.Code = "513210", ind.Title = "Software Publishers", ind.Level = NationalIndustry

reg.Valid("513210") // true
reg.Valid("999999") // false

results := reg.Search("software", naics.MaxResults(5))
// [{Industry: &{Code:"5132", Title:"Software Publishers", ...}, Score: 102.5}, ...]
```

## Examples

### Lookup & Validation

```go
reg, err := naics.New()
if err != nil {
    log.Fatal(err)
}

// Look up any NAICS code -- O(1) map lookup
ind, ok := reg.Lookup("513210")
if ok {
    fmt.Println(ind.Code)        // "513210"
    fmt.Println(ind.Title)       // "Software Publishers"
    fmt.Println(ind.Level)       // "National Industry"
    fmt.Println(ind.Description) // "This industry comprises establishments primarily ..."
}

// Validate codes from user input or a database
codes := []string{"513210", "111110", "INVALID", "999999"}
for _, code := range codes {
    fmt.Printf("%-8s valid=%t\n", code, reg.Valid(code))
}
// 513210   valid=true
// 111110   valid=true
// INVALID  valid=false
// 999999   valid=false

// Iterate all 2,125 codes (sorted)
for _, ind := range reg.All() {
    fmt.Printf("%s %s\n", ind.Code, ind.Title)
}
```

### Search

```go
// Basic keyword search -- results ranked by relevance
results := reg.Search("software")
for _, r := range results {
    fmt.Printf("[%.0f] %s - %s\n", r.Score, r.Industry.Code, r.Industry.Title)
}
// [102.5] 5132  - Software Publishers
// [102.0] 51321 - Software Publishers
// [101.5] 513210 - Software Publishers
// ...

// Limit results
top3 := reg.Search("manufacturing", naics.MaxResults(3))

// Filter by hierarchy level -- only 6-digit codes
national := reg.Search("software", naics.AtLevel(naics.NationalIndustry))

// Filter by minimum relevance score
highConf := reg.Search("software", naics.MinScore(50.0))

// Combine options
results = reg.Search("farming",
    naics.MaxResults(10),
    naics.AtLevel(naics.NationalIndustry),
    naics.MinScore(5.0),
)
```

### Paginated Search

```go
// Page 1
page1 := reg.SearchPage("farming", naics.MaxResults(10))
fmt.Printf("Showing %d of %d matches\n", len(page1.Results), page1.Total)
// Showing 10 of 55 matches

// Page 2
page2 := reg.SearchPage("farming", naics.MaxResults(10), naics.Offset(10))
fmt.Printf("Page 2: %d results, more=%t\n", len(page2.Results), page2.HasMore())
// Page 2: 10 results, more=true

// Build a generic paginator
pageSize := 10
for page := 0; ; page++ {
    resp := reg.SearchPage("farming", naics.MaxResults(pageSize), naics.Offset(page*pageSize))
    for _, r := range resp.Results {
        fmt.Printf("  %s %s\n", r.Industry.Code, r.Industry.Title)
    }
    if !resp.HasMore() {
        break
    }
}

// Just need the count? Skip materializing results entirely.
count := reg.Count("software")
fmt.Printf("%d industries match 'software'\n", count)
```

### Hierarchy Navigation

```go
// The NAICS tree has 5 levels:
//   51       Sector (2-digit)
//   513      Subsector (3-digit)
//   5132     Industry Group (4-digit)
//   51321    NAICS Industry (5-digit)
//   513210   National Industry (6-digit)

// Walk up: get the parent
parent, ok := reg.Parent("513210")
fmt.Println(parent) // "51321 - Software Publishers"

// Walk up to the top: full ancestor chain
for _, a := range reg.Ancestors("513210") {
    fmt.Printf("  %s - %s (%s)\n", a.Code, a.Title, a.Level)
}
//   51321 - Software Publishers (NAICS Industry)
//   5132  - Software Publishers (Industry Group)
//   513   - Publishing Industries (Subsector)
//   51    - Information (Sector)

// Walk down: direct children
for _, c := range reg.Children("51") {
    fmt.Printf("  %s %s\n", c.Code, c.Title)
}
//   511 Publishing Industries (except Internet)
//   512 Motion Picture and Sound Recording Industries
//   513 Publishing Industries
//   ...

// Walk down: ALL descendants (entire subtree)
desc := reg.Descendants("5132")
fmt.Printf("Industry group 5132 has %d descendants\n", len(desc))

// Siblings: same parent, same level
siblings := reg.Siblings("513")
fmt.Printf("Subsector 513 has %d siblings under sector 51\n", len(siblings))

// All 20 top-level sectors
for _, s := range reg.Sectors() {
    fmt.Printf("  %-6s %s\n", s.Code, s.Title)
}
//   11     Agriculture, Forestry, Fishing and Hunting
//   21     Mining, Quarrying, and Oil and Gas Extraction
//   ...
//   31-33  Manufacturing
//   ...
//   92     Public Administration
```

### SIC-to-NAICS Crosswalk

```go
import "github.com/faisalmushtaq007/naics/crosswalk"

sic, err := crosswalk.NewSIC()
if err != nil {
    log.Fatal(err)
}

// SIC -> NAICS (one SIC code can map to multiple NAICS codes)
mappings, ok := sic.SICToNAICS("7372")
if ok {
    for _, m := range mappings {
        fmt.Printf("SIC %s (%s) -> NAICS %s (%s)\n",
            m.SICCode, m.SICTitle, m.NAICSCode, m.NAICSTitle)
    }
}

// NAICS -> SIC (reverse lookup)
mappings, ok = sic.NAICSToSIC("513210")
if ok {
    for _, m := range mappings {
        fmt.Printf("NAICS %s -> SIC %s (%s)\n", m.NAICSCode, m.SICCode, m.SICTitle)
    }
}

fmt.Printf("Total SIC-NAICS mappings: %d\n", sic.Len())
```

### NAICS Version Crosswalk (2017 to 2022)

```go
import "github.com/faisalmushtaq007/naics/crosswalk"

ver, err := crosswalk.NewVersion()
if err != nil {
    log.Fatal(err)
}

// 2017 -> 2022 (an industry may have been split into multiple codes)
mappings, ok := ver.MapTo2022("511210")
if ok {
    for _, m := range mappings {
        fmt.Printf("2017 %s (%s) -> 2022 %s (%s)\n",
            m.Code2017, m.Title2017, m.Code2022, m.Title2022)
    }
}
// 2017 511210 (Software Publishers) -> 2022 513210 (Software Publishers)

// 2022 -> 2017 (reverse: an industry may have been merged from multiple codes)
mappings, ok = ver.MapTo2017("513210")
```

### Loading External Data

```go
// Load your own NAICS data from any io.Reader
f, err := os.Open("custom_naics.csv")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

reg, err := naics.NewFromCSV(f)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Loaded %d custom codes\n", reg.Len())
```

CSV format (header row required):

```csv
code,title,description
51,Information,The Information sector
513210,Software Publishers,This industry comprises establishments...
```

The `description` column is optional.

## CLI

```bash
go install github.com/faisalmushtaq007/naics/cmd/naics@latest
```

```bash
$ naics lookup 513210
Code:        513210
Title:       Software Publishers
Level:       National Industry
Description: This industry comprises establishments primarily ...

$ naics search -n 5 farming
Results 1-5 of 55 for "farming":

  1111     Oilseed and Grain Farming (Industry Group)
  1112     Vegetable and Melon Farming (Industry Group)
  1113     Fruit and Tree Nut Farming (Industry Group)
  1119     Other Crop Farming (Industry Group)
  1121     Cattle Ranching and Farming (Industry Group)

  Page 1 of 11. Next: naics search -n 5 -p 2 farming

$ naics count software
8 results for "software"

$ naics validate 513210 999999
  513210  valid
  999999  INVALID

$ naics tree 5132
5132 Software Publishers
  51321 Software Publishers
    513210 Software Publishers

$ naics parent 513210
Ancestor chain for 513210:
  51321 Software Publishers (NAICS Industry)
  5132 Software Publishers (Industry Group)
  513 Publishing Industries (Subsector)
  51 Information (Sector)

$ naics sectors
NAICS 2022 Sectors (20):

  11     Agriculture, Forestry, Fishing and Hunting
  21     Mining, Quarrying, and Oil and Gas Extraction
  ...

$ naics sic 7372
SIC 7372 -> NAICS mappings:

  511210   Software Publishers

$ naics version 511210
NAICS 2017 511210 -> 2022 mappings:

  513210   Software Publishers
```

All commands:

| Command | Description |
|---------|-------------|
| `lookup <code>` | Look up an industry by NAICS code |
| `search [-n limit] [-p page] <query>` | Search with pagination |
| `count <query>` | Count matching industries |
| `validate <code> [codes...]` | Validate codes (exits 1 if any invalid) |
| `tree <code>` | Display hierarchy tree |
| `parent <code>` | Show ancestor chain to sector |
| `sectors` | List all 20 NAICS sectors |
| `sic <sic_code>` | SIC-to-NAICS crosswalk |
| `version <2017_code>` | Map 2017 NAICS code to 2022 |

## API Reference

<details>
<summary>Full type and method listing</summary>

### Registry

| Method | Description |
|--------|-------------|
| `New() (*Registry, error)` | Load embedded NAICS 2022 data |
| `NewFromCSV(r io.Reader) (*Registry, error)` | Load from custom CSV |
| `Lookup(code) (*Industry, bool)` | O(1) code lookup |
| `Valid(code) bool` | Check if code exists |
| `All() []*Industry` | All codes sorted |
| `Len() int` | Total code count |
| `Search(query, ...SearchOption) []SearchResult` | Keyword search |
| `SearchPage(query, ...SearchOption) SearchResponse` | Search with pagination metadata |
| `Count(query, ...SearchOption) int` | Count matches |
| `Parent(code) (*Industry, bool)` | One level up |
| `Children(code) []*Industry` | One level down |
| `Descendants(code) []*Industry` | Entire subtree |
| `Ancestors(code) []*Industry` | Chain up to sector |
| `Siblings(code) []*Industry` | Same parent, same level |
| `Sectors() []*Industry` | All 20 top-level sectors |

### Search Options

| Option | Description |
|--------|-------------|
| `MaxResults(n)` | Limit results per page |
| `Offset(n)` | Skip first N results |
| `MinScore(s)` | Minimum relevance score |
| `AtLevel(l)` | Filter by hierarchy level |

### Types

```go
type Industry struct {
    Code        string
    Title       string
    Description string
    Level       Level
}

type Level int
// Sector=2, Subsector=3, IndustryGroup=4, NAICSIndustry=5, NationalIndustry=6

type SearchResult struct {
    Industry *Industry
    Score    float64
}

type SearchResponse struct {
    Results []SearchResult
    Total   int
    Offset  int
    Limit   int
}
func (r SearchResponse) HasMore() bool
```

### Crosswalk (`crosswalk` subpackage)

| Function/Method | Description |
|-----------------|-------------|
| `NewSIC() (*SIC, error)` | Load SIC-NAICS mappings |
| `NewSICFromCSV(r) (*SIC, error)` | Load from custom CSV |
| `SICToNAICS(sic) ([]SICMapping, bool)` | SIC to NAICS |
| `NAICSToSIC(naics) ([]SICMapping, bool)` | NAICS to SIC |
| `NewVersion() (*Version, error)` | Load 2017-2022 mappings |
| `NewVersionFromCSV(r) (*Version, error)` | Load from custom CSV |
| `MapTo2022(code) ([]VersionMapping, bool)` | 2017 to 2022 |
| `MapTo2017(code) ([]VersionMapping, bool)` | 2022 to 2017 |

</details>

## Project Structure

```
naics/
├── naics.go                 # Public API (thin facade)
├── example_test.go          # Godoc examples
├── crosswalk/               # SIC + version crosswalk subpackage
│   ├── sic.go
│   └── version.go
├── cmd/naics/               # CLI tool
├── internal/
│   ├── core/                # Implementation (types, registry, search, hierarchy)
│   ├── data/                # Embedded CSV data (NAICS 2022, SIC, version concordance)
│   └── cmd/fetchdata/       # Utility to refresh data from upstream
├── Makefile
├── go.mod
├── LICENSE
└── README.md
```

## Benchmarks

```
BenchmarkNew             ~1.5 ms/op
BenchmarkLookup           ~9 ns/op     0 allocs/op
BenchmarkValid            ~9 ns/op     0 allocs/op
BenchmarkSearch_Short    ~3 ms/op
BenchmarkParent          ~27 ns/op     0 allocs/op
BenchmarkChildren        ~4 µs/op      4 allocs/op
BenchmarkAncestors      ~570 ns/op     3 allocs/op
BenchmarkSectors         ~7 µs/op      6 allocs/op
```

Run locally: `make bench`

## Contributing

```bash
git clone https://github.com/faisalmushtaq007/naics.git
cd naics
make test        # run tests
make lint        # vet + fmt check
make bench       # benchmarks
make build       # build CLI
```

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Run `make` (runs lint, test, build)
4. Commit and open a pull request

## License

[MIT](LICENSE)
