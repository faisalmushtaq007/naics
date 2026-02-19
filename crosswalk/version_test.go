package crosswalk

import (
	"strings"
	"testing"
)

func mustVersion(t *testing.T) *Version {
	t.Helper()
	v, err := NewVersion()
	if err != nil {
		t.Fatalf("NewVersion() error: %v", err)
	}
	return v
}

func TestNewVersion(t *testing.T) {
	v, err := NewVersion()
	if err != nil {
		t.Fatalf("NewVersion() error: %v", err)
	}
	if v.Len() == 0 {
		t.Fatal("version crosswalk is empty")
	}
	if v.Len() < 1000 {
		t.Errorf("expected at least 1000 mappings, got %d", v.Len())
	}
}

func TestNewVersionFromCSV(t *testing.T) {
	csv := `naics_2017,title_2017,naics_2022,title_2022
511210,Software Publishers,513210,Software Publishers
511110,Newspaper Publishers,513110,Newspaper Publishers`

	v, err := NewVersionFromCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("NewVersionFromCSV() error: %v", err)
	}
	if v.Len() != 2 {
		t.Errorf("expected 2 mappings, got %d", v.Len())
	}
}

func TestMapTo2022(t *testing.T) {
	v := mustVersion(t)

	mappings, ok := v.MapTo2022("511210")
	if !ok {
		t.Fatal("expected mapping for 2017 code 511210")
	}
	found := false
	for _, m := range mappings {
		if m.Code2022 == "513210" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 2022 code 513210 in mappings for 2017 code 511210")
	}
}

func TestMapTo2022_UnchangedCode(t *testing.T) {
	v := mustVersion(t)

	mappings, ok := v.MapTo2022("111110")
	if !ok {
		t.Fatal("expected mapping for 2017 code 111110")
	}
	if len(mappings) != 1 {
		t.Errorf("expected 1 mapping for unchanged code, got %d", len(mappings))
	}
	if mappings[0].Code2022 != "111110" {
		t.Errorf("expected same code 111110, got %s", mappings[0].Code2022)
	}
}

func TestMapTo2022_NotFound(t *testing.T) {
	v := mustVersion(t)
	_, ok := v.MapTo2022("999999")
	if ok {
		t.Error("expected no mapping for invalid code")
	}
}

func TestMapTo2017(t *testing.T) {
	v := mustVersion(t)

	mappings, ok := v.MapTo2017("513210")
	if !ok {
		t.Fatal("expected mapping for 2022 code 513210")
	}
	found := false
	for _, m := range mappings {
		if m.Code2017 == "511210" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 2017 code 511210 in mappings for 2022 code 513210")
	}
}

func TestMapTo2017_NotFound(t *testing.T) {
	v := mustVersion(t)
	_, ok := v.MapTo2017("999999")
	if ok {
		t.Error("expected no mapping for invalid code")
	}
}

func TestVersion_AllMappings(t *testing.T) {
	v := mustVersion(t)
	all := v.AllMappings()
	if len(all) != v.Len() {
		t.Errorf("AllMappings() returned %d, Len() = %d", len(all), v.Len())
	}
}
