package updatecheck

import (
	"log"
	"strings"
	"time"

	"claudex/internal/services/globalprefs"
	"claudex/internal/services/npmregistry"
	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
)

const packageName = "@claudex/cli"

type UseCase struct {
	npmSvc         *npmregistry.Client
	prefsSvc       globalprefs.Service
	currentVersion string
	latestVersion  string
}

func New(fs afero.Fs, currentVersion string) *UseCase {
	return &UseCase{
		npmSvc:         npmregistry.New(),
		prefsSvc:       globalprefs.New(fs),
		currentVersion: currentVersion,
	}
}

// ShouldPrompt checks if user should be prompted for update
func (uc *UseCase) ShouldPrompt() Result {
	// 1. Check if user opted out
	prefs, _ := uc.prefsSvc.Load()
	if prefs.UpdateCheck.NeverAskAgain {
		return ResultNeverAskAgain
	}

	// 2. Check cache validity
	if prefs.IsUpdateCacheValid() && prefs.UpdateCheck.CheckSucceeded {
		uc.latestVersion = prefs.UpdateCheck.CachedVersion
		// Compare cached version
		if !uc.isNewerVersion(uc.latestVersion) {
			return ResultCached
		}
		return ResultPromptUser
	}

	// 3. Fetch from npm registry
	latest, err := uc.npmSvc.GetLatestVersion(packageName)
	if err != nil {
		log.Printf("Update check failed: %v", err)
		// Update cache as failed
		prefs.SetUpdateCache("", false)
		uc.prefsSvc.Save(prefs)
		return ResultNetworkError
	}

	// 4. Cache the result
	uc.latestVersion = latest
	prefs.SetUpdateCache(latest, true)
	uc.prefsSvc.Save(prefs)

	// 5. Compare versions
	if !uc.isNewerVersion(latest) {
		return ResultUpToDate
	}

	return ResultPromptUser
}

// isNewerVersion returns true if latest > current
func (uc *UseCase) isNewerVersion(latest string) bool {
	// Clean version strings (remove 'v' prefix if present)
	current := strings.TrimPrefix(uc.currentVersion, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Handle dirty versions (e.g., "0.1.2-dirty")
	current = strings.Split(current, "-")[0]

	currentV, err := semver.NewVersion(current)
	if err != nil {
		log.Printf("Failed to parse current version %q: %v", current, err)
		return false
	}

	latestV, err := semver.NewVersion(latest)
	if err != nil {
		log.Printf("Failed to parse latest version %q: %v", latest, err)
		return false
	}

	return latestV.GreaterThan(currentV)
}

// GetLatestVersion returns the latest version found
func (uc *UseCase) GetLatestVersion() string {
	return uc.latestVersion
}

// GetCurrentVersion returns the current version
func (uc *UseCase) GetCurrentVersion() string {
	return uc.currentVersion
}

// SaveNeverAsk saves the user's preference to never ask again
func (uc *UseCase) SaveNeverAsk() error {
	prefs, _ := uc.prefsSvc.Load()
	prefs.UpdateCheck.NeverAskAgain = true
	prefs.UpdateCheck.DeclinedAt = time.Now().Format(time.RFC3339)
	return uc.prefsSvc.Save(prefs)
}
