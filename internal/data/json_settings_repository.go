package data

// JSONSettingsRepository stores the settings as one JSON object in a file.
type JSONSettingsRepository struct {
	store FileStore
	path  string
}

// NewJSONSettingsRepository builds a repository that keeps the settings in path, through store.
func NewJSONSettingsRepository(store FileStore, path string) *JSONSettingsRepository {
	return &JSONSettingsRepository{store: store, path: path}
}

// Load returns the saved settings; fields missing from the file keep their defaults.
func (repository *JSONSettingsRepository) Load() (Settings, error) {
	settings := DefaultSettings()
	if err := readJSON(repository.store, repository.path, "settings", &settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

// Update applies change to the saved settings and saves the result.
func (repository *JSONSettingsRepository) Update(change func(settings *Settings)) error {
	settings, err := repository.Load()
	if err != nil {
		return err
	}
	change(&settings)
	return repository.Save(settings)
}

// Save replaces the settings file.
func (repository *JSONSettingsRepository) Save(settings Settings) error {
	return writeJSON(repository.store, repository.path, "settings", settings)
}
