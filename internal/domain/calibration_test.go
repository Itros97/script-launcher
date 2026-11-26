package domain

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

func TestParsesAPatternWithADecimalComma(t *testing.T) {
	pattern, err := ParsePattern("bloque_25", "25,4", " 20.05 ", "10")
	if err != nil {
		t.Fatal(err)
	}
	if pattern != (Pattern{Name: "bloque_25", Length: 25.4, Width: 20.05, Thickness: 10}) {
		t.Errorf("got %+v", pattern)
	}
}

func TestRejectsAPatternSizeThatIsNotAboveZero(t *testing.T) {
	for _, size := range []string{"", "abc", "0", "-3", "1,2,3"} {
		if _, err := ParsePattern("bloque", size, "20", "10"); !errors.Is(err, ErrInvalidSize) {
			t.Errorf("%q: got %v, want ErrInvalidSize", size, err)
		}
	}
}

func TestRejectsAPatternWithAnInvalidName(t *testing.T) {
	if _, err := ParsePattern("bloque 1", "25", "20", "10"); !errors.Is(err, ErrInvalidAlias) {
		t.Errorf("got %v, want ErrInvalidAlias", err)
	}
	if _, err := ParsePattern("", "25", "20", "10"); !errors.Is(err, ErrEmptyAlias) {
		t.Errorf("got %v, want ErrEmptyAlias", err)
	}
}

func TestAPatternRunsCalibre3DPatronWithItsRealSizes(t *testing.T) {
	pattern := Pattern{Name: "bloque", Length: 60, Width: 40.05, Thickness: 20}
	want := []string{"-m", "calibre3d", "patron", "bloque", "--real", "60", "40.05", "20"}
	if got := pattern.Args(); !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAPatternLeavesItsViewWithAPrefix(t *testing.T) {
	want := filepath.Join("p", "datos", "resultados", "patron_bloque_vista.png")
	if got := StepPattern.ResultFile("p", "bloque", "_vista"); got != want {
		t.Errorf("got %q", got)
	}
}

func TestParsesARealCalibre3DCalibration(t *testing.T) {
	calibration, err := ParseCalibration(readTestdata(t, "calibracion.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := Calibration{Patterns: 2, Correction: Correction{KXY: 1.0012, OffsetXY: 0.95, KZ: 1, OffsetZ: 0.015}}
	if calibration != want {
		t.Errorf("got %+v, want %+v", calibration, want)
	}
}

func TestACalibrationWithoutCorrectionIsAnError(t *testing.T) {
	for _, content := range []string{"{", `{"patrones": []}`} {
		if _, err := ParseCalibration([]byte(content)); err == nil {
			t.Errorf("%q: no error", content)
		}
	}
}
