package ui

import (
	"fmt"
	"strings"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/display"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/visualization"
	"github.com/gdamore/tcell/v2"
)

type MoonListModal struct {
	*Modal
	planet        models.CelestialBody
	moonNames     []string
	selectedIndex int
	scrollIndex   int
	renderer      *visualization.Renderer
}

func NewMoonListModal(
	screen tcell.Screen,
	planet models.CelestialBody,
	moonNames []string,
	selectedIndex int,
	scrollIndex int,
	renderer *visualization.Renderer,
) *MoonListModal {
	config := ModalConfig{
		Width:    constants.ModalWidth,
		Height:   constants.ModalHeight,
		Title:    fmt.Sprintf(" %s Moons (%d total) ", planet.EnglishName, len(planet.Moons)),
		Content:  []string{},
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &MoonListModal{
		Modal:         modal,
		planet:        planet,
		moonNames:     moonNames,
		selectedIndex: selectedIndex,
		scrollIndex:   scrollIndex,
		renderer:      renderer,
	}
}

func (mlm *MoonListModal) Render() {
	mlm.Modal.Render()
	mlm.renderMoonList()
	mlm.renderScrollIndicators()
	mlm.DrawInstructions("↑/↓ to navigate • Enter to select • Escape/'b' to go back")
}

func (mlm *MoonListModal) renderMoonList() {
	visibleItems := constants.MaxVisibleItems
	startY := mlm.y + 3

	for i := 0; i < visibleItems && i+mlm.scrollIndex < len(mlm.moonNames); i++ {
		moonIndex := i + mlm.scrollIndex
		moonName := mlm.moonNames[moonIndex]

		style := mlm.contentStyle
		if moonIndex == mlm.selectedIndex {
			style = style.Reverse(true).Bold(true)
		}

		prefix := "  "
		if moonIndex == mlm.selectedIndex {
			prefix = "► "
		}

		text := fmt.Sprintf("%s%d. %s", prefix, moonIndex+1, moonName)
		mlm.drawTextAt(mlm.x+2, startY+i, style, text)
	}
}

func (mlm *MoonListModal) renderScrollIndicators() {
	visibleItems := constants.MaxVisibleItems
	startY := mlm.y + 3
	modalWidth := mlm.width

	scrollAreaStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)

	for i := 0; i < visibleItems; i++ {
		mlm.screen.SetContent(mlm.x+modalWidth-3, startY+i, '│', nil, scrollAreaStyle)
	}

	if len(mlm.moonNames) > visibleItems {
		totalScrollable := len(mlm.moonNames) - visibleItems
		scrollPosition := int(float64(mlm.scrollIndex) / float64(totalScrollable) * float64(visibleItems-1))
		mlm.screen.SetContent(mlm.x+modalWidth-3, startY+scrollPosition, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue))
	}

	if mlm.scrollIndex > 0 {
		mlm.drawTextAt(mlm.x+modalWidth-2, mlm.y+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
		mlm.drawTextAt(mlm.x+modalWidth-8, mlm.y+2, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "More")
	}
	if mlm.scrollIndex+visibleItems < len(mlm.moonNames) {
		mlm.drawTextAt(mlm.x+modalWidth-2, mlm.y+mlm.height-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
		mlm.drawTextAt(mlm.x+modalWidth-8, mlm.y+mlm.height-3, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "More")
	}

	statusStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)
	statusText := fmt.Sprintf("Showing %d-%d of %d moons",
		mlm.scrollIndex+1,
		Minimum(mlm.scrollIndex+visibleItems, len(mlm.moonNames)),
		len(mlm.moonNames))
	mlm.drawTextAt(mlm.x+2, mlm.y+mlm.height-3, statusStyle, statusText)
}

type MoonDetailsModal struct {
	*Modal
	moon   models.CelestialBody
	planet models.CelestialBody
}

func NewMoonDetailsModal(screen tcell.Screen, moon, planet models.CelestialBody) *MoonDetailsModal {
	content := generateMoonDetailsFromFields(moon, planet)

	config := ModalConfig{
		Width:    constants.ModalWidth,
		Height:   Minimum(len(content)+6, 25),
		Title:    fmt.Sprintf(" %s (Moon of %s) ", moon.EnglishName, planet.EnglishName),
		Content:  content,
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &MoonDetailsModal{
		Modal:  modal,
		moon:   moon,
		planet: planet,
	}
}

func (mdm *MoonDetailsModal) Render() {
	mdm.Modal.Render()

	if isAPIMoon(mdm.moon) {
		noteStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)
		wrappedLines := mdm.WrapText("Note: Limited moon data available from API", constants.ModalContentWidth)
		currentY := mdm.y + mdm.height - 3
		for _, line := range wrappedLines {
			mdm.drawTextAt(mdm.x+2, currentY, noteStyle, line)
			currentY++
		}
	}

	mdm.DrawInstructions("Press Enter, Escape, or 'b' to go back to moon list")
}

func generateMoonDetailsFromFields(moon models.CelestialBody, planet models.CelestialBody) []string {
	var content []string

	content = append(content, "Type: Moon")
	content = append(content, fmt.Sprintf("Orbits: %s", planet.EnglishName))

	if moon.Name != "" && moon.Name != moon.EnglishName {
		content = append(content, fmt.Sprintf("Original Name: %s", moon.Name))
	}

	stringFields := display.GetCelestialBodyStringFields()
	for _, field := range stringFields {
		if field.Condition(moon) && field.Label != "Type" {
			content = append(content, field.FormatStringFieldValue(moon))
		}
	}

	fields := display.GetCelestialBodyFields()
	for _, field := range fields {
		if field.Condition(moon) {
			content = append(content, field.FormatFieldValue(moon))
		}
	}

	return content
}

func isAPIMoon(moon models.CelestialBody) bool {
	return moon.MeanRadius > 0 || moon.Mass.MassValue > 0 || moon.Density > 0 ||
		(moon.ID != "" && !strings.Contains(moon.ID, "moon"))
}
