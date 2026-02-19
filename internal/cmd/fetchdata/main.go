// Command fetchdata downloads and processes NAICS code data from official sources
// into the CSV format used by the embedded data in this library.
//
// Usage:
//
//	go run ./internal/cmd/fetchdata
//	go run ./internal/cmd/fetchdata -source /path/to/raw.csv -out internal/data/naics2022.csv
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

const defaultURL = "https://raw.githubusercontent.com/owings1/naics/main/2022.csv"

func main() {
	source := flag.String("source", "", "path to local raw NAICS CSV (downloads from GitHub if empty)")
	out := flag.String("out", "internal/data/naics2022.csv", "output CSV file path")
	flag.Parse()

	var r io.ReadCloser
	if *source != "" {
		f, err := os.Open(*source)
		if err != nil {
			log.Fatalf("open source file: %v", err)
		}
		r = f
	} else {
		fmt.Fprintf(os.Stderr, "Downloading from %s ...\n", defaultURL)
		resp, err := http.Get(defaultURL)
		if err != nil {
			log.Fatalf("download: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("download: status %s", resp.Status)
		}
		r = resp.Body
	}
	defer r.Close()

	records, err := parseRaw(r)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}

	if err := writeCSV(*out, records); err != nil {
		log.Fatalf("write: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Wrote %d records to %s\n", len(records), *out)
}

type record struct {
	code, title, description string
}

func parseRaw(r io.Reader) ([]record, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	_ = header

	var records []record
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		if len(row) < 3 {
			continue
		}

		code := strings.TrimSpace(row[1])
		title := strings.TrimSpace(row[2])
		desc := ""
		if len(row) >= 4 {
			desc = strings.TrimSpace(row[3])
		}

		if code == "" {
			continue
		}

		title = strings.TrimRightFunc(title, func(r rune) bool {
			return r == 'T' && !unicode.IsLower(r)
		})
		if strings.HasSuffix(title, "T") && !strings.HasSuffix(title, " T") {
			title = strings.TrimSuffix(title, "T")
		}

		desc = cleanDescription(desc)

		records = append(records, record{
			code:        code,
			title:       title,
			description: desc,
		})
	}
	return records, nil
}

func cleanDescription(s string) string {
	if s == "" || s == "NULL" {
		return ""
	}

	if strings.HasPrefix(s, "See industry description for") {
		return ""
	}

	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Cross-References") {
			break
		}
		if strings.HasPrefix(line, "Illustrative Examples:") {
			break
		}
		cleaned = append(cleaned, line)
	}
	result := strings.Join(cleaned, " ")

	if strings.HasPrefix(result, "The Sector as a Whole") {
		result = strings.TrimPrefix(result, "The Sector as a Whole")
		result = strings.TrimSpace(result)
	}

	if len(result) > 500 {
		cut := strings.LastIndex(result[:500], ".")
		if cut > 200 {
			result = result[:cut+1]
		} else {
			result = result[:500] + "..."
		}
	}

	return result
}

func writeCSV(path string, records []record) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"code", "title", "description"}); err != nil {
		return err
	}

	for _, rec := range records {
		if err := w.Write([]string{rec.code, rec.title, rec.description}); err != nil {
			return err
		}
	}

	return nil
}
