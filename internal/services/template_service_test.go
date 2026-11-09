package services

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
	"github.com/Itros97/script-launcher/internal/services/templates"
)

func TestTemplateBecomesTheRequirementsWhenThereAreNone(t *testing.T) {
	project := t.TempDir()
	path, err := SaveRequirementsTemplate(project)
	if err != nil || path != filepath.Join(project, domain.RequirementsFile) {
		t.Fatalf("got %s, %v", path, err)
	}
	if content, _ := os.ReadFile(path); !bytes.Equal(content, templates.Requirements) {
		t.Errorf("got %q", content)
	}
}

func TestTemplateNeverOverwritesExistingRequirements(t *testing.T) {
	project := t.TempDir()
	own := filepath.Join(project, domain.RequirementsFile)
	if err := os.WriteFile(own, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := SaveRequirementsTemplate(project)
	if err != nil || path != filepath.Join(project, RequirementsTemplateCopy) {
		t.Fatalf("got %s, %v", path, err)
	}
	if content, _ := os.ReadFile(own); string(content) != "mine\n" {
		t.Errorf("requirements.txt was changed: %q", content)
	}
	if _, err := SaveRequirementsTemplate(project); !errors.Is(err, ErrTemplateExists) {
		t.Errorf("third save: got %v, want ErrTemplateExists", err)
	}
}

func TestTemplateNeedsAnExistingProject(t *testing.T) {
	if _, err := SaveRequirementsTemplate(filepath.Join(t.TempDir(), "gone")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("got %v, want ErrNotExist", err)
	}
}

func TestTemplateListsTheCameraLibrary(t *testing.T) {
	if !bytes.Contains(templates.Requirements, []byte("\npyrealsense2>=")) {
		t.Error("the template must install pyrealsense2, which the photo capture needs")
	}
}

func TestTemplateInstallsTheCameraLibraryOnLinuxAndOnMac(t *testing.T) {
	for _, line := range []string{`pyrealsense2>=2.55; sys_platform != "darwin"`, `pyrealsense2-macosx>=2.56; sys_platform == "darwin" and platform_machine == "arm64"`} {
		if !bytes.Contains(templates.Requirements, []byte(line+"\n")) {
			t.Errorf("the template lacks %q", line)
		}
	}
}
