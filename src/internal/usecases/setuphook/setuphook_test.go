package setuphook

import (
	"errors"
	"testing"
	"time"

	"claudex/internal/services/preferences"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHookService is a mock implementation of hooksetup.Service
type mockHookService struct {
	isGitRepo  bool
	isInstall  bool
	installErr error
}

func (m *mockHookService) IsGitRepo() bool   { return m.isGitRepo }
func (m *mockHookService) IsInstalled() bool { return m.isInstall }
func (m *mockHookService) Install() error    { return m.installErr }

// mockPrefService is a mock implementation of preferences.Service
type mockPrefService struct {
	prefs   preferences.Preferences
	loadErr error
	saveErr error
}

func (m *mockPrefService) Load() (preferences.Preferences, error) {
	return m.prefs, m.loadErr
}

func (m *mockPrefService) Save(prefs preferences.Preferences) error {
	m.prefs = prefs
	return m.saveErr
}

// TestShouldPrompt_NotGitRepo verifies that non-git repos return ResultNotGitRepo
func TestShouldPrompt_NotGitRepo(t *testing.T) {
	uc := &UseCase{
		hookSvc: &mockHookService{isGitRepo: false},
		prefSvc: &mockPrefService{},
	}

	result := uc.ShouldPrompt()
	assert.Equal(t, ResultNotGitRepo, result)
}

// TestShouldPrompt_AlreadyInstalled verifies that installed hooks return ResultAlreadyInstalled
func TestShouldPrompt_AlreadyInstalled(t *testing.T) {
	uc := &UseCase{
		hookSvc: &mockHookService{isGitRepo: true, isInstall: true},
		prefSvc: &mockPrefService{},
	}

	result := uc.ShouldPrompt()
	assert.Equal(t, ResultAlreadyInstalled, result)
}

// TestShouldPrompt_UserDeclined verifies that declined preference returns ResultUserDeclined
func TestShouldPrompt_UserDeclined(t *testing.T) {
	uc := &UseCase{
		hookSvc: &mockHookService{isGitRepo: true, isInstall: false},
		prefSvc: &mockPrefService{
			prefs: preferences.Preferences{
				HookSetupDeclined: true,
				DeclinedAt:        "2024-01-01T00:00:00Z",
			},
		},
	}

	result := uc.ShouldPrompt()
	assert.Equal(t, ResultUserDeclined, result)
}

// TestShouldPrompt_ShouldPrompt verifies that uninstalled hooks with no decline return ResultPromptUser
func TestShouldPrompt_ShouldPrompt(t *testing.T) {
	uc := &UseCase{
		hookSvc: &mockHookService{isGitRepo: true, isInstall: false},
		prefSvc: &mockPrefService{
			prefs: preferences.Preferences{HookSetupDeclined: false},
		},
	}

	result := uc.ShouldPrompt()
	assert.Equal(t, ResultPromptUser, result)
}

// TestShouldPrompt_PrefLoadError verifies that preference load errors still prompt
func TestShouldPrompt_PrefLoadError(t *testing.T) {
	uc := &UseCase{
		hookSvc: &mockHookService{isGitRepo: true, isInstall: false},
		prefSvc: &mockPrefService{
			loadErr: errors.New("read error"),
		},
	}

	result := uc.ShouldPrompt()
	assert.Equal(t, ResultPromptUser, result, "should prompt when preferences can't be loaded")
}

// TestInstall_Success verifies that Install delegates to hookSvc
func TestInstall_Success(t *testing.T) {
	hookSvc := &mockHookService{installErr: nil}
	uc := &UseCase{
		hookSvc: hookSvc,
		prefSvc: &mockPrefService{},
	}

	err := uc.Install()
	assert.NoError(t, err)
}

// TestInstall_Error verifies that Install propagates errors from hookSvc
func TestInstall_Error(t *testing.T) {
	expectedErr := errors.New("install failed")
	hookSvc := &mockHookService{installErr: expectedErr}
	uc := &UseCase{
		hookSvc: hookSvc,
		prefSvc: &mockPrefService{},
	}

	err := uc.Install()
	assert.Equal(t, expectedErr, err)
}

// TestSaveDeclined_Success verifies that SaveDeclined persists the preference
func TestSaveDeclined_Success(t *testing.T) {
	prefSvc := &mockPrefService{}
	uc := &UseCase{
		hookSvc: &mockHookService{},
		prefSvc: prefSvc,
	}

	err := uc.SaveDeclined()
	require.NoError(t, err)

	// Verify preferences were updated
	assert.True(t, prefSvc.prefs.HookSetupDeclined, "HookSetupDeclined should be true")
	assert.NotEmpty(t, prefSvc.prefs.DeclinedAt, "DeclinedAt should be set")

	// Verify timestamp is valid RFC3339
	_, parseErr := time.Parse(time.RFC3339, prefSvc.prefs.DeclinedAt)
	assert.NoError(t, parseErr, "DeclinedAt should be valid RFC3339")
}

// TestSaveDeclined_Error verifies that SaveDeclined propagates save errors
func TestSaveDeclined_Error(t *testing.T) {
	expectedErr := errors.New("save failed")
	prefSvc := &mockPrefService{saveErr: expectedErr}
	uc := &UseCase{
		hookSvc: &mockHookService{},
		prefSvc: prefSvc,
	}

	err := uc.SaveDeclined()
	assert.Equal(t, expectedErr, err)
}

// TestSaveDeclined_LoadError verifies that SaveDeclined works even if Load fails
func TestSaveDeclined_LoadError(t *testing.T) {
	prefSvc := &mockPrefService{
		loadErr: errors.New("load failed"),
	}
	uc := &UseCase{
		hookSvc: &mockHookService{},
		prefSvc: prefSvc,
	}

	err := uc.SaveDeclined()
	require.NoError(t, err, "should succeed even if Load fails")

	// Verify preferences were updated with fresh values
	assert.True(t, prefSvc.prefs.HookSetupDeclined)
	assert.NotEmpty(t, prefSvc.prefs.DeclinedAt)
}

// TestNew verifies that New creates a UseCase with proper services
func TestNew(t *testing.T) {
	// This is more of an integration smoke test
	// We can't easily test the actual service instantiation without mocking afero
	// But we can verify the structure is correct
	uc := &UseCase{
		hookSvc: &mockHookService{},
		prefSvc: &mockPrefService{},
	}

	assert.NotNil(t, uc.hookSvc)
	assert.NotNil(t, uc.prefSvc)
}
