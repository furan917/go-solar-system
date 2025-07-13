package visualization

import "github.com/furan917/go-solar-system/internal/models"

// DebrisBeltRenderer handles rendering of asteroid and Kuiper belts
type DebrisBeltRenderer struct {
	circleDrawer *CircleDrawer
	scaler       *DistanceScaler
}

// NewDebrisBeltRenderer creates a new debris belt renderer
func NewDebrisBeltRenderer(circleDrawer *CircleDrawer, scaler *DistanceScaler) *DebrisBeltRenderer {
	return &DebrisBeltRenderer{
		circleDrawer: circleDrawer,
		scaler:       scaler,
	}
}

// RenderAsteroidBelt renders the asteroid belt between Mars and Jupiter
func (dbr *DebrisBeltRenderer) RenderAsteroidBelt(grid [][]rune, centerX, centerY int, planets []models.CelestialBody) {
	marsDistance, jupiterDistance := dbr.findPlanetDistances(planets, "Mars", "Jupiter")

	innerRadius := dbr.scaler.ScaleDistance(marsDistance*1.5, planets)
	outerRadius := dbr.scaler.ScaleDistance(jupiterDistance*0.6, planets)

	dbr.renderDebrisBelt(grid, centerX, centerY, innerRadius, outerRadius, 10, 3, '∗')
}

// RenderKuiperBelt renders the Kuiper belt beyond Neptune
func (dbr *DebrisBeltRenderer) RenderKuiperBelt(grid [][]rune, centerX, centerY int, planets []models.CelestialBody) {
	neptuneDistance := dbr.findPlanetDistance(planets, "Neptune")

	innerRadius := dbr.scaler.ScaleDistance(neptuneDistance*1.2, planets)
	outerRadius := dbr.scaler.ScaleDistance(neptuneDistance*1.7, planets)

	dbr.renderDebrisBelt(grid, centerX, centerY, innerRadius, outerRadius, 12, 4, '◦')
}

// findPlanetDistances finds distances for two planets
func (dbr *DebrisBeltRenderer) findPlanetDistances(planets []models.CelestialBody, planet1, planet2 string) (float64, float64) {
	var dist1, dist2 float64
	defaults := map[string]float64{
		"Mars":    1.5,
		"Jupiter": 5.2,
		"Neptune": 30.0,
	}

	for _, planet := range planets {
		if planet.EnglishName == planet1 && planet.EnglishName != "Sun" {
			dist1 = planet.SemimajorAxis
		}
		if planet.EnglishName == planet2 && planet.EnglishName != "Sun" {
			dist2 = planet.SemimajorAxis
		}
	}

	if dist1 == 0 {
		dist1 = defaults[planet1]
	}
	if dist2 == 0 {
		dist2 = defaults[planet2]
	}

	return dist1, dist2
}

// findPlanetDistance finds distance for a single planet
func (dbr *DebrisBeltRenderer) findPlanetDistance(planets []models.CelestialBody, planetName string) float64 {
	for _, planet := range planets {
		if planet.EnglishName == planetName && planet.EnglishName != "Sun" {
			return planet.SemimajorAxis
		}
	}

	defaults := map[string]float64{
		"Neptune": 30.0,
	}

	return defaults[planetName]
}

// renderDebrisBelt renders a debris belt with specified parameters
func (dbr *DebrisBeltRenderer) renderDebrisBelt(grid [][]rune, centerX, centerY int, innerRadius, outerRadius float64, angleStep, rings int, symbol rune) {
	for angle := 0; angle < 360; angle += angleStep {
		radians := float64(angle) * 3.14159 / 180

		for i := 0; i < rings; i++ {
			radius := innerRadius + float64(i)*(outerRadius-innerRadius)/float64(rings)
			x, y := dbr.circleDrawer.CalculatePosition(centerX, centerY, radius, radians)

			if dbr.circleDrawer.isInBounds(x, y, len(grid[0]), len(grid)) && grid[y][x] == ' ' {
				grid[y][x] = symbol
			}
		}
	}
}
