package formats

import (
	"fmt"

	"github.com/furan917/go-solar-system/internal/models"
)

// SystemData represents an external star system with metadata
type SystemData struct {
	SystemName    string                 `json:"systemName"`
	Description   string                 `json:"description"`
	DiscoveryYear string                 `json:"discoveryYear"`
	Distance      string                 `json:"distance"`
	Galaxy        string                 `json:"galaxy"`
	Bodies        []models.CelestialBody `json:"bodies"`
}

// SystemMetadata represents just the metadata portion (without celestial bodies)
type SystemMetadata struct {
	SystemName    string `json:"systemName"`
	Description   string `json:"description"`
	DiscoveryYear string `json:"discoveryYear"`
	Distance      string `json:"distance"`
	Galaxy        string `json:"galaxy"`
}

// FileFormat defines the interface that all file format handlers must implement
// This allows for easy extension to support additional formats beyond JSON
type FileFormat interface {
	// GetSupportedExtensions returns the file extensions this handler supports (e.g., [".json", ".yaml"])
	GetSupportedExtensions() []string

	// GetFormatName returns a human-readable name for this format (e.g., "JSON", "YAML")
	GetFormatName() string

	// ParseSystemData parses the complete system data from file content
	ParseSystemData(data []byte) (*SystemData, error)

	// ParseSystemMetadata parses only the metadata (for performance) from file content
	ParseSystemMetadata(data []byte) (*SystemMetadata, error)

	// ValidateFormat performs basic validation to ensure the data is in the expected format
	ValidateFormat(data []byte) error

	// GetMimeType returns the MIME type for this format (optional, can return empty string)
	GetMimeType() string
}

// FormatRegistry manages all available file format handlers
type FormatRegistry struct {
	handlers map[string]FileFormat // extension -> handler mapping
	formats  []FileFormat          // list of all handlers
}

// NewFormatRegistry creates a new format registry with all available formats
func NewFormatRegistry() *FormatRegistry {
	registry := &FormatRegistry{
		handlers: make(map[string]FileFormat),
		formats:  make([]FileFormat, 0),
	}

	// Register built-in formats
	registry.RegisterFormat(NewJSONFormat())

	// Example: To add YAML support, uncomment the line below and ensure yaml.go has proper implementation
	// registry.RegisterFormat(NewYAMLFormat())

	return registry
}

// RegisterFormat registers a new file format handler
func (fr *FormatRegistry) RegisterFormat(format FileFormat) {
	fr.formats = append(fr.formats, format)

	for _, ext := range format.GetSupportedExtensions() {
		fr.handlers[ext] = format
	}
}

// GetHandlerForExtension returns the file format handler for a given extension
func (fr *FormatRegistry) GetHandlerForExtension(extension string) (FileFormat, bool) {
	handler, exists := fr.handlers[extension]
	return handler, exists
}

// GetAllFormats returns all registered format handlers
func (fr *FormatRegistry) GetAllFormats() []FileFormat {
	return fr.formats
}

// GetSupportedExtensions returns all supported file extensions
func (fr *FormatRegistry) GetSupportedExtensions() []string {
	var extensions []string
	for ext := range fr.handlers {
		extensions = append(extensions, ext)
	}
	return extensions
}

// DetectFormat attempts to detect the format of file content by trying all registered handlers
func (fr *FormatRegistry) DetectFormat(data []byte) (FileFormat, error) {
	var lastError error

	for _, format := range fr.formats {
		if err := format.ValidateFormat(data); err == nil {
			return format, nil
		} else {
			lastError = err
		}
	}

	if lastError != nil {
		return nil, lastError
	}

	return nil, fmt.Errorf("no supported format detected")
}
