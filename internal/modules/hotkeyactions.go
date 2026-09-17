//go:build windows

package modules

import (
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
)

// --- Hotkey Actions: lines of "accelerator = command" bind global hotkeys
// to launching programs or shell commands. ---

func newHotkeyActions(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "hotkey-actions",
		Name:        "Hotkey Actions",
		Description: "Bind global shortcuts to actions: launch apps, run commands, toggle utilities.",
		Icon:        "KeyCommand",
		Category:    core.CategoryInput,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key: "bindings", Label: "Hotkey bindings",
				Description: "One per line: shortcut = command (e.g. ctrl+alt+n = notepad).",
				Type:        core.SettingText, Value: "ctrl+alt+n = notepad",
			},
		},
	})

	var (
		mu         sync.Mutex
		registered []string
	)

	unregisterAll := func() {
		mu.Lock()
		defer mu.Unlock()
		for _, combo := range registered {
			_ = app.GlobalShortcut.Unregister(combo)
		}
		registered = nil
	}

	registerAll := func() {
		unregisterAll()
		var ok []string
		for _, line := range strings.Split(m.SettingString("bindings"), "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			combo := strings.ToLower(strings.TrimSpace(parts[0]))
			cmd := strings.TrimSpace(parts[1])
			if combo == "" || cmd == "" {
				continue
			}
			err := app.GlobalShortcut.Register(combo, func() {
				log.Printf("[hotkey-actions] %s -> %s", combo, cmd)
				c := exec.Command("cmd", "/c", cmd)
				c.SysProcAttr = sysProcAttrHide
				if err := c.Start(); err != nil {
					log.Printf("[hotkey-actions] exec failed: %v", err)
				}
			})
			if err != nil {
				log.Printf("[hotkey-actions] register %q failed: %v", combo, err)
				continue
			}
			ok = append(ok, combo)
		}
		mu.Lock()
		registered = ok
		mu.Unlock()
	}

	return m.WithHandlers(func() error {
		registerAll()
		return nil
	}, func() error {
		unregisterAll()
		return nil
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "bindings" || !m.Info().Enabled {
			return nil
		}
		registerAll()
		return nil
	})
}
