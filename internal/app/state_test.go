package app

import (
	"testing"

	"github.com/furan917/go-solar-system/internal/constants"
)

func TestClampSystemSelectionEmptyList(t *testing.T) {
	s := NewAppState()
	s.SystemSelectedIndex = 5
	s.SystemScrollIndex = 3

	s.ClampSystemSelection(0)

	if s.SystemSelectedIndex != 0 || s.SystemScrollIndex != 0 {
		t.Errorf("got selected=%d scroll=%d, want 0/0", s.SystemSelectedIndex, s.SystemScrollIndex)
	}
}

func TestClampSystemSelectionShrinkingList(t *testing.T) {
	s := NewAppState()
	s.SystemSelectedIndex = 9
	s.SystemScrollIndex = 5

	s.ClampSystemSelection(3)

	if s.SystemSelectedIndex != 2 {
		t.Errorf("selection should clamp to last entry, got %d", s.SystemSelectedIndex)
	}
	if s.SystemScrollIndex != 0 {
		t.Errorf("scroll should reset when the list fits, got %d", s.SystemScrollIndex)
	}
}

func TestClampSystemSelectionGrowingListKeepsSelectionVisible(t *testing.T) {
	s := NewAppState()
	count := constants.SystemListVisibleItems * 3
	s.SystemSelectedIndex = count - 1
	s.SystemScrollIndex = 0

	s.ClampSystemSelection(count)

	if s.SystemSelectedIndex < s.SystemScrollIndex ||
		s.SystemSelectedIndex >= s.SystemScrollIndex+constants.SystemListVisibleItems {
		t.Errorf("selection %d outside viewport starting at %d", s.SystemSelectedIndex, s.SystemScrollIndex)
	}
}

func TestClampSystemSelectionNegativeIndices(t *testing.T) {
	s := NewAppState()
	s.SystemSelectedIndex = -4
	s.SystemScrollIndex = -2

	s.ClampSystemSelection(5)

	if s.SystemSelectedIndex != 0 || s.SystemScrollIndex != 0 {
		t.Errorf("got selected=%d scroll=%d, want 0/0", s.SystemSelectedIndex, s.SystemScrollIndex)
	}
}

func TestHandleSystemNavigationStaysInBounds(t *testing.T) {
	s := NewAppState()
	count := 4

	for i := 0; i < 10; i++ {
		s.HandleSystemNavigation(1, count)
	}
	if s.SystemSelectedIndex != count-1 {
		t.Errorf("down navigation overran the list: %d", s.SystemSelectedIndex)
	}

	for i := 0; i < 10; i++ {
		s.HandleSystemNavigation(-1, count)
	}
	if s.SystemSelectedIndex != 0 {
		t.Errorf("up navigation underran the list: %d", s.SystemSelectedIndex)
	}
}

func TestHandleSystemNavigationScrollsViewport(t *testing.T) {
	s := NewAppState()
	count := constants.SystemListVisibleItems + 5

	for i := 0; i < count; i++ {
		s.HandleSystemNavigation(1, count)

		if s.SystemSelectedIndex < s.SystemScrollIndex ||
			s.SystemSelectedIndex >= s.SystemScrollIndex+constants.SystemListVisibleItems {
			t.Fatalf("step %d: selection %d outside viewport starting at %d", i, s.SystemSelectedIndex, s.SystemScrollIndex)
		}
	}
}

func TestClearSystemFilter(t *testing.T) {
	s := NewAppState()
	s.SystemFilter = "kep"
	s.SystemFilterActive = true

	s.ClearSystemFilter()

	if s.SystemFilter != "" || s.SystemFilterActive {
		t.Errorf("filter not cleared: %q active=%v", s.SystemFilter, s.SystemFilterActive)
	}
}

func TestTrimLastRuneHandlesMultibyte(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"a":    "",
		"kep":  "ke",
		"日本語":  "日本",
		"café": "caf",
	}

	for in, want := range cases {
		if got := trimLastRune(in); got != want {
			t.Errorf("trimLastRune(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResetModalsClearsSystemFilter(t *testing.T) {
	s := NewAppState()
	s.ShowingSystemList = true
	s.SystemFilter = "kep"
	s.SystemFilterActive = true

	s.ResetModals()

	if s.SystemFilter != "" || s.SystemFilterActive {
		t.Errorf("error recovery left a stale filter: %q active=%v", s.SystemFilter, s.SystemFilterActive)
	}
}
