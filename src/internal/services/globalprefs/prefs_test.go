package globalprefs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestLoadPreferences(t *testing.T) {
	tests := []struct {
		name       string
		setupPrefs *MCPPreferences
		expectZero bool
	}{
		{
			name:       "no preferences file",
			setupPrefs: nil,
			expectZero: true,
		},
		{
			name: "existing preferences",
			setupPrefs: &MCPPreferences{
				MCPSetupDeclined: true,
				DeclinedAt:       "2024-01-01T00:00:00Z",
			},
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			svc := New(fs)

			// Setup preferences file if provided
			if tt.setupPrefs != nil {
				prefsPath, _ := svc.(*FileService).getPrefsPath()
				data, _ := json.Marshal(tt.setupPrefs)
				fs.MkdirAll(configDir, 0755)
				afero.WriteFile(fs, prefsPath, data, 0644)
			}

			prefs, err := svc.Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectZero {
				if prefs.MCPSetupDeclined {
					t.Error("expected zero value, got declined=true")
				}
			} else {
				if prefs.MCPSetupDeclined != tt.setupPrefs.MCPSetupDeclined {
					t.Errorf("expected declined=%v, got %v",
						tt.setupPrefs.MCPSetupDeclined, prefs.MCPSetupDeclined)
				}
				if prefs.DeclinedAt != tt.setupPrefs.DeclinedAt {
					t.Errorf("expected declinedAt=%s, got %s",
						tt.setupPrefs.DeclinedAt, prefs.DeclinedAt)
				}
			}
		})
	}
}

func TestSavePreferences(t *testing.T) {
	fs := afero.NewMemMapFs()
	svc := New(fs)

	prefs := MCPPreferences{
		MCPSetupDeclined: true,
		DeclinedAt:       time.Now().Format(time.RFC3339),
	}

	err := svc.Save(prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify saved preferences
	loaded, err := svc.Load()
	if err != nil {
		t.Fatalf("failed to load saved preferences: %v", err)
	}

	if loaded.MCPSetupDeclined != prefs.MCPSetupDeclined {
		t.Errorf("expected declined=%v, got %v",
			prefs.MCPSetupDeclined, loaded.MCPSetupDeclined)
	}

	if loaded.DeclinedAt != prefs.DeclinedAt {
		t.Errorf("expected declinedAt=%s, got %s",
			prefs.DeclinedAt, loaded.DeclinedAt)
	}
}

func TestSavePreferencesCreatesDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	svc := New(fs)

	// Verify directory doesn't exist
	prefsPath, _ := svc.(*FileService).getPrefsPath()
	_, err := fs.Stat(prefsPath)
	if err == nil {
		t.Fatal("preferences file should not exist yet")
	}

	// Save should create directory
	prefs := MCPPreferences{
		MCPSetupDeclined: true,
		DeclinedAt:       time.Now().Format(time.RFC3339),
	}

	err = svc.Save(prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	_, err = fs.Stat(prefsPath)
	if err != nil {
		t.Errorf("preferences file should exist: %v", err)
	}
}

func TestUpdatePreferencesSerialization(t *testing.T) {
	fs := afero.NewMemMapFs()
	svc := New(fs)

	prefs := MCPPreferences{
		MCPSetupDeclined: true,
		DeclinedAt:       "2024-01-01T00:00:00Z",
		UpdateCheck: UpdatePreferences{
			NeverAskAgain:  true,
			DeclinedAt:     "2024-01-02T00:00:00Z",
			LastCheckedAt:  "2024-01-03T00:00:00Z",
			CachedVersion:  "1.2.3",
			CheckSucceeded: true,
		},
	}

	err := svc.Save(prefs)
	if err != nil {
		t.Fatalf("failed to save preferences: %v", err)
	}

	loaded, err := svc.Load()
	if err != nil {
		t.Fatalf("failed to load preferences: %v", err)
	}

	if loaded.UpdateCheck.NeverAskAgain != prefs.UpdateCheck.NeverAskAgain {
		t.Errorf("expected NeverAskAgain=%v, got %v",
			prefs.UpdateCheck.NeverAskAgain, loaded.UpdateCheck.NeverAskAgain)
	}
	if loaded.UpdateCheck.DeclinedAt != prefs.UpdateCheck.DeclinedAt {
		t.Errorf("expected DeclinedAt=%s, got %s",
			prefs.UpdateCheck.DeclinedAt, loaded.UpdateCheck.DeclinedAt)
	}
	if loaded.UpdateCheck.LastCheckedAt != prefs.UpdateCheck.LastCheckedAt {
		t.Errorf("expected LastCheckedAt=%s, got %s",
			prefs.UpdateCheck.LastCheckedAt, loaded.UpdateCheck.LastCheckedAt)
	}
	if loaded.UpdateCheck.CachedVersion != prefs.UpdateCheck.CachedVersion {
		t.Errorf("expected CachedVersion=%s, got %s",
			prefs.UpdateCheck.CachedVersion, loaded.UpdateCheck.CachedVersion)
	}
	if loaded.UpdateCheck.CheckSucceeded != prefs.UpdateCheck.CheckSucceeded {
		t.Errorf("expected CheckSucceeded=%v, got %v",
			prefs.UpdateCheck.CheckSucceeded, loaded.UpdateCheck.CheckSucceeded)
	}
}

func TestIsUpdateCacheValid(t *testing.T) {
	tests := []struct {
		name          string
		prefs         MCPPreferences
		expectedValid bool
	}{
		{
			name: "fresh cache (1 hour old)",
			prefs: MCPPreferences{
				UpdateCheck: UpdatePreferences{
					LastCheckedAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			expectedValid: true,
		},
		{
			name: "expired cache (25 hours old)",
			prefs: MCPPreferences{
				UpdateCheck: UpdatePreferences{
					LastCheckedAt: time.Now().Add(-25 * time.Hour).Format(time.RFC3339),
				},
			},
			expectedValid: false,
		},
		{
			name:          "empty cache",
			prefs:         MCPPreferences{},
			expectedValid: false,
		},
		{
			name: "invalid timestamp",
			prefs: MCPPreferences{
				UpdateCheck: UpdatePreferences{
					LastCheckedAt: "invalid-timestamp",
				},
			},
			expectedValid: false,
		},
		{
			name: "exactly 24 hours old (boundary)",
			prefs: MCPPreferences{
				UpdateCheck: UpdatePreferences{
					LastCheckedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.prefs.IsUpdateCacheValid()
			if valid != tt.expectedValid {
				t.Errorf("expected IsUpdateCacheValid()=%v, got %v",
					tt.expectedValid, valid)
			}
		})
	}
}

func TestSetUpdateCache(t *testing.T) {
	prefs := MCPPreferences{}

	// Capture time before SetUpdateCache
	before := time.Now()

	prefs.SetUpdateCache("1.2.3", true)

	// Verify CachedVersion is set correctly
	if prefs.UpdateCheck.CachedVersion != "1.2.3" {
		t.Errorf("expected CachedVersion=1.2.3, got %s",
			prefs.UpdateCheck.CachedVersion)
	}

	// Verify CheckSucceeded is set correctly
	if !prefs.UpdateCheck.CheckSucceeded {
		t.Error("expected CheckSucceeded=true, got false")
	}

	// Verify LastCheckedAt is a valid timestamp
	if prefs.UpdateCheck.LastCheckedAt == "" {
		t.Error("expected LastCheckedAt to be set, got empty string")
	}

	lastChecked, err := time.Parse(time.RFC3339, prefs.UpdateCheck.LastCheckedAt)
	if err != nil {
		t.Fatalf("failed to parse LastCheckedAt: %v", err)
	}

	// Verify timestamp is recent (within last second, allowing for clock precision)
	if time.Since(lastChecked) > time.Second {
		t.Errorf("expected LastCheckedAt to be recent, but was %s ago",
			time.Since(lastChecked))
	}

	// Verify timestamp is not in the future (allowing for tiny clock skew)
	if lastChecked.After(before.Add(time.Millisecond * 10)) {
		t.Errorf("expected LastCheckedAt to not be in the future, got %s (before was %s)",
			lastChecked.Format(time.RFC3339), before.Format(time.RFC3339))
	}

	// Test with failed check
	prefs.SetUpdateCache("1.2.4", false)
	if prefs.UpdateCheck.CachedVersion != "1.2.4" {
		t.Errorf("expected CachedVersion=1.2.4, got %s",
			prefs.UpdateCheck.CachedVersion)
	}
	if prefs.UpdateCheck.CheckSucceeded {
		t.Error("expected CheckSucceeded=false, got true")
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test loading old preferences without UpdateCheck field
	fs := afero.NewMemMapFs()
	svc := New(fs)

	// Manually create old-style preferences JSON
	oldPrefsJSON := `{
		"mcpSetupDeclined": true,
		"declinedAt": "2024-01-01T00:00:00Z"
	}`

	prefsPath, _ := svc.(*FileService).getPrefsPath()
	fs.MkdirAll(configDir, 0755)
	afero.WriteFile(fs, prefsPath, []byte(oldPrefsJSON), 0644)

	// Load should work without error
	loaded, err := svc.Load()
	if err != nil {
		t.Fatalf("failed to load old preferences: %v", err)
	}

	// Verify existing fields are loaded correctly
	if !loaded.MCPSetupDeclined {
		t.Error("expected MCPSetupDeclined=true, got false")
	}
	if loaded.DeclinedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("expected DeclinedAt=2024-01-01T00:00:00Z, got %s",
			loaded.DeclinedAt)
	}

	// Verify UpdateCheck has zero values (backward compatible)
	if loaded.UpdateCheck.NeverAskAgain {
		t.Error("expected UpdateCheck.NeverAskAgain=false, got true")
	}
	if loaded.UpdateCheck.LastCheckedAt != "" {
		t.Errorf("expected UpdateCheck.LastCheckedAt='', got %s",
			loaded.UpdateCheck.LastCheckedAt)
	}
	if loaded.UpdateCheck.CachedVersion != "" {
		t.Errorf("expected UpdateCheck.CachedVersion='', got %s",
			loaded.UpdateCheck.CachedVersion)
	}
}
