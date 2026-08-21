package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

func TestWrapTextMeasuresDisplayWidthNotBytes(t *testing.T) {
	m := &Modal{}

	text := "α • β • γ • δ"
	if got := m.WrapText(text, runewidth.StringWidth(text)); len(got) != 1 {
		t.Errorf("text that fits by display width was wrapped: %q", got)
	}
}

func TestWrapTextTruncationStaysValidUTF8(t *testing.T) {
	m := &Modal{}

	for _, word := range []string{
		strings.Repeat("é", 40),
		strings.Repeat("→", 40),
		strings.Repeat("🌍", 40),
	} {
		for _, maxWidth := range []int{8, 9, 10, 11, 12} {
			for _, line := range m.WrapText(word, maxWidth) {
				if !utf8.ValidString(line) {
					t.Errorf("width %d produced invalid UTF-8 from %q", maxWidth, word[:8])
				}
				if runewidth.StringWidth(line) > maxWidth {
					t.Errorf("width %d exceeded: %q is %d columns", maxWidth, line, runewidth.StringWidth(line))
				}
			}
		}
	}
}

func TestWrapTextKeepsWordsWithinWidth(t *testing.T) {
	m := &Modal{}

	lines := m.WrapText("the quick brown fox jumps over the lazy dog", 12)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", lines)
	}
	for _, line := range lines {
		if runewidth.StringWidth(line) > 12 {
			t.Errorf("line %q is %d columns, want <= 12", line, runewidth.StringWidth(line))
		}
	}
}
