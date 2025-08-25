package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Version    string     `mapstructure:"version"`
	Project    Project    `mapstructure:"project"`
	Paths      Paths      `mapstructure:"paths"`
	Generation Generation `mapstructure:"generation"`
}

type Project struct {
	Module string `mapstructure:"module"` // Go module name from go.mod
}

type Paths struct {
	ScanDirs  []string `mapstructure:"scan_dirs"`
	OutputDir string   `mapstructure:"output_dir"`
}

type Generation struct {
	Routes       RouteConfig `mapstructure:"routes"`
	Dependencies DepConfig   `mapstructure:"dependencies"`
}

type RouteConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	OutputFile string `mapstructure:"output_file"`
}

type DepConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	OutputFile string `mapstructure:"output_file"`
}

// ProvideConfig loads taskw.yaml from current directory or creates default config using Viper
func ProvideConfig() (*Config, error) {
	v := viper.New()

	// Set config file details
	configDir := filepath.Dir(".")
	configName := strings.TrimSuffix(filepath.Base("taskw.yaml"), filepath.Ext("taskw.yaml"))
	configType := strings.TrimPrefix(filepath.Ext("taskw.yaml"), ".")

	if configDir == "." {
		v.AddConfigPath(".")
	} else {
		v.AddConfigPath(configDir)
	}
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Set defaults
	if err := setDefaults(v); err != nil {
		return nil, fmt.Errorf("error setting defaults: %w", err)
	}

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		// If config doesn't exist, create it with defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			config := &Config{}
			if err := v.Unmarshal(config); err != nil {
				return nil, fmt.Errorf("error unmarshaling default config: %w", err)
			}

			return config, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values using Viper
func setDefaults(v *viper.Viper) error {
	// Auto-detect Go module
	module, err := detectGoModule()
	if err != nil {
		return fmt.Errorf("error detecting Go module: %w", err)
	}

	v.SetDefault("version", "1.0")
	v.SetDefault("project.module", module)
	v.SetDefault("paths.scan_dirs", []string{"."})
	v.SetDefault("paths.output_dir", ".")
	v.SetDefault("generation.routes.enabled", true)
	v.SetDefault("generation.routes.output_file", "routes_gen.go")
	v.SetDefault("generation.dependencies.enabled", true)
	v.SetDefault("generation.dependencies.output_file", "dependencies_gen.go")

	return nil
}

// detectGoModule reads go.mod to extract the module name
// Returns empty string if go.mod doesn't exist (e.g., during init)
func detectGoModule() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		// If go.mod doesn't exist, return empty string (will be handled during init)
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("could not read go.mod: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(line[7:]), nil
		}
	}

	return "", fmt.Errorf("could not detect Go module name from go.mod")
}

// Save writes the config to a YAML file
func (c *Config) Save(path string) error {
	if path == "" {
		path = "taskw.yaml"
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Set all values from config struct
	v.Set("version", c.Version)
	v.Set("project.module", c.Project.Module)
	v.Set("paths.scan_dirs", c.Paths.ScanDirs)
	v.Set("paths.output_dir", c.Paths.OutputDir)
	v.Set("generation.routes.enabled", c.Generation.Routes.Enabled)
	v.Set("generation.routes.output_file", c.Generation.Routes.OutputFile)
	v.Set("generation.dependencies.enabled", c.Generation.Dependencies.Enabled)
	v.Set("generation.dependencies.output_file", c.Generation.Dependencies.OutputFile)

	// Write config file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
