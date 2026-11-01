package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// calibreDate is how calibre3d writes dates: local time, to the second, without a zone.
const calibreDate = "2006-01-02T15:04:05"

// Measurement is what calibre3d measured of a piece, after its calibration correction, in millimetres.
type Measurement struct {
	Name      string
	Date      time.Time
	Length    float64
	Width     float64
	Thickness float64
	Feret     float64
	Warnings  []string
}

// PlaneInfo is the reference plane calibre3d saved from the empty mat.
type PlaneInfo struct {
	Date         time.Time
	NoiseMM      float64
	CameraHeight float64
}

// measurementFile is the part of calibre3d's result file the launcher shows.
type measurementFile struct {
	Name   string `json:"nombre"`
	Date   string `json:"fecha"`
	Result *struct {
		Length    float64  `json:"largo_mm"`
		Width     float64  `json:"ancho_mm"`
		Thickness float64  `json:"espesor_mm"`
		Feret     float64  `json:"feret_max_mm"`
		Warnings  []string `json:"aviso"`
	} `json:"resultado"`
}

// planeFile is the part of calibre3d's plano.json the launcher shows.
type planeFile struct {
	Date         string   `json:"fecha"`
	NoiseMM      *float64 `json:"rms_mm"`
	CameraHeight float64  `json:"altura_camara_mm"`
}

// ParseMeasurement reads a calibre3d result file (datos/resultados/<name>.json).
func ParseMeasurement(content []byte) (Measurement, error) {
	var file measurementFile
	if err := json.Unmarshal(content, &file); err != nil {
		return Measurement{}, fmt.Errorf("reading the measurement: %w", err)
	}
	if file.Result == nil {
		return Measurement{}, fmt.Errorf("reading the measurement: it has no \"resultado\"")
	}
	date, _ := time.ParseInLocation(calibreDate, file.Date, time.Local) // a missing date is shown as unknown
	result := file.Result
	return Measurement{
		Name: file.Name, Date: date, Length: result.Length, Width: result.Width,
		Thickness: result.Thickness, Feret: result.Feret, Warnings: result.Warnings,
	}, nil
}

// ParsePlane reads calibre3d's plano.json.
func ParsePlane(content []byte) (PlaneInfo, error) {
	var file planeFile
	if err := json.Unmarshal(content, &file); err != nil {
		return PlaneInfo{}, fmt.Errorf("reading the plane: %w", err)
	}
	if file.NoiseMM == nil {
		return PlaneInfo{}, fmt.Errorf("reading the plane: it has no \"rms_mm\"")
	}
	date, _ := time.ParseInLocation(calibreDate, file.Date, time.Local)
	return PlaneInfo{Date: date, NoiseMM: *file.NoiseMM, CameraHeight: file.CameraHeight}, nil
}
