package core

import "strings"

// Parent returns the parent industry of the given code, moving one level
// up in the NAICS hierarchy. For example, the parent of "511210" (6-digit)
// is "51121" (5-digit).
//
// For range-based sectors (31-33, 44-45, 48-49), subsector codes like "311"
// correctly resolve to the range sector "31-33".
//
// Returns false if code is not found, is a top-level sector, or has no parent.
func (reg *Registry) Parent(code string) (*Industry, bool) {
	ind, ok := reg.codes[code]
	if !ok {
		return nil, false
	}
	if ind.Level == Sector {
		return nil, false
	}

	parentCode := deriveParentCode(code)
	if parentCode == "" {
		return nil, false
	}

	parent, ok := reg.codes[parentCode]
	return parent, ok
}

// Children returns the direct children of the given code -- industries
// exactly one level deeper in the hierarchy. For example, children of
// sector "51" are all 3-digit subsectors starting with "51".
//
// Returns nil if code is not found or has no children.
func (reg *Registry) Children(code string) []*Industry {
	ind, ok := reg.codes[code]
	if !ok {
		return nil
	}

	childLevel := ind.Level + 1
	if childLevel > NationalIndustry {
		return nil
	}

	prefixes := childPrefixes(code)
	var children []*Industry
	for _, candidate := range reg.list {
		if candidate.Level != childLevel {
			continue
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(candidate.Code, prefix) {
				children = append(children, candidate)
				break
			}
		}
	}
	return children
}

// Descendants returns all industries below the given code in the hierarchy
// (excluding the code itself). Results are sorted by code.
//
// Returns nil if code is not found or has no descendants.
func (reg *Registry) Descendants(code string) []*Industry {
	if _, ok := reg.codes[code]; !ok {
		return nil
	}

	prefixes := childPrefixes(code)
	var descendants []*Industry
	for _, candidate := range reg.list {
		if candidate.Code == code {
			continue
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(candidate.Code, prefix) {
				descendants = append(descendants, candidate)
				break
			}
		}
	}
	return descendants
}

// Ancestors returns the chain of parent industries from the given code up to
// the sector level. The returned slice is ordered from immediate parent to
// the top-level sector. For example, Ancestors("511210") returns:
// [51121, 5112, 511, 51].
//
// Returns nil if code is not found or is already a sector.
func (reg *Registry) Ancestors(code string) []*Industry {
	if _, ok := reg.codes[code]; !ok {
		return nil
	}

	var chain []*Industry
	current := code
	for {
		parent, ok := reg.Parent(current)
		if !ok {
			break
		}
		chain = append(chain, parent)
		current = parent.Code
	}
	return chain
}

// Siblings returns all industries at the same level sharing the same parent
// as the given code, excluding the code itself.
//
// Returns nil if code is not found or has no siblings.
func (reg *Registry) Siblings(code string) []*Industry {
	parent, ok := reg.Parent(code)
	if !ok {
		return nil
	}

	children := reg.Children(parent.Code)
	var siblings []*Industry
	for _, child := range children {
		if child.Code != code {
			siblings = append(siblings, child)
		}
	}
	return siblings
}

// Sectors returns all top-level sector industries (Level == Sector).
func (reg *Registry) Sectors() []*Industry {
	var sectors []*Industry
	for _, ind := range reg.list {
		if ind.Level == Sector {
			sectors = append(sectors, ind)
		}
	}
	return sectors
}

// deriveParentCode computes the parent code by removing the last digit.
// For 3-digit codes under range sectors (e.g., "311"), it returns the
// range form (e.g., "31-33").
func deriveParentCode(code string) string {
	lvl := codeLevel(code)
	if lvl <= Sector {
		return ""
	}

	if lvl == Subsector {
		return sectorCode(code[:2])
	}

	return code[:len(code)-1]
}

// childPrefixes returns the code prefixes that children of the given code
// would start with. For most codes this is just the code itself.
// For range-based sectors, it returns all prefixes in the range.
func childPrefixes(code string) []string {
	switch code {
	case "31-33":
		return []string{"31", "32", "33"}
	case "44-45":
		return []string{"44", "45"}
	case "48-49":
		return []string{"48", "49"}
	default:
		return []string{code}
	}
}
