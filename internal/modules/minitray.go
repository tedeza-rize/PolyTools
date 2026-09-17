//go:build windows

package modules

import (
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Minimize To Tray: hotkey hides the foreground window; the PolyTools
// tray menu gains a "Restore" item per hidden window. ---

func newMinimizeToTray(app *application.App, tray *application.SystemTray, showMain func()) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "minimize-to-tray",
		Name:        "Minimize To Tray",
		Description: "Send any window to the system tray.",
		Icon:        "ArrowMinimize",
		Category:    core.CategoryWindowing,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Send window to tray", Type: core.SettingShortcut, Value: "super+shift+down"},
		},
	})

	type hiddenWin struct {
		hwnd  uintptr
		title string
	}
	var (
		mu     sync.Mutex
		hidden []hiddenWin
	)

	var rebuildMenu func()
	rebuildMenu = func() {
		mu.Lock()
		// prune windows whose process exited while hidden
		alive := hidden[:0]
		for _, h := range hidden {
			if win32.IsWindowAlive(h.hwnd) {
				alive = append(alive, h)
			}
		}
		hidden = alive
		list := append([]hiddenWin(nil), hidden...)
		mu.Unlock()

		menu := application.NewMenu()
		menu.Add("Open PolyTools").OnClick(func(*application.Context) {
			showMain()
		})
		for _, hw := range list {
			hw := hw
			label := "Restore: " + hw.title
			if len([]rune(hw.title)) > 40 {
				label = "Restore: " + string([]rune(hw.title)[:40]) + "…"
			}
			menu.Add(label).OnClick(func(*application.Context) {
				mu.Lock()
				for i, h := range hidden {
					if h.hwnd == hw.hwnd {
						hidden = append(hidden[:i], hidden[i+1:]...)
						break
					}
				}
				mu.Unlock()
				win32.ShowWindowS(hw.hwnd)
				win32.RestoreWindow(hw.hwnd)
				win32.SetForeground(hw.hwnd)
				rebuildMenu()
			})
		}
		if len(list) > 0 {
			menu.Add("Restore all").OnClick(func(*application.Context) {
				mu.Lock()
				all := hidden
				hidden = nil
				mu.Unlock()
				for _, hw := range all {
					win32.ShowWindowS(hw.hwnd)
					win32.RestoreWindow(hw.hwnd)
				}
				rebuildMenu()
			})
		}
		menu.AddSeparator()
		menu.Add("Quit").OnClick(func(*application.Context) { app.Quit() })
		tray.SetMenu(menu)
	}

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		hwnd := win32.ForegroundWindow()
		if hwnd == 0 {
			return
		}
		title := win32.WindowTitle(hwnd)
		if title == "" {
			title = win32.WindowExe(hwnd)
		}
		mu.Lock()
		hidden = append(hidden, hiddenWin{hwnd: hwnd, title: title})
		mu.Unlock()
		win32.HideWindow(hwnd)
		log.Printf("[minimize-to-tray] hid %q (%d)", title, hwnd)
		rebuildMenu()
	})

	return m.WithHandlers(func() error {
		if err := hk.register(); err != nil {
			return err
		}
		rebuildMenu()
		return nil
	}, func() error {
		mu.Lock()
		all := hidden
		hidden = nil
		mu.Unlock()
		for _, hw := range all {
			win32.ShowWindowS(hw.hwnd)
		}
		rebuildMenu()
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}
