package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// TestLoad_EmptyConfig_ReturnsDefaults verifies that an empty config file returns default feature values
func TestLoad_EmptyConfig_ReturnsDefaults(t *testing.T) {
	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	// Create empty config file
	err := afero.WriteFile(fs, configPath, []byte(""), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// Assert defaults
	require.True(t, cfg.Features.AutodocSessionProgress, "AutodocSessionProgress should default to true")
	require.True(t, cfg.Features.AutodocSessionEnd, "AutodocSessionEnd should default to true")
	require.Equal(t, 5, cfg.Features.AutodocFrequency, "AutodocFrequency should default to 5")

	// Assert ProcessProtection defaults
	require.Equal(t, runtime.NumCPU()*2, cfg.Features.ProcessProtection.MaxProcesses)
	require.Equal(t, 5, cfg.Features.ProcessProtection.RateLimitPerSecond)
	require.Equal(t, 300, cfg.Features.ProcessProtection.TimeoutSeconds)
}

// TestLoad_NoConfigFile_ReturnsDefaults verifies that missing config file returns defaults
func TestLoad_NoConfigFile_ReturnsDefaults(t *testing.T) {
	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	// Don't create config file

	// Load config
	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// Assert defaults
	require.True(t, cfg.Features.AutodocSessionProgress)
	require.True(t, cfg.Features.AutodocSessionEnd)
	require.Equal(t, 5, cfg.Features.AutodocFrequency)
	require.Empty(t, cfg.Doc)
	require.False(t, cfg.NoOverwrite)

	// Assert ProcessProtection defaults
	require.Equal(t, runtime.NumCPU()*2, cfg.Features.ProcessProtection.MaxProcesses)
	require.Equal(t, 5, cfg.Features.ProcessProtection.RateLimitPerSecond)
	require.Equal(t, 300, cfg.Features.ProcessProtection.TimeoutSeconds)
}

// TestLoad_PartialFeatures_UsesDefaults verifies partial [features] section uses defaults for missing fields
func TestLoad_PartialFeatures_UsesDefaults(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Features
	}{
		{
			name: "Only autodoc_session_progress",
			content: `[features]
autodoc_session_progress = false`,
			expected: Features{
				AutodocSessionProgress: false,
				AutodocSessionEnd:      true, // default
				AutodocFrequency:       5,    // default
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
		{
			name: "Only autodoc_session_end",
			content: `[features]
autodoc_session_end = false`,
			expected: Features{
				AutodocSessionProgress: true, // default
				AutodocSessionEnd:      false,
				AutodocFrequency:       5, // default
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
		{
			name: "Only autodoc_frequency",
			content: `[features]
autodoc_frequency = 10`,
			expected: Features{
				AutodocSessionProgress: true, // default
				AutodocSessionEnd:      true, // default
				AutodocFrequency:       10,
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
		{
			name: "Two fields set",
			content: `[features]
autodoc_session_progress = false
autodoc_frequency = 15`,
			expected: Features{
				AutodocSessionProgress: false,
				AutodocSessionEnd:      true, // default
				AutodocFrequency:       15,
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			configPath := "/test/.claudex/config.toml"

			err := afero.WriteFile(fs, configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			cfg, err := Load(fs, configPath)
			require.NoError(t, err)

			require.Equal(t, tt.expected.AutodocSessionProgress, cfg.Features.AutodocSessionProgress)
			require.Equal(t, tt.expected.AutodocSessionEnd, cfg.Features.AutodocSessionEnd)
			require.Equal(t, tt.expected.AutodocFrequency, cfg.Features.AutodocFrequency)
		})
	}
}

// TestLoad_FullFeatures_ParsesAllValues verifies full [features] section parses all values correctly
func TestLoad_FullFeatures_ParsesAllValues(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Features
	}{
		{
			name: "All features disabled",
			content: `[features]
autodoc_session_progress = false
autodoc_session_end = false
autodoc_frequency = 1`,
			expected: Features{
				AutodocSessionProgress: false,
				AutodocSessionEnd:      false,
				AutodocFrequency:       1,
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
		{
			name: "All features enabled with custom frequency",
			content: `[features]
autodoc_session_progress = true
autodoc_session_end = true
autodoc_frequency = 20`,
			expected: Features{
				AutodocSessionProgress: true,
				AutodocSessionEnd:      true,
				AutodocFrequency:       20,
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
		{
			name: "Mixed settings",
			content: `[features]
autodoc_session_progress = false
autodoc_session_end = true
autodoc_frequency = 10`,
			expected: Features{
				AutodocSessionProgress: false,
				AutodocSessionEnd:      true,
				AutodocFrequency:       10,
				ProcessProtection: ProcessProtection{
					MaxProcesses:       runtime.NumCPU() * 2,
					RateLimitPerSecond: 5,
					TimeoutSeconds:     300,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			configPath := "/test/.claudex/config.toml"

			err := afero.WriteFile(fs, configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			cfg, err := Load(fs, configPath)
			require.NoError(t, err)

			require.Equal(t, tt.expected, cfg.Features)
		})
	}
}

// TestLoad_InvalidFrequency_StillLoads verifies config loads even with problematic frequency values
func TestLoad_InvalidFrequency_StillLoads(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantValue int
	}{
		{
			name: "Negative frequency",
			content: `[features]
autodoc_frequency = -5`,
			wantValue: -5, // TOML will parse it, validation happens elsewhere
		},
		{
			name: "Zero frequency",
			content: `[features]
autodoc_frequency = 0`,
			wantValue: 0, // TOML will parse it, validation happens elsewhere
		},
		{
			name: "Large frequency",
			content: `[features]
autodoc_frequency = 1000`,
			wantValue: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			configPath := "/test/.claudex/config.toml"

			err := afero.WriteFile(fs, configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			cfg, err := Load(fs, configPath)
			require.NoError(t, err)

			require.Equal(t, tt.wantValue, cfg.Features.AutodocFrequency)
		})
	}
}

// TestLoad_WithOtherConfig_PreservesFeatures verifies features work alongside other config sections
func TestLoad_WithOtherConfig_PreservesFeatures(t *testing.T) {
	content := `
doc = ["docs/api.md", "docs/guide.md"]
no_overwrite = true

[features]
autodoc_session_progress = false
autodoc_session_end = true
autodoc_frequency = 15
`

	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	err := afero.WriteFile(fs, configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// Assert non-features config
	require.Equal(t, []string{"docs/api.md", "docs/guide.md"}, cfg.Doc)
	require.True(t, cfg.NoOverwrite)

	// Assert features config
	require.False(t, cfg.Features.AutodocSessionProgress)
	require.True(t, cfg.Features.AutodocSessionEnd)
	require.Equal(t, 15, cfg.Features.AutodocFrequency)
}

// TestLoad_MalformedTOML_ReturnsError verifies malformed TOML returns an error
func TestLoad_MalformedTOML_ReturnsError(t *testing.T) {
	content := `[features
autodoc_session_progress = false` // Missing closing bracket

	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	err := afero.WriteFile(fs, configPath, []byte(content), 0644)
	require.NoError(t, err)

	_, err = Load(fs, configPath)
	require.Error(t, err, "malformed TOML should return error")
}

// TestLoad_FromTestdataFile verifies loading from actual testdata config file
func TestLoad_FromTestdataFile(t *testing.T) {
	fs := afero.NewOsFs()
	configPath := "../../../testdata/configs/features.toml"

	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// This test verifies the actual testdata file content
	// Expected values based on testdata/configs/features.toml
	require.False(t, cfg.Features.AutodocSessionProgress)
	require.True(t, cfg.Features.AutodocSessionEnd)
	require.Equal(t, 10, cfg.Features.AutodocFrequency)
}

// TestLoad_ProcessProtection_Defaults verifies ProcessProtection defaults
func TestLoad_ProcessProtection_Defaults(t *testing.T) {
	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	// Create empty config file
	err := afero.WriteFile(fs, configPath, []byte(""), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// Assert ProcessProtection defaults
	require.Equal(t, runtime.NumCPU()*2, cfg.Features.ProcessProtection.MaxProcesses)
	require.Equal(t, 5, cfg.Features.ProcessProtection.RateLimitPerSecond)
	require.Equal(t, 300, cfg.Features.ProcessProtection.TimeoutSeconds)
}

// TestLoad_ProcessProtection_FromTOML verifies ProcessProtection loads from TOML
func TestLoad_ProcessProtection_FromTOML(t *testing.T) {
	content := `[features.process_protection]
max_processes = 25
rate_limit_per_second = 10
timeout_seconds = 600`

	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	err := afero.WriteFile(fs, configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// Assert custom values
	require.Equal(t, 25, cfg.Features.ProcessProtection.MaxProcesses)
	require.Equal(t, 10, cfg.Features.ProcessProtection.RateLimitPerSecond)
	require.Equal(t, 600, cfg.Features.ProcessProtection.TimeoutSeconds)
}

// TestLoad_ProcessProtection_PartialTOML verifies partial ProcessProtection uses defaults
func TestLoad_ProcessProtection_PartialTOML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected ProcessProtection
	}{
		{
			name: "Only max_processes",
			content: `[features.process_protection]
max_processes = 100`,
			expected: ProcessProtection{
				MaxProcesses:       100,
				RateLimitPerSecond: 5,   // default
				TimeoutSeconds:     300, // default
			},
		},
		{
			name: "Only rate_limit_per_second",
			content: `[features.process_protection]
rate_limit_per_second = 2`,
			expected: ProcessProtection{
				MaxProcesses:       runtime.NumCPU() * 2, // default
				RateLimitPerSecond: 2,
				TimeoutSeconds:     300, // default
			},
		},
		{
			name: "Two fields",
			content: `[features.process_protection]
max_processes = 75
timeout_seconds = 120`,
			expected: ProcessProtection{
				MaxProcesses:       75,
				RateLimitPerSecond: 5,   // default
				TimeoutSeconds:     120,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			configPath := "/test/.claudex/config.toml"

			err := afero.WriteFile(fs, configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			cfg, err := Load(fs, configPath)
			require.NoError(t, err)

			require.Equal(t, tt.expected, cfg.Features.ProcessProtection)
		})
	}
}

// TestLoad_ProcessProtection_EnvOverrides verifies environment variable overrides
func TestLoad_ProcessProtection_EnvOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected ProcessProtection
	}{
		{
			name: "CLAUDEX_MAX_PROCESSES override",
			envVars: map[string]string{
				"CLAUDEX_MAX_PROCESSES": "75",
			},
			expected: ProcessProtection{
				MaxProcesses:       75,
				RateLimitPerSecond: 5,
				TimeoutSeconds:     300,
			},
		},
		{
			name: "CLAUDEX_RATE_LIMIT override",
			envVars: map[string]string{
				"CLAUDEX_RATE_LIMIT": "10",
			},
			expected: ProcessProtection{
				MaxProcesses:       runtime.NumCPU() * 2,
				RateLimitPerSecond: 10,
				TimeoutSeconds:     300,
			},
		},
		{
			name: "CLAUDEX_TIMEOUT override",
			envVars: map[string]string{
				"CLAUDEX_TIMEOUT": "600",
			},
			expected: ProcessProtection{
				MaxProcesses:       runtime.NumCPU() * 2,
				RateLimitPerSecond: 5,
				TimeoutSeconds:     600,
			},
		},
		{
			name: "All env vars override",
			envVars: map[string]string{
				"CLAUDEX_MAX_PROCESSES": "100",
				"CLAUDEX_RATE_LIMIT":    "15",
				"CLAUDEX_TIMEOUT":       "900",
			},
			expected: ProcessProtection{
				MaxProcesses:       100,
				RateLimitPerSecond: 15,
				TimeoutSeconds:     900,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tt.envVars {
				originalValue := os.Getenv(key)
				os.Setenv(key, val)
				defer func(k, v string) {
					if v == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, v)
					}
				}(key, originalValue)
			}

			fs := afero.NewMemMapFs()
			configPath := "/test/.claudex/config.toml"

			// Create empty config file
			err := afero.WriteFile(fs, configPath, []byte(""), 0644)
			require.NoError(t, err)

			cfg, err := Load(fs, configPath)
			require.NoError(t, err)

			require.Equal(t, tt.expected, cfg.Features.ProcessProtection)
		})
	}
}

// TestLoad_ProcessProtection_EnvOverridesToml verifies env vars override TOML values
func TestLoad_ProcessProtection_EnvOverridesToml(t *testing.T) {
	content := `[features.process_protection]
max_processes = 25
rate_limit_per_second = 3
timeout_seconds = 120`

	// Set env vars to override TOML
	os.Setenv("CLAUDEX_MAX_PROCESSES", "100")
	os.Setenv("CLAUDEX_TIMEOUT", "600")
	defer func() {
		os.Unsetenv("CLAUDEX_MAX_PROCESSES")
		os.Unsetenv("CLAUDEX_TIMEOUT")
	}()

	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	err := afero.WriteFile(fs, configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// MaxProcesses and TimeoutSeconds should be from env vars
	require.Equal(t, 100, cfg.Features.ProcessProtection.MaxProcesses)
	require.Equal(t, 600, cfg.Features.ProcessProtection.TimeoutSeconds)

	// RateLimitPerSecond should be from TOML
	require.Equal(t, 3, cfg.Features.ProcessProtection.RateLimitPerSecond)
}

// TestLoad_ProcessProtection_InvalidEnvVars verifies invalid env vars are ignored
func TestLoad_ProcessProtection_InvalidEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "Non-numeric max_processes",
			envVars: map[string]string{
				"CLAUDEX_MAX_PROCESSES": "not-a-number",
			},
		},
		{
			name: "Empty env var",
			envVars: map[string]string{
				"CLAUDEX_RATE_LIMIT": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tt.envVars {
				originalValue := os.Getenv(key)
				if val != "" {
					os.Setenv(key, val)
				}
				defer func(k, v string) {
					if v == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, v)
					}
				}(key, originalValue)
			}

			fs := afero.NewMemMapFs()
			configPath := "/test/.claudex/config.toml"

			err := afero.WriteFile(fs, configPath, []byte(""), 0644)
			require.NoError(t, err)

			cfg, err := Load(fs, configPath)
			require.NoError(t, err)

			// Should use defaults when env vars are invalid
			expected := ProcessProtection{
				MaxProcesses:       runtime.NumCPU() * 2,
				RateLimitPerSecond: 5,
				TimeoutSeconds:     300,
			}

			require.Equal(t, expected, cfg.Features.ProcessProtection)
		})
	}
}

// TestLoad_ProcessProtection_WithOtherFeatures verifies ProcessProtection works with other features
func TestLoad_ProcessProtection_WithOtherFeatures(t *testing.T) {
	content := `[features]
autodoc_session_progress = false
autodoc_session_end = true
autodoc_frequency = 15

[features.process_protection]
max_processes = 75
rate_limit_per_second = 8
timeout_seconds = 450`

	fs := afero.NewMemMapFs()
	configPath := "/test/.claudex/config.toml"

	err := afero.WriteFile(fs, configPath, []byte(content), 0644)
	require.NoError(t, err)

	cfg, err := Load(fs, configPath)
	require.NoError(t, err)

	// Assert other features
	require.False(t, cfg.Features.AutodocSessionProgress)
	require.True(t, cfg.Features.AutodocSessionEnd)
	require.Equal(t, 15, cfg.Features.AutodocFrequency)

	// Assert ProcessProtection
	require.Equal(t, 75, cfg.Features.ProcessProtection.MaxProcesses)
	require.Equal(t, 8, cfg.Features.ProcessProtection.RateLimitPerSecond)
	require.Equal(t, 450, cfg.Features.ProcessProtection.TimeoutSeconds)
}
