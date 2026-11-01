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

// ProjectMarker is the file that tells the calibre3d project's folder apart.
const ProjectMarker = "prueba.sh"

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
// says what to do: the project's own package is missing (from a wrong folder, or from the right one
// when hasMarker says ProjectMarker is there), a library is missing, or another error.
func DescribeImportFailure(module, stderr string, hasMarker bool) string {
	if match := missingModule.FindStringSubmatch(stderr); match != nil {
		missing := strings.Split(match[1], ".")[0]
		if missing == module && hasMarker {
			return "Esta carpeta tiene " + ProjectMarker + " pero le falta la carpeta " + module + "/ con el programa. Copia el proyecto completo."
		}
		if missing == module {
			return "En esta carpeta no está el programa " + module + ". Elige la carpeta del proyecto (la que tiene " + ProjectMarker + ")."
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
