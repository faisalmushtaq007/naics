package naics_test

import (
	"fmt"
	"log"

	"github.com/faisalmushtaq007/naics"
)

func ExampleNew() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Loaded %d NAICS codes\n", reg.Len())
}

func ExampleRegistry_Lookup() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	ind, ok := reg.Lookup("513210")
	if !ok {
		fmt.Println("not found")
		return
	}
	fmt.Printf("Code:  %s\n", ind.Code)
	fmt.Printf("Title: %s\n", ind.Title)
	fmt.Printf("Level: %s\n", ind.Level)
	// Output:
	// Code:  513210
	// Title: Software Publishers
	// Level: National Industry
}

func ExampleRegistry_Valid() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reg.Valid("513210"))
	fmt.Println(reg.Valid("999999"))
	// Output:
	// true
	// false
}

func ExampleRegistry_Search() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	results := reg.Search("software", naics.MaxResults(3))
	for _, r := range results {
		fmt.Printf("%-6s %s (score=%.0f)\n", r.Industry.Code, r.Industry.Title, r.Score)
	}
}

func ExampleRegistry_Parent() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	parent, ok := reg.Parent("513210")
	if ok {
		fmt.Printf("Parent of 513210: %s\n", parent)
	}
	// Output:
	// Parent of 513210: 51321 - Software Publishers
}

func ExampleRegistry_Children() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	children := reg.Children("51")
	fmt.Printf("Sector 51 has %d subsectors\n", len(children))
	if len(children) > 0 {
		fmt.Printf("First: %s\n", children[0])
	}
}

func ExampleRegistry_Ancestors() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	ancestors := reg.Ancestors("513210")
	fmt.Println("Ancestor chain for 513210:")
	for _, a := range ancestors {
		fmt.Printf("  %s (%s)\n", a, a.Level)
	}
	// Output:
	// Ancestor chain for 513210:
	//   51321 - Software Publishers (NAICS Industry)
	//   5132 - Software Publishers (Industry Group)
	//   513 - Publishing Industries (Subsector)
	//   51 - Information (Sector)
}

func ExampleRegistry_Sectors() {
	reg, err := naics.New()
	if err != nil {
		log.Fatal(err)
	}

	sectors := reg.Sectors()
	fmt.Printf("Total sectors: %d\n", len(sectors))
}
