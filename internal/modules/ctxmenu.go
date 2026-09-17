//go:build windows

package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
	"polytools/internal/core"
)

// --- Context Menu: register classic Explorer verbs via HKCU registry.
// Windows 11 modern top-level entries require a packaged identity
// (MSIX + COM) and are tracked but not yet registered. ---

func newContextMenu() *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "context-menu",
		Name:        "Context Menu",
		Description: "Add custom items to the Explorer right-click menu.",
		Icon:        "MoreHorizontal",
		Category:    core.CategoryFiles,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "menuLabel", Label: "Menu label", Type: core.SettingText, Value: "Copy file path"},
			{Key: "command", Label: "Command", Description: `Command template. %1 = selected file, %V = folder. Default copies the path to the clipboard.`, Type: core.SettingText, Value: ""},
			{
				Key: "scope", Label: "Applies to", Type: core.SettingSelect, Value: "files",
				Options: []core.SelectOption{
					{Value: "files", Label: "Files"}, {Value: "folders", Label: "Folders"},
					{Value: "background", Label: "Folder background"}, {Value: "both", Label: "Files and folders"},
				},
			},
			{Key: "classicMenu", Label: "Classic menu", Description: "Show items in the legacy menu (Show more options).", Type: core.SettingToggle, Value: true},
			{Key: "modernMenu", Label: "Modern menu", Description: "Windows 11 top-level menu — needs a packaged component; not registered yet.", Type: core.SettingToggle, Value: false},
		},
	})

	exe, _ := os.Executable()

	sanitized := regexp.MustCompile(`[^A-Za-z0-9]`)

	var remove func()
	apply := func() error {
		remove() // clean previous entries first
		if !m.SettingBool("classicMenu") {
			return nil
		}
		label := m.SettingString("menuLabel")
		if label == "" {
			label = "PolyTools"
		}
		keyName := "PolyTools." + sanitized.ReplaceAllString(label, "")
		cmd := m.SettingString("command")
		if cmd == "" {
			cmd = fmt.Sprintf(`"%s" --copypath "%%1"`, exe)
		}
		scope := m.SettingString("scope")
		bases := []string{}
		switch scope {
		case "folders":
			bases = []string{`Software\Classes\Directory\shell`}
		case "background":
			bases = []string{`Software\Classes\Directory\Background\shell`}
		case "both":
			bases = []string{`Software\Classes\*\shell`, `Software\Classes\Directory\shell`}
		default:
			bases = []string{`Software\Classes\*\shell`}
		}
		for _, base := range bases {
			k, _, err := registry.CreateKey(registry.CURRENT_USER,
				base+`\`+keyName, registry.SET_VALUE)
			if err != nil {
				return err
			}
			_ = k.SetStringValue("", label)
			iconPath := filepath.Join(filepath.Dir(exe), "appicon.ico")
			if _, err := os.Stat(iconPath); err == nil {
				_ = k.SetStringValue("Icon", iconPath)
			}
			k.Close()
			ck, _, err := registry.CreateKey(registry.CURRENT_USER,
				base+`\`+keyName+`\command`, registry.SET_VALUE)
			if err != nil {
				return err
			}
			_ = ck.SetStringValue("", cmd)
			ck.Close()
		}
		return nil
	}

	remove = func() {
		for _, base := range []string{
			`Software\Classes\*\shell`,
			`Software\Classes\Directory\shell`,
			`Software\Classes\Directory\Background\shell`,
		} {
			k, err := registry.OpenKey(registry.CURRENT_USER, base, registry.ENUMERATE_SUB_KEYS|registry.SET_VALUE)
			if err != nil {
				continue
			}
			subs, _ := k.ReadSubKeyNames(-1)
			k.Close()
			for _, s := range subs {
				if strings.HasPrefix(s, "PolyTools.") {
					registry.DeleteKey(registry.CURRENT_USER, base+`\`+s+`\command`)
					registry.DeleteKey(registry.CURRENT_USER, base+`\`+s)
				}
			}
		}
	}

	return m.WithHandlers(apply, func() error {
		remove()
		return nil
	}).WithSettingHandler(func(key string, _ any) error {
		if !m.Info().Enabled {
			return nil
		}
		return apply()
	})
}
