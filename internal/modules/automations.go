//go:build windows

package modules

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Automations: trigger → steps flows. v1 supports hotkey / interval /
// startup triggers and run / open / type / keys / delay / beep steps.
// Rules are edited on the module page and stored in automations.json. ---

type AutomationRule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Trigger struct {
		Type  string `json:"type"`  // hotkey | interval | startup
		Value string `json:"value"` // accel combo or minutes
	} `json:"trigger"`
	Steps []AutomationStep `json:"steps"`
}

type AutomationStep struct {
	Type  string `json:"type"`  // run | open | type | keys | delay | beep
	Value string `json:"value"` // command / path-or-url / text / key combo / ms
}

func automationsPath() string { return filepath.Join(stateDir(), "automations.json") }

// LoadAutomationRules reads the rules file (service-facing).
func LoadAutomationRules() []AutomationRule {
	raw, err := os.ReadFile(automationsPath())
	if err != nil {
		return nil
	}
	var rules []AutomationRule
	if json.Unmarshal(raw, &rules) != nil {
		return nil
	}
	return rules
}

// SaveAutomationRules writes the rules file (service-facing).
func SaveAutomationRules(rules []AutomationRule) error {
	raw, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(automationsPath(), raw, 0o644)
}

func newAutomations(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "automations",
		Name:        "Automations",
		Description: "Build block-based automation flows: trigger, condition, action.",
		Icon:        "Flow",
		Category:    core.CategoryAdvanced,
		Available:   true,
	})

	var (
		mu       sync.Mutex
		hotkeys  []string
		tickers  []*time.Ticker
		stopTick chan struct{}
	)

	runRule := func(r AutomationRule) {
		log.Printf("[automations] running %q", r.Name)
		for _, step := range r.Steps {
			switch step.Type {
			case "run":
				c := exec.Command("cmd", "/c", step.Value)
				c.SysProcAttr = sysProcAttrHide
				_ = c.Start()
			case "open":
				c := exec.Command("cmd", "/c", "start", "", step.Value)
				c.SysProcAttr = sysProcAttrHide
				_ = c.Start()
			case "type":
				win32.SendText(step.Value)
			case "keys":
				var vks []uint16
				for _, part := range strings.Split(step.Value, "+") {
					if vk, ok := win32.VKByName[strings.ToLower(strings.TrimSpace(part))]; ok {
						vks = append(vks, vk)
					}
				}
				if len(vks) > 0 {
					win32.SendKeys(vks...)
				}
			case "delay":
				if ms, err := time.ParseDuration(step.Value + "ms"); err == nil {
					time.Sleep(ms)
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			case "beep":
				win32.Beep()
			}
		}
	}

	start := func() {
		rules := LoadAutomationRules()
		stopTick = make(chan struct{})
		for _, r := range rules {
			r := r
			if !r.Enabled {
				continue
			}
			switch r.Trigger.Type {
			case "hotkey":
				combo := r.Trigger.Value
				if err := app.GlobalShortcut.Register(combo, func() { go runRule(r) }); err == nil {
					mu.Lock()
					hotkeys = append(hotkeys, combo)
					mu.Unlock()
				}
			case "interval":
				min := 5.0
				if v, err := time.ParseDuration(r.Trigger.Value + "m"); err == nil {
					tk := time.NewTicker(v)
					tickers = append(tickers, tk)
					go func() {
						for {
							select {
							case <-stopTick:
								return
							case <-tk.C:
								go runRule(r)
							}
						}
					}()
				} else {
					_ = min
				}
			case "startup":
				go runRule(r)
			}
		}
	}

	stop := func() {
		mu.Lock()
		defer mu.Unlock()
		for _, c := range hotkeys {
			_ = app.GlobalShortcut.Unregister(c)
		}
		hotkeys = nil
		for _, t := range tickers {
			t.Stop()
		}
		tickers = nil
		if stopTick != nil {
			select {
			case <-stopTick:
			default:
				close(stopTick)
			}
			stopTick = nil
		}
	}

	return m.WithHandlers(func() error {
		start()
		return nil
	}, func() error {
		stop()
		return nil
	})
}
