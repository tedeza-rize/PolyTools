//go:build windows

package modules

import "polytools/internal/core"

// Coming-soon modules: metadata only (Available=false), so the settings UI
// shows the full catalog while implementations land later.

func f64(v float64) *float64 { return &v }

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

func newBorderlessGaming() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "borderless-gaming",
		Name:        "Borderless Gaming",
		Description: "Force games into borderless windowed mode.",
		Icon:        "BorderNone",
		Category:    core.CategoryGaming,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle borderless on foreground window", Type: core.SettingShortcut, Value: "super+f11"},
			{Key: "autoApply", Label: "Apply automatically", Description: "Apply borderless when a listed game launches.", Type: core.SettingToggle, Value: false},
			{Key: "gameList", Label: "Game list", Description: "Comma-separated executable names, e.g. game.exe, othergame.exe", Type: core.SettingText, Value: ""},
		},
	})
}

func newFpsOverlay() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "fps-overlay",
		Name:        "FPS Overlay",
		Description: "Show the frame rate of the current game as an overlay.",
		Icon:        "Gauge",
		Category:    core.CategoryGaming,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle overlay", Type: core.SettingShortcut, Value: "super+shift+f"},
			{
				Key: "displayStyle", Label: "Display style", Type: core.SettingSelect, Value: "number",
				Options: []core.SelectOption{
					{Value: "number", Label: "Number only"},
					{Value: "graph", Label: "Graph only"},
					{Value: "number-graph", Label: "Number + graph"},
				},
			},
			{
				Key: "position", Label: "Position", Type: core.SettingSelect, Value: "top-left",
				Options: []core.SelectOption{
					{Value: "top-left", Label: "Top left"}, {Value: "top-right", Label: "Top right"},
					{Value: "bottom-left", Label: "Bottom left"}, {Value: "bottom-right", Label: "Bottom right"},
				},
			},
			{Key: "showFrametime", Label: "Show frame time", Type: core.SettingToggle, Value: true},
		},
	})
}

func newLosslessScaling() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "lossless-scaling",
		Name:        "Lossless Scaling",
		Description: "Scale game windows up to fullscreen resolution.",
		Icon:        "ArrowMaximize",
		Category:    core.CategoryGaming,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Scale foreground window", Type: core.SettingShortcut, Value: "ctrl+alt+s"},
			{
				Key: "scalingMode", Label: "Scaling mode", Type: core.SettingSelect, Value: "integer",
				Options: []core.SelectOption{
					{Value: "integer", Label: "Integer (nearest)"},
					{Value: "fsr1", Label: "FSR 1 (sharp)"},
				},
			},
		},
	})
}

func newBatteryManager() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "battery-manager",
		Name:        "Battery Manager",
		Description: "Limit battery charge level and get notified at a threshold.",
		Icon:        "BatteryCharge",
		Category:    core.CategorySystem,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "chargeLimitEnabled", Label: "Charge limit", Description: "Requires a supported hardware interface.", Type: core.SettingToggle, Value: false},
			{Key: "chargeLimit", Label: "Limit charge to", Type: core.SettingSlider, Value: 80.0, Min: f64(50), Max: f64(100), Step: f64(5)},
			{Key: "notifyEnabled", Label: "Charge notification", Description: "Notify when the battery reaches the level below.", Type: core.SettingToggle, Value: false},
			{Key: "notifyAt", Label: "Notify at", Type: core.SettingSlider, Value: 80.0, Min: f64(20), Max: f64(100), Step: f64(5)},
		},
	})
}

func newAutomations() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "automations",
		Name:        "Automations",
		Description: "Build block-based automation flows: trigger, condition, action.",
		Icon:        "Flow",
		Category:    core.CategoryAdvanced,
		Available:   false,
	})
}

func newHotkeyActions() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "hotkey-actions",
		Name:        "Hotkey Actions",
		Description: "Bind global shortcuts to actions: launch apps, run commands, toggle utilities.",
		Icon:        "KeyCommand",
		Category:    core.CategoryInput,
		Available:   false,
	})
}

func newContextMenu() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "context-menu",
		Name:        "Context Menu",
		Description: "Add custom items to the Explorer right-click menu.",
		Icon:        "MoreHorizontal",
		Category:    core.CategoryFiles,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "classicMenu", Label: "Classic menu", Description: "Show items in the legacy menu (Show more options).", Type: core.SettingToggle, Value: true},
			{Key: "modernMenu", Label: "Modern menu", Description: "Show items in the Windows 11 top-level menu. Requires an extra package component.", Type: core.SettingToggle, Value: false},
		},
	})
}

func newScreensaver() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "screensaver",
		Name:        "Screensaver",
		Description: "Custom screensaver: play videos, web pages, or image slideshows.",
		Icon:        "Tv",
		Category:    core.CategorySystem,
		Available:   false,
		Settings: []core.SettingField{
			{
				Key: "contentType", Label: "Content type", Type: core.SettingSelect, Value: "video",
				Options: []core.SelectOption{
					{Value: "video", Label: "Video"}, {Value: "web", Label: "Web page"}, {Value: "images", Label: "Image slideshow"},
				},
			},
			{Key: "source", Label: "Source", Description: "File path or URL.", Type: core.SettingText, Value: ""},
		},
	})
}

func newWindowTransparency() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "window-transparency",
		Name:        "Window Transparency",
		Description: "Make any window see-through with a shortcut.",
		Icon:        "SquareHint",
		Category:    core.CategoryWindowing,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle transparency", Type: core.SettingShortcut, Value: "super+shift+o"},
			{Key: "opacity", Label: "Opacity", Type: core.SettingSlider, Value: 80.0, Min: f64(10), Max: f64(100), Step: f64(5)},
			{Key: "excludeOwn", Label: "Skip PolyTools window", Type: core.SettingToggle, Value: true},
		},
	})
}

func newMinimizeToTray() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "minimize-to-tray",
		Name:        "Minimize To Tray",
		Description: "Send any window to the system tray.",
		Icon:        "ArrowMinimize",
		Category:    core.CategoryWindowing,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Send window to tray", Type: core.SettingShortcut, Value: "super+shift+down"},
		},
	})
}

func newTextExpander() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "text-expander",
		Name:        "Text Expander",
		Description: "Expand short abbreviations into full text as you type.",
		Icon:        "TextExpand",
		Category:    core.CategoryInput,
		Available:   false,
		Settings: []core.SettingField{
			{
				Key: "trigger", Label: "Expand trigger", Type: core.SettingSelect, Value: "tab",
				Options: []core.SelectOption{
					{Value: "instant", Label: "Instant"}, {Value: "tab", Label: "On Tab"}, {Value: "space", Label: "On Space"},
				},
			},
		},
	})
}

func newSystemInfo() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "system-info",
		Name:        "System Info",
		Description: "Show CPU, memory, GPU and system information.",
		Icon:        "DesktopPulse",
		Category:    core.CategorySystem,
		Available:   false,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Show system info", Type: core.SettingShortcut, Value: "super+shift+i"},
		},
	})
}
