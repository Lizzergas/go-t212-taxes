// Package main provides the entry point for the t212-taxes CLI application.
// This application processes Trading 212 CSV exports and calculates financial metrics
// including yearly reports, capital gains, dividends, and overall investment performance.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"

	"t212-taxes/internal/app/cli"
)

func main() {
	// Initialize configuration
	initConfig()

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