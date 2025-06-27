package calculator

import (
	"sort"
	"strings"
	"time"

	"t212-taxes/internal/domain/types"
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
	// Filter transactions up to end of year
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	var relevantTransactions []types.Transaction
	
	for _, tx := range transactions {
		if tx.Time.Before(endOfYear) || tx.Time.Equal(endOfYear) {
			relevantTransactions = append(relevantTransactions, tx)
		}
	}

	// Build positions map and track last prices
	positions := make(map[string]*types.PortfolioPosition)
	lastPrices := make(map[string]*PriceInfo)
	
	// Track yearly metrics
	var yearlyDeposits, yearlyDividends, yearlyInterest float64
	
	for _, tx := range relevantTransactions {
		// Handle yearly metrics (only for the specific year)
		if tx.Time.Year() == year {
			switch {
			case tx.Action == types.TransactionTypeDeposit:
				if tx.Total != nil {
					amount := pc.convertToBaseCurrency(*tx.Total, tx.CurrencyTotal, tx.ExchangeRate)
					yearlyDeposits += amount
				}
			case strings.Contains(strings.ToLower(string(tx.Action)), "dividend"):
				var amount float64
				var currency *string
				var exchangeRate *float64
				
				// Get dividend amount - check both Result and Total fields
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
					convertedAmount := pc.convertToBaseCurrency(amount, currency, exchangeRate)
					yearlyDividends += convertedAmount
				}
			case strings.Contains(strings.ToLower(string(tx.Action)), "interest"):
				var amount float64
				var currency *string
				var exchangeRate *float64
				
				// Get interest amount - check both Result and Total fields
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
					convertedAmount := pc.convertToBaseCurrency(amount, currency, exchangeRate)
					yearlyInterest += convertedAmount
				}
			}
		}

		// Handle trade transactions for portfolio positions
		if pc.isTradeTransaction(tx) && tx.Ticker != nil && tx.ISIN != nil {
			ticker := *tx.Ticker
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

			// Update position based on transaction type
			if pc.isBuyTransaction(tx) {
				pc.handleBuyTransaction(position, tx)
			} else if pc.isSellTransaction(tx) {
				pc.handleSellTransaction(position, tx)
			}
			
			// Track last price information
			if tx.PricePerShare != nil && *tx.PricePerShare > 0 {
				priceInBaseCurrency := pc.convertToBaseCurrency(*tx.PricePerShare, tx.CurrencyPricePerShare, tx.ExchangeRate)
				
				lastPrices[ticker] = &PriceInfo{
					Price:    priceInBaseCurrency,
					Date:     tx.Time,
					Currency: pc.baseCurrency,
					OriginalPrice: *tx.PricePerShare,
					OriginalCurrency: pc.safeString(tx.CurrencyPricePerShare),
				}
			}
		}
	}

	// Convert positions map to slice and filter out zero positions
	var finalPositions []types.PortfolioPosition
	var totalShares, totalInvested, totalMarketValue float64
	
	for _, position := range positions {
		if position.Shares > 0.001 { // Filter out tiny remaining positions
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
					position.UnrealizedGainLossPercent = (position.UnrealizedGainLoss / position.TotalCost) * 100
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
			
			finalPositions = append(finalPositions, *position)
			totalShares += position.Shares
			totalInvested += position.TotalCost
			totalMarketValue += position.MarketValue
		}
	}

	// Sort finalPositions by MarketValue in descending order
	sort.Slice(finalPositions, func(i, j int) bool {
		return finalPositions[i].MarketValue > finalPositions[j].MarketValue
	})

	// Calculate total unrealized gains/losses
	totalUnrealizedGainLoss := totalMarketValue - totalInvested
	var totalUnrealizedGainLossPercent float64
	if totalInvested > 0 {
		totalUnrealizedGainLossPercent = (totalUnrealizedGainLoss / totalInvested) * 100
	}

	return &types.PortfolioSummary{
		Year:            year,
		AsOfDate:        endOfYear,
		Positions:       finalPositions,
		TotalPositions:  len(finalPositions),
		TotalShares:     totalShares,
		TotalInvested:   totalInvested,
		TotalMarketValue: totalMarketValue,
		TotalUnrealizedGainLoss: totalUnrealizedGainLoss,
		TotalUnrealizedGainLossPercent: totalUnrealizedGainLossPercent,
		Currency:        pc.baseCurrency,
		YearlyDeposits:  yearlyDeposits,
		YearlyDividends: yearlyDividends,
		YearlyInterest:  yearlyInterest,
	}
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
func (pc *PortfolioCalculator) convertToBaseCurrency(amount float64, currency *string, exchangeRate *float64) float64 {
	if currency == nil || *currency == pc.baseCurrency {
		return amount
	}
	
	if exchangeRate != nil && *exchangeRate > 0 {
		return amount / *exchangeRate
	}
	
	return amount // Fallback if no exchange rate
} 