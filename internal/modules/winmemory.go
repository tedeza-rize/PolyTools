//go:build windows

package modules

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Window Memory: remembers each app's window rect (keyed by exe) and
// restores it when a matching window is shown again. ---

func newWindowMemory() *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "window-memory",
		Name:        "Window Memory",
		Description: "Remember window positions and sizes, and restore them automatically.",
		Icon:        "WindowMultiple",
		Category:    core.CategoryWindowing,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "restoreOnLaunch", Label: "Restore on launch", Type: core.SettingToggle, Value: true},
		},
	})

	path := filepath.Join(stateDir(), "window-memory.json")

	var (
		mu    sync.Mutex
		known = map[string]win32.Rect{} // exe -> rect
		stop  func()
		dirty bool
	)

	load := func() {
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &known)
		}
	}
	save := func() {
		raw, err := json.MarshalIndent(known, "", "  ")
		if err == nil {
			_ = os.WriteFile(path, raw, 0o644)
		}
	}

	onEvent := func(ev win32.WinEvent) {
		hwnd := win32.RootWindow(ev.Hwnd)
		exe := win32.WindowExe(hwnd)
		if exe == "" || exe == "polytools.exe" {
			return
		}
		switch ev.Event {
		case 0x000B: // EVENT_SYSTEM_MOVESIZEEND — user finished moving/sizing
			if rc, ok := win32.PlacementRect(hwnd); ok {
				mu.Lock()
				known[exe] = rc
				dirty = true
				mu.Unlock()
			}
		case 0x8002: // EVENT_OBJECT_SHOW — window appeared
			if !m.SettingBool("restoreOnLaunch") {
				return
			}
			mu.Lock()
			rc, ok := known[exe]
			mu.Unlock()
			if !ok {
				return
			}
			// Small delay: many apps position themselves after first show.
			go func(hwnd uintptr, want win32.Rect) {
				time.Sleep(300 * time.Millisecond)
				if !win32.IsWindowAlive(hwnd) {
					return
				}
				cur, _ := win32.WindowRect(hwnd)
				if cur == want {
					return
				}
				win32.SetWindowPosTo(hwnd, want)
				log.Printf("[window-memory] restored %s to %+v", exe, want)
			}(hwnd, rc)
		}
	}

	return m.WithHandlers(func() error {
		load()
		stop = win32.StartWinEventHook(0x0001, 0x8FFF, onEvent)
		// periodic flush of dirty state
		go func() {
			for range time.Tick(10 * time.Second) {
				mu.Lock()
				d := dirty
				dirty = false
				mu.Unlock()
				if d {
					save()
				}
			}
		}()
		return nil
	}, func() error {
		if stop != nil {
			stop()
			stop = nil
		}
		save()
		return nil
	})
}

// stateDir returns %LOCALAPPDATA%\PolyTools (same place as settings.json).
func stateDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "."
	}
	d := filepath.Join(dir, "PolyTools")
	_ = os.MkdirAll(d, 0o755)
	return d
}
