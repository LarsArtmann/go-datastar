package datastartest

import (
	"strings"
)

// Diff returns a readable line-based diff between the want and got event
// streams: lines only in want are prefixed "-", lines only in got "+", common
// lines unchanged. Each event renders as its header line followed by its
// decoded datalines, so a mismatch in one selector or one element is visible
// without reading the whole stream. Empty string means the streams are equal.
func Diff(want, got []Event) string {
	wantLines := renderEvents(want)
	gotLines := renderEvents(got)

	if strings.Join(wantLines, "\n") == strings.Join(gotLines, "\n") {
		return ""
	}

	return strings.TrimRight(diffLines(wantLines, gotLines), "\n")
}

// renderEvents renders events as one header line per event followed by one
// indented line per raw dataline — the shape Diff and Snapshot compare.
func renderEvents(events []Event) []string {
	if len(events) == 0 {
		return []string{"(no events)"}
	}

	lines := make([]string, 0, len(events)*2)

	for _, evt := range events {
		lines = append(lines, evt.String())

		for _, dataline := range evt.DataLines {
			lines = append(lines, "  "+dataline)
		}
	}

	return lines
}

// diffLines returns a longest-common-subsequence line diff of the two line
// slices. Streams are small (test outputs), so the O(n*m) table is fine.
func diffLines(want, got []string) string {
	n, m := len(want), len(got)

	// lcs[i][j] = length of the LCS of want[i:] and got[j:].
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if want[i] == got[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}

	var b strings.Builder

	i, j := 0, 0
	for i < n && j < m {
		switch {
		case want[i] == got[j]:
			b.WriteString("  " + want[i] + "\n")
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			b.WriteString("- " + want[i] + "\n")
			i++
		default:
			b.WriteString("+ " + got[j] + "\n")
			j++
		}
	}

	for ; i < n; i++ {
		b.WriteString("- " + want[i] + "\n")
	}

	for ; j < m; j++ {
		b.WriteString("+ " + got[j] + "\n")
	}

	return b.String()
}
