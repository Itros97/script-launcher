package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Range bounds a size in millimetres; a nil end does not bound it. Both ends are inclusive.
type Range struct {
	Min *float64
	Max *float64
}

// MeasurementFilter picks measurements by part of their name and by ranges on the three axes.
type MeasurementFilter struct {
	Name      string
	Length    Range
	Width     Range
	Thickness Range
}

// Contains tells whether value is inside the range.
func (bounds Range) Contains(value float64) bool {
	if bounds.Min != nil && value < *bounds.Min {
		return false
	}
	return bounds.Max == nil || value <= *bounds.Max
}

// Matches tells whether the measurement's name holds the filter's text, ignoring case, and its sizes fit every range.
func (filter MeasurementFilter) Matches(measurement Measurement) bool {
	if !strings.Contains(strings.ToLower(measurement.Name), strings.ToLower(strings.TrimSpace(filter.Name))) {
		return false
	}
	return filter.Length.Contains(measurement.Length) &&
		filter.Width.Contains(measurement.Width) &&
		filter.Thickness.Contains(measurement.Thickness)
}

// FilterMeasurements returns the measurements that match, in the same order.
func FilterMeasurements(measurements []Measurement, filter MeasurementFilter) []Measurement {
	matching := make([]Measurement, 0, len(measurements))
	for _, measurement := range measurements {
		if filter.Matches(measurement) {
			matching = append(matching, measurement)
		}
	}
	return matching
}

// ParseBound reads a typed limit in millimetres, with a decimal comma or point; empty is no limit.
func ParseBound(text string) (*float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(strings.Replace(text, ",", ".", 1), 64)
	if err != nil {
		return nil, fmt.Errorf("%q is not a number of millimetres", text)
	}
	return &value, nil
}
