package systems

import (
	"errors"
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
	searchPaths      []fs.FS
	availableSystems map[string]systemSource
	currentSystem    string
	loadedSystems    map[string]SystemData
	cachedSystemInfo map[string]string
	formatRegistry   *formats.FormatRegistry
}

// NewSystemManager creates a system manager reading bundled systems and userDirs, where later sources take precedence
func NewSystemManager(bundled fs.FS, userDirs ...string) *SystemManager {
	var searchPaths []fs.FS

	if bundled != nil {
		searchPaths = append(searchPaths, bundled)
	}

	for _, dir := range userDirs {
		if dir == "" {
			continue
		}
		searchPaths = append(searchPaths, os.DirFS(dir))
	}

	return &SystemManager{
		searchPaths:      searchPaths,
		availableSystems: make(map[string]systemSource),
		loadedSystems:    make(map[string]SystemData),
		cachedSystemInfo: make(map[string]string),
		currentSystem:    "solar-system",
		formatRegistry:   formats.NewFormatRegistry(),
	}
}

// ScanSystems scans every configured source for available system files
func (sm *SystemManager) ScanSystems() error {
	for _, fsys := range sm.searchPaths {
		if err := sm.scanSource(fsys); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SystemManager) scanSource(fsys fs.FS) error {
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if _, supported := sm.formatRegistry.GetHandlerForExtension(ext); !supported {
			return nil
		}

		systemName := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		if err := validateSystemName(systemName); err != nil {
			return fmt.Errorf("invalid system name %s: %w", systemName, err)
		}

		sm.availableSystems[systemName] = systemSource{fsys: fsys, path: path}

		return nil
	})

	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}

// readSystemFile returns a registered system's contents and its format handler
func (sm *SystemManager) readSystemFile(systemName string) ([]byte, formats.FileFormat, error) {
	source, exists := sm.availableSystems[systemName]
	if !exists {
		return nil, nil, fmt.Errorf("system '%s' not found", systemName)
	}

	data, err := fs.ReadFile(source.fsys, source.path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read system file %s: %w", source.path, err)
	}

	ext := strings.ToLower(filepath.Ext(source.path))
	handler, exists := sm.formatRegistry.GetHandlerForExtension(ext)
	if !exists {
		return nil, nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return data, handler, nil
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

	data, handler, err := sm.readSystemFile(systemName)
	if err != nil {
		return nil, err
	}

	systemData, err := handler.ParseSystemData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse system file %s: %w", systemName, err)
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
	data, handler, err := sm.readSystemFile(systemName)
	if err != nil {
		return nil, err
	}

	metadata, err := handler.ParseSystemMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse system metadata %s: %w", systemName, err)
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

// SystemEntry pairs a system identifier with its display line
type SystemEntry struct {
	Name string
	Info string
}

// ListSystems returns every available system with its display line
func (sm *SystemManager) ListSystems() []SystemEntry {
	names := sm.GetAvailableSystems()
	entries := make([]SystemEntry, 0, len(names))

	for _, name := range names {
		info, err := sm.GetSystemInfo(name)
		if err != nil {
			continue
		}

		marker := " "
		if name == sm.currentSystem {
			marker = "*"
		}

		entries = append(entries, SystemEntry{
			Name: name,
			Info: fmt.Sprintf(" %s %s", marker, info),
		})
	}

	return entries
}

// FilterSystems returns systems matching every whitespace-separated term in query, case-insensitively
func (sm *SystemManager) FilterSystems(query string) []SystemEntry {
	entries := sm.ListSystems()

	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return entries
	}

	matches := make([]SystemEntry, 0, len(entries))
	for _, entry := range entries {
		haystack := strings.ToLower(entry.Name + " " + entry.Info)

		matchesAll := true
		for _, term := range terms {
			if !strings.Contains(haystack, term) {
				matchesAll = false
				break
			}
		}

		if matchesAll {
			matches = append(matches, entry)
		}
	}

	return matches
}

// ListSystemsWithInfo returns a formatted list of all available systems with descriptions
func (sm *SystemManager) ListSystemsWithInfo() ([]string, error) {
	entries := sm.ListSystems()

	info := make([]string, 0, len(entries))
	for _, entry := range entries {
		info = append(info, entry.Info)
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
	var f []string
	for _, format := range sm.formatRegistry.GetAllFormats() {
		extensions := strings.Join(format.GetSupportedExtensions(), ", ")
		f = append(f, fmt.Sprintf("%s (%s)", format.GetFormatName(), extensions))
	}
	return f
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
