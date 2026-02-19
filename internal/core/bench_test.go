package core

import "testing"

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		reg, err := New()
		if err != nil {
			b.Fatal(err)
		}
		_ = reg
	}
}

func BenchmarkLookup(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Lookup("513210")
	}
}

func BenchmarkValid(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Valid("513210")
	}
}

func BenchmarkSearch_Short(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Search("software")
	}
}

func BenchmarkSearch_Long(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Search("computer systems design")
	}
}

func BenchmarkSearch_MaxResults(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Search("manufacturing", MaxResults(10))
	}
}

func BenchmarkParent(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Parent("513210")
	}
}

func BenchmarkChildren(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Children("51")
	}
}

func BenchmarkDescendants(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Descendants("51")
	}
}

func BenchmarkAncestors(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Ancestors("513210")
	}
}

func BenchmarkSectors(b *testing.B) {
	reg, err := New()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		reg.Sectors()
	}
}
