package core

import (
	"fmt"
	"log"
	"sync"
)

// Registry owns the module list, applies persisted state, and emits
// change notifications through a callback (wired to Wails events in main).
type Registry struct {
	mu        sync.RWMutex
	mods      []Module
	byKey     map[string]Module
	store     *Store
	onChanged func(key string)
}

func NewRegistry(store *Store) *Registry {
	return &Registry{byKey: map[string]Module{}, store: store}
}

func (r *Registry) OnChanged(f func(key string)) { r.onChanged = f }

func (r *Registry) Register(m Module) {
	r.mu.Lock()
	defer r.mu.Unlock()
	info := m.Info()
	if _, exists := r.byKey[info.Key]; exists {
		panic(fmt.Sprintf("duplicate module key %q", info.Key))
	}
	r.mods = append(r.mods, m)
	r.byKey[info.Key] = m
}

// ApplyPersisted restores saved enabled flags + setting values into modules
// (without side effects). Call Boot() afterwards to actually enable modules.
func (r *Registry) ApplyPersisted() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.mods {
		bm, ok := m.(*BaseModule)
		if !ok {
			continue
		}
		if enabled, settings, found := r.store.ModuleState(m.Info().Key); found {
			bm.ApplyState(enabled, settings)
		}
	}
}

// Boot enables every module marked enabled. Call after the app + services
// exist so callbacks (hotkeys, hooks) have a live runtime.
func (r *Registry) Boot() {
	for _, m := range r.snapshot() {
		if m.Info().Enabled {
			if err := m.SetEnabled(true); err != nil {
				log.Printf("module %s failed to enable: %v", m.Info().Key, err)
				if bm, ok := m.(*BaseModule); ok {
					bm.info.Enabled = false
				}
			}
		}
	}
}

func (r *Registry) snapshot() []Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Module(nil), r.mods...)
}

// Modules returns metadata for the settings UI, in registration order.
func (r *Registry) Modules() []Info {
	mods := r.snapshot()
	infos := make([]Info, 0, len(mods))
	for _, m := range mods {
		infos = append(infos, m.Info())
	}
	return infos
}

func (r *Registry) SetEnabled(key string, enabled bool) error {
	r.mu.RLock()
	m, ok := r.byKey[key]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown module %q", key)
	}
	if err := m.SetEnabled(enabled); err != nil {
		return err
	}
	r.store.SetModuleState(key, enabled, nil)
	_ = r.store.Save()
	if r.onChanged != nil {
		r.onChanged(key)
	}
	return nil
}

func (r *Registry) UpdateSetting(key, field string, value any) error {
	r.mu.RLock()
	m, ok := r.byKey[key]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown module %q", key)
	}
	if err := m.UpdateSetting(field, value); err != nil {
		return err
	}
	r.store.SetModuleState(key, m.Info().Enabled, map[string]any{field: value})
	_ = r.store.Save()
	return nil
}

func (r *Registry) General() General { return r.store.General() }

func (r *Registry) SetGeneral(g General) {
	r.store.SetGeneral(g)
	_ = r.store.Save()
}
