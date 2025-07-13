package ui

import (
	"fmt"
	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/gdamore/tcell/v2"
)

type MoonListModal struct {
	*Modal
	planet        models.CelestialBody
	moonNames     []string
	selectedIndex int
	scrollIndex   int
}

func NewMoonListModal(screen tcell.Screen, planet models.CelestialBody, moonNames []string) *MoonListModal {
	config := ModalConfig{
		Width:    60,
		Height:   20,
		Title:    fmt.Sprintf(" %s Moons (%d total) ", planet.EnglishName, len(planet.Moons)),
		Content:  []string{},
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &MoonListModal{
		Modal:         modal,
		planet:        planet,
		moonNames:     moonNames,
		selectedIndex: 0,
		scrollIndex:   0,
	}
}

func (mlm *MoonListModal) Render() {
	mlm.Modal.Render()
	mlm.renderMoonList()
	mlm.DrawInstructions("↑/↓ to navigate • Enter to select • Escape/'b' to go back")
}

func (mlm *MoonListModal) renderMoonList() {
	visibleLines := mlm.height - 6

	for i := 0; i < visibleLines && i+mlm.scrollIndex < len(mlm.moonNames); i++ {
		lineIndex := i + mlm.scrollIndex
		moonName := mlm.moonNames[lineIndex]

		style := mlm.contentStyle
		if lineIndex == mlm.selectedIndex {
			style = style.Reverse(true).Bold(true)
		}

		text := fmt.Sprintf("%d. %s", lineIndex+1, moonName)
		mlm.drawTextAt(mlm.x+2, mlm.y+3+i, style, text)
	}
}

func (mlm *MoonListModal) HandleNavigation(key tcell.Key) {
	switch key {
	case tcell.KeyUp:
		if mlm.selectedIndex > 0 {
			mlm.selectedIndex--
			if mlm.selectedIndex < mlm.scrollIndex {
				mlm.scrollIndex = mlm.selectedIndex
			}
		}
	case tcell.KeyDown:
		if mlm.selectedIndex < len(mlm.moonNames)-1 {
			mlm.selectedIndex++
			visibleLines := mlm.height - 6
			if mlm.selectedIndex >= mlm.scrollIndex+visibleLines {
				mlm.scrollIndex = mlm.selectedIndex - visibleLines + 1
			}
		}
	default:
	}
}

func (mlm *MoonListModal) GetSelectedMoon() models.CelestialBody {
	if mlm.selectedIndex < len(mlm.planet.Moons) {
		moon := mlm.planet.Moons[mlm.selectedIndex]
		return models.CelestialBody{
			ID:          moon.ID,
			Name:        moon.Name,
			EnglishName: moon.EnglishName,
			BodyType:    "Moon",
		}
	}
	return models.CelestialBody{}
}

type MoonDetailsModal struct {
	*Modal
	moon   models.CelestialBody
	planet models.CelestialBody
}

func NewMoonDetailsModal(screen tcell.Screen, moon, planet models.CelestialBody) *MoonDetailsModal {
	content := generateMoonDetails(moon)

	config := ModalConfig{
		Width:    60,
		Height:   minimum(len(content)+6, 25),
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
	mdm.DrawInstructions("Escape/'b' to go back")
}

func generateMoonDetails(moon models.CelestialBody) []string {
	var content []string

	content = append(content, "Type: Moon")

	if moon.MeanRadius > 0 {
		content = append(content, fmt.Sprintf("Radius: %.0f km", moon.MeanRadius))
	}

	if moon.Mass.MassValue > 0 {
		content = append(content, fmt.Sprintf("Mass: %.2f × 10^%d kg", moon.Mass.MassValue, moon.Mass.MassExponent))
	}

	if moon.Density > 0 {
		content = append(content, fmt.Sprintf("Density: %.2f g/cm³", moon.Density))
	}

	if moon.Gravity > 0 {
		content = append(content, fmt.Sprintf("Gravity: %.1f m/s²", moon.Gravity))
	}

	if moon.SemimajorAxis > 0 {
		content = append(content, fmt.Sprintf("Distance from Planet: %.0f km", moon.SemimajorAxis))
	}

	if moon.SideralOrbit > 0 {
		content = append(content, fmt.Sprintf("Orbital Period: %.2f days", moon.SideralOrbit))
	}

	if moon.DiscoveredBy != "" {
		content = append(content, fmt.Sprintf("Discovered by: %s", moon.DiscoveredBy))
	}

	if moon.DiscoveryDate != "" {
		content = append(content, fmt.Sprintf("Discovery Date: %s", moon.DiscoveryDate))
	}

	return content
}
