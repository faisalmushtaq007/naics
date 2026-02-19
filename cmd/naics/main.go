// Command naics provides a CLI for looking up, searching, and navigating
// NAICS industry codes.
//
// Usage:
//
//	naics lookup <code>                        Look up an industry by code
//	naics search [-n limit] [-p page] <query>  Search industries by keyword
//	naics count <query>                        Count matching industries
//	naics validate <code> [codes...]           Validate one or more NAICS codes
//	naics tree <code>                Display hierarchy tree for a code
//	naics parent <code>              Show ancestor chain to sector
//	naics sectors                    List all sectors
//	naics sic <sic_code>             Look up NAICS codes for a SIC code
//	naics version <2017_code>        Map a 2017 NAICS code to 2022
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/faisalmushtaq007/naics"
	"github.com/faisalmushtaq007/naics/crosswalk"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	reg, err := naics.New()
	if err != nil {
		fatal("load NAICS data: %v", err)
	}

	switch cmd {
	case "lookup":
		cmdLookup(reg, args)
	case "search":
		cmdSearch(reg, args)
	case "count":
		cmdCount(reg, args)
	case "validate":
		cmdValidate(reg, args)
	case "tree":
		cmdTree(reg, args)
	case "parent":
		cmdParent(reg, args)
	case "sectors":
		cmdSectors(reg)
	case "sic":
		cmdSIC(args)
	case "version":
		cmdVersion(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func cmdLookup(reg *naics.Registry, args []string) {
	if len(args) < 1 {
		fatal("usage: naics lookup <code>")
	}
	code := args[0]
	ind, ok := reg.Lookup(code)
	if !ok {
		fatal("code %q not found", code)
	}

	fmt.Printf("Code:        %s\n", ind.Code)
	fmt.Printf("Title:       %s\n", ind.Title)
	fmt.Printf("Level:       %s\n", ind.Level)
	if ind.Description != "" {
		fmt.Printf("Description: %s\n", ind.Description)
	}
}

func cmdSearch(reg *naics.Registry, args []string) {
	if len(args) < 1 {
		fatal("usage: naics search [-n limit] [-p page] <query>")
	}

	limit := 20
	page := 1
	var queryParts []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n":
			i++
			if i >= len(args) {
				fatal("-n requires a number")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				fatal("-n must be a positive integer")
			}
			limit = n
		case "-p":
			i++
			if i >= len(args) {
				fatal("-p requires a number")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				fatal("-p must be a positive integer")
			}
			page = n
		default:
			queryParts = append(queryParts, args[i])
		}
	}

	if len(queryParts) == 0 {
		fatal("usage: naics search [-n limit] [-p page] <query>")
	}

	query := strings.Join(queryParts, " ")
	offset := (page - 1) * limit

	resp := reg.SearchPage(query, naics.MaxResults(limit), naics.Offset(offset))
	if resp.Total == 0 {
		fmt.Println("No results found.")
		return
	}

	if len(resp.Results) == 0 {
		fmt.Printf("No results on page %d (total: %d).\n", page, resp.Total)
		return
	}

	start := offset + 1
	end := offset + len(resp.Results)
	fmt.Printf("Results %d-%d of %d for %q:\n\n", start, end, resp.Total, query)

	for _, r := range resp.Results {
		fmt.Printf("  %-8s %s (%s)\n", r.Industry.Code, r.Industry.Title, r.Industry.Level)
	}

	if resp.HasMore() {
		fmt.Printf("\n  Page %d of %d. Next: naics search -n %d -p %d %s\n",
			page, (resp.Total+limit-1)/limit, limit, page+1, query)
	}
}

func cmdCount(reg *naics.Registry, args []string) {
	if len(args) < 1 {
		fatal("usage: naics count <query>")
	}
	query := strings.Join(args, " ")
	count := reg.Count(query)
	fmt.Printf("%d results for %q\n", count, query)
}

func cmdValidate(reg *naics.Registry, args []string) {
	if len(args) < 1 {
		fatal("usage: naics validate <code> [codes...]")
	}
	allValid := true
	for _, code := range args {
		if reg.Valid(code) {
			fmt.Printf("  %s  valid\n", code)
		} else {
			fmt.Printf("  %s  INVALID\n", code)
			allValid = false
		}
	}
	if !allValid {
		os.Exit(1)
	}
}

func cmdTree(reg *naics.Registry, args []string) {
	if len(args) < 1 {
		fatal("usage: naics tree <code>")
	}
	code := args[0]
	ind, ok := reg.Lookup(code)
	if !ok {
		fatal("code %q not found", code)
	}

	printTree(reg, ind, 0)
}

func printTree(reg *naics.Registry, ind *naics.Industry, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s%s %s\n", indent, ind.Code, ind.Title)
	for _, child := range reg.Children(ind.Code) {
		printTree(reg, child, depth+1)
	}
}

func cmdParent(reg *naics.Registry, args []string) {
	if len(args) < 1 {
		fatal("usage: naics parent <code>")
	}
	code := args[0]
	if _, ok := reg.Lookup(code); !ok {
		fatal("code %q not found", code)
	}

	ancestors := reg.Ancestors(code)
	if len(ancestors) == 0 {
		fmt.Println("No parent (top-level sector).")
		return
	}

	fmt.Printf("Ancestor chain for %s:\n", code)
	for _, a := range ancestors {
		fmt.Printf("  %s %s (%s)\n", a.Code, a.Title, a.Level)
	}
}

func cmdSectors(reg *naics.Registry) {
	sectors := reg.Sectors()
	fmt.Printf("NAICS 2022 Sectors (%d):\n\n", len(sectors))
	for _, s := range sectors {
		fmt.Printf("  %-6s %s\n", s.Code, s.Title)
	}
}

func cmdSIC(args []string) {
	if len(args) < 1 {
		fatal("usage: naics sic <sic_code>")
	}
	sicCode := args[0]

	sic, err := crosswalk.NewSIC()
	if err != nil {
		fatal("load SIC crosswalk: %v", err)
	}

	mappings, ok := sic.SICToNAICS(sicCode)
	if !ok {
		fatal("SIC code %q not found in crosswalk", sicCode)
	}

	fmt.Printf("SIC %s -> NAICS mappings:\n\n", sicCode)
	for _, m := range mappings {
		fmt.Printf("  %-8s %s\n", m.NAICSCode, m.NAICSTitle)
	}
}

func cmdVersion(args []string) {
	if len(args) < 1 {
		fatal("usage: naics version <2017_code>")
	}
	code := args[0]

	ver, err := crosswalk.NewVersion()
	if err != nil {
		fatal("load version crosswalk: %v", err)
	}

	mappings, ok := ver.MapTo2022(code)
	if !ok {
		fatal("2017 NAICS code %q not found in concordance", code)
	}

	fmt.Printf("NAICS 2017 %s -> 2022 mappings:\n\n", code)
	for _, m := range mappings {
		if m.Code2017 == m.Code2022 {
			fmt.Printf("  %-8s %s (unchanged)\n", m.Code2022, m.Title2022)
		} else {
			fmt.Printf("  %-8s %s\n", m.Code2022, m.Title2022)
		}
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `naics - NAICS industry code lookup tool

Usage:
  naics <command> [arguments]

Commands:
  lookup <code>                        Look up an industry by NAICS code
  search [-n limit] [-p page] <query>  Search with pagination (default: 20/page)
  count <query>                        Count matching industries
  validate <code> [codes...]           Validate one or more NAICS codes
  tree <code>                Display hierarchy tree for a code
  parent <code>              Show ancestor chain to sector
  sectors                    List all 20 NAICS sectors
  sic <sic_code>             Look up NAICS codes for a SIC code
  version <2017_code>        Map a 2017 NAICS code to 2022

Examples:
  naics lookup 513210
  naics search "software"
  naics search -n 10 -p 2 "farming"
  naics count "software"
  naics validate 513210 999999
  naics tree 51
  naics parent 513210
  naics sic 7372
  naics version 511210`)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "naics: "+format+"\n", args...)
	os.Exit(1)
}
