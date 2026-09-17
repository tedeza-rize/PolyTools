//go:build windows

package modules

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
)

// RegisterAll registers every module (implemented + coming soon).
// app is passed to modules that need runtime services like global shortcuts;
// win is the main window, for modules that must know their own HWND;
// tray + showMain let Minimize To Tray extend the tray menu.
func RegisterAll(reg *core.Registry, app *application.App, win *application.WebviewWindow, tray *application.SystemTray, showMain func()) {
	reg.Register(newAwake())
	reg.Register(newTextExtractor(app))
	reg.Register(newColorPicker(app))
	reg.Register(newScreenRuler(app))
	reg.Register(newBatteryManager())
	reg.Register(newScreensaver(app))
	reg.Register(newSystemInfo())

	reg.Register(newAlwaysOnTop(app))
	reg.Register(newGrabAndMove(app, win))
	reg.Register(newWindowMemory())
	reg.Register(newZoneLayouts(app, win))
	reg.Register(newWindowTransparency(app, win))
	reg.Register(newMinimizeToTray(app, tray, showMain))

	reg.Register(newBorderlessGaming(app))
	reg.Register(newFpsOverlay(app))
	reg.Register(newLosslessScaling(app))

	reg.Register(newKeyboardManager())
	reg.Register(newMouseUtilities(app))
	reg.Register(newHotkeyActions(app))
	reg.Register(newTextExpander())
	reg.Register(newPastePlain(app))

	reg.Register(newBatchRename())
	reg.Register(newQuickPeek(app))
	reg.Register(newContextMenu())

	reg.Register(newEnvVars())
	reg.Register(newHostsEditor())
	reg.Register(newAutomations(app))
}
