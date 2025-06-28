package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	// Save original viper state
	originalConfig := viper.AllSettings()
	
	// Reset viper for clean test
	viper.Reset()
	
	// Restore original config after test
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	tests := []struct {
		name           string
		envVars        map[string]string
		configFile     string
		configContent  string
		expectedValues map[string]interface{}
	}{
		{
			name: "default configuration",
			expectedValues: map[string]interface{}{
				"currency": "EUR",
				"verbose":  false,
			},
		},
		{
			name: "environment variables",
			envVars: map[string]string{
				"T212_CURRENCY": "USD",
				"T212_VERBOSE":  "true",
			},
			expectedValues: map[string]interface{}{
				"currency": "USD",
				"verbose":  true,
			},
		},
		{
			name:        "config file",
			configFile:  "test_config.yaml",
			configContent: `
currency: GBP
verbose: true
custom_setting: test_value
`,
			expectedValues: map[string]interface{}{
				"currency":       "GBP",
				"verbose":        true,
				"custom_setting": "test_value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create temporary config file if needed
			if tt.configFile != "" && tt.configContent != "" {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, tt.configFile)
				err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
				viper.Set("config", configPath)
			}

			// Run initConfig
			initConfig()

			// Verify expected values
			for key, expectedValue := range tt.expectedValues {
				var actualValue interface{}
				
				// Handle boolean values specifically
				if key == "verbose" {
					actualValue = viper.GetBool(key)
				} else {
					actualValue = viper.Get(key)
				}
				
				if actualValue != expectedValue {
					t.Errorf("Expected %s to be %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestInitConfigWithInvalidFile(t *testing.T) {
	// Save original viper state
	originalConfig := viper.AllSettings()
	
	// Reset viper for clean test
	viper.Reset()
	
	// Restore original config after test
	defer func() {
		viper.Reset()
		for key, value := range originalConfig {
			viper.Set(key, value)
		}
	}()

	// Set a non-existent config file
	viper.Set("config", "/non/existent/config.yaml")

	// This should not panic, just log a warning
	initConfig()

	// Verify defaults are still set
	if viper.GetString("currency") != "EUR" {
		t.Errorf("Expected default currency EUR, got %s", viper.GetString("currency"))
	}

	if viper.GetBool("verbose") != false {
		t.Errorf("Expected default verbose false, got %t", viper.GetBool("verbose"))
	}
}

func TestMainFunction(t *testing.T) {
	// This is a basic test to ensure main doesn't panic during initialization
	// We can't easily test the full execution without mocking cli.RootCmd.Execute()
	
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set args to help command to avoid actual processing
	os.Args = []string{"t212-taxes", "--help"}

	// This should not panic during initialization phase
	// Note: This will exit with help, so we can't test the full flow easily
	// But we can test that the initialization functions work
	
	// Test version info setup
	v, c, d, b := getVersionInfo()
	if v == "" || c == "" || d == "" || b == "" {
		t.Error("Version info should not be empty")
	}
} 