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

// hotkeyCtl tracks the combo a module has registered so it can be
// re-registered cleanly when the user edits the shortcut (the store only
// knows the new value, so the previous combo must be kept here).
type hotkeyCtl struct {
	app     *application.App
	comboOf func() string // current combo, read from the module's settings
	onFire  func()
	last    string // combo actually registered
}

func newHotkeyCtl(app *application.App, comboOf func() string, onFire func()) *hotkeyCtl {
	return &hotkeyCtl{app: app, comboOf: comboOf, onFire: onFire}
}

func (h *hotkeyCtl) register() error {
	h.last = h.comboOf()
	return h.app.GlobalShortcut.Register(h.last, h.onFire)
}

func (h *hotkeyCtl) unregister() error {
	return h.app.GlobalShortcut.Unregister(h.last)
}

// rebind swaps the previously registered combo for the current one.
func (h *hotkeyCtl) rebind() error {
	_ = h.app.GlobalShortcut.Unregister(h.last)
	return h.register()
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

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		hwnd := win32.ForegroundWindow()
		log.Printf("[always-on-top] hotkey fired, hwnd=%d topmost=%v", hwnd, win32.IsTopMost(hwnd))
		if hwnd != 0 {
			win32.SetTopMost(hwnd, !win32.IsTopMost(hwnd))
		}
	})

	return m.WithHandlers(hk.register, hk.unregister).
		WithSettingHandler(func(key string, value any) error {
			if key != "hotkey" || !m.Info().Enabled {
				return nil
			}
			return hk.rebind()
		})
}
