package domain

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// csvHeader names the columns of an exported measurement list.
var csvHeader = []string{"nombre", "fecha", "largo_mm", "ancho_mm", "espesor_mm", "feret_max_mm", "avisos"}

// WriteMeasurementsCSV writes the measurements as CSV for a spreadsheet in Spanish: ";" between
// fields and a decimal comma.
func WriteMeasurementsCSV(writer io.Writer, measurements []Measurement) error {
	output := csv.NewWriter(writer)
	output.Comma = ';'
	if err := output.Write(csvHeader); err != nil {
		return fmt.Errorf("writing the CSV header: %w", err)
	}
	for _, measurement := range measurements {
		if err := output.Write(csvRecord(measurement)); err != nil {
			return fmt.Errorf("writing %s to the CSV: %w", measurement.Name, err)
		}
	}
	output.Flush()
	return output.Error()
}

// csvRecord is one measurement's row; an unknown date is left empty.
func csvRecord(measurement Measurement) []string {
	date := ""
	if !measurement.Date.IsZero() {
		date = measurement.Date.Format(calibreDate)
	}
	return []string{
		measurement.Name, date,
		decimalComma(measurement.Length), decimalComma(measurement.Width),
		decimalComma(measurement.Thickness), decimalComma(measurement.Feret),
		strings.Join(measurement.Warnings, " | "),
	}
}

// decimalComma writes millimetres to four decimals, as calibre3d does, with a decimal comma.
func decimalComma(value float64) string {
	return strings.Replace(strconv.FormatFloat(value, 'f', 4, 64), ".", ",", 1)
}
