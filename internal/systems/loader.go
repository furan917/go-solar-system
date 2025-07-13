package systems

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/furan917/go-solar-system/internal/systems/formats"
)

// SystemData represents an external star system (now using interface-based loading)
type SystemData = formats.SystemData

// SystemManager handles loading and switching between star systems
type SystemManager struct {
	systemsDir       string
	availableSystems map[string]string
	currentSystem    string
	loadedSystems    map[string]SystemData
	cachedSystemInfo map[string]string
	formatRegistry   *formats.FormatRegistry
}

// NewSystemManager creates a new system manager
func NewSystemManager(systemsDir string) *SystemManager {
	return &SystemManager{
		systemsDir:       systemsDir,
		availableSystems: make(map[string]string),
		loadedSystems:    make(map[string]SystemData),
		cachedSystemInfo: make(map[string]string),
		currentSystem:    "solar-system",
		formatRegistry:   formats.NewFormatRegistry(),
	}
}

// ScanSystems scans the systems directory for available system files
func (sm *SystemManager) ScanSystems() error {
	if _, err := os.Stat(sm.systemsDir); os.IsNotExist(err) {
		return nil
	}

	baseDir, err := filepath.Abs(sm.systemsDir)
	if err != nil {
		return fmt.Errorf("invalid systems directory: %w", err)
	}

	err = filepath.WalkDir(sm.systemsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", path, err)
		}

		if !strings.HasPrefix(absPath, baseDir) {
			return fmt.Errorf("path traversal detected: %s", path)
		}

		if d.IsDir() {
			return nil
		}

		// Check if file extension is supported by any registered format
		ext := strings.ToLower(filepath.Ext(path))
		if _, supported := sm.formatRegistry.GetHandlerForExtension(ext); supported {
			filename := d.Name()
			systemName := strings.TrimSuffix(filename, filepath.Ext(filename))

			if err := validateSystemName(systemName); err != nil {
				return fmt.Errorf("invalid system name %s: %w", systemName, err)
			}

			sm.availableSystems[systemName] = path
		}

		return nil
	})

	return err
}

// GetAvailableSystems returns a list of available system names in alphabetical order
func (sm *SystemManager) GetAvailableSystems() []string {
	systems := []string{"solar-system"}

	var externalSystems []string
	for name := range sm.availableSystems {
		externalSystems = append(externalSystems, name)
	}
	sort.Strings(externalSystems)

	systems = append(systems, externalSystems...)

	return systems
}

// GetCurrentSystem returns the name of the currently selected system
func (sm *SystemManager) GetCurrentSystem() string {
	return sm.currentSystem
}

// LoadSystem loads a specific star system
func (sm *SystemManager) LoadSystem(systemName string) (*SystemData, error) {
	if system, exists := sm.loadedSystems[systemName]; exists {
		return &system, nil
	}

	if systemName == "solar-system" {
		return nil, fmt.Errorf("solar system should be loaded via API")
	}

	filePath, exists := sm.availableSystems[systemName]
	if !exists {
		return nil, fmt.Errorf("system '%s' not found", systemName)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read system file %s: %w", filePath, err)
	}

	// Detect format and get appropriate handler
	ext := strings.ToLower(filepath.Ext(filePath))
	handler, exists := sm.formatRegistry.GetHandlerForExtension(ext)
	if !exists {
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	// Parse system data using the appropriate format handler
	systemData, err := handler.ParseSystemData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse system file %s: %w", filePath, err)
	}

	system := *systemData

	sm.loadedSystems[systemName] = system

	return &system, nil
}

// SwitchToSystem switches to a different star system
func (sm *SystemManager) SwitchToSystem(systemName string) error {
	if systemName == "solar-system" {
		sm.currentSystem = systemName
		return nil
	}

	_, err := sm.LoadSystem(systemName)
	if err != nil {
		return err
	}

	sm.currentSystem = systemName
	return nil
}

// GetSystemData returns the data for the currently selected system
func (sm *SystemManager) GetSystemData() (*SystemData, error) {
	if sm.currentSystem == "solar-system" {
		return nil, fmt.Errorf("solar system data should be fetched via API")
	}

	return sm.LoadSystem(sm.currentSystem)
}

// GetCurrentSystemDisplayName returns the current system name with galaxy
func (sm *SystemManager) GetCurrentSystemDisplayName() string {
	if sm.currentSystem == "solar-system" {
		return "Solar System, Milky Way"
	}

	metadata, err := sm.LoadSystemMetadata(sm.currentSystem)
	if err != nil {
		return sm.currentSystem
	}

	if metadata.Galaxy != "" {
		return fmt.Sprintf("%s, %s", metadata.SystemName, metadata.Galaxy)
	}

	return metadata.SystemName
}

// GetSystemInfo returns descriptive information about a system
func (sm *SystemManager) GetSystemInfo(systemName string) (string, error) {
	if cached, exists := sm.cachedSystemInfo[systemName]; exists {
		return cached, nil
	}

	var info string
	if systemName == "solar-system" {
		info = "Our Solar System - The system containing Earth and 8 planets orbiting the Sun"
	} else {
		metadata, err := sm.LoadSystemMetadata(systemName)
		if err != nil {
			return "", err
		}

		info = fmt.Sprintf("%s - %s (Discovered: %s, Distance: %s)",
			metadata.SystemName, metadata.Description, metadata.DiscoveryYear, metadata.Distance)
	}

	sm.cachedSystemInfo[systemName] = info

	return info, nil
}

// LoadSystemMetadata loads only the metadata (not celestial bodies) for performance
func (sm *SystemManager) LoadSystemMetadata(systemName string) (*SystemData, error) {
	filePath, exists := sm.availableSystems[systemName]
	if !exists {
		return nil, fmt.Errorf("system '%s' not found", systemName)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read system file %s: %w", filePath, err)
	}

	// Detect format and get appropriate handler
	ext := strings.ToLower(filepath.Ext(filePath))
	handler, exists := sm.formatRegistry.GetHandlerForExtension(ext)
	if !exists {
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	// Parse system metadata using the appropriate format handler
	metadata, err := handler.ParseSystemMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse system metadata %s: %w", filePath, err)
	}

	return &SystemData{
		SystemName:    metadata.SystemName,
		Description:   metadata.Description,
		DiscoveryYear: metadata.DiscoveryYear,
		Distance:      metadata.Distance,
		Galaxy:        metadata.Galaxy,
		Bodies:        nil,
	}, nil
}

// ListSystemsWithInfo returns a formatted list of all available systems with descriptions
func (sm *SystemManager) ListSystemsWithInfo() ([]string, error) {
	systems := sm.GetAvailableSystems()
	var info []string

	for _, systemName := range systems {
		systemInfo, err := sm.GetSystemInfo(systemName)
		if err != nil {
			continue
		}

		marker := " "
		if systemName == sm.currentSystem {
			marker = "*"
		}

		info = append(info, fmt.Sprintf(" %s %s", marker, systemInfo))
	}

	return info, nil
}

// validateSystemName validates system names to prevent injection attacks
func validateSystemName(name string) error {
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("system name contains invalid character: %c", char)
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("system name cannot be empty")
	}

	if len(name) > 64 {
		return fmt.Errorf("system name too long: %d characters (max: 64)", len(name))
	}

	if name[0] == '-' || name[0] == '_' {
		return fmt.Errorf("system name cannot start with special character: %c", name[0])
	}

	return nil
}

// GetSupportedFormats returns information about all supported file formats
func (sm *SystemManager) GetSupportedFormats() []string {
	var formats []string
	for _, format := range sm.formatRegistry.GetAllFormats() {
		extensions := strings.Join(format.GetSupportedExtensions(), ", ")
		formats = append(formats, fmt.Sprintf("%s (%s)", format.GetFormatName(), extensions))
	}
	return formats
}

// ValidateSystemFile validates a system file using format detection
func (sm *SystemManager) ValidateSystemFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Try extension-based detection first
	ext := strings.ToLower(filepath.Ext(filePath))
	if handler, exists := sm.formatRegistry.GetHandlerForExtension(ext); exists {
		return handler.ValidateFormat(data)
	}

	// Fall back to content-based detection
	_, err = sm.formatRegistry.DetectFormat(data)
	return err
}
