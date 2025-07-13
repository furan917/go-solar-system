// Package constants provides centralized configuration and constants
// for the solar system application.
package constants

import "time"

// API Configuration
const (
	SolarSystemAPIBase = "https://api.le-systeme-solaire.net/rest"
	DefaultTimeout     = 10 * time.Second
)

// UI Layout Constants
const (
	ModalWidth        = 70
	ModalMargin       = 2
	ModalContentWidth = 64
	ModalHeight       = 20
	MaxVisibleItems   = 10

	AspectRatio = 2.0

	DisplayUpdateRate = 100 * time.Millisecond
)

// Modal position enumeration
type ModalPosition int

const (
	TopRight ModalPosition = iota
	Center
	TopLeft
	BottomRight
)
