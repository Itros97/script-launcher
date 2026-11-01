package profiles

import (
	"slices"
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
)

func TestIndexOfFindsAKnownProfile(t *testing.T) {
	if index := NewRegistry(Default{}, &Calibre3D{}).IndexOf(domain.Calibre3DProfileID); index != 1 {
		t.Errorf("got %d, want 1", index)
	}
}

func TestIndexOfFallsBackToTheFirstProfileWhenUnknown(t *testing.T) {
	if index := NewRegistry(Default{}, &Calibre3D{}).IndexOf("gone"); index != 0 {
		t.Errorf("got %d, want 0", index)
	}
}

func TestNamesFollowTheSelectorOrder(t *testing.T) {
	if names := NewRegistry(Default{}, &Calibre3D{}).Names(); !slices.Equal(names, []string{"General", "Calibre3D"}) {
		t.Errorf("got %v", names)
	}
}
