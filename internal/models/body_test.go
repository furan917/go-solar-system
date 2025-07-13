package models

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance*math.Max(math.Abs(a), math.Abs(b))
}

func TestCelestialBody_GetMassKg(t *testing.T) {
	tests := []struct {
		name     string
		body     CelestialBody
		expected float64
	}{
		{
			name: "Earth mass",
			body: CelestialBody{
				Mass: Mass{
					MassValue:    5.97237,
					MassExponent: 24,
				},
			},
			expected: 5.97237e24,
		},
		{
			name: "Mars mass",
			body: CelestialBody{
				Mass: Mass{
					MassValue:    6.4171,
					MassExponent: 23,
				},
			},
			expected: 6.4171e23,
		},
		{
			name: "Zero mass",
			body: CelestialBody{
				Mass: Mass{
					MassValue:    0,
					MassExponent: 0,
				},
			},
			expected: 0,
		},
		{
			name: "Negative exponent",
			body: CelestialBody{
				Mass: Mass{
					MassValue:    1.5,
					MassExponent: -3,
				},
			},
			expected: 0.0015,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.body.GetMassKg()
			if !almostEqual(result, tt.expected, 1e-10) {
				t.Errorf("GetMassKg() = %g, want %g", result, tt.expected)
			}
		})
	}
}

func TestCelestialBody_GetVolumeKm3(t *testing.T) {
	tests := []struct {
		name     string
		body     CelestialBody
		expected float64
	}{
		{
			name: "Earth volume",
			body: CelestialBody{
				Vol: Vol{
					VolValue:    1.08321,
					VolExponent: 12,
				},
			},
			expected: 1.08321e12,
		},
		{
			name: "Zero volume",
			body: CelestialBody{
				Vol: Vol{
					VolValue:    0,
					VolExponent: 0,
				},
			},
			expected: 0,
		},
		{
			name: "Small volume",
			body: CelestialBody{
				Vol: Vol{
					VolValue:    2.5,
					VolExponent: 8,
				},
			},
			expected: 2.5e8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.body.GetVolumeKm3()
			if result != tt.expected {
				t.Errorf("GetVolumeKm3() = %e, want %e", result, tt.expected)
			}
		})
	}
}
