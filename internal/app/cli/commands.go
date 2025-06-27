package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

// incomeCmd represents the income command
var incomeCmd = &cobra.Command{
	Use:   "income",
	Short: "Generate detailed dividend and interest income reports",
	Long: `Generate comprehensive dividend and interest income reports from Trading 212 CSV files.

Features:
- Detailed dividend analysis with withholding tax tracking
- Interest rate analysis and source breakdown
- Monthly and yearly income breakdowns
- Top dividend payers identification
- Currency conversion support

Examples:
  # Generate income report for all CSV files in exports directory
  t212-taxes income --dir ./exports

  # Generate income report with JSON output
  t212-taxes income --dir ./exports --format json --output income_report.json

  # Generate income report for specific files
  t212-taxes income --files file1.csv,file2.csv`,
	Run: generateIncomeReport,
}

// portfolioCmd represents the portfolio command
var portfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Generate portfolio valuation reports across years",
	Long: `Generate comprehensive portfolio valuation reports showing holdings and market values across years.

Features:
- End-of-year portfolio positions for each year
- Market values based on last transaction prices
- Unrealized gains/losses calculations
- Total portfolio value tracking
- Year-over-year portfolio growth analysis

Examples:
  # Generate portfolio valuation report for all CSV files
  t212-taxes portfolio --dir ./exports

  # Generate portfolio report with JSON output
  t212-taxes portfolio --dir ./exports --format json --output portfolio_report.json

  # Show specific number of holdings per year
  t212-taxes portfolio --dir ./exports --max-holdings 15`,
	Run: generatePortfolioReport,
}

func init() {
	// Add subcommands
	RootCmd.AddCommand(processCmd)
	RootCmd.AddCommand(analyzeCmd)
	RootCmd.AddCommand(validateCmd)
	RootCmd.AddCommand(incomeCmd)
	RootCmd.AddCommand(portfolioCmd)

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

	// Income command flags
	incomeCmd.Flags().String("dir", "", "Directory containing CSV files")
	incomeCmd.Flags().String("files", "", "Comma-separated list of CSV files")
	incomeCmd.Flags().String("output", "", "Output file for results (JSON format)")
	incomeCmd.Flags().String("format", "table", "Output format (table, json)")
	incomeCmd.Flags().Int("top-payers", 10, "Number of top dividend payers to display")

	// Portfolio command flags
	portfolioCmd.Flags().String("dir", "", "Directory containing CSV files")
	portfolioCmd.Flags().StringSlice("files", []string{}, "Comma-separated list of CSV files")
	portfolioCmd.Flags().String("output", "", "Output file path")
	portfolioCmd.Flags().String("format", "table", "Output format (table, json)")
	portfolioCmd.Flags().Int("max-holdings", 10, "Maximum number of holdings to display per year")
	portfolioCmd.Flags().Bool("show-all", false, "Show all positions (ignores max-holdings limit)")

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

	// Calculate portfolio valuation report
	portfolioReport, err := finCalc.CalculatePortfolioReports(files)
	if err != nil {
		log.Printf("Warning: Could not calculate portfolio report: %v", err)
		portfolioReport = nil
	}

	// Calculate income report
	incomeCalc := calculator.NewIncomeCalculator(currency)
	incomeReport, err := incomeCalc.CalculateIncomeReport(result.Transactions)
	if err != nil {
		log.Printf("Warning: Could not calculate income report: %v", err)
		incomeReport = nil
	}

	// Show TUI with all available data
	app := tui.NewAppWithAllData(yearlyReports, overallReport, result.Transactions, portfolioReport, incomeReport)
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

// generateIncomeReport handles the income command
func generateIncomeReport(cmd *cobra.Command, args []string) {
	files, err := getCSVFiles(cmd)
	if err != nil {
		log.Fatalf("Error getting CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No CSV files found")
	}

	// Initialize parser and income calculator
	csvParser := parser.NewCSVParser()
	currency := viper.GetString("currency")
	incomeCalc := calculator.NewIncomeCalculator(currency)

	// Parse files
	fmt.Printf("Processing %d CSV files for income analysis...\n", len(files))
	result, err := csvParser.ParseMultipleFiles(files)
	if err != nil {
		log.Fatalf("Error parsing CSV files: %v", err)
	}

	// Calculate income report
	incomeReport, err := incomeCalc.CalculateIncomeReport(result.Transactions)
	if err != nil {
		log.Fatalf("Error calculating income report: %v", err)
	}

	// Output results
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")

	if outputFile != "" {
		err = saveIncomeReportToFile(incomeReport, outputFile, format)
		if err != nil {
			log.Fatalf("Error saving income report: %v", err)
		}
		fmt.Printf("Income report saved to %s\n", outputFile)
	} else {
		printIncomeReport(incomeReport, format)
	}
}

// printIncomeReport prints income report to console
func printIncomeReport(report *types.IncomeReport, format string) {
	if format == "json" {
		// Print JSON format
		fmt.Printf("%+v\n", report)
	} else {
		// Print table format
		printIncomeReportTable(report)
	}
}

// printIncomeReportTable prints income report in table format
func printIncomeReportTable(report *types.IncomeReport) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                    INCOME REPORT")
	fmt.Println(strings.Repeat("=", 80))

	// Summary section
	fmt.Printf("\n📊 SUMMARY (%s)\n", report.Currency)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Total Income:           %10.2f %s\n", report.TotalIncome, report.Currency)
	fmt.Printf("Date Range:             %s to %s\n",
		report.DateRange.From.Format("2006-01-02"),
		report.DateRange.To.Format("2006-01-02"))

	// Dividend section
	fmt.Printf("\n💰 DIVIDENDS (%s)\n", report.Currency)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Total Dividends:        %10.2f %s\n", report.Dividends.TotalDividends, report.Currency)
	fmt.Printf("Withholding Tax:        %10.2f %s\n", report.Dividends.TotalWithholdingTax, report.Currency)
	fmt.Printf("Net Dividends:          %10.2f %s\n", report.Dividends.NetDividends, report.Currency)
	fmt.Printf("Dividend Count:         %10d\n", report.Dividends.DividendCount)
	if report.Dividends.AverageYield > 0 {
		fmt.Printf("Average Yield:          %10.2f%%\n", report.Dividends.AverageYield)
	}

	// Interest section
	fmt.Printf("\n🏦 INTEREST (%s)\n", report.Currency)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Total Interest:         %10.2f %s\n", report.Interest.TotalInterest, report.Currency)
	fmt.Printf("Interest Count:         %10d\n", report.Interest.InterestCount)
	if report.Interest.AverageRate > 0 {
		fmt.Printf("Average Rate:           %10.2f%%\n", report.Interest.AverageRate)
	}

	// Top dividend payers
	if len(report.Dividends.BySecurity) > 0 {
		fmt.Printf("\n🏆 TOP DIVIDEND PAYERS (%s)\n", report.Currency)
		fmt.Println(strings.Repeat("-", 40))

		// Convert map to slice for sorting
		type securityAmount struct {
			security string
			amount   float64
		}
		var securities []securityAmount
		for security, amount := range report.Dividends.BySecurity {
			securities = append(securities, securityAmount{security, amount})
		}

		// Sort by amount (descending)
		sort.Slice(securities, func(i, j int) bool {
			return securities[i].amount > securities[j].amount
		})

		// Display top 10
		limit := 10
		if len(securities) < limit {
			limit = len(securities)
		}

		for i := 0; i < limit; i++ {
			fmt.Printf("%-20s %10.2f %s\n", securities[i].security, securities[i].amount, report.Currency)
		}
	}

	// Interest by source
	if len(report.Interest.BySource) > 0 {
		fmt.Printf("\n📈 INTEREST BY SOURCE (%s)\n", report.Currency)
		fmt.Println(strings.Repeat("-", 40))

		for source, amount := range report.Interest.BySource {
			fmt.Printf("%-20s %10.2f %s\n", source, amount, report.Currency)
		}
	}

	// Monthly breakdown
	if len(report.Dividends.ByMonth) > 0 || len(report.Interest.ByMonth) > 0 {
		fmt.Printf("\n📅 MONTHLY BREAKDOWN (%s)\n", report.Currency)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("%-10s %12s %12s %12s\n", "Month", "Dividends", "Interest", "Total")
		fmt.Println(strings.Repeat("-", 50))

		// Combine all months
		allMonths := make(map[string]bool)
		for month := range report.Dividends.ByMonth {
			allMonths[month] = true
		}
		for month := range report.Interest.ByMonth {
			allMonths[month] = true
		}

		// Convert to slice and sort
		var months []string
		for month := range allMonths {
			months = append(months, month)
		}
		sort.Strings(months)

		for _, month := range months {
			dividends := report.Dividends.ByMonth[month]
			interest := report.Interest.ByMonth[month]
			total := dividends + interest
			fmt.Printf("%-10s %12.2f %12.2f %12.2f\n", month, dividends, interest, total)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

// saveIncomeReportToFile saves income report to a file
func saveIncomeReportToFile(report *types.IncomeReport, filename, format string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if format == "json" {
		// Save as JSON
		file.WriteString(fmt.Sprintf("%+v\n", report))
	} else {
		// Save as text table
		file.WriteString("Trading 212 Income Report\n")
		file.WriteString("=========================\n\n")

		file.WriteString(fmt.Sprintf("Total Income: %.2f %s\n", report.TotalIncome, report.Currency))
		file.WriteString(fmt.Sprintf("Date Range: %s to %s\n\n",
			report.DateRange.From.Format("2006-01-02"),
			report.DateRange.To.Format("2006-01-02")))

		file.WriteString("Dividends:\n")
		file.WriteString(fmt.Sprintf("  Total: %.2f %s\n", report.Dividends.TotalDividends, report.Currency))
		file.WriteString(fmt.Sprintf("  Withholding Tax: %.2f %s\n", report.Dividends.TotalWithholdingTax, report.Currency))
		file.WriteString(fmt.Sprintf("  Net: %.2f %s\n", report.Dividends.NetDividends, report.Currency))
		file.WriteString(fmt.Sprintf("  Count: %d\n\n", report.Dividends.DividendCount))

		file.WriteString("Interest:\n")
		file.WriteString(fmt.Sprintf("  Total: %.2f %s\n", report.Interest.TotalInterest, report.Currency))
		file.WriteString(fmt.Sprintf("  Count: %d\n", report.Interest.InterestCount))
	}

	return nil
}

// generatePortfolioReport handles the portfolio command
func generatePortfolioReport(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")
	format, _ := cmd.Flags().GetString("format")
	maxHoldings, _ := cmd.Flags().GetInt("max-holdings")
	showAll, _ := cmd.Flags().GetBool("show-all")

	if dir == "" {
		log.Fatal("Directory must be specified")
	}

	// Initialize calculator
	currency := viper.GetString("currency")
	finCalc := calculator.NewFinancialCalculator(currency)

	// Parse files
	fmt.Printf("Generating portfolio valuation report for %s...\n", dir)
	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		log.Fatalf("Error globbing CSV files: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("No CSV files found")
	}

	// Calculate portfolio reports
	portfolioReport, err := finCalc.CalculatePortfolioReports(files)
	if err != nil {
		log.Fatalf("Error calculating portfolio reports: %v", err)
	}

	// Display results
	if format == "json" {
		printPortfolioReportJSON(portfolioReport)
	} else {
		printPortfolioReportTable(portfolioReport, maxHoldings, showAll)
	}
}

// printPortfolioReportJSON prints portfolio report in JSON format
func printPortfolioReportJSON(report *types.PortfolioValuationReport) {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling portfolio report to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}

// printPortfolioReportTable prints portfolio report in table format
func printPortfolioReportTable(report *types.PortfolioValuationReport, maxHoldings int, showAll bool) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                   📊 PORTFOLIO VALUATION REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Data Source: %s\n", report.DataSource)
	fmt.Printf("Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Note: %s\n", report.PriceNote)

	if len(report.YearlyPortfolios) == 0 {
		fmt.Println("\nNo portfolio data found.")
		return
	}

	// Summary across all years
	fmt.Printf("\n💰 PORTFOLIO SUMMARY (%s)\n", report.Currency)
	fmt.Println(strings.Repeat("-", 60))

	for _, yearly := range report.YearlyPortfolios {
		fmt.Printf("\n📅 YEAR %d (as of %s)\n", yearly.Year, yearly.AsOfDate.Format("2006-01-02"))
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("Total Positions:        %10d\n", yearly.TotalPositions)
		fmt.Printf("Total Shares:           %10.2f\n", yearly.TotalShares)
		fmt.Printf("Total Invested:         %10.2f %s\n", yearly.TotalInvested, yearly.Currency)
		fmt.Printf("Total Market Value:     %10.2f %s\n", yearly.TotalMarketValue, yearly.Currency)
		fmt.Printf("Unrealized P&L:         %10.2f %s (%.2f%%)\n",
			yearly.TotalUnrealizedGainLoss, yearly.Currency, yearly.TotalUnrealizedGainLossPercent)

		// Show yearly activity
		fmt.Printf("\n💰 %d ACTIVITY\n", yearly.Year)
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("Deposits:               %10.2f %s\n", yearly.YearlyDeposits, yearly.Currency)
		fmt.Printf("Dividends:              %10.2f %s\n", yearly.YearlyDividends, yearly.Currency)
		if yearly.YearlyInterest > 0 {
			fmt.Printf("Interest:               %10.2f %s\n", yearly.YearlyInterest, yearly.Currency)
		}

		// Show top holdings for this year
		if len(yearly.Positions) > 0 {
			if showAll {
				fmt.Printf("\n🏆 ALL HOLDINGS %d (%d positions)\n", yearly.Year, len(yearly.Positions))
			} else {
				fmt.Printf("\n🏆 TOP HOLDINGS %d\n", yearly.Year)
			}
			fmt.Println(strings.Repeat("-", 100))
			fmt.Printf("%-8s %-6s %-12s %-12s %-12s %-12s %-8s\n",
				"Ticker", "Shares", "Avg Cost", "Last Price", "Total Cost", "Market Val", "P&L %")
			fmt.Println(strings.Repeat("-", 100))

			// Display top N holdings
			limit := maxHoldings
			if showAll {
				limit = len(yearly.Positions)
			} else {
				if len(yearly.Positions) < limit {
					limit = len(yearly.Positions)
				}
			}

			for i := 0; i < limit; i++ {
				pos := yearly.Positions[i]
				fmt.Printf("%-8s %6.2f %12.2f %12.2f %12.2f %12.2f %7.1f%%\n",
					pos.Ticker,
					pos.Shares,
					pos.AverageCost,
					pos.LastPrice,
					pos.TotalCost,
					pos.MarketValue,
					pos.UnrealizedGainLossPercent)
			}

			// Show expand/collapse hint
			if !showAll && len(yearly.Positions) > maxHoldings {
				remaining := len(yearly.Positions) - maxHoldings
				fmt.Printf("\n📋 ... and %d more positions\n", remaining)
				fmt.Printf("💡 Use --show-all flag to see all positions: t212-taxes portfolio --dir exports --show-all\n")
			} else if showAll && len(yearly.Positions) > 10 {
				fmt.Printf("\n📋 All %d positions shown\n", len(yearly.Positions))
				fmt.Printf("💡 Use --max-holdings=N to limit display: t212-taxes portfolio --dir exports --max-holdings=5\n")
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}
