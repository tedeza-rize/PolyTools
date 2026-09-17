//go:build windows

package modules

import (
	"math"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Mouse Utilities: Find My Mouse — double-Ctrl (or a hotkey) dims the
// screen except a spotlight circle around the cursor. ---

func newMouseUtilities(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "mouse-utilities",
		Name:        "Mouse Utilities",
		Description: "Find My Mouse spotlight and cursor utilities.",
		Icon:        "Cursor",
		Category:    core.CategoryInput,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Spotlight shortcut", Type: core.SettingShortcut, Value: "ctrl+alt+f"},
			{Key: "doubleCtrl", Label: "Double-tap Ctrl", Description: "Also trigger by pressing Ctrl twice quickly.", Type: core.SettingToggle, Value: true},
			{Key: "radius", Label: "Spotlight radius", Type: core.SettingSlider, Value: 140.0, Min: f64(60), Max: f64(400), Step: f64(10)},
		},
	})

	var (
		mu        sync.Mutex
		overlay   *win32.Overlay
		stopKeys  func()
		lastCtrl  time.Time
		visibleAt time.Time
	)

	showSpotlight := func() {
		x, y := win32.CursorPos()
		mr := win32.MonitorRectAt(x, y)
		mu.Lock()
		if overlay != nil {
			overlay.Destroy()
		}
		o, err := win32.NewOverlay(mr.Left, mr.Top, mr.Width(), mr.Height())
		if err != nil {
			mu.Unlock()
			return
		}
		overlay = o
		visibleAt = time.Now()
		mu.Unlock()
		o.Fill(0, 0, 0, 150) // dim everything
		o.CircleHole(x-mr.Left, y-mr.Top, int32(m.SettingFloat("radius")))
		// Soft white ring around the hole.
		r := int32(m.SettingFloat("radius"))
		for t := int32(0); t < 4; t++ {
			drawRing(o, x-mr.Left, y-mr.Top, r+t, 255, 255, 255, uint8(200-t*40))
		}
		o.Present()
	}

	hideSpotlight := func() {
		mu.Lock()
		defer mu.Unlock()
		if overlay != nil {
			overlay.Destroy()
			overlay = nil
		}
	}

	// Keyboard hook only watches for double-Ctrl + any key to dismiss.
	onKey := func(ev win32.KeyEvent) bool {
		if ev.Down {
			mu.Lock()
			shown := overlay != nil && time.Since(visibleAt) > 250*time.Millisecond
			mu.Unlock()
			if shown {
				hideSpotlight()
				return false
			}
			if ev.VK == 0x11 && m.SettingBool("doubleCtrl") { // Ctrl
				if time.Since(lastCtrl) < 350*time.Millisecond {
					go showSpotlight()
					lastCtrl = time.Time{}
				} else {
					lastCtrl = time.Now()
				}
			}
		}
		return false
	}

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		mu.Lock()
		shown := overlay != nil
		mu.Unlock()
		if shown {
			hideSpotlight()
		} else {
			showSpotlight()
		}
	})

	// Dismiss on click too — cheap mouse hook while overlay shown.
	var stopMouse func()
	onMouse := func(ev win32.MouseEvent) bool {
		if ev.Msg == 0x0201 || ev.Msg == 0x0204 {
			mu.Lock()
			shown := overlay != nil && time.Since(visibleAt) > 250*time.Millisecond
			mu.Unlock()
			if shown {
				hideSpotlight()
			}
		}
		return false
	}

	return m.WithHandlers(func() error {
		if err := hk.register(); err != nil {
			return err
		}
		stopKeys = win32.StartKeyboardHook(onKey)
		stopMouse = win32.StartMouseHook(onMouse)
		return nil
	}, func() error {
		hideSpotlight()
		if stopKeys != nil {
			stopKeys()
			stopKeys = nil
		}
		if stopMouse != nil {
			stopMouse()
			stopMouse = nil
		}
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}

// drawRing draws a thin circle outline (midpoint circle, approximate).
func drawRing(o *win32.Overlay, cx, cy, r int32, red, g, b, a uint8) {
	for deg := 0; deg < 360; deg += 2 {
		rad := float64(deg) * 0.0174533
		x := cx + int32(float64(r)*math.Cos(rad))
		y := cy + int32(float64(r)*math.Sin(rad))
		o.FillRect(x, y, 2, 2, b, g, red, a)
	}
}
