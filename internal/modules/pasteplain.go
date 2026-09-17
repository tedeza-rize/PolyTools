//go:build windows

package modules

import (
	"log"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// sysProcAttrHide prevents a console window flashing for spawned commands.
var sysProcAttrHide = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW

// --- Paste as Plain Text: hotkey replaces clipboard content with plain
// text and pastes it (like Ctrl+Shift+V everywhere). ---

func newPastePlain(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "paste-plain",
		Name:        "Paste Plain Text",
		Description: "Paste clipboard content without formatting.",
		Icon:        "TextFont",
		Category:    core.CategoryInput,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Paste plain text", Type: core.SettingShortcut, Value: "super+shift+v"},
		},
	})

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		text := win32.ClipboardText()
		if text == "" {
			return
		}
		win32.SetClipboardText(text)
		time.Sleep(30 * time.Millisecond)
		// Ctrl+V into the focused window.
		win32.SendKeys(0x11, 0x56) // VK_CONTROL, 'V'
		log.Printf("[paste-plain] pasted %d chars", len(text))
	})

	return m.WithHandlers(hk.register, hk.unregister).
		WithSettingHandler(func(key string, _ any) error {
			if key != "hotkey" || !m.Info().Enabled {
				return nil
			}
			return hk.rebind()
		})
}
