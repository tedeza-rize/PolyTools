//go:build windows

package services

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
)

// PolyToolsService is bound to the frontend. It exposes module state and
// general settings, plus window controls for the custom titlebar / tray.
type PolyToolsService struct {
	app *application.App
	reg *core.Registry
	win *application.WebviewWindow
}

func (s *PolyToolsService) Attach(app *application.App, reg *core.Registry, win *application.WebviewWindow) {
	s.app = app
	s.reg = reg
	s.win = win
}

// --- modules ---

func (s *PolyToolsService) Modules() []core.Info {
	return s.reg.Modules()
}

func (s *PolyToolsService) SetModuleEnabled(key string, enabled bool) error {
	return s.reg.SetEnabled(key, enabled)
}

func (s *PolyToolsService) UpdateModuleSetting(key, field string, value any) error {
	return s.reg.UpdateSetting(key, field, value)
}

// --- general settings ---

func (s *PolyToolsService) General() core.General {
	g := s.reg.General()
	if enabled, err := s.app.Autostart.IsEnabled(); err == nil {
		g.RunAtStartup = enabled
	}
	return g
}

func (s *PolyToolsService) SetRunAtStartup(enabled bool) error {
	var err error
	if enabled {
		err = s.app.Autostart.Enable()
	} else {
		err = s.app.Autostart.Disable()
	}
	if err != nil {
		return err
	}
	g := s.reg.General()
	g.RunAtStartup = enabled
	s.reg.SetGeneral(g)
	return nil
}

func (s *PolyToolsService) SetTheme(theme string) {
	g := s.reg.General()
	g.Theme = theme
	s.reg.SetGeneral(g)
}

func (s *PolyToolsService) SetLanguage(lang string) {
	g := s.reg.General()
	g.Language = lang
	s.reg.SetGeneral(g)
}

// --- window controls for the custom titlebar / tray ---

func (s *PolyToolsService) ShowWindow() {
	s.win.Show()
	s.win.UnMinimise()
}

func (s *PolyToolsService) HideWindow() {
	s.win.Hide()
}

func (s *PolyToolsService) Minimise() {
	s.win.Minimise()
}

func (s *PolyToolsService) ToggleMaximise() {
	s.win.ToggleMaximise()
}

func (s *PolyToolsService) Quit() {
	s.app.Quit()
}
