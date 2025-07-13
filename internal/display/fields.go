// Package display provides shared configuration and utilities for displaying
// celestial body information across different parts of the application.
package display

import (
	"fmt"

	"github.com/furan917/go-solar-system/internal/models"
)

// FieldConfig defines how to display a specific field of a celestial body
type FieldConfig struct {
	Label     string
	Format    string
	Unit      string
	Condition func(models.CelestialBody) bool
	Value     func(models.CelestialBody) interface{}
}

// StringFieldConfig defines how to display string fields of a celestial body
type StringFieldConfig struct {
	Label     string
	Condition func(models.CelestialBody) bool
	Value     func(models.CelestialBody) string
}

// GetCelestialBodyFields returns the standardized field configurations
// for displaying celestial body numeric data across the application
func GetCelestialBodyFields() []FieldConfig {
	return []FieldConfig{
		{
			Label:     "Mean Radius",
			Format:    "%.0f",
			Unit:      "km",
			Condition: func(cb models.CelestialBody) bool { return cb.MeanRadius > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.MeanRadius },
		},
		{
			Label:     "Mass",
			Format:    "%.2e",
			Unit:      "kg",
			Condition: func(cb models.CelestialBody) bool { return cb.GetMassKg() > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.GetMassKg() },
		},
		{
			Label:     "Density",
			Format:    "%.2f",
			Unit:      "g/cm³",
			Condition: func(cb models.CelestialBody) bool { return cb.Density > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Density },
		},
		{
			Label:     "Gravity",
			Format:    "%.2f",
			Unit:      "m/s²",
			Condition: func(cb models.CelestialBody) bool { return cb.Gravity > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Gravity },
		},
		{
			Label:     "Distance from Sun",
			Format:    "%.0f",
			Unit:      "km",
			Condition: func(cb models.CelestialBody) bool { return cb.SemimajorAxis > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.SemimajorAxis },
		},
		{
			Label:     "Orbital Period",
			Format:    "%.2f",
			Unit:      "days",
			Condition: func(cb models.CelestialBody) bool { return cb.SideralOrbit > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.SideralOrbit },
		},
		{
			Label:     "Rotation Period",
			Format:    "%.2f",
			Unit:      "hours",
			Condition: func(cb models.CelestialBody) bool { return cb.SideralRotation != 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.SideralRotation },
		},
		{
			Label:     "Escape Velocity",
			Format:    "%.2f",
			Unit:      "km/s",
			Condition: func(cb models.CelestialBody) bool { return cb.Escape > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Escape },
		},
		{
			Label:     "Equatorial Radius",
			Format:    "%.0f",
			Unit:      "km",
			Condition: func(cb models.CelestialBody) bool { return cb.EquaRadius > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.EquaRadius },
		},
		{
			Label:     "Polar Radius",
			Format:    "%.0f",
			Unit:      "km",
			Condition: func(cb models.CelestialBody) bool { return cb.PolarRadius > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.PolarRadius },
		},
		{
			Label:     "Flattening",
			Format:    "%.6f",
			Unit:      "",
			Condition: func(cb models.CelestialBody) bool { return cb.Flattening > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Flattening },
		},
		{
			Label:     "Orbital Eccentricity",
			Format:    "%.6f",
			Unit:      "",
			Condition: func(cb models.CelestialBody) bool { return cb.Eccentricity > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Eccentricity },
		},
		{
			Label:     "Orbital Inclination",
			Format:    "%.2f",
			Unit:      "degrees",
			Condition: func(cb models.CelestialBody) bool { return cb.Inclination != 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Inclination },
		},
		{
			Label:     "Volume",
			Format:    "%.2e",
			Unit:      "km³",
			Condition: func(cb models.CelestialBody) bool { return cb.GetVolumeKm3() > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.GetVolumeKm3() },
		},
		{
			Label:     "Perihelion",
			Format:    "%.0f",
			Unit:      "km",
			Condition: func(cb models.CelestialBody) bool { return cb.Perihelion > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Perihelion },
		},
		{
			Label:     "Aphelion",
			Format:    "%.0f",
			Unit:      "km",
			Condition: func(cb models.CelestialBody) bool { return cb.Aphelion > 0 },
			Value:     func(cb models.CelestialBody) interface{} { return cb.Aphelion },
		},
	}
}

// GetCelestialBodyStringFields returns the standardized string field configurations
// for displaying celestial body text data across the application
func GetCelestialBodyStringFields() []StringFieldConfig {
	return []StringFieldConfig{
		{
			Label:     "Dimension",
			Condition: func(cb models.CelestialBody) bool { return cb.Dimension != "" },
			Value:     func(cb models.CelestialBody) string { return cb.Dimension },
		},
		{
			Label:     "Discovered By",
			Condition: func(cb models.CelestialBody) bool { return cb.DiscoveredBy != "" },
			Value:     func(cb models.CelestialBody) string { return cb.DiscoveredBy },
		},
		{
			Label:     "Discovery Date",
			Condition: func(cb models.CelestialBody) bool { return cb.DiscoveryDate != "" },
			Value:     func(cb models.CelestialBody) string { return cb.DiscoveryDate },
		},
		{
			Label:     "Alternative Name",
			Condition: func(cb models.CelestialBody) bool { return cb.AlternativeName != "" },
			Value:     func(cb models.CelestialBody) string { return cb.AlternativeName },
		},
	}
}

// FormatFieldValue formats a field value according to its configuration
func (fc FieldConfig) FormatFieldValue(body models.CelestialBody) string {
	if !fc.Condition(body) {
		return ""
	}

	value := fc.Value(body)
	if fc.Unit != "" {
		return fmt.Sprintf("%s: %s %s", fc.Label, fmt.Sprintf(fc.Format, value), fc.Unit)
	}
	return fmt.Sprintf("%s: %s", fc.Label, fmt.Sprintf(fc.Format, value))
}

// FormatStringFieldValue formats a string field value according to its configuration
func (sfc StringFieldConfig) FormatStringFieldValue(body models.CelestialBody) string {
	if !sfc.Condition(body) {
		return ""
	}
	return fmt.Sprintf("%s: %s", sfc.Label, sfc.Value(body))
}
