//go:build windows

package modules

import (
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Screensaver: after N idle minutes, open a fullscreen borderless
// window rendering the configured content (video / web page / slideshow).
// Any input dismisses it. ---

// screensaverDismiss is wired to the module's dismiss closure so the
// service can close the window through the module's own bookkeeping
// (clearing the shown flag, not just the window handle).
var screensaverDismiss func()

// sourceKeys maps each content type to the setting key holding its source.
// "source" is the legacy single key — kept as a hidden field so older
// configs still resolve.
var sourceKeys = map[string]string{
	"video":  "videoFile",
	"web":    "webUrl",
	"images": "imageFolder",
}

func newScreensaver(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "screensaver",
		Name:        "Screensaver",
		Description: "Custom screensaver: play videos, web pages, or image slideshows.",
		Icon:        "Tv",
		Category:    core.CategorySystem,
		Available:   true,
		Settings: []core.SettingField{
			{
				Key: "contentType", Label: "Content type", Type: core.SettingSelect, Value: "video",
				Options: []core.SelectOption{
					{Value: "video", Label: "Video"}, {Value: "web", Label: "Web page"}, {Value: "images", Label: "Image slideshow"},
				},
			},
			{
				Key: "webUrl", Label: "URL", Type: core.SettingText, Value: "",
				Description: "Web page to display.",
				ShowIf:      &core.ShowIf{Key: "contentType", Equals: "web"},
			},
			{
				Key: "videoFile", Label: "Video file", Type: core.SettingFile, Value: "",
				Description: "Video to loop.",
				ShowIf:      &core.ShowIf{Key: "contentType", Equals: "video"},
			},
			{
				Key: "imageFolder", Label: "Image folder", Type: core.SettingFolder, Value: "",
				Description: "Folder of images for the slideshow.",
				ShowIf:      &core.ShowIf{Key: "contentType", Equals: "images"},
			},
			// Legacy storage key — superseded by the per-type keys above
			// but still read as a fallback for older configurations.
			{Key: "source", Type: core.SettingText, Value: "", Hidden: true},
			{Key: "idleMinutes", Label: "Start after idle", Type: core.SettingSlider, Value: 5.0, Min: f64(1), Max: f64(60), Step: f64(1)},
			{Key: "hotkey", Label: "Preview now", Type: core.SettingShortcut, Value: "ctrl+alt+shift+s"},
		},
	})

	var (
		mu      sync.Mutex
		saver   *application.WebviewWindow
		shownAt time.Time
		stopCh  chan struct{}
	)

	source := func() string {
		s := m.SettingString(sourceKeys[m.SettingString("contentType")])
		if s == "" {
			s = m.SettingString("source")
		}
		return s
	}

	show := func() {
		mu.Lock()
		if saver != nil {
			mu.Unlock()
			return
		}
		mu.Unlock()
		contentType := m.SettingString("contentType")
		src := source()
		q := url.Values{}
		q.Set("type", contentType)
		q.Set("src", src)
		target := "/?page=screensaver&" + q.Encode()
		// Web content opens as a top-level navigation, not an iframe —
		// sites with CSP frame-ancestors / X-Frame-Options refuse to load
		// inside an embedded frame.
		if contentType == "web" &&
			(strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")) {
			target = src
		}
		opts := application.WebviewWindowOptions{
			Title:       "PolyTools Screensaver",
			Frameless:   true,
			AlwaysOnTop: true,
			URL:         target,
			Windows: application.WindowsWindow{
				BackdropType:    application.None,
				HiddenOnTaskbar: true,
			},
		}
		// Cover the monitor holding the foreground window. Window options
		// take DIP coordinates while MonitorRect reports physical pixels,
		// so look up the matching screen and use its (already-DIP) bounds.
		mr := win32.MonitorRect(win32.ForegroundWindow())
		if screen := app.Screen.ScreenNearestPhysicalPoint(application.Point{
			X: int(mr.Left + mr.Width()/2),
			Y: int(mr.Top + mr.Height()/2),
		}); screen != nil {
			opts.X, opts.Y = screen.Bounds.X, screen.Bounds.Y
			opts.Width, opts.Height = screen.Bounds.Width, screen.Bounds.Height
		} else {
			opts.X, opts.Y = int(mr.Left), int(mr.Top)
			opts.Width, opts.Height = int(mr.Width()), int(mr.Height())
		}
		w := app.Window.NewWithOptions(opts)
		// If the window is closed by anything other than dismiss() (user,
		// OS, crash), clear the handle so the module doesn't stay stuck in
		// the "shown" state forever.
		w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
			mu.Lock()
			if saver == w {
				saver = nil
			}
			mu.Unlock()
		})
		mu.Lock()
		saver = w
		shownAt = time.Now()
		mu.Unlock()
		log.Printf("[screensaver] shown")
	}

	dismiss := func() {
		mu.Lock()
		w := saver
		saver = nil
		mu.Unlock()
		if w != nil {
			w.Close()
		}
	}
	screensaverDismiss = dismiss

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		mu.Lock()
		shown := saver != nil
		mu.Unlock()
		if shown {
			dismiss()
		} else {
			show()
		}
	})

	watch := func() {
		threshold := uint32(m.SettingFloat("idleMinutes") * 60)
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(2 * time.Second):
			}
			idle := win32.IdleSeconds()
			mu.Lock()
			shown := saver != nil
			up := time.Since(shownAt)
			mu.Unlock()
			// Dismiss on input only after a short grace period — the
			// preview hotkey press itself resets the idle counter, which
			// would otherwise kill the window on the next tick.
			if shown && idle < 2 && up > 3*time.Second {
				dismiss()
			} else if !shown && threshold > 0 && idle >= threshold {
				show()
			}
		}
	}

	return m.WithHandlers(func() error {
		// The preview hotkey is auxiliary — if the combo is taken by
		// another app, still enable the idle watcher.
		if err := hk.register(); err != nil {
			log.Printf("[screensaver] preview hotkey not registered: %v", err)
		}
		stopCh = make(chan struct{})
		go watch()
		return nil
	}, func() error {
		close(stopCh)
		dismiss()
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}

// ScreensaverDismiss closes the screensaver window (called from the service).
func ScreensaverDismiss() {
	if screensaverDismiss != nil {
		screensaverDismiss()
	}
}
