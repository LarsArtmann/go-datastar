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
	wantLen, gotLen := len(want), len(got)

	// lcs[i][j] = length of the LCS of want[i:] and got[j:].
	lcs := make([][]int, wantLen+1) //nolint:makezero // dense DP table, no appends
	for i := range lcs {
		lcs[i] = make([]int, gotLen+1) //nolint:makezero // dense DP table, no appends
	}

	for i := wantLen - 1; i >= 0; i-- {
		for j := gotLen - 1; j >= 0; j-- {
			if want[i] == got[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}

	var b strings.Builder

	wantIdx, gotIdx := 0, 0
	for wantIdx < wantLen && gotIdx < gotLen {
		switch {
		case want[wantIdx] == got[gotIdx]:
			b.WriteString("  " + want[wantIdx] + "\n")
			wantIdx++
			gotIdx++
		case lcs[wantIdx+1][gotIdx] >= lcs[wantIdx][gotIdx+1]:
			b.WriteString("- " + want[wantIdx] + "\n")
			wantIdx++
		default:
			b.WriteString("+ " + got[gotIdx] + "\n")
			gotIdx++
		}
	}

	for ; wantIdx < wantLen; wantIdx++ {
		b.WriteString("- " + want[wantIdx] + "\n")
	}

	for ; gotIdx < gotLen; gotIdx++ {
		b.WriteString("+ " + got[gotIdx] + "\n")
	}

	return b.String()
}
