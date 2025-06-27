package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"t212-taxes/internal/domain/calculator"
	"t212-taxes/internal/domain/types"
)

var (
	// Styles for the TUI
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Margin(1, 0)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8C00"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	currencyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF"))
)

// Model represents the TUI application state
type Model struct {
	YearlyReports     []types.YearlyReport
	OverallReport     *types.OverallReport
	AllTransactions   []types.Transaction // Added for portfolio calculation
	CurrentView       string              // "yearly", "overall", "portfolio", "help"
	SelectedYear      int                 // Track which year's portfolio we're viewing
	CurrentPortfolio  *types.PortfolioSummary // Current portfolio data
	GridCursor        GridCursor
	GridLayout        GridLayout
	Width             int
	Height            int
}

// GridCursor represents the current position in the grid
type GridCursor struct {
	Row int
	Col int
}

// GridLayout holds grid configuration
type GridLayout struct {
	Columns     int
	Rows        int
	TotalItems  int
	ItemWidth   int
	ItemHeight  int
}

// NewApp creates a new TUI application
func NewApp() *Model {
	return &Model{
		CurrentView:   "help",
		YearlyReports: []types.YearlyReport{},
		OverallReport: &types.OverallReport{},
		GridCursor:    GridCursor{Row: 0, Col: 0},
	}
}

// NewAppWithData creates a new TUI application with data
func NewAppWithData(yearlyReports []types.YearlyReport, overallReport *types.OverallReport) *Model {
	view := "yearly"
	if len(yearlyReports) == 0 {
		view = "overall"
	}

	model := &Model{
		CurrentView:   view,
		YearlyReports: yearlyReports,
		OverallReport: overallReport,
		GridCursor:    GridCursor{Row: 0, Col: 0},
	}

	// Calculate initial grid layout
	model.updateGridLayout()

	return model
}

// NewAppWithPortfolioData creates a new TUI application with full data including transactions
func NewAppWithPortfolioData(yearlyReports []types.YearlyReport, overallReport *types.OverallReport, transactions []types.Transaction) *Model {
	view := "yearly"
	if len(yearlyReports) == 0 {
		view = "overall"
	}

	model := &Model{
		CurrentView:     view,
		YearlyReports:   yearlyReports,
		OverallReport:   overallReport,
		AllTransactions: transactions,
		GridCursor:      GridCursor{Row: 0, Col: 0},
	}

	// Calculate initial grid layout
	model.updateGridLayout()

	return model
}

// Run starts the TUI application
func (m *Model) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Init implements the bubbletea.Model interface
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements the bubbletea.Model interface
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y":
			m.CurrentView = "yearly"
		case "o":
			m.CurrentView = "overall"
		case "h", "?":
			m.CurrentView = "help"
		case "up", "k":
			if m.CurrentView == "yearly" && len(m.YearlyReports) > 0 {
				newRow := m.GridCursor.Row - 1
				if m.isValidPosition(newRow, m.GridCursor.Col) {
					m.GridCursor.Row = newRow
				}
			}
		case "down", "j":
			if m.CurrentView == "yearly" && len(m.YearlyReports) > 0 {
				newRow := m.GridCursor.Row + 1
				if m.isValidPosition(newRow, m.GridCursor.Col) {
					m.GridCursor.Row = newRow
				}
			}
		case "left":
			if m.CurrentView == "yearly" && len(m.YearlyReports) > 0 {
				newCol := m.GridCursor.Col - 1
				if m.isValidPosition(m.GridCursor.Row, newCol) {
					m.GridCursor.Col = newCol
				}
			}
		case "right", "l":
			if m.CurrentView == "yearly" && len(m.YearlyReports) > 0 {
				newCol := m.GridCursor.Col + 1
				if m.isValidPosition(m.GridCursor.Row, newCol) {
					m.GridCursor.Col = newCol
				}
			}
		case "enter", " ":
			if m.CurrentView == "yearly" && len(m.YearlyReports) > 0 && len(m.AllTransactions) > 0 {
				// Navigate to portfolio view for selected year
				selectedIndex := m.getSelectedIndex()
				if selectedIndex < len(m.YearlyReports) {
					selectedYear := m.YearlyReports[selectedIndex].Year
					m.SelectedYear = selectedYear
					
					// Calculate portfolio for the selected year
					portfolioCalc := calculator.NewPortfolioCalculator("EUR") // TODO: Make currency configurable
					m.CurrentPortfolio = portfolioCalc.CalculateEndOfYearPortfolio(m.AllTransactions, selectedYear)
					m.CurrentView = "portfolio"
				}
			}
		case "b":
			if m.CurrentView == "portfolio" {
				// Go back to yearly view
				m.CurrentView = "yearly"
				m.CurrentPortfolio = nil
			}
		case "p":
			if m.CurrentPortfolio != nil {
				m.CurrentView = "portfolio"
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Recalculate grid layout when window is resized
		m.updateGridLayout()
	}

	return m, nil
}

// View implements the bubbletea.Model interface
func (m Model) View() string {
	var content string

	switch m.CurrentView {
	case "yearly":
		content = m.renderYearlyView()
	case "overall":
		content = m.renderOverallView()
	case "portfolio":
		content = m.renderPortfolioView()
	default:
		content = m.renderHelpView()
	}

	// Add navigation help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("y: yearly grid • o: overall • p: portfolio • h: help • ↑↓←→: navigate • enter: drill down • b: back • q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("🏦 Trading 212 Tax Calculator"),
		content,
		"",
		help,
	)
}

// renderYearlyView renders the yearly reports view in a grid layout
func (m Model) renderYearlyView() string {
	if len(m.YearlyReports) == 0 {
		return boxStyle.Render(warningStyle.Render("No yearly data available. Please process CSV files first."))
	}

	// Build grid row by row
	var gridRows []string

	for row := 0; row < m.GridLayout.Rows; row++ {
		var rowItems []string

		for col := 0; col < m.GridLayout.Columns; col++ {
			index := row*m.GridLayout.Columns + col
			
			// Check if this position has a valid item
			if index >= len(m.YearlyReports) {
				// Add empty space for positioning
				emptyCard := lipgloss.NewStyle().
					Width(m.GridLayout.ItemWidth).
					Height(m.GridLayout.ItemHeight).
					Render("")
				rowItems = append(rowItems, emptyCard)
				continue
			}

			report := m.YearlyReports[index]
			
			// Determine if this item is selected
			isSelected := (row == m.GridCursor.Row && col == m.GridCursor.Col)
			
			// Create the year card
			yearCard := m.createYearCard(report, isSelected)
			rowItems = append(rowItems, yearCard)
		}

		// Join items in this row horizontally
		rowContent := lipgloss.JoinHorizontal(lipgloss.Left, rowItems...)
		gridRows = append(gridRows, rowContent)
	}

	// Join all rows vertically
	gridContent := lipgloss.JoinVertical(lipgloss.Left, gridRows...)
	
	// Add grid navigation info
	selectedIndex := m.getSelectedIndex()
	navInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(fmt.Sprintf("Grid: %d×%d | Selected: %d/%d | ↑↓←→ navigate • Enter: select", 
			m.GridLayout.Rows, m.GridLayout.Columns, selectedIndex+1, len(m.YearlyReports)))

	return lipgloss.JoinVertical(lipgloss.Left, gridContent, "", navInfo)
}

// createYearCard creates a formatted card for a single year report
func (m Model) createYearCard(report types.YearlyReport, isSelected bool) string {
	// Choose style based on selection
	cardStyle := boxStyle.Copy().
		Width(m.GridLayout.ItemWidth).
		Height(m.GridLayout.ItemHeight)

	if isSelected {
		cardStyle = cardStyle.BorderForeground(lipgloss.Color("#F25D94"))
	}

	// Create labeled content with compact formatting
	var content strings.Builder
	
	// Header with year
	content.WriteString(headerStyle.Render(fmt.Sprintf("📅 %d", report.Year)))
	content.WriteString("\n")

	// Key metrics with clear labels
	content.WriteString(fmt.Sprintf("💰 Deposits: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalDeposits, report.Currency))))
	
	content.WriteString(fmt.Sprintf("📈 Gains: %s\n",
		currencyStyle.Render(formatCurrency(report.CapitalGains, report.Currency))))
	
	content.WriteString(fmt.Sprintf("💎 Dividends: %s\n",
		currencyStyle.Render(formatCurrency(report.Dividends, report.Currency))))
	
	if report.Interest > 0 {
		content.WriteString(fmt.Sprintf("🏦 Interest: %s\n",
			currencyStyle.Render(formatCurrency(report.Interest, report.Currency))))
	}

	// Total gains line
	content.WriteString(fmt.Sprintf("🎯 Total: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalGains, report.Currency))))

	// Transactions count and percentage on last line
	percentageText := fmt.Sprintf("%.1f%%", report.PercentageIncrease)
	content.WriteString(fmt.Sprintf("📊 %d txns • %s",
		report.TotalTransactions, percentageText))

	return cardStyle.Render(content.String())
}

// renderOverallView renders the overall report view
func (m Model) renderOverallView() string {
	if m.OverallReport == nil {
		return boxStyle.Render(warningStyle.Render("No overall data available."))
	}

	content := m.formatOverallReport(*m.OverallReport)
	return boxStyle.Render(content)
}

// renderPortfolioView renders the portfolio view for a specific year
func (m Model) renderPortfolioView() string {
	if m.CurrentPortfolio == nil {
		return boxStyle.Render(warningStyle.Render("No portfolio data available."))
	}

	portfolio := *m.CurrentPortfolio
	var content strings.Builder

	// Header
	content.WriteString(headerStyle.Render(fmt.Sprintf("📊 Portfolio as of Dec 31, %d", portfolio.Year)))
	content.WriteString("\n\n")

	// Summary statistics
	content.WriteString(headerStyle.Render("📈 Summary"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("📦 Total Positions: %s\n",
		valueStyle.Render(fmt.Sprintf("%d", portfolio.TotalPositions))))
	content.WriteString(fmt.Sprintf("📊 Total Shares: %s\n",
		valueStyle.Render(fmt.Sprintf("%.2f", portfolio.TotalShares))))
	content.WriteString(fmt.Sprintf("💰 Total Invested: %s\n",
		currencyStyle.Render(formatCurrency(portfolio.TotalInvested, portfolio.Currency))))

	// Yearly activity
	content.WriteString("\n")
	content.WriteString(headerStyle.Render(fmt.Sprintf("💡 %d Activity", portfolio.Year)))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("💰 Deposits: %s\n",
		currencyStyle.Render(formatCurrency(portfolio.YearlyDeposits, portfolio.Currency))))
	content.WriteString(fmt.Sprintf("💎 Dividends: %s\n",
		currencyStyle.Render(formatCurrency(portfolio.YearlyDividends, portfolio.Currency))))
	if portfolio.YearlyInterest > 0 {
		content.WriteString(fmt.Sprintf("🏦 Interest: %s\n",
			currencyStyle.Render(formatCurrency(portfolio.YearlyInterest, portfolio.Currency))))
	}

	// Positions table
	if len(portfolio.Positions) > 0 {
		content.WriteString("\n")
		content.WriteString(headerStyle.Render("🎯 Holdings"))
		content.WriteString("\n")
		
		// Table header
		content.WriteString(fmt.Sprintf("%-12s %-10s %-12s %-12s %-8s\n",
			"Ticker", "Shares", "Avg Cost", "Total Cost", "Txns"))
		content.WriteString(strings.Repeat("-", 70))
		content.WriteString("\n")

		// Position rows
		for _, pos := range portfolio.Positions {
			content.WriteString(fmt.Sprintf("%-12s %-10.2f %-12s %-12s %-8d\n",
				pos.Ticker,
				pos.Shares,
				formatCurrency(pos.AverageCost, pos.Currency),
				formatCurrency(pos.TotalCost, pos.Currency),
				pos.TransactionCount))
		}
	}

	// Navigation help
	content.WriteString("\n")
	navHelp := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("b: back to yearly view • p: portfolio • o: overall • h: help • q: quit")
	content.WriteString(navHelp)

	return boxStyle.Render(content.String())
}

// renderHelpView renders the help view
func (m Model) renderHelpView() string {
	help := `Welcome to Trading 212 Tax Calculator!

This tool helps you analyze your Trading 212 CSV exports and calculate
financial metrics for tax purposes.

🚀 Getting Started:
   1. Export your Trading 212 data as CSV files (yearly exports recommended)
   2. Use the command line to process your files:
      t212-taxes analyze --dir ./exports

📊 Navigation:
   y - View yearly reports (grid layout)
   o - View overall summary
   p - View portfolio (if available)
   h - Show this help
   ↑↓←→ or k/j - Navigate grid (in yearly view)
   Enter/Space - Drill down to portfolio (in yearly view)
   b - Go back (from portfolio view)
   q - Quit

🎯 Grid Features:
   • Years displayed in adaptive grid layout
   • Navigate with arrow keys in all directions  
   • Grid automatically adjusts to terminal size
   • Selected year highlighted with pink border
   • Press Enter to view end-of-year portfolio

📊 Portfolio Features:
   • Shows holdings as of December 31st
   • Position details with shares and costs
   • Yearly activity summary
   • Transaction counts per position

📁 File Structure:
   Your CSV files should follow this naming pattern:
   from_YYYY-MM-DD_to_YYYY-MM-DD_[hash].csv

💡 Features:
   • Yearly financial breakdowns in grid format
   • Capital gains calculations
   • Dividend and interest tracking
   • Deposit summaries
   • Investment performance metrics

⚠️  Disclaimer:
   This tool provides estimates for informational purposes.
   Always consult a tax professional for official tax advice.`

	return boxStyle.Render(infoStyle.Render(help))
}

// formatYearlyReport formats a yearly report for display
func (m Model) formatYearlyReport(report types.YearlyReport) string {
	var content strings.Builder

	// Header
	content.WriteString(headerStyle.Render(fmt.Sprintf("📅 %d Financial Overview", report.Year)))
	content.WriteString("\n")

	// Metrics
	content.WriteString(fmt.Sprintf("💰 Deposits: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalDeposits, report.Currency))))
	
	content.WriteString(fmt.Sprintf("💳 Transactions: %s\n",
		valueStyle.Render(fmt.Sprintf("%d", report.TotalTransactions))))
	
	content.WriteString(fmt.Sprintf("📈 Capital Gains: %s\n",
		currencyStyle.Render(formatCurrency(report.CapitalGains, report.Currency))))
	
	content.WriteString(fmt.Sprintf("💎 Dividends: %s\n",
		currencyStyle.Render(formatCurrency(report.Dividends, report.Currency))))
	
	if report.Interest > 0 {
		content.WriteString(fmt.Sprintf("🏦 Interest: %s\n",
			currencyStyle.Render(formatCurrency(report.Interest, report.Currency))))
	}
	
	content.WriteString(fmt.Sprintf("🎯 Total Gains: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalGains, report.Currency))))
	
	// Percentage with color coding
	percentageText := fmt.Sprintf("%.2f%%", report.PercentageIncrease)
	var percentageStyled string
	if report.PercentageIncrease > 0 {
		percentageStyled = infoStyle.Render(percentageText)
	} else if report.PercentageIncrease < 0 {
		percentageStyled = errorStyle.Render(percentageText)
	} else {
		percentageStyled = valueStyle.Render(percentageText)
	}
	
	content.WriteString(fmt.Sprintf("📊 Money Increase: %s", percentageStyled))

	return content.String()
}

// formatOverallReport formats the overall report for display
func (m Model) formatOverallReport(report types.OverallReport) string {
	var content strings.Builder

	// Header
	content.WriteString(headerStyle.Render("🌟 Overall Investment Summary"))
	content.WriteString("\n\n")

	// Years covered
	if len(report.Years) > 0 {
		yearsStr := fmt.Sprintf("%d", report.Years[0])
		if len(report.Years) > 1 {
			yearsStr = fmt.Sprintf("%d - %d", report.Years[0], report.Years[len(report.Years)-1])
		}
		content.WriteString(fmt.Sprintf("📅 Period: %s\n", valueStyle.Render(yearsStr)))
	}

	content.WriteString("\n")

	// Overall metrics
	content.WriteString(fmt.Sprintf("💰 Total Deposits: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalDeposits, report.Currency))))
	
	content.WriteString(fmt.Sprintf("💳 Total Transactions: %s\n",
		valueStyle.Render(fmt.Sprintf("%d", report.TotalTransactions))))
	
	content.WriteString(fmt.Sprintf("📈 Total Capital Gains: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalCapitalGains, report.Currency))))
	
	content.WriteString(fmt.Sprintf("💎 Total Dividends: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalDividends, report.Currency))))
	
	if report.TotalInterest > 0 {
		content.WriteString(fmt.Sprintf("🏦 Total Interest: %s\n",
			currencyStyle.Render(formatCurrency(report.TotalInterest, report.Currency))))
	}
	
	content.WriteString(fmt.Sprintf("🎯 Total Gains: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalGains, report.Currency))))

	content.WriteString("\n")

	// Overall performance with color coding
	percentageText := fmt.Sprintf("%.2f%%", report.OverallPercentage)
	var percentageStyled string
	if report.OverallPercentage > 0 {
		percentageStyled = infoStyle.Render(percentageText)
	} else if report.OverallPercentage < 0 {
		percentageStyled = errorStyle.Render(percentageText)
	} else {
		percentageStyled = valueStyle.Render(percentageText)
	}
	
	content.WriteString(fmt.Sprintf("📊 Overall Performance: %s", percentageStyled))

	// Investment efficiency
	if report.TotalDeposits > 0 {
		content.WriteString("\n\n")
		content.WriteString(headerStyle.Render("📊 Investment Efficiency"))
		content.WriteString("\n")
		
		avgPerYear := report.OverallPercentage / float64(len(report.Years))
		content.WriteString(fmt.Sprintf("📈 Average Annual Return: %s\n",
			valueStyle.Render(fmt.Sprintf("%.2f%%", avgPerYear))))
		
		if len(report.Years) > 1 {
			totalValue := report.TotalDeposits + report.TotalGains
			content.WriteString(fmt.Sprintf("💼 Total Invested + Realized Gains: %s\n",
				currencyStyle.Render(formatCurrency(totalValue, report.Currency))))
			content.WriteString(fmt.Sprintf("💡 Note: This is deposits + realized gains, not current market value\n"))
		}
	}

	return content.String()
}

// formatCurrency formats currency values for display
func formatCurrency(amount float64, currency string) string {
	symbol := getCurrencySymbol(currency)
	
	if amount >= 0 {
		return fmt.Sprintf("%s%.2f", symbol, amount)
	} else {
		return fmt.Sprintf("-%s%.2f", symbol, -amount)
	}
}

// getCurrencySymbol returns the symbol for a currency
func getCurrencySymbol(currency string) string {
	switch currency {
	case "EUR":
		return "€"
	case "USD":
		return "$"
	case "GBP":
		return "£"
	case "BGN":
		return "лв"
	default:
		return currency + " "
	}
}

// PrintReportsTable prints reports in table format to console
func PrintReportsTable(yearlyReports []types.YearlyReport, overallReport *types.OverallReport) {
	fmt.Println("📊 Trading 212 Financial Reports")
	fmt.Println("================================")
	fmt.Println()

	// Print yearly reports
	for _, report := range yearlyReports {
		fmt.Printf("📅 %d Financial Overview\n", report.Year)
		fmt.Printf("💰 Deposits: %s\n", formatCurrency(report.TotalDeposits, report.Currency))
		fmt.Printf("💳 Transactions: %d\n", report.TotalTransactions)
		fmt.Printf("📈 Capital Gains: %s\n", formatCurrency(report.CapitalGains, report.Currency))
		fmt.Printf("💎 Dividends: %s\n", formatCurrency(report.Dividends, report.Currency))
		if report.Interest > 0 {
			fmt.Printf("🏦 Interest: %s\n", formatCurrency(report.Interest, report.Currency))
		}
		fmt.Printf("🎯 Total Gains: %s\n", formatCurrency(report.TotalGains, report.Currency))
		fmt.Printf("📊 Money Increase: %.2f%%\n", report.PercentageIncrease)
		fmt.Println()
	}

	// Print overall report
	if overallReport != nil {
		fmt.Println("🌟 Overall Investment Summary")
		fmt.Println("============================")
		if len(overallReport.Years) > 0 {
			yearsStr := fmt.Sprintf("%d", overallReport.Years[0])
			if len(overallReport.Years) > 1 {
				yearsStr = fmt.Sprintf("%d - %d", overallReport.Years[0], overallReport.Years[len(overallReport.Years)-1])
			}
			fmt.Printf("📅 Period: %s\n", yearsStr)
		}
		fmt.Printf("💰 Total Deposits: %s\n", formatCurrency(overallReport.TotalDeposits, overallReport.Currency))
		fmt.Printf("💳 Total Transactions: %d\n", overallReport.TotalTransactions)
		fmt.Printf("📈 Total Capital Gains: %s\n", formatCurrency(overallReport.TotalCapitalGains, overallReport.Currency))
		fmt.Printf("💎 Total Dividends: %s\n", formatCurrency(overallReport.TotalDividends, overallReport.Currency))
		if overallReport.TotalInterest > 0 {
			fmt.Printf("🏦 Total Interest: %s\n", formatCurrency(overallReport.TotalInterest, overallReport.Currency))
		}
		fmt.Printf("🎯 Total Gains: %s\n", formatCurrency(overallReport.TotalGains, overallReport.Currency))
		fmt.Printf("📊 Overall Performance: %.2f%%\n", overallReport.OverallPercentage)
		fmt.Println()
	}
}

// updateGridLayout calculates the optimal grid layout based on available space and data
func (m *Model) updateGridLayout() {
	if len(m.YearlyReports) == 0 {
		return
	}

	// Calculate optimal grid dimensions
	totalItems := len(m.YearlyReports)
	
	// Base item dimensions (minimum required space for a year card)
	minItemWidth := 35   // Minimum width for year card
	minItemHeight := 14  // Minimum height for year card (increased for labels)
	
	// Calculate available space (accounting for borders, padding, and navigation help)
	availableWidth := m.Width - 4   // Account for margins
	availableHeight := m.Height - 8 // Account for title, help text, margins
	
	// Calculate optimal number of columns
	maxCols := max(1, availableWidth/minItemWidth)
	optimalCols := min(maxCols, totalItems)
	
	// Try different column counts to find the best fit
	bestCols := 1
	for cols := 1; cols <= optimalCols; cols++ {
		rows := (totalItems + cols - 1) / cols // Ceiling division
		if rows*minItemHeight <= availableHeight {
			bestCols = cols
		}
	}
	
	// Calculate final layout
	columns := bestCols
	rows := (totalItems + columns - 1) / columns
	itemWidth := min(availableWidth/columns, 45) // Max width to prevent overly wide cards
	itemHeight := minItemHeight
	
	m.GridLayout = GridLayout{
		Columns:     columns,
		Rows:        rows,
		TotalItems:  totalItems,
		ItemWidth:   itemWidth,
		ItemHeight:  itemHeight,
	}
	
	// Ensure cursor is within bounds
	maxRow := rows - 1
	maxCol := columns - 1
	
	if m.GridCursor.Row > maxRow {
		m.GridCursor.Row = maxRow
	}
	if m.GridCursor.Col > maxCol {
		m.GridCursor.Col = maxCol
	}
	
	// Ensure cursor points to a valid item
	if m.getSelectedIndex() >= totalItems {
		// Move to last valid position
		lastItemIndex := totalItems - 1
		m.GridCursor.Row = lastItemIndex / columns
		m.GridCursor.Col = lastItemIndex % columns
	}
}

// getSelectedIndex returns the array index of the currently selected item
func (m *Model) getSelectedIndex() int {
	return m.GridCursor.Row*m.GridLayout.Columns + m.GridCursor.Col
}

// isValidPosition checks if the cursor position points to a valid item
func (m *Model) isValidPosition(row, col int) bool {
	if row < 0 || col < 0 || row >= m.GridLayout.Rows || col >= m.GridLayout.Columns {
		return false
	}
	index := row*m.GridLayout.Columns + col
	return index < m.GridLayout.TotalItems
}

// Helper functions for grid calculations
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}