//go:build windows

package modules

import "polytools/internal/core"

// Coming-soon modules: metadata only (Available=false), so the settings UI
// shows the full catalog while implementations land later.

func f64(v float64) *float64 { return &v }

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
