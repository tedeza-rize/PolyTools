//go:build windows

package services

import (
	"fmt"
	"math"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows/registry"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// PolyToolsService is bound to the frontend. It exposes module state and
// general settings, plus window controls for the custom titlebar / tray.
type PolyToolsService struct {
	app *application.App
	reg *core.Registry
	win *application.WebviewWindow

	lastCPUIdle  uint64
	lastCPUTotal uint64
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

// --- system info ---

// SystemStats is the live snapshot shown on the System Info module page.
type SystemStats struct {
	OS             string  `json:"os"`
	CPUName        string  `json:"cpuName"`
	CPUPercent     float64 `json:"cpuPercent"`
	MemTotalBytes  uint64  `json:"memTotalBytes"`
	MemUsedBytes   uint64  `json:"memUsedBytes"`
	MemLoadPercent uint32  `json:"memLoadPercent"`
	GPUName        string  `json:"gpuName"`
	UptimeSeconds  uint64  `json:"uptimeSeconds"`
	HasBattery     bool    `json:"hasBattery"`
	BatteryPercent int     `json:"batteryPercent"`
	BatteryOnAC    bool    `json:"batteryOnAC"`
}

func regString(path, name string) string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	v, _, err := k.GetStringValue(name)
	if err != nil {
		return ""
	}
	return v
}

func (s *PolyToolsService) SystemInfo() SystemStats {
	st := SystemStats{
		GPUName:       win32.PrimaryGPUName(),
		UptimeSeconds: win32.UptimeSeconds(),
	}

	cv := regString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "ProductName")
	dv := regString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "DisplayVersion")
	bd := regString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "CurrentBuild")
	if dv != "" || bd != "" {
		cv = fmt.Sprintf("%s %s (Build %s)", cv, dv, bd)
	}
	st.OS = cv
	st.CPUName = regString(`HARDWARE\DESCRIPTION\System\CentralProcessor\0`, "ProcessorNameString")

	total, avail, load := win32.MemoryStatus()
	st.MemTotalBytes, st.MemUsedBytes, st.MemLoadPercent = total, total-avail, load
	st.BatteryOnAC, st.BatteryPercent, st.HasBattery = win32.PowerStatus()

	// CPU% = delta of (kernel+user) minus idle between calls.
	idle, kern, usr := win32.SystemTimes()
	tot := kern + usr
	if s.lastCPUTotal == 0 {
		time.Sleep(150 * time.Millisecond)
		s.lastCPUIdle, s.lastCPUTotal = idle, tot
		idle, kern, usr = win32.SystemTimes()
		tot = kern + usr
	}
	if dTot := tot - s.lastCPUTotal; dTot > 0 {
		dIdle := idle - s.lastCPUIdle
		st.CPUPercent = math.Round((1-float64(dIdle)/float64(dTot))*1000) / 10
	}
	s.lastCPUIdle, s.lastCPUTotal = idle, tot
	return st
}
