package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// General holds app-level (non-module) settings.
type General struct {
	RunAtStartup bool   `json:"runAtStartup"`
	Theme        string `json:"theme"` // system | light | dark
}

type persistedModule struct {
	Enabled  bool           `json:"enabled"`
	Settings map[string]any `json:"settings,omitempty"`
}

type persistedFile struct {
	General General                   `json:"general"`
	Modules map[string]persistedModule `json:"modules"`
}

// Store persists all app state in %LOCALAPPDATA%\PolyTools\settings.json.
type Store struct {
	path string
	data persistedFile
}

func NewStore() (*Store, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "."
	}
	dir = filepath.Join(dir, "PolyTools")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		path: filepath.Join(dir, "settings.json"),
		data: persistedFile{
			General: General{Theme: "system"},
			Modules: map[string]persistedModule{},
		},
	}
	_ = s.load()
	return s, nil
}

func (s *Store) load() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &s.data)
}

func (s *Store) Save() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o644)
}

func (s *Store) General() General { return s.data.General }

func (s *Store) SetGeneral(g General) { s.data.General = g }

func (s *Store) ModuleState(key string) (enabled bool, settings map[string]any, ok bool) {
	m, ok := s.data.Modules[key]
	return m.Enabled, m.Settings, ok
}

func (s *Store) SetModuleState(key string, enabled bool, settings map[string]any) {
	m := s.data.Modules[key]
	m.Enabled = enabled
	if settings != nil {
		if m.Settings == nil {
			m.Settings = map[string]any{}
		}
		for k, v := range settings {
			m.Settings[k] = v
		}
	}
	s.data.Modules[key] = m
}
