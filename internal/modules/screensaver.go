//go:build windows

package modules

import (
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- Screensaver: after N idle minutes, open a fullscreen borderless
// window rendering the configured content (video / web page / slideshow).
// Any input dismisses it. ---

var ScreensaverWindow *application.WebviewWindow // shown flag for service

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
			{Key: "source", Label: "Source", Description: "File path or URL. For slideshows: a folder path.", Type: core.SettingText, Value: ""},
			{Key: "idleMinutes", Label: "Start after idle", Type: core.SettingSlider, Value: 5.0, Min: f64(1), Max: f64(60), Step: f64(1)},
			{Key: "hotkey", Label: "Preview now", Type: core.SettingShortcut, Value: "super+shift+s"},
		},
	})

	var (
		mu      sync.Mutex
		saver   *application.WebviewWindow
		stopCh  chan struct{}
	)

	show := func() {
		mu.Lock()
		if saver != nil {
			mu.Unlock()
			return
		}
		mu.Unlock()
		mr := win32.MonitorRect(win32.ForegroundWindow())
		q := url.Values{}
		q.Set("type", m.SettingString("contentType"))
		q.Set("src", m.SettingString("source"))
		w := app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:     "PolyTools Screensaver",
			Frameless: true,
			URL:       "/?page=screensaver&" + q.Encode(),
			X:         int(mr.Left),
			Y:         int(mr.Top),
			Width:     int(mr.Width()),
			Height:    int(mr.Height()),
			Windows: application.WindowsWindow{
				BackdropType: application.None,
			},
		})
		mu.Lock()
		saver = w
		ScreensaverWindow = w
		mu.Unlock()
		log.Printf("[screensaver] shown")
	}

	dismiss := func() {
		mu.Lock()
		w := saver
		saver = nil
		ScreensaverWindow = nil
		mu.Unlock()
		if w != nil {
			w.Close()
		}
	}

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
			mu.Unlock()
			if shown && idle < 2 {
				// input resumed → dismiss
				dismiss()
			} else if !shown && threshold > 0 && idle >= threshold {
				show()
			}
		}
	}

	return m.WithHandlers(func() error {
		if err := hk.register(); err != nil {
			return err
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

// helper for service: currently-showing flag
func ScreensaverShowing() bool { return ScreensaverWindow != nil }
func ScreensaverDismiss() {
	if ScreensaverWindow != nil {
		ScreensaverWindow.Close()
		ScreensaverWindow = nil
	}
}
