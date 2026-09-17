//go:build windows

package modules

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
)

// RegisterAll registers every module (implemented + coming soon).
// app is passed to modules that need runtime services like global shortcuts.
func RegisterAll(reg *core.Registry, app *application.App) {
	reg.Register(newAwake())
	reg.Register(newTextExtractor())
	reg.Register(newColorPicker())
	reg.Register(newScreenRuler())
	reg.Register(newBatteryManager())
	reg.Register(newScreensaver())
	reg.Register(newSystemInfo())

	reg.Register(newAlwaysOnTop(app))
	reg.Register(newGrabAndMove())
	reg.Register(newWindowMemory())
	reg.Register(newZoneLayouts())
	reg.Register(newWindowTransparency())
	reg.Register(newMinimizeToTray())

	reg.Register(newBorderlessGaming())
	reg.Register(newFpsOverlay())
	reg.Register(newLosslessScaling())

	reg.Register(newKeyboardManager())
	reg.Register(newMouseUtilities())
	reg.Register(newHotkeyActions())
	reg.Register(newTextExpander())

	reg.Register(newBatchRename())
	reg.Register(newQuickPeek())
	reg.Register(newContextMenu())

	reg.Register(newEnvVars())
	reg.Register(newHostsEditor())
	reg.Register(newAutomations())
}
