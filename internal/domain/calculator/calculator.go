package calculator

import (
	"log"

	"t212-taxes/internal/domain/types"
)

// Calculator handles tax calculations for different jurisdictions
type Calculator interface {
	Calculate(transactions []types.Transaction, options types.ProcessingOptions) (*types.TaxCalculation, error)
	CalculateCapitalGains(transactions []types.Transaction, options types.ProcessingOptions) (float64, float64, error)
	CalculateDividends(transactions []types.Transaction, options types.ProcessingOptions) (float64, float64, error)
}

// TaxCalculator implements Calculator
type TaxCalculator struct {
	jurisdictions map[string]TaxJurisdiction
}

// TaxJurisdiction represents tax rules for a jurisdiction
type TaxJurisdiction struct {
	Code                string
	Name                string
	CapitalGainsTaxRate float64
	DividendTaxRate     float64
	Allowances          TaxAllowances
}

// TaxAllowances represents tax-free allowances
type TaxAllowances struct {
	CapitalGains float64
	Dividends    float64
}

// NewTaxCalculator creates a new tax calculator
func NewTaxCalculator() *TaxCalculator {
	return &TaxCalculator{
		jurisdictions: map[string]TaxJurisdiction{
			"US": {
				Code:                "US",
				Name:                "United States",
				CapitalGainsTaxRate: 0.15,
				DividendTaxRate:     0.15,
				Allowances: TaxAllowances{
					CapitalGains: 0,
					Dividends:    0,
				},
			},
			"UK": {
				Code:                "UK",
				Name:                "United Kingdom",
				CapitalGainsTaxRate: 0.10,
				DividendTaxRate:     0.075,
				Allowances: TaxAllowances{
					CapitalGains: 6000,
					Dividends:    2000,
				},
			},
			"BG": {
				Code:                "BG",
				Name:                "Bulgaria",
				CapitalGainsTaxRate: 0.10,
				DividendTaxRate:     0.05,
				Allowances: TaxAllowances{
					CapitalGains: 0,
					Dividends:    0,
				},
			},
		},
	}
}

// Calculate performs comprehensive tax calculations
func (c *TaxCalculator) Calculate(transactions []types.Transaction, options types.ProcessingOptions) (*types.TaxCalculation, error) {
	log.Println("🚧 Tax Calculator is under development")
	log.Printf("Processing %d transactions for jurisdiction %s", len(transactions), options.Jurisdiction)

	// Hello World implementation - return mock calculation
	return &types.TaxCalculation{
		TotalGains:         0,
		TotalLosses:        0,
		NetGainLoss:        0,
		DividendIncome:     0,
		WithholdingTaxPaid: 0,
		TaxableIncome:      0,
		EstimatedTax:       0,
	}, nil
}

// CalculateCapitalGains calculates capital gains and losses
func (c *TaxCalculator) CalculateCapitalGains(transactions []types.Transaction, options types.ProcessingOptions) (float64, float64, error) {
	log.Printf("🚧 Capital Gains Calculator is under development - processing %d transactions", len(transactions))

	return 0.0, 0.0, nil
}

// CalculateDividends calculates dividend income and withholding tax
func (c *TaxCalculator) CalculateDividends(transactions []types.Transaction, options types.ProcessingOptions) (float64, float64, error) {
	log.Printf("🚧 Dividend Calculator is under development - processing %d transactions", len(transactions))

	return 0.0, 0.0, nil
}

// GetJurisdiction returns tax jurisdiction details
func (c *TaxCalculator) GetJurisdiction(code string) (*TaxJurisdiction, bool) {
	jurisdiction, exists := c.jurisdictions[code]
	return &jurisdiction, exists
}

// GetSupportedJurisdictions returns all supported jurisdictions
func (c *TaxCalculator) GetSupportedJurisdictions() []TaxJurisdiction {
	jurisdictions := make([]TaxJurisdiction, 0, len(c.jurisdictions))
	for _, jurisdiction := range c.jurisdictions {
		jurisdictions = append(jurisdictions, jurisdiction)
	}
	return jurisdictions
}
