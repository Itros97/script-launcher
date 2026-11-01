package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"unicode"
)

// PhotosDir is where photos are saved, relative to the project.
const PhotosDir = "datos/fotos"

// MaxAliasLength is the longest name a photo can have, in characters.
const MaxAliasLength = 64

var (
	// ErrEmptyAlias is returned for a photo name with nothing in it.
	ErrEmptyAlias = errors.New("empty photo name")
	// ErrInvalidAlias is returned for a photo name with characters that do not belong in a file name.
	ErrInvalidAlias = errors.New("invalid photo name")
	// ErrAliasTaken is returned when a photo with that name already exists.
	ErrAliasTaken = errors.New("photo name already used")
)

// ValidatePhotoAlias accepts names made of letters, digits, '-' and '_', up to MaxAliasLength characters.
func ValidatePhotoAlias(alias string) error {
	if alias == "" {
		return ErrEmptyAlias
	}
	runes := []rune(alias)
	if len(runes) > MaxAliasLength {
		return fmt.Errorf("%w: longer than %d characters", ErrInvalidAlias, MaxAliasLength)
	}
	for _, char := range runes {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '-' && char != '_' {
			return fmt.Errorf("%w: %q is not allowed", ErrInvalidAlias, char)
		}
	}
	return nil
}

// PhotoPath is where the photo with this name is saved inside the project.
func PhotoPath(project, alias string) string {
	return filepath.Join(project, PhotosDir, alias+".png")
}
