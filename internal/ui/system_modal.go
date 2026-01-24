package ui

import (
	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/systems"
	"github.com/gdamore/tcell/v2"
)

type SystemModal struct {
	*Modal
	systemManager *systems.SystemManager
	selectedIndex int
	scrollIndex   int
	systemInfo    []string
}

func NewSystemModal(
	screen tcell.Screen,
	systemManager *systems.SystemManager,
	selectedIndex int,
	scrollIndex int,
) *SystemModal {
	systemInfo, _ := systemManager.ListSystemsWithInfo()

	config := ModalConfig{
		Width:    80,
		Height:   20,
		Title:    " 🌌 Star System Selection ",
		Content:  []string{},
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &SystemModal{
		Modal:         modal,
		systemManager: systemManager,
		systemInfo:    systemInfo,
		selectedIndex: selectedIndex,
		scrollIndex:   scrollIndex,
	}
}

func (sm *SystemModal) Render() {
	sm.Modal.Render()

	sm.renderSystemList()

	sm.DrawInstructions("↑/↓ to navigate • Enter to select • Escape/'b' to cancel")
}

func (sm *SystemModal) renderSystemList() {
	visibleLines := sm.height - 6

	for i := 0; i < visibleLines && i+sm.scrollIndex < len(sm.systemInfo); i++ {
		lineIndex := i + sm.scrollIndex
		line := sm.systemInfo[lineIndex]

		style := sm.contentStyle
		if lineIndex == sm.selectedIndex {
			style = style.Reverse(true).Bold(true)
		}

		maxLineLength := sm.width - 10
		wrappedLines := sm.WrapText(line, maxLineLength)

		if len(wrappedLines) > 0 {
			sm.drawTextAt(sm.x+2, sm.y+3+i, style, wrappedLines[0])
		}
	}

	if sm.scrollIndex > 0 {
		sm.drawTextAt(sm.x+sm.width-2, sm.y+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
	}
	if sm.scrollIndex+visibleLines < len(sm.systemInfo) {
		sm.drawTextAt(sm.x+sm.width-2, sm.y+sm.height-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
	}
}
