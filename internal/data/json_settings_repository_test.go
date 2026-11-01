package data

import (
	"testing"

	"github.com/Itros97/script-launcher/internal/domain"
)

const settingsFile = "/config/settings.json"

func TestSettingsDefaultToTheDefaultProfileWhenNothingWasSaved(t *testing.T) {
	settings, err := NewJSONSettingsRepository(memoryStore{}, settingsFile).Load()
	if err != nil || settings.Profile != domain.DefaultProfileID {
		t.Fatalf("got %+v, %v", settings, err)
	}
}

func TestSavedProfileSurvivesANewRepository(t *testing.T) {
	store := memoryStore{}
	if err := NewJSONSettingsRepository(store, settingsFile).Save(Settings{Profile: domain.Calibre3DProfileID}); err != nil {
		t.Fatal(err)
	}
	settings, err := NewJSONSettingsRepository(store, settingsFile).Load()
	if err != nil || settings.Profile != domain.Calibre3DProfileID {
		t.Errorf("got %+v, %v", settings, err)
	}
}

func TestCorruptSettingsAreAnError(t *testing.T) {
	store := memoryStore{settingsFile: []byte("[oops")}
	if _, err := NewJSONSettingsRepository(store, settingsFile).Load(); err == nil {
		t.Fatal("expected an error for a corrupt settings file")
	}
}

func TestUpdateKeepsTheFieldsItDoesNotChange(t *testing.T) {
	repository := NewJSONSettingsRepository(memoryStore{}, settingsFile)
	if err := repository.Update(func(settings *Settings) { settings.Calibre3DProject = "/projects/calibre" }); err != nil {
		t.Fatal(err)
	}
	if err := repository.Update(func(settings *Settings) { settings.Profile = domain.Calibre3DProfileID }); err != nil {
		t.Fatal(err)
	}
	settings, err := repository.Load()
	if err != nil || settings.Profile != domain.Calibre3DProfileID || settings.Calibre3DProject != "/projects/calibre" {
		t.Errorf("got %+v, %v", settings, err)
	}
}
