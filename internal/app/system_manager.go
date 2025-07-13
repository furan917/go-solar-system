package app

import (
	"fmt"
	"sort"

	"github.com/furan917/go-solar-system/internal/models"
)

type SystemManager struct {
	state         *AppState
	planetService *PlanetService
	uiRenderer    *UIRenderer
	errorHandler  *ErrorHandler
	logger        interface{}
}

func NewSystemManager(state *AppState, planetService *PlanetService, uiRenderer *UIRenderer, errorHandler *ErrorHandler, logger interface{}) *SystemManager {
	return &SystemManager{
		state:         state,
		planetService: planetService,
		uiRenderer:    uiRenderer,
		errorHandler:  errorHandler,
		logger:        logger,
	}
}

func (sm *SystemManager) LoadCurrentSystem() error {
	defer func() {
		if r := recover(); r != nil {
			if logger, ok := sm.logger.(interface{ Printf(string, ...interface{}) }); ok {
				logger.Printf("Panic in loadCurrentSystem: %v", r)
			}
		}
	}()

	currentSystem := sm.uiRenderer.GetSystemManager().GetCurrentSystem()

	if currentSystem == "solar-system" {
		planets, err := sm.planetService.GetClient().GetPlanets()
		if err != nil {
			return NewAPIError("failed to load Solar System from API", err).
				WithContext("system", currentSystem)
		}

		if len(planets) == 0 {
			return NewValidationError("no planets received from API", nil).
				WithContext("system", currentSystem)
		}

		sm.state.SetPlanets(planets)
	} else {
		systemData, err := sm.uiRenderer.GetSystemManager().GetSystemData()
		if err != nil {
			return NewFileError("failed to load external system", err).
				WithContext("system", currentSystem)
		}

		if len(systemData.Bodies) == 0 {
			return NewValidationError("no celestial bodies in system file", nil).
				WithContext("system", currentSystem)
		}

		sm.state.SetPlanets(systemData.Bodies)
	}

	return nil
}

func (sm *SystemManager) SortPlanetsByDistance() error {
	defer func() {
		if r := recover(); r != nil {
			if logger, ok := sm.logger.(interface{ Printf(string, ...interface{}) }); ok {
				logger.Printf("Panic in sortPlanetsByDistance: %v", r)
			}
		}
	}()

	planets := sm.state.GetPlanets()
	sort.Slice(planets, func(i, j int) bool {
		return planets[i].SemimajorAxis < planets[j].SemimajorAxis
	})
	sm.state.SetPlanets(planets)
	return nil
}

func (sm *SystemManager) NormalizePlanetNames(planets []models.CelestialBody) []models.CelestialBody {
	if !sm.isOurSolarSystem(planets) {
		return planets
	}

	normalized := make([]models.CelestialBody, len(planets))
	copy(normalized, planets)

	for i, planet := range normalized {
		if sm.isSunBody(planet) {
			normalized[i].EnglishName = "Sun"
			normalized[i].Name = "Sun"
			normalized[i].BodyType = "Star"
		}
	}

	return normalized
}

func (sm *SystemManager) FindOrCreateCentralStar(planets []models.CelestialBody) models.CelestialBody {
	for _, planet := range planets {
		if planet.SemimajorAxis == 0 || planet.BodyType == "Star" || sm.isSunBody(planet) {
			return planet
		}
	}

	largestRadius := 0.0
	for _, planet := range planets {
		if planet.MeanRadius > largestRadius {
			largestRadius = planet.MeanRadius
		}
	}

	centralStarRadius := largestRadius * 10
	if centralStarRadius < 100000 {
		centralStarRadius = 695700
	}

	starName := "Central Star"
	starID := "central-star"

	if sm.isOurSolarSystem(planets) {
		starName = "Sun"
		starID = "sun"
		centralStarRadius = 695700
	}

	return models.CelestialBody{
		ID:          starID,
		Name:        starName,
		EnglishName: starName,
		IsPlanet:    false,
		BodyType:    "Star",
		MeanRadius:  centralStarRadius,
		Mass: models.Mass{
			MassValue:    1.9891,
			MassExponent: 30,
		},
		Density:         1.408,
		Gravity:         274.0,
		SemimajorAxis:   0,
		SideralRotation: 609.12,
		DiscoveredBy:    "Ancient",
		DiscoveryDate:   "Prehistoric",
		Moons:           []models.Moon{},
	}
}

func (sm *SystemManager) ContainsCentralStar(planets []models.CelestialBody) bool {
	for _, planet := range planets {
		if planet.SemimajorAxis == 0 || planet.BodyType == "Star" {
			return true
		}
	}
	return false
}

func (sm *SystemManager) SwitchToSelectedSystem() {
	defer func() {
		if r := recover(); r != nil {
			if logger, ok := sm.logger.(interface{ Printf(string, ...interface{}) }); ok {
				logger.Printf("Panic in switchToSelectedSystem: %v", r)
			}
			sm.errorHandler.HandleError(NewSystemError("panic during system switch", fmt.Errorf("%v", r)))
		}
	}()

	availableSystems := sm.uiRenderer.GetSystemManager().GetAvailableSystems()
	if sm.state.SystemSelectedIndex >= len(availableSystems) {
		sm.errorHandler.HandleError(NewValidationError("invalid system index", nil).
			WithContext("index", sm.state.SystemSelectedIndex).
			WithContext("available", len(availableSystems)))
		return
	}

	selectedSystem := availableSystems[sm.state.SystemSelectedIndex]

	if err := sm.uiRenderer.GetSystemManager().SwitchToSystem(selectedSystem); err != nil {
		sm.errorHandler.HandleError(NewSystemError("failed to switch system", err).
			WithContext("target_system", selectedSystem))
		return
	}

	if err := sm.LoadCurrentSystem(); err != nil {
		sm.errorHandler.HandleError(NewSystemError("failed to reload system data after switch", err).
			WithContext("target_system", selectedSystem))
		return
	}

	if err := sm.SortPlanetsByDistance(); err != nil {
		sm.errorHandler.HandleError(NewStateError("failed to sort planets after system switch", err))
	}

	sm.state.SetPlanets(sm.NormalizePlanetNames(sm.state.GetPlanets()))
	centralStar := sm.FindOrCreateCentralStar(sm.state.GetPlanets())

	if !sm.ContainsCentralStar(sm.state.GetPlanets()) {
		sm.state.SetPlanets(append([]models.CelestialBody{centralStar}, sm.state.GetPlanets()...))
	}

	sm.state.SelectedIndex = 0
	sm.state.ShowingSystemList = false
}

func (sm *SystemManager) isOurSolarSystem(planets []models.CelestialBody) bool {
	knownPlanets := map[string]bool{
		"Mercury": false, "Venus": false, "Earth": false, "Mars": false,
		"Jupiter": false, "Saturn": false, "Uranus": false, "Neptune": false, "Pluto": false,
	}

	matchCount := 0
	for _, planet := range planets {
		if _, exists := knownPlanets[planet.EnglishName]; exists {
			matchCount++
		}
	}

	return matchCount >= 4
}

func (sm *SystemManager) isSunBody(body models.CelestialBody) bool {
	sunNames := map[string]bool{
		"Sun": true, "Sol": true, "Soleil": true, "Sole": true, "Sonne": true,
		"sun": true, "sol": true, "soleil": true, "sole": true, "sonne": true,
	}

	if sunNames[body.EnglishName] || sunNames[body.Name] {
		return true
	}

	if body.SemimajorAxis == 0 && body.MeanRadius > 600000 {
		return true
	}

	if body.BodyType == "Star" {
		return true
	}

	return false
}
