package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestAcceptsLettersDigitsDashesAndUnderscores(t *testing.T) {
	for _, alias := range []string{"pieza1", "pendrive_lado-2", "piñón", "A"} {
		if err := ValidatePhotoAlias(alias); err != nil {
			t.Errorf("%q: unexpected error %v", alias, err)
		}
	}
}

func TestRejectsAnEmptyName(t *testing.T) {
	if err := ValidatePhotoAlias(""); !errors.Is(err, ErrEmptyAlias) {
		t.Errorf("got %v, want ErrEmptyAlias", err)
	}
}

func TestRejectsNamesThatAreNotSafeFileNames(t *testing.T) {
	for _, alias := range []string{"con espacio", "../fuera", "a/b", "punto.png", strings.Repeat("a", MaxAliasLength+1)} {
		if err := ValidatePhotoAlias(alias); !errors.Is(err, ErrInvalidAlias) {
			t.Errorf("%q: got %v, want ErrInvalidAlias", alias, err)
		}
	}
}

func TestPhotoPathIsInsideThePhotosFolder(t *testing.T) {
	if got := PhotoPath("/projects/calibre", "pieza1"); got != "/projects/calibre/datos/fotos/pieza1.png" {
		t.Errorf("got %s", got)
	}
}
