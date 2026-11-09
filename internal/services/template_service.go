package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/templates"
)

// SaveRequirementsTemplate writes system's requirements template into the project under that
// system's own name, such as requirements-macos.txt, and returns where. It never overwrites a file.
func SaveRequirementsTemplate(project string, system domain.Platform) (string, error) {
	if !isDir(project) {
		return "", fmt.Errorf("cannot save the template in %s: %w", project, os.ErrNotExist)
	}
	path := filepath.Join(project, system.RequirementsFile())
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if os.IsExist(err) {
		return "", fmt.Errorf("%w: %s", ErrTemplateExists, path)
	}
	if err != nil {
		return "", fmt.Errorf("saving the template as %s: %w", path, err)
	}
	_, writeErr := file.Write(templates.Requirements(system))
	if closeErr := file.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		return "", fmt.Errorf("writing the template to %s: %w", path, writeErr)
	}
	return path, nil
}
