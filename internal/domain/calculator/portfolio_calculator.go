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

	// Build positions map
	positions := make(map[string]*types.PortfolioPosition)
	
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
				if tx.Result != nil {
					amount := pc.convertToBaseCurrency(*tx.Result, tx.CurrencyResult, tx.ExchangeRate)
					yearlyDividends += amount
				}
			case strings.Contains(strings.ToLower(string(tx.Action)), "interest"):
				if tx.Result != nil {
					amount := pc.convertToBaseCurrency(*tx.Result, tx.CurrencyResult, tx.ExchangeRate)
					yearlyInterest += amount
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
		}
	}

	// Convert positions map to slice and filter out zero positions
	var finalPositions []types.PortfolioPosition
	var totalShares, totalInvested float64
	
	for _, position := range positions {
		if position.Shares > 0.001 { // Filter out tiny remaining positions
			if position.Shares > 0 {
				position.AverageCost = position.TotalCost / position.Shares
			}
			finalPositions = append(finalPositions, *position)
			totalShares += position.Shares
			totalInvested += position.TotalCost
		}
	}

	// Sort finalPositions by TotalCost in descending order
	sort.Slice(finalPositions, func(i, j int) bool {
		return finalPositions[i].TotalCost > finalPositions[j].TotalCost
	})

	return &types.PortfolioSummary{
		Year:            year,
		AsOfDate:        endOfYear,
		Positions:       finalPositions,
		TotalPositions:  len(finalPositions),
		TotalShares:     totalShares,
		TotalInvested:   totalInvested,
		Currency:        pc.baseCurrency,
		YearlyDeposits:  yearlyDeposits,
		YearlyDividends: yearlyDividends,
		YearlyInterest:  yearlyInterest,
	}
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