package types_test

import (
	"testing"
	"time"

	"t212-taxes/internal/domain/types"
)

func TestTransactionTypes(t *testing.T) {
	tests := []struct {
		name        string
		txType      types.TransactionType
		expectedStr string
	}{
		{"Market Buy", types.TransactionTypeMarketBuy, "Market buy"},
		{"Market Sell", types.TransactionTypeMarketSell, "Market sell"},
		{"Dividend", types.TransactionTypeDividend, "Dividend"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.txType) != tt.expectedStr {
				t.Errorf("expected %s, got %s", tt.expectedStr, string(tt.txType))
			}
		})
	}
}

func TestCurrencyTypes(t *testing.T) {
	tests := []struct {
		name        string
		currency    types.Currency
		expectedStr string
	}{
		{"USD", types.CurrencyUSD, "USD"},
		{"EUR", types.CurrencyEUR, "EUR"},
		{"GBP", types.CurrencyGBP, "GBP"},
		{"BGN", types.CurrencyBGN, "BGN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.currency) != tt.expectedStr {
				t.Errorf("expected %s, got %s", tt.expectedStr, string(tt.currency))
			}
		})
	}
}

func TestTransaction(t *testing.T) {
	// Helper functions for pointers
	stringPtr := func(s string) *string { return &s }
	floatPtr := func(f float64) *float64 { return &f }

	transaction := types.Transaction{
		Action:                types.TransactionTypeMarketBuy,
		Time:                  time.Now(),
		ISIN:                  stringPtr("US0378331005"),
		Ticker:                stringPtr("AAPL"),
		Name:                  stringPtr("Apple Inc."),
		Shares:                floatPtr(10),
		PricePerShare:         floatPtr(150.0),
		CurrencyPricePerShare: stringPtr("USD"),
		ExchangeRate:          floatPtr(1.0),
		Result:                floatPtr(-1500.0),
		Total:                 floatPtr(-1500.0),
	}

	if transaction.Action != types.TransactionTypeMarketBuy {
		t.Errorf("expected %s, got %s", types.TransactionTypeMarketBuy, transaction.Action)
	}

	if transaction.ISIN == nil || *transaction.ISIN != "US0378331005" {
		t.Errorf("expected ISIN US0378331005, got %v", transaction.ISIN)
	}

	if transaction.Ticker == nil || *transaction.Ticker != "AAPL" {
		t.Errorf("expected Ticker AAPL, got %v", transaction.Ticker)
	}
}

func TestYearlyReport(t *testing.T) {
	report := types.YearlyReport{
		Year:               2024,
		TotalDeposits:      1000.0,
		TotalTransactions:  5,
		CapitalGains:       100.0,
		Dividends:          25.0,
		Interest:           5.0,
		TotalGains:         130.0,
		PercentageIncrease: 13.0,
		Currency:           "EUR",
	}

	if report.Year != 2024 {
		t.Errorf("expected Year 2024, got %d", report.Year)
	}

	if report.TotalGains != 130.0 {
		t.Errorf("expected TotalGains 130.0, got %f", report.TotalGains)
	}

	if report.Currency != "EUR" {
		t.Errorf("expected Currency EUR, got %s", report.Currency)
	}
}

func TestOverallReport(t *testing.T) {
	yearlyReports := []types.YearlyReport{
		{Year: 2023, TotalDeposits: 500.0, TotalGains: 50.0, Currency: "EUR"},
		{Year: 2024, TotalDeposits: 500.0, TotalGains: 80.0, Currency: "EUR"},
	}

	overall := types.OverallReport{
		TotalDeposits:     1000.0,
		TotalTransactions: 10,
		TotalCapitalGains: 100.0,
		TotalDividends:    20.0,
		TotalInterest:     10.0,
		TotalGains:        130.0,
		OverallPercentage: 13.0,
		Years:             []int{2023, 2024},
		YearlyReports:     yearlyReports,
		Currency:          "EUR",
	}

	if overall.TotalDeposits != 1000.0 {
		t.Errorf("expected TotalDeposits 1000.0, got %f", overall.TotalDeposits)
	}

	if len(overall.Years) != 2 {
		t.Errorf("expected 2 years, got %d", len(overall.Years))
	}

	if overall.Years[0] != 2023 || overall.Years[1] != 2024 {
		t.Errorf("expected years [2023, 2024], got %v", overall.Years)
	}
}
