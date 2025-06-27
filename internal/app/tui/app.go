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
	YearlyReports       []types.YearlyReport
	OverallReport       *types.OverallReport
	AllTransactions     []types.Transaction              // Added for portfolio calculation
	PortfolioReport     *types.PortfolioValuationReport  // New: Full portfolio valuation data
	IncomeReport        *types.IncomeReport              // New: Income/dividend data
	CurrentView         string                           // "yearly", "overall", "portfolio", "income", "help"
	SelectedYear        int                              // Track which year's portfolio we're viewing
	CurrentPortfolio    *types.PortfolioSummary          // Current portfolio data
	PortfolioExpanded   bool                             // Track if portfolio positions are expanded
	GridCursor          GridCursor
	GridLayout          GridLayout
	Width               int
	Height              int
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

// NewAppWithAllData creates a new TUI application with all available data
func NewAppWithAllData(yearlyReports []types.YearlyReport, overallReport *types.OverallReport, transactions []types.Transaction, portfolioReport *types.PortfolioValuationReport, incomeReport *types.IncomeReport) *Model {
	view := "yearly"
	if len(yearlyReports) == 0 {
		view = "overall"
	}

	model := &Model{
		CurrentView:     view,
		YearlyReports:   yearlyReports,
		OverallReport:   overallReport,
		AllTransactions: transactions,
		PortfolioReport: portfolioReport,
		IncomeReport:    incomeReport,
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
		case "i":
			if m.IncomeReport != nil {
				m.CurrentView = "income"
			}
		case "p":
			// Show portfolio valuation if available
			if m.PortfolioReport != nil && len(m.PortfolioReport.YearlyPortfolios) > 0 {
				// Default to latest year if no year selected
				if m.SelectedYear == 0 && len(m.YearlyReports) > 0 {
					m.SelectedYear = m.YearlyReports[len(m.YearlyReports)-1].Year
				}
				// Find the portfolio for the selected year
				for _, portfolio := range m.PortfolioReport.YearlyPortfolios {
					if portfolio.Year == m.SelectedYear {
						m.CurrentPortfolio = &portfolio
						break
					}
				}
				m.CurrentView = "portfolio"
			}
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
		case "e", "x":
			if m.CurrentView == "portfolio" {
				// Toggle expand/collapse portfolio positions
				m.PortfolioExpanded = !m.PortfolioExpanded
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
	var title string
	var content string

	switch m.CurrentView {
	case "yearly":
		title = "📅 Yearly Reports"
		content = m.renderYearlyView()
	case "overall":
		title = "🌟 Overall Summary"
		content = m.renderOverallView()
	case "portfolio":
		title = "📊 Portfolio View"
		content = m.renderPortfolioView()
	case "income":
		title = "💰 Income Report"
		content = m.renderIncomeView()
	case "help":
		title = "❓ Help"
		content = m.renderHelpView()
	default:
		title = "🏠 Home"
		content = m.renderHelpView()
	}

	// Navigation hints
	navHints := "y: yearly • o: overall • p: portfolio • i: income • h: help • q: quit"
	if m.CurrentView == "portfolio" {
		navHints = "b: back to yearly • " + navHints
	}

	titleBar := titleStyle.Render(title)
	navigation := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(navHints)

	return fmt.Sprintf("%s\n%s\n%s", titleBar, content, navigation)
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
	content.WriteString(headerStyle.Render("📈 Portfolio Summary"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("📦 Total Positions: %s\n",
		valueStyle.Render(fmt.Sprintf("%d", portfolio.TotalPositions))))
	content.WriteString(fmt.Sprintf("📊 Total Shares: %s\n",
		valueStyle.Render(fmt.Sprintf("%.2f", portfolio.TotalShares))))
	content.WriteString(fmt.Sprintf("💰 Total Invested: %s\n",
		currencyStyle.Render(formatCurrency(portfolio.TotalInvested, portfolio.Currency))))
	content.WriteString(fmt.Sprintf("📈 Market Value: %s\n",
		currencyStyle.Render(formatCurrency(portfolio.TotalMarketValue, portfolio.Currency))))
	
	// Unrealized gains/losses with color coding
	gainLossText := fmt.Sprintf("%.2f %s (%.2f%%)", 
		portfolio.TotalUnrealizedGainLoss, portfolio.Currency, portfolio.TotalUnrealizedGainLossPercent)
	var gainLossStyled string
	if portfolio.TotalUnrealizedGainLoss > 0 {
		gainLossStyled = infoStyle.Render("📈 " + gainLossText)
	} else if portfolio.TotalUnrealizedGainLoss < 0 {
		gainLossStyled = errorStyle.Render("📉 " + gainLossText)
	} else {
		gainLossStyled = valueStyle.Render("➖ " + gainLossText)
	}
	content.WriteString(fmt.Sprintf("Unrealized P&L: %s\n", gainLossStyled))

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

	// Positions table with market data
	if len(portfolio.Positions) > 0 {
		content.WriteString("\n")
		if m.PortfolioExpanded {
			content.WriteString(headerStyle.Render(fmt.Sprintf("🎯 All Holdings (%d positions)", len(portfolio.Positions))))
		} else {
			content.WriteString(headerStyle.Render("🎯 Top Holdings"))
		}
		content.WriteString("\n")
		
		// Table header with market data
		content.WriteString(fmt.Sprintf("%-8s %-6s %-10s %-10s %-10s %-10s %-8s\n",
			"Ticker", "Shares", "Last Price", "Total Cost", "Market Val", "P&L", "P&L %"))
		content.WriteString(strings.Repeat("-", 75))
		content.WriteString("\n")

		// Position rows - expand/collapse logic
		limit := len(portfolio.Positions)
		if !m.PortfolioExpanded && limit > 10 {
			limit = 10
		}

		for i := 0; i < limit; i++ {
			pos := portfolio.Positions[i]
			plSign := ""
			if pos.UnrealizedGainLoss > 0 {
				plSign = "+"
			}
			
			content.WriteString(fmt.Sprintf("%-8s %6.1f %10.2f %10.2f %10.2f %s%7.0f %6.1f%%\n",
				pos.Ticker,
				pos.Shares,
				pos.LastPrice,
				pos.TotalCost,
				pos.MarketValue,
				plSign,
				pos.UnrealizedGainLoss,
				pos.UnrealizedGainLossPercent))
		}
		
		// Show expand/collapse information
		if !m.PortfolioExpanded && len(portfolio.Positions) > 10 {
			remaining := len(portfolio.Positions) - 10
			content.WriteString(fmt.Sprintf("\n📋 ... and %d more positions\n", remaining))
			content.WriteString(infoStyle.Render("Press 'e' or 'x' to expand all positions"))
		} else if m.PortfolioExpanded && len(portfolio.Positions) > 10 {
			content.WriteString(fmt.Sprintf("\n📋 All %d positions shown\n", len(portfolio.Positions)))
			content.WriteString(infoStyle.Render("Press 'e' or 'x' to collapse to top 10"))
		}
	}

	// Navigation help
	content.WriteString("\n\n")
	var navHelp string
	if m.PortfolioExpanded {
		navHelp = "e/x: collapse • b: back • i: income • o: overall • h: help • q: quit"
	} else {
		navHelp = "e/x: expand all • b: back • i: income • o: overall • h: help • q: quit"
	}
	
	navHelpStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(navHelp)
	content.WriteString(navHelpStyled)

	return boxStyle.Render(content.String())
}

// renderIncomeView renders the income report view
func (m Model) renderIncomeView() string {
	if m.IncomeReport == nil {
		return boxStyle.Render(warningStyle.Render("No income data available."))
	}

	content := m.formatIncomeReport(*m.IncomeReport)
	return boxStyle.Render(content)
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
   i - View income report
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
   • Market values based on last transaction prices
   • Unrealized gains/losses calculations
   • Total portfolio value tracking
   • Position details with P&L percentages
   • Top holdings ranked by market value
   • Yearly activity summary (deposits, dividends, interest)
   • Expand/collapse all positions (e/x keys)

💰 Income Features:
   • Detailed dividend analysis with withholding tax
   • Interest income tracking
   • Total income summaries
   • Dividend yield calculations
   • Transaction counts and breakdowns

🎮 Portfolio Controls:
   • e/x - Expand/collapse all positions
   • b - Go back to yearly view
   • ↑↓←→ - Navigate between years
   • Enter - Drill down to specific year

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

// formatIncomeReport formats the income report for display
func (m Model) formatIncomeReport(report types.IncomeReport) string {
	var content strings.Builder

	// Header
	content.WriteString(headerStyle.Render("💰 Income Report"))
	content.WriteString("\n\n")

	// Date range
	content.WriteString(fmt.Sprintf("📅 Period: %s - %s\n",
		report.DateRange.From.Format("2006-01-02"),
		report.DateRange.To.Format("2006-01-02")))

	content.WriteString("\n")

	// Dividend details
	content.WriteString(headerStyle.Render("💎 Dividends"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("Total Dividends: %s\n",
		currencyStyle.Render(formatCurrency(report.Dividends.TotalDividends, report.Currency))))
	content.WriteString(fmt.Sprintf("Withholding Tax: %s\n",
		currencyStyle.Render(formatCurrency(report.Dividends.TotalWithholdingTax, report.Currency))))
	content.WriteString(fmt.Sprintf("Net Dividends: %s\n",
		currencyStyle.Render(formatCurrency(report.Dividends.NetDividends, report.Currency))))
	content.WriteString(fmt.Sprintf("Dividend Count: %d\n", report.Dividends.DividendCount))
	if report.Dividends.AverageYield > 0 {
		content.WriteString(fmt.Sprintf("Average Yield: %.2f%%\n", report.Dividends.AverageYield))
	}

	// Interest details (if any)
	if report.Interest.TotalInterest > 0 {
		content.WriteString("\n")
		content.WriteString(headerStyle.Render("🏦 Interest"))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Total Interest: %s\n",
			currencyStyle.Render(formatCurrency(report.Interest.TotalInterest, report.Currency))))
		content.WriteString(fmt.Sprintf("Interest Count: %d\n", report.Interest.InterestCount))
		if report.Interest.AverageRate > 0 {
			content.WriteString(fmt.Sprintf("Average Rate: %.2f%%\n", report.Interest.AverageRate))
		}
	}

	// Total income
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("🎯 Total Income: %s\n",
		currencyStyle.Render(formatCurrency(report.TotalIncome, report.Currency))))

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