package domain

import (
	"strings"
	"testing"
	"time"
)

func TestTheCSVHasAHeaderSemicolonsAndDecimalCommas(t *testing.T) {
	var output strings.Builder
	measurements := []Measurement{{
		Name: "pieza1", Date: time.Date(2026, 11, 24, 15, 30, 12, 0, time.Local),
		Length: 61.23, Width: 40.5, Thickness: 19.875, Feret: 62, Warnings: []string{"lejos", "inclinada"},
	}}
	if err := WriteMeasurementsCSV(&output, measurements); err != nil {
		t.Fatal(err)
	}
	want := "nombre;fecha;largo_mm;ancho_mm;espesor_mm;feret_max_mm;avisos\n" +
		"pieza1;2026-11-24T15:30:12;61,2300;40,5000;19,8750;62,0000;lejos | inclinada\n"
	if output.String() != want {
		t.Errorf("got\n%s\nwant\n%s", output.String(), want)
	}
}

func TestTheCSVQuotesANameWithASemicolonAndLeavesAnUnknownDateEmpty(t *testing.T) {
	var output strings.Builder
	if err := WriteMeasurementsCSV(&output, []Measurement{{Name: "a;b"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "\n\"a;b\";;0,0000;") {
		t.Errorf("got %q", output.String())
	}
}
