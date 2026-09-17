//go:build windows

package modules

import "polytools/internal/core"

// Modules driven entirely by their settings-page panels (no background
// runtime) — Batch Rename, Environment Variables, Hosts File Editor.

func newBatchRename() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "batch-rename",
		Name:        "Batch Rename",
		Description: "Bulk-rename files with search/replace and regular expressions.",
		Icon:        "Rename",
		Category:    core.CategoryFiles,
		Available:   true,
	})
}

func newEnvVars() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "env-vars",
		Name:        "Environment Variables",
		Description: "Manage user and system environment variables.",
		Icon:        "BracesVariable",
		Category:    core.CategoryAdvanced,
		Available:   true,
	})
}

func newHostsEditor() *core.BaseModule {
	return core.NewModule(core.Info{
		Key:         "hosts-editor",
		Name:        "Hosts File Editor",
		Description: "Edit the Windows hosts file.",
		Icon:        "Globe",
		Category:    core.CategoryAdvanced,
		Available:   true,
	})
}
