package domain

import (
	"slices"
	"testing"
)

func bound(value float64) *float64 { return &value }

func piece(name string, length, width, thickness float64) Measurement {
	return Measurement{Name: name, Length: length, Width: width, Thickness: thickness}
}

func TestAnEmptyFilterMatchesEveryMeasurement(t *testing.T) {
	if !(MeasurementFilter{}).Matches(piece("any", 1, 2, 3)) {
		t.Error("an empty filter left a measurement out")
	}
}

func TestTheNameFilterFindsPartOfTheNameIgnoringCase(t *testing.T) {
	filter := MeasurementFilter{Name: " TAPA "}
	if !filter.Matches(piece("caja-tapa-2", 0, 0, 0)) {
		t.Error("part of the name in another case did not match")
	}
	if filter.Matches(piece("base", 0, 0, 0)) {
		t.Error("a name without the text matched")
	}
}

func TestEachAxisIsFilteredByItsOwnRange(t *testing.T) {
	measurement := piece("p", 60, 40, 20)
	cases := []struct {
		name   string
		filter MeasurementFilter
		want   bool
	}{
		{"length above the minimum", MeasurementFilter{Length: Range{Min: bound(50)}}, true},
		{"length below the minimum", MeasurementFilter{Length: Range{Min: bound(61)}}, false},
		{"width over the maximum", MeasurementFilter{Width: Range{Max: bound(39.9)}}, false},
		{"thickness inside both ends", MeasurementFilter{Thickness: Range{Min: bound(10), Max: bound(30)}}, true},
		{"thickness outside both ends", MeasurementFilter{Thickness: Range{Min: bound(21), Max: bound(30)}}, false},
	}
	for _, test := range cases {
		if got := test.filter.Matches(measurement); got != test.want {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
		}
	}
}

func TestRangeEndsAreInclusive(t *testing.T) {
	if !(Range{Min: bound(20), Max: bound(20)}).Contains(20) {
		t.Error("a value equal to both ends was left out")
	}
}

func TestFilterMeasurementsKeepsTheOrder(t *testing.T) {
	all := []Measurement{piece("a1", 10, 0, 0), piece("b", 50, 0, 0), piece("a2", 30, 0, 0)}
	got := FilterMeasurements(all, MeasurementFilter{Name: "a"})
	if names := measurementNames(got); !slices.Equal(names, []string{"a1", "a2"}) {
		t.Errorf("got %v", names)
	}
}

func TestParseBoundReadsCommaPointAndEmpty(t *testing.T) {
	for text, want := range map[string]float64{"12,5": 12.5, " 7.25 ": 7.25, "3": 3} {
		got, err := ParseBound(text)
		if err != nil || got == nil || *got != want {
			t.Errorf("%q: got %v, %v", text, got, err)
		}
	}
	if got, err := ParseBound("  "); got != nil || err != nil {
		t.Errorf("empty: got %v, %v", got, err)
	}
	if _, err := ParseBound("doce"); err == nil {
		t.Error("a word was read as a number")
	}
}

func measurementNames(measurements []Measurement) []string {
	names := make([]string, len(measurements))
	for index, measurement := range measurements {
		names[index] = measurement.Name
	}
	return names
}
