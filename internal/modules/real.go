//go:build windows

package modules

import (
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Awake: prevent sleep via SetThreadExecutionState ---

func newAwake() *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "awake",
		Name:        "Awake",
		Description: "Keep the computer awake without changing power & sleep settings.",
		Icon:        "WeatherMoon",
		Category:    core.CategorySystem,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key:         "keepDisplayOn",
				Label:       "Keep screen on",
				Description: "Also prevent the display from turning off.",
				Type:        core.SettingToggle,
				Value:       true,
			},
		},
	})
	return m.WithHandlers(
		func() error {
			win32.KeepAwake(m.SettingBool("keepDisplayOn"))
			return nil
		},
		func() error {
			win32.ClearAwake()
			return nil
		},
	).WithSettingHandler(func(key string, _ any) error {
		if key == "keepDisplayOn" && m.Info().Enabled {
			win32.KeepAwake(m.SettingBool("keepDisplayOn"))
		}
		return nil
	})
}

// --- Always On Top: Win+Ctrl+T pins the foreground window ---

func newAlwaysOnTop(app *application.App) *core.BaseModule {
	const defaultHotkey = "ctrl+alt+t"

	m := core.NewModule(core.Info{
		Key:         "always-on-top",
		Name:        "Always On Top",
		Description: "Pin a window above all others with a quick shortcut.",
		Icon:        "Pin",
		Category:    core.CategoryWindowing,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key:         "hotkey",
				Label:       "Activation shortcut",
				Description: "Press on any window to pin / unpin it.",
				Type:        core.SettingShortcut,
				Value:       defaultHotkey,
			},
			{
				Key:         "showBorder",
				Label:       "Show border",
				Description: "Draw a highlight around pinned windows.",
				Type:        core.SettingToggle,
				Value:       true,
			},
			{
				Key:         "playSound",
				Label:       "Play sound",
				Description: "Play a sound when a window is pinned.",
				Type:        core.SettingToggle,
				Value:       false,
			},
		},
	})

	register := func() error {
		return app.GlobalShortcut.Register(m.SettingString("hotkey"), func() {
			hwnd := win32.ForegroundWindow()
			log.Printf("[always-on-top] hotkey fired, hwnd=%d topmost=%v", hwnd, win32.IsTopMost(hwnd))
			if hwnd != 0 {
				win32.SetTopMost(hwnd, !win32.IsTopMost(hwnd))
			}
		})
	}

	return m.WithHandlers(
		register,
		func() error {
			return app.GlobalShortcut.Unregister(m.SettingString("hotkey"))
		},
	).WithSettingHandler(func(key string, value any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		old := value // value already applied; re-register with the new combo
		_ = old
		_ = app.GlobalShortcut.UnregisterAll()
		return register()
	})
}
