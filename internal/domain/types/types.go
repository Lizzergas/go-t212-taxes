package types

import (
	"time"
)

// TransactionType represents the type of transaction in T212
type TransactionType string

const (
	TransactionTypeMarketBuy  TransactionType = "Market buy"
	TransactionTypeMarketSell TransactionType = "Market sell"
	TransactionTypeLimitBuy   TransactionType = "Limit buy"
	TransactionTypeLimitSell  TransactionType = "Limit sell"
	TransactionTypeStopBuy    TransactionType = "Stop buy"
	TransactionTypeStopSell   TransactionType = "Stop sell"
	TransactionTypeDividend   TransactionType = "Dividend"
	TransactionTypeInterest   TransactionType = "Interest"
	TransactionTypeDeposit    TransactionType = "Deposit"
	TransactionTypeWithdrawal TransactionType = "Withdrawal"
)

// Currency represents supported currencies
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyBGN Currency = "BGN"
)

// Transaction represents a T212 transaction with full CSV format support
type Transaction struct {
	Action                                TransactionType `csv:"Action" json:"action"`
	Time                                  time.Time       `csv:"Time" json:"time"`
	ISIN                                  *string         `csv:"ISIN" json:"isin,omitempty"`
	Ticker                                *string         `csv:"Ticker" json:"ticker,omitempty"`
	Name                                  *string         `csv:"Name" json:"name,omitempty"`
	Notes                                 *string         `csv:"Notes" json:"notes,omitempty"`
	ID                                    *string         `csv:"ID" json:"id,omitempty"`
	Shares                                *float64        `csv:"No. of shares" json:"shares,omitempty"`
	PricePerShare                         *float64        `csv:"Price / share" json:"price_per_share,omitempty"`
	CurrencyPricePerShare                 *string         `csv:"Currency (Price / share)" json:"currency_price_per_share,omitempty"`
	ExchangeRate                          *float64        `csv:"Exchange rate" json:"exchange_rate,omitempty"`
	Result                                *float64        `csv:"Result" json:"result,omitempty"`
	CurrencyResult                        *string         `csv:"Currency (Result)" json:"currency_result,omitempty"`
	Total                                 *float64        `csv:"Total" json:"total,omitempty"`
	CurrencyTotal                         *string         `csv:"Currency (Total)" json:"currency_total,omitempty"`
	WithholdingTax                        *float64        `csv:"Withholding tax" json:"withholding_tax,omitempty"`
	CurrencyWithholdingTax                *string         `csv:"Currency (Withholding tax)" json:"currency_withholding_tax,omitempty"`
	ChargeAmount                          *float64        `csv:"Charge amount" json:"charge_amount,omitempty"`
	CurrencyChargeAmount                  *string         `csv:"Currency (Charge amount)" json:"currency_charge_amount,omitempty"`
	DepositFee                            *float64        `csv:"Deposit fee" json:"deposit_fee,omitempty"`
	CurrencyDepositFee                    *string         `csv:"Currency (Deposit fee)" json:"currency_deposit_fee,omitempty"`
	CurrencyConversionFromAmount          *float64        `csv:"Currency conversion from amount" json:"currency_conversion_from_amount,omitempty"`
	CurrencyCurrencyConversionFromAmount  *string         `csv:"Currency (Currency conversion from amount)" json:"currency_currency_conversion_from_amount,omitempty"`
	CurrencyConversionToAmount            *float64        `csv:"Currency conversion to amount" json:"currency_conversion_to_amount,omitempty"`
	CurrencyCurrencyConversionToAmount    *string         `csv:"Currency (Currency conversion to amount)" json:"currency_currency_conversion_to_amount,omitempty"`
	CurrencyConversionFee                 *float64        `csv:"Currency conversion fee" json:"currency_conversion_fee,omitempty"`
	CurrencyCurrencyConversionFee         *string         `csv:"Currency (Currency conversion fee)" json:"currency_currency_conversion_fee,omitempty"`
}

// TaxCalculation represents the result of tax calculations
type TaxCalculation struct {
	TotalGains        float64 `json:"total_gains"`
	TotalLosses       float64 `json:"total_losses"`
	NetGainLoss       float64 `json:"net_gain_loss"`
	DividendIncome    float64 `json:"dividend_income"`
	WithholdingTaxPaid float64 `json:"withholding_tax_paid"`
	TaxableIncome     float64 `json:"taxable_income"`
	EstimatedTax      float64 `json:"estimated_tax"`
}

// ProcessingOptions holds configuration for processing
type ProcessingOptions struct {
	TaxYear             int      `json:"tax_year"`
	Currency            Currency `json:"currency"`
	Jurisdiction        string   `json:"jurisdiction"`
	IncludeWithholdingTax bool   `json:"include_withholding_tax"`
}

// ProcessingResult represents the complete result of CSV processing
type ProcessingResult struct {
	Transactions    []Transaction     `json:"transactions"`
	TaxCalculation  TaxCalculation    `json:"tax_calculation"`
	Options         ProcessingOptions `json:"options"`
	ProcessedAt     time.Time         `json:"processed_at"`
	Summary         ProcessingSummary `json:"summary"`
}

// ProcessingSummary provides high-level statistics
type ProcessingSummary struct {
	TotalTransactions   int       `json:"total_transactions"`
	UniqueInstruments  int       `json:"unique_instruments"`
	DateRange          DateRange `json:"date_range"`
}

// YearlyReport represents financial report for a specific year
type YearlyReport struct {
	Year                  int     `json:"year"`
	TotalDeposits         float64 `json:"total_deposits"`
	TotalTransactions     int     `json:"total_transactions"`
	CapitalGains          float64 `json:"capital_gains"`
	Dividends             float64 `json:"dividends"`
	Interest              float64 `json:"interest"`
	TotalGains            float64 `json:"total_gains"`
	PercentageIncrease    float64 `json:"percentage_increase"`
	Currency              string  `json:"currency"`
}

// OverallReport represents total investment summary across all years
type OverallReport struct {
	TotalDeposits         float64       `json:"total_deposits"`
	TotalTransactions     int           `json:"total_transactions"`
	TotalCapitalGains     float64       `json:"total_capital_gains"`
	TotalDividends        float64       `json:"total_dividends"`
	TotalInterest         float64       `json:"total_interest"`
	TotalGains            float64       `json:"total_gains"`
	OverallPercentage     float64       `json:"overall_percentage"`
	Years                 []int         `json:"years"`
	YearlyReports         []YearlyReport `json:"yearly_reports"`
	Currency              string        `json:"currency"`
}

// SecurityPosition represents holdings for a specific security
type SecurityPosition struct {
	Ticker        string              `json:"ticker"`
	ISIN          string              `json:"isin"`
	Name          string              `json:"name"`
	Purchases     []PurchaseRecord    `json:"purchases"`
	CurrentShares float64             `json:"current_shares"`
}

// PurchaseRecord represents a single purchase of a security
type PurchaseRecord struct {
	Date          time.Time `json:"date"`
	Shares        float64   `json:"shares"`
	PricePerShare float64   `json:"price_per_share"`
	TotalCost     float64   `json:"total_cost"`
	Currency      string    `json:"currency"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
