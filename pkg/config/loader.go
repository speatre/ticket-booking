package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader provides configuration loading with environment-specific features
type Loader struct {
	configPath string
	env        Environment
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) (*Loader, error) {
	envStr := os.Getenv("APP_ENV")
	if envStr == "" {
		envStr = string(EnvDevelopment)
	}

	env, err := ParseEnvironment(envStr)
	if err != nil {
		return nil, fmt.Errorf("parse environment: %w", err)
	}

	return &Loader{
		configPath: configPath,
		env:        env,
	}, nil
}

// Load loads and validates configuration for the current environment
func (l *Loader) Load() (*Config, error) {
	// Try environment-specific config first
	envConfigPath := l.getEnvConfigPath()
	if _, err := os.Stat(envConfigPath); err == nil {
		// Environment-specific config exists
		config := Load(envConfigPath)
		config.App.Env = string(l.env)
		return config, nil
	}

	// Fall back to default config
	if _, err := os.Stat(l.configPath); err != nil {
		return nil, fmt.Errorf("config file not found: %s (also tried %s)", l.configPath, envConfigPath)
	}

	config := Load(l.configPath)
	config.App.Env = string(l.env)
	return config, nil
}

// LoadWithDefaults loads configuration with environment-specific defaults
func (l *Loader) LoadWithDefaults() (*Config, error) {
	config, err := l.Load()
	if err != nil {
		return nil, err
	}

	// Apply environment-specific defaults
	envDefaults := GetConfigForEnv(l.env)
	l.applyEnvironmentOverrides(config, envDefaults)

	return config, nil
}

// getEnvConfigPath returns the path for environment-specific config file
func (l *Loader) getEnvConfigPath() string {
	ext := filepath.Ext(l.configPath)
	name := strings.TrimSuffix(filepath.Base(l.configPath), ext)
	dir := filepath.Dir(l.configPath)

	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, l.env, ext))
}

// applyEnvironmentOverrides applies environment-specific overrides
func (l *Loader) applyEnvironmentOverrides(config, defaults *Config) {
	// Only override values that are still at their zero/default values
	if config.Logging.RetentionDays == 0 || config.Logging.RetentionDays == DefaultLogRetentionDays {
		config.Logging.RetentionDays = defaults.Logging.RetentionDays
	}

	if config.Worker.PaymentSuccessRate == 0 || config.Worker.PaymentSuccessRate == DefaultPaymentSuccessRate {
		config.Worker.PaymentSuccessRate = defaults.Worker.PaymentSuccessRate
	}

	// Add more environment-specific overrides as needed
}

// GetEnvironment returns the current environment
func (l *Loader) GetEnvironment() Environment {
	return l.env
}

// IsProduction returns true if running in production
func (l *Loader) IsProduction() bool {
	return l.env.IsProduction()
}

// IsDevelopment returns true if running in development
func (l *Loader) IsDevelopment() bool {
	return l.env.IsDevelopment()
}

// MustLoad loads configuration and panics on error (for simple applications)
func MustLoad(configPath string) *Config {
	loader, err := NewLoader(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to create config loader: %v", err))
	}

	config, err := loader.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	return config
}

// MustLoadWithDefaults loads configuration with defaults and panics on error
func MustLoadWithDefaults(configPath string) *Config {
	loader, err := NewLoader(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to create config loader: %v", err))
	}

	config, err := loader.LoadWithDefaults()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	return config
}

