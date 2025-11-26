package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProfile_GetToken(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		want    string
	}{
		{
			name: "bearer token provided",
			profile: Profile{
				BearerToken: "token-12345:secret-67890",
			},
			want: "token-12345:secret-67890",
		},
		{
			name: "access key and secret key provided",
			profile: Profile{
				AccessKey: "token-abcde",
				SecretKey: "fghij-klmno",
			},
			want: "token-abcde:fghij-klmno",
		},
		{
			name: "bearer token takes precedence",
			profile: Profile{
				BearerToken: "token-bearer",
				AccessKey:   "token-access",
				SecretKey:   "secret-key",
			},
			want: "token-bearer",
		},
		{
			name:    "no credentials provided",
			profile: Profile{},
			want:    "",
		},
		{
			name: "only access key provided",
			profile: Profile{
				AccessKey: "token-access",
			},
			want: "",
		},
		{
			name: "only secret key provided",
			profile: Profile{
				SecretKey: "secret-only",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.GetToken(); got != tt.want {
				t.Errorf("Profile.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				CurrentProfile: "default",
				Profiles: []Profile{
					{Name: "default", URL: "https://rancher.example.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "no profiles",
			config: Config{
				CurrentProfile: "default",
				Profiles:       []Profile{},
			},
			wantErr: true,
			errMsg:  "no profiles configured",
		},
		{
			name: "current profile not found",
			config: Config{
				CurrentProfile: "nonexistent",
				Profiles: []Profile{
					{Name: "default", URL: "https://rancher.example.com"},
				},
			},
			wantErr: true,
			errMsg:  "profile 'nonexistent' not found",
		},
		{
			name: "multiple profiles with valid current",
			config: Config{
				CurrentProfile: "prod",
				Profiles: []Profile{
					{Name: "default", URL: "https://rancher.example.com"},
					{Name: "prod", URL: "https://rancher-prod.example.com"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Config.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestConfig_GetCurrentProfile(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantProfile *Profile
		wantErr     bool
	}{
		{
			name: "profile found",
			config: Config{
				CurrentProfile: "default",
				Profiles: []Profile{
					{Name: "default", URL: "https://rancher.example.com"},
				},
			},
			wantProfile: &Profile{Name: "default", URL: "https://rancher.example.com"},
			wantErr:     false,
		},
		{
			name: "profile not found",
			config: Config{
				CurrentProfile: "missing",
				Profiles: []Profile{
					{Name: "default", URL: "https://rancher.example.com"},
				},
			},
			wantProfile: nil,
			wantErr:     true,
		},
		{
			name: "multiple profiles, correct one returned",
			config: Config{
				CurrentProfile: "prod",
				Profiles: []Profile{
					{Name: "default", URL: "https://rancher.example.com"},
					{Name: "prod", URL: "https://rancher-prod.example.com"},
					{Name: "dev", URL: "https://rancher-dev.example.com"},
				},
			},
			wantProfile: &Profile{Name: "prod", URL: "https://rancher-prod.example.com"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.GetCurrentProfile()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.GetCurrentProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name != tt.wantProfile.Name || got.URL != tt.wantProfile.URL {
					t.Errorf("Config.GetCurrentProfile() = %v, want %v", got, tt.wantProfile)
				}
			}
		})
	}
}

func TestConfig_GetRefreshInterval(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   time.Duration
	}{
		{
			name:   "default when empty",
			config: Config{RefreshInterval: ""},
			want:   5 * time.Second,
		},
		{
			name:   "custom valid duration",
			config: Config{RefreshInterval: "10s"},
			want:   10 * time.Second,
		},
		{
			name:   "minutes",
			config: Config{RefreshInterval: "2m"},
			want:   2 * time.Minute,
		},
		{
			name:   "invalid duration falls back to default",
			config: Config{RefreshInterval: "invalid"},
			want:   5 * time.Second,
		},
		{
			name:   "combined units",
			config: Config{RefreshInterval: "1m30s"},
			want:   90 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetRefreshInterval(); got != tt.want {
				t.Errorf("Config.GetRefreshInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "r9s-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "config.yaml")

	config := &Config{
		CurrentProfile: "test",
		Profiles: []Profile{
			{Name: "test", URL: "https://test.example.com", BearerToken: "token-test"},
		},
		RefreshInterval: "5s",
		LogLevel:        "info",
	}

	// Test save
	err = config.Save(configFile)
	if err != nil {
		t.Errorf("Config.Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify file permissions (should be 0600)
	info, err := os.Stat(configFile)
	if err != nil {
		t.Errorf("Failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Config file permissions = %v, want 0600", info.Mode().Perm())
	}

	// Verify content by loading it back
	loadedConfig, err := Load(configFile, "")
	if err != nil {
		t.Errorf("Failed to load saved config: %v", err)
	}

	if loadedConfig.CurrentProfile != config.CurrentProfile {
		t.Errorf("Loaded CurrentProfile = %v, want %v", loadedConfig.CurrentProfile, config.CurrentProfile)
	}
	if len(loadedConfig.Profiles) != len(config.Profiles) {
		t.Errorf("Loaded Profiles count = %v, want %v", len(loadedConfig.Profiles), len(config.Profiles))
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	// Load should create default config and exit
	// Since it calls os.Exit(0), we can't test it directly in a unit test
	// This is a known limitation - createDefaultConfig needs refactoring for testability
	// For now, we just document this behavior
	t.Skip("Load() with non-existent file calls os.Exit(0), cannot test directly")
}

func TestLoad_ValidFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "r9s-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create a valid config file
	validYaml := `currentProfile: default
profiles:
  - name: default
    url: https://rancher.example.com
    bearerToken: token-12345:secret-67890
    insecure: false
refreshInterval: 5s
logLevel: info
`
	err = os.WriteFile(configFile, []byte(validYaml), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	config, err := Load(configFile, "")
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}

	// Verify loaded values
	if config.CurrentProfile != "default" {
		t.Errorf("CurrentProfile = %v, want 'default'", config.CurrentProfile)
	}
	if len(config.Profiles) != 1 {
		t.Errorf("Profiles count = %v, want 1", len(config.Profiles))
	}
	if config.Profiles[0].URL != "https://rancher.example.com" {
		t.Errorf("Profile URL = %v, want 'https://rancher.example.com'", config.Profiles[0].URL)
	}
	if config.RefreshInterval != "5s" {
		t.Errorf("RefreshInterval = %v, want '5s'", config.RefreshInterval)
	}
}

func TestLoad_ProfileOverride(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "r9s-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create config with multiple profiles
	validYaml := `currentProfile: default
profiles:
  - name: default
    url: https://rancher.example.com
  - name: prod
    url: https://rancher-prod.example.com
refreshInterval: 5s
logLevel: info
`
	err = os.WriteFile(configFile, []byte(validYaml), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load with profile override
	config, err := Load(configFile, "prod")
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}

	// Verify profile was overridden
	if config.CurrentProfile != "prod" {
		t.Errorf("CurrentProfile = %v, want 'prod'", config.CurrentProfile)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "r9s-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create invalid YAML
	invalidYaml := `this is not valid yaml: [[[`
	err = os.WriteFile(configFile, []byte(invalidYaml), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load should return error
	_, err = Load(configFile, "")
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}
