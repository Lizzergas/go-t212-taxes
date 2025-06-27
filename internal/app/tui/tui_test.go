package tui

import (
	"testing"
	"time"

	"github.com/Lizzergas/go-t212-taxes/internal/domain/types"
)

func TestNewAppWithAllData(t *testing.T) {
	// Create test data
	yearlyReports := []types.YearlyReport{
		{
			Year:               2021,
			TotalDeposits:      1000.0,
			TotalTransactions:  5,
			CapitalGains:       100.0,
			Dividends:          50.0,
			Interest:           10.0,
			TotalGains:         160.0,
			PercentageIncrease: 16.0,
			Currency:           "EUR",
		},
	}

	overallReport := &types.OverallReport{
		TotalDeposits:     1000.0,
		TotalTransactions: 5,
		TotalCapitalGains: 100.0,
		TotalDividends:    50.0,
		TotalInterest:     10.0,
		TotalGains:        160.0,
		OverallPercentage: 16.0,
		Years:             []int{2021},
		YearlyReports:     yearlyReports,
		Currency:          "EUR",
	}

	transactions := []types.Transaction{
		{
			Action: types.TransactionTypeMarketBuy,
			Time:   time.Date(2021, 10, 15, 10, 0, 0, 0, time.UTC),
		},
	}

	portfolioReport := &types.PortfolioValuationReport{
		YearlyPortfolios: []types.PortfolioSummary{
			{
				Year:                           2021,
				AsOfDate:                       time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC),
				TotalPositions:                 1,
				TotalShares:                    10.0,
				TotalInvested:                  1000.0,
				TotalMarketValue:               1100.0,
				TotalUnrealizedGainLoss:        100.0,
				TotalUnrealizedGainLossPercent: 10.0,
				Currency:                       "EUR",
			},
		},
		Currency:    "EUR",
		GeneratedAt: time.Now(),
		DataSource:  "Trading 212 CSV Export",
		PriceNote:   "Portfolio values based on last transaction price for each security",
	}

	incomeReport := &types.IncomeReport{
		Dividends: types.DividendSummary{
			TotalDividends:      50.0,
			TotalWithholdingTax: 5.0,
			NetDividends:        45.0,
			DividendCount:       2,
			Currency:            "EUR",
		},
		Interest: types.InterestSummary{
			TotalInterest: 10.0,
			InterestCount: 1,
			Currency:      "EUR",
		},
		TotalIncome: 55.0,
		Currency:    "EUR",
		DateRange: types.DateRange{
			From: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	// Create TUI model
	model := NewAppWithAllData(yearlyReports, overallReport, transactions, portfolioReport, incomeReport)

	// Verify model was created correctly
	if model == nil {
		t.Fatal("Expected model to be created, got nil")
	}

	if model.CurrentView != "yearly" {
		t.Errorf("Expected current view to be 'yearly', got '%s'", model.CurrentView)
	}

	if len(model.YearlyReports) != 1 {
		t.Errorf("Expected 1 yearly report, got %d", len(model.YearlyReports))
	}

	if model.OverallReport == nil {
		t.Error("Expected overall report to be set")
	}

	if len(model.AllTransactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(model.AllTransactions))
	}

	if model.PortfolioReport == nil {
		t.Error("Expected portfolio report to be set")
	}

	if model.IncomeReport == nil {
		t.Error("Expected income report to be set")
	}
}

func TestTUINavigation(t *testing.T) {
	// Create minimal test data
	model := NewApp()
	model.IncomeReport = &types.IncomeReport{
		TotalIncome: 100.0,
		Currency:    "EUR",
	}
	model.PortfolioReport = &types.PortfolioValuationReport{
		YearlyPortfolios: []types.PortfolioSummary{
			{Year: 2021, TotalInvested: 1000.0},
		},
		Currency: "EUR",
	}

	// Test view rendering doesn't crash
	views := []string{"yearly", "overall", "portfolio", "income", "help"}
	for _, view := range views {
		model.CurrentView = view
		result := model.View()
		if result == "" {
			t.Errorf("Expected non-empty view for %s", view)
		}
	}
}

func TestFormatIncomeReport(t *testing.T) {
	model := NewApp()

	incomeReport := types.IncomeReport{
		Dividends: types.DividendSummary{
			TotalDividends:      100.0,
			TotalWithholdingTax: 10.0,
			NetDividends:        90.0,
			DividendCount:       5,
			AverageYield:        2.5,
			Currency:            "EUR",
		},
		Interest: types.InterestSummary{
			TotalInterest: 20.0,
			InterestCount: 2,
			AverageRate:   1.5,
			Currency:      "EUR",
		},
		TotalIncome: 110.0,
		Currency:    "EUR",
		DateRange: types.DateRange{
			From: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	result := model.formatIncomeReport(incomeReport)

	// Check that the result contains expected elements
	if result == "" {
		t.Error("Expected non-empty income report format")
	}

	// Check for key income report elements
	expectedElements := []string{
		"Income Report",
		"2021-01-01",
		"2021-12-31",
		"100.00", // Total dividends
		"10.00",  // Withholding tax
		"90.00",  // Net dividends
		"20.00",  // Interest
		"110.00", // Total income
	}

	for _, element := range expectedElements {
		if !contains(result, element) {
			t.Errorf("Expected income report to contain '%s', but it didn't. Got: %s", element, result)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsInMiddle(str, substr)))
}

func containsInMiddle(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
			name:     "BGN amount",
			amount:   150.25,
			currency: "BGN",
			want:     "лв150.25",
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
			got := formatCurrency(tt.amount, tt.currency)
			if got != tt.want {
				t.Errorf("formatCurrency(%v, %s) = %s, want %s", tt.amount, tt.currency, got, tt.want)
			}
		})
	}
}
