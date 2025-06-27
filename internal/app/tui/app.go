package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	YearlyReports []types.YearlyReport
	OverallReport *types.OverallReport
	CurrentView   string // "yearly", "overall", "help"
	SelectedYear  int
	Width         int
	Height        int
}

// NewApp creates a new TUI application
func NewApp() *Model {
	return &Model{
		CurrentView:   "help",
		YearlyReports: []types.YearlyReport{},
		OverallReport: &types.OverallReport{},
	}
}

// NewAppWithData creates a new TUI application with data
func NewAppWithData(yearlyReports []types.YearlyReport, overallReport *types.OverallReport) *Model {
	view := "yearly"
	if len(yearlyReports) == 0 {
		view = "overall"
	}

	return &Model{
		CurrentView:   view,
		YearlyReports: yearlyReports,
		OverallReport: overallReport,
	}
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
				if m.SelectedYear > 0 {
					m.SelectedYear--
				}
			}
		case "down", "j":
			if m.CurrentView == "yearly" && len(m.YearlyReports) > 0 {
				if m.SelectedYear < len(m.YearlyReports)-1 {
					m.SelectedYear++
				}
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
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
	default:
		content = m.renderHelpView()
	}

	// Add navigation help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("y: yearly • o: overall • h: help • q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("🏦 Trading 212 Tax Calculator"),
		content,
		"",
		help,
	)
}

// renderYearlyView renders the yearly reports view
func (m Model) renderYearlyView() string {
	if len(m.YearlyReports) == 0 {
		return boxStyle.Render(warningStyle.Render("No yearly data available. Please process CSV files first."))
	}

	var content strings.Builder

	for i, report := range m.YearlyReports {
		// Highlight selected year
		style := boxStyle
		if i == m.SelectedYear {
			style = boxStyle.Copy().BorderForeground(lipgloss.Color("#F25D94"))
		}

		yearContent := m.formatYearlyReport(report)
		content.WriteString(style.Render(yearContent))
		content.WriteString("\n")
	}

	return content.String()
}

// renderOverallView renders the overall report view
func (m Model) renderOverallView() string {
	if m.OverallReport == nil {
		return boxStyle.Render(warningStyle.Render("No overall data available."))
	}

	content := m.formatOverallReport(*m.OverallReport)
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
   y - View yearly reports
   o - View overall summary
   h - Show this help
   ↑/↓ or k/j - Navigate between years (in yearly view)
   q - Quit

📁 File Structure:
   Your CSV files should follow this naming pattern:
   from_YYYY-MM-DD_to_YYYY-MM-DD_[hash].csv

💡 Features:
   • Yearly financial breakdowns
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