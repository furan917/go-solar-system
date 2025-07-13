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

func NewSystemModal(screen tcell.Screen, systemManager *systems.SystemManager) *SystemModal {
	systemInfo, _ := systemManager.ListSystemsWithInfo()

	config := ModalConfig{
		Width:    80,
		Height:   20,
		Title:    " 🌌 Star System Selection ",
		Content:  systemInfo,
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &SystemModal{
		Modal:         modal,
		systemManager: systemManager,
		systemInfo:    systemInfo,
		selectedIndex: 0,
		scrollIndex:   0,
	}
}

func (sm *SystemModal) Render() {
	sm.Modal.Render()

	sm.renderSystemList()

	sm.DrawInstructions("↑/↓ to navigate • Enter to select • Escape to cancel")
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

		sm.drawTextAt(sm.x+2, sm.y+3+i, style, line)
	}
}

func (sm *SystemModal) HandleNavigation(key tcell.Key) {
	switch key {
	case tcell.KeyUp:
		if sm.selectedIndex > 0 {
			sm.selectedIndex--
			if sm.selectedIndex < sm.scrollIndex {
				sm.scrollIndex = sm.selectedIndex
			}
		}
	case tcell.KeyDown:
		if sm.selectedIndex < len(sm.systemInfo)-1 {
			sm.selectedIndex++
			visibleLines := sm.height - 6
			if sm.selectedIndex >= sm.scrollIndex+visibleLines {
				sm.scrollIndex = sm.selectedIndex - visibleLines + 1
			}
		}
	default:
	}
}

func (sm *SystemModal) GetSelectedSystem() string {
	availableSystems := sm.systemManager.GetAvailableSystems()
	if sm.selectedIndex < len(availableSystems) {
		return availableSystems[sm.selectedIndex]
	}
	return ""
}
