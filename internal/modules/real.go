//go:build windows

package modules

import (
	"fmt"
	"log"
	"math"

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

// --- Window Transparency: hotkey toggles see-through on the foreground window ---

func newWindowTransparency(app *application.App, win *application.WebviewWindow) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "window-transparency",
		Name:        "Window Transparency",
		Description: "Make any window see-through with a shortcut.",
		Icon:        "SquareHint",
		Category:    core.CategoryWindowing,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle transparency", Type: core.SettingShortcut, Value: "super+shift+o"},
			{Key: "opacity", Label: "Opacity", Type: core.SettingSlider, Value: 80.0, Min: f64(10), Max: f64(100), Step: f64(5)},
			{Key: "excludeOwn", Label: "Skip PolyTools window", Type: core.SettingToggle, Value: true},
		},
	})

	// hwnd -> whether we added WS_EX_LAYERED ourselves
	applied := map[uintptr]bool{}
	var own uintptr // resolved lazily; the handle is valid only after Run

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		hwnd := win32.ForegroundWindow()
		if hwnd == 0 {
			return
		}
		if own == 0 && win != nil {
			own = uintptr(win.NativeWindow())
		}
		if m.SettingBool("excludeOwn") && hwnd == own {
			return
		}
		if added, ok := applied[hwnd]; ok {
			win32.ClearTransparency(hwnd, added)
			delete(applied, hwnd)
			log.Printf("[window-transparency] hwnd=%d restored", hwnd)
			return
		}
		applied[hwnd] = !win32.IsLayered(hwnd)
		pct := int(m.SettingFloat("opacity"))
		win32.SetTransparency(hwnd, pct)
		log.Printf("[window-transparency] hwnd=%d opacity=%d%%", hwnd, pct)
	})

	restoreAll := func() {
		for hwnd, added := range applied {
			win32.ClearTransparency(hwnd, added)
		}
		applied = map[uintptr]bool{}
	}

	return m.WithHandlers(hk.register, func() error {
		restoreAll()
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}

// --- Color Picker: hotkey copies the color under the cursor ---

func newColorPicker(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "color-picker",
		Name:        "Color Picker",
		Description: "Pick colors from anywhere on the screen.",
		Icon:        "Color",
		Category:    core.CategorySystem,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation shortcut", Type: core.SettingShortcut, Value: "super+shift+c"},
			{
				Key: "format", Label: "Color format", Type: core.SettingSelect, Value: "hex",
				Options: []core.SelectOption{
					{Value: "hex", Label: "HEX"}, {Value: "rgb", Label: "RGB"}, {Value: "hsl", Label: "HSL"},
				},
			},
			{Key: "playSound", Label: "Play sound on copy", Type: core.SettingToggle, Value: true},
		},
	})

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		x, y := win32.CursorPos()
		r, g, b, ok := win32.PixelColor(x, y)
		if !ok {
			log.Printf("[color-picker] GetPixel failed at %d,%d", x, y)
			return
		}
		text := formatColor(m.SettingString("format"), r, g, b)
		app.Clipboard.SetText(text)
		if m.SettingBool("playSound") {
			win32.Beep()
		}
		log.Printf("[color-picker] %d,%d -> %s", x, y, text)
	})

	return m.WithHandlers(hk.register, hk.unregister).
		WithSettingHandler(func(key string, _ any) error {
			if key != "hotkey" || !m.Info().Enabled {
				return nil
			}
			return hk.rebind()
		})
}

func formatColor(format string, r, g, b uint8) string {
	switch format {
	case "rgb":
		return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
	case "hsl":
		h, s, l := rgbToHSL(r, g, b)
		return fmt.Sprintf("hsl(%.0f, %.0f%%, %.0f%%)", h, s*100, l*100)
	default:
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	default:
		h = (rf-gf)/d + 4
	}
	return h * 60, s, l
}

// --- System Info: live stats rendered on the module page ---

func newSystemInfo() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "system-info",
		Name:        "System Info",
		Description: "Show CPU, memory, GPU and system information.",
		Icon:        "DesktopPulse",
		Category:    core.CategorySystem,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "refreshSeconds", Label: "Refresh interval", Description: "Seconds between stat updates.", Type: core.SettingSlider, Value: 2.0, Min: f64(1), Max: f64(10), Step: f64(1)},
		},
	})
}
