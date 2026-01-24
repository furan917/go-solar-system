// Package constants provides centralized configuration and constants
// for the solar system application.
package constants

import (
	"encoding/base64"
	"time"
)

// API Configuration
const (
	SolarSystemAPIBase = "https://api.le-systeme-solaire.net/rest"
	// This is a light obscuring of a key used for ID only, its simply to avoid repo scrapers, nothing secure here
	SolarSystemAPIKeyObfuscated = "OWprYm1pPmh3PGNpP3dubTk5dztqYjt3azhibGtoOWk8Pzhv"
	DefaultTimeout              = 10 * time.Second
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

// UnobscureKey decodes the API key. This is NOT for security - the API key is a public
// identifier. This light obfuscation just helps avoid automated scrapers.
func UnobscureKey(obscured string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(obscured)
	if err != nil {
		return "", err
	}
	for i := range data {
		data[i] ^= 0x5A
	}
	return string(data), nil
}
