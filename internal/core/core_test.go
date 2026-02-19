package core

import (
	"strings"
	"testing"
)

func mustRegistry(t *testing.T) *Registry {
	t.Helper()
	reg, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return reg
}

func TestNew(t *testing.T) {
	reg, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if reg.Len() == 0 {
		t.Fatal("registry is empty")
	}
	if reg.Len() < 2000 {
		t.Errorf("expected at least 2000 codes, got %d", reg.Len())
	}
}

func TestNewFromCSV(t *testing.T) {
	csv := `code,title,description
51,Information,The Information sector
513,Publishing Industries,Publishing subsector
5132,Software Publishers,Software publishing group
51321,Software Publishers,Software publishing industry
513210,Software Publishers,This industry comprises software publishers`

	reg, err := NewFromCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("NewFromCSV() error: %v", err)
	}
	if reg.Len() != 5 {
		t.Errorf("expected 5 codes, got %d", reg.Len())
	}
}

func TestNewFromCSV_MinimalColumns(t *testing.T) {
	csv := `code,title
51,Information
513,Publishing`

	reg, err := NewFromCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("NewFromCSV() error: %v", err)
	}
	if reg.Len() != 2 {
		t.Errorf("expected 2 codes, got %d", reg.Len())
	}
	ind, ok := reg.Lookup("51")
	if !ok {
		t.Fatal("expected to find code 51")
	}
	if ind.Description != "" {
		t.Errorf("expected empty description, got %q", ind.Description)
	}
}

func TestNewFromCSV_EmptyData(t *testing.T) {
	csv := `code,title,description`
	_, err := NewFromCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for empty CSV, got nil")
	}
}

func TestNewFromCSV_TooFewColumns(t *testing.T) {
	csv := `code`
	_, err := NewFromCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for CSV with too few columns, got nil")
	}
}

func TestLookup(t *testing.T) {
	reg := mustRegistry(t)

	tests := []struct {
		code      string
		wantFound bool
		wantTitle string
	}{
		{"51", true, "Information"},
		{"513210", true, "Software Publishers"},
		{"11", true, "Agriculture, Forestry, Fishing and Hunting"},
		{"31-33", true, "Manufacturing"},
		{"44-45", true, "Retail Trade"},
		{"48-49", true, "Transportation and Warehousing"},
		{"111110", true, "Soybean Farming"},
		{"999999", false, ""},
		{"", false, ""},
		{"X", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			ind, ok := reg.Lookup(tt.code)
			if ok != tt.wantFound {
				t.Errorf("Lookup(%q) found = %v, want %v", tt.code, ok, tt.wantFound)
			}
			if ok && ind.Title != tt.wantTitle {
				t.Errorf("Lookup(%q) title = %q, want %q", tt.code, ind.Title, tt.wantTitle)
			}
		})
	}
}

func TestValid(t *testing.T) {
	reg := mustRegistry(t)

	if !reg.Valid("513210") {
		t.Error("expected 513210 to be valid")
	}
	if !reg.Valid("111110") {
		t.Error("expected 111110 to be valid")
	}
	if reg.Valid("999999") {
		t.Error("expected 999999 to be invalid")
	}
	if reg.Valid("") {
		t.Error("expected empty string to be invalid")
	}
}

func TestAll(t *testing.T) {
	reg := mustRegistry(t)
	all := reg.All()

	if len(all) != reg.Len() {
		t.Errorf("All() returned %d, Len() = %d", len(all), reg.Len())
	}

	for i := 1; i < len(all); i++ {
		if all[i].Code < all[i-1].Code {
			t.Errorf("All() not sorted: %s appears after %s", all[i].Code, all[i-1].Code)
			break
		}
	}
}

func TestSearch(t *testing.T) {
	reg := mustRegistry(t)

	tests := []struct {
		query       string
		wantMinHits int
		wantFirst   string
	}{
		{"Software Publishers", 1, "5132"},
		{"software", 1, ""},
		{"agriculture", 1, ""},
		{"soybean farming", 1, ""},
		{"", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := reg.Search(tt.query)
			if len(results) < tt.wantMinHits {
				t.Errorf("Search(%q) got %d results, want at least %d", tt.query, len(results), tt.wantMinHits)
			}
			if tt.wantFirst != "" && len(results) > 0 && results[0].Industry.Code != tt.wantFirst {
				t.Errorf("Search(%q) first result = %s, want %s", tt.query, results[0].Industry.Code, tt.wantFirst)
			}
		})
	}
}

func TestSearch_MaxResults(t *testing.T) {
	reg := mustRegistry(t)
	results := reg.Search("farming", MaxResults(5))
	if len(results) > 5 {
		t.Errorf("MaxResults(5) returned %d results", len(results))
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	reg := mustRegistry(t)

	lower := reg.Search("software publishers")
	upper := reg.Search("SOFTWARE PUBLISHERS")
	mixed := reg.Search("Software Publishers")

	if len(lower) == 0 || len(upper) == 0 || len(mixed) == 0 {
		t.Fatal("expected results for all case variants")
	}
	if lower[0].Industry.Code != upper[0].Industry.Code || lower[0].Industry.Code != mixed[0].Industry.Code {
		t.Error("case-insensitive search returned different first results")
	}
}

func TestSearch_Offset(t *testing.T) {
	reg := mustRegistry(t)

	all := reg.Search("farming")
	if len(all) < 5 {
		t.Fatalf("expected at least 5 results for 'farming', got %d", len(all))
	}

	page := reg.Search("farming", MaxResults(3), Offset(2))
	if len(page) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(page))
	}
	if page[0].Industry.Code != all[2].Industry.Code {
		t.Errorf("offset result[0] = %s, want %s", page[0].Industry.Code, all[2].Industry.Code)
	}
}

func TestSearch_OffsetBeyondTotal(t *testing.T) {
	reg := mustRegistry(t)

	results := reg.Search("soybean farming", Offset(9999))
	if len(results) != 0 {
		t.Errorf("offset beyond total should return empty, got %d", len(results))
	}
}

func TestSearchPage(t *testing.T) {
	reg := mustRegistry(t)

	resp := reg.SearchPage("farming", MaxResults(5))
	if resp.Total < 5 {
		t.Fatalf("expected total >= 5 for 'farming', got %d", resp.Total)
	}
	if len(resp.Results) != 5 {
		t.Errorf("expected 5 results, got %d", len(resp.Results))
	}
	if resp.Offset != 0 {
		t.Errorf("expected offset 0, got %d", resp.Offset)
	}
	if resp.Limit != 5 {
		t.Errorf("expected limit 5, got %d", resp.Limit)
	}
	if !resp.HasMore() {
		t.Error("expected HasMore() = true")
	}
}

func TestSearchPage_Pagination(t *testing.T) {
	reg := mustRegistry(t)

	page1 := reg.SearchPage("farming", MaxResults(3))
	page2 := reg.SearchPage("farming", MaxResults(3), Offset(3))

	if page1.Total != page2.Total {
		t.Errorf("total changed between pages: %d vs %d", page1.Total, page2.Total)
	}
	if len(page1.Results) == 0 || len(page2.Results) == 0 {
		t.Fatal("expected results on both pages")
	}
	if page1.Results[0].Industry.Code == page2.Results[0].Industry.Code {
		t.Error("page 1 and page 2 should have different first results")
	}
}

func TestSearchPage_EmptyQuery(t *testing.T) {
	reg := mustRegistry(t)
	resp := reg.SearchPage("")
	if resp.Total != 0 || len(resp.Results) != 0 {
		t.Error("empty query should return empty response")
	}
}

func TestSearchPage_LastPage(t *testing.T) {
	reg := mustRegistry(t)
	resp := reg.SearchPage("soybean farming")
	if resp.Total == 0 {
		t.Fatal("expected results for 'soybean farming'")
	}

	last := reg.SearchPage("soybean farming", Offset(resp.Total-1), MaxResults(10))
	if len(last.Results) != 1 {
		t.Errorf("last page should have 1 result, got %d", len(last.Results))
	}
	if last.HasMore() {
		t.Error("last page should not have more")
	}
}

func TestCount(t *testing.T) {
	reg := mustRegistry(t)

	count := reg.Count("farming")
	all := reg.Search("farming")
	if count != len(all) {
		t.Errorf("Count() = %d, len(Search()) = %d", count, len(all))
	}
}

func TestCount_Empty(t *testing.T) {
	reg := mustRegistry(t)
	if reg.Count("") != 0 {
		t.Error("empty query count should be 0")
	}
}

func TestCount_NoMatch(t *testing.T) {
	reg := mustRegistry(t)
	if reg.Count("zzzzzzxxxxxnonexistent") != 0 {
		t.Error("nonexistent query count should be 0")
	}
}

func TestMinScore(t *testing.T) {
	reg := mustRegistry(t)

	all := reg.Search("software")
	filtered := reg.Search("software", MinScore(50.0))

	if len(filtered) >= len(all) {
		t.Errorf("MinScore should reduce results: all=%d, filtered=%d", len(all), len(filtered))
	}
	for _, r := range filtered {
		if r.Score < 50.0 {
			t.Errorf("result %s has score %.1f, below min 50.0", r.Industry.Code, r.Score)
		}
	}
}

func TestAtLevel(t *testing.T) {
	reg := mustRegistry(t)

	results := reg.Search("software", AtLevel(NationalIndustry))
	for _, r := range results {
		if r.Industry.Level != NationalIndustry {
			t.Errorf("result %s is level %v, want NationalIndustry", r.Industry.Code, r.Industry.Level)
		}
	}
}

func TestAtLevel_Count(t *testing.T) {
	reg := mustRegistry(t)

	allCount := reg.Count("manufacturing")
	sectorCount := reg.Count("manufacturing", AtLevel(Sector))

	if sectorCount >= allCount {
		t.Errorf("sector-only count (%d) should be less than all (%d)", sectorCount, allCount)
	}
}

func TestParent(t *testing.T) {
	reg := mustRegistry(t)

	tests := []struct {
		code       string
		wantParent string
		wantFound  bool
	}{
		{"513210", "51321", true},
		{"51321", "5132", true},
		{"5132", "513", true},
		{"513", "51", true},
		{"51", "", false},
		{"XXXXX", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			parent, ok := reg.Parent(tt.code)
			if ok != tt.wantFound {
				t.Errorf("Parent(%q) found = %v, want %v", tt.code, ok, tt.wantFound)
			}
			if ok && parent.Code != tt.wantParent {
				t.Errorf("Parent(%q) = %s, want %s", tt.code, parent.Code, tt.wantParent)
			}
		})
	}
}

func TestParent_RangeSectors(t *testing.T) {
	reg := mustRegistry(t)

	tests := []struct {
		subsector  string
		wantParent string
	}{
		{"311", "31-33"},
		{"321", "31-33"},
		{"332", "31-33"},
		{"441", "44-45"},
		{"455", "44-45"},
		{"481", "48-49"},
		{"492", "48-49"},
	}

	for _, tt := range tests {
		t.Run(tt.subsector, func(t *testing.T) {
			parent, ok := reg.Parent(tt.subsector)
			if !ok {
				t.Fatalf("Parent(%q) not found", tt.subsector)
			}
			if parent.Code != tt.wantParent {
				t.Errorf("Parent(%q) = %s, want %s", tt.subsector, parent.Code, tt.wantParent)
			}
		})
	}
}

func TestChildren(t *testing.T) {
	reg := mustRegistry(t)

	children := reg.Children("51")
	if len(children) == 0 {
		t.Fatal("expected children for sector 51")
	}
	for _, child := range children {
		if child.Level != Subsector {
			t.Errorf("child of sector should be subsector, got %v", child.Level)
		}
		if !strings.HasPrefix(child.Code, "51") {
			t.Errorf("child code %s does not start with 51", child.Code)
		}
	}
}

func TestChildren_RangeSector(t *testing.T) {
	reg := mustRegistry(t)

	children := reg.Children("31-33")
	if len(children) == 0 {
		t.Fatal("expected children for sector 31-33")
	}
	for _, child := range children {
		if child.Level != Subsector {
			t.Errorf("child of sector should be subsector, got %v for %s", child.Level, child.Code)
		}
		prefix := child.Code[:2]
		if prefix != "31" && prefix != "32" && prefix != "33" {
			t.Errorf("child %s doesn't belong to 31-33 sector", child.Code)
		}
	}
}

func TestChildren_NationalIndustry(t *testing.T) {
	reg := mustRegistry(t)
	children := reg.Children("513210")
	if len(children) != 0 {
		t.Errorf("6-digit code should have no children, got %d", len(children))
	}
}

func TestChildren_InvalidCode(t *testing.T) {
	reg := mustRegistry(t)
	children := reg.Children("XXXXXX")
	if children != nil {
		t.Errorf("invalid code should return nil, got %v", children)
	}
}

func TestDescendants(t *testing.T) {
	reg := mustRegistry(t)

	desc := reg.Descendants("513")
	if len(desc) == 0 {
		t.Fatal("expected descendants for 513")
	}
	for _, d := range desc {
		if !strings.HasPrefix(d.Code, "513") {
			t.Errorf("descendant %s does not start with 513", d.Code)
		}
		if d.Code == "513" {
			t.Error("descendants should not include the code itself")
		}
	}
}

func TestDescendants_InvalidCode(t *testing.T) {
	reg := mustRegistry(t)
	desc := reg.Descendants("XXXXXX")
	if desc != nil {
		t.Error("invalid code should return nil")
	}
}

func TestAncestors(t *testing.T) {
	reg := mustRegistry(t)

	ancestors := reg.Ancestors("513210")
	if len(ancestors) == 0 {
		t.Fatal("expected ancestors for 513210")
	}

	expectedCodes := []string{"51321", "5132", "513", "51"}
	if len(ancestors) != len(expectedCodes) {
		t.Fatalf("expected %d ancestors, got %d: %v", len(expectedCodes), len(ancestors), ancestors)
	}
	for i, want := range expectedCodes {
		if ancestors[i].Code != want {
			t.Errorf("ancestor[%d] = %s, want %s", i, ancestors[i].Code, want)
		}
	}
}

func TestAncestors_Sector(t *testing.T) {
	reg := mustRegistry(t)
	ancestors := reg.Ancestors("51")
	if len(ancestors) != 0 {
		t.Errorf("sector should have no ancestors, got %d", len(ancestors))
	}
}

func TestAncestors_RangeSector(t *testing.T) {
	reg := mustRegistry(t)
	ancestors := reg.Ancestors("311")
	if len(ancestors) == 0 {
		t.Fatal("expected ancestors for 311")
	}
	if ancestors[0].Code != "31-33" {
		t.Errorf("parent of 311 should be 31-33, got %s", ancestors[0].Code)
	}
}

func TestSiblings(t *testing.T) {
	reg := mustRegistry(t)

	siblings := reg.Siblings("513")
	if len(siblings) == 0 {
		t.Fatal("expected siblings for 513")
	}
	for _, sib := range siblings {
		if sib.Code == "513" {
			t.Error("siblings should not include the code itself")
		}
		if sib.Level != Subsector {
			t.Errorf("sibling should be at subsector level, got %v", sib.Level)
		}
	}
}

func TestSectors(t *testing.T) {
	reg := mustRegistry(t)
	sectors := reg.Sectors()

	if len(sectors) < 17 {
		t.Errorf("expected at least 17 sectors (20 logical), got %d", len(sectors))
	}

	for _, s := range sectors {
		if s.Level != Sector {
			t.Errorf("Sectors() returned non-sector: %s (level=%v)", s.Code, s.Level)
		}
	}

	sectorCodes := make(map[string]bool)
	for _, s := range sectors {
		sectorCodes[s.Code] = true
	}
	for _, expected := range []string{"11", "21", "22", "23", "31-33", "42", "44-45", "48-49", "51", "52", "53", "54", "55", "56", "61", "62", "71", "72", "81", "92"} {
		if !sectorCodes[expected] {
			t.Errorf("missing expected sector %s", expected)
		}
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{Sector, "Sector"},
		{Subsector, "Subsector"},
		{IndustryGroup, "Industry Group"},
		{NAICSIndustry, "NAICS Industry"},
		{NationalIndustry, "National Industry"},
		{Level(99), "Level(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level(%d).String() = %q, want %q", int(tt.level), got, tt.want)
			}
		})
	}
}

func TestIndustry_String(t *testing.T) {
	ind := &Industry{Code: "513210", Title: "Software Publishers"}
	want := "513210 - Software Publishers"
	if got := ind.String(); got != want {
		t.Errorf("Industry.String() = %q, want %q", got, want)
	}

	var nilInd *Industry
	if got := nilInd.String(); got != "<nil>" {
		t.Errorf("nil Industry.String() = %q, want %q", got, "<nil>")
	}
}

func TestCodeLevel(t *testing.T) {
	tests := []struct {
		code string
		want Level
	}{
		{"51", Sector},
		{"513", Subsector},
		{"5132", IndustryGroup},
		{"51321", NAICSIndustry},
		{"513210", NationalIndustry},
		{"31-33", Sector},
		{"44-45", Sector},
		{"48-49", Sector},
		{"X", Level(-1)},
		{"1234567", Level(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := codeLevel(tt.code); got != tt.want {
				t.Errorf("codeLevel(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}
