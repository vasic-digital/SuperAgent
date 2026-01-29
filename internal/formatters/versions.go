package formatters

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// VersionsManifest tracks pinned versions for all formatters
type VersionsManifest struct {
	Native  map[string]VersionInfo `yaml:"native"`
	Service map[string]VersionInfo `yaml:"service"`
	Builtin map[string]VersionInfo `yaml:"builtin"`
}

// VersionInfo contains version information for a formatter
type VersionInfo struct {
	Version string `yaml:"version"`
	Commit  string `yaml:"commit"`
	GitRef  string `yaml:"git_ref"`
	Note    string `yaml:"note,omitempty"`
}

// LoadVersionsManifest loads the versions manifest from a file
func LoadVersionsManifest(path string) (*VersionsManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions manifest: %w", err)
	}

	var manifest VersionsManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse versions manifest: %w", err)
	}

	return &manifest, nil
}

// SaveVersionsManifest saves the versions manifest to a file
func SaveVersionsManifest(manifest *VersionsManifest, path string) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal versions manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write versions manifest: %w", err)
	}

	return nil
}

// GetVersion returns the version info for a formatter
func (m *VersionsManifest) GetVersion(name string, ftype FormatterType) (VersionInfo, bool) {
	switch ftype {
	case FormatterTypeNative:
		info, ok := m.Native[name]
		return info, ok
	case FormatterTypeService:
		info, ok := m.Service[name]
		return info, ok
	case FormatterTypeBuiltin:
		info, ok := m.Builtin[name]
		return info, ok
	default:
		return VersionInfo{}, false
	}
}

// SetVersion sets the version info for a formatter
func (m *VersionsManifest) SetVersion(name string, ftype FormatterType, info VersionInfo) {
	switch ftype {
	case FormatterTypeNative:
		if m.Native == nil {
			m.Native = make(map[string]VersionInfo)
		}
		m.Native[name] = info
	case FormatterTypeService:
		if m.Service == nil {
			m.Service = make(map[string]VersionInfo)
		}
		m.Service[name] = info
	case FormatterTypeBuiltin:
		if m.Builtin == nil {
			m.Builtin = make(map[string]VersionInfo)
		}
		m.Builtin[name] = info
	}
}

// AllVersions returns all version info
func (m *VersionsManifest) AllVersions() map[string]VersionInfo {
	all := make(map[string]VersionInfo)

	for name, info := range m.Native {
		all[name] = info
	}

	for name, info := range m.Service {
		all[name] = info
	}

	for name, info := range m.Builtin {
		all[name] = info
	}

	return all
}

// Count returns the total number of formatters
func (m *VersionsManifest) Count() int {
	return len(m.Native) + len(m.Service) + len(m.Builtin)
}
