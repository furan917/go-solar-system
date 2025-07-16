package orbital

import (
	"math"
	"time"

	"github.com/furan917/go-solar-system/internal/models"
)

// Calculator interface defines orbital position calculation methods
type Calculator interface {
	CalculateMeanAnomaly(body models.CelestialBody, currentTime time.Time) float64
	GetSystemType() SystemType
}

// SystemType represents the type of orbital calculation system
type SystemType string

const (
	SystemTypeSolar   SystemType = "solar"
	SystemTypeGeneric SystemType = "generic"
	SystemTypeExact   SystemType = "exact"
)

// SolarSystemCalculator handles our Solar System with J2000.0 epoch data
type SolarSystemCalculator struct {
	epochTime time.Time
}

// GenericCalculator handles external systems with pseudo-random positioning
type GenericCalculator struct {
	epochTime time.Time
}

// ExactCalculator handles systems with precise orbital element data
type ExactCalculator struct{}

// NewSolarSystemCalculator creates calculator for our Solar System
func NewSolarSystemCalculator(epochTime time.Time) *SolarSystemCalculator {
	return &SolarSystemCalculator{epochTime: epochTime}
}

// NewGenericCalculator creates calculator for external systems without orbital data
func NewGenericCalculator(epochTime time.Time) *GenericCalculator {
	return &GenericCalculator{epochTime: epochTime}
}

// NewExactCalculator creates calculator for systems with precise orbital elements
func NewExactCalculator() *ExactCalculator {
	return &ExactCalculator{}
}

// CalculateMeanAnomaly for Solar System using J2000.0 epoch positions
func (sc *SolarSystemCalculator) CalculateMeanAnomaly(body models.CelestialBody, currentTime time.Time) float64 {
	// J2000.0 epoch: January 1, 2000, 12:00 TT
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)

	// Accurate mean anomalies at J2000.0 epoch (in radians)
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

	j2000MeanAnomaly, exists := j2000MeanAnomalies[body.EnglishName]
	if !exists {
		generic := NewGenericCalculator(sc.epochTime)
		return generic.CalculateMeanAnomaly(body, currentTime)
	}

	daysSinceJ2000 := currentTime.Sub(j2000).Hours() / 24.0

	if body.SideralOrbit <= 0 {
		return j2000MeanAnomaly
	}

	meanMotionPerDay := 2 * math.Pi / body.SideralOrbit
	currentMeanAnomaly := j2000MeanAnomaly + meanMotionPerDay*daysSinceJ2000

	return math.Mod(currentMeanAnomaly, 2*math.Pi)
}

func (sc *SolarSystemCalculator) GetSystemType() SystemType {
	return SystemTypeSolar
}

// CalculateMeanAnomaly for generic systems using deterministic pseudo-random positioning
func (gc *GenericCalculator) CalculateMeanAnomaly(body models.CelestialBody, currentTime time.Time) float64 {
	// Generate deterministic starting position based on body properties
	seed := body.SemimajorAxis + body.SideralOrbit + body.MeanRadius
	initialAngle := math.Mod(seed*0.01745329, 2*math.Pi) // 0.01745329 ≈ π/180

	daysSinceEpoch := currentTime.Sub(gc.epochTime).Hours() / 24.0

	if body.SideralOrbit <= 0 {
		return initialAngle
	}

	meanMotionPerDay := 2 * math.Pi / body.SideralOrbit
	currentMeanAnomaly := initialAngle + meanMotionPerDay*daysSinceEpoch

	return math.Mod(currentMeanAnomaly, 2*math.Pi)
}

func (gc *GenericCalculator) GetSystemType() SystemType {
	return SystemTypeGeneric
}

// CalculateMeanAnomaly for systems with precise orbital element data
func (ec *ExactCalculator) CalculateMeanAnomaly(body models.CelestialBody, currentTime time.Time) float64 {
	if body.OrbitalElements == nil {
		return 0
	}

	epochTime := body.OrbitalElements.Epoch
	daysSinceEpoch := currentTime.Sub(epochTime).Hours() / 24.0

	if body.SideralOrbit <= 0 {
		return math.Mod(body.OrbitalElements.MeanAnomaly*math.Pi/180.0, 2*math.Pi)
	}

	meanMotionPerDay := 2 * math.Pi / body.SideralOrbit
	epochMeanAnomaly := body.OrbitalElements.MeanAnomaly * math.Pi / 180.0 // Convert to radians
	currentMeanAnomaly := epochMeanAnomaly + meanMotionPerDay*daysSinceEpoch

	return math.Mod(currentMeanAnomaly, 2*math.Pi)
}

func (ec *ExactCalculator) GetSystemType() SystemType {
	return SystemTypeExact
}

// CalculatorFactory creates appropriate calculator based on body and system characteristics
type CalculatorFactory struct{}

func NewCalculatorFactory() *CalculatorFactory {
	return &CalculatorFactory{}
}

// CreateCalculator returns appropriate calculator for the given body and system
func (cf *CalculatorFactory) CreateCalculator(body models.CelestialBody, epochTime time.Time) Calculator {
	if cf.isSolarSystemBody(body) {
		return NewSolarSystemCalculator(epochTime)
	}

	if body.OrbitalElements != nil {
		return NewExactCalculator()
	}

	return NewGenericCalculator(epochTime)
}

// isSolarSystemBody detects if body belongs to our Solar System
func (cf *CalculatorFactory) isSolarSystemBody(body models.CelestialBody) bool {
	knownSolarSystemBodies := map[string]bool{
		"Mercury": true, "Venus": true, "Earth": true, "Mars": true,
		"Jupiter": true, "Saturn": true, "Uranus": true, "Neptune": true, "Pluto": true,
	}
	return knownSolarSystemBodies[body.EnglishName]
}
