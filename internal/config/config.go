package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config holds all CLI configuration
type Config struct {
	Auth     AuthConfig     `yaml:"auth" mapstructure:"auth"`
	Defaults DefaultsConfig `yaml:"defaults" mapstructure:"defaults"`
	Output   OutputConfig   `yaml:"output" mapstructure:"output"`
}

// AuthConfig holds OAuth credentials and tokens
type AuthConfig struct {
	ClientID     string    `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string    `yaml:"client_secret" mapstructure:"client_secret"`
	AccessToken  string    `yaml:"access_token" mapstructure:"access_token"`
	RefreshToken string    `yaml:"refresh_token" mapstructure:"refresh_token"`
	TokenExpiry  time.Time `yaml:"token_expiry" mapstructure:"token_expiry"`
}

// DefaultsConfig holds default values for common options
type DefaultsConfig struct {
	TeamID     string `yaml:"team_id" mapstructure:"team_id"`
	ProjectID  string `yaml:"project_id" mapstructure:"project_id"`
	DataCenter string `yaml:"data_center" mapstructure:"data_center"` // com, eu, in, au, jp
}

// OutputConfig holds output formatting preferences
type OutputConfig struct {
	Format string `yaml:"format" mapstructure:"format"` // json, table, yaml, plain
	Color  bool   `yaml:"color" mapstructure:"color"`
}

var (
	cfg        *Config
	configPath string
)

// GetConfigDir returns the config directory path
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".zs"
	}
	return filepath.Join(home, ".zs")
}

// GetConfigPath returns the full config file path
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.yaml")
}

// Init initializes the configuration system
func Init() error {
	configDir := GetConfigDir()
	configPath = GetConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("auth.client_id", "")
	viper.SetDefault("auth.client_secret", "")
	viper.SetDefault("auth.access_token", "")
	viper.SetDefault("auth.refresh_token", "")
	viper.SetDefault("defaults.team_id", "")
	viper.SetDefault("defaults.project_id", "")
	viper.SetDefault("defaults.data_center", "com")
	viper.SetDefault("output.format", "table")
	viper.SetDefault("output.color", true)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
		// Config file doesn't exist, create it with defaults
		if err := Save(); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	// Unmarshal into struct
	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		cfg = &Config{
			Defaults: DefaultsConfig{
				DataCenter: "com",
			},
			Output: OutputConfig{
				Format: "table",
				Color:  true,
			},
		}
	}
	return cfg
}

// Save writes the current configuration to disk
func Save() error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update viper values from struct
	c := Get()
	viper.Set("auth.client_id", c.Auth.ClientID)
	viper.Set("auth.client_secret", c.Auth.ClientSecret)
	viper.Set("auth.access_token", c.Auth.AccessToken)
	viper.Set("auth.refresh_token", c.Auth.RefreshToken)
	viper.Set("auth.token_expiry", c.Auth.TokenExpiry)
	viper.Set("defaults.team_id", c.Defaults.TeamID)
	viper.Set("defaults.project_id", c.Defaults.ProjectID)
	viper.Set("defaults.data_center", c.Defaults.DataCenter)
	viper.Set("output.format", c.Output.Format)
	viper.Set("output.color", c.Output.Color)

	// Write to file
	return viper.WriteConfigAs(GetConfigPath())
}

// SaveConfig saves a config struct to disk (used during initial setup)
func SaveConfig(c *Config) error {
	cfg = c
	return Save()
}

// UpdateAuth updates authentication tokens and saves
func UpdateAuth(accessToken, refreshToken string, expiry time.Time) error {
	c := Get()
	c.Auth.AccessToken = accessToken
	c.Auth.RefreshToken = refreshToken
	c.Auth.TokenExpiry = expiry
	return Save()
}

// SetCredentials sets OAuth client credentials
func SetCredentials(clientID, clientSecret string) error {
	c := Get()
	c.Auth.ClientID = clientID
	c.Auth.ClientSecret = clientSecret
	return Save()
}

// SetDefault sets a default value
func SetDefault(key, value string) error {
	c := Get()
	switch key {
	case "team_id", "team":
		c.Defaults.TeamID = value
	case "project_id", "project":
		c.Defaults.ProjectID = value
	case "data_center", "dc":
		c.Defaults.DataCenter = value
	default:
		return fmt.Errorf("unknown default key: %s", key)
	}
	return Save()
}

// SetOutputFormat sets the output format preference
func SetOutputFormat(format string) error {
	c := Get()
	c.Output.Format = format
	return Save()
}

// ClearAuth clears all authentication data
func ClearAuth() error {
	c := Get()
	c.Auth.AccessToken = ""
	c.Auth.RefreshToken = ""
	c.Auth.TokenExpiry = time.Time{}
	return Save()
}

// IsAuthenticated checks if valid auth tokens exist
func IsAuthenticated() bool {
	c := Get()
	return c.Auth.AccessToken != "" && c.Auth.RefreshToken != ""
}

// IsTokenExpired checks if the access token has expired
func IsTokenExpired() bool {
	c := Get()
	if c.Auth.TokenExpiry.IsZero() {
		return true
	}
	// Consider token expired 5 minutes before actual expiry
	return time.Now().Add(5 * time.Minute).After(c.Auth.TokenExpiry)
}

// GetBaseURL returns the API base URL for the configured data center
func GetBaseURL() string {
	c := Get()
	dc := c.Defaults.DataCenter
	if dc == "" {
		dc = "com"
	}

	switch dc {
	case "eu":
		return "https://sprints.zoho.eu"
	case "in":
		return "https://sprints.zoho.in"
	case "au":
		return "https://sprints.zoho.com.au"
	case "jp":
		return "https://sprints.zoho.jp"
	case "ca":
		return "https://sprints.zohocloud.ca"
	default:
		return "https://sprints.zoho.com"
	}
}

// GetAccountsURL returns the OAuth accounts URL for the configured data center
func GetAccountsURL() string {
	c := Get()
	dc := c.Defaults.DataCenter
	if dc == "" {
		dc = "com"
	}

	switch dc {
	case "eu":
		return "https://accounts.zoho.eu"
	case "in":
		return "https://accounts.zoho.in"
	case "au":
		return "https://accounts.zoho.com.au"
	case "jp":
		return "https://accounts.zoho.jp"
	case "ca":
		return "https://accounts.zohocloud.ca"
	default:
		return "https://accounts.zoho.com"
	}
}

// Export returns the config as YAML bytes (for debugging)
func Export() ([]byte, error) {
	return yaml.Marshal(Get())
}
