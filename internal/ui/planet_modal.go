package ui

import (
	"fmt"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/display"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/visualization"
	"github.com/gdamore/tcell/v2"
)

type PlanetModal struct {
	*Modal
	planet   models.CelestialBody
	renderer *visualization.Renderer
}

func NewPlanetModal(screen tcell.Screen, planet models.CelestialBody, renderer *visualization.Renderer) *PlanetModal {
	content := generatePlanetDetailsFromFields(planet, renderer)
	symbol := renderer.GetPlanetSymbol(planet.EnglishName)

	config := ModalConfig{
		Width:    constants.ModalWidth,
		Height:   Minimum(len(content)+6, 30),
		Title:    fmt.Sprintf(" %c %s ", symbol, planet.EnglishName),
		Content:  content,
		Position: constants.TopRight,
	}

	modal := NewModal(screen, config)

	return &PlanetModal{
		Modal:    modal,
		planet:   planet,
		renderer: renderer,
	}
}

func (pm *PlanetModal) Render() {
	pm.Modal.Render()

	instructions := "Press Enter, Escape, or 'b' to close"
	if len(pm.planet.Moons) > 0 {
		instructions += " • 'm' for moons"
	}
	pm.DrawInstructions(instructions)
}

func generatePlanetDetailsFromFields(planet models.CelestialBody, renderer *visualization.Renderer) []string {
	var content []string

	stringFields := display.GetCelestialBodyStringFields()
	for _, field := range stringFields {
		if field.Condition(planet) {
			content = append(content, field.FormatStringFieldValue(planet))
		}
	}

	fields := display.GetCelestialBodyFields()
	for _, field := range fields {
		if field.Condition(planet) {
			content = append(content, field.FormatFieldValue(planet))
		}
	}

	if len(planet.Moons) > 0 {
		moonHandler := renderer.GetMoonHandler()
		moonLines := moonHandler.FormatMoonDisplay(planet, 5)
		for i, line := range moonLines {
			if i == 0 {
				content = append(content, "")
				content = append(content, line)
			} else {
				content = append(content, "  "+line)
			}
		}
	}

	return content
}
