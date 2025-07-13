package app

import (
	"fmt"
	"strings"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/display"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/systems"
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
		ur.drawMoonDetailsModal(width, height)
	} else if ur.state.IsShowingMoons() {
		ur.drawMoonListModal(width, height)
	} else if ur.state.IsShowingSystemList() {
		ur.drawSystemListModal(width, height)
	} else if ur.state.IsShowingDetails() {
		ur.drawPlanetDetailsModal(width, height)
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

// Modal rendering methods moved from app.go
func (ur *UIRenderer) drawPlanetDetailsModal(width, height int) {
	planet := ur.state.SelectedPlanet
	contentLines := ur.calculatePlanetDetailsLines(planet)
	dynamicHeight := minimum(contentLines+6, height-4) // 6 for borders, title, instructions
	modalX, modalY, _, modalHeight := ur.setupModal(width, height, dynamicHeight)

	symbol := ur.renderer.GetPlanetSymbol(planet.EnglishName)
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
	title := fmt.Sprintf(" %c %s ", symbol, planet.EnglishName)
	ur.drawText(modalX+2, modalY+1, titleStyle, title)

	detailStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	currentY := modalY + 3

	currentY = ur.drawCelestialBodyDetails(planet, modalX+2, currentY, detailStyle)

	if len(planet.Moons) > 0 {
		moonHandler := ur.renderer.GetMoonHandler()
		moonLines := moonHandler.FormatMoonDisplay(planet, 5)

		for i, line := range moonLines {
			if i == 0 {
				ur.drawText(modalX+2, currentY, detailStyle, line)
				currentY += 2
			} else {
				ur.drawText(modalX+4, currentY, detailStyle, line)
				currentY++
			}
		}
	}

	instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
	instruction := "Press Enter, Escape, or 'b' to close"
	if len(planet.Moons) > 0 {
		instruction += " • 'm' for moons"
	}
	ur.drawText(modalX+2, modalY+modalHeight-2, instructionStyle, instruction)
}

func (ur *UIRenderer) drawMoonListModal(width, height int) {
	modalX, modalY, modalWidth, modalHeight := ur.setupModal(width, height)

	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
	title := fmt.Sprintf(" %s Moons (%d total) ", ur.state.SelectedPlanet.EnglishName, len(ur.state.SelectedPlanet.Moons))
	ur.drawText(modalX+2, modalY+1, titleStyle, title)

	moonHandler := ur.renderer.GetMoonHandler()
	moonNames := moonHandler.GetMoonNames(ur.state.SelectedPlanet)

	if len(moonNames) == 0 {
		for i := 0; i < len(ur.state.SelectedPlanet.Moons); i++ {
			moonNames = append(moonNames, fmt.Sprintf("Moon %d", i+1))
		}
	}

	visibleItems := constants.MaxVisibleItems
	startY := modalY + 3

	scrollAreaStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)

	for i := 0; i < visibleItems; i++ {
		ur.screen.SetContent(modalX+modalWidth-3, startY+i, '│', nil, scrollAreaStyle)
	}

	if len(moonNames) > visibleItems {
		totalScrollable := len(moonNames) - visibleItems
		scrollPosition := int(float64(ur.state.MoonScrollIndex) / float64(totalScrollable) * float64(visibleItems-1))
		ur.screen.SetContent(modalX+modalWidth-3, startY+scrollPosition, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue))
	}

	if ur.state.MoonScrollIndex > 0 {
		ur.drawText(modalX+modalWidth-2, modalY+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
		ur.drawText(modalX+modalWidth-8, modalY+2, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "More")
	}
	if ur.state.MoonScrollIndex+visibleItems < len(moonNames) {
		ur.drawText(modalX+modalWidth-2, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
		ur.drawText(modalX+modalWidth-8, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "More")
	}

	for i := 0; i < visibleItems && i+ur.state.MoonScrollIndex < len(moonNames); i++ {
		moonIndex := i + ur.state.MoonScrollIndex
		moonName := moonNames[moonIndex]

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
		if moonIndex == ur.state.MoonSelectedIndex {
			style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true).Reverse(true)
		}

		prefix := "  "
		if moonIndex == ur.state.MoonSelectedIndex {
			prefix = "► "
		}

		moonText := fmt.Sprintf("%s%d. %s", prefix, moonIndex+1, moonName)
		ur.drawText(modalX+2, startY+i, style, moonText)
	}

	statusStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)
	statusText := fmt.Sprintf("Showing %d-%d of %d moons",
		ur.state.MoonScrollIndex+1,
		minimum(ur.state.MoonScrollIndex+visibleItems, len(moonNames)),
		len(moonNames))
	ur.drawText(modalX+2, modalY+modalHeight-3, statusStyle, statusText)

	instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
	ur.drawWrappedTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, "↑/↓ to navigate • Enter to select • Escape/'b' to go back", constants.ModalContentWidth)
}

func (ur *UIRenderer) drawMoonDetailsModal(width, height int) {
	contentLines := ur.calculateMoonDetailsLines(ur.state.SelectedMoon)
	dynamicHeight := minimum(contentLines+6, height-4) // 6 for borders, title, instructions
	modalX, modalY, _, modalHeight := ur.setupModal(width, height, dynamicHeight)

	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
	title := fmt.Sprintf(" %s (Moon of %s) ", ur.state.SelectedMoon.EnglishName, ur.state.SelectedPlanet.EnglishName)
	ur.drawText(modalX+2, modalY+1, titleStyle, title)

	detailStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
	currentY := modalY + 2
	currentY++

	if ur.state.SelectedMoon.ID != "" {
		currentY = ur.drawWrappedTextAt(modalX+2, currentY, detailStyle, fmt.Sprintf("ID: %s", ur.state.SelectedMoon.ID), constants.ModalContentWidth)
		currentY++
	}

	if ur.state.SelectedMoon.Name != "" && ur.state.SelectedMoon.Name != ur.state.SelectedMoon.EnglishName {
		currentY = ur.drawWrappedTextAt(modalX+2, currentY, detailStyle, fmt.Sprintf("Original Name: %s", ur.state.SelectedMoon.Name), constants.ModalContentWidth)
		currentY++
	}

	currentY = ur.drawWrappedTextAt(modalX+2, currentY, detailStyle, fmt.Sprintf("Orbits: %s", ur.state.SelectedPlanet.EnglishName), constants.ModalContentWidth)
	currentY++

	currentY = ur.drawCelestialBodyDetails(ur.state.SelectedMoon, modalX+2, currentY, detailStyle)

	ur.drawWrappedTextAt(modalX+2, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "Note: Limited moon data available from API", constants.ModalContentWidth)

	instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
	ur.drawWrappedTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, "Press Enter, Escape, or 'b' to go back to moon list", constants.ModalContentWidth)
}

func (ur *UIRenderer) drawSystemListModal(width, height int) {
	modalX, modalY, modalWidth, modalHeight := ur.setupModal(width, height)

	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
	title := " 🌌 Star System Selection "
	ur.drawText(modalX+2, modalY+1, titleStyle, title)

	systemInfo, err := ur.systemManager.ListSystemsWithInfo()
	if err != nil {
		ur.drawText(modalX+2, modalY+3, tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorDarkBlue), "Error loading system information")
		return
	}

	visibleItems := constants.MaxVisibleItems
	startY := modalY + 3

	if ur.state.SystemScrollIndex > 0 {
		ur.drawText(modalX+modalWidth-2, modalY+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
	}
	if ur.state.SystemScrollIndex+visibleItems < len(systemInfo) {
		ur.drawText(modalX+modalWidth-2, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
	}

	for i := 0; i < visibleItems && i+ur.state.SystemScrollIndex < len(systemInfo); i++ {
		systemIndex := i + ur.state.SystemScrollIndex
		systemLine := systemInfo[systemIndex]

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
		if systemIndex == ur.state.SystemSelectedIndex {
			style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true).Reverse(true)
		}

		maxLineLength := constants.ModalContentWidth
		wrappedLines := ur.wrapText(systemLine, maxLineLength)

		if len(wrappedLines) > 0 {
			ur.drawText(modalX+2, startY+i, style, wrappedLines[0])
		}
	}

	instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
	ur.drawWrappedTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, "↑/↓ to navigate • Enter to select • Escape/'b' to cancel", constants.ModalContentWidth)
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

// Supporting methods for modal rendering

// setupModal handles all common modal configuration and drawing setup
func (ur *UIRenderer) setupModal(screenWidth, screenHeight int, dynamicHeight ...int) (modalX, modalY, modalWidth, modalHeight int) {
	modalWidth = constants.ModalWidth
	if len(dynamicHeight) > 0 {
		modalHeight = dynamicHeight[0]
	} else {
		modalHeight = constants.ModalHeight
	}
	modalX = screenWidth - modalWidth - constants.ModalMargin
	modalY = 1

	for y := modalY; y < modalY+modalHeight; y++ {
		for x := modalX; x < modalX+modalWidth; x++ {
			ur.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(tcell.ColorDarkBlue))
		}
	}

	ur.drawModalBorder(modalX, modalY, modalWidth, modalHeight)

	return modalX, modalY, modalWidth, modalHeight
}

// drawModalBorder draws the modal border
func (ur *UIRenderer) drawModalBorder(x, y, width, height int) {
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue).Bold(true)

	for i := x; i < x+width; i++ {
		ur.screen.SetContent(i, y, '═', nil, borderStyle)
		ur.screen.SetContent(i, y+height-1, '═', nil, borderStyle)
	}

	for i := y; i < y+height; i++ {
		ur.screen.SetContent(x, i, '║', nil, borderStyle)
		ur.screen.SetContent(x+width-1, i, '║', nil, borderStyle)
	}

	// Corners
	ur.screen.SetContent(x, y, '╔', nil, borderStyle)
	ur.screen.SetContent(x+width-1, y, '╗', nil, borderStyle)
	ur.screen.SetContent(x, y+height-1, '╚', nil, borderStyle)
	ur.screen.SetContent(x+width-1, y+height-1, '╝', nil, borderStyle)
}

func (ur *UIRenderer) wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+1+len(word) > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				lines = append(lines, word[:maxWidth-3]+"...")
				currentLine = ""
			}
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// drawWrappedTextAt draws text with wrapping and returns the next Y position
func (ur *UIRenderer) drawWrappedTextAt(x, y int, style tcell.Style, text string, maxWidth int) int {
	lines := ur.wrapText(text, maxWidth)
	for i, line := range lines {
		ur.drawText(x, y+i, style, line)
	}
	return y + len(lines)
}

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

// drawCelestialBodyDetails draws celestial body details using a data-driven approach
func (ur *UIRenderer) drawCelestialBodyDetails(body models.CelestialBody, x, y int, style tcell.Style) int {
	currentY := y

	stringFields := display.GetCelestialBodyStringFields()
	for _, field := range stringFields {
		if field.Condition(body) {
			detail := field.FormatStringFieldValue(body)
			currentY = ur.drawWrappedTextAt(x, currentY, style, detail, constants.ModalContentWidth)
		}
	}

	fields := display.GetCelestialBodyFields()
	for _, field := range fields {
		if field.Condition(body) {
			detail := field.FormatFieldValue(body)
			currentY = ur.drawWrappedTextAt(x, currentY, style, detail, constants.ModalContentWidth)
		}
	}

	return currentY
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
		dynamicHeight := minimum(contentLines+6, screenHeight-4)
		modalX, modalY, modalWidth, modalHeight = ur.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)
	} else if ur.state.ShowingMoonDetails {
		contentLines := ur.calculateMoonDetailsLines(ur.state.SelectedMoon)
		dynamicHeight := minimum(contentLines+6, screenHeight-4)
		modalX, modalY, modalWidth, modalHeight = ur.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)
	} else {
		modalX, modalY, modalWidth, modalHeight = ur.GetModalDimensions(screenWidth, screenHeight)
	}

	return mouseX >= modalX && mouseX < modalX+modalWidth &&
		mouseY >= modalY && mouseY < modalY+modalHeight
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
