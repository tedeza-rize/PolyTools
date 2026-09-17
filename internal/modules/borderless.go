//go:build windows

package modules

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Borderless Gaming: strip the foreground window's frame and stretch it
// across its monitor. Toggle restores the original style + rect. ---

func newBorderlessGaming(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "borderless-gaming",
		Name:        "Borderless Gaming",
		Description: "Force games into borderless windowed mode.",
		Icon:        "BorderNone",
		Category:    core.CategoryGaming,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle borderless on foreground window", Type: core.SettingShortcut, Value: "super+f11"},
			{Key: "autoApply", Label: "Apply automatically", Description: "Apply borderless when a listed game launches.", Type: core.SettingToggle, Value: false},
			{Key: "gameList", Label: "Game list", Description: "Comma-separated executable names, e.g. game.exe, othergame.exe", Type: core.SettingText, Value: ""},
		},
	})

	type saved struct {
		style uintptr
		rc    win32.Rect
	}
	var (
		mu      sync.Mutex
		applied = map[uintptr]saved{}
		stop    func()
	)

	gameSet := func() map[string]bool {
		set := map[string]bool{}
		for _, s := range strings.Split(m.SettingString("gameList"), ",") {
			if s = strings.ToLower(strings.TrimSpace(s)); s != "" {
				set[s] = true
			}
		}
		return set
	}

	apply := func(hwnd uintptr) {
		mu.Lock()
		if _, ok := applied[hwnd]; ok {
			mu.Unlock()
			return
		}
		mu.Unlock()
		style := win32.WindowStyle(hwnd)
		rc, ok := win32.PlacementRect(hwnd)
		if !ok {
			return
		}
		mr := win32.MonitorRect(hwnd)
		mu.Lock()
		applied[hwnd] = saved{style: style, rc: rc}
		mu.Unlock()
		win32.SetWindowStyle(hwnd, style&^win32.WSOverlapped)
		win32.SetWindowPosTo(hwnd, mr)
		log.Printf("[borderless] applied to %s", win32.WindowExe(hwnd))
	}

	unapply := func(hwnd uintptr) {
		mu.Lock()
		s, ok := applied[hwnd]
		if ok {
			delete(applied, hwnd)
		}
		mu.Unlock()
		if !ok || !win32.IsWindowAlive(hwnd) {
			return
		}
		win32.SetWindowStyle(hwnd, s.style)
		win32.SetWindowPosTo(hwnd, s.rc)
	}

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		hwnd := win32.ForegroundWindow()
		if hwnd == 0 {
			return
		}
		mu.Lock()
		_, isApplied := applied[hwnd]
		mu.Unlock()
		if isApplied {
			unapply(hwnd)
		} else {
			apply(hwnd)
		}
	})

	onEvent := func(ev win32.WinEvent) {
		if ev.Event != 0x8002 || !m.SettingBool("autoApply") { // EVENT_OBJECT_SHOW
			return
		}
		hwnd := win32.RootWindow(ev.Hwnd)
		if !gameSet()[win32.WindowExe(hwnd)] {
			return
		}
		go func() {
			time.Sleep(1500 * time.Millisecond)
			if win32.IsWindowAlive(hwnd) {
				apply(hwnd)
			}
		}()
	}

	return m.WithHandlers(func() error {
		if err := hk.register(); err != nil {
			return err
		}
		stop = win32.StartWinEventHook(0x8002, 0x8002, onEvent)
		return nil
	}, func() error {
		if stop != nil {
			stop()
			stop = nil
		}
		mu.Lock()
		restores := applied
		applied = map[uintptr]saved{}
		mu.Unlock()
		for hwnd, s := range restores {
			if win32.IsWindowAlive(hwnd) {
				win32.SetWindowStyle(hwnd, s.style)
				win32.SetWindowPosTo(hwnd, s.rc)
			}
		}
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}
