package calculator

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"t212-taxes/internal/domain/types"
)

// FinancialCalculator handles financial calculations and reporting
type FinancialCalculator struct {
	baseCurrency string
}

// NewFinancialCalculator creates a new financial calculator
func NewFinancialCalculator(baseCurrency string) *FinancialCalculator {
	return &FinancialCalculator{
		baseCurrency: baseCurrency,
	}
}

// CalculateYearlyReports generates yearly financial reports from transactions
func (fc *FinancialCalculator) CalculateYearlyReports(transactions []types.Transaction) ([]types.YearlyReport, error) {
	if len(transactions) == 0 {
		return []types.YearlyReport{}, nil
	}

	// Group transactions by year
	yearlyTransactions := fc.groupTransactionsByYear(transactions)
	
	var reports []types.YearlyReport
	for year, yearTransactions := range yearlyTransactions {
		report, err := fc.calculateYearlyReport(year, yearTransactions)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate report for year %d: %w", year, err)
		}
		reports = append(reports, *report)
	}

	// Sort reports by year
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Year < reports[j].Year
	})

	return reports, nil
}

// CalculateOverallReport generates an overall investment summary
func (fc *FinancialCalculator) CalculateOverallReport(yearlyReports []types.YearlyReport) *types.OverallReport {
	if len(yearlyReports) == 0 {
		return &types.OverallReport{
			Currency: fc.baseCurrency,
		}
	}

	overall := &types.OverallReport{
		YearlyReports: yearlyReports,
		Currency:      fc.baseCurrency,
	}

	// Extract years
	years := make([]int, len(yearlyReports))
	for i, report := range yearlyReports {
		years[i] = report.Year
	}
	overall.Years = years

	// Sum up totals
	for _, report := range yearlyReports {
		overall.TotalDeposits += report.TotalDeposits
		overall.TotalTransactions += report.TotalTransactions
		overall.TotalCapitalGains += report.CapitalGains
		overall.TotalDividends += report.Dividends
		overall.TotalInterest += report.Interest
		overall.TotalGains += report.TotalGains
	}

	// Calculate overall percentage
	if overall.TotalDeposits > 0 {
		overall.OverallPercentage = (overall.TotalGains / overall.TotalDeposits) * 100
	}

	return overall
}

// groupTransactionsByYear groups transactions by their year
func (fc *FinancialCalculator) groupTransactionsByYear(transactions []types.Transaction) map[int][]types.Transaction {
	yearlyTransactions := make(map[int][]types.Transaction)
	
	for _, transaction := range transactions {
		year := transaction.Time.Year()
		yearlyTransactions[year] = append(yearlyTransactions[year], transaction)
	}
	
	return yearlyTransactions
}

// calculateYearlyReport calculates financial metrics for a specific year
func (fc *FinancialCalculator) calculateYearlyReport(year int, transactions []types.Transaction) (*types.YearlyReport, error) {
	report := &types.YearlyReport{
		Year:              year,
		TotalTransactions: len(transactions),
		Currency:          fc.baseCurrency,
	}

	for _, transaction := range transactions {
		switch transaction.Action {
		case types.TransactionTypeDeposit:
			// Add deposits to total
			if transaction.Total != nil {
				amount := fc.convertToBaseCurrency(*transaction.Total, transaction.CurrencyTotal, transaction.ExchangeRate)
				report.TotalDeposits += amount
			}

		default:
			// Check for dividend transactions (handle different formats)
			actionStr := string(transaction.Action)
			if strings.Contains(strings.ToLower(actionStr), "dividend") {
				// Add dividends to total
				if transaction.Result != nil {
					amount := fc.convertToBaseCurrency(*transaction.Result, transaction.CurrencyResult, transaction.ExchangeRate)
					report.Dividends += amount
				} else if transaction.Total != nil {
					amount := fc.convertToBaseCurrency(*transaction.Total, transaction.CurrencyTotal, transaction.ExchangeRate)
					report.Dividends += amount
				}
			} else if strings.Contains(strings.ToLower(actionStr), "interest") {
				// Add interest to total
				if transaction.Result != nil {
					amount := fc.convertToBaseCurrency(*transaction.Result, transaction.CurrencyResult, transaction.ExchangeRate)
					report.Interest += amount
				} else if transaction.Total != nil {
					amount := fc.convertToBaseCurrency(*transaction.Total, transaction.CurrencyTotal, transaction.ExchangeRate)
					report.Interest += amount
				}
			}

		case types.TransactionTypeMarketSell, types.TransactionTypeLimitSell, types.TransactionTypeStopSell:
			// For sells, we need to calculate capital gains
			// This is a simplified approach - in reality, we'd need to track purchase prices
			if transaction.Result != nil {
				amount := fc.convertToBaseCurrency(*transaction.Result, transaction.CurrencyResult, transaction.ExchangeRate)
				if amount > 0 {
					report.CapitalGains += amount
				}
			}
		}
	}

	// Calculate total gains
	report.TotalGains = report.CapitalGains + report.Dividends + report.Interest

	// Calculate percentage increase
	if report.TotalDeposits > 0 {
		report.PercentageIncrease = (report.TotalGains / report.TotalDeposits) * 100
	}

	return report, nil
}

// convertToBaseCurrency converts an amount to the base currency
func (fc *FinancialCalculator) convertToBaseCurrency(amount float64, currency *string, exchangeRate *float64) float64 {
	if currency == nil || *currency == fc.baseCurrency {
		return amount
	}
	
	if exchangeRate == nil || *exchangeRate == 0 {
		// If no exchange rate provided, assume 1:1 (this should be handled better in production)
		return amount
	}
	
	// Convert using exchange rate
	// Note: Exchange rate semantics may vary, this is a simplified approach
	return amount / *exchangeRate
}

// CalculateCapitalGains calculates capital gains using FIFO method
func (fc *FinancialCalculator) CalculateCapitalGains(transactions []types.Transaction) (float64, float64, error) {
	// Group transactions by security
	securityTransactions := make(map[string][]types.Transaction)
	
	for _, transaction := range transactions {
		if transaction.Ticker == nil {
			continue
		}
		
		ticker := *transaction.Ticker
		if fc.isTradeTransaction(transaction.Action) {
			securityTransactions[ticker] = append(securityTransactions[ticker], transaction)
		}
	}
	
	totalGains := 0.0
	totalLosses := 0.0
	
	// Calculate gains/losses for each security using FIFO
	for ticker, secTrans := range securityTransactions {
		gains, losses, err := fc.calculateSecurityGainsLosses(ticker, secTrans)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to calculate gains/losses for %s: %w", ticker, err)
		}
		totalGains += gains
		totalLosses += losses
	}
	
	return totalGains, totalLosses, nil
}

// isTradeTransaction checks if the transaction is a trade (buy/sell)
func (fc *FinancialCalculator) isTradeTransaction(action types.TransactionType) bool {
	switch action {
	case types.TransactionTypeMarketBuy, types.TransactionTypeMarketSell,
		 types.TransactionTypeLimitBuy, types.TransactionTypeLimitSell,
		 types.TransactionTypeStopBuy, types.TransactionTypeStopSell:
		return true
	default:
		return false
	}
}

// isBuyTransaction checks if the transaction is a buy order
func (fc *FinancialCalculator) isBuyTransaction(action types.TransactionType) bool {
	switch action {
	case types.TransactionTypeMarketBuy, types.TransactionTypeLimitBuy, types.TransactionTypeStopBuy:
		return true
	default:
		return false
	}
}

// calculateSecurityGainsLosses calculates gains/losses for a specific security using FIFO
func (fc *FinancialCalculator) calculateSecurityGainsLosses(ticker string, transactions []types.Transaction) (float64, float64, error) {
	// Sort transactions by time
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Time.Before(transactions[j].Time)
	})
	
	// FIFO queue for purchases
	var purchases []PurchaseRecord
	totalGains := 0.0
	totalLosses := 0.0
	
	for _, transaction := range transactions {
		if fc.isBuyTransaction(transaction.Action) {
			// Add to purchases queue
			if transaction.Shares != nil && transaction.PricePerShare != nil {
				shares := *transaction.Shares
				pricePerShare := *transaction.PricePerShare
				
				// Convert to base currency
				convertedPrice := fc.convertToBaseCurrency(pricePerShare, transaction.CurrencyPricePerShare, transaction.ExchangeRate)
				
				purchases = append(purchases, PurchaseRecord{
					Date:          transaction.Time,
					Shares:        shares,
					PricePerShare: convertedPrice,
					TotalCost:     shares * convertedPrice,
				})
			}
		} else {
			// Sell transaction - calculate gains/losses using FIFO
			if transaction.Shares != nil && transaction.PricePerShare != nil {
				sellShares := *transaction.Shares
				sellPrice := *transaction.PricePerShare
				
				// Convert to base currency
				convertedSellPrice := fc.convertToBaseCurrency(sellPrice, transaction.CurrencyPricePerShare, transaction.ExchangeRate)
				
				remainingShares := sellShares
				
				// Process FIFO
				for i := 0; i < len(purchases) && remainingShares > 0; i++ {
					purchase := &purchases[i]
					if purchase.Shares <= 0 {
						continue
					}
					
					sharesToProcess := math.Min(remainingShares, purchase.Shares)
					
					// Calculate gain/loss
					costBasis := sharesToProcess * purchase.PricePerShare
					saleProceeds := sharesToProcess * convertedSellPrice
					gainLoss := saleProceeds - costBasis
					
					if gainLoss > 0 {
						totalGains += gainLoss
					} else {
						totalLosses += math.Abs(gainLoss)
					}
					
					// Update remaining shares
					purchase.Shares -= sharesToProcess
					remainingShares -= sharesToProcess
				}
			}
		}
	}
	
	return totalGains, totalLosses, nil
}

// PurchaseRecord represents a purchase for FIFO calculation
type PurchaseRecord struct {
	Date          time.Time
	Shares        float64
	PricePerShare float64
	TotalCost     float64
}