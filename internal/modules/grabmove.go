//go:build windows

package modules

import (
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Grab And Move: modifier + left-drag anywhere inside a window moves it;
// modifier + right-drag resizes. ---

func newGrabAndMove(app *application.App, win *application.WebviewWindow) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "grab-and-move",
		Name:        "Grab And Move",
		Description: "Move and resize windows by holding a modifier and dragging anywhere inside the window.",
		Icon:        "ArrowMove",
		Category:    core.CategoryWindowing,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key: "modifier", Label: "Modifier key", Type: core.SettingSelect, Value: "alt",
				Options: []core.SelectOption{{Value: "alt", Label: "Alt"}, {Value: "win", Label: "Win"}, {Value: "ctrl", Label: "Ctrl"}},
			},
			{Key: "skipPolyTools", Label: "Skip PolyTools window", Type: core.SettingToggle, Value: true},
		},
	})

	var (
		mu      sync.Mutex
		stop    func()
		dragHwnd uintptr
		resize   bool
		startPt  struct{ X, Y int32 }
		startRc  win32.Rect
		ownHwnd  uintptr
	)

	onMouse := func(ev win32.MouseEvent) bool {
		mod := m.SettingString("modifier")
		if !win32.ModifierHeld(mod) {
			mu.Lock()
			dragHwnd = 0
			mu.Unlock()
			return false
		}
		switch ev.Msg {
		case 0x0201, 0x0204: // L/R button down
			hwnd := win32.WindowAtPoint(ev.X, ev.Y)
			if hwnd == 0 {
				return false
			}
			if ownHwnd == 0 && win != nil {
				ownHwnd = uintptr(win.NativeWindow())
			}
			if m.SettingBool("skipPolyTools") && hwnd == ownHwnd {
				return false
			}
			rc, ok := win32.WindowRect(hwnd)
			if !ok {
				return false
			}
			mu.Lock()
			dragHwnd = hwnd
			resize = ev.Msg == 0x0204
			startPt.X, startPt.Y = ev.X, ev.Y
			startRc = rc
			mu.Unlock()
			return false // let the click through
		case 0x0200: // move
			mu.Lock()
			h, rs, sp, sr := dragHwnd, resize, startPt, startRc
			mu.Unlock()
			if h == 0 || !win32.IsWindowAlive(h) {
				return false
			}
			dx, dy := ev.X-sp.X, ev.Y-sp.Y
			if rs {
				w, hgt := sr.Width()+dx, sr.Height()+dy
				if w < 100 {
					w = 100
				}
				if hgt < 60 {
					hgt = 60
				}
				win32.MoveWindowTo(h, win32.Rect{Left: sr.Left, Top: sr.Top, Right: sr.Left + w, Bottom: sr.Top + hgt})
			} else {
				win32.MoveWindowTo(h, win32.Rect{
					Left:   sr.Left + dx,
					Top:    sr.Top + dy,
					Right:  sr.Right + dx,
					Bottom: sr.Bottom + dy,
				})
			}
			return false
		case 0x0202, 0x0205: // button up
			mu.Lock()
			dragHwnd = 0
			mu.Unlock()
		}
		return false
	}

	return m.WithHandlers(func() error {
		stop = win32.StartMouseHook(onMouse)
		return nil
	}, func() error {
		if stop != nil {
			stop()
			stop = nil
		}
		return nil
	})
}
