package app

import (
	"fmt"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/display"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/systems"
	"github.com/furan917/go-solar-system/internal/ui"
	"github.com/furan917/go-solar-system/internal/visualization"
	"github.com/gdamore/tcell/v2"
)

// UIRenderer handles all UI rendering concerns for the solar system application
type UIRenderer struct {
	screen        tcell.Screen
	renderer      *visualization.Renderer
	systemManager *systems.SystemManager
	state         *AppState
}

// NewUIRenderer creates a new UI renderer with necessary dependencies
func NewUIRenderer(
	screen tcell.Screen,
	renderer *visualization.Renderer,
	systemManager *systems.SystemManager,
	state *AppState,
) *UIRenderer {
	return &UIRenderer{
		screen:        screen,
		renderer:      renderer,
		systemManager: systemManager,
		state:         state,
	}
}

// DrawScreen renders the complete UI
func (ur *UIRenderer) DrawScreen() {
	ur.screen.Clear()

	width, height := ur.screen.Size()

	ur.drawText(2, 1, tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true), "🌌 Solar System Explorer")

	modalWidth := constants.ModalWidth
	availableWidth := width - modalWidth - (constants.ModalMargin * 3)
	ur.drawPlanetList(2, 3, availableWidth)

	ur.drawSolarSystem(2, 6, width-4, height-8)

	instructions := "Arrow keys to navigate • Enter/Click to select • S for systems • Q to quit • 1-9 for direct selection"
	systemDisplayName := ur.systemManager.GetCurrentSystemDisplayName()

	instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorLightBlue)
	systemStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	ur.drawText(2, height-2, instructionStyle, instructions)
	ur.drawText(2+len(instructions)+3, height-2, systemStyle, fmt.Sprintf("• Current System: %s", systemDisplayName))

	// Draw modals based on current state
	if ur.state.IsShowingMoonDetails() {
		modal := ur.createMoonDetailsModal()
		modal.Render()
	} else if ur.state.IsShowingMoons() {
		modal := ur.createMoonListModal()
		modal.Render()
	} else if ur.state.IsShowingSystemList() {
		modal := ur.createSystemModal()
		modal.Render()
	} else if ur.state.IsShowingDetails() {
		modal := ur.createPlanetModal()
		modal.Render()
	}

	ur.screen.Show()
}

// drawText renders text at the specified position with given style
func (ur *UIRenderer) drawText(x, y int, style tcell.Style, text string) {
	for i, r := range text {
		ur.screen.SetContent(x+i, y, r, nil, style)
	}
}

// drawPlanetList renders the horizontal list of planets
func (ur *UIRenderer) drawPlanetList(x, y, maxWidth int) {
	currentX := x
	currentY := y

	ur.state.ClearPlanetListPositions()

	for i, planet := range ur.state.GetPlanets() {
		symbol := ur.renderer.GetPlanetSymbol(planet.EnglishName)
		name := planet.EnglishName

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if i == ur.state.SelectedIndex {
			style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true).Reverse(true)
		}

		planetText := fmt.Sprintf(" %c %s ", symbol, name)
		textWidth := len(planetText)

		if currentX+textWidth > x+maxWidth {
			currentY++
			currentX = x
		}

		ur.drawText(currentX, currentY, style, planetText)

		ur.state.AddPlanetListPosition(PlanetListPosition{
			Index: i,
			X:     currentX,
			Y:     currentY,
			Width: textWidth,
		})

		currentX += textWidth
	}
}

// drawSolarSystem renders the orbital visualization
func (ur *UIRenderer) drawSolarSystem(x, y, width, height int) {
	screenWidth, screenHeight := ur.screen.Size()
	grid, planetPositions := ur.renderer.RenderSolarSystemDataWithPositions(ur.state.GetPlanets(), width, height, screenWidth, screenHeight)
	ur.state.UpdatePlanetPositions(x, y, planetPositions)

	for row := 0; row < len(grid) && row < height; row++ {
		for col := 0; col < len(grid[row]) && col < width; col++ {
			if grid[row][col] != ' ' {
				style := ur.getPlanetStyle(grid[row][col])
				ur.screen.SetContent(x+col, y+row, grid[row][col], nil, style)
			}
		}
	}
}

// getPlanetStyle returns the appropriate style for a planet symbol
func (ur *UIRenderer) getPlanetStyle(symbol rune) tcell.Style {
	switch symbol {
	case '☉': // Sun
		return tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	case '☿': // Mercury
		return tcell.StyleDefault.Foreground(tcell.ColorGray)
	case '♀': // Venus
		return tcell.StyleDefault.Foreground(tcell.ColorOrange)
	case '♁': // Earth
		return tcell.StyleDefault.Foreground(tcell.ColorBlue)
	case '♂': // Mars
		return tcell.StyleDefault.Foreground(tcell.ColorRed)
	case '♃': // Jupiter
		return tcell.StyleDefault.Foreground(tcell.ColorBrown)
	case '♄': // Saturn
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	case '♅': // Uranus
		return tcell.StyleDefault.Foreground(tcell.ColorAqua)
	case '♆': // Neptune
		return tcell.StyleDefault.Foreground(tcell.ColorBlue)
	case '♇': // Pluto
		return tcell.StyleDefault.Foreground(tcell.ColorGray)
	case '.': // Asteroids/debris
		return tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
	case '·': // Kuiper belt
		return tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
	default:
		return tcell.StyleDefault.Foreground(tcell.ColorWhite)
	}
}

// UpdateDimensions handles screen resize events
func (ur *UIRenderer) UpdateDimensions(width, height int) {
	ur.renderer.UpdateDimensions(width, height)
}

// GetRenderer returns the visualization renderer
func (ur *UIRenderer) GetRenderer() *visualization.Renderer {
	return ur.renderer
}

// GetSystemManager returns the system manager
func (ur *UIRenderer) GetSystemManager() *systems.SystemManager {
	return ur.systemManager
}

func (ur *UIRenderer) createPlanetModal() *ui.PlanetModal {
	return ui.NewPlanetModal(
		ur.screen,
		ur.state.SelectedPlanet,
		ur.renderer,
	)
}

func (ur *UIRenderer) createMoonListModal() *ui.MoonListModal {
	moonHandler := ur.renderer.GetMoonHandler()
	moonNames := moonHandler.GetMoonNames(ur.state.SelectedPlanet)

	if len(moonNames) == 0 {
		for i := 0; i < len(ur.state.SelectedPlanet.Moons); i++ {
			moonNames = append(moonNames, fmt.Sprintf("Moon %d", i+1))
		}
	}

	return ui.NewMoonListModal(
		ur.screen,
		ur.state.SelectedPlanet,
		moonNames,
		ur.state.MoonSelectedIndex,
		ur.state.MoonScrollIndex,
		ur.renderer,
	)
}

func (ur *UIRenderer) createMoonDetailsModal() *ui.MoonDetailsModal {
	return ui.NewMoonDetailsModal(
		ur.screen,
		ur.state.SelectedMoon,
		ur.state.SelectedPlanet,
	)
}

func (ur *UIRenderer) createSystemModal() *ui.SystemModal {
	return ui.NewSystemModal(
		ur.screen,
		ur.systemManager,
		ur.state.SystemSelectedIndex,
		ur.state.SystemScrollIndex,
	)
}

// Supporting methods for modal click detection

// calculatePlanetDetailsLines calculates how many lines are needed for planet details
func (ur *UIRenderer) calculatePlanetDetailsLines(planet models.CelestialBody) int {
	lines := 0

	fields := display.GetCelestialBodyFields()
	for _, field := range fields {
		if field.Condition(planet) {
			lines++
		}
	}

	stringFields := display.GetCelestialBodyStringFields()
	for _, field := range stringFields {
		if field.Condition(planet) {
			lines++
		}
	}

	// Count moon lines
	if len(planet.Moons) > 0 {
		moonHandler := ur.renderer.GetMoonHandler()
		moonLines := moonHandler.FormatMoonDisplay(planet, 5)
		lines += len(moonLines) + 1 // +1 for spacing
	}

	return lines
}

// calculateMoonDetailsLines calculates how many lines are needed for moon details
func (ur *UIRenderer) calculateMoonDetailsLines(moon models.CelestialBody) int {
	lines := 1 // Type line (base)

	fields := []func(models.CelestialBody) bool{
		func(cb models.CelestialBody) bool { return cb.MeanRadius > 0 },
		func(cb models.CelestialBody) bool { return cb.GetMassKg() > 0 },
		func(cb models.CelestialBody) bool { return cb.Density > 0 },
		func(cb models.CelestialBody) bool { return cb.Gravity > 0 },
		func(cb models.CelestialBody) bool { return cb.SemimajorAxis > 0 },
		func(cb models.CelestialBody) bool { return cb.SideralOrbit > 0 },
		func(cb models.CelestialBody) bool { return cb.SideralRotation != 0 },
		func(cb models.CelestialBody) bool { return cb.Escape > 0 },
		func(cb models.CelestialBody) bool { return cb.EquaRadius > 0 },
		func(cb models.CelestialBody) bool { return cb.PolarRadius > 0 },
		func(cb models.CelestialBody) bool { return cb.Flattening > 0 },
		func(cb models.CelestialBody) bool { return cb.Eccentricity > 0 },
		func(cb models.CelestialBody) bool { return cb.Inclination != 0 },
		func(cb models.CelestialBody) bool { return cb.GetVolumeKm3() > 0 },
		func(cb models.CelestialBody) bool { return cb.Perihelion > 0 },
		func(cb models.CelestialBody) bool { return cb.Aphelion > 0 },
		func(cb models.CelestialBody) bool { return cb.Dimension != "" },
		func(cb models.CelestialBody) bool { return cb.DiscoveredBy != "" },
		func(cb models.CelestialBody) bool { return cb.DiscoveryDate != "" },
		func(cb models.CelestialBody) bool { return cb.AlternativeName != "" },
	}

	for _, fieldCheck := range fields {
		if fieldCheck(moon) {
			lines++
		}
	}

	if moon.ID != "" {
		lines++
	}
	if moon.Name != "" && moon.Name != moon.EnglishName {
		lines++
	}

	lines += 2 // Note about limited data + spacing

	return lines
}

func (ur *UIRenderer) GetModalDimensions(screenWidth, screenHeight int, dynamicHeight ...int) (modalX, modalY, modalWidth, modalHeight int) {
	modalWidth = constants.ModalWidth
	if len(dynamicHeight) > 0 {
		modalHeight = dynamicHeight[0]
	} else {
		modalHeight = constants.ModalHeight
	}
	modalX = screenWidth - modalWidth - constants.ModalMargin
	modalY = 1
	return
}

func (ur *UIRenderer) IsClickInModalArea(mouseX, mouseY int) bool {
	if !ur.state.ShowingDetails && !ur.state.ShowingMoons && !ur.state.ShowingMoonDetails && !ur.state.ShowingSystemList {
		return false
	}

	screenWidth, screenHeight := ur.screen.Size()
	var modalX, modalY, modalWidth, modalHeight int

	if ur.state.ShowingDetails {
		contentLines := ur.calculatePlanetDetailsLines(ur.state.SelectedPlanet)
		dynamicHeight := ui.Minimum(contentLines+6, screenHeight-4)
		modalX, modalY, modalWidth, modalHeight = ur.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)
	} else if ur.state.ShowingMoonDetails {
		contentLines := ur.calculateMoonDetailsLines(ur.state.SelectedMoon)
		dynamicHeight := ui.Minimum(contentLines+6, screenHeight-4)
		modalX, modalY, modalWidth, modalHeight = ur.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)
	} else {
		modalX, modalY, modalWidth, modalHeight = ur.GetModalDimensions(screenWidth, screenHeight)
	}

	return mouseX >= modalX && mouseX < modalX+modalWidth &&
		mouseY >= modalY && mouseY < modalY+modalHeight
}
