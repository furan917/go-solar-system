package formats

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONFormat implements the FileFormat interface for JSON files
type JSONFormat struct{}

// NewJSONFormat creates a new JSON format handler
func NewJSONFormat() *JSONFormat {
	return &JSONFormat{}
}

// GetSupportedExtensions returns the file extensions this handler supports
func (jf *JSONFormat) GetSupportedExtensions() []string {
	return []string{".json"}
}

// GetFormatName returns a human-readable name for this format
func (jf *JSONFormat) GetFormatName() string {
	return "JSON"
}

// ParseSystemData parses the complete system data from JSON content
func (jf *JSONFormat) ParseSystemData(data []byte) (*SystemData, error) {
	var system SystemData
	if err := json.Unmarshal(data, &system); err != nil {
		return nil, fmt.Errorf("failed to parse JSON system data: %w", err)
	}

	// Validate required fields
	if err := jf.validateSystemData(&system); err != nil {
		return nil, fmt.Errorf("invalid system data: %w", err)
	}

	return &system, nil
}

// ParseSystemMetadata parses only the metadata (for performance) from JSON content
func (jf *JSONFormat) ParseSystemMetadata(data []byte) (*SystemMetadata, error) {
	var metadata SystemMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse JSON system metadata: %w", err)
	}

	// Validate required fields
	if err := jf.validateSystemMetadata(&metadata); err != nil {
		return nil, fmt.Errorf("invalid system metadata: %w", err)
	}

	return &metadata, nil
}

// ValidateFormat performs basic validation to ensure the data is valid JSON
func (jf *JSONFormat) ValidateFormat(data []byte) error {
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Additional validation to ensure it looks like system data
	var system map[string]interface{}
	if err := json.Unmarshal(data, &system); err != nil {
		return fmt.Errorf("JSON data is not an object: %w", err)
	}

	// Check for required fields
	requiredFields := []string{"systemName", "bodies"}
	for _, field := range requiredFields {
		if _, exists := system[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// GetMimeType returns the MIME type for JSON
func (jf *JSONFormat) GetMimeType() string {
	return "application/json"
}

// validateSystemData validates the complete system data structure
func (jf *JSONFormat) validateSystemData(system *SystemData) error {
	if strings.TrimSpace(system.SystemName) == "" {
		return fmt.Errorf("systemName cannot be empty")
	}

	if len(system.Bodies) == 0 {
		return fmt.Errorf("system must contain at least one celestial body")
	}

	// Validate each celestial body has required fields
	for i, body := range system.Bodies {
		if strings.TrimSpace(body.EnglishName) == "" {
			return fmt.Errorf("celestial body at index %d missing englishName", i)
		}
	}

	return nil
}

// validateSystemMetadata validates the system metadata structure
func (jf *JSONFormat) validateSystemMetadata(metadata *SystemMetadata) error {
	if strings.TrimSpace(metadata.SystemName) == "" {
		return fmt.Errorf("systemName cannot be empty")
	}

	return nil
}
