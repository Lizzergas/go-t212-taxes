package calculator

import (
	"testing"
	"time"

	"github.com/Lizzergas/go-t212-taxes/internal/domain/types"
)

func TestPortfolioCalculator_CalculatePortfolioValuation(t *testing.T) {
	calculator := NewPortfolioCalculator("EUR")

	msftISIN := "US5949181045"
	msftTicker := "MSFT"
	msftName := "Microsoft"
	shares10 := 10.0
	shares5 := 5.0
	price100 := 100.0
	price120 := 120.0
	total1000 := 1000.0
	total600 := 600.0
	eurCurrency := "EUR"

	transactions := []types.Transaction{
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2021, 10, 15, 10, 0, 0, 0, time.UTC),
			ISIN:                  &msftISIN,
			Ticker:                &msftTicker,
			Name:                  &msftName,
			Shares:                &shares10,
			PricePerShare:         &price100,
			CurrencyPricePerShare: &eurCurrency,
			Total:                 &total1000,
			CurrencyTotal:         &eurCurrency,
		},
		{
			Action:                types.TransactionTypeMarketBuy,
			Time:                  time.Date(2021, 12, 15, 10, 0, 0, 0, time.UTC),
			ISIN:                  &msftISIN,
			Ticker:                &msftTicker,
			Name:                  &msftName,
			Shares:                &shares5,
			PricePerShare:         &price120,
			CurrencyPricePerShare: &eurCurrency,
			Total:                 &total600,
			CurrencyTotal:         &eurCurrency,
		},
	}

	report := calculator.CalculatePortfolioValuation(transactions)

	// Verify report structure
	if report == nil {
		t.Fatal("Expected portfolio valuation report, got nil")
	}

	if report.Currency != "EUR" {
		t.Errorf("Expected currency EUR, got %s", report.Currency)
	}

	if report.DataSource != "Trading 212 CSV Export" {
		t.Errorf("Expected correct data source")
	}

	// Should have 1 year of data
	if len(report.YearlyPortfolios) != 1 {
		t.Errorf("Expected 1 yearly portfolio, got %d", len(report.YearlyPortfolios))
	}

	portfolio := report.YearlyPortfolios[0]
	if portfolio.Year != 2021 {
		t.Errorf("Expected year 2021, got %d", portfolio.Year)
	}

	// Should have 1 position
	if portfolio.TotalPositions != 1 {
		t.Errorf("Expected 1 position, got %d", portfolio.TotalPositions)
	}

	if len(portfolio.Positions) > 0 {
		position := portfolio.Positions[0]
		// Should use the latest transaction price (120.0)
		if position.LastPrice != 120.0 {
			t.Errorf("Expected last price 120.0, got %.2f", position.LastPrice)
		}

		// Should have 15 shares total
		if position.Shares != 15.0 {
			t.Errorf("Expected 15 shares, got %.2f", position.Shares)
		}

		// Market value should be calculated
		expectedMarketValue := 15.0 * 120.0 // 1800
		if position.MarketValue != expectedMarketValue {
			t.Errorf("Expected market value %.2f, got %.2f", expectedMarketValue, position.MarketValue)
		}
	}
}

func TestPortfolioCalculator_ExtractYears(t *testing.T) {
	calculator := NewPortfolioCalculator("EUR")

	msftTicker := "MSFT"

	transactions := []types.Transaction{
		{
			Action: types.TransactionTypeMarketBuy,
			Time:   time.Date(2021, 10, 15, 10, 0, 0, 0, time.UTC),
			Ticker: &msftTicker,
		},
		{
			Action: types.TransactionTypeMarketBuy,
			Time:   time.Date(2022, 6, 15, 10, 0, 0, 0, time.UTC),
			Ticker: &msftTicker,
		},
	}

	years := calculator.extractYears(transactions)

	expectedYears := []int{2021, 2022}
	if len(years) != len(expectedYears) {
		t.Errorf("Expected %d years, got %d", len(expectedYears), len(years))
	}

	for i, expectedYear := range expectedYears {
		if years[i] != expectedYear {
			t.Errorf("Expected year %d at index %d, got %d", expectedYear, i, years[i])
		}
	}
}

func TestPortfolioCalculator_EmptyTransactions(t *testing.T) {
	calculator := NewPortfolioCalculator("EUR")

	report := calculator.CalculatePortfolioValuation([]types.Transaction{})

	if report == nil {
		t.Fatal("Expected portfolio valuation report, got nil")
	}

	if len(report.YearlyPortfolios) != 0 {
		t.Errorf("Expected no yearly portfolios for empty transactions, got %d", len(report.YearlyPortfolios))
	}
}
