package crosswalk

import (
	"strings"
	"testing"
)

func mustSIC(t *testing.T) *SIC {
	t.Helper()
	s, err := NewSIC()
	if err != nil {
		t.Fatalf("NewSIC() error: %v", err)
	}
	return s
}

func TestNewSIC(t *testing.T) {
	s, err := NewSIC()
	if err != nil {
		t.Fatalf("NewSIC() error: %v", err)
	}
	if s.Len() == 0 {
		t.Fatal("SIC crosswalk is empty")
	}
	if s.Len() < 2000 {
		t.Errorf("expected at least 2000 mappings, got %d", s.Len())
	}
}

func TestNewSICFromCSV(t *testing.T) {
	csv := `naics_code,naics_title,sic_code,sic_title
111110,Soybean Farming,0116,Soybeans
111120,Oilseed Farming,0119,Cash Grains`

	s, err := NewSICFromCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("NewSICFromCSV() error: %v", err)
	}
	if s.Len() != 2 {
		t.Errorf("expected 2 mappings, got %d", s.Len())
	}
}

func TestSICToNAICS(t *testing.T) {
	s := mustSIC(t)

	mappings, ok := s.SICToNAICS("0116")
	if !ok {
		t.Fatal("expected mappings for SIC 0116")
	}
	found := false
	for _, m := range mappings {
		if m.NAICSCode == "111110" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected NAICS 111110 in SIC 0116 mappings")
	}
}

func TestSICToNAICS_NotFound(t *testing.T) {
	s := mustSIC(t)
	_, ok := s.SICToNAICS("9999")
	if ok {
		t.Error("expected no mappings for SIC 9999")
	}
}

func TestSICToNAICS_MultipleResults(t *testing.T) {
	s := mustSIC(t)

	mappings, ok := s.SICToNAICS("0119")
	if !ok {
		t.Fatal("expected mappings for SIC 0119")
	}
	if len(mappings) < 2 {
		t.Errorf("expected multiple NAICS codes for SIC 0119, got %d", len(mappings))
	}
}

func TestNAICSToSIC(t *testing.T) {
	s := mustSIC(t)

	mappings, ok := s.NAICSToSIC("111110")
	if !ok {
		t.Fatal("expected mappings for NAICS 111110")
	}
	found := false
	for _, m := range mappings {
		if m.SICCode == "0116" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SIC 0116 in NAICS 111110 mappings")
	}
}

func TestNAICSToSIC_NotFound(t *testing.T) {
	s := mustSIC(t)
	_, ok := s.NAICSToSIC("999999")
	if ok {
		t.Error("expected no mappings for NAICS 999999")
	}
}

func TestSIC_AllMappings(t *testing.T) {
	s := mustSIC(t)
	all := s.AllMappings()
	if len(all) != s.Len() {
		t.Errorf("AllMappings() returned %d, Len() = %d", len(all), s.Len())
	}
}
