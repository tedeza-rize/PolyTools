//go:build windows

package services

import (
	"polytools/internal/modules"
	"polytools/internal/win32"
)

// --- file pickers for settings panels ---

func (s *PolyToolsService) PickFile(title string) (string, error) {
	return s.app.Dialog.OpenFile().SetTitle(title).PromptForSingleSelection()
}

func (s *PolyToolsService) PickFolder(title string) (string, error) {
	return s.app.Dialog.OpenFile().SetTitle(title).
		CanChooseFiles(false).CanChooseDirectories(true).
		PromptForSingleSelection()
}

// --- Automations rules CRUD ---

func (s *PolyToolsService) AutomationRules() []modules.AutomationRule {
	return modules.LoadAutomationRules()
}

func (s *PolyToolsService) SaveAutomationRules(rules []modules.AutomationRule) error {
	if err := modules.SaveAutomationRules(rules); err != nil {
		return err
	}
	// restart the engine so triggers/hotkeys pick up changes
	_ = s.reg.SetEnabled("automations", false)
	_ = s.reg.SetEnabled("automations", true)
	return nil
}

// --- Screensaver ---

func (s *PolyToolsService) DismissScreensaver() {
	modules.ScreensaverDismiss()
}

// WinScreensaverInfo describes the built-in Windows screensaver
// configuration, plus whether PolyTools currently has it suspended.
type WinScreensaverInfo struct {
	Active    bool   `json:"active"`
	Timeout   uint32 `json:"timeout"` // seconds
	ScrPath   string `json:"scrPath"`
	Suspended bool   `json:"suspended"`
}

func (s *PolyToolsService) WindowsScreensaver() WinScreensaverInfo {
	return WinScreensaverInfo{
		Active:    win32.ScreensaverActive(),
		Timeout:   win32.ScreensaverTimeout(),
		ScrPath:   win32.ScreensaverPath(),
		Suspended: modules.WinScreensaverSuspended(),
	}
}
