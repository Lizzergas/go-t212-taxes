package calculator

import (
	"testing"
	"time"

	"t212-taxes/internal/domain/types"
)

// Helper functions

func TestIncomeCalculator_CalculateIncomeReport(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	transactions := []types.Transaction{
		// Dividend transactions
		{
			Action:         types.TransactionTypeDividend,
			Time:           time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         stringPtr("AAPL"),
			ISIN:           stringPtr("US0378331005"),
			Name:           stringPtr("Apple Inc."),
			Result:         floatPtr(25.0),
			CurrencyResult: stringPtr("EUR"),
			WithholdingTax: floatPtr(3.75),
		},
		{
			Action:         types.TransactionTypeDividend,
			Time:           time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         stringPtr("MSFT"),
			ISIN:           stringPtr("US5949181045"),
			Name:           stringPtr("Microsoft Corporation"),
			Result:         floatPtr(30.0),
			CurrencyResult: stringPtr("EUR"),
			WithholdingTax: floatPtr(4.5),
		},
		// Interest transactions
		{
			Action:         types.TransactionTypeInterest,
			Time:           time.Date(2024, 1, 31, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(10.0),
			CurrencyResult: stringPtr("EUR"),
			Notes:          stringPtr("Monthly interest on cash balance"),
		},
		{
			Action:         types.TransactionTypeInterest,
			Time:           time.Date(2024, 2, 29, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(12.0),
			CurrencyResult: stringPtr("EUR"),
			Notes:          stringPtr("Monthly interest on margin account"),
		},
	}

	report, err := calc.CalculateIncomeReport(transactions)
	if err != nil {
		t.Fatalf("CalculateIncomeReport() error = %v", err)
	}

	// Test dividend summary
	if report.Dividends.TotalDividends != 55.0 {
		t.Errorf("Dividends.TotalDividends = %f, want 55.0", report.Dividends.TotalDividends)
	}

	if report.Dividends.TotalWithholdingTax != 8.25 {
		t.Errorf("Dividends.TotalWithholdingTax = %f, want 8.25", report.Dividends.TotalWithholdingTax)
	}

	if report.Dividends.NetDividends != 46.75 {
		t.Errorf("Dividends.NetDividends = %f, want 46.75", report.Dividends.NetDividends)
	}

	if report.Dividends.DividendCount != 2 {
		t.Errorf("Dividends.DividendCount = %d, want 2", report.Dividends.DividendCount)
	}

	// Test interest summary
	if report.Interest.TotalInterest != 22.0 {
		t.Errorf("Interest.TotalInterest = %f, want 22.0", report.Interest.TotalInterest)
	}

	if report.Interest.InterestCount != 2 {
		t.Errorf("Interest.InterestCount = %d, want 2", report.Interest.InterestCount)
	}

	// Test total income
	expectedTotalIncome := 46.75 + 22.0 // Net dividends + total interest
	if report.TotalIncome != expectedTotalIncome {
		t.Errorf("TotalIncome = %f, want %f", report.TotalIncome, expectedTotalIncome)
	}

	// Test currency
	if report.Currency != "EUR" {
		t.Errorf("Currency = %s, want EUR", report.Currency)
	}
}

func TestIncomeCalculator_CalculateIncomeReport_EmptyTransactions(t *testing.T) {
	calc := NewIncomeCalculator("EUR")
	transactions := []types.Transaction{}

	report, err := calc.CalculateIncomeReport(transactions)
	if err != nil {
		t.Fatalf("CalculateIncomeReport() error = %v", err)
	}

	if report.TotalIncome != 0 {
		t.Errorf("TotalIncome = %f, want 0", report.TotalIncome)
	}

	if report.Dividends.TotalDividends != 0 {
		t.Errorf("Dividends.TotalDividends = %f, want 0", report.Dividends.TotalDividends)
	}

	if report.Interest.TotalInterest != 0 {
		t.Errorf("Interest.TotalInterest = %f, want 0", report.Interest.TotalInterest)
	}

	if report.Currency != "EUR" {
		t.Errorf("Currency = %s, want EUR", report.Currency)
	}
}

func TestIncomeCalculator_ExtractDividendRecords(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	transactions := []types.Transaction{
		{
			Action:         types.TransactionTypeDividend,
			Time:           time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         stringPtr("AAPL"),
			ISIN:           stringPtr("US0378331005"),
			Name:           stringPtr("Apple Inc."),
			Result:         floatPtr(25.0),
			CurrencyResult: stringPtr("EUR"),
			WithholdingTax: floatPtr(3.75),
		},
		{
			Action:                types.TransactionTypeMarketBuy, // Non-dividend transaction
			Time:                  time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC),
			Ticker:                stringPtr("AAPL"),
			Shares:                floatPtr(10),
			PricePerShare:         floatPtr(150.0),
			CurrencyPricePerShare: stringPtr("USD"),
		},
		{
			Action:         types.TransactionTypeDividend,
			Time:           time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         stringPtr("MSFT"),
			Result:         floatPtr(30.0),
			CurrencyResult: stringPtr("EUR"),
			WithholdingTax: floatPtr(4.5),
		},
	}

	records := calc.extractDividendRecords(transactions)

	if len(records) != 2 {
		t.Errorf("Expected 2 dividend records, got %d", len(records))
	}

	// Check first record
	if records[0].Ticker != "AAPL" {
		t.Errorf("First record ticker = %s, want AAPL", records[0].Ticker)
	}

	if records[0].Amount != 25.0 {
		t.Errorf("First record amount = %f, want 25.0", records[0].Amount)
	}

	if records[0].WithholdingTax != 3.75 {
		t.Errorf("First record withholding tax = %f, want 3.75", records[0].WithholdingTax)
	}

	if records[0].NetAmount != 21.25 {
		t.Errorf("First record net amount = %f, want 21.25", records[0].NetAmount)
	}

	// Check second record
	if records[1].Ticker != "MSFT" {
		t.Errorf("Second record ticker = %s, want MSFT", records[1].Ticker)
	}

	if records[1].Amount != 30.0 {
		t.Errorf("Second record amount = %f, want 30.0", records[1].Amount)
	}
}

func TestIncomeCalculator_ExtractInterestRecords(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	transactions := []types.Transaction{
		{
			Action:         types.TransactionTypeInterest,
			Time:           time.Date(2024, 1, 31, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(10.0),
			CurrencyResult: stringPtr("EUR"),
			Notes:          stringPtr("Monthly interest on cash balance"),
		},
		{
			Action:        types.TransactionTypeMarketBuy, // Non-interest transaction
			Time:          time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
			Ticker:        stringPtr("AAPL"),
			Shares:        floatPtr(10),
			PricePerShare: floatPtr(150.0),
		},
		{
			Action:         types.TransactionTypeInterest,
			Time:           time.Date(2024, 2, 29, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(12.0),
			CurrencyResult: stringPtr("EUR"),
			Notes:          stringPtr("Monthly interest on margin account"),
		},
	}

	records := calc.extractInterestRecords(transactions)

	if len(records) != 2 {
		t.Errorf("Expected 2 interest records, got %d", len(records))
	}

	// Check first record
	if records[0].Amount != 10.0 {
		t.Errorf("First record amount = %f, want 10.0", records[0].Amount)
	}

	if records[0].Source != "Cash" {
		t.Errorf("First record source = %s, want Cash", records[0].Source)
	}

	if records[0].Period != "Monthly" {
		t.Errorf("First record period = %s, want Monthly", records[0].Period)
	}

	// Check second record
	if records[1].Amount != 12.0 {
		t.Errorf("Second record amount = %f, want 12.0", records[1].Amount)
	}

	if records[1].Source != "Margin" {
		t.Errorf("Second record source = %s, want Margin", records[1].Source)
	}
}

func TestIncomeCalculator_ExtractInterestDetails(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	tests := []struct {
		name     string
		notes    string
		expected struct {
			source string
			period string
		}
	}{
		{
			name:  "Cash monthly interest",
			notes: "Monthly interest on cash balance",
			expected: struct {
				source string
				period string
			}{
				source: "Cash",
				period: "Monthly",
			},
		},
		{
			name:  "Margin quarterly interest",
			notes: "Quarterly interest on margin account",
			expected: struct {
				source string
				period string
			}{
				source: "Margin",
				period: "Quarterly",
			},
		},
		{
			name:  "Account annual interest",
			notes: "Annual interest on account balance",
			expected: struct {
				source string
				period string
			}{
				source: "Account",
				period: "Annual",
			},
		},
		{
			name:  "Unknown source and period",
			notes: "Some other interest payment",
			expected: struct {
				source string
				period string
			}{
				source: "Unknown",
				period: "Unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &types.InterestRecord{}
			calc.extractInterestDetails(record, tt.notes)

			if record.Source != tt.expected.source {
				t.Errorf("Source = %s, want %s", record.Source, tt.expected.source)
			}

			if record.Period != tt.expected.period {
				t.Errorf("Period = %s, want %s", record.Period, tt.expected.period)
			}
		})
	}
}

func TestIncomeCalculator_CalculateDividendSummary(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	records := []types.DividendRecord{
		{
			Date:           time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         "AAPL",
			Amount:         25.0,
			WithholdingTax: 3.75,
			NetAmount:      21.25,
		},
		{
			Date:           time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         "MSFT",
			Amount:         30.0,
			WithholdingTax: 4.5,
			NetAmount:      25.5,
		},
		{
			Date:           time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         "AAPL",
			Amount:         20.0,
			WithholdingTax: 3.0,
			NetAmount:      17.0,
		},
	}

	summary := calc.calculateDividendSummary(records)

	// Test totals
	if summary.TotalDividends != 75.0 {
		t.Errorf("TotalDividends = %f, want 75.0", summary.TotalDividends)
	}

	if summary.TotalWithholdingTax != 11.25 {
		t.Errorf("TotalWithholdingTax = %f, want 11.25", summary.TotalWithholdingTax)
	}

	if summary.NetDividends != 63.75 {
		t.Errorf("NetDividends = %f, want 63.75", summary.NetDividends)
	}

	if summary.DividendCount != 3 {
		t.Errorf("DividendCount = %d, want 3", summary.DividendCount)
	}

	// Test by security
	if summary.BySecurity["AAPL"] != 38.25 {
		t.Errorf("BySecurity[AAPL] = %f, want 38.25", summary.BySecurity["AAPL"])
	}

	if summary.BySecurity["MSFT"] != 25.5 {
		t.Errorf("BySecurity[MSFT] = %f, want 25.5", summary.BySecurity["MSFT"])
	}

	// Test by year
	if summary.ByYear[2024] != 63.75 {
		t.Errorf("ByYear[2024] = %f, want 63.75", summary.ByYear[2024])
	}

	// Test by month
	if summary.ByMonth["2024-01"] != 21.25 {
		t.Errorf("ByMonth[2024-01] = %f, want 21.25", summary.ByMonth["2024-01"])
	}

	if summary.ByMonth["2024-02"] != 25.5 {
		t.Errorf("ByMonth[2024-02] = %f, want 25.5", summary.ByMonth["2024-02"])
	}

	if summary.ByMonth["2024-03"] != 17.0 {
		t.Errorf("ByMonth[2024-03] = %f, want 17.0", summary.ByMonth["2024-03"])
	}
}

func TestIncomeCalculator_CalculateInterestSummary(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	records := []types.InterestRecord{
		{
			Date:   time.Date(2024, 1, 31, 9, 0, 0, 0, time.UTC),
			Amount: 10.0,
			Source: "Cash",
			Period: "Monthly",
		},
		{
			Date:   time.Date(2024, 2, 29, 9, 0, 0, 0, time.UTC),
			Amount: 12.0,
			Source: "Margin",
			Period: "Monthly",
		},
		{
			Date:   time.Date(2024, 3, 31, 9, 0, 0, 0, time.UTC),
			Amount: 8.0,
			Source: "Cash",
			Period: "Monthly",
		},
	}

	summary := calc.calculateInterestSummary(records)

	// Test totals
	if summary.TotalInterest != 30.0 {
		t.Errorf("TotalInterest = %f, want 30.0", summary.TotalInterest)
	}

	if summary.InterestCount != 3 {
		t.Errorf("InterestCount = %d, want 3", summary.InterestCount)
	}

	// Test by source
	if summary.BySource["Cash"] != 18.0 {
		t.Errorf("BySource[Cash] = %f, want 18.0", summary.BySource["Cash"])
	}

	if summary.BySource["Margin"] != 12.0 {
		t.Errorf("BySource[Margin] = %f, want 12.0", summary.BySource["Margin"])
	}

	// Test by year
	if summary.ByYear[2024] != 30.0 {
		t.Errorf("ByYear[2024] = %f, want 30.0", summary.ByYear[2024])
	}

	// Test by month
	if summary.ByMonth["2024-01"] != 10.0 {
		t.Errorf("ByMonth[2024-01] = %f, want 10.0", summary.ByMonth["2024-01"])
	}

	if summary.ByMonth["2024-02"] != 12.0 {
		t.Errorf("ByMonth[2024-02] = %f, want 12.0", summary.ByMonth["2024-02"])
	}

	if summary.ByMonth["2024-03"] != 8.0 {
		t.Errorf("ByMonth[2024-03] = %f, want 8.0", summary.ByMonth["2024-03"])
	}
}

func TestIncomeCalculator_GetTopDividendPayers(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	records := []types.DividendRecord{
		{
			Ticker:    "AAPL",
			NetAmount: 25.0,
		},
		{
			Ticker:    "MSFT",
			NetAmount: 30.0,
		},
		{
			Ticker:    "AAPL",
			NetAmount: 15.0,
		},
		{
			Ticker:    "GOOGL",
			NetAmount: 20.0,
		},
	}

	payers := calc.GetTopDividendPayers(records, 2)

	if len(payers) != 2 {
		t.Errorf("Expected 2 payers, got %d", len(payers))
	}

	// Check order (should be sorted by amount descending)
	if payers[0].Security != "AAPL" || payers[0].Amount != 40.0 {
		t.Errorf("First payer = %s (%.2f), want AAPL (40.00)", payers[0].Security, payers[0].Amount)
	}

	if payers[1].Security != "MSFT" || payers[1].Amount != 30.0 {
		t.Errorf("Second payer = %s (%.2f), want MSFT (30.00)", payers[1].Security, payers[1].Amount)
	}
}

func TestIncomeCalculator_GetMonthlyIncomeBreakdown(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	dividendRecords := []types.DividendRecord{
		{
			Date:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			NetAmount: 25.0,
		},
		{
			Date:      time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC),
			NetAmount: 30.0,
		},
	}

	interestRecords := []types.InterestRecord{
		{
			Date:   time.Date(2024, 1, 31, 9, 0, 0, 0, time.UTC),
			Amount: 10.0,
		},
		{
			Date:   time.Date(2024, 2, 29, 9, 0, 0, 0, time.UTC),
			Amount: 12.0,
		},
	}

	breakdown := calc.GetMonthlyIncomeBreakdown(dividendRecords, interestRecords)

	// Check January
	janData := breakdown["2024-01"]
	if janData.Dividends != 25.0 {
		t.Errorf("January dividends = %f, want 25.0", janData.Dividends)
	}

	if janData.Interest != 10.0 {
		t.Errorf("January interest = %f, want 10.0", janData.Interest)
	}

	if janData.TotalIncome != 35.0 {
		t.Errorf("January total income = %f, want 35.0", janData.TotalIncome)
	}

	// Check February
	febData := breakdown["2024-02"]
	if febData.Dividends != 30.0 {
		t.Errorf("February dividends = %f, want 30.0", febData.Dividends)
	}

	if febData.Interest != 12.0 {
		t.Errorf("February interest = %f, want 12.0", febData.Interest)
	}

	if febData.TotalIncome != 42.0 {
		t.Errorf("February total income = %f, want 42.0", febData.TotalIncome)
	}
}

func TestIncomeCalculator_CalculateDividendYield(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	tests := []struct {
		name           string
		dividendAmount float64
		sharePrice     float64
		shares         float64
		expected       float64
	}{
		{
			name:           "Normal dividend yield",
			dividendAmount: 25.0,
			sharePrice:     100.0,
			shares:         10,
			expected:       2.5, // (25 / (100 * 10)) * 100
		},
		{
			name:           "Zero share price",
			dividendAmount: 25.0,
			sharePrice:     0.0,
			shares:         10,
			expected:       0.0,
		},
		{
			name:           "Zero shares",
			dividendAmount: 25.0,
			sharePrice:     100.0,
			shares:         0,
			expected:       0.0,
		},
		{
			name:           "Zero dividend",
			dividendAmount: 0.0,
			sharePrice:     100.0,
			shares:         10,
			expected:       0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateDividendYield(tt.dividendAmount, tt.sharePrice, tt.shares)
			if result != tt.expected {
				t.Errorf("CalculateDividendYield() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestIncomeCalculator_CalculateEffectiveInterestRate(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	tests := []struct {
		name           string
		interestAmount float64
		principal      float64
		days           int
		expected       float64
	}{
		{
			name:           "Normal interest rate",
			interestAmount: 10.0,
			principal:      1000.0,
			days:           30,
			expected:       12.166666666666666, // (10 / 30) * 365 / 1000 * 100
		},
		{
			name:           "Zero principal",
			interestAmount: 10.0,
			principal:      0.0,
			days:           30,
			expected:       0.0,
		},
		{
			name:           "Zero days",
			interestAmount: 10.0,
			principal:      1000.0,
			days:           0,
			expected:       0.0,
		},
		{
			name:           "Zero interest",
			interestAmount: 0.0,
			principal:      1000.0,
			days:           30,
			expected:       0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateEffectiveInterestRate(tt.interestAmount, tt.principal, tt.days)
			if result != tt.expected {
				t.Errorf("CalculateEffectiveInterestRate() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestIncomeCalculator_CurrencyConversion(t *testing.T) {
	calc := NewIncomeCalculator("EUR")

	transactions := []types.Transaction{
		{
			Action:         types.TransactionTypeDividend,
			Time:           time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			Ticker:         stringPtr("AAPL"),
			Result:         floatPtr(100.0), // USD
			CurrencyResult: stringPtr("USD"),
			ExchangeRate:   floatPtr(0.9),  // 1 USD = 0.9 EUR
			WithholdingTax: floatPtr(15.0), // USD
		},
		{
			Action:         types.TransactionTypeInterest,
			Time:           time.Date(2024, 1, 31, 9, 0, 0, 0, time.UTC),
			Result:         floatPtr(50.0), // USD
			CurrencyResult: stringPtr("USD"),
			ExchangeRate:   floatPtr(0.9), // 1 USD = 0.9 EUR
		},
	}

	report, err := calc.CalculateIncomeReport(transactions)
	if err != nil {
		t.Fatalf("CalculateIncomeReport() error = %v", err)
	}

	// Check dividend conversion (100 USD / 0.9 = 111.11 EUR)
	expectedDividend := 100.0 / 0.9
	if report.Dividends.TotalDividends != expectedDividend {
		t.Errorf("Dividends.TotalDividends = %f, want %f", report.Dividends.TotalDividends, expectedDividend)
	}

	// Check withholding tax conversion (15 USD / 0.9 = 16.67 EUR)
	expectedWithholdingTax := 15.0 / 0.9
	if report.Dividends.TotalWithholdingTax != expectedWithholdingTax {
		t.Errorf("Dividends.TotalWithholdingTax = %f, want %f", report.Dividends.TotalWithholdingTax, expectedWithholdingTax)
	}

	// Check interest conversion (50 USD / 0.9 = 55.56 EUR)
	expectedInterest := 50.0 / 0.9
	if report.Interest.TotalInterest != expectedInterest {
		t.Errorf("Interest.TotalInterest = %f, want %f", report.Interest.TotalInterest, expectedInterest)
	}
}
