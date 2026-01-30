package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName = "wappd.json"
)

// ConfigFileName returns the name of the config file
func ConfigFileName() string {
	return configFileName
}

// ConfigFile represents the JSON configuration file structure
type ConfigFile struct {
	UpdateModified   *bool  `json:"updateModified,omitempty"`
	OverwriteExif   *bool  `json:"overwriteExif,omitempty"`
	OverrideOriginal *bool  `json:"overrideOriginal,omitempty"`
	OutputDir        string `json:"outputDir,omitempty"`
	Verbose          *bool  `json:"verbose,omitempty"`
}

// LoadConfigFile loads configuration from wappd.json if it exists in the specified directory
// Returns nil if file doesn't exist (not an error)
func LoadConfigFile(dirPath string) (*ConfigFile, error) {
	configPath := filepath.Join(dirPath, configFileName)
	return LoadConfigFileFromPath(configPath)
}

// LoadConfigFileFromPath loads configuration from a specific file path
// Returns nil if file doesn't exist (not an error)
func LoadConfigFileFromPath(configPath string) (*ConfigFile, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // No config file is fine
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	return &config, nil
}

// MergeConfig merges config file values with CLI flags
// CLI flags take precedence over config file values
// For boolean flags: if CLI flag is true (explicitly set), it overrides config.
//                    if CLI flag is false (default), config file value is used if present.
// For strings: if CLI flag is non-empty, it overrides config.
//              if CLI flag is empty, config file value is used if present.
func MergeConfig(fileConfig *ConfigFile, cliConfig Config) Config {
	result := cliConfig
	
	if fileConfig == nil {
		return result
	}
	
	// Boolean flags: CLI true overrides, CLI false allows config file default
	if fileConfig.UpdateModified != nil {
		if cliConfig.UpdateModified {
			// CLI explicitly set to true, use it
			result.UpdateModified = true
		} else {
			// CLI is false (default), use config file value
			result.UpdateModified = *fileConfig.UpdateModified
		}
	}
	
	if fileConfig.OverwriteExif != nil {
		if cliConfig.OverwriteExif {
			result.OverwriteExif = true
		} else {
			result.OverwriteExif = *fileConfig.OverwriteExif
		}
	}
	
	if fileConfig.OverrideOriginal != nil {
		if cliConfig.OverrideOriginal {
			result.OverrideOriginal = true
		} else {
			result.OverrideOriginal = *fileConfig.OverrideOriginal
		}
	}
	
	if fileConfig.Verbose != nil {
		if cliConfig.Verbose {
			result.Verbose = true
		} else {
			result.Verbose = *fileConfig.Verbose
		}
	}
	
	// String flags: CLI non-empty overrides, CLI empty allows config file default
	if fileConfig.OutputDir != "" {
		if cliConfig.OutputDir != "" {
			// CLI explicitly set, use it
			result.OutputDir = cliConfig.OutputDir
		} else {
			// CLI is empty, use config file value
			result.OutputDir = fileConfig.OutputDir
		}
	}
	
	// Note: DryRun is not in config file - always CLI-only for safety
	
	return result
}
