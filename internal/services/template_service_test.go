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

func TestTemplateIsSavedUnderItsSystemsName(t *testing.T) {
	project := t.TempDir()
	for _, system := range domain.Platforms {
		path, err := SaveRequirementsTemplate(project, system)
		if err != nil || path != filepath.Join(project, system.RequirementsFile()) {
			t.Fatalf("%s: got %s, %v", system, path, err)
		}
		if content, _ := os.ReadFile(path); !bytes.Equal(content, templates.Requirements(system)) {
			t.Errorf("%s: got %q", system, content)
		}
	}
}

func TestTemplateNeverOverwritesAFile(t *testing.T) {
	project := t.TempDir()
	own := filepath.Join(project, domain.MacOS.RequirementsFile())
	if err := os.WriteFile(own, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveRequirementsTemplate(project, domain.MacOS); !errors.Is(err, ErrTemplateExists) {
		t.Errorf("got %v, want ErrTemplateExists", err)
	}
	if content, _ := os.ReadFile(own); string(content) != "mine\n" {
		t.Errorf("the file was changed: %q", content)
	}
}

func TestTemplateNeedsAnExistingProject(t *testing.T) {
	if _, err := SaveRequirementsTemplate(filepath.Join(t.TempDir(), "gone"), domain.Linux); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("got %v, want ErrNotExist", err)
	}
}

func TestEachTemplateInstallsItsOwnCameraLibrary(t *testing.T) {
	cases := map[domain.Platform]struct{ want, never string }{
		domain.Linux: {"\npyrealsense2>=2.55\n", "pyrealsense2-macosx"},
		domain.MacOS: {"\npyrealsense2-macosx>=2.56; platform_machine == \"arm64\"\n", "\npyrealsense2>="},
	}
	for system, c := range cases {
		content := templates.Requirements(system)
		if !bytes.Contains(content, []byte(c.want)) {
			t.Errorf("%s template lacks %q", system, c.want)
		}
		for _, line := range bytes.Split(content, []byte("\n")) {
			if !bytes.HasPrefix(line, []byte("#")) && bytes.Contains(append([]byte("\n"), line...), []byte(c.never)) {
				t.Errorf("%s template must not install %q: %q", system, c.never, line)
			}
		}
	}
}
