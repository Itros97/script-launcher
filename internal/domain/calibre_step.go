package domain

import "path/filepath"

// CalibreModule is the Python module of the calibre3d project.
const CalibreModule = "calibre3d"

// CalibreStep is one of the calibre3d commands the launcher runs.
type CalibreStep int

const (
	StepPlane   CalibreStep = iota // save the empty mat as the reference plane
	StepCenter                     // say how far a coin on the cross is from the camera's axis
	StepMeasure                    // measure a named piece
	StepDemo                       // a full measurement of a simulated piece, without the camera
	StepPattern                    // measure a piece of known size and update the calibration
)

// calibre3d's folders and names, relative to the project.
const (
	calibreDataDir = "datos"
	calibreDemoDir = "demo"
	resultsDir     = "resultados"
	demoResult     = "demo_canto"
	centerResult   = "centrado"
	patternPrefix  = "patron_"
)

// Args are the arguments that run the step with `python`; name is used by StepMeasure and StepPattern
// only, and StepPattern also needs the real sizes that Pattern.Args adds.
func (step CalibreStep) Args(name string) []string {
	switch step {
	case StepPlane:
		return []string{"-m", CalibreModule, "plano"}
	case StepCenter:
		return []string{"-m", CalibreModule, "centrar"}
	case StepMeasure:
		return []string{"-m", CalibreModule, "medir", name}
	case StepPattern:
		return []string{"-m", CalibreModule, "patron", name}
	default:
		return []string{"-m", CalibreModule, "demo"}
	}
}

// Title names the step in the activity panel.
func (step CalibreStep) Title(name string) string {
	switch step {
	case StepPlane:
		return "Plano"
	case StepCenter:
		return "Centrado"
	case StepMeasure:
		return "Medida «" + name + "»"
	case StepPattern:
		return "Patrón «" + name + "»"
	default:
		return "Demo"
	}
}

// ResultName is the name calibre3d gives the step's result files.
func (step CalibreStep) ResultName(name string) string {
	switch step {
	case StepCenter:
		return centerResult
	case StepDemo:
		return demoResult
	case StepPattern:
		return patternPrefix + name
	default:
		return name
	}
}

// ResultFile is where a step leaves a result file with this suffix ("" for the .json).
func (step CalibreStep) ResultFile(project, name, suffix string) string {
	folder := calibreDataDir
	if step == StepDemo {
		folder = calibreDemoDir
	}
	extension := ".json"
	if suffix != "" {
		extension = suffix + ".png"
	}
	return filepath.Join(project, folder, resultsDir, step.ResultName(name)+extension)
}

// PlaneFile is where calibre3d keeps the reference plane.
func PlaneFile(project string) string { return filepath.Join(project, calibreDataDir, "plano.json") }

// ResultsDir is where calibre3d leaves the result of every measured piece.
func ResultsDir(project string) string {
	return filepath.Join(project, calibreDataDir, resultsDir)
}

// IsPieceResult tells whether a file in ResultsDir is a measured piece's result, not the centring's.
func IsPieceResult(fileName string) bool {
	return filepath.Ext(fileName) == ".json" && fileName != centerResult+".json"
}

// PieceFiles are every file calibre3d leaves for a measured piece: its result and its two pictures.
func PieceFiles(project, name string) []string {
	return []string{
		StepMeasure.ResultFile(project, name, ""),
		StepMeasure.ResultFile(project, name, "_mapa"),
		StepMeasure.ResultFile(project, name, "_vista"),
	}
}
