package calculator

import (
	"sort"
	"strings"

	"t212-taxes/internal/domain/types"
)

// IncomeCalculator handles dividend and interest calculations and reporting
type IncomeCalculator struct {
	baseCurrency string
}

// NewIncomeCalculator creates a new income calculator
func NewIncomeCalculator(baseCurrency string) *IncomeCalculator {
	return &IncomeCalculator{
		baseCurrency: baseCurrency,
	}
}

// CalculateIncomeReport generates comprehensive income report from transactions
func (ic *IncomeCalculator) CalculateIncomeReport(transactions []types.Transaction) (*types.IncomeReport, error) {
	if len(transactions) == 0 {
		return &types.IncomeReport{
			Dividends: types.DividendSummary{Currency: ic.baseCurrency},
			Interest:  types.InterestSummary{Currency: ic.baseCurrency},
			Currency:  ic.baseCurrency,
		}, nil
	}

	// Extract dividend and interest transactions
	dividendRecords := ic.extractDividendRecords(transactions)
	interestRecords := ic.extractInterestRecords(transactions)

	// Calculate summaries
	dividendSummary := ic.calculateDividendSummary(dividendRecords)
	interestSummary := ic.calculateInterestSummary(interestRecords)

	// Calculate date range
	dateRange := ic.calculateDateRange(transactions)

	// Calculate total income
	totalIncome := dividendSummary.NetDividends + interestSummary.TotalInterest

	return &types.IncomeReport{
		Dividends:   dividendSummary,
		Interest:    interestSummary,
		TotalIncome: totalIncome,
		Currency:    ic.baseCurrency,
		DateRange:   dateRange,
	}, nil
}

// extractDividendRecords extracts and processes dividend transactions
func (ic *IncomeCalculator) extractDividendRecords(transactions []types.Transaction) []types.DividendRecord {
	var records []types.DividendRecord

	for _, tx := range transactions {
		// Check if transaction is a dividend (handle different formats)
		actionStr := string(tx.Action)
		if !strings.Contains(strings.ToLower(actionStr), "dividend") {
			continue
		}

		// Extract basic dividend information
		record := types.DividendRecord{
			Date:   tx.Time,
			Amount: 0,
		}

		// Get dividend amount
		if tx.Result != nil {
			record.Amount = *tx.Result
		} else if tx.Total != nil {
			record.Amount = *tx.Total
		}

		// Get currency
		if tx.CurrencyResult != nil {
			record.Currency = *tx.CurrencyResult
		} else if tx.CurrencyTotal != nil {
			record.Currency = *tx.CurrencyTotal
		} else {
			record.Currency = ic.baseCurrency
		}

		// Get exchange rate
		if tx.ExchangeRate != nil {
			record.ExchangeRate = *tx.ExchangeRate
		}

		// Get withholding tax
		if tx.WithholdingTax != nil {
			record.WithholdingTax = *tx.WithholdingTax
		}

		// Get security information
		if tx.Ticker != nil {
			record.Ticker = *tx.Ticker
		}
		if tx.ISIN != nil {
			record.ISIN = *tx.ISIN
		}
		if tx.Name != nil {
			record.Name = *tx.Name
		}

		// Calculate net amount
		record.NetAmount = record.Amount - record.WithholdingTax

		// Convert to base currency if needed
		if record.Currency != ic.baseCurrency && record.ExchangeRate > 0 {
			record.Amount = record.Amount / record.ExchangeRate
			record.WithholdingTax = record.WithholdingTax / record.ExchangeRate
			record.NetAmount = record.NetAmount / record.ExchangeRate
			record.Currency = ic.baseCurrency
		}

		records = append(records, record)
	}

	return records
}

// extractInterestRecords extracts and processes interest transactions
func (ic *IncomeCalculator) extractInterestRecords(transactions []types.Transaction) []types.InterestRecord {
	var records []types.InterestRecord

	for _, tx := range transactions {
		// Check if transaction is an interest (handle different formats)
		actionStr := string(tx.Action)
		if !strings.Contains(strings.ToLower(actionStr), "interest") {
			continue
		}

		// Extract basic interest information
		record := types.InterestRecord{
			Date:   tx.Time,
			Amount: 0,
		}

		// Get interest amount
		if tx.Result != nil {
			record.Amount = *tx.Result
		} else if tx.Total != nil {
			record.Amount = *tx.Total
		}

		// Get currency
		if tx.CurrencyResult != nil {
			record.Currency = *tx.CurrencyResult
		} else if tx.CurrencyTotal != nil {
			record.Currency = *tx.CurrencyTotal
		} else {
			record.Currency = ic.baseCurrency
		}

		// Get exchange rate
		if tx.ExchangeRate != nil {
			record.ExchangeRate = *tx.ExchangeRate
		}

		// Get notes for additional information
		if tx.Notes != nil {
			record.Notes = *tx.Notes
			// Try to extract source and period from notes
			ic.extractInterestDetails(&record, *tx.Notes)
		}

		// Convert to base currency if needed
		if record.Currency != ic.baseCurrency && record.ExchangeRate > 0 {
			record.Amount = record.Amount / record.ExchangeRate
			record.Currency = ic.baseCurrency
		}

		records = append(records, record)
	}

	return records
}

// extractInterestDetails tries to extract additional interest information from notes
func (ic *IncomeCalculator) extractInterestDetails(record *types.InterestRecord, notes string) {
	notes = strings.ToLower(notes)

	// Try to identify source
	if strings.Contains(notes, "cash") {
		record.Source = "Cash"
	} else if strings.Contains(notes, "margin") {
		record.Source = "Margin"
	} else if strings.Contains(notes, "account") {
		record.Source = "Account"
	} else {
		record.Source = "Unknown"
	}

	// Try to identify period
	if strings.Contains(notes, "daily") {
		record.Period = "Daily"
	} else if strings.Contains(notes, "weekly") {
		record.Period = "Weekly"
	} else if strings.Contains(notes, "monthly") {
		record.Period = "Monthly"
	} else if strings.Contains(notes, "quarterly") {
		record.Period = "Quarterly"
	} else if strings.Contains(notes, "annual") || strings.Contains(notes, "yearly") {
		record.Period = "Annual"
	} else {
		record.Period = "Unknown"
	}
}

// calculateDividendSummary calculates comprehensive dividend statistics
func (ic *IncomeCalculator) calculateDividendSummary(records []types.DividendRecord) types.DividendSummary {
	summary := types.DividendSummary{
		Currency:   ic.baseCurrency,
		BySecurity: make(map[string]float64),
		ByYear:     make(map[int]float64),
		ByMonth:    make(map[string]float64),
	}

	if len(records) == 0 {
		return summary
	}

	var totalYield float64
	securityCount := 0

	for _, record := range records {
		// Sum totals
		summary.TotalDividends += record.Amount
		summary.TotalWithholdingTax += record.WithholdingTax
		summary.NetDividends += record.NetAmount
		summary.DividendCount++

		// Group by security
		securityKey := record.Ticker
		if securityKey == "" {
			securityKey = record.ISIN
		}
		if securityKey == "" {
			securityKey = "Unknown"
		}
		summary.BySecurity[securityKey] += record.NetAmount

		// Group by year
		year := record.Date.Year()
		summary.ByYear[year] += record.NetAmount

		// Group by month (YYYY-MM format)
		monthKey := record.Date.Format("2006-01")
		summary.ByMonth[monthKey] += record.NetAmount

		// Calculate yield if we have the data
		if record.DividendYield > 0 {
			totalYield += record.DividendYield
			securityCount++
		}
	}

	// Calculate average yield
	if securityCount > 0 {
		summary.AverageYield = totalYield / float64(securityCount)
	}

	return summary
}

// calculateInterestSummary calculates comprehensive interest statistics
func (ic *IncomeCalculator) calculateInterestSummary(records []types.InterestRecord) types.InterestSummary {
	summary := types.InterestSummary{
		Currency: ic.baseCurrency,
		BySource: make(map[string]float64),
		ByYear:   make(map[int]float64),
		ByMonth:  make(map[string]float64),
	}

	if len(records) == 0 {
		return summary
	}

	var totalRate float64
	rateCount := 0

	for _, record := range records {
		// Sum totals
		summary.TotalInterest += record.Amount
		summary.InterestCount++

		// Group by source
		source := record.Source
		if source == "" {
			source = "Unknown"
		}
		summary.BySource[source] += record.Amount

		// Group by year
		year := record.Date.Year()
		summary.ByYear[year] += record.Amount

		// Group by month (YYYY-MM format)
		monthKey := record.Date.Format("2006-01")
		summary.ByMonth[monthKey] += record.Amount

		// Calculate average rate if we have the data
		if record.InterestRate > 0 {
			totalRate += record.InterestRate
			rateCount++
		}
	}

	// Calculate average rate
	if rateCount > 0 {
		summary.AverageRate = totalRate / float64(rateCount)
	}

	return summary
}

// calculateDateRange calculates the date range from transactions
func (ic *IncomeCalculator) calculateDateRange(transactions []types.Transaction) types.DateRange {
	if len(transactions) == 0 {
		return types.DateRange{}
	}

	minTime := transactions[0].Time
	maxTime := transactions[0].Time

	for _, tx := range transactions {
		if tx.Time.Before(minTime) {
			minTime = tx.Time
		}
		if tx.Time.After(maxTime) {
			maxTime = tx.Time
		}
	}

	return types.DateRange{
		From: minTime,
		To:   maxTime,
	}
}

// GetTopDividendPayers returns the top dividend-paying securities
func (ic *IncomeCalculator) GetTopDividendPayers(records []types.DividendRecord, limit int) []DividendPayer {
	securityMap := make(map[string]float64)

	for _, record := range records {
		securityKey := record.Ticker
		if securityKey == "" {
			securityKey = record.ISIN
		}
		if securityKey == "" {
			securityKey = "Unknown"
		}
		securityMap[securityKey] += record.NetAmount
	}

	// Convert to slice for sorting
	var payers []DividendPayer
	for security, amount := range securityMap {
		payers = append(payers, DividendPayer{
			Security: security,
			Amount:   amount,
		})
	}

	// Sort by amount (descending)
	sort.Slice(payers, func(i, j int) bool {
		return payers[i].Amount > payers[j].Amount
	})

	// Return top N
	if limit > 0 && limit < len(payers) {
		return payers[:limit]
	}
	return payers
}

// GetMonthlyIncomeBreakdown returns monthly income breakdown
func (ic *IncomeCalculator) GetMonthlyIncomeBreakdown(dividendRecords []types.DividendRecord, interestRecords []types.InterestRecord) map[string]MonthlyIncome {
	monthlyData := make(map[string]MonthlyIncome)

	// Process dividends
	for _, record := range dividendRecords {
		monthKey := record.Date.Format("2006-01")
		data := monthlyData[monthKey]
		data.Month = monthKey
		data.Dividends += record.NetAmount
		data.TotalIncome += record.NetAmount
		monthlyData[monthKey] = data
	}

	// Process interest
	for _, record := range interestRecords {
		monthKey := record.Date.Format("2006-01")
		data := monthlyData[monthKey]
		data.Month = monthKey
		data.Interest += record.Amount
		data.TotalIncome += record.Amount
		monthlyData[monthKey] = data
	}

	return monthlyData
}

// DividendPayer represents a dividend-paying security with total amount
type DividendPayer struct {
	Security string  `json:"security"`
	Amount   float64 `json:"amount"`
}

// MonthlyIncome represents monthly income breakdown
type MonthlyIncome struct {
	Month       string  `json:"month"`
	Dividends   float64 `json:"dividends"`
	Interest    float64 `json:"interest"`
	TotalIncome float64 `json:"total_income"`
}

// CalculateDividendYield calculates dividend yield for a security
func (ic *IncomeCalculator) CalculateDividendYield(dividendAmount, sharePrice, shares float64) float64 {
	if sharePrice <= 0 || shares <= 0 {
		return 0
	}

	totalValue := sharePrice * shares
	if totalValue <= 0 {
		return 0
	}

	return (dividendAmount / totalValue) * 100
}

// CalculateEffectiveInterestRate calculates effective interest rate
func (ic *IncomeCalculator) CalculateEffectiveInterestRate(interestAmount, principal float64, days int) float64 {
	if principal <= 0 || days <= 0 {
		return 0
	}

	// Calculate annualized rate
	annualizedInterest := (interestAmount / float64(days)) * 365
	return (annualizedInterest / principal) * 100
}
