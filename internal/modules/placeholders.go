//go:build windows

package modules

import "polytools/internal/core"

// Coming-soon modules: metadata only (Available=false), so the settings UI
// shows the full catalog while implementations land later.

func newTextExtractor() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "text-extractor",
		Name:        "Text Extractor",
		Description: "Copy text from anywhere on the screen using OCR.",
		Icon:        "TextFont",
		Category:    core.CategorySystem,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation shortcut", Type: core.SettingShortcut, Value: "super+shift+t"},
		},
	})
}

func newColorPicker() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "color-picker",
		Name:        "Color Picker",
		Description: "Pick colors from anywhere on the screen.",
		Icon:        "Color",
		Category:    core.CategorySystem,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation shortcut", Type: core.SettingShortcut, Value: "super+shift+c"},
			{
				Key:   "format", Label: "Color format", Type: core.SettingSelect, Value: "hex",
				Options: []core.SelectOption{{Value: "hex", Label: "HEX"}, {Value: "rgb", Label: "RGB"}, {Value: "hsl", Label: "HSL"}},
			},
		},
	})
}

func newScreenRuler() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "screen-ruler",
		Name:        "Screen Ruler",
		Description: "Measure pixel distances on screen.",
		Icon:        "Ruler",
		Category:    core.CategorySystem,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation shortcut", Type: core.SettingShortcut, Value: "super+shift+m"},
		},
	})
}

func newGrabAndMove() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "grab-and-move",
		Name:        "Grab And Move",
		Description: "Move and resize windows by holding a modifier and dragging anywhere inside the window.",
		Icon:        "ArrowMove",
		Category:    core.CategoryWindowing,
		Available:   false,
		Settings: []core.SettingField{
			{
				Key:   "modifier", Label: "Modifier key", Type: core.SettingSelect, Value: "alt",
				Options: []core.SelectOption{{Value: "alt", Label: "Alt"}, {Value: "win", Label: "Win"}, {Value: "ctrl", Label: "Ctrl"}},
			},
		},
	})
}

func newWindowMemory() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "window-memory",
		Name:        "Window Memory",
		Description: "Remember window positions and sizes, and restore them automatically.",
		Icon:        "WindowMultiple",
		Category:    core.CategoryWindowing,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "restoreOnLaunch", Label: "Restore on launch", Type: core.SettingToggle, Value: true},
		},
	})
}

func newZoneLayouts() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "zone-layouts",
		Name:        "Zone Layouts",
		Description: "Create custom window layouts and snap windows into zones.",
		Icon:        "Grid",
		Category:    core.CategoryWindowing,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "shiftToSnap", Label: "Hold Shift to snap", Type: core.SettingToggle, Value: true},
		},
	})
}

func newKeyboardManager() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "keyboard-manager",
		Name:        "Keyboard Manager",
		Description: "Remap keys and shortcuts.",
		Icon:        "Keyboard",
		Category:    core.CategoryInput,
		Available:   false,
	})
}

func newMouseUtilities() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "mouse-utilities",
		Name:        "Mouse Utilities",
		Description: "Find My Mouse, highlighter, crosshairs and more.",
		Icon:        "Cursor",
		Category:    core.CategoryInput,
		Available:   false,
	})
}

func newBatchRename() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "batch-rename",
		Name:        "Batch Rename",
		Description: "Bulk-rename files with search/replace and regular expressions.",
		Icon:        "Rename",
		Category:    core.CategoryFiles,
		Available:   false,
	})
}

func newQuickPeek() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "quick-peek",
		Name:        "Quick Peek",
		Description: "Preview file content without opening the application.",
		Icon:        "Eye",
		Category:    core.CategoryFiles,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Activation key", Type: core.SettingShortcut, Value: "ctrl+space"},
		},
	})
}

func newEnvVars() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "env-vars",
		Name:        "Environment Variables",
		Description: "Manage user and system environment variables.",
		Icon:        "BracesVariable",
		Category:    core.CategoryAdvanced,
		Available:   false,
	})
}

func newHostsEditor() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "hosts-editor",
		Name:        "Hosts File Editor",
		Description: "Edit the Windows hosts file.",
		Icon:        "Globe",
		Category:    core.CategoryAdvanced,
		Available:   false,
	})
}
