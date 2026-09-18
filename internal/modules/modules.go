//go:build windows

package modules

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
)

// sysProcAttrHide prevents a console window flashing for spawned commands.
var sysProcAttrHide = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW

// stateDir returns %LOCALAPPDATA%\PolyTools (same place as settings.json).
func stateDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "."
	}
	d := filepath.Join(dir, "PolyTools")
	_ = os.MkdirAll(d, 0o755)
	return d
}

// RegisterAll registers every module (implemented + coming soon).
// app is passed to modules that need runtime services like global shortcuts.
func RegisterAll(reg *core.Registry, app *application.App) {
	reg.Register(newAwake())
	reg.Register(newBatteryManager())
	reg.Register(newScreensaver(app))

	reg.Register(newAlwaysOnTop(app))

	reg.Register(newBorderlessGaming(app))
	reg.Register(newFpsOverlay(app))
	reg.Register(newLosslessScaling(app))

	reg.Register(newHotkeyActions(app))

	reg.Register(newContextMenu())

	reg.Register(newAutomations(app))
}
