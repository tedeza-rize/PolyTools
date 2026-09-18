//go:build windows

package services

import (
	"polytools/internal/modules"
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
