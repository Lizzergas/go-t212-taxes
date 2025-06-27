package calculator

import (
	"sort"
	"strings"
	"time"

	"t212-taxes/internal/domain/types"
)

const (
	// MinSharesThreshold is the minimum shares to consider a position significant
	MinSharesThreshold = 0.001
)

// PortfolioCalculator calculates portfolio positions and summaries
type PortfolioCalculator struct {
	baseCurrency string
}

// NewPortfolioCalculator creates a new portfolio calculator
func NewPortfolioCalculator(baseCurrency string) *PortfolioCalculator {
	return &PortfolioCalculator{
		baseCurrency: baseCurrency,
	}
}

// CalculatePortfolioValuation generates portfolio valuations across multiple years
func (pc *PortfolioCalculator) CalculatePortfolioValuation(transactions []types.Transaction) *types.PortfolioValuationReport {
	// Get all unique years from transactions
	years := pc.extractYears(transactions)

	var yearlyPortfolios []types.PortfolioSummary

	for _, year := range years {
		portfolio := pc.CalculateEndOfYearPortfolio(transactions, year)
		yearlyPortfolios = append(yearlyPortfolios, *portfolio)
	}

	return &types.PortfolioValuationReport{
		YearlyPortfolios: yearlyPortfolios,
		Currency:         pc.baseCurrency,
		GeneratedAt:      time.Now(),
		DataSource:       "Trading 212 CSV Export",
		PriceNote:        "Portfolio values based on last transaction price for each security",
	}
}

// extractYears gets all unique years from transactions
func (pc *PortfolioCalculator) extractYears(transactions []types.Transaction) []int {
	yearMap := make(map[int]bool)

	for _, tx := range transactions {
		if pc.isTradeTransaction(tx) || tx.Action == types.TransactionTypeDeposit {
			yearMap[tx.Time.Year()] = true
		}
	}

	var years []int
	for year := range yearMap {
		years = append(years, year)
	}

	// Sort years
	sort.Ints(years)
	return years
}

// CalculateEndOfYearPortfolio calculates the portfolio state at the end of a given year
func (pc *PortfolioCalculator) CalculateEndOfYearPortfolio(transactions []types.Transaction, year int) *types.PortfolioSummary {
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	relevantTransactions := pc.filterTransactionsUpToDate(transactions, endOfYear)

	// Process transactions to build positions and calculate metrics
	positions := make(map[string]*types.PortfolioPosition)
	lastPrices := make(map[string]*PriceInfo)
	yearlyMetrics := pc.calculateYearlyMetrics(relevantTransactions, year)

	pc.processTransactionsForPositions(relevantTransactions, positions, lastPrices)
	finalPositions, totals := pc.buildFinalPositions(positions, lastPrices)

	return &types.PortfolioSummary{
		Year:                           year,
		AsOfDate:                       endOfYear,
		Positions:                      finalPositions,
		TotalPositions:                 len(finalPositions),
		TotalShares:                    totals.TotalShares,
		TotalInvested:                  totals.TotalInvested,
		TotalMarketValue:               totals.TotalMarketValue,
		TotalUnrealizedGainLoss:        totals.TotalMarketValue - totals.TotalInvested,
		TotalUnrealizedGainLossPercent: pc.calculatePercentage(totals.TotalMarketValue-totals.TotalInvested, totals.TotalInvested),
		Currency:                       pc.baseCurrency,
		YearlyDeposits:                 yearlyMetrics.Deposits,
		YearlyDividends:                yearlyMetrics.Dividends,
		YearlyInterest:                 yearlyMetrics.Interest,
	}
}

// filterTransactionsUpToDate filters transactions up to the specified date
func (pc *PortfolioCalculator) filterTransactionsUpToDate(transactions []types.Transaction, endDate time.Time) []types.Transaction {
	var filtered []types.Transaction
	for _, tx := range transactions {
		if tx.Time.Before(endDate) || tx.Time.Equal(endDate) {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

// YearlyMetrics holds yearly metric calculations
type YearlyMetrics struct {
	Deposits  float64
	Dividends float64
	Interest  float64
}

// calculateYearlyMetrics calculates deposits, dividends, and interest for a specific year
func (pc *PortfolioCalculator) calculateYearlyMetrics(transactions []types.Transaction, year int) *YearlyMetrics {
	metrics := &YearlyMetrics{}

	for _, tx := range transactions {
		if tx.Time.Year() != year {
			continue
		}

		switch {
		case tx.Action == types.TransactionTypeDeposit:
			if tx.Total != nil {
				amount := pc.convertToBaseCurrency(*tx.Total, tx.CurrencyTotal, tx.ExchangeRate)
				metrics.Deposits += amount
			}
		case strings.Contains(strings.ToLower(string(tx.Action)), "dividend"):
			convertedAmount := pc.extractTransactionAmount(tx)
			metrics.Dividends += convertedAmount
		case strings.Contains(strings.ToLower(string(tx.Action)), "interest"):
			convertedAmount := pc.extractTransactionAmount(tx)
			metrics.Interest += convertedAmount
		}
	}

	return metrics
}

// processTransactionsForPositions processes trade transactions to build positions and track prices
func (pc *PortfolioCalculator) processTransactionsForPositions(transactions []types.Transaction, positions map[string]*types.PortfolioPosition, lastPrices map[string]*PriceInfo) {
	for _, tx := range transactions {
		if !pc.isTradeTransaction(tx) || tx.Ticker == nil || tx.ISIN == nil {
			continue
		}

		ticker := *tx.Ticker
		position := pc.getOrCreatePosition(positions, ticker, tx)

		// Update position based on transaction type
		if pc.isBuyTransaction(tx) {
			pc.handleBuyTransaction(position, tx)
		} else if pc.isSellTransaction(tx) {
			pc.handleSellTransaction(position, tx)
		}

		// Track last price information
		pc.updateLastPrice(lastPrices, ticker, tx)
	}
}

// getOrCreatePosition gets existing position or creates a new one
func (pc *PortfolioCalculator) getOrCreatePosition(positions map[string]*types.PortfolioPosition, ticker string, tx types.Transaction) *types.PortfolioPosition {
	position, exists := positions[ticker]
	if !exists {
		position = &types.PortfolioPosition{
			Ticker:           ticker,
			ISIN:             *tx.ISIN,
			Name:             pc.getSecurityName(tx),
			Shares:           0,
			TotalCost:        0,
			Currency:         pc.baseCurrency,
			TransactionCount: 0,
		}
		positions[ticker] = position
	}
	return position
}

// updateLastPrice updates the last price information for a ticker
func (pc *PortfolioCalculator) updateLastPrice(lastPrices map[string]*PriceInfo, ticker string, tx types.Transaction) {
	if tx.PricePerShare != nil && *tx.PricePerShare > 0 {
		priceInBaseCurrency := pc.convertToBaseCurrency(*tx.PricePerShare, tx.CurrencyPricePerShare, tx.ExchangeRate)

		lastPrices[ticker] = &PriceInfo{
			Price:            priceInBaseCurrency,
			Date:             tx.Time,
			Currency:         pc.baseCurrency,
			OriginalPrice:    *tx.PricePerShare,
			OriginalCurrency: pc.safeString(tx.CurrencyPricePerShare),
		}
	}
}

// PositionTotals holds aggregated position totals
type PositionTotals struct {
	TotalShares      float64
	TotalInvested    float64
	TotalMarketValue float64
}

// buildFinalPositions converts positions map to sorted slice and calculates totals
func (pc *PortfolioCalculator) buildFinalPositions(positions map[string]*types.PortfolioPosition, lastPrices map[string]*PriceInfo) ([]types.PortfolioPosition, *PositionTotals) {
	var finalPositions []types.PortfolioPosition
	totals := &PositionTotals{}

	for _, position := range positions {
		if position.Shares <= MinSharesThreshold { // Filter out tiny remaining positions
			continue
		}

		pc.finalizePosition(position, lastPrices)

		finalPositions = append(finalPositions, *position)
		totals.TotalShares += position.Shares
		totals.TotalInvested += position.TotalCost
		totals.TotalMarketValue += position.MarketValue
	}

	// Sort by market value in descending order
	sort.Slice(finalPositions, func(i, j int) bool {
		return finalPositions[i].MarketValue > finalPositions[j].MarketValue
	})

	return finalPositions, totals
}

// finalizePosition calculates final position metrics including market value and P&L
func (pc *PortfolioCalculator) finalizePosition(position *types.PortfolioPosition, lastPrices map[string]*PriceInfo) {
	if position.Shares > 0 {
		position.AverageCost = position.TotalCost / position.Shares
	}

	// Add market pricing information
	if priceInfo, hasPriceInfo := lastPrices[position.Ticker]; hasPriceInfo {
		position.LastPrice = priceInfo.Price
		position.LastPriceDate = priceInfo.Date
		position.LastPriceCurrency = priceInfo.Currency
		position.MarketValue = position.Shares * priceInfo.Price
		position.UnrealizedGainLoss = position.MarketValue - position.TotalCost

		if position.TotalCost > 0 {
			position.UnrealizedGainLossPercent = (position.UnrealizedGainLoss / position.TotalCost) * PercentMultiplier
		}
	} else {
		// No price information available - use cost basis
		position.LastPrice = position.AverageCost
		position.LastPriceDate = position.LastPurchase
		position.LastPriceCurrency = position.Currency
		position.MarketValue = position.TotalCost
		position.UnrealizedGainLoss = 0
		position.UnrealizedGainLossPercent = 0
	}
}

// calculatePercentage calculates percentage safely, avoiding division by zero
func (pc *PortfolioCalculator) calculatePercentage(numerator, denominator float64) float64 {
	if denominator > 0 {
		return (numerator / denominator) * PercentMultiplier
	}
	return 0
}

// PriceInfo holds price information for a security
type PriceInfo struct {
	Price            float64
	Date             time.Time
	Currency         string
	OriginalPrice    float64
	OriginalCurrency string
}

// safeString safely dereferences a string pointer
func (pc *PortfolioCalculator) safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// handleBuyTransaction processes a buy transaction
func (pc *PortfolioCalculator) handleBuyTransaction(position *types.PortfolioPosition, tx types.Transaction) {
	if tx.Shares != nil && tx.Total != nil {
		shares := *tx.Shares
		cost := pc.convertToBaseCurrency(*tx.Total, tx.CurrencyTotal, tx.ExchangeRate)

		// Update position
		position.Shares += shares
		position.TotalCost += cost
		position.TransactionCount++

		// Update dates
		if position.FirstPurchase.IsZero() || tx.Time.Before(position.FirstPurchase) {
			position.FirstPurchase = tx.Time
		}
		if position.LastPurchase.IsZero() || tx.Time.After(position.LastPurchase) {
			position.LastPurchase = tx.Time
		}
	}
}

// handleSellTransaction processes a sell transaction
func (pc *PortfolioCalculator) handleSellTransaction(position *types.PortfolioPosition, tx types.Transaction) {
	if tx.Shares != nil && tx.Total != nil {
		shares := *tx.Shares

		// Calculate cost basis to remove (FIFO method)
		if position.Shares > 0 {
			avgCost := position.TotalCost / position.Shares
			costToRemove := avgCost * shares

			position.Shares -= shares
			position.TotalCost -= costToRemove
			position.TransactionCount++

			// Ensure we don't go negative
			if position.Shares < 0 {
				position.Shares = 0
			}
			if position.TotalCost < 0 {
				position.TotalCost = 0
			}
		}
	}
}

// getSecurityName extracts the security name from transaction
func (pc *PortfolioCalculator) getSecurityName(tx types.Transaction) string {
	if tx.Name != nil {
		return *tx.Name
	}
	if tx.Ticker != nil {
		return *tx.Ticker
	}
	return "Unknown"
}

// isTradeTransaction checks if transaction is a trade (buy/sell)
func (pc *PortfolioCalculator) isTradeTransaction(tx types.Transaction) bool {
	switch tx.Action {
	case types.TransactionTypeMarketBuy, types.TransactionTypeMarketSell,
		types.TransactionTypeLimitBuy, types.TransactionTypeLimitSell,
		types.TransactionTypeStopBuy, types.TransactionTypeStopSell:
		return true
	default:
		return false
	}
}

// isBuyTransaction checks if transaction is a buy
func (pc *PortfolioCalculator) isBuyTransaction(tx types.Transaction) bool {
	switch tx.Action {
	case types.TransactionTypeMarketBuy, types.TransactionTypeLimitBuy, types.TransactionTypeStopBuy:
		return true
	default:
		return false
	}
}

// isSellTransaction checks if transaction is a sell
func (pc *PortfolioCalculator) isSellTransaction(tx types.Transaction) bool {
	switch tx.Action {
	case types.TransactionTypeMarketSell, types.TransactionTypeLimitSell, types.TransactionTypeStopSell:
		return true
	default:
		return false
	}
}

// convertToBaseCurrency converts amount to base currency
// extractTransactionAmount extracts and converts transaction amount from Result or Total fields
func (pc *PortfolioCalculator) extractTransactionAmount(tx types.Transaction) float64 {
	var amount float64
	var currency *string
	var exchangeRate *float64

	// Get amount - check both Result and Total fields
	if tx.Result != nil && *tx.Result != 0 {
		amount = *tx.Result
		currency = tx.CurrencyResult
		exchangeRate = tx.ExchangeRate
	} else if tx.Total != nil && *tx.Total != 0 {
		amount = *tx.Total
		currency = tx.CurrencyTotal
		exchangeRate = tx.ExchangeRate
	}

	if amount != 0 {
		return pc.convertToBaseCurrency(amount, currency, exchangeRate)
	}
	return 0
}

func (pc *PortfolioCalculator) convertToBaseCurrency(amount float64, currency *string, exchangeRate *float64) float64 {
	if currency == nil || *currency == pc.baseCurrency {
		return amount
	}

	if exchangeRate != nil && *exchangeRate > 0 {
		return amount / *exchangeRate
	}

	return amount // Fallback if no exchange rate
}
