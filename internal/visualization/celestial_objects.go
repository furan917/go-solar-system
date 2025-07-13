package visualization

import (
	"math"
	"time"

	"github.com/furan917/go-solar-system/internal/models"
)

// StarPosition represents the position of a star in the visualization
type StarPosition struct {
	X, Y int
}

// CelestialObjectRenderer handles rendering of celestial objects
type CelestialObjectRenderer struct {
	circleDrawer *CircleDrawer
	startTime    time.Time
	epochTime    time.Time
	width        int
	height       int
}

// NewCelestialObjectRenderer creates a new celestial object renderer
func NewCelestialObjectRenderer(circleDrawer *CircleDrawer, width, height int) *CelestialObjectRenderer {
	epoch := time.Now()
	return &CelestialObjectRenderer{
		circleDrawer: circleDrawer,
		startTime:    time.Now(),
		epochTime:    epoch,
		width:        width,
		height:       height,
	}
}

// RenderSun renders the sun at the center
func (cor *CelestialObjectRenderer) RenderSun(grid [][]rune, centerX, centerY int) {
	sunRadius := cor.scaleSunSize()
	cor.circleDrawer.DrawFilledCircle(grid, centerX, centerY, sunRadius, '☉')
}

// RenderStars renders multiple stars for multi-star systems
func (cor *CelestialObjectRenderer) RenderStars(grid [][]rune, centerX, centerY int, stars []models.CelestialBody) {
	if len(stars) == 1 {
		starRadius := cor.scaleStarSize(stars[0].MeanRadius, len(stars))
		symbol := cor.getStarSymbol(stars[0])
		cor.circleDrawer.DrawFilledCircle(grid, centerX, centerY, starRadius, symbol)
		return
	}

	positions := cor.calculateStarPositions(stars, centerX, centerY)

	for i, star := range stars {
		if i < len(positions) {
			starRadius := cor.scaleStarSize(star.MeanRadius, len(stars))
			symbol := cor.getStarSymbol(star)

			px, py := positions[i].X, positions[i].Y
			if starRadius <= 1 {
				if cor.circleDrawer.isInBounds(px, py, len(grid[0]), len(grid)) {
					grid[py][px] = symbol
				}
			} else {
				cor.circleDrawer.DrawFilledCircle(grid, px, py, starRadius, symbol)
			}
		}
	}
}

// RenderPlanet renders a planet at its orbital position
func (cor *CelestialObjectRenderer) RenderPlanet(grid [][]rune, centerX, centerY int, planet models.CelestialBody, radius float64) {
	angle := cor.getOrbitalAngle(planet)
	px, py := cor.circleDrawer.CalculatePosition(centerX, centerY, radius, angle)

	planetRadius := cor.scalePlanetSize(planet.MeanRadius)
	symbol := cor.GetPlanetSymbol(planet.EnglishName)

	if planetRadius <= 1 {
		if cor.circleDrawer.isInBounds(px, py, len(grid[0]), len(grid)) {
			grid[py][px] = symbol
		}
	} else {
		cor.circleDrawer.DrawFilledCircle(grid, px, py, planetRadius, symbol)
	}
}

// RenderOrbit renders an orbital path
func (cor *CelestialObjectRenderer) RenderOrbit(grid [][]rune, centerX, centerY int, radius float64) {
	cor.circleDrawer.DrawCircle(grid, centerX, centerY, radius, '·')
}

// getOrbitalAngle calculates the current orbital angle for a planet using realistic orbital mechanics
func (cor *CelestialObjectRenderer) getOrbitalAngle(planet models.CelestialBody) float64 {
	if planet.SideralOrbit <= 0 {
		return 0
	}

	// Calculate mean anomaly based on real orbital mechanics
	meanAnomaly := cor.calculateMeanAnomaly(planet)

	// For visualization purposes, we'll use a simplified approach
	// In a real implementation, you'd solve Kepler's equation for eccentric anomaly
	// then calculate true anomaly, but for circular orbits this approximation works well

	// Apply eccentricity correction (simplified)
	if planet.Eccentricity > 0 {
		// Simple approximation: true anomaly ≈ mean anomaly + 2*e*sin(mean anomaly)
		trueAnomaly := meanAnomaly + 2*planet.Eccentricity*math.Sin(meanAnomaly)
		return math.Mod(trueAnomaly, 2*math.Pi)
	}

	return math.Mod(meanAnomaly, 2*math.Pi)
}

// scalePlanetSize scales planet size based on actual radius data and terminal size
func (cor *CelestialObjectRenderer) scalePlanetSize(meanRadius float64) int {
	if meanRadius <= 0 {
		return 1
	}

	terminalSizeFactor := cor.getTerminalSizeFactor()

	// Use logarithmic scaling for more realistic size representation
	// Enhanced scaling to show clear differences between planet sizes
	logRadius := math.Log10(meanRadius)

	baseSize := 1
	switch {
	case logRadius >= 4.8: // >= ~63,000 km (Jupiter-class) - Much larger
		baseSize = 3
	case logRadius >= 4.7: // >= ~50,000 km (Saturn-class) - Large
		baseSize = 2
	case logRadius >= 4.3: // >= ~20,000 km (Ice giant-class) - Medium-large
		baseSize = 2
	case logRadius >= 3.7: // >= ~5,000 km (Earth-class) - Medium
		baseSize = 1
	case logRadius >= 3.3: // >= ~2,000 km (Mars-class) - Small-medium
		baseSize = 1
	default: // < ~2,000 km (Small body-class) - Small
		baseSize = 1
	}

	scaledSize := int(float64(baseSize) * terminalSizeFactor)
	if scaledSize < 1 {
		scaledSize = 1
	}

	maxSizes := map[int]int{
		3: 3, // Jupiter-class
		2: 2, // Saturn/Ice giant-class:
		1: 1, // Earth/Small planets
	}

	if maxSize, exists := maxSizes[baseSize]; exists && scaledSize > maxSize {
		scaledSize = maxSize
	}

	return scaledSize
}

// scaleSunSize scales the sun's size based on terminal dimensions
func (cor *CelestialObjectRenderer) scaleSunSize() int {
	terminalSizeFactor := cor.getTerminalSizeFactor()

	// Much smaller base sun size to stay within first orbit
	baseSize := 3.0
	scaledSize := int(baseSize * terminalSizeFactor)

	// Ensure minimum size of 2 for visibility
	if scaledSize < 2 {
		scaledSize = 2
	}

	// Strict maximum to never extend beyond first orbit
	if scaledSize > 4 {
		scaledSize = 4
	}

	return scaledSize
}

// scaleStarSize scales star size for both single and multi-star systems
func (cor *CelestialObjectRenderer) scaleStarSize(meanRadius float64, numStars int) int {
	if numStars == 1 {
		// Single star system - use the existing sun scaling logic
		return cor.scaleSunSize()
	}

	// Multi-star systems - return radius 2 for 3 vertical lines
	// Radius 2 = 3 vertical lines (center + 1 above + 1 below)
	return 2
}

// getTerminalSizeFactor calculates a scaling factor based on terminal dimensions
func (cor *CelestialObjectRenderer) getTerminalSizeFactor() float64 {
	// Use the smaller dimension to determine scaling factor
	// This ensures planets scale appropriately in both narrow and wide terminals
	minDimension := math.Min(float64(cor.width), float64(cor.height))

	// Moderate scaling with reasonable bounds
	// Reference: 80x24 (small) = 0.67, 120x36 (medium) = 1.0, 200x60 (large) = 1.67, 300x80 (very large) = 2.22
	baseDimension := 36.0 // Medium terminal height as reference
	sizeFactor := minDimension / baseDimension

	if sizeFactor < 0.5 {
		sizeFactor = 0.5
	} else if sizeFactor > 2.5 {
		sizeFactor = 2.5
	}

	return sizeFactor
}

// GetPlanetSymbol returns the Unicode symbol for a celestial body
func (cor *CelestialObjectRenderer) GetPlanetSymbol(name string) rune {
	// Known solar system symbols for backward compatibility
	knownSymbols := map[string]rune{
		"Sun":     '☉',
		"Mercury": '☿',
		"Venus":   '♀',
		"Earth":   '♁',
		"Mars":    '♂',
		"Jupiter": '♃',
		"Saturn":  '♄',
		"Uranus":  '♅',
		"Neptune": '♆',
		"Pluto":   '♇',
	}

	if symbol, exists := knownSymbols[name]; exists {
		return symbol
	}

	return cor.generateGenericSymbol(name)
}

// generateGenericSymbol creates a symbol for unknown celestial bodies
func (cor *CelestialObjectRenderer) generateGenericSymbol(name string) rune {
	genericSymbols := []rune{'●', '◉', '◎', '○', '◯', '⬤', '⚫', '⚪', '🪐', '🌍', '🌎', '🌏', '🌑', '🌒', '🌓', '🌔', '🌕', '🌖', '🌗', '🌘'}

	hash := 0
	for _, char := range name {
		hash = (hash + int(char)) % len(genericSymbols)
	}

	return genericSymbols[hash]
}

// GetOrbitalAngle returns the current orbital angle for a planet (exposed for position calculation)
func (cor *CelestialObjectRenderer) GetOrbitalAngle(planet models.CelestialBody) float64 {
	return cor.getOrbitalAngle(planet)
}

// GetPlanetSize returns the scaled planet size (exposed for click detection)
func (cor *CelestialObjectRenderer) GetPlanetSize(meanRadius float64) int {
	return cor.scalePlanetSize(meanRadius)
}

// UpdateDimensions updates the terminal dimensions for dynamic scaling
func (cor *CelestialObjectRenderer) UpdateDimensions(width, height int) {
	cor.width = width
	cor.height = height
}

// GetSunSize returns the scaled sun size (exposed for click detection)
func (cor *CelestialObjectRenderer) GetSunSize() int {
	return cor.scaleSunSize()
}

// calculateMeanAnomaly calculates the mean anomaly for a planet based on its orbital period
func (cor *CelestialObjectRenderer) calculateMeanAnomaly(planet models.CelestialBody) float64 {
	currentMeanAnomaly := cor.calculateCurrentMeanAnomaly(planet)
	elapsed := time.Since(cor.startTime).Seconds()
	orbitalPeriodSeconds := planet.SideralOrbit * 24 * 3600
	meanMotion := 2 * math.Pi / orbitalPeriodSeconds

	// Scale time for animation purposes (make it much faster for visualization)
	// Each real day = 0.1 seconds in animation (10x faster than before)
	animationSpeedFactor := 864000.0

	animatedMeanAnomaly := currentMeanAnomaly + meanMotion*elapsed*animationSpeedFactor

	return animatedMeanAnomaly
}

// calculateCurrentMeanAnomaly calculates where a planet should be in its orbit today
func (cor *CelestialObjectRenderer) calculateCurrentMeanAnomaly(planet models.CelestialBody) float64 {
	if cor.isOurSolarSystem(planet) {
		return cor.calculateSolarSystemMeanAnomaly(planet)
	}

	// Use generic approach for unknown systems
	// Generate a pseudo-random but deterministic starting position based on planet properties
	// This ensures planets don't all start at the same position

	seed := planet.SemimajorAxis + planet.SideralOrbit + planet.MeanRadius
	initialAngle := math.Mod(seed*0.01745329, 2*math.Pi) // 0.01745329 ≈ π/180
	daysSinceEpoch := time.Since(cor.epochTime).Hours() / 24.0

	if planet.SideralOrbit <= 0 {
		return initialAngle
	}
	meanMotionPerDay := 2 * math.Pi / planet.SideralOrbit

	currentMeanAnomaly := initialAngle + meanMotionPerDay*daysSinceEpoch

	return math.Mod(currentMeanAnomaly, 2*math.Pi)
}

// calculateStarPositions calculates positions for multiple stars around their barycenter
func (cor *CelestialObjectRenderer) calculateStarPositions(stars []models.CelestialBody, centerX, centerY int) []StarPosition {
	if len(stars) <= 1 {
		return []StarPosition{{centerX, centerY}}
	}

	if len(stars) == 2 {
		return cor.calculateBinaryStarPositions(stars, centerX, centerY)
	}

	return cor.calculateMultipleStarPositions(stars, centerX, centerY)
}

// calculateBinaryStarPositions calculates positions for binary star systems
func (cor *CelestialObjectRenderer) calculateBinaryStarPositions(stars []models.CelestialBody, centerX, centerY int) []StarPosition {
	if len(stars) != 2 {
		return []StarPosition{{centerX, centerY}}
	}

	// Calculate masses for proper barycenter
	mass1 := cor.getStarMass(stars[0])
	mass2 := cor.getStarMass(stars[1])
	totalMass := mass1 + mass2

	baseSeparation := cor.calculateBinarySeparation(stars)

	// Calculate distance from barycenter for each star
	r1 := baseSeparation * (mass2 / totalMass)
	r2 := baseSeparation * (mass1 / totalMass)

	elapsed := time.Since(cor.startTime).Seconds()
	orbitalPeriod := cor.calculateBinaryOrbitalPeriod(stars, baseSeparation)
	angle := 2 * math.Pi * elapsed / orbitalPeriod

	x1 := centerX + int(r1*math.Cos(angle))
	y1 := centerY + int(r1*math.Sin(angle)*0.5) // 0.5 for terminal aspect ratio
	x2 := centerX - int(r2*math.Cos(angle))     // Opposite side
	y2 := centerY - int(r2*math.Sin(angle)*0.5)

	return []StarPosition{
		{x1, y1},
		{x2, y2},
	}
}

// calculateMultipleStarPositions calculates positions for systems with 3+ stars
func (cor *CelestialObjectRenderer) calculateMultipleStarPositions(stars []models.CelestialBody, centerX, centerY int) []StarPosition {
	positions := make([]StarPosition, len(stars))

	if len(stars) == 0 {
		return positions
	}

	avgRadius := 0.0
	for _, star := range stars {
		avgRadius += float64(cor.scaleStarSize(star.MeanRadius, len(stars)))
	}
	avgRadius /= float64(len(stars))

	ringRadius := cor.calculateMultiStarRadius(len(stars))

	for i := range stars {
		angle := 2 * math.Pi * float64(i) / float64(len(stars))

		elapsed := time.Since(cor.startTime).Seconds()
		rotationPeriod := cor.calculateMultiStarRotationPeriod(len(stars))
		rotationAngle := 2 * math.Pi * elapsed / rotationPeriod
		angle += rotationAngle

		x := centerX + int(ringRadius*math.Cos(angle))
		y := centerY + int(ringRadius*math.Sin(angle)*0.5) // 0.5 for terminal aspect ratio

		positions[i] = StarPosition{x, y}
	}

	return positions
}

// getStarMass extracts or estimates the mass of a star
func (cor *CelestialObjectRenderer) getStarMass(star models.CelestialBody) float64 {
	if star.Mass.MassValue > 0 {
		return star.Mass.MassValue * math.Pow(10, float64(star.Mass.MassExponent))
	}

	if star.MeanRadius > 0 {
		// Rough approximation: M/M_sun ≈ (R/R_sun)^2.5 for main sequence stars
		sunRadius := 695700.0
		radiusRatio := star.MeanRadius / sunRadius
		return math.Pow(radiusRatio, 2.5) * 1.989e30 // Solar mass in kg
	}

	return 1.989e30
}

// GetBarycenter calculates the barycenter for multi-star systems
func (cor *CelestialObjectRenderer) GetBarycenter(stars []models.CelestialBody, centerX, centerY int) (int, int) {
	if len(stars) <= 1 {
		return centerX, centerY
	}

	// For visualization purposes, we keep the barycenter at screen center
	// Real barycenter calculation would be more complex and might move the center off-screen
	return centerX, centerY
}

// calculateBinarySeparation calculates appropriate separation for binary stars
func (cor *CelestialObjectRenderer) calculateBinarySeparation(stars []models.CelestialBody) float64 {
	terminalSizeFactor := cor.getTerminalSizeFactor()

	// Reduced base separation for tighter orbits for on screen display
	baseSeparation := 6.0 * terminalSizeFactor

	minSeparation := 6.0 // 2 star radii + 2 buffer (reduced)

	return math.Max(baseSeparation, minSeparation)
}

// calculateMultiStarRadius calculates ring radius for 3+ star systems
func (cor *CelestialObjectRenderer) calculateMultiStarRadius(numStars int) float64 {
	terminalSizeFactor := cor.getTerminalSizeFactor()

	baseRadius := 3.0 * terminalSizeFactor

	if numStars > 3 {
		baseRadius *= 1.3 // Reduced from 1.5
	}

	return baseRadius
}

// calculateBinaryOrbitalPeriod calculates realistic orbital period for binary stars
func (cor *CelestialObjectRenderer) calculateBinaryOrbitalPeriod(stars []models.CelestialBody, separation float64) float64 {
	if len(stars) != 2 {
		return 30.0 // fallback
	}

	mass1 := cor.getStarMass(stars[0])
	mass2 := cor.getStarMass(stars[1])
	totalMass := mass1 + mass2

	// Use Kepler's third law: P² ∝ a³/M
	// For visualization, we use a scaled version
	// Real separation would be in AU, but we use screen units

	// Scale mass relative to solar masses for period calculation
	solarMass := 1.989e30
	totalMassRatio := totalMass / solarMass

	// Base period scales with mass (more massive = faster orbit)
	// and separation (wider = slower orbit)
	basePeriod := 20.0 // Base period in seconds for solar-mass binary

	periodScaling := math.Pow(separation/10.0, 1.5) / math.Sqrt(totalMassRatio)

	period := basePeriod * periodScaling

	if period < 10.0 {
		period = 10.0
	} else if period > 120.0 {
		period = 120.0
	}

	return period
}

// calculateMultiStarRotationPeriod calculates rotation period for multi-star systems
func (cor *CelestialObjectRenderer) calculateMultiStarRotationPeriod(numStars int) float64 {
	basePeriod := 30.0

	period := basePeriod * (1.0 + float64(numStars-3)*0.5)

	if period < 20.0 {
		period = 20.0
	} else if period > 90.0 {
		period = 90.0
	}

	return period
}

// getStarSymbol returns appropriate symbol and color for a star based on its type
func (cor *CelestialObjectRenderer) getStarSymbol(star models.CelestialBody) rune {
	// Check if we have stellar classification
	stellarClass := cor.getStellarClass(star)

	// Return symbol based on stellar class
	switch stellarClass[0] {
	case 'O', 'B': // Hot blue/blue-white stars
		return '✦' // Blue star symbol
	case 'A', 'F': // White/yellow-white stars
		return '✧' // White star symbol
	case 'G': // Yellow stars (like Sun)
		return '☉' // Sun symbol
	case 'K': // Orange stars
		return '✩' // Orange star symbol
	case 'M': // Red dwarf stars
		return '✪' // Red star symbol
	default:
		// Unknown or special cases
		if star.EnglishName == "Sun" {
			return '☉'
		}
		return '⭐' // Generic star
	}
}

// getStellarClass extracts stellar classification from star data
func (cor *CelestialObjectRenderer) getStellarClass(star models.CelestialBody) string {
	if stellarClass := cor.getStellarClassField(star); stellarClass != "" {
		return stellarClass
	}

	if temp, ok := cor.getTemperature(star); ok {
		return cor.classifyByTemperature(temp)
	}

	mass := cor.getStarMass(star)
	solarMass := 1.989e30
	massRatio := mass / solarMass

	if massRatio > 16 {
		return "O5V" // Blue supergiant
	} else if massRatio > 2.1 {
		return "B5V" // Blue-white main sequence
	} else if massRatio > 1.4 {
		return "A5V" // White main sequence
	} else if massRatio > 1.04 {
		return "F5V" // Yellow-white main sequence
	} else if massRatio > 0.8 {
		return "G5V" // Yellow main sequence (Sun-like)
	} else if massRatio > 0.45 {
		return "K5V" // Orange main sequence
	} else {
		return "M5V" // Red dwarf
	}
}

// getStellarClassField extracts stellar class field from star data
func (cor *CelestialObjectRenderer) getStellarClassField(star models.CelestialBody) string {
	return star.StellarClass
}

// getTemperature extracts temperature from star data if available
func (cor *CelestialObjectRenderer) getTemperature(star models.CelestialBody) (float64, bool) {
	if star.Temperature > 0 {
		return star.Temperature, true
	}
	return 0, false
}

// classifyByTemperature returns stellar class based on effective temperature
func (cor *CelestialObjectRenderer) classifyByTemperature(temp float64) string {
	if temp >= 30000 {
		return "O5V"
	} else if temp >= 10000 {
		return "B5V"
	} else if temp >= 7500 {
		return "A5V"
	} else if temp >= 6000 {
		return "F5V"
	} else if temp >= 5200 {
		return "G5V"
	} else if temp >= 3700 {
		return "K5V"
	} else {
		return "M5V"
	}
}

// isOurSolarSystem detects if we're working with our Solar System based on planet characteristics
func (cor *CelestialObjectRenderer) isOurSolarSystem(planet models.CelestialBody) bool {
	knownPlanets := map[string]bool{
		"Mercury": true, "Venus": true, "Earth": true, "Mars": true,
		"Jupiter": true, "Saturn": true, "Uranus": true, "Neptune": true, "Pluto": true,
	}

	return knownPlanets[planet.EnglishName]
}

// calculateSolarSystemMeanAnomaly calculates accurate positions for our Solar System
func (cor *CelestialObjectRenderer) calculateSolarSystemMeanAnomaly(planet models.CelestialBody) float64 {
	// Reference epoch: J2000.0 (January 1, 2000, 12:00 TT)
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)

	// Accurate mean anomalies at J2000.0 epoch for our Solar System (in radians)
	j2000MeanAnomalies := map[string]float64{
		"Mercury": 174.7948 * math.Pi / 180.0,
		"Venus":   50.4161 * math.Pi / 180.0,
		"Earth":   357.5291 * math.Pi / 180.0,
		"Mars":    19.3730 * math.Pi / 180.0,
		"Jupiter": 20.0202 * math.Pi / 180.0,
		"Saturn":  317.0207 * math.Pi / 180.0,
		"Uranus":  141.0498 * math.Pi / 180.0,
		"Neptune": 256.2250 * math.Pi / 180.0,
		"Pluto":   14.8820 * math.Pi / 180.0,
	}

	// Get J2000.0 mean anomaly for this planet
	j2000MeanAnomaly, exists := j2000MeanAnomalies[planet.EnglishName]
	if !exists {
		// Fallback to generic calculation if planet not found
		return cor.calculateGenericMeanAnomaly(planet)
	}

	currentDate := time.Now()
	daysSinceJ2000 := currentDate.Sub(j2000).Hours() / 24.0

	if planet.SideralOrbit <= 0 {
		return j2000MeanAnomaly
	}
	meanMotionPerDay := 2 * math.Pi / planet.SideralOrbit

	currentMeanAnomaly := j2000MeanAnomaly + meanMotionPerDay*daysSinceJ2000

	return math.Mod(currentMeanAnomaly, 2*math.Pi)
}

// calculateGenericMeanAnomaly provides fallback calculation for unknown planets
func (cor *CelestialObjectRenderer) calculateGenericMeanAnomaly(planet models.CelestialBody) float64 {
	seed := planet.SemimajorAxis + planet.SideralOrbit + planet.MeanRadius

	initialAngle := math.Mod(seed*0.01745329, 2*math.Pi) // 0.01745329 ≈ π/180

	daysSinceEpoch := time.Since(cor.epochTime).Hours() / 24.0

	if planet.SideralOrbit <= 0 {
		return initialAngle
	}
	meanMotionPerDay := 2 * math.Pi / planet.SideralOrbit

	currentMeanAnomaly := initialAngle + meanMotionPerDay*daysSinceEpoch

	return math.Mod(currentMeanAnomaly, 2*math.Pi)
}
