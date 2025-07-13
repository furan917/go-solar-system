package app

import (
    "fmt"
    "strings"

    "github.com/furan917/go-solar-system/internal/constants"
    "github.com/furan917/go-solar-system/internal/display"
    "github.com/furan917/go-solar-system/internal/models"
    "github.com/gdamore/tcell/v2"
)

type AppRenderer struct {
    screen     tcell.Screen
    uiRenderer *UIRenderer
    state      *AppState
}

func NewAppRenderer(screen tcell.Screen, uiRenderer *UIRenderer, state *AppState) *AppRenderer {
    return &AppRenderer{
        screen:     screen,
        uiRenderer: uiRenderer,
        state:      state,
    }
}

func (ar *AppRenderer) DrawScreen() {
    ar.screen.Clear()

    width, height := ar.screen.Size()

    ar.drawText(2, 1, tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true), "🌌 Solar System Explorer")

    modalWidth := constants.ModalWidth
    availableWidth := width - modalWidth - (constants.ModalMargin * 3)
    ar.drawPlanetList(2, 3, availableWidth)

    ar.drawSolarSystem(2, 6, width-4, height-8)

    instructions := "Arrow keys to navigate • Enter/Click to select • S for systems • Q to quit • 1-9 for direct selection"
    systemDisplayName := ar.uiRenderer.GetSystemManager().GetCurrentSystemDisplayName()

    instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorLightBlue)
    systemStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

    ar.drawText(2, height-2, instructionStyle, instructions)
    ar.drawText(2+len(instructions)+3, height-2, systemStyle, fmt.Sprintf("• Current System: %s", systemDisplayName))

    if ar.state.IsShowingMoonDetails() {
        ar.drawMoonDetailsModal(width, height)
    } else if ar.state.IsShowingMoons() {
        ar.drawMoonListModal(width, height)
    } else if ar.state.IsShowingSystemList() {
        ar.drawSystemListModal(width, height)
    } else if ar.state.IsShowingDetails() {
        ar.drawPlanetDetailsModal(width, height)
    }

    ar.screen.Show()
}

func (ar *AppRenderer) drawText(x, y int, style tcell.Style, text string) {
    for i, r := range text {
        ar.screen.SetContent(x+i, y, r, nil, style)
    }
}

func (ar *AppRenderer) drawPlanetList(x, y, maxWidth int) {
    planets := ar.state.GetPlanets()
    if len(planets) == 0 {
        return
    }

    currentX := x
    currentY := y

    ar.state.ClearPlanetListPositions()

    positions := make([]PlanetListPosition, 0, len(planets))

    for i, planet := range planets {
        symbol := ar.uiRenderer.GetRenderer().GetPlanetSymbol(planet.EnglishName)

        style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
        if i == ar.state.SelectedIndex {
            style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true).Reverse(true)
        }

        planetTextLen := 4 + len(planet.EnglishName)

        if currentX+planetTextLen > maxWidth {
            currentY++
            currentX = x
        }

        positions = append(positions, PlanetListPosition{
            Index: i,
            X:     currentX,
            Y:     currentY,
            Width: planetTextLen,
        })

        ar.screen.SetContent(currentX, currentY, ' ', nil, style)
        ar.screen.SetContent(currentX+1, currentY, symbol, nil, style)
        ar.screen.SetContent(currentX+2, currentY, ' ', nil, style)

        nameStart := currentX + 3
        for j, r := range planet.EnglishName {
            ar.screen.SetContent(nameStart+j, currentY, r, nil, style)
        }
        ar.screen.SetContent(nameStart+len(planet.EnglishName), currentY, ' ', nil, style)

        currentX += planetTextLen + 2
    }

    for _, pos := range positions {
        ar.state.AddPlanetListPosition(pos)
    }
}

func (ar *AppRenderer) drawSolarSystem(x, y, width, height int) {
    screenWidth, screenHeight := ar.screen.Size()
    planets := ar.state.GetPlanets()

    solarSystemData, planetPositions := ar.uiRenderer.GetRenderer().RenderSolarSystemDataWithPositions(planets, width, height, screenWidth, screenHeight)

    ar.state.UpdatePlanetPositions(x, y, planetPositions)

    maxRow := minimum(len(solarSystemData), height)
    for row := 0; row < maxRow; row++ {
        line := solarSystemData[row]
        maxCol := minimum(len(line), width)

        for col := 0; col < maxCol; col++ {
            char := line[col]
            if char == ' ' {
                continue
            }

            var style tcell.Style

            switch char {
            case '·':
                style = tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
            case '∗':
                style = tcell.StyleDefault.Foreground(tcell.ColorOrange)
            case '◦':
                style = tcell.StyleDefault.Foreground(tcell.ColorBlue)
            case '☉', '☿', '♀', '♁', '♂', '♃', '♄', '♅', '♆':
                style = ar.getPlanetStyle(char)
            default:
                style = tcell.StyleDefault.Foreground(tcell.ColorWhite)
            }

            ar.screen.SetContent(x+col, y+row, char, nil, style)
        }
    }
}

func (ar *AppRenderer) getPlanetStyle(planet rune) tcell.Style {
    switch planet {
    case '☉':
        return tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
    case '☿':
        return tcell.StyleDefault.Foreground(tcell.ColorGray).Bold(true)
    case '♀':
        return tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
    case '♁':
        return tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
    case '♂':
        return tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
    case '♃':
        return tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true)
    case '♄':
        return tcell.StyleDefault.Foreground(tcell.ColorPurple).Bold(true)
    case '♅':
        return tcell.StyleDefault.Foreground(tcell.ColorTeal).Bold(true)
    case '♆':
        return tcell.StyleDefault.Foreground(tcell.ColorNavy).Bold(true)
    default:
        return tcell.StyleDefault.Foreground(tcell.ColorWhite)
    }
}

func (ar *AppRenderer) drawPlanetDetailsModal(screenWidth, screenHeight int) {
    planet := ar.state.SelectedPlanet
    contentLines := ar.calculatePlanetDetailsLines(planet)
    dynamicHeight := minimum(contentLines+6, screenHeight-4)
    modalX, modalY, _, modalHeight := ar.setupModal(screenWidth, screenHeight, dynamicHeight)

    symbol := ar.uiRenderer.GetRenderer().GetPlanetSymbol(planet.EnglishName)
    titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
    title := fmt.Sprintf(" %c %s ", symbol, planet.EnglishName)
    ar.drawTextAt(modalX+2, modalY+1, titleStyle, title)

    detailStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
    currentY := modalY + 3

    currentY = ar.drawCelestialBodyDetails(planet, modalX+2, currentY, detailStyle)

    if len(planet.Moons) > 0 {
        moonHandler := ar.uiRenderer.GetRenderer().GetMoonHandler()
        moonLines := moonHandler.FormatMoonDisplay(planet, 5)

        for i, line := range moonLines {
            if i == 0 {
                ar.drawTextAt(modalX+2, currentY, detailStyle, line)
                currentY += 2
            } else {
                ar.drawTextAt(modalX+4, currentY, detailStyle, line)
                currentY++
            }
        }
    }

    instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
    instruction := "Press Enter, Escape, or 'b' to close"
    if len(planet.Moons) > 0 {
        instruction += " • 'm' for moons"
    }
    ar.drawTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, instruction)
}

func (ar *AppRenderer) drawMoonListModal(screenWidth, screenHeight int) {
    modalX, modalY, modalWidth, modalHeight := ar.setupModal(screenWidth, screenHeight)

    titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
    title := fmt.Sprintf(" %s Moons (%d total) ", ar.state.SelectedPlanet.EnglishName, len(ar.state.SelectedPlanet.Moons))
    ar.drawTextAt(modalX+2, modalY+1, titleStyle, title)

    moonHandler := ar.uiRenderer.GetRenderer().GetMoonHandler()
    moonNames := moonHandler.GetMoonNames(ar.state.SelectedPlanet)

    if len(moonNames) == 0 {
        for i := 0; i < len(ar.state.SelectedPlanet.Moons); i++ {
            moonNames = append(moonNames, fmt.Sprintf("Moon %d", i+1))
        }
    }

    visibleItems := constants.MaxVisibleItems
    startY := modalY + 3

    scrollAreaStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)

    for i := 0; i < visibleItems; i++ {
        ar.screen.SetContent(modalX+modalWidth-3, startY+i, '│', nil, scrollAreaStyle)
    }

    if len(moonNames) > visibleItems {
        totalScrollable := len(moonNames) - visibleItems
        scrollPosition := int(float64(ar.state.MoonScrollIndex) / float64(totalScrollable) * float64(visibleItems-1))
        ar.screen.SetContent(modalX+modalWidth-3, startY+scrollPosition, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue))
    }

    if ar.state.MoonScrollIndex > 0 {
        ar.drawTextAt(modalX+modalWidth-2, modalY+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
        ar.drawTextAt(modalX+modalWidth-8, modalY+2, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "More")
    }
    if ar.state.MoonScrollIndex+visibleItems < len(moonNames) {
        ar.drawTextAt(modalX+modalWidth-2, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
        ar.drawTextAt(modalX+modalWidth-8, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "More")
    }

    for i := 0; i < visibleItems && i+ar.state.MoonScrollIndex < len(moonNames); i++ {
        moonIndex := i + ar.state.MoonScrollIndex
        moonName := moonNames[moonIndex]

        style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
        if moonIndex == ar.state.MoonSelectedIndex {
            style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true).Reverse(true)
        }

        prefix := "  "
        if moonIndex == ar.state.MoonSelectedIndex {
            prefix = "► "
        }

        moonText := fmt.Sprintf("%s%d. %s", prefix, moonIndex+1, moonName)
        ar.drawTextAt(modalX+2, startY+i, style, moonText)
    }

    statusStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue)
    statusText := fmt.Sprintf("Showing %d-%d of %d moons",
        ar.state.MoonScrollIndex+1,
        minimum(ar.state.MoonScrollIndex+visibleItems, len(moonNames)),
        len(moonNames))
    ar.drawTextAt(modalX+2, modalY+modalHeight-3, statusStyle, statusText)

    instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
    ar.drawWrappedTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, "↑/↓ to navigate • Enter to select • Escape/'b' to go back", constants.ModalContentWidth)
}

func (ar *AppRenderer) drawMoonDetailsModal(screenWidth, screenHeight int) {
    contentLines := ar.calculateMoonDetailsLines(ar.state.SelectedMoon)
    dynamicHeight := minimum(contentLines+6, screenHeight-4)
    modalX, modalY, _, modalHeight := ar.setupModal(screenWidth, screenHeight, dynamicHeight)

    titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
    title := fmt.Sprintf(" %s (Moon of %s) ", ar.state.SelectedMoon.EnglishName, ar.state.SelectedPlanet.EnglishName)
    ar.drawTextAt(modalX+2, modalY+1, titleStyle, title)

    detailStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
    currentY := modalY + 2
    currentY++

    if ar.state.SelectedMoon.ID != "" {
        currentY = ar.drawWrappedTextAt(modalX+2, currentY, detailStyle, fmt.Sprintf("ID: %s", ar.state.SelectedMoon.ID), constants.ModalContentWidth)
        currentY++
    }

    if ar.state.SelectedMoon.Name != "" && ar.state.SelectedMoon.Name != ar.state.SelectedMoon.EnglishName {
        currentY = ar.drawWrappedTextAt(modalX+2, currentY, detailStyle, fmt.Sprintf("Original Name: %s", ar.state.SelectedMoon.Name), constants.ModalContentWidth)
        currentY++
    }

    currentY = ar.drawWrappedTextAt(modalX+2, currentY, detailStyle, fmt.Sprintf("Orbits: %s", ar.state.SelectedPlanet.EnglishName), constants.ModalContentWidth)
    currentY++

    currentY = ar.drawCelestialBodyDetails(ar.state.SelectedMoon, modalX+2, currentY, detailStyle)

    ar.drawWrappedTextAt(modalX+2, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDarkBlue), "Note: Limited moon data available from API", constants.ModalContentWidth)

    instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
    ar.drawWrappedTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, "Press Enter, Escape, or 'b' to go back to moon list", constants.ModalContentWidth)
}

func (ar *AppRenderer) drawSystemListModal(screenWidth, screenHeight int) {
    modalX, modalY, modalWidth, modalHeight := ar.setupModal(screenWidth, screenHeight)

    titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true)
    title := " 🌌 Star System Selection "
    ar.drawTextAt(modalX+2, modalY+1, titleStyle, title)

    systemInfo, err := ar.uiRenderer.GetSystemManager().ListSystemsWithInfo()
    if err != nil {
        ar.drawTextAt(modalX+2, modalY+3, tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorDarkBlue), "Error loading system information")
        return
    }

    visibleItems := constants.MaxVisibleItems
    startY := modalY + 3

    if ar.state.SystemScrollIndex > 0 {
        ar.drawTextAt(modalX+modalWidth-2, modalY+2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↑")
    }
    if ar.state.SystemScrollIndex+visibleItems < len(systemInfo) {
        ar.drawTextAt(modalX+modalWidth-2, modalY+modalHeight-3, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true), "↓")
    }

    for i := 0; i < visibleItems && i+ar.state.SystemScrollIndex < len(systemInfo); i++ {
        systemIndex := i + ar.state.SystemScrollIndex
        systemLine := systemInfo[systemIndex]

        style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)
        if systemIndex == ar.state.SystemSelectedIndex {
            style = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true).Reverse(true)
        }

        maxLineLength := constants.ModalContentWidth
        wrappedLines := ar.wrapText(systemLine, maxLineLength)

        if len(wrappedLines) > 0 {
            ar.drawTextAt(modalX+2, startY+i, style, wrappedLines[0])
        }
    }

    instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue)
    ar.drawWrappedTextAt(modalX+2, modalY+modalHeight-2, instructionStyle, "↑/↓ to navigate • Enter to select • Escape/'b' to cancel", constants.ModalContentWidth)
}

func (ar *AppRenderer) drawTextAt(x, y int, style tcell.Style, text string) {
    for i, r := range text {
        ar.screen.SetContent(x+i, y, r, nil, style)
    }
}

func (ar *AppRenderer) wrapText(text string, maxWidth int) []string {
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

func (ar *AppRenderer) drawWrappedTextAt(x, y int, style tcell.Style, text string, maxWidth int) int {
    lines := ar.wrapText(text, maxWidth)
    for i, line := range lines {
        ar.drawTextAt(x, y+i, style, line)
    }
    return y + len(lines)
}

func (ar *AppRenderer) setupModal(screenWidth, screenHeight int, dynamicHeight ...int) (modalX, modalY, modalWidth, modalHeight int) {
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
            ar.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(tcell.ColorDarkBlue))
        }
    }

    ar.drawModalBorder(modalX, modalY, modalWidth, modalHeight)

    return modalX, modalY, modalWidth, modalHeight
}

func (ar *AppRenderer) drawModalBorder(x, y, width, height int) {
    borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue).Bold(true)

    for i := x; i < x+width; i++ {
        ar.screen.SetContent(i, y, '═', nil, borderStyle)
        ar.screen.SetContent(i, y+height-1, '═', nil, borderStyle)
    }

    for i := y; i < y+height; i++ {
        ar.screen.SetContent(x, i, '║', nil, borderStyle)
        ar.screen.SetContent(x+width-1, i, '║', nil, borderStyle)
    }

    ar.screen.SetContent(x, y, '╔', nil, borderStyle)
    ar.screen.SetContent(x+width-1, y, '╗', nil, borderStyle)
    ar.screen.SetContent(x, y+height-1, '╚', nil, borderStyle)
    ar.screen.SetContent(x+width-1, y+height-1, '╝', nil, borderStyle)
}

func (ar *AppRenderer) calculatePlanetDetailsLines(planet models.CelestialBody) int {
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

    if len(planet.Moons) > 0 {
        moonHandler := ar.uiRenderer.GetRenderer().GetMoonHandler()
        moonLines := moonHandler.FormatMoonDisplay(planet, 5)
        lines += len(moonLines) + 1
    }

    return lines
}

func (ar *AppRenderer) calculateMoonDetailsLines(moon models.CelestialBody) int {
    lines := 1

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

    lines += 2

    return lines
}

func (ar *AppRenderer) drawCelestialBodyDetails(body models.CelestialBody, x, y int, style tcell.Style) int {
    currentY := y

    stringFields := display.GetCelestialBodyStringFields()
    for _, field := range stringFields {
        if field.Condition(body) {
            detail := field.FormatStringFieldValue(body)
            currentY = ar.drawWrappedTextAt(x, currentY, style, detail, constants.ModalContentWidth)
        }
    }

    fields := display.GetCelestialBodyFields()
    for _, field := range fields {
        if field.Condition(body) {
            detail := field.FormatFieldValue(body)
            currentY = ar.drawWrappedTextAt(x, currentY, style, detail, constants.ModalContentWidth)
        }
    }

    return currentY
}

func (ar *AppRenderer) GetModalDimensions(screenWidth, screenHeight int, dynamicHeight ...int) (modalX, modalY, modalWidth, modalHeight int) {
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

func (ar *AppRenderer) IsClickInModalArea(mouseX, mouseY int) bool {
    if !ar.state.ShowingDetails && !ar.state.ShowingMoons && !ar.state.ShowingMoonDetails && !ar.state.ShowingSystemList {
        return false
    }

    screenWidth, screenHeight := ar.screen.Size()
    var modalX, modalY, modalWidth, modalHeight int

    if ar.state.ShowingDetails {
        contentLines := ar.calculatePlanetDetailsLines(ar.state.SelectedPlanet)
        dynamicHeight := minimum(contentLines+6, screenHeight-4)
        modalX, modalY, modalWidth, modalHeight = ar.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)
    } else if ar.state.ShowingMoonDetails {
        contentLines := ar.calculateMoonDetailsLines(ar.state.SelectedMoon)
        dynamicHeight := minimum(contentLines+6, screenHeight-4)
        modalX, modalY, modalWidth, modalHeight = ar.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)
    } else {
        modalX, modalY, modalWidth, modalHeight = ar.GetModalDimensions(screenWidth, screenHeight)
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
