package ui

import (
	"fmt"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/systems"
	"github.com/gdamore/tcell/v2"
)

type SystemModal struct {
	*Modal
	entries       []systems.SystemEntry
	selectedIndex int
	scrollIndex   int
	filter        string
	filterActive  bool
}

func NewSystemModal(
	screen tcell.Screen,
	entries []systems.SystemEntry,
	selectedIndex int,
	scrollIndex int,
	filter string,
	filterActive bool,
) *SystemModal {
	config := ModalConfig{
		Width:    constants.SystemModalWidth,
		Height:   constants.SystemModalHeight,
		Title:    " 🌌 Star System Selection ",
		Content:  []string{},
		Position: constants.TopRight,
	}

	return &SystemModal{
		Modal:         NewModal(screen, config),
		entries:       entries,
		selectedIndex: selectedIndex,
		scrollIndex:   scrollIndex,
		filter:        filter,
		filterActive:  filterActive,
	}
}

func (sm *SystemModal) Render() {
	sm.Modal.Render()

	sm.renderFilterLine()
	sm.renderSystemList()

	if sm.filterActive {
		sm.DrawInstructions("Type to filter • ↑/↓ to navigate • Enter to select • Escape to clear")
	} else {
		sm.DrawInstructions("↑/↓ to navigate • Enter to select • '/' to filter • Escape/'b' to cancel")
	}
}

func (sm *SystemModal) renderFilterLine() {
	style := sm.contentStyle

	var text string
	switch {
	case sm.filterActive:
		style = style.Foreground(tcell.ColorYellow).Bold(true)
		text = fmt.Sprintf("Filter: %s_  (%s)", sm.filter, matchCount(len(sm.entries)))
	case sm.filter != "":
		style = style.Foreground(tcell.ColorYellow)
		text = fmt.Sprintf("Filter: %s  (%s)", sm.filter, matchCount(len(sm.entries)))
	default:
		style = style.Foreground(tcell.ColorLightBlue)
		text = "Press '/' to filter"
	}

	sm.drawTextAt(sm.x+2, sm.y+2, style, text)
}

func matchCount(n int) string {
	if n == 1 {
		return "1 match"
	}
	return fmt.Sprintf("%d matches", n)
}

func (sm *SystemModal) renderSystemList() {
	if len(sm.entries) == 0 {
		sm.drawTextAt(sm.x+2, sm.y+3, sm.contentStyle.Foreground(tcell.ColorGray), "No systems match this filter")
		return
	}

	visibleLines := constants.SystemListVisibleItems

	for i := 0; i < visibleLines && i+sm.scrollIndex < len(sm.entries); i++ {
		lineIndex := i + sm.scrollIndex

		style := sm.contentStyle
		if lineIndex == sm.selectedIndex {
			style = style.Reverse(true).Bold(true)
		}

		wrappedLines := sm.WrapText(sm.entries[lineIndex].Info, sm.width-10)
		if len(wrappedLines) > 0 {
			sm.drawTextAt(sm.x+2, sm.y+3+i, style, wrappedLines[0])
		}
	}

	if sm.scrollIndex > 0 {
		sm.drawTextAt(sm.x+sm.width-2, sm.y+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
	}
	if sm.scrollIndex+visibleLines < len(sm.entries) {
		sm.drawTextAt(sm.x+sm.width-2, sm.y+sm.height-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
	}
}
