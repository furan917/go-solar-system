package visualization

import (
	"math"

	"github.com/furan917/go-solar-system/internal/models"
)

// DistanceScaler handles scaling of astronomical distances to screen coordinates
type DistanceScaler struct {
	width  int
	height int
}

// NewDistanceScaler creates a new distance scaler
func NewDistanceScaler(width, height int) *DistanceScaler {
	return &DistanceScaler{
		width:  width,
		height: height,
	}
}

// ScaleDistance scales an astronomical distance to fit the display
func (ds *DistanceScaler) ScaleDistance(distance float64, planets []models.CelestialBody) float64 {
	if distance <= 0 {
		return 0
	}

	minDistance, maxDistance := ds.findDistanceRange(planets)

	logMin := math.Log(minDistance)
	logMax := math.Log(maxDistance)
	logCurrent := math.Log(distance)

	normalized := (logCurrent - logMin) / (logMax - logMin)

	minRadius := 8.0
	maxRadius := math.Min(float64(ds.width/2-3), float64(ds.height/2-3)) * 0.95

	return minRadius + normalized*(maxRadius-minRadius)
}

// findDistanceRange finds the minimum and maximum distances among planets (excluding Sun)
func (ds *DistanceScaler) findDistanceRange(planets []models.CelestialBody) (float64, float64) {
	if len(planets) == 0 {
		return 1.0, 100.0
	}

	var minDistance, maxDistance float64
	first := true

	for _, planet := range planets {
		if planet.EnglishName == "Sun" || planet.SemimajorAxis <= 0 {
			continue
		}

		if first {
			minDistance = planet.SemimajorAxis
			maxDistance = planet.SemimajorAxis
			first = false
		} else {
			if planet.SemimajorAxis < minDistance {
				minDistance = planet.SemimajorAxis
			}
			if planet.SemimajorAxis > maxDistance {
				maxDistance = planet.SemimajorAxis
			}
		}
	}

	if first {
		return 1.0, 100.0
	}

	return minDistance, maxDistance
}
