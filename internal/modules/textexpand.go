//go:build windows

package modules

import (
	"strings"
	"sync"
	"time"
	"unicode"

	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Text Expander: typed abbreviations expand into full text.
// Rules are edited on the module page (lines of "trigger = expansion"). ---

func newTextExpander() *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "text-expander",
		Name:        "Text Expander",
		Description: "Expand short abbreviations into full text as you type.",
		Icon:        "TextExpand",
		Category:    core.CategoryInput,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key: "trigger", Label: "Expand trigger", Type: core.SettingSelect, Value: "tab",
				Options: []core.SelectOption{
					{Value: "instant", Label: "Instant"}, {Value: "tab", Label: "On Tab"}, {Value: "space", Label: "On Space"},
				},
			},
			{
				Key: "rules", Label: "Expansions",
				Description: "One per line: trigger = expansion",
				Type:        core.SettingText, Value: ";mail = name@example.com\n;sig = Best regards,\n;addr = 123 Main Street",
			},
		},
	})

	var (
		mu    sync.Mutex
		buf   []rune // recent typed characters
		state [256]byte
		stop  func()
	)

	rules := func() map[string]string {
		out := map[string]string{}
		for _, line := range strings.Split(m.SettingString("rules"), "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				if k := strings.TrimSpace(parts[0]); k != "" {
					out[k] = strings.TrimSpace(parts[1])
				}
			}
		}
		return out
	}

	expand := func(trigger string, expansion string, swallowTrigger bool) {
		go func() {
			// Let the trigger key land first unless swallowed, then erase.
			time.Sleep(15 * time.Millisecond)
			win32.SendBackspaces(len([]rune(trigger)))
			time.Sleep(10 * time.Millisecond)
			win32.SendText(expansion)
		}()
	}

	onKey := func(ev win32.KeyEvent) bool {
		// Keep the modifier state array fresh for ToUnicode.
		if ev.VK < 256 {
			if ev.Down {
				state[ev.VK] = 0x80
			} else {
				state[ev.VK] = 0
			}
		}
		if !ev.Down {
			return false
		}

		triggerMode := m.SettingString("trigger")

		// Buffer bookkeeping.
		switch ev.VK {
		case 0x08: // backspace
			mu.Lock()
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
			mu.Unlock()
			return false
		case 0x09: // tab
			if triggerMode == "tab" {
				if trig, exp, ok := matchRules(rules(), bufferString(&mu, &buf)); ok {
					mu.Lock()
					buf = buf[:0]
					mu.Unlock()
					expand(trig, exp, true)
					return true // swallow the tab
				}
			}
			mu.Lock()
			buf = buf[:0]
			mu.Unlock()
			return false
		case 0x20: // space
			if triggerMode == "space" {
				if trig, exp, ok := matchRules(rules(), bufferString(&mu, &buf)); ok {
					mu.Lock()
					buf = buf[:0]
					mu.Unlock()
					expand(trig, exp, true)
					return true
				}
			}
		case 0x25, 0x26, 0x27, 0x28, 0x1B, 0x0D: // arrows, esc, enter
			mu.Lock()
			buf = buf[:0]
			mu.Unlock()
			return false
		}

		// Translate vk+scan into a rune using our tracked modifier state.
		if r := toChar(ev, &state); r != 0 && !unicode.IsControl(r) {
			mu.Lock()
			buf = append(buf, r)
			if len(buf) > 64 {
				buf = buf[len(buf)-64:]
			}
			s := string(buf)
			mu.Unlock()
			if triggerMode == "instant" || triggerMode == "space" {
				if trig, exp, ok := matchRules(rules(), s); ok {
					mu.Lock()
					buf = buf[:0]
					mu.Unlock()
					expand(trig, exp, false)
					return triggerMode == "space"
				}
			}
			return false
		}
		return false
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
	})
}

func bufferString(mu *sync.Mutex, buf *[]rune) string {
	mu.Lock()
	defer mu.Unlock()
	return string(*buf)
}

// matchRules finds the longest trigger that is a suffix of typed text.
func matchRules(rules map[string]string, typed string) (string, string, bool) {
	var best string
	for trig := range rules {
		if len(trig) > len(best) && strings.HasSuffix(typed, trig) {
			best = trig
		}
	}
	if best == "" {
		return "", "", false
	}
	return best, rules[best], true
}

// toChar converts a key event to a rune via ToUnicode with tracked state.
func toChar(ev win32.KeyEvent, state *[256]byte) rune {
	return win32.VKToChar(ev.VK, ev.ScanCode, state)
}
