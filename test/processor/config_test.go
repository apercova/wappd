package processor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apercova/wappd/internal/processor"
)

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "wappd.json")

	// Create a valid config file
	configContent := `{
		"updateModified": true,
		"overwriteExif": false,
		"outputDir": "./processed",
		"verbose": true
	}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load config file
	config, err := processor.LoadConfigFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfigFile() error = %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfigFile() returned nil config")
	}

	if config.UpdateModified == nil || !*config.UpdateModified {
		t.Error("LoadConfigFile() updateModified should be true")
	}

	if config.OverwriteExif == nil || *config.OverwriteExif {
		t.Error("LoadConfigFile() overwriteExif should be false")
	}

	if config.OutputDir != "./processed" {
		t.Errorf("LoadConfigFile() outputDir = %v, want ./processed", config.OutputDir)
	}

	if config.Verbose == nil || !*config.Verbose {
		t.Error("LoadConfigFile() verbose should be true")
	}
}

func TestLoadConfigFile_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load non-existent config file
	config, err := processor.LoadConfigFile(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfigFile() should not error when file doesn't exist: %v", err)
	}

	if config != nil {
		t.Error("LoadConfigFile() should return nil when file doesn't exist")
	}
}

func TestLoadConfigFileFromPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.json")

	// Create a valid config file
	configContent := `{
		"updateModified": true,
		"overrideOriginal": true
	}`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load config file from custom path
	config, err := processor.LoadConfigFileFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFileFromPath() error = %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfigFileFromPath() returned nil config")
	}

	if config.UpdateModified == nil || !*config.UpdateModified {
		t.Error("LoadConfigFileFromPath() updateModified should be true")
	}

	if config.OverrideOriginal == nil || !*config.OverrideOriginal {
		t.Error("LoadConfigFileFromPath() overrideOriginal should be true")
	}
}

func TestLoadConfigFileFromPath_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	// Create invalid JSON
	err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Try to load invalid config file
	_, err = processor.LoadConfigFileFromPath(configPath)
	if err == nil {
		t.Error("LoadConfigFileFromPath() should error on invalid JSON")
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig *processor.ConfigFile
		cliConfig  processor.Config
		want       processor.Config
	}{
		{
			name:       "No config file",
			fileConfig: nil,
			cliConfig: processor.Config{
				UpdateModified:   false,
				OverwriteExif:    false,
				OverrideOriginal: false,
				OutputDir:        "",
				Verbose:          false,
				DryRun:           false,
			},
			want: processor.Config{
				UpdateModified:   false,
				OverwriteExif:    false,
				OverrideOriginal: false,
				OutputDir:        "",
				Verbose:          false,
				DryRun:           false,
			},
		},
		{
			name: "Config file provides defaults",
			fileConfig: &processor.ConfigFile{
				UpdateModified:   boolPtr(true),
				OverwriteExif:    boolPtr(false),
				OverrideOriginal: boolPtr(true),
				OutputDir:        "./processed",
				Verbose:          boolPtr(true),
			},
			cliConfig: processor.Config{
				UpdateModified:   false, // CLI default
				OverwriteExif:    false, // CLI default
				OverrideOriginal: false, // CLI default
				OutputDir:        "",    // CLI default
				Verbose:          false, // CLI default
				DryRun:           false,
			},
			want: processor.Config{
				UpdateModified:   true,  // From config file
				OverwriteExif:    false, // From config file
				OverrideOriginal: true,  // From config file
				OutputDir:        "./processed", // From config file
				Verbose:          true,  // From config file
				DryRun:           false, // Always from CLI
			},
		},
		{
			name: "CLI flags override config file",
			fileConfig: &processor.ConfigFile{
				UpdateModified:   boolPtr(false),
				OverwriteExif:    boolPtr(false),
				OverrideOriginal: boolPtr(false),
				OutputDir:        "./processed",
				Verbose:          boolPtr(false),
			},
			cliConfig: processor.Config{
				UpdateModified:   true,  // CLI explicitly set
				OverwriteExif:    true,  // CLI explicitly set
				OverrideOriginal: true,  // CLI explicitly set
				OutputDir:        "./custom", // CLI explicitly set
				Verbose:          true,  // CLI explicitly set
				DryRun:           true,
			},
			want: processor.Config{
				UpdateModified:   true,  // CLI wins
				OverwriteExif:    true,  // CLI wins
				OverrideOriginal: true,  // CLI wins
				OutputDir:        "./custom", // CLI wins
				Verbose:          true,  // CLI wins
				DryRun:           true, // Always from CLI
			},
		},
		{
			name: "Mixed: some CLI, some config",
			fileConfig: &processor.ConfigFile{
				UpdateModified:   boolPtr(true),
				OverwriteExif:    boolPtr(false),
				OutputDir:        "./processed",
				Verbose:          boolPtr(true),
			},
			cliConfig: processor.Config{
				UpdateModified:   true,  // CLI explicitly set to true
				OverwriteExif:    false, // CLI default (false), use config file value
				OverrideOriginal: false, // CLI default, config not set
				OutputDir:        "",    // CLI default (empty), use config file value
				Verbose:          false, // CLI default (false), use config file value
				DryRun:           false,
			},
			want: processor.Config{
				UpdateModified:   true,  // CLI explicitly set
				OverwriteExif:    false, // From config file (CLI false = use config)
				OverrideOriginal: false, // Default (config not set)
				OutputDir:        "./processed", // From config file (CLI empty = use config)
				Verbose:          true,  // From config file (CLI false = use config)
				DryRun:           false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.MergeConfig(tt.fileConfig, tt.cliConfig)
			if got.UpdateModified != tt.want.UpdateModified {
				t.Errorf("MergeConfig() UpdateModified = %v, want %v", got.UpdateModified, tt.want.UpdateModified)
			}
			if got.OverwriteExif != tt.want.OverwriteExif {
				t.Errorf("MergeConfig() OverwriteExif = %v, want %v", got.OverwriteExif, tt.want.OverwriteExif)
			}
			if got.OverrideOriginal != tt.want.OverrideOriginal {
				t.Errorf("MergeConfig() OverrideOriginal = %v, want %v", got.OverrideOriginal, tt.want.OverrideOriginal)
			}
			if got.OutputDir != tt.want.OutputDir {
				t.Errorf("MergeConfig() OutputDir = %v, want %v", got.OutputDir, tt.want.OutputDir)
			}
			if got.Verbose != tt.want.Verbose {
				t.Errorf("MergeConfig() Verbose = %v, want %v", got.Verbose, tt.want.Verbose)
			}
			if got.DryRun != tt.want.DryRun {
				t.Errorf("MergeConfig() DryRun = %v, want %v", got.DryRun, tt.want.DryRun)
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

func TestConfigFileName(t *testing.T) {
	name := processor.ConfigFileName()
	if name != "wappd.json" {
		t.Errorf("ConfigFileName() = %v, want wappd.json", name)
	}
}
