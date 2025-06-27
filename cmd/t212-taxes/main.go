// Package main provides the entry point for the t212-taxes CLI application.
// This application processes Trading 212 CSV exports and calculates financial metrics
// including yearly reports, capital gains, dividends, and overall investment performance.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"

	"github.com/Lizzergas/go-t212-taxes/internal/app/cli"
)

// Build-time variables injected via ldflags
var (
	version = "dev"      // Version will be set during build: -X main.version=v1.0.0
	commit  = "unknown"  // Git commit hash: -X main.commit=abc123
	date    = "unknown"  // Build date: -X main.date=2024-01-01T00:00:00Z
	builtBy = "source"   // How it was built: -X main.builtBy=goreleaser
)

func main() {
	// Initialize configuration
	initConfig()

	// Get version information (dynamic if not set via ldflags)
	v, c, d, b := getVersionInfo()

	// Set version information for CLI
	cli.SetVersionInfo(v, c, d, b)

	// Execute the root command
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// Set default configuration values
	viper.SetDefault("currency", "EUR")
	viper.SetDefault("verbose", false)

	// Read configuration from environment variables
	viper.SetEnvPrefix("T212")
	viper.AutomaticEnv()

	// Read configuration from config file if specified
	configFile := viper.GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Look for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// Only log if a config file was explicitly specified
		if configFile != "" {
			log.Printf("Warning: Could not read config file %s: %v", viper.ConfigFileUsed(), err)
		}
	} else {
		if viper.GetBool("verbose") {
			log.Printf("Using config file: %s", viper.ConfigFileUsed())
		}
	}

	// Set log level based on verbose flag
	if viper.GetBool("verbose") {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	} else {
		log.SetFlags(0)
	}
}
