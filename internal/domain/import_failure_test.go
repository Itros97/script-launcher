package domain

import (
	"strings"
	"testing"
)

const traceback = "Traceback (most recent call last):\n  File \"<string>\", line 1, in <module>\n"

func TestMissingProjectPackageAsksForTheRightFolder(t *testing.T) {
	reason := DescribeImportFailure("calibre3d", traceback+"ModuleNotFoundError: No module named 'calibre3d'\n")
	if reason != "En esta carpeta no está el programa calibre3d. Elige la carpeta del proyecto (la que tiene prueba.sh)." {
		t.Errorf("got %q", reason)
	}
}

func TestMissingLibraryNamesThePipPackage(t *testing.T) {
	reason := DescribeImportFailure("calibre3d", traceback+"ModuleNotFoundError: No module named 'cv2'\n")
	if !strings.Contains(reason, "Falta la librería cv2: añade opencv-python a requirements.txt") {
		t.Errorf("got %q", reason)
	}
}

func TestMissingSubmoduleNamesItsTopPackage(t *testing.T) {
	reason := DescribeImportFailure("calibre3d", "ModuleNotFoundError: No module named 'scipy.spatial'\n")
	if !strings.Contains(reason, "Falta la librería scipy: añade scipy") {
		t.Errorf("got %q", reason)
	}
}

func TestOtherErrorsKeepPythonsLastLine(t *testing.T) {
	reason := DescribeImportFailure("calibre3d", traceback+"SyntaxError: invalid syntax\n")
	if reason != "Error al cargar calibre3d: SyntaxError: invalid syntax" {
		t.Errorf("got %q", reason)
	}
}

func TestNoOutputStillSaysSomething(t *testing.T) {
	if reason := DescribeImportFailure("calibre3d", ""); reason != "calibre3d no se puede cargar en el entorno." {
		t.Errorf("got %q", reason)
	}
}
