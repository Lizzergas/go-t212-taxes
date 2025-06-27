package parser

import (
	"strings"
	"testing"
	"time"

	"t212-taxes/internal/domain/types"
)

func TestCSVParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		csvData string
		want    int // expected number of transactions
		wantErr bool
	}{
		{
			name: "valid CSV with multiple transactions",
			csvData: `Action,Time,ISIN,Ticker,Name,Notes,ID,No. of shares,Price / share,Currency (Price / share),Exchange rate,Result,Currency (Result),Total,Currency (Total),Withholding tax,Currency (Withholding tax),Charge amount,Currency (Charge amount),Deposit fee,Currency (Deposit fee),Currency conversion from amount,Currency (Currency conversion from amount),Currency conversion to amount,Currency (Currency conversion to amount),Currency conversion fee,Currency (Currency conversion fee)
Market buy,2024-01-15 10:30:00,US0378331005,AAPL,Apple Inc.,,1,10,150.00,USD,1.00,-1500.00,USD,-1500.00,USD,0.00,USD,0.00,USD,0.00,USD,0.00,USD,0.00,USD,0.00,USD
Dividend,2024-03-01 09:00:00,US0378331005,AAPL,Apple Inc.,Quarterly dividend,2,,0.25,USD,1.00,2.50,USD,2.50,USD,0.00,USD,0.00,USD,0.00,USD,0.00,USD,0.00,USD,0.00,USD
Deposit,2024-01-01 08:00:00,,,,,3,,,,,1000.00,EUR,1000.00,EUR,0.00,EUR,0.00,EUR,0.00,EUR,0.00,EUR,0.00,EUR,0.00,EUR`,
			want:    3,
			wantErr: false,
		},
		{
			name:    "empty CSV",
			csvData: "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "CSV with header only",
			csvData: `Action,Time,ISIN,Ticker,Name,Notes,ID,No. of shares,Price / share,Currency (Price / share),Exchange rate,Result,Currency (Result),Total,Currency (Total),Withholding tax,Currency (Withholding tax),Charge amount,Currency (Charge amount),Deposit fee,Currency (Deposit fee),Currency conversion from amount,Currency (Currency conversion from amount),Currency conversion to amount,Currency (Currency conversion to amount),Currency conversion fee,Currency (Currency conversion fee)`,
			want:    0,
			wantErr: false,
		},
		{
			name: "invalid CSV header",
			csvData: `Wrong,Headers
Market buy,2024-01-15 10:30:00`,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser()
			reader := strings.NewReader(tt.csvData)

			result, err := parser.Parse(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && len(result.Transactions) != tt.want {
				t.Errorf("Parse() got %d transactions, want %d", len(result.Transactions), tt.want)
			}
		})
	}
}

func TestCSVParser_validateHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  []string
		wantErr bool
	}{
		{
			name:    "valid header - 22 columns (2022-2023 format)",
			header:  []string{"Action", "Time", "ISIN", "Ticker", "Name", "No. of shares", "Price / share", "Currency (Price / share)", "Exchange rate", "Total", "Currency (Total)", "Withholding tax", "Currency (Withholding tax)", "Charge amount", "Currency (Charge amount)", "Deposit fee", "Currency (Deposit fee)", "ID", "Currency conversion fee", "Currency (Currency conversion fee)", "Notes", "Extra"},
			wantErr: false,
		},
		{
			name:    "missing required field",
			header:  []string{"Action", "Time", "Ticker", "Name"},
			wantErr: true,
		},
		{
			name:    "empty header",
			header:  []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser()
			err := parser.validateHeader(tt.header)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSVParser_parseTransaction(t *testing.T) {
	header := []string{"Action", "Time", "ISIN", "Ticker", "Name", "No. of shares", "Price / share", "Result"}

	tests := []struct {
		name       string
		record     []string
		wantAction types.TransactionType
		wantErr    bool
	}{
		{
			name:       "valid market buy",
			record:     []string{"Market buy", "2024-01-15 10:30:00", "US0378331005", "AAPL", "Apple Inc.", "10", "150.00", "-1500.00"},
			wantAction: types.TransactionTypeMarketBuy,
			wantErr:    false,
		},
		{
			name:       "valid dividend",
			record:     []string{"Dividend", "2024-03-01 09:00:00", "US0378331005", "AAPL", "Apple Inc.", "", "0.25", "2.50"},
			wantAction: types.TransactionTypeDividend,
			wantErr:    false,
		},
		{
			name:    "missing action",
			record:  []string{"", "2024-01-15 10:30:00", "US0378331005", "AAPL", "Apple Inc.", "10", "150.00", "-1500.00"},
			wantErr: true,
		},
		{
			name:    "invalid time format",
			record:  []string{"Market buy", "invalid-time", "US0378331005", "AAPL", "Apple Inc.", "10", "150.00", "-1500.00"},
			wantErr: true,
		},
		{
			name:    "header/record length mismatch",
			record:  []string{"Market buy", "2024-01-15 10:30:00", "US0378331005"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser()
			transaction, err := parser.parseTransaction(header, tt.record)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && transaction.Action != tt.wantAction {
				t.Errorf("parseTransaction() action = %v, want %v", transaction.Action, tt.wantAction)
			}
		})
	}
}

func TestCSVParser_validateYearlyStructure(t *testing.T) {
	tests := []struct {
		name      string
		filenames []string
		wantErr   bool
	}{
		{
			name: "valid yearly structure",
			filenames: []string{
				"from_2022-01-01_to_2022-12-31_abc123.csv",
				"from_2023-01-01_to_2023-12-31_def456.csv",
			},
			wantErr: false,
		},
		{
			name: "invalid filename format",
			filenames: []string{
				"invalid_filename.csv",
			},
			wantErr: true,
		},
		{
			name: "cross-year range",
			filenames: []string{
				"from_2022-12-01_to_2023-01-31_abc123.csv",
			},
			wantErr: true,
		},
		{
			name: "duplicate year",
			filenames: []string{
				"from_2022-01-01_to_2022-06-30_abc123.csv",
				"from_2022-07-01_to_2022-12-31_def456.csv",
			},
			wantErr: true,
		},
		{
			name: "invalid date range",
			filenames: []string{
				"from_2022-12-31_to_2022-01-01_abc123.csv",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser()
			err := parser.ValidateYearlyStructure(tt.filenames)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateYearlyStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSVParser_parseOptionalFloat(t *testing.T) {
	tests := []struct {
		name      string
		fieldMap  map[string]string
		fieldName string
		want      *float64
		wantErr   bool
	}{
		{
			name:      "valid float",
			fieldMap:  map[string]string{"test": "123.45"},
			fieldName: "test",
			want:      floatPtr(123.45),
			wantErr:   false,
		},
		{
			name:      "empty field",
			fieldMap:  map[string]string{"test": ""},
			fieldName: "test",
			want:      nil,
			wantErr:   false,
		},
		{
			name:      "zero value",
			fieldMap:  map[string]string{"test": "0"},
			fieldName: "test",
			want:      nil,
			wantErr:   false,
		},
		{
			name:      "invalid float",
			fieldMap:  map[string]string{"test": "invalid"},
			fieldName: "test",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing field",
			fieldMap:  map[string]string{},
			fieldName: "test",
			want:      nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser()
			var target *float64
			err := parser.parseOptionalFloat(tt.fieldMap, tt.fieldName, &target)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseOptionalFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !floatPtrEqual(target, tt.want) {
				t.Errorf("parseOptionalFloat() target = %v, want %v", target, tt.want)
			}
		})
	}
}

func TestCSVParser_calculateSummary(t *testing.T) {
	transactions := []types.Transaction{
		{
			Action: types.TransactionTypeMarketBuy,
			Time:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Ticker: stringPtr("AAPL"),
		},
		{
			Action: types.TransactionTypeDividend,
			Time:   time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC),
			Ticker: stringPtr("AAPL"),
		},
		{
			Action: types.TransactionTypeMarketBuy,
			Time:   time.Date(2024, 3, 1, 14, 0, 0, 0, time.UTC),
			Ticker: stringPtr("GOOGL"),
		},
	}

	parser := NewCSVParser()
	summary := parser.calculateSummary(transactions)

	if summary.TotalTransactions != 3 {
		t.Errorf("calculateSummary() TotalTransactions = %d, want 3", summary.TotalTransactions)
	}

	if summary.UniqueInstruments != 2 {
		t.Errorf("calculateSummary() UniqueInstruments = %d, want 2", summary.UniqueInstruments)
	}

	expectedFrom := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	expectedTo := time.Date(2024, 3, 1, 14, 0, 0, 0, time.UTC)

	if !summary.DateRange.From.Equal(expectedFrom) {
		t.Errorf("calculateSummary() DateRange.From = %v, want %v", summary.DateRange.From, expectedFrom)
	}

	if !summary.DateRange.To.Equal(expectedTo) {
		t.Errorf("calculateSummary() DateRange.To = %v, want %v", summary.DateRange.To, expectedTo)
	}
}

// Helper functions for tests
func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

func floatPtrEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
