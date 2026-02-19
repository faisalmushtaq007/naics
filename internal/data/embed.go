// Package data provides embedded NAICS code data files.
// This is an internal package; external consumers should use the top-level naics package.
package data

import "embed"

// CSV contains the embedded NAICS 2022 code data in CSV format.
// The CSV has columns: code, title, description.
//
//go:embed naics2022.csv
var CSV embed.FS

// SICCrosswalk contains the SIC-to-NAICS crosswalk mapping in CSV format.
// The CSV has columns: naics_code, naics_title, sic_code, sic_title.
//
//go:embed sic_naics.csv
var SICCrosswalk embed.FS

// NAICSVersionCrosswalk contains the NAICS 2017-to-2022 concordance in CSV format.
// The CSV has columns: naics_2017, title_2017, naics_2022, title_2022.
//
//go:embed naics_2017_to_2022.csv
var NAICSVersionCrosswalk embed.FS
