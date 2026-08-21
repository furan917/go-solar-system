package visualization

import (
	"fmt"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/gdamore/tcell/v2"
)

// PlanetPosition stores the screen coordinates and size of a planet
type PlanetPosition struct {
	X, Y   int
	Radius int
	Planet models.CelestialBody
}

// RendererDependencies encapsulates all dependencies for the Renderer
type RendererDependencies struct {
	CircleDrawer       *CircleDrawer
	CelestialRenderer  *CelestialObjectRenderer
	DebrisBeltRenderer *DebrisBeltRenderer
	DistanceScaler     *DistanceScaler
	MoonHandler        *MoonHandler
}

type Renderer struct {
	width              int
	height             int
	centerX            int
	centerY            int
	circleDrawer       *CircleDrawer
	celestialRenderer  *CelestialObjectRenderer
	debrisBeltRenderer *DebrisBeltRenderer
	distanceScaler     *DistanceScaler
	moonHandler        *MoonHandler
	simpleOrbits       bool
}

// NewRenderer creates a renderer with dependency injection
func NewRenderer(width, height int, deps RendererDependencies) *Renderer {
	return &Renderer{
		width:              width,
		height:             height,
		centerX:            width / 2,
		centerY:            height / 2,
		circleDrawer:       deps.CircleDrawer,
		celestialRenderer:  deps.CelestialRenderer,
		debrisBeltRenderer: deps.DebrisBeltRenderer,
		distanceScaler:     deps.DistanceScaler,
		moonHandler:        deps.MoonHandler,
		simpleOrbits:       SimpleOrbitsRequested(),
	}
}

// newOrbitCanvas returns a braille canvas, or nil when plain orbits are requested
func (r *Renderer) newOrbitCanvas(width, height int) *BrailleCanvas {
	if r.simpleOrbits {
		return nil
	}
	return NewBrailleCanvas(width, height)
}

// renderOrbit draws one orbital path into the canvas, or the grid when canvas is nil
func (r *Renderer) renderOrbit(grid [][]rune, canvas *BrailleCanvas, centerX, centerY int, radius float64) {
	if canvas == nil {
		r.celestialRenderer.RenderOrbit(grid, centerX, centerY, radius)
		return
	}
	r.celestialRenderer.RenderOrbitBraille(canvas, centerX, centerY, radius)
}

// NewRendererWithDefaults creates a renderer with default dependencies
func NewRendererWithDefaults(width, height int) *Renderer {
	circleDrawer := NewCircleDrawer(constants.AspectRatio)
	celestialRenderer := NewCelestialObjectRenderer(circleDrawer, width, height)
	distanceScaler := NewDistanceScaler(width, height)
	debrisBeltRenderer := NewDebrisBeltRenderer(circleDrawer, distanceScaler)
	moonHandler := NewMoonHandler()

	deps := RendererDependencies{
		CircleDrawer:       circleDrawer,
		CelestialRenderer:  celestialRenderer,
		DebrisBeltRenderer: debrisBeltRenderer,
		DistanceScaler:     distanceScaler,
		MoonHandler:        moonHandler,
	}

	return NewRenderer(width, height, deps)
}

func (r *Renderer) RenderSolarSystemData(planets []models.CelestialBody, width, height int) [][]rune {
	centerX := width / 2
	centerY := height / 2

	r.celestialRenderer.UpdateDimensions(r.width, r.height)

	grid := r.createGrid(width, height)
	canvas := r.newOrbitCanvas(width, height)

	stars, actualPlanets := r.separateStarsAndPlanets(planets)

	if len(stars) > 0 {
		r.celestialRenderer.RenderStars(grid, centerX, centerY, stars)
	} else {
		r.celestialRenderer.RenderSun(grid, centerX, centerY)
	}

	r.debrisBeltRenderer.RenderAsteroidBelt(grid, centerX, centerY, actualPlanets)
	r.debrisBeltRenderer.RenderKuiperBelt(grid, centerX, centerY, actualPlanets)

	for _, planet := range actualPlanets {
		if planet.SemimajorAxis <= 0 {
			continue
		}

		radius := r.distanceScaler.ScaleDistance(planet.SemimajorAxis, actualPlanets)

		r.renderOrbit(grid, canvas, centerX, centerY, radius)

		r.celestialRenderer.RenderPlanet(grid, centerX, centerY, planet, radius)
	}

	if canvas != nil {
		canvas.CompositeInto(grid)
	}

	return grid
}

// RenderSolarSystemDataWithPositions renders and returns planet positions for mouse interaction
func (r *Renderer) RenderSolarSystemDataWithPositions(planets []models.CelestialBody, width, height, screenWidth, screenHeight int) ([][]rune, map[string]PlanetPosition) {
	centerX := width / 2
	centerY := height / 2
	planetPositions := make(map[string]PlanetPosition)

	r.celestialRenderer.UpdateDimensions(screenWidth, screenHeight)

	grid := r.createGrid(width, height)
	canvas := r.newOrbitCanvas(width, height)

	stars, actualPlanets := r.separateStarsAndPlanets(planets)

	if len(stars) > 0 {
		r.celestialRenderer.RenderStars(grid, centerX, centerY, stars)
	} else {
		r.celestialRenderer.RenderSun(grid, centerX, centerY)
	}

	r.debrisBeltRenderer.RenderAsteroidBelt(grid, centerX, centerY, actualPlanets)
	r.debrisBeltRenderer.RenderKuiperBelt(grid, centerX, centerY, actualPlanets)

	for _, star := range stars {
		starRadius := r.celestialRenderer.GetSunSize() // Use sun size for now
		planetPositions[star.EnglishName] = PlanetPosition{
			X:      centerX, // Simplified - stars are at barycenter for interaction
			Y:      centerY,
			Radius: starRadius,
			Planet: star,
		}
	}

	for _, planet := range actualPlanets {
		if planet.SemimajorAxis <= 0 {
			continue
		}

		radius := r.distanceScaler.ScaleDistance(planet.SemimajorAxis, actualPlanets)

		r.renderOrbit(grid, canvas, centerX, centerY, radius)

		angle := r.celestialRenderer.GetOrbitalAngle(planet)
		px, py := r.circleDrawer.CalculatePosition(centerX, centerY, radius, angle)
		planetRadius := r.celestialRenderer.GetPlanetSize(planet.MeanRadius)

		planetPositions[planet.EnglishName] = PlanetPosition{
			X:      px,
			Y:      py,
			Radius: planetRadius,
			Planet: planet,
		}

		r.celestialRenderer.RenderPlanet(grid, centerX, centerY, planet, radius)
	}

	if canvas != nil {
		canvas.CompositeInto(grid)
	}

	return grid, planetPositions
}

// createGrid creates a new grid filled with spaces
func (r *Renderer) createGrid(width, height int) [][]rune {
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	return grid
}

// GetPlanetSymbol returns the Unicode symbol for a celestial body (delegated to celestial renderer)
func (r *Renderer) GetPlanetSymbol(name string) rune {
	return r.celestialRenderer.GetPlanetSymbol(name)
}

// GetMoonHandler returns the moon handler for external use
func (r *Renderer) GetMoonHandler() *MoonHandler {
	return r.moonHandler
}

// GetPlanetSize returns the scaled planet size for debugging
func (r *Renderer) GetPlanetSize(meanRadius float64) int {
	return r.celestialRenderer.GetPlanetSize(meanRadius)
}

// GetSunSize returns the scaled sun size for debugging
func (r *Renderer) GetSunSize() int {
	return r.celestialRenderer.GetSunSize()
}

// UpdateDimensions updates all renderer dimensions for dynamic resizing
func (r *Renderer) UpdateDimensions(width, height int) {
	r.width = width
	r.height = height
	r.centerX = width / 2
	r.centerY = height / 2

	r.celestialRenderer.UpdateDimensions(width, height)
	r.distanceScaler = NewDistanceScaler(width, height)
	r.debrisBeltRenderer = NewDebrisBeltRenderer(r.circleDrawer, r.distanceScaler)
}

// separateStarsAndPlanets separates celestial bodies into stars and planets
func (r *Renderer) separateStarsAndPlanets(bodies []models.CelestialBody) ([]models.CelestialBody, []models.CelestialBody) {
	var stars []models.CelestialBody
	var planets []models.CelestialBody

	for _, body := range bodies {
		if body.BodyType == "Star" || body.EnglishName == "Sun" || (body.SemimajorAxis == 0 && !body.IsPlanet) {
			stars = append(stars, body)
		} else {
			planets = append(planets, body)
		}
	}

	return stars, planets
}

func (r *Renderer) GetColorForSymbol(symbol rune) tcell.Color {
	return r.symbolToTcellColor(symbol)
}

func (r *Renderer) symbolToTcellColor(symbol rune) tcell.Color {
	colorMap := map[rune]tcell.Color{
		'☿': tcell.ColorGray,   // Mercury
		'♀': tcell.ColorYellow, // Venus
		'♁': tcell.ColorBlue,   // Earth
		'♂': tcell.ColorRed,    // Mars
		'♃': tcell.ColorOrange, // Jupiter
		'♄': tcell.ColorPurple, // Saturn
		'♅': tcell.ColorTeal,   // Uranus
		'♆': tcell.ColorNavy,   // Neptune
		'♇': tcell.ColorGray,   // Pluto
		'☉': tcell.ColorYellow, // Sun
		'✦': tcell.ColorBlue,   // Blue star
		'✧': tcell.ColorWhite,  // White star
		'✩': tcell.ColorOrange, // Orange star
		'✪': tcell.ColorRed,    // Red star
		'⭐': tcell.ColorWhite,  // Generic star
	}

	if assignedColor, exists := colorMap[symbol]; exists {
		return assignedColor
	}

	return tcell.ColorWhite
}

func (r *Renderer) RenderPlanetDetails(planet models.CelestialBody) []string {
	var details []string

	details = append(details, fmt.Sprintf("╔═══════════════════════════════════════════════════════════════════════════════╗"))
	details = append(details, fmt.Sprintf("║ %c %s", r.GetPlanetSymbol(planet.EnglishName), planet.EnglishName))
	details = append(details, fmt.Sprintf("╠═══════════════════════════════════════════════════════════════════════════════╣"))

	fields := r.getPlanetDetailFields(planet)
	for _, field := range fields {
		details = append(details, fmt.Sprintf("║ %s", field))
	}

	details = append(details, fmt.Sprintf("╚═══════════════════════════════════════════════════════════════════════════════╝"))

	return details
}

// getPlanetDetailFields returns formatted planet details using data-driven approach
func (r *Renderer) getPlanetDetailFields(planet models.CelestialBody) []string {
	type fieldConfig struct {
		label     string
		format    string
		unit      string
		condition func() bool
		value     func() interface{}
	}

	configs := []fieldConfig{
		{"Mean Radius", "%.0f", "km", func() bool { return planet.MeanRadius > 0 }, func() interface{} { return planet.MeanRadius }},
		{"Mass", "%.2e", "kg", func() bool { return planet.GetMassKg() > 0 }, func() interface{} { return planet.GetMassKg() }},
		{"Density", "%.2f", "g/cm³", func() bool { return planet.Density > 0 }, func() interface{} { return planet.Density }},
		{"Gravity", "%.2f", "m/s²", func() bool { return planet.Gravity > 0 }, func() interface{} { return planet.Gravity }},
		{"Distance from Sun", "%.0f", "km", func() bool { return planet.SemimajorAxis > 0 }, func() interface{} { return planet.SemimajorAxis }},
		{"Orbital Period", "%.2f", "days", func() bool { return planet.SideralOrbit > 0 }, func() interface{} { return planet.SideralOrbit }},
		{"Rotation Period", "%.2f", "hours", func() bool { return planet.SideralRotation != 0 }, func() interface{} { return planet.SideralRotation }},
	}

	var fields []string
	for _, config := range configs {
		if config.condition() {
			fields = append(fields, fmt.Sprintf("%s: %s %s", config.label, fmt.Sprintf(config.format, config.value()), config.unit))
		}
	}

	if len(planet.Moons) > 0 {
		fields = append(fields, fmt.Sprintf("Moons: %d", len(planet.Moons)))
		for i, moon := range planet.Moons {
			if i < 5 {
				fields = append(fields, fmt.Sprintf("  • %s", moon.EnglishName))
			} else if i == 5 {
				fields = append(fields, fmt.Sprintf("  • ... and %d more", len(planet.Moons)-5))
				break
			}
		}
	}

	stringFields := []struct {
		label     string
		condition func() bool
		value     func() string
	}{
		{"Discovered by", func() bool { return planet.DiscoveredBy != "" }, func() string { return planet.DiscoveredBy }},
		{"Discovery Date", func() bool { return planet.DiscoveryDate != "" }, func() string { return planet.DiscoveryDate }},
	}

	for _, field := range stringFields {
		if field.condition() {
			fields = append(fields, fmt.Sprintf("%s: %s", field.label, field.value()))
		}
	}

	return fields
}

func (r *Renderer) RenderMenu(planets []models.CelestialBody, selectedIndex int) []string {
	var menu []string

	menuComponents := []string{
		"╔══════════════════════════════════════════════════════════════════════════════════╗",
		"║                           🌌 Solar System Explorer                              ║",
		"╠══════════════════════════════════════════════════════════════════════════════════╣",
		"║ Use 'u'/'d' to navigate, Enter to select, 'q' to quit, 'b' to go back        ║",
		"╠══════════════════════════════════════════════════════════════════════════════════╣",
	}

	menu = append(menu, menuComponents...)

	for i, planet := range planets {
		menu = append(menu, r.formatPlanetMenuItem(planet, i == selectedIndex))
	}

	menu = append(menu, "╚══════════════════════════════════════════════════════════════════════════════════╝")

	return menu
}

// formatPlanetMenuItem formats a single planet menu item
func (r *Renderer) formatPlanetMenuItem(planet models.CelestialBody, isSelected bool) string {
	prefix := "║ "
	if isSelected {
		prefix = "║►"
	}

	symbol := r.GetPlanetSymbol(planet.EnglishName)
	moonCount := ""
	if len(planet.Moons) > 0 {
		moonCount = fmt.Sprintf(" (%d moons)", len(planet.Moons))
	}

	return fmt.Sprintf("%s %c %s%s", prefix, symbol, planet.EnglishName, moonCount)
}

func (r *Renderer) ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

func (r *Renderer) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func (r *Renderer) PrintLines(lines []string) {
	for _, line := range lines {
		fmt.Println(line)
	}
}
