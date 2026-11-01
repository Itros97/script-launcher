package domain

import (
	"regexp"
	"strings"
)

// EnvCheck is an environment's status and, when it is not ready, why, in words for the user.
type EnvCheck struct {
	Status EnvStatus
	Reason string
}

// missingModule finds the module Python could not import.
var missingModule = regexp.MustCompile(`No module named '([^']+)'`)

// pipNames are the pip packages of imports whose name differs from the package's.
var pipNames = map[string]string{
	"cv2":     "opencv-python",
	"PIL":     "Pillow",
	"sklearn": "scikit-learn",
	"yaml":    "PyYAML",
	"skimage": "scikit-image",
}

// DescribeImportFailure turns what Python printed when `import module` failed into a sentence that
// says what to do: the project's own package is missing, a library is missing, or another error.
func DescribeImportFailure(module, stderr string) string {
	if match := missingModule.FindStringSubmatch(stderr); match != nil {
		missing := strings.Split(match[1], ".")[0]
		if missing == module {
			return "En esta carpeta no está el programa " + module + ". Elige la carpeta del proyecto (la que tiene prueba.sh)."
		}
		return "Falta la librería " + missing + ": añade " + pipName(missing) + " a " + RequirementsFile + " y vuelve a preparar el entorno."
	}
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	if last := strings.TrimSpace(lines[len(lines)-1]); last != "" {
		return "Error al cargar " + module + ": " + last
	}
	return module + " no se puede cargar en el entorno."
}

// pipName is the pip package that provides an import.
func pipName(module string) string {
	if name, ok := pipNames[module]; ok {
		return name
	}
	return module
}
