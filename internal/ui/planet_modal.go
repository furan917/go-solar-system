package ui

import (
	"fmt"
	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/gdamore/tcell/v2"
)

type PlanetModal struct {
	*Modal
	planet models.CelestialBody
}

func NewPlanetModal(screen tcell.Screen, planet models.CelestialBody) *PlanetModal {
	content := generatePlanetDetails(planet)

	config := ModalConfig{
		Width:    60,
		Height:   minimum(len(content)+6, 30),
		Title:    fmt.Sprintf(" %s Details ", planet.EnglishName),
		Content:  content,
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &PlanetModal{
		Modal:  modal,
		planet: planet,
	}
}

func (pm *PlanetModal) Render() {
	pm.Modal.Render()

	instructions := "Escape/'b' to close"
	if len(pm.planet.Moons) > 0 {
		instructions += " • 'm' for moons"
	}
	pm.DrawInstructions(instructions)
}

func generatePlanetDetails(planet models.CelestialBody) []string {
	var content []string

	content = append(content, fmt.Sprintf("Type: %s", planet.BodyType))

	if planet.MeanRadius > 0 {
		content = append(content, fmt.Sprintf("Radius: %.0f km", planet.MeanRadius))
	}

	if planet.Mass.MassValue > 0 {
		content = append(content, fmt.Sprintf("Mass: %.2f × 10^%d kg", planet.Mass.MassValue, planet.Mass.MassExponent))
	}

	if planet.Density > 0 {
		content = append(content, fmt.Sprintf("Density: %.2f g/cm³", planet.Density))
	}

	if planet.Gravity > 0 {
		content = append(content, fmt.Sprintf("Gravity: %.1f m/s²", planet.Gravity))
	}

	if planet.SemimajorAxis > 0 {
		content = append(content, fmt.Sprintf("Distance: %.0f km", planet.SemimajorAxis))
	}

	if planet.SideralOrbit > 0 {
		content = append(content, fmt.Sprintf("Orbital Period: %.2f days", planet.SideralOrbit))
	}

	if planet.Eccentricity > 0 {
		content = append(content, fmt.Sprintf("Eccentricity: %.3f", planet.Eccentricity))
	}

	if planet.Temperature > 0 {
		content = append(content, fmt.Sprintf("Temperature: %.0f K", planet.Temperature))
	}

	if planet.StellarClass != "" {
		content = append(content, fmt.Sprintf("Stellar Class: %s", planet.StellarClass))
	}

	if planet.DiscoveredBy != "" {
		content = append(content, fmt.Sprintf("Discovered by: %s", planet.DiscoveredBy))
	}

	if planet.DiscoveryDate != "" {
		content = append(content, fmt.Sprintf("Discovery Date: %s", planet.DiscoveryDate))
	}

	if len(planet.Moons) > 0 {
		content = append(content, fmt.Sprintf("Moons: %d", len(planet.Moons)))
	}

	return content
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
