package main

import (
	"fmt"
	"sort"
	"strings"
)

// formatExtensionStats renders a top-N breakdown of file extension counts as
// a single comma-separated line. Extensions are sorted by count descending;
// counts beyond top-N are summed into an "others(N)" tail.
func formatExtensionStats(counts map[string]int, topN int) string {
	if len(counts) == 0 {
		return ""
	}

	type extCount struct {
		ext   string
		count int
	}
	sorted := make([]extCount, 0, len(counts))
	for ext, count := range counts {
		sorted = append(sorted, extCount{ext, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].count != sorted[j].count {
			return sorted[i].count > sorted[j].count
		}
		// Stable, deterministic order when counts tie.
		return sorted[i].ext < sorted[j].ext
	})

	var b strings.Builder
	others := 0
	for i, ec := range sorted {
		if i >= topN {
			others += ec.count
			continue
		}
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%s(%d)", ec.ext, ec.count)
	}
	if others > 0 {
		fmt.Fprintf(&b, ", others(%d)", others)
	}
	return b.String()
}
