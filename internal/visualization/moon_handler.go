package visualization

import (
	"fmt"
	"strings"

	"github.com/furan917/go-solar-system/internal/models"
)

// MoonHandler handles moon name resolution and display
type MoonHandler struct {
	famousMoons map[string][]string
}

// NewMoonHandler creates a new moon handler with well-known moon names
func NewMoonHandler() *MoonHandler {
	return &MoonHandler{
		famousMoons: map[string][]string{
			"Earth":   {"Moon"},
			"Mars":    {"Phobos", "Deimos"},
			"Jupiter": {"Io", "Europa", "Ganymede", "Callisto"},
			"Saturn":  {"Titan", "Enceladus", "Mimas", "Rhea"},
			"Uranus":  {"Titania", "Oberon", "Umbriel", "Ariel"},
			"Neptune": {"Triton", "Nereid"},
		},
	}
}

// GetMoonNames returns appropriate moon names for display
func (mh *MoonHandler) GetMoonNames(planet models.CelestialBody) []string {
	moonCount := len(planet.Moons)
	if moonCount == 0 {
		return []string{}
	}

	var moonNames []string
	for _, moon := range planet.Moons {
		name := mh.GetMoonNameFromAPI(moon)
		if name != "" {
			moonNames = append(moonNames, name)
		}
	}

	if len(moonNames) == 0 {
		if famousMoons, exists := mh.famousMoons[planet.EnglishName]; exists {
			for i, name := range famousMoons {
				if i < moonCount {
					moonNames = append(moonNames, name)
				}
			}
		}
	}

	return moonNames
}

// GetMoonNameFromAPI extracts moon name from API data (exported for use in app)
func (mh *MoonHandler) GetMoonNameFromAPI(moon models.Moon) string {
	if moon.EnglishName != "" {
		return moon.EnglishName
	}

	if moon.Name != "" {
		return moon.Name
	}

	if moon.ID != "" {
		return moon.ID
	}

	if moon.Rel != "" {
		return mh.extractMoonNameFromURL(moon.Rel)
	}

	return ""
}

// extractMoonNameFromURL extracts moon name from API URL
func (mh *MoonHandler) extractMoonNameFromURL(url string) string {
	if url == "" {
		return ""
	}

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		id := parts[len(parts)-1]
		return mh.prettifyMoonName(id)
	}

	return ""
}

// prettifyMoonName converts API IDs to readable names
func (mh *MoonHandler) prettifyMoonName(id string) string {
	nameMap := map[string]string{
		"lune":     "Moon",
		"phobos":   "Phobos",
		"deimos":   "Deimos",
		"io":       "Io",
		"europa":   "Europa",
		"ganymede": "Ganymede",
		"callisto": "Callisto",
		"titan":    "Titan",
		"encelade": "Enceladus",
		"mimas":    "Mimas",
		"rhea":     "Rhea",
		"titania":  "Titania",
		"oberon":   "Oberon",
		"umbriel":  "Umbriel",
		"ariel":    "Ariel",
		"triton":   "Triton",
		"nereid":   "Nereid",
	}

	if prettyName, exists := nameMap[strings.ToLower(id)]; exists {
		return prettyName
	}

	if len(id) > 0 {
		return strings.ToUpper(id[:1]) + strings.ToLower(id[1:])
	}

	return id
}

// FormatMoonDisplay formats moon information for display
func (mh *MoonHandler) FormatMoonDisplay(planet models.CelestialBody, maxMoons int) []string {
	moonCount := len(planet.Moons)
	if moonCount == 0 {
		return []string{}
	}

	var lines []string
	moonNames := mh.GetMoonNames(planet)

	lines = append(lines, fmt.Sprintf("Moons: %d", moonCount))

	displayCount := len(moonNames)
	if displayCount > maxMoons {
		displayCount = maxMoons
	}

	for i := 0; i < displayCount; i++ {
		lines = append(lines, fmt.Sprintf("  • %s", moonNames[i]))
	}

	if moonCount > displayCount {
		remaining := moonCount - displayCount
		lines = append(lines, fmt.Sprintf("  • ... and %d more", remaining))
	}

	return lines
}
