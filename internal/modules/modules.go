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

	reg.Register(newAlwaysOnTop(app))
	reg.Register(newGrabAndMove())
	reg.Register(newWindowMemory())
	reg.Register(newZoneLayouts())

	reg.Register(newKeyboardManager())
	reg.Register(newMouseUtilities())

	reg.Register(newBatchRename())
	reg.Register(newQuickPeek())

	reg.Register(newEnvVars())
	reg.Register(newHostsEditor())
}
