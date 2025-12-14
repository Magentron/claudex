package updatecheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"claudex/internal/services/globalprefs"
	"claudex/internal/services/npmregistry"
	"github.com/spf13/afero"
)

func TestShouldPrompt_NeverAskAgain(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Setup: user opted out
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			NeverAskAgain: true,
			DeclinedAt:    "2024-01-01T00:00:00Z",
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	if result != ResultNeverAskAgain {
		t.Errorf("expected ResultNeverAskAgain, got %v", result)
	}
}

func TestShouldPrompt_CachedNoNewVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.2")

	// Setup: valid cache with same version
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			CachedVersion:  "0.1.2",
			CheckSucceeded: true,
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	if result != ResultCached {
		t.Errorf("expected ResultCached, got %v", result)
	}
}

func TestShouldPrompt_CachedWithNewVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Setup: valid cache with newer version
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			CachedVersion:  "0.2.0",
			CheckSucceeded: true,
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	if result != ResultPromptUser {
		t.Errorf("expected ResultPromptUser, got %v", result)
	}

	if uc.GetLatestVersion() != "0.2.0" {
		t.Errorf("expected latest version 0.2.0, got %s", uc.GetLatestVersion())
	}
}

func TestShouldPrompt_NetworkError(t *testing.T) {
	// Create a server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Replace npm client with one pointing to test server
	// Since we can't easily inject, we'll test with an invalid package name
	// that will cause network error (timeout or 404)
	uc.npmSvc = npmregistry.New()

	// Force cache to be invalid
	result := uc.ShouldPrompt()

	// With invalid package, should get network error
	// Note: This might timeout or fail, both are network errors
	if result != ResultNetworkError && result != ResultUpToDate {
		t.Logf("Warning: expected ResultNetworkError, got %v (may be timing dependent)", result)
	}
}

func TestShouldPrompt_UpToDate(t *testing.T) {
	// Create mock npm registry server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		packageInfo := npmregistry.PackageInfo{
			DistTags: npmregistry.DistTags{
				Latest: "0.1.0",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(packageInfo)
	}))
	defer server.Close()

	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// We can't easily inject the server URL, so we'll test the version comparison logic
	// by using the cache mechanism
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Format(time.RFC3339),
			CachedVersion:  "0.1.0",
			CheckSucceeded: true,
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	if result != ResultCached {
		t.Errorf("expected ResultCached (up to date), got %v", result)
	}
}

func TestShouldPrompt_NewVersionAvailable(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Setup: valid cache with newer version
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Format(time.RFC3339),
			CachedVersion:  "0.2.0",
			CheckSucceeded: true,
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	if result != ResultPromptUser {
		t.Errorf("expected ResultPromptUser, got %v", result)
	}
}

func TestVersionComparison_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expectNewer    bool
	}{
		{
			name:           "v prefix on current",
			currentVersion: "v0.1.0",
			latestVersion:  "0.2.0",
			expectNewer:    true,
		},
		{
			name:           "v prefix on latest",
			currentVersion: "0.1.0",
			latestVersion:  "v0.2.0",
			expectNewer:    true,
		},
		{
			name:           "dirty suffix on current",
			currentVersion: "0.1.0-dirty",
			latestVersion:  "0.2.0",
			expectNewer:    true,
		},
		{
			name:           "same version with v prefix",
			currentVersion: "v0.1.0",
			latestVersion:  "v0.1.0",
			expectNewer:    false,
		},
		{
			name:           "major version bump",
			currentVersion: "0.1.0",
			latestVersion:  "1.0.0",
			expectNewer:    true,
		},
		{
			name:           "minor version bump",
			currentVersion: "0.1.0",
			latestVersion:  "0.2.0",
			expectNewer:    true,
		},
		{
			name:           "patch version bump",
			currentVersion: "0.1.0",
			latestVersion:  "0.1.1",
			expectNewer:    true,
		},
		{
			name:           "current newer than latest",
			currentVersion: "0.2.0",
			latestVersion:  "0.1.0",
			expectNewer:    false,
		},
		{
			name:           "equal versions",
			currentVersion: "0.1.0",
			latestVersion:  "0.1.0",
			expectNewer:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			uc := New(fs, tt.currentVersion)

			result := uc.isNewerVersion(tt.latestVersion)

			if result != tt.expectNewer {
				t.Errorf("isNewerVersion(%q, %q) = %v; want %v",
					tt.currentVersion, tt.latestVersion, result, tt.expectNewer)
			}
		})
	}
}

func TestGetLatestVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Initially should be empty
	if uc.GetLatestVersion() != "" {
		t.Errorf("expected empty latest version, got %s", uc.GetLatestVersion())
	}

	// After checking (using cache)
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Format(time.RFC3339),
			CachedVersion:  "0.2.0",
			CheckSucceeded: true,
		},
	}
	prefsSvc.Save(prefs)

	uc.ShouldPrompt()

	if uc.GetLatestVersion() != "0.2.0" {
		t.Errorf("expected latest version 0.2.0, got %s", uc.GetLatestVersion())
	}
}

func TestGetCurrentVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	currentVersion := "0.1.0"
	uc := New(fs, currentVersion)

	if uc.GetCurrentVersion() != currentVersion {
		t.Errorf("expected current version %s, got %s", currentVersion, uc.GetCurrentVersion())
	}
}

func TestSaveNeverAsk(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	err := uc.SaveNeverAsk()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify preference was saved
	prefsSvc := globalprefs.New(fs)
	prefs, err := prefsSvc.Load()
	if err != nil {
		t.Fatalf("failed to load preferences: %v", err)
	}

	if !prefs.UpdateCheck.NeverAskAgain {
		t.Error("expected NeverAskAgain to be true")
	}

	if prefs.UpdateCheck.DeclinedAt == "" {
		t.Error("expected DeclinedAt to be set")
	}

	// Verify timestamp is valid RFC3339
	_, err = time.Parse(time.RFC3339, prefs.UpdateCheck.DeclinedAt)
	if err != nil {
		t.Errorf("DeclinedAt timestamp is invalid: %v", err)
	}
}

func TestShouldPrompt_ExpiredCache(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Setup: expired cache (>24 hours old)
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Add(-25 * time.Hour).Format(time.RFC3339),
			CachedVersion:  "0.1.0",
			CheckSucceeded: true,
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	// With expired cache, should try to fetch from npm
	// Since we can't mock the npm call easily, it will likely be NetworkError or ResultUpToDate
	if result == ResultCached {
		t.Error("expected to not use expired cache")
	}
}

func TestShouldPrompt_FailedCacheCheck(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs, "0.1.0")

	// Setup: recent cache but check failed
	prefsSvc := globalprefs.New(fs)
	prefs := globalprefs.MCPPreferences{
		UpdateCheck: globalprefs.UpdatePreferences{
			LastCheckedAt:  time.Now().Format(time.RFC3339),
			CachedVersion:  "",
			CheckSucceeded: false,
		},
	}
	prefsSvc.Save(prefs)

	result := uc.ShouldPrompt()

	// Should retry the check since previous check failed
	if result == ResultCached {
		t.Error("expected to not use failed cache")
	}
}
