package domain

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func TestParsesARealCalibre3DResult(t *testing.T) {
	measurement, err := ParseMeasurement(readTestdata(t, "medida_demo.json"))
	if err != nil {
		t.Fatal(err)
	}
	if measurement.Name != "demo_canto" || measurement.Length != 59.97 || measurement.Width != 40.0476 ||
		measurement.Thickness != 19.9875 || measurement.Feret != 60.0375 || measurement.Date.IsZero() {
		t.Errorf("got %+v", measurement)
	}
}

func TestParsesWarnings(t *testing.T) {
	content := []byte(`{"nombre": "p", "resultado": {"largo_mm": 1, "aviso": ["Pieza lejos del centro"]}}`)
	measurement, err := ParseMeasurement(content)
	if err != nil || !slices.Equal(measurement.Warnings, []string{"Pieza lejos del centro"}) {
		t.Errorf("got %+v, %v", measurement, err)
	}
}

func TestRejectsAResultWithoutMeasurements(t *testing.T) {
	for _, content := range []string{"{roto", `{"nombre": "p"}`} {
		if _, err := ParseMeasurement([]byte(content)); err == nil {
			t.Errorf("%q: expected an error", content)
		}
	}
}

func TestParsesARealCalibre3DPlane(t *testing.T) {
	plane, err := ParsePlane(readTestdata(t, "plano_demo.json"))
	if err != nil || plane.NoiseMM < 0.05 || plane.NoiseMM > 0.051 || plane.CameraHeight < 132 || plane.Date.IsZero() {
		t.Errorf("got %+v, %v", plane, err)
	}
	if _, err := ParsePlane([]byte(`{"fecha": "x"}`)); err == nil {
		t.Error("a plane without noise should be an error")
	}
}

func TestStepsRunTheMatchingCommand(t *testing.T) {
	cases := map[CalibreStep][]string{
		StepPlane:   {"-m", "calibre3d", "plano"},
		StepCenter:  {"-m", "calibre3d", "centrar"},
		StepMeasure: {"-m", "calibre3d", "medir", "pieza1"},
		StepDemo:    {"-m", "calibre3d", "demo"},
	}
	for step, want := range cases {
		if got := step.Args("pieza1"); !slices.Equal(got, want) {
			t.Errorf("step %d: got %v, want %v", step, got, want)
		}
	}
}

func TestStepsFindTheirResultFiles(t *testing.T) {
	cases := map[string]string{
		StepMeasure.ResultFile("/p", "pieza1", ""):      "/p/datos/resultados/pieza1.json",
		StepMeasure.ResultFile("/p", "pieza1", "_mapa"): "/p/datos/resultados/pieza1_mapa.png",
		StepCenter.ResultFile("/p", "", "_vista"):       "/p/datos/resultados/centrado_vista.png",
		StepDemo.ResultFile("/p", "", ""):               "/p/demo/resultados/demo_canto.json",
		PlaneFile("/p"):                                 "/p/datos/plano.json",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}
