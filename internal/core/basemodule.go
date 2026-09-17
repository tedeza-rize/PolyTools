package core

import "fmt"

// BaseModule carries module state (enabled flag + setting values) and
// delegates the real work to optional callbacks. Most modules can be built
// with NewModule + WithHandlers instead of implementing Module by hand.
type BaseModule struct {
	info      Info
	onEnable  func() error
	onDisable func() error
	onSetting func(key string, value any) error
}

func NewModule(info Info) *BaseModule {
	return &BaseModule{info: info}
}

func (m *BaseModule) WithHandlers(onEnable, onDisable func() error) *BaseModule {
	m.onEnable = onEnable
	m.onDisable = onDisable
	return m
}

func (m *BaseModule) WithSettingHandler(f func(key string, value any) error) *BaseModule {
	m.onSetting = f
	return m
}

func (m *BaseModule) Info() Info { return m.info }

func (m *BaseModule) SetEnabled(enabled bool) error {
	if !m.info.Available && enabled {
		return fmt.Errorf("module %q is not available yet", m.info.Key)
	}
	var err error
	if enabled && m.onEnable != nil {
		err = m.onEnable()
	} else if !enabled && m.onDisable != nil {
		err = m.onDisable()
	}
	if err != nil {
		return err
	}
	m.info.Enabled = enabled
	return nil
}

func (m *BaseModule) UpdateSetting(key string, value any) error {
	for i := range m.info.Settings {
		if m.info.Settings[i].Key == key {
			m.info.Settings[i].Value = value
			if m.onSetting != nil {
				return m.onSetting(key, value)
			}
			return nil
		}
	}
	return fmt.Errorf("module %q: unknown setting %q", m.info.Key, key)
}

// Setting returns the current value of a setting field.
func (m *BaseModule) Setting(key string) any {
	for _, s := range m.info.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return nil
}

func (m *BaseModule) SettingBool(key string) bool {
	v, _ := m.Setting(key).(bool)
	return v
}

func (m *BaseModule) SettingString(key string) string {
	v, _ := m.Setting(key).(string)
	return v
}

func (m *BaseModule) SettingFloat(key string) float64 {
	switch v := m.Setting(key).(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0
}

// ApplyState restores persisted state without running enable/disable side
// effects. Boot() calls SetEnabled afterwards for modules that were enabled.
func (m *BaseModule) ApplyState(enabled bool, settings map[string]any) {
	m.info.Enabled = enabled
	for k, v := range settings {
		for i := range m.info.Settings {
			if m.info.Settings[i].Key == k {
				m.info.Settings[i].Value = v
			}
		}
	}
}
