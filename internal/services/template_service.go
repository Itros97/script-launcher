package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/templates"
)

// RequirementsTemplateCopy is the name the template is saved under when the project already has a requirements.txt.
const RequirementsTemplateCopy = "requirements.plantilla.txt"

// SaveRequirementsTemplate writes the requirements template into the project and returns where: as
// requirements.txt when there is none, and beside it otherwise. It never overwrites a file.
func SaveRequirementsTemplate(project string) (string, error) {
	if !isDir(project) {
		return "", fmt.Errorf("cannot save the template in %s: %w", project, os.ErrNotExist)
	}
	for _, name := range []string{domain.RequirementsFile, RequirementsTemplateCopy} {
		path := filepath.Join(project, name)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("saving the template as %s: %w", path, err)
		}
		_, writeErr := file.Write(templates.Requirements)
		if closeErr := file.Close(); writeErr == nil {
			writeErr = closeErr
		}
		if writeErr != nil {
			return "", fmt.Errorf("writing the template to %s: %w", path, writeErr)
		}
		return path, nil
	}
	return "", fmt.Errorf("%w: %s and %s", ErrTemplateExists, domain.RequirementsFile, RequirementsTemplateCopy)
}
