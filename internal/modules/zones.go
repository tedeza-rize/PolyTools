//go:build windows

package modules

import (
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Zone Layouts: hold Shift while dragging a window → zones preview
// appears; release to snap the window into the hovered zone. ---

func newZoneLayouts(app *application.App, win *application.WebviewWindow) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "zone-layouts",
		Name:        "Zone Layouts",
		Description: "Create custom window layouts and snap windows into zones.",
		Icon:        "Grid",
		Category:    core.CategoryWindowing,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "shiftToSnap", Label: "Hold Shift to snap", Type: core.SettingToggle, Value: true},
			{
				Key: "layout", Label: "Layout", Type: core.SettingSelect, Value: "columns2",
				Options: []core.SelectOption{
					{Value: "columns2", Label: "Two columns"},
					{Value: "columns3", Label: "Three columns"},
					{Value: "grid4", Label: "2 × 2 grid"},
					{Value: "priority", Label: "Priority (wide left + two right)"},
				},
			},
		},
	})

	var (
		mu       sync.Mutex
		stop     func()
		overlay  *win32.Overlay
		dragHwnd uintptr
		monRect  win32.Rect
		zones    []win32.Rect
		hovered  int
		ownHwnd  uintptr
	)

	computeZones := func(area win32.Rect) []win32.Rect {
		w, h := area.Width(), area.Height()
		gap := int32(8)
		switch m.SettingString("layout") {
		case "columns3":
			cw := w / 3
			return []win32.Rect{
				{area.Left, area.Top, area.Left + cw - gap, area.Bottom},
				{area.Left + cw, area.Top, area.Left + 2*cw - gap, area.Bottom},
				{area.Left + 2 * cw, area.Top, area.Right, area.Bottom},
			}
		case "grid4":
			cw, ch := w/2, h/2
			return []win32.Rect{
				{area.Left, area.Top, area.Left + cw - gap, area.Top + ch - gap},
				{area.Left + cw, area.Top, area.Right, area.Top + ch - gap},
				{area.Left, area.Top + ch, area.Left + cw - gap, area.Bottom},
				{area.Left + cw, area.Top + ch, area.Right, area.Bottom},
			}
		case "priority":
			cw := w * 2 / 3
			return []win32.Rect{
				{area.Left, area.Top, area.Left + cw - gap, area.Bottom},
				{area.Left + cw, area.Top, area.Right, area.Top + h/2 - gap},
				{area.Left + cw, area.Top + h/2, area.Right, area.Bottom},
			}
		default: // columns2
			cw := w / 2
			return []win32.Rect{
				{area.Left, area.Top, area.Left + cw - gap, area.Bottom},
				{area.Left + cw, area.Top, area.Right, area.Bottom},
			}
		}
	}

	showOverlay := func() {
		mu.Lock()
		defer mu.Unlock()
		if overlay != nil {
			overlay.Destroy()
		}
		o, err := win32.NewOverlay(monRect.Left, monRect.Top, monRect.Width(), monRect.Height())
		if err != nil {
			return
		}
		overlay = o
		drawZones(o, monRect, zones, hovered)
		o.Present()
	}

	hideOverlay := func() {
		mu.Lock()
		defer mu.Unlock()
		if overlay != nil {
			overlay.Destroy()
			overlay = nil
		}
	}

	onMouse := func(ev win32.MouseEvent) bool {
		shiftSnap := m.SettingBool("shiftToSnap")
		switch ev.Msg {
		case 0x0201: // LButton down — begin tracking a potential drag
			hwnd := win32.WindowAtPoint(ev.X, ev.Y)
			if hwnd == 0 {
				return false
			}
			if ownHwnd == 0 && win != nil {
				ownHwnd = uintptr(win.NativeWindow())
			}
			if hwnd == ownHwnd {
				return false
			}
			mu.Lock()
			dragHwnd = hwnd
			monRect = win32.MonitorRect(hwnd)
			zones = computeZones(monRect)
			hovered = -1
			mu.Unlock()
		case 0x0200: // move
			mu.Lock()
			h, zs, mr := dragHwnd, zones, monRect
			shown := overlay != nil
			mu.Unlock()
			if h == 0 {
				return false
			}
			if shiftSnap && win32.KeyDownState(0x10) { // Shift
				// which zone contains the cursor?
				idx := -1
				for i, z := range zs {
					if ev.X >= z.Left && ev.X < z.Right && ev.Y >= z.Top && ev.Y < z.Bottom {
						idx = i
						break
					}
				}
				if !shown || idx != hovered {
					mu.Lock()
					hovered = idx
					mu.Unlock()
					if shown {
						// redraw with new highlight
						mu.Lock()
						if overlay != nil {
							o := overlay
							drawZones(o, mr, zs, idx)
							o.Present()
						}
						mu.Unlock()
					} else {
						showOverlay()
					}
				}
			} else {
				mu.Lock()
				dragHwnd = 0
				mu.Unlock()
				hideOverlay()
			}
		case 0x0202: // LButton up
			mu.Lock()
			h, zs, idx := dragHwnd, zones, hovered
			dragHwnd = 0
			mu.Unlock()
			hideOverlay()
			if h != 0 && idx >= 0 && idx < len(zs) && win32.IsWindowAlive(h) {
				z := zs[idx]
				win32.SetWindowPosTo(h, z)
				win32.SetForeground(h)
				log.Printf("[zone-layouts] snapped to zone %d %+v", idx, z)
			}
		}
		return false
	}

	return m.WithHandlers(func() error {
		stop = win32.StartMouseHook(onMouse)
		return nil
	}, func() error {
		hideOverlay()
		if stop != nil {
			stop()
			stop = nil
		}
		return nil
	})
}

func drawZones(o *win32.Overlay, mon win32.Rect, zones []win32.Rect, hovered int) {
	o.Fill(0, 0, 0, 0)
	for i, z := range zones {
		rx, ry := z.Left-mon.Left, z.Top-mon.Top
		rw, rh := z.Width(), z.Height()
		if i == hovered {
			o.FillRect(rx, ry, rw, rh, 15, 108, 189, 90) // brand fill
			o.BorderRect(rx, ry, rw, rh, 3, 15, 108, 189, 255)
		} else {
			o.FillRect(rx, ry, rw, rh, 15, 108, 189, 40)
			o.BorderRect(rx, ry, rw, rh, 2, 200, 200, 200, 180)
		}
	}
}
