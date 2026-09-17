//go:build windows

package modules

import (
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// Shared management of the fullscreen tool overlay window used by
// Screen Ruler (measure) and Text Extractor (region pick → OCR).

var (
	toolMu  sync.Mutex
	toolWin *application.WebviewWindow
)

// openToolWindow shows a fullscreen translucent overlay window loading the
// given route ("/?page=..."), positioned on the monitor under the cursor.
func openToolWindow(app *application.App, route string) *application.WebviewWindow {
	toolMu.Lock()
	defer toolMu.Unlock()
	if toolWin != nil {
		toolWin.Close()
		toolWin = nil
	}
	x, y := win32.CursorPos()
	mr := win32.MonitorRectAt(x, y)
	w := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "PolyTools",
		Frameless:      true,
		AlwaysOnTop:    true,
		BackgroundType: application.BackgroundTypeTranslucent,
		URL:            route,
		X:              int(mr.Left),
		Y:              int(mr.Top),
		Width:          int(mr.Width()),
		Height:         int(mr.Height()),
	})
	toolWin = w
	return w
}

// CloseToolWindow hides the current tool overlay (called from the page).
func CloseToolWindow() {
	toolMu.Lock()
	defer toolMu.Unlock()
	if toolWin != nil {
		toolWin.Close()
		toolWin = nil
	}
}

// --- Screen Ruler ---

func newScreenRuler(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "screen-ruler",
		Name:        "Screen Ruler",
		Description: "Measure pixel distances on screen.",
		Icon:        "Ruler",
		Category:    core.CategorySystem,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation shortcut", Type: core.SettingShortcut, Value: "super+shift+m"},
		},
	})

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		openToolWindow(app, "/?page=ruler")
		log.Printf("[ruler] opened")
	})

	return m.WithHandlers(hk.register, hk.unregister).
		WithSettingHandler(func(key string, _ any) error {
			if key != "hotkey" || !m.Info().Enabled {
				return nil
			}
			return hk.rebind()
		})
}

// --- Text Extractor ---

func newTextExtractor(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "text-extractor",
		Name:        "Text Extractor",
		Description: "Copy text from anywhere on the screen using OCR.",
		Icon:        "TextFont",
		Category:    core.CategorySystem,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation shortcut", Type: core.SettingShortcut, Value: "super+shift+t"},
		},
	})

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		openToolWindow(app, "/?page=extract")
	})

	return m.WithHandlers(hk.register, hk.unregister).
		WithSettingHandler(func(key string, _ any) error {
			if key != "hotkey" || !m.Info().Enabled {
				return nil
			}
			return hk.rebind()
		})
}

// --- Quick Peek: hotkey grabs the Explorer selection (via Ctrl+C /
// clipboard files) or clipboard text and previews it in a window. ---

func newQuickPeek(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "quick-peek",
		Name:        "Quick Peek",
		Description: "Preview file content without opening the application.",
		Icon:        "Eye",
		Category:    core.CategoryFiles,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation key", Type: core.SettingShortcut, Value: "ctrl+space"},
		},
	})

	var peekWin *application.WebviewWindow

	closePeek := func() {
		if peekWin != nil {
			peekWin.Close()
			peekWin = nil
		}
	}

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		if peekWin != nil {
			closePeek()
			return
		}
		// Grab Explorer selection: send Ctrl+C, then read CF_HDROP.
		path := ""
		if exe := win32.WindowExe(win32.ForegroundWindow()); exe == "explorer.exe" {
			win32.SendKeys(0x11, 0x43) // Ctrl+C
			time.Sleep(120 * time.Millisecond)
			if files := win32.ClipboardFiles(); len(files) > 0 {
				path = files[0]
			}
		}
		route := "/?page=peek"
		if path != "" {
			route += "&path=" + url.QueryEscape(path)
		}
		peekWin = app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:       "Quick Peek",
			Width:       860,
			Height:      620,
			AlwaysOnTop: true,
			URL:         route,
		})
	})

	return m.WithHandlers(hk.register, func() error {
		closePeek()
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}
