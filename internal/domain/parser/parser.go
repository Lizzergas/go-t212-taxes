// Package parser provides CSV parsing functionality for Trading 212 export files
package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Lizzergas/go-t212-taxes/internal/domain/types"
)

// Constants
const (
	MaxErrorsDisplayed = 10
	LineNumberOffset   = 2
	MinHeaderColumns   = 22
	RegexMatchGroups   = 3
)

// Parser handles CSV parsing for T212 files
type Parser interface {
	Parse(reader io.Reader) (*types.ProcessingResult, error)
	ParseFile(filename string) (*types.ProcessingResult, error)
	ValidateFormat(reader io.Reader) error
}

// CSVParser implements Parser for CSV files
type CSVParser struct {
	skipHeader bool
	delimiter  rune
}

// FileService interface for file operations (useful for testing)
type FileService interface {
	ReadFile(filename string) ([]byte, error)
	ListFiles(directory string) ([]string, error)
}

// NewCSVParser creates a new CSV parser
func NewCSVParser() *CSVParser {
	return &CSVParser{
		skipHeader: true,
		delimiter:  ',',
	}
}

// Parse processes CSV data from a reader
func (p *CSVParser) Parse(reader io.Reader) (*types.ProcessingResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = p.delimiter
	csvReader.LazyQuotes = true       // Handle malformed quotes more gracefully
	csvReader.TrimLeadingSpace = true // Handle leading spaces
	csvReader.FieldsPerRecord = -1    // Allow variable number of fields

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Parse header
	header := records[0]
	if err := p.validateHeader(header); err != nil {
		return nil, fmt.Errorf("invalid CSV header: %w", err)
	}

	// Parse transactions
	transactions := make([]types.Transaction, 0, len(records)-1)
	successfulParsed := 0
	failedParsed := 0

	for i, record := range records[1:] {
		transaction, err := p.parseTransaction(header, record)
		if err != nil {
			failedParsed++
			if failedParsed <= MaxErrorsDisplayed { // Only show first 10 errors to avoid spam
				log.Printf("Warning: failed to parse transaction at line %d: %v", i+LineNumberOffset, err)
			}
			continue
		}
		transactions = append(transactions, *transaction)
		successfulParsed++
	}

	if failedParsed > 0 {
		log.Printf("Summary: Successfully parsed %d transactions, failed to parse %d transactions", successfulParsed, failedParsed)
	}

	// Sort transactions by time
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Time.Before(transactions[j].Time)
	})

	// Calculate summary
	summary := p.calculateSummary(transactions)

	return &types.ProcessingResult{
		Transactions:   transactions,
		TaxCalculation: types.TaxCalculation{},
		Options: types.ProcessingOptions{
			TaxYear:      time.Now().Year(),
			Currency:     types.CurrencyEUR,
			Jurisdiction: "EU",
		},
		ProcessedAt: time.Now(),
		Summary:     summary,
	}, nil
}

// ParseFile processes a CSV file
func (p *CSVParser) ParseFile(filename string) (*types.ProcessingResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close() //nolint:errcheck

	return p.Parse(file)
}

// ValidateFormat checks if the CSV format is valid for T212
func (p *CSVParser) ValidateFormat(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = p.delimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	return p.validateHeader(records[0])
}

// SetDelimiter sets the CSV delimiter
func (p *CSVParser) SetDelimiter(delimiter rune) {
	p.delimiter = delimiter
}

// SetSkipHeader sets whether to skip the header row
func (p *CSVParser) SetSkipHeader(skip bool) {
	p.skipHeader = skip
}

// ParseMultipleFiles processes multiple CSV files and combines results
func (p *CSVParser) ParseMultipleFiles(filenames []string) (*types.ProcessingResult, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// Validate yearly structure
	if err := p.ValidateYearlyStructure(filenames); err != nil {
		return nil, fmt.Errorf("yearly validation failed: %w", err)
	}

	allTransactions := make([]types.Transaction, 0)

	for _, filename := range filenames {
		result, err := p.ParseFile(filename)
		if err != nil {
			log.Printf("Warning: failed to parse file %s: %v", filename, err)
			continue
		}
		allTransactions = append(allTransactions, result.Transactions...)
	}

	// Sort all transactions by time
	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].Time.Before(allTransactions[j].Time)
	})

	// Calculate combined summary
	summary := p.calculateSummary(allTransactions)

	return &types.ProcessingResult{
		Transactions:   allTransactions,
		TaxCalculation: types.TaxCalculation{},
		Options: types.ProcessingOptions{
			TaxYear:      time.Now().Year(),
			Currency:     types.CurrencyEUR,
			Jurisdiction: "EU",
		},
		ProcessedAt: time.Now(),
		Summary:     summary,
	}, nil
}

// validateHeader checks if the CSV header contains required fields
func (p *CSVParser) validateHeader(header []string) error {
	requiredFields := []string{"Action", "Time", "ISIN", "Ticker", "Name"}

	headerMap := make(map[string]bool)
	for _, field := range header {
		headerMap[strings.TrimSpace(field)] = true
	}

	for _, required := range requiredFields {
		if !headerMap[required] {
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	// Trading 212 has evolved their CSV format over time:
	// - 2022-2023: 22 columns (missing Result field)
	// - 2021: 23 columns
	// - 2024+: 27 columns (added currency conversion fields)
	if len(header) < MinHeaderColumns {
		return fmt.Errorf("CSV has %d columns, expected at least 22", len(header))
	}

	if len(header) != 22 && len(header) != 23 && len(header) != 27 {
		return fmt.Errorf("CSV has %d columns, expected 22, 23, or 27 (different Trading 212 format versions)", len(header))
	}

	return nil
}

// parseTransaction converts a CSV record to a Transaction
func (p *CSVParser) parseTransaction(header []string, record []string) (*types.Transaction, error) {
	// Handle field count mismatches
	record, err := p.normalizeRecord(header, record)
	if err != nil {
		return nil, err
	}

	// Create field map
	fieldMap := p.createFieldMap(header, record)

	transaction := &types.Transaction{}

	// Parse required fields
	if err := p.parseRequiredFields(fieldMap, transaction); err != nil {
		return nil, err
	}

	// Parse optional fields
	p.parseOptionalStringFields(fieldMap, transaction)

	if err := p.parseOptionalNumericFields(fieldMap, transaction); err != nil {
		return nil, err
	}

	p.parseOptionalCurrencyFields(fieldMap, transaction)

	return transaction, nil
}

// normalizeRecord handles field count mismatches between header and record
func (p *CSVParser) normalizeRecord(header, record []string) ([]string, error) {
	if len(record) == len(header) {
		return record, nil
	}

	// If record is severely shorter or longer, skip it
	if len(record) < len(header)/2 || len(record) > len(header)*2 {
		return nil, fmt.Errorf("severe field count mismatch (skipping): expected %d, got %d", len(header), len(record))
	}

	// For minor mismatches, try to fix
	if len(record) < len(header) {
		// Pad with empty strings
		for len(record) < len(header) {
			record = append(record, "")
		}
	} else {
		// Truncate extra fields
		record = record[:len(header)]
	}

	return record, nil
}

// createFieldMap creates a map from header to record values
func (p *CSVParser) createFieldMap(header, record []string) map[string]string {
	fieldMap := make(map[string]string)
	for i, field := range header {
		fieldName := strings.TrimSpace(field)
		fieldValue := strings.TrimSpace(record[i])
		fieldMap[fieldName] = fieldValue
	}
	return fieldMap
}

// parseRequiredFields parses required action and time fields
func (p *CSVParser) parseRequiredFields(fieldMap map[string]string, transaction *types.Transaction) error {
	// Parse Action (required)
	action := fieldMap["Action"]
	if action == "" {
		return fmt.Errorf("missing action field")
	}
	transaction.Action = types.TransactionType(action)

	// Parse Time (required)
	timeStr := fieldMap["Time"]
	if timeStr == "" {
		return fmt.Errorf("missing time field")
	}

	parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		// Try alternative format
		parsedTime, err = time.Parse("2006-01-02T15:04:05", timeStr)
		if err != nil {
			return fmt.Errorf("failed to parse time %s: %w", timeStr, err)
		}
	}
	transaction.Time = parsedTime

	return nil
}

// parseOptionalStringFields parses optional string fields
func (p *CSVParser) parseOptionalStringFields(fieldMap map[string]string, transaction *types.Transaction) {
	stringFields := []struct {
		fieldName string
		target    **string
	}{
		{"ISIN", &transaction.ISIN},
		{"Ticker", &transaction.Ticker},
		{"Name", &transaction.Name},
		{"Notes", &transaction.Notes},
		{"ID", &transaction.ID},
	}

	for _, field := range stringFields {
		if value := fieldMap[field.fieldName]; value != "" {
			*field.target = &value
		}
	}
}

// parseOptionalNumericFields parses all optional numeric fields
func (p *CSVParser) parseOptionalNumericFields(fieldMap map[string]string, transaction *types.Transaction) error {
	// Standard numeric fields
	numericFields := []struct {
		fieldName string
		target    **float64
	}{
		{"No. of shares", &transaction.Shares},
		{"Price / share", &transaction.PricePerShare},
		{"Exchange rate", &transaction.ExchangeRate},
		{"Total", &transaction.Total},
		{"Withholding tax", &transaction.WithholdingTax},
		{"Charge amount", &transaction.ChargeAmount},
		{"Deposit fee", &transaction.DepositFee},
		{"Currency conversion fee", &transaction.CurrencyConversionFee},
	}

	for _, field := range numericFields {
		if err := p.parseOptionalFloat(fieldMap, field.fieldName, field.target); err != nil {
			return err
		}
	}

	// Conditional fields that may not exist in older formats
	conditionalFields := []struct {
		fieldName string
		target    **float64
	}{
		{"Result", &transaction.Result},
		{"Currency conversion from amount", &transaction.CurrencyConversionFromAmount},
		{"Currency conversion to amount", &transaction.CurrencyConversionToAmount},
	}

	for _, field := range conditionalFields {
		if _, exists := fieldMap[field.fieldName]; exists {
			if err := p.parseOptionalFloat(fieldMap, field.fieldName, field.target); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseOptionalCurrencyFields parses all optional currency fields
func (p *CSVParser) parseOptionalCurrencyFields(fieldMap map[string]string, transaction *types.Transaction) {
	// Standard currency fields
	currencyFields := []struct {
		fieldName string
		target    **string
	}{
		{"Currency (Price / share)", &transaction.CurrencyPricePerShare},
		{"Currency (Total)", &transaction.CurrencyTotal},
		{"Currency (Withholding tax)", &transaction.CurrencyWithholdingTax},
		{"Currency (Charge amount)", &transaction.CurrencyChargeAmount},
		{"Currency (Deposit fee)", &transaction.CurrencyDepositFee},
	}

	for _, field := range currencyFields {
		if curr := fieldMap[field.fieldName]; curr != "" {
			*field.target = &curr
		}
	}

	// Conditional currency fields that may not exist in older formats
	conditionalCurrencyFields := []struct {
		fieldName string
		target    **string
	}{
		{"Currency (Result)", &transaction.CurrencyResult},
		{"Currency (Currency conversion from amount)", &transaction.CurrencyCurrencyConversionFromAmount},
		{"Currency (Currency conversion to amount)", &transaction.CurrencyCurrencyConversionToAmount},
		{"Currency (Currency conversion fee)", &transaction.CurrencyCurrencyConversionFee},
	}

	for _, field := range conditionalCurrencyFields {
		if curr, exists := fieldMap[field.fieldName]; exists && curr != "" {
			*field.target = &curr
		}
	}
}

// parseOptionalFloat parses a float field if it exists and is not empty
func (p *CSVParser) parseOptionalFloat(fieldMap map[string]string, fieldName string, target **float64) error {
	valueStr := fieldMap[fieldName]
	if valueStr == "" || valueStr == "0" {
		return nil
	}

	// Handle Trading 212's "Not available" values
	if valueStr == "Not available" {
		return nil
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", fieldName, err)
	}

	*target = &value
	return nil
}

// calculateSummary calculates processing summary from transactions
func (p *CSVParser) calculateSummary(transactions []types.Transaction) types.ProcessingSummary {
	if len(transactions) == 0 {
		return types.ProcessingSummary{}
	}

	// Calculate date range
	minTime := transactions[0].Time
	maxTime := transactions[0].Time

	// Track unique instruments
	uniqueTickers := make(map[string]bool)

	for _, transaction := range transactions {
		if transaction.Time.Before(minTime) {
			minTime = transaction.Time
		}
		if transaction.Time.After(maxTime) {
			maxTime = transaction.Time
		}

		if transaction.Ticker != nil && *transaction.Ticker != "" {
			uniqueTickers[*transaction.Ticker] = true
		}
	}

	return types.ProcessingSummary{
		TotalTransactions: len(transactions),
		UniqueInstruments: len(uniqueTickers),
		DateRange: types.DateRange{
			From: minTime,
			To:   maxTime,
		},
	}
}

// ValidateYearlyStructure validates that CSV files follow yearly naming convention
func (p *CSVParser) ValidateYearlyStructure(filenames []string) error {
	pattern := regexp.MustCompile(`from_(\d{4}-\d{2}-\d{2})_to_(\d{4}-\d{2}-\d{2})_[A-Za-z0-9]+\.csv$`)

	yearsSeen := make(map[int]bool)

	for _, filename := range filenames {
		base := filepath.Base(filename)
		matches := pattern.FindStringSubmatch(base)

		if len(matches) != RegexMatchGroups {
			return fmt.Errorf("invalid filename format: %s (expected: from_YYYY-MM-DD_to_YYYY-MM-DD_[hash].csv)", base)
		}

		startDate, err := time.Parse("2006-01-02", matches[1])
		if err != nil {
			return fmt.Errorf("invalid start date in filename %s: %w", base, err)
		}

		endDate, err := time.Parse("2006-01-02", matches[2])
		if err != nil {
			return fmt.Errorf("invalid end date in filename %s: %w", base, err)
		}

		if startDate.After(endDate) {
			return fmt.Errorf("start date after end date in filename %s", base)
		}

		startYear := startDate.Year()
		endYear := endDate.Year()

		if startYear != endYear {
			return fmt.Errorf("date range spans multiple years in filename %s", base)
		}

		if yearsSeen[startYear] {
			return fmt.Errorf("duplicate year %d found in filenames", startYear)
		}

		yearsSeen[startYear] = true
	}

	return nil
}
