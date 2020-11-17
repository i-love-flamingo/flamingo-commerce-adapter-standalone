package csv

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
)

type (
	// RowDto is a simple Data Transfer Object to receive CSV Rows
	RowDto map[string]string

	// ReadOption option to modify the csv.Reader
	ReadOption func(reader *csv.Reader)
)

// DelimiterOption to set the csv delimiter
func DelimiterOption(delimiter rune) ReadOption {
	return func(reader *csv.Reader) {
		if delimiter > 0 {
			reader.Comma = delimiter
		}
	}
}

// ReadCSV reads a CSV File and returns its Contents
func ReadCSV(csvFile string, options ...ReadOption) ([]RowDto, error) {
	f, err := os.Open(csvFile)
	if err != nil {
		log.Printf("Error - RowDto %v", err)
		return nil, err
	}

	var csvContents []RowDto
	var headerRow []string

	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(f))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	if options != nil {
		for _, option := range options {
			option(r)
		}
	}
	rowCount := 0
	isFirstRow := true

	for {
		rowCount++
		record, err := r.Read()

		// Stop at EOF.
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			continue
		}

		if isFirstRow {
			isFirstRow = false
			headerRow = record
			continue
		}

		row := make(map[string]string)

		for k, colName := range headerRow {
			if len(record) > k {
				row[colName] = record[k]
			}
		}

		csvContents = append(csvContents, row)
	}

	return csvContents, nil
}
