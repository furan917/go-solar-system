// Package interfaces defines the core interfaces for the solar system application.
// These interfaces improve testability, modularity, and enable dependency injection.
package interfaces

import (
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/gdamore/tcell/v2"
)

// APIClient defines the interface for fetching celestial body data
type APIClient interface {
	GetAllBodies() ([]models.CelestialBody, error)
	GetBody(id string) (*models.CelestialBody, error)
}

// Renderer defines the interface for solar system visualization
type Renderer interface {
	RenderSolarSystemData(planets []models.CelestialBody, width, height int) [][]rune
	UpdateDimensions(width, height int)
}

// StateManager defines the interface for application state management
type StateManager interface {
	// Navigation state
	GetSelectedIndex() int
	GetSelectedPlanet() models.CelestialBody
	GetSelectedMoon() models.CelestialBody
	UpdatePlanetSelection(index int, planet models.CelestialBody)

	// UI visibility state
	IsShowingDetails() bool
	IsShowingMoons() bool
	IsShowingMoonDetails() bool
	IsShowingSystemList() bool
	IsAnyModalShowing() bool

	// Modal management
	ShowPlanetDetails(planet models.CelestialBody, index int)
	ShowMoonList()
	ShowMoonDetails(moon models.CelestialBody)
	ShowSystemList()
	ResetModals()

	// Navigation
	HandleMoonNavigation(direction int, moonCount int)
	HandleSystemNavigation(direction int, systemCount int)

	// Application control
	IsRunning() bool
	SetRunning(running bool)
}

// SystemManager defines the interface for star system management
type SystemManager interface {
	GetCurrentSystemDisplayName() string
	ScanSystems() error
}

// Screen wraps tcell.Screen for easier testing
type Screen interface {
	Init() error
	Fini()
	Clear()
	Show()
	Size() (int, int)
	PollEvent() tcell.Event
	SetContent(x, y int, mainc rune, combc []rune, style tcell.Style)
	Sync()
}

// CircleDrawer defines the interface for drawing circular shapes
type CircleDrawer interface {
	DrawCircle(grid [][]rune, centerX, centerY int, radius float64, symbol rune)
}

// DistanceScaler defines the interface for scaling astronomical distances
type DistanceScaler interface {
	ScaleDistance(actualDistance float64, maxDisplayRadius float64) float64
}

// MoonHandler defines the interface for moon data management
type MoonHandler interface {
	ResolveMoonNames(moons []models.Moon) []models.Moon
}
