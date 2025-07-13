package app

import (
	"sync"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/visualization"
)

// AppState manages all application state for the solar system application
// Critical fields use thread-safe access, others maintain backward compatibility
type AppState struct {
	// Protects critical concurrent access
	mu sync.RWMutex

	// Core data - centralized to avoid scattered state
	Planets             []models.CelestialBody
	PlanetPositions     map[string]visualization.PlanetPosition
	PlanetListPositions []PlanetListPosition
	CurrentSystem       string

	// Navigation state
	SelectedIndex  int
	SelectedPlanet models.CelestialBody
	SelectedMoon   models.CelestialBody

	// UI visibility state
	ShowingDetails     bool
	ShowingMoons       bool
	ShowingMoonDetails bool
	ShowingSystemList  bool

	// Scroll state for lists
	MoonScrollIndex     int
	MoonSelectedIndex   int
	SystemScrollIndex   int
	SystemSelectedIndex int

	// Application control - CRITICAL: Use thread-safe access only
	running bool
}

// PlanetListPosition represents a clickable planet position in the UI
type PlanetListPosition struct {
	Index int
	X     int
	Y     int
	Width int
}

// NewAppState creates a new application state with default values
func NewAppState() *AppState {
	return &AppState{
		Planets:             make([]models.CelestialBody, 0),
		PlanetPositions:     make(map[string]visualization.PlanetPosition),
		PlanetListPositions: make([]PlanetListPosition, 0),
		CurrentSystem:       "solar-system",
		SelectedIndex:       0,
		MoonScrollIndex:     0,
		MoonSelectedIndex:   0,
		SystemScrollIndex:   0,
		SystemSelectedIndex: 0,
		running:             true,
		ShowingDetails:      false,
		ShowingMoons:        false,
		ShowingMoonDetails:  false,
		ShowingSystemList:   false,
	}
}

// ResetModals closes all modal windows
func (s *AppState) ResetModals() {
	s.ShowingDetails = false
	s.ShowingMoons = false
	s.ShowingMoonDetails = false
	s.ShowingSystemList = false
}

// IsAnyModalShowing returns true if any modal is currently visible
func (s *AppState) IsAnyModalShowing() bool {
	return s.ShowingDetails || s.ShowingMoons || s.ShowingMoonDetails || s.ShowingSystemList
}

// ShowPlanetDetails opens the planet details modal
func (s *AppState) ShowPlanetDetails(planet models.CelestialBody, index int) {
	s.ResetModals()
	s.SelectedPlanet = planet
	s.SelectedIndex = index
	s.ShowingDetails = true
}

// ShowMoonList opens the moon list modal
func (s *AppState) ShowMoonList() {
	s.ResetModals()
	s.ShowingMoons = true
	s.MoonScrollIndex = 0
	s.MoonSelectedIndex = 0
}

// ShowMoonDetails opens the moon details modal
func (s *AppState) ShowMoonDetails(moon models.CelestialBody) {
	s.ResetModals()
	s.SelectedMoon = moon
	s.ShowingMoonDetails = true
}

// ShowSystemList opens the system selection modal
func (s *AppState) ShowSystemList() {
	s.ResetModals()
	s.ShowingSystemList = true
}

// HandleMoonNavigation updates moon navigation state
func (s *AppState) HandleMoonNavigation(direction int, moonCount int) {
	switch direction {
	case -1: // Up
		if s.MoonSelectedIndex > 0 {
			s.MoonSelectedIndex--
			if s.MoonSelectedIndex < s.MoonScrollIndex {
				s.MoonScrollIndex = s.MoonSelectedIndex
			}
		}
	case 1: // Down
		if s.MoonSelectedIndex < moonCount-1 {
			s.MoonSelectedIndex++
			if s.MoonSelectedIndex >= s.MoonScrollIndex+constants.MaxVisibleItems {
				s.MoonScrollIndex = s.MoonSelectedIndex - constants.MaxVisibleItems + 1
			}
		}
	}
}

// HandleSystemNavigation updates system navigation state
func (s *AppState) HandleSystemNavigation(direction int, systemCount int) {
	switch direction {
	case -1: // Up
		if s.SystemSelectedIndex > 0 {
			s.SystemSelectedIndex--
			if s.SystemSelectedIndex < s.SystemScrollIndex {
				s.SystemScrollIndex = s.SystemSelectedIndex
			}
		}
	case 1: // Down
		if s.SystemSelectedIndex < systemCount-1 {
			s.SystemSelectedIndex++
			if s.SystemSelectedIndex >= s.SystemScrollIndex+constants.MaxVisibleItems {
				s.SystemScrollIndex = s.SystemSelectedIndex - constants.MaxVisibleItems + 1
			}
		}
	}
}

// UpdatePlanetSelection updates the currently selected planet
func (s *AppState) UpdatePlanetSelection(index int, planet models.CelestialBody) {
	s.SelectedIndex = index
	s.SelectedPlanet = planet
}

// Thread-safe accessors for critical concurrent fields

func (s *AppState) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *AppState) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = running
}

// Convenience getters for interface compliance (not thread-safe - only use from main thread)

func (s *AppState) GetSelectedIndex() int {
	return s.SelectedIndex
}

func (s *AppState) GetSelectedPlanet() models.CelestialBody {
	return s.SelectedPlanet
}

func (s *AppState) GetSelectedMoon() models.CelestialBody {
	return s.SelectedMoon
}

func (s *AppState) IsShowingDetails() bool {
	return s.ShowingDetails
}

func (s *AppState) IsShowingMoons() bool {
	return s.ShowingMoons
}

func (s *AppState) IsShowingMoonDetails() bool {
	return s.ShowingMoonDetails
}

func (s *AppState) IsShowingSystemList() bool {
	return s.ShowingSystemList
}

// Data accessors for centralized state

func (s *AppState) GetPlanets() []models.CelestialBody {
	return s.Planets
}

func (s *AppState) SetPlanets(planets []models.CelestialBody) {
	s.Planets = planets
}

func (s *AppState) GetPlanetPositions() map[string]visualization.PlanetPosition {
	return s.PlanetPositions
}

func (s *AppState) SetPlanetPositions(positions map[string]visualization.PlanetPosition) {
	s.PlanetPositions = positions
}

func (s *AppState) GetPlanetListPositions() []PlanetListPosition {
	return s.PlanetListPositions
}

func (s *AppState) SetPlanetListPositions(positions []PlanetListPosition) {
	s.PlanetListPositions = positions
}

func (s *AppState) GetCurrentSystem() string {
	return s.CurrentSystem
}

func (s *AppState) SetCurrentSystem(system string) {
	s.CurrentSystem = system
}

// Data manipulation methods for better encapsulation

func (s *AppState) ClearPlanetListPositions() {
	s.PlanetListPositions = s.PlanetListPositions[:0]
}

func (s *AppState) AddPlanetListPosition(pos PlanetListPosition) {
	s.PlanetListPositions = append(s.PlanetListPositions, pos)
}

func (s *AppState) UpdatePlanetPositions(x, y int, positions map[string]visualization.PlanetPosition) {
	s.PlanetPositions = make(map[string]visualization.PlanetPosition)
	for name, pos := range positions {
		adjustedPos := pos
		adjustedPos.X += x
		adjustedPos.Y += y
		s.PlanetPositions[name] = adjustedPos
	}
}

// Data consistency and validation methods

func (s *AppState) ValidateState() error {
	if s.SelectedIndex < 0 || s.SelectedIndex >= len(s.Planets) {
		s.SelectedIndex = 0
	}

	if s.MoonSelectedIndex < 0 {
		s.MoonSelectedIndex = 0
	}

	if s.MoonScrollIndex < 0 {
		s.MoonScrollIndex = 0
	}

	if s.SystemSelectedIndex < 0 {
		s.SystemSelectedIndex = 0
	}

	if s.SystemScrollIndex < 0 {
		s.SystemScrollIndex = 0
	}

	return nil
}

// Thread-safe planet access with bounds checking
func (s *AppState) GetPlanetSafely(index int) (models.CelestialBody, bool) {
	if index < 0 || index >= len(s.Planets) {
		return models.CelestialBody{}, false
	}
	return s.Planets[index], true
}
