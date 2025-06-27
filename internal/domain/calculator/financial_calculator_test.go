package calculator

import (
	"testing"
	"time"

	"github.com/Lizzergas/go-t212-taxes/internal/domain/types"
)

func TestFinancialCalculator_CalculateYearlyReports(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	transactions := []types.Transaction{
		// 2023 transactions
		{
			Action:        types.TransactionTypeDeposit,
			Time:          time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Total:         floatPtr(1000.0),
			CurrencyTotal: stringPtr("EUR"),
		},
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
			Ticker:                stringPtr("AAPL"),
			Shares:                floatPtr(10),
			PricePerShare:         floatPtr(150.0),
			CurrencyPricePerShare: stringPtr("USD"),
			ExchangeRate:          floatPtr(0.9),
		},
		{
			Action:         types.TransactionTypeDividend,
			Time:           time.Date(2023, 3, 1, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(25.0),
			CurrencyResult: stringPtr("EUR"),
		},
		// 2024 transactions
		{
			Action:        types.TransactionTypeDeposit,
			Time:          time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			Total:         floatPtr(500.0),
			CurrencyTotal: stringPtr("EUR"),
		},
		{
			Action:         types.TransactionTypeInterest,
			Time:           time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(10.0),
			CurrencyResult: stringPtr("EUR"),
		},
	}

	reports, err := calc.CalculateYearlyReports(transactions)
	if err != nil {
		t.Fatalf("CalculateYearlyReports() error = %v", err)
	}

	if len(reports) != 2 {
		t.Errorf("CalculateYearlyReports() got %d reports, want 2", len(reports))
	}

	// Check 2023 report
	report2023 := findReportByYear(reports, 2023)
	if report2023 == nil {
		t.Fatal("2023 report not found")
	}

	if report2023.TotalDeposits != 1000.0 {
		t.Errorf("2023 TotalDeposits = %f, want 1000.0", report2023.TotalDeposits)
	}

	if report2023.Dividends != 25.0 {
		t.Errorf("2023 Dividends = %f, want 25.0", report2023.Dividends)
	}

	if report2023.TotalGains != 25.0 {
		t.Errorf("2023 TotalGains = %f, want 25.0", report2023.TotalGains)
	}

	expectedPercentage := (25.0 / 1000.0) * 100
	if report2023.PercentageIncrease != expectedPercentage {
		t.Errorf("2023 PercentageIncrease = %f, want %f", report2023.PercentageIncrease, expectedPercentage)
	}

	// Check 2024 report
	report2024 := findReportByYear(reports, 2024)
	if report2024 == nil {
		t.Fatal("2024 report not found")
	}

	if report2024.TotalDeposits != 500.0 {
		t.Errorf("2024 TotalDeposits = %f, want 500.0", report2024.TotalDeposits)
	}

	if report2024.Interest != 10.0 {
		t.Errorf("2024 Interest = %f, want 10.0", report2024.Interest)
	}
}

func TestFinancialCalculator_CalculateOverallReport(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	yearlyReports := []types.YearlyReport{
		{
			Year:               2023,
			TotalDeposits:      1000.0,
			TotalTransactions:  3,
			CapitalGains:       50.0,
			Dividends:          25.0,
			Interest:           0.0,
			TotalGains:         75.0,
			PercentageIncrease: 7.5,
			Currency:           "EUR",
		},
		{
			Year:               2024,
			TotalDeposits:      500.0,
			TotalTransactions:  2,
			CapitalGains:       30.0,
			Dividends:          15.0,
			Interest:           10.0,
			TotalGains:         55.0,
			PercentageIncrease: 11.0,
			Currency:           "EUR",
		},
	}

	overall := calc.CalculateOverallReport(yearlyReports)

	if overall.TotalDeposits != 1500.0 {
		t.Errorf("OverallReport TotalDeposits = %f, want 1500.0", overall.TotalDeposits)
	}

	if overall.TotalTransactions != 5 {
		t.Errorf("OverallReport TotalTransactions = %d, want 5", overall.TotalTransactions)
	}

	if overall.TotalCapitalGains != 80.0 {
		t.Errorf("OverallReport TotalCapitalGains = %f, want 80.0", overall.TotalCapitalGains)
	}

	if overall.TotalDividends != 40.0 {
		t.Errorf("OverallReport TotalDividends = %f, want 40.0", overall.TotalDividends)
	}

	if overall.TotalInterest != 10.0 {
		t.Errorf("OverallReport TotalInterest = %f, want 10.0", overall.TotalInterest)
	}

	if overall.TotalGains != 130.0 {
		t.Errorf("OverallReport TotalGains = %f, want 130.0", overall.TotalGains)
	}

	expectedOverallPercentage := (130.0 / 1500.0) * 100
	tolerance := 0.000001
	if abs(overall.OverallPercentage-expectedOverallPercentage) > tolerance {
		t.Errorf("OverallReport OverallPercentage = %f, want %f", overall.OverallPercentage, expectedOverallPercentage)
	}

	if len(overall.Years) != 2 {
		t.Errorf("OverallReport Years length = %d, want 2", len(overall.Years))
	}

	if overall.Years[0] != 2023 || overall.Years[1] != 2024 {
		t.Errorf("OverallReport Years = %v, want [2023 2024]", overall.Years)
	}
}

func TestFinancialCalculator_convertToBaseCurrency(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	tests := []struct {
		name         string
		amount       float64
		currency     *string
		exchangeRate *float64
		want         float64
	}{
		{
			name:     "same currency",
			amount:   100.0,
			currency: stringPtr("EUR"),
			want:     100.0,
		},
		{
			name:     "nil currency",
			amount:   100.0,
			currency: nil,
			want:     100.0,
		},
		{
			name:         "USD to EUR with exchange rate",
			amount:       100.0,
			currency:     stringPtr("USD"),
			exchangeRate: floatPtr(0.9),
			want:         111.11111111111111, // 100 / 0.9
		},
		{
			name:         "no exchange rate",
			amount:       100.0,
			currency:     stringPtr("USD"),
			exchangeRate: nil,
			want:         100.0,
		},
		{
			name:         "zero exchange rate",
			amount:       100.0,
			currency:     stringPtr("USD"),
			exchangeRate: floatPtr(0.0),
			want:         100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.convertToBaseCurrency(tt.amount, tt.currency, tt.exchangeRate)
			if got != tt.want {
				t.Errorf("convertToBaseCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinancialCalculator_CalculateCapitalGains(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	transactions := []types.Transaction{
		// Buy AAPL
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC),
			Ticker:                stringPtr("AAPL"),
			Shares:                floatPtr(10),
			PricePerShare:         floatPtr(100.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
		// Sell part of AAPL at profit
		{
			Action:                types.TransactionTypeMarketSell,
			Time:                  time.Date(2023, 3, 15, 10, 0, 0, 0, time.UTC),
			Ticker:                stringPtr("AAPL"),
			Shares:                floatPtr(5),
			PricePerShare:         floatPtr(120.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
		// Buy GOOGL
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC),
			Ticker:                stringPtr("GOOGL"),
			Shares:                floatPtr(5),
			PricePerShare:         floatPtr(200.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
		// Sell GOOGL at loss
		{
			Action:                types.TransactionTypeMarketSell,
			Time:                  time.Date(2023, 4, 1, 10, 0, 0, 0, time.UTC),
			Ticker:                stringPtr("GOOGL"),
			Shares:                floatPtr(5),
			PricePerShare:         floatPtr(180.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
	}

	gains, losses, err := calc.CalculateCapitalGains(transactions)
	if err != nil {
		t.Fatalf("CalculateCapitalGains() error = %v", err)
	}

	// AAPL: bought 5 shares at 100, sold at 120 = 100 profit
	expectedGains := 100.0
	if gains != expectedGains {
		t.Errorf("CalculateCapitalGains() gains = %f, want %f", gains, expectedGains)
	}

	// GOOGL: bought 5 shares at 200, sold at 180 = 100 loss
	expectedLosses := 100.0
	if losses != expectedLosses {
		t.Errorf("CalculateCapitalGains() losses = %f, want %f", losses, expectedLosses)
	}
}

func TestFinancialCalculator_isTradeTransaction(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	tests := []struct {
		action types.TransactionType
		want   bool
	}{
		{types.TransactionTypeMarketBuy, true},
		{types.TransactionTypeMarketSell, true},
		{types.TransactionTypeLimitBuy, true},
		{types.TransactionTypeLimitSell, true},
		{types.TransactionTypeStopBuy, true},
		{types.TransactionTypeStopSell, true},
		{types.TransactionTypeDividend, false},
		{types.TransactionTypeInterest, false},
		{types.TransactionTypeDeposit, false},
		{types.TransactionTypeWithdrawal, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			got := calc.isTradeTransaction(tt.action)
			if got != tt.want {
				t.Errorf("isTradeTransaction(%v) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestFinancialCalculator_isBuyTransaction(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	tests := []struct {
		action types.TransactionType
		want   bool
	}{
		{types.TransactionTypeMarketBuy, true},
		{types.TransactionTypeLimitBuy, true},
		{types.TransactionTypeStopBuy, true},
		{types.TransactionTypeMarketSell, false},
		{types.TransactionTypeLimitSell, false},
		{types.TransactionTypeStopSell, false},
		{types.TransactionTypeDividend, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			got := calc.isBuyTransaction(tt.action)
			if got != tt.want {
				t.Errorf("isBuyTransaction(%v) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestFinancialCalculator_calculateSecurityGainsLosses(t *testing.T) {
	calc := NewFinancialCalculator("EUR")

	// Test FIFO calculation for a single security
	transactions := []types.Transaction{
		// First buy
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Shares:                floatPtr(10),
			PricePerShare:         floatPtr(100.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
		// Second buy at different price
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC),
			Shares:                floatPtr(10),
			PricePerShare:         floatPtr(110.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
		// Sell - should use FIFO (first 5 at 100, next 5 at 110)
		{
			Action:                types.TransactionTypeMarketSell,
			Time:                  time.Date(2023, 3, 1, 10, 0, 0, 0, time.UTC),
			Shares:                floatPtr(10),
			PricePerShare:         floatPtr(120.0),
			CurrencyPricePerShare: stringPtr("EUR"),
		},
	}

	gains, losses, err := calc.calculateSecurityGainsLosses(transactions)
	if err != nil {
		t.Fatalf("calculateSecurityGainsLosses() error = %v", err)
	}

	// Bought: 10 shares at 100, 10 shares at 110
	// Sold: 10 shares at 120 using FIFO (first 10 shares bought at 100)
	// Gains: 10 * 120 - 10 * 100 = 1200 - 1000 = 200
	expectedGains := 200.0
	if gains != expectedGains {
		t.Errorf("calculateSecurityGainsLosses() gains = %f, want %f", gains, expectedGains)
	}

	expectedLosses := 0.0
	if losses != expectedLosses {
		t.Errorf("calculateSecurityGainsLosses() losses = %f, want %f", losses, expectedLosses)
	}
}

// Helper functions
func findReportByYear(reports []types.YearlyReport, year int) *types.YearlyReport {
	for _, report := range reports {
		if report.Year == year {
			return &report
		}
	}
	return nil
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
