package csv

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"

	"log"
)

type (
	// ProductCsvRowDto is a simple Data Transfer Object to receive CSV Rows
	ProductCsvRowDto map[string]string
)

// ReadProductCSV reads a CSV File and returns its Contents
func ReadProductCSV(csvFile string) ([]ProductCsvRowDto, error) {
	f, err := os.Open(csvFile)
	if err != nil {
		log.Printf("Error - ProductCsvRowDto %v", err)
		return nil, err
	}

	var csvContents []ProductCsvRowDto
	var headerRow []string

	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(f))
	r.LazyQuotes = true
	r.Comma = ';'
	rowCount := 0
	isFirstRow := true

	for {
		rowCount++
		record, err := r.Read()

		// Stop at EOF.
		if err == io.EOF {
			break
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
