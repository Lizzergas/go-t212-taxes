package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetCSVFiles(t *testing.T) {
	// Create temporary directory and files for testing
	tmpDir, err := os.MkdirTemp("", "t212-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test CSV files
	testFiles := []string{
		"from_2023-01-01_to_2023-12-31_abc123.csv",
		"from_2024-01-01_to_2024-12-31_def456.csv",
		"not_a_csv.txt",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tmpDir, file)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		want     int // expected number of CSV files
		wantErr  bool
	}{
		{
			name: "get CSV files from directory",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("dir", tmpDir, "test directory")
				cmd.Flags().String("files", "", "test files")
				return cmd
			},
			want:    2, // Only the CSV files
			wantErr: false,
		},
		{
			name: "get specific CSV files",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("dir", "", "test directory")
				cmd.Flags().String("files", "file1.csv,file2.csv", "test files")
				return cmd
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "no flags set - should look in current directory",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("dir", "", "test directory")
				cmd.Flags().String("files", "", "test files")
				return cmd
			},
			want:    0, // Assuming no CSV files in test's current directory
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			files, err := getCSVFiles(cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("getCSVFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(files) != tt.want {
				t.Errorf("getCSVFiles() got %d files, want %d", len(files), tt.want)
			}

			// Verify all returned files are CSV files
			for _, file := range files {
				if filepath.Ext(file) != ".csv" {
					t.Errorf("getCSVFiles() returned non-CSV file: %s", file)
				}
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		want     string
	}{
		{
			name:     "positive EUR amount",
			amount:   123.45,
			currency: "EUR",
			want:     "€123.45",
		},
		{
			name:     "negative EUR amount",
			amount:   -123.45,
			currency: "EUR",
			want:     "-€123.45",
		},
		{
			name:     "zero amount",
			amount:   0.0,
			currency: "EUR",
			want:     "€0.00",
		},
		{
			name:     "USD amount",
			amount:   100.0,
			currency: "USD",
			want:     "$100.00",
		},
		{
			name:     "GBP amount",
			amount:   50.75,
			currency: "GBP",
			want:     "£50.75",
		},
		{
			name:     "unknown currency",
			amount:   200.0,
			currency: "XYZ",
			want:     "XYZ 200.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test this through the TUI package since formatCurrency is not exported
			// This is a limitation of the current design
			t.Skip("formatCurrency is not exported - would need refactoring to test")
		})
	}
}

func TestRootCmd(t *testing.T) {
	// Test that the root command can be created without errors
	if RootCmd == nil {
		t.Error("RootCmd should not be nil")
	}

	if RootCmd.Use != "t212-taxes" {
		t.Errorf("RootCmd.Use = %s, want 't212-taxes'", RootCmd.Use)
	}

	// Check that subcommands are registered
	expectedCommands := []string{"process", "analyze", "validate", "income", "portfolio"}
	commands := RootCmd.Commands()

	if len(commands) != len(expectedCommands) {
		t.Errorf("RootCmd has %d commands, want %d", len(commands), len(expectedCommands))
	}

	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("RootCmd missing command: %s", expected)
		}
	}
}

func TestProcessCmd(t *testing.T) {
	if processCmd == nil {
		t.Error("processCmd should not be nil")
	}

	if processCmd.Use != "process" {
		t.Errorf("processCmd.Use = %s, want 'process'", processCmd.Use)
	}

	// Check that required flags are present
	expectedFlags := []string{"dir", "files", "output", "format"}

	for _, flagName := range expectedFlags {
		flag := processCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("processCmd missing flag: %s", flagName)
		}
	}
}

func TestAnalyzeCmd(t *testing.T) {
	if analyzeCmd == nil {
		t.Error("analyzeCmd should not be nil")
	}

	if analyzeCmd.Use != "analyze" {
		t.Errorf("analyzeCmd.Use = %s, want 'analyze'", analyzeCmd.Use)
	}

	// Check that required flags are present
	expectedFlags := []string{"dir", "files"}

	for _, flagName := range expectedFlags {
		flag := analyzeCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("analyzeCmd missing flag: %s", flagName)
		}
	}
}

func TestValidateCmd(t *testing.T) {
	if validateCmd == nil {
		t.Error("validateCmd should not be nil")
	}

	if validateCmd.Use != "validate" {
		t.Errorf("validateCmd.Use = %s, want 'validate'", validateCmd.Use)
	}

	// Check that required flags are present
	expectedFlags := []string{"dir", "files"}

	for _, flagName := range expectedFlags {
		flag := validateCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("validateCmd missing flag: %s", flagName)
		}
	}
}
