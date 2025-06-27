package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"t212-taxes/internal/app/tui"
	"t212-taxes/internal/domain/calculator"
	"t212-taxes/internal/domain/parser"
	"t212-taxes/internal/domain/types"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "t212-taxes",
	Short: "Trading 212 CSV processor and tax calculator",
	Long: `A comprehensive tool for processing Trading 212 CSV exports and calculating tax obligations.

Features:
- Process multiple CSV files with yearly validation
- Calculate financial metrics (deposits, gains, dividends)
- Generate yearly and overall investment reports
- Beautiful terminal UI with detailed breakdowns`,
	Run: func(cmd *cobra.Command, args []string) {
		// Interactive mode - show TUI
		app := tui.NewApp()
		if err := app.Run(); err != nil {
			log.Fatalf("Failed to start TUI: %v", err)
		}
	},
}

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process CSV files and generate reports",
	Long: `Process Trading 212 CSV files and generate detailed financial reports.

Examples:
  # Process all CSV files in exports directory
  t212-taxes process --dir ./exports

  # Process specific files
  t212-taxes process --files file1.csv,file2.csv,file3.csv

  # Process with specific currency
  t212-taxes process --dir ./exports --currency EUR`,
	Run: processFiles,
}

// analyzeCmd represents the analyze command  
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze CSV files and show detailed reports",
	Long:  `Analyze Trading 212 CSV files and show detailed yearly and overall reports with TUI.`,
	Run:   analyzeFiles,
}

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate CSV file format and structure",
	Long:  `Validate Trading 212 CSV files for correct format and yearly structure.`,
	Run:   validateFiles,
}

func init() {
	// Add subcommands
	RootCmd.AddCommand(processCmd)
	RootCmd.AddCommand(analyzeCmd)
	RootCmd.AddCommand(validateCmd)

	// Global flags
	RootCmd.PersistentFlags().String("currency", "EUR", "Base currency for calculations")
	RootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	RootCmd.PersistentFlags().String("config", "", "Config file (default is ./config.yaml)")

	// Process command flags
	processCmd.Flags().String("dir", "", "Directory containing CSV files")
	processCmd.Flags().String("files", "", "Comma-separated list of CSV files")
	processCmd.Flags().String("output", "", "Output file for results (JSON format)")
	processCmd.Flags().String("format", "table", "Output format (table, json)")

	// Analyze command flags
	analyzeCmd.Flags().String("dir", "", "Directory containing CSV files")
	analyzeCmd.Flags().String("files", "", "Comma-separated list of CSV files")

	// Validate command flags
	validateCmd.Flags().String("dir", "", "Directory containing CSV files")
	validateCmd.Flags().String("files", "", "Comma-separated list of CSV files")

	// Bind flags to viper
	viper.BindPFlag("currency", RootCmd.PersistentFlags().Lookup("currency"))
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
}

// processFiles handles the process command
func processFiles(cmd *cobra.Command, args []string) {
	files, err := getCSVFiles(cmd)
	if err != nil {
		log.Fatalf("Error getting CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No CSV files found")
	}

	// Initialize parser and calculator
	csvParser := parser.NewCSVParser()
	currency := viper.GetString("currency")
	finCalc := calculator.NewFinancialCalculator(currency)

	// Parse files
	fmt.Printf("Processing %d CSV files...\n", len(files))
	result, err := csvParser.ParseMultipleFiles(files)
	if err != nil {
		log.Fatalf("Error parsing CSV files: %v", err)
	}

	// Calculate reports
	yearlyReports, err := finCalc.CalculateYearlyReports(result.Transactions)
	if err != nil {
		log.Fatalf("Error calculating yearly reports: %v", err)
	}

	overallReport := finCalc.CalculateOverallReport(yearlyReports)

	// Output results
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")

	if outputFile != "" {
		err = saveReportsToFile(yearlyReports, overallReport, outputFile, format)
		if err != nil {
			log.Fatalf("Error saving reports: %v", err)
		}
		fmt.Printf("Reports saved to %s\n", outputFile)
	} else {
		printReports(yearlyReports, overallReport, format)
	}
}

// analyzeFiles handles the analyze command
func analyzeFiles(cmd *cobra.Command, args []string) {
	files, err := getCSVFiles(cmd)
	if err != nil {
		log.Fatalf("Error getting CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No CSV files found")
	}

	// Initialize parser and calculator
	csvParser := parser.NewCSVParser()
	currency := viper.GetString("currency")
	finCalc := calculator.NewFinancialCalculator(currency)

	// Parse files
	result, err := csvParser.ParseMultipleFiles(files)
	if err != nil {
		log.Fatalf("Error parsing CSV files: %v", err)
	}

	// Calculate reports
	yearlyReports, err := finCalc.CalculateYearlyReports(result.Transactions)
	if err != nil {
		log.Fatalf("Error calculating yearly reports: %v", err)
	}

	overallReport := finCalc.CalculateOverallReport(yearlyReports)

	// Show TUI with reports
	app := tui.NewAppWithData(yearlyReports, overallReport)
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}

// validateFiles handles the validate command
func validateFiles(cmd *cobra.Command, args []string) {
	files, err := getCSVFiles(cmd)
	if err != nil {
		log.Fatalf("Error getting CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No CSV files found")
	}

	csvParser := parser.NewCSVParser()

	fmt.Printf("Validating %d CSV files...\n", len(files))
	
	// Validate yearly structure
	err = csvParser.ValidateYearlyStructure(files)
	if err != nil {
		fmt.Printf("❌ Yearly structure validation failed: %v\n", err)
		return
	}
	fmt.Println("✅ Yearly structure validation passed")

	// Validate individual files
	allValid := true
	for _, file := range files {
		fileHandle, err := os.Open(file)
		if err != nil {
			fmt.Printf("❌ %s: Failed to open file: %v\n", filepath.Base(file), err)
			allValid = false
			continue
		}

		err = csvParser.ValidateFormat(fileHandle)
		fileHandle.Close()
		
		if err != nil {
			fmt.Printf("❌ %s: %v\n", filepath.Base(file), err)
			allValid = false
		} else {
			fmt.Printf("✅ %s: Valid format\n", filepath.Base(file))
		}
	}

	if allValid {
		fmt.Println("\n🎉 All files are valid!")
	} else {
		fmt.Println("\n⚠️  Some files have validation errors")
		os.Exit(1)
	}
}

// getCSVFiles gets CSV files from command flags
func getCSVFiles(cmd *cobra.Command) ([]string, error) {
	dir, _ := cmd.Flags().GetString("dir")
	filesFlag, _ := cmd.Flags().GetString("files")

	var files []string

	if dir != "" {
		// Get all CSV files from directory
		matches, err := filepath.Glob(filepath.Join(dir, "*.csv"))
		if err != nil {
			return nil, fmt.Errorf("error globbing CSV files: %w", err)
		}
		files = append(files, matches...)
	}

	if filesFlag != "" {
		// Parse comma-separated files
		fileList := strings.Split(filesFlag, ",")
		for _, file := range fileList {
			files = append(files, strings.TrimSpace(file))
		}
	}

	// If no files specified, look in current directory
	if len(files) == 0 {
		matches, err := filepath.Glob("*.csv")
		if err != nil {
			return nil, fmt.Errorf("error globbing CSV files in current directory: %w", err)
		}
		files = append(files, matches...)
	}

	return files, nil
}

// printReports prints reports to console
func printReports(yearlyReports []types.YearlyReport, overallReport *types.OverallReport, format string) {
	if format == "json" {
		// Print JSON format
		fmt.Println("Yearly Reports:")
		for _, report := range yearlyReports {
			fmt.Printf("%+v\n", report)
		}
		fmt.Println("\nOverall Report:")
		fmt.Printf("%+v\n", overallReport)
	} else {
		// Print table format
		tui.PrintReportsTable(yearlyReports, overallReport)
	}
}

// saveReportsToFile saves reports to a file
func saveReportsToFile(yearlyReports []types.YearlyReport, overallReport *types.OverallReport, filename, format string) error {
	// This is a placeholder - in a real implementation, you'd serialize to JSON/CSV/etc
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if format == "json" {
		// Save as JSON
		file.WriteString("{\n")
		file.WriteString("  \"yearly_reports\": [\n")
		for i, report := range yearlyReports {
			file.WriteString(fmt.Sprintf("    %+v", report))
			if i < len(yearlyReports)-1 {
				file.WriteString(",")
			}
			file.WriteString("\n")
		}
		file.WriteString("  ],\n")
		file.WriteString(fmt.Sprintf("  \"overall_report\": %+v\n", overallReport))
		file.WriteString("}\n")
	} else {
		// Save as text table
		file.WriteString("Trading 212 Tax Report\n")
		file.WriteString("======================\n\n")
		
		for _, report := range yearlyReports {
			file.WriteString(fmt.Sprintf("Year %d:\n", report.Year))
			file.WriteString(fmt.Sprintf("  Deposits: %.2f %s\n", report.TotalDeposits, report.Currency))
			file.WriteString(fmt.Sprintf("  Transactions: %d\n", report.TotalTransactions))
			file.WriteString(fmt.Sprintf("  Capital Gains: %.2f %s\n", report.CapitalGains, report.Currency))
			file.WriteString(fmt.Sprintf("  Dividends: %.2f %s\n", report.Dividends, report.Currency))
			file.WriteString(fmt.Sprintf("  Total Gains: %.2f %s\n", report.TotalGains, report.Currency))
			file.WriteString(fmt.Sprintf("  Percentage Increase: %.2f%%\n\n", report.PercentageIncrease))
		}
		
		file.WriteString("Overall Summary:\n")
		file.WriteString(fmt.Sprintf("  Total Deposits: %.2f %s\n", overallReport.TotalDeposits, overallReport.Currency))
		file.WriteString(fmt.Sprintf("  Total Transactions: %d\n", overallReport.TotalTransactions))
		file.WriteString(fmt.Sprintf("  Total Gains: %.2f %s\n", overallReport.TotalGains, overallReport.Currency))
		file.WriteString(fmt.Sprintf("  Overall Percentage: %.2f%%\n", overallReport.OverallPercentage))
	}

	return nil
}