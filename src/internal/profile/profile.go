// Package profile provides profile loading and management for Claudex agents.
// It supports loading profiles from both embedded FS and the filesystem
// .claude/agents/ directory, with profile composition capabilities.
package profile

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// GetProfiles returns a sorted list of all available profile names from both
// embedded FS and filesystem .claude/agents/ directory.
func GetProfiles(profilesFS fs.FS) ([]string, error) {
	profileSet := make(map[string]bool)

	// Look for profiles in embedded FS profiles/agents/ directory
	entries, err := fs.ReadDir(profilesFS, "profiles/agents")
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && !strings.HasPrefix(name, ".") {
				profileSet[name] = true
			}
		}
	}

	// Also look for profiles in filesystem .claude/agents/ directory
	fsAgentsDir := filepath.Join(".claude", "agents")
	if fsEntries, err := os.ReadDir(fsAgentsDir); err == nil {
		for _, entry := range fsEntries {
			name := entry.Name()
			if !entry.IsDir() && !strings.HasPrefix(name, ".") {
				// Remove .md extension for consistent naming
				name = strings.TrimSuffix(name, ".md")
				profileSet[name] = true
			}
		}
	}

	// Convert set to sorted slice
	var profiles []string
	for name := range profileSet {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	return profiles, nil
}

// ExtractDescription extracts a profile description from the first line
// that contains role-related keywords.
func ExtractDescription(profilesFS fs.FS, profilePath string) string {
	file, err := profilesFS.Open(profilePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`(?i)(role:|principal|agent)`)

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			desc := strings.TrimLeft(line, "#*- ")
			desc = regexp.MustCompile(`(?i)role:`).ReplaceAllString(desc, "")
			desc = strings.TrimSpace(desc)
			if len(desc) > 60 {
				desc = desc[:60]
			}
			return desc
		}
	}

	return ""
}

// Load loads a profile from the embedded FS profiles/agents/ directory.
func Load(profilesFS fs.FS, profileName string) ([]byte, error) {
	// Look for profile in profiles/agents/ directory
	agentPath := "profiles/agents/" + profileName
	return fs.ReadFile(profilesFS, agentPath)
}

// LoadFromFS loads a profile from the filesystem (.claude/agents/).
func LoadFromFS(profileName string) ([]byte, error) {
	// Try with .md extension first
	agentPath := filepath.Join(".claude", "agents", profileName+".md")
	if data, err := os.ReadFile(agentPath); err == nil {
		return data, nil
	}

	// Try without extension
	agentPath = filepath.Join(".claude", "agents", profileName)
	return os.ReadFile(agentPath)
}

// LoadComposed tries to load a profile from embedded FS first, then filesystem.
func LoadComposed(profilesFS fs.FS, profileName string) ([]byte, error) {
	// First try embedded FS
	if data, err := Load(profilesFS, profileName); err == nil {
		return data, nil
	}

	// Then try filesystem
	return LoadFromFS(profileName)
}

// ResolvePath resolves a profile path in the embedded FS.
func ResolvePath(profilesFS fs.FS, profileName string) string {
	// Look for profile in profiles/agents/ directory
	agentPath := "profiles/agents/" + profileName
	if _, err := fs.Stat(profilesFS, agentPath); err == nil {
		return agentPath
	}

	return ""
}
