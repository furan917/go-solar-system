package app

import (
	"fmt"
	"sort"

	"github.com/furan917/go-solar-system/internal/api"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/systems"
)

// PlanetService handles business logic for celestial body operations
type PlanetService struct {
	client        *api.Client
	systemManager *systems.SystemManager
}

// NewPlanetService creates a new planet service with necessary dependencies
func NewPlanetService(client *api.Client, systemManager *systems.SystemManager) *PlanetService {
	return &PlanetService{
		client:        client,
		systemManager: systemManager,
	}
}

// LoadCurrentSystem loads celestial bodies for the current system
func (ps *PlanetService) LoadCurrentSystem() ([]models.CelestialBody, error) {
	currentSystem := ps.systemManager.GetCurrentSystem()

	if currentSystem == "solar-system" {
		return ps.loadSolarSystem()
	}

	return ps.loadExternalSystem(currentSystem)
}

// loadSolarSystem loads our solar system from the API
func (ps *PlanetService) loadSolarSystem() ([]models.CelestialBody, error) {
	bodies, err := ps.client.GetAllBodies()
	if err != nil {
		return nil, fmt.Errorf("failed to load solar system: %w", err)
	}

	var planets []models.CelestialBody
	for _, body := range bodies {
		if body.IsPlanet {
			planets = append(planets, body)
		}
	}

	sort.Slice(planets, func(i, j int) bool {
		return planets[i].SemimajorAxis < planets[j].SemimajorAxis
	})

	return planets, nil
}

// loadExternalSystem loads an external star system from JSON files
func (ps *PlanetService) loadExternalSystem(systemName string) ([]models.CelestialBody, error) {
	systemData, err := ps.systemManager.LoadSystem(systemName)
	if err != nil {
		return nil, fmt.Errorf("failed to load external system %s: %w", systemName, err)
	}

	planets := systemData.Bodies
	sort.Slice(planets, func(i, j int) bool {
		return planets[i].SemimajorAxis < planets[j].SemimajorAxis
	})

	return planets, nil
}

// SwitchToSystem changes the current system and loads its data
func (ps *PlanetService) SwitchToSystem(systemName string) ([]models.CelestialBody, error) {
	if err := ps.systemManager.SwitchToSystem(systemName); err != nil {
		return nil, fmt.Errorf("failed to switch to system %s: %w", systemName, err)
	}

	return ps.LoadCurrentSystem()
}

// GetMoonData attempts to fetch detailed moon data
func (ps *PlanetService) GetMoonData(moonID string) (*models.CelestialBody, error) {
	return ps.client.GetMoonData(moonID)
}

// ValidatePlanetData performs basic validation on planet data
func (ps *PlanetService) ValidatePlanetData(planets []models.CelestialBody) error {
	if len(planets) == 0 {
		return fmt.Errorf("no planets loaded")
	}

	for i, planet := range planets {
		if planet.EnglishName == "" {
			return fmt.Errorf("planet at index %d has no name", i)
		}

		if planet.MeanRadius < 0 {
			return fmt.Errorf("planet %s has invalid radius: %.2f", planet.EnglishName, planet.MeanRadius)
		}
	}

	return nil
}

// GetClient returns the API client
func (ps *PlanetService) GetClient() *api.Client {
	return ps.client
}
