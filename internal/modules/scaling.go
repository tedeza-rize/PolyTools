//go:build windows

package modules

import (
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Lossless Scaling: integer display magnification via the Magnification
// API — equivalent of Lossless Scaling's integer mode. Hotkey toggles. ---

func newLosslessScaling(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "lossless-scaling",
		Name:        "Lossless Scaling",
		Description: "Scale game windows up to fullscreen resolution.",
		Icon:        "ArrowMaximize",
		Category:    core.CategoryGaming,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle scaling", Type: core.SettingShortcut, Value: "ctrl+alt+s"},
			{
				Key: "scalingMode", Label: "Scaling mode", Type: core.SettingSelect, Value: "integer",
				Options: []core.SelectOption{
					{Value: "integer", Label: "Integer (nearest)"},
					{Value: "smooth", Label: "Smooth (fractional allowed)"},
				},
			},
			{
				Key: "factor", Label: "Scale factor", Type: core.SettingSlider,
				Value: 2.0, Min: f64(1.5), Max: f64(4), Step: f64(0.5),
			},
		},
	})

	var (
		mu      sync.Mutex
		scaling bool
		magUp   bool
	)

	setScaled := func(on bool) {
		mu.Lock()
		defer mu.Unlock()
		if on == scaling {
			return
		}
		if on && !magUp {
			if !win32.MagInit() {
				log.Printf("[lossless] MagInitialize failed")
				return
			}
			magUp = true
		}
		factor := float32(m.SettingFloat("factor"))
		if m.SettingString("scalingMode") == "integer" {
			factor = float32(int32(factor + 0.5))
			if factor < 2 {
				factor = 2
			}
		}
		var ox, oy int32
		if on {
			// Center the magnification on the foreground window.
			if hwnd := win32.ForegroundWindow(); hwnd != 0 {
				if rc, ok := win32.WindowRect(hwnd); ok {
					mr := win32.MonitorRect(hwnd)
					cx := (rc.Left + rc.Right) / 2
					cy := (rc.Top + rc.Bottom) / 2
					ox = cx - mr.Left - int32(float32(mr.Width())/factor/2)
					oy = cy - mr.Top - int32(float32(mr.Height())/factor/2)
					_ = mr
				}
			}
			if !win32.SetFullscreenScale(factor, ox, oy) {
				log.Printf("[lossless] MagSetFullscreenTransform failed")
				return
			}
			scaling = true
			log.Printf("[lossless] scaling on x%.2f", factor)
		} else {
			win32.SetFullscreenScale(1.0, 0, 0)
			scaling = false
			log.Printf("[lossless] scaling off")
		}
	}

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		mu.Lock()
		on := !scaling
		mu.Unlock()
		setScaled(on)
	})

	return m.WithHandlers(hk.register, func() error {
		setScaled(false)
		mu.Lock()
		if magUp {
			win32.MagShutdown()
			magUp = false
		}
		mu.Unlock()
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}
