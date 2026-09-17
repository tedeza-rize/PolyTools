//go:build windows

package modules

import (
	"strings"
	"sync"

	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Keyboard Manager: single-key remapping via a low-level keyboard hook.
// Rules: lines of "Source = Target" using key names (CapsLock = Escape). ---

func newKeyboardManager() *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "keyboard-manager",
		Name:        "Keyboard Manager",
		Description: "Remap keys and shortcuts.",
		Icon:        "Keyboard",
		Category:    core.CategoryInput,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key: "remap", Label: "Key remapping",
				Description: "One per line: Source = Target (e.g. CapsLock = Escape). Single keys only.",
				Type:        core.SettingText, Value: "CapsLock = Escape",
			},
		},
	})

	var (
		mu   sync.Mutex
		stop func()
	)

	mappings := func() map[uint16]uint16 {
		out := map[uint16]uint16{}
		for _, line := range strings.Split(m.SettingString("remap"), "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			src, ok1 := win32.VKByName[strings.ToLower(strings.TrimSpace(parts[0]))]
			dst, ok2 := win32.VKByName[strings.ToLower(strings.TrimSpace(parts[1]))]
			if ok1 && ok2 && src != dst {
				out[src] = dst
			}
		}
		return out
	}

	// injected tracks keys we sent so the hook ignores its own events.
	var injected sync.Map // map[uint16]int

	onKey := func(ev win32.KeyEvent) bool {
		if ev.Down {
			if n, ok := injected.Load(ev.VK); ok && n.(int) > 0 {
				injected.Store(ev.VK, n.(int)-1)
				return false // our own injected press
			}
		}
		mp := mappings()
		dst, ok := mp[ev.VK]
		if !ok {
			return false
		}
		// Swallow the original and inject the mapped key.
		go func(down bool, vk uint16) {
			if down {
				injected.Store(vk, 1)
				win32.KeyDown(vk)
			} else {
				win32.KeyUp(vk)
			}
		}(ev.Down, dst)
		return true
	}

	return m.WithHandlers(func() error {
		stop = win32.StartKeyboardHook(onKey)
		return nil
	}, func() error {
		if stop != nil {
			stop()
			stop = nil
		}
		return nil
	}).WithSettingHandler(func(key string, _ any) error {
		mu.Lock()
		defer mu.Unlock()
		return nil
	})
}
