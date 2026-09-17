package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"polytools/internal/core"
	"polytools/internal/modules"
	"polytools/internal/services"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIcon []byte

func main() {
	store, err := core.NewStore()
	if err != nil {
		log.Fatal(err)
	}
	reg := core.NewRegistry(store)
	svc := &services.PolyToolsService{}

	app := application.New(application.Options{
		Name:        "PolyTools",
		Description: "Windows utilities for power users",
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.polytools.app",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				svc.ShowWindow()
			},
			ExitCode: 0,
		},
	})

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:    "PolyTools",
		Width:    1180,
		Height:   780,
		MinWidth: 900,
		MinHeight: 560,
		Frameless: true,
		// Translucent so the Mica backdrop shows through the webview.
		BackgroundType: application.BackgroundTypeTranslucent,
		Windows: application.WindowsWindow{
			BackdropType:           application.Mica,
			NonClientRegionSupport: true,
		},
		URL: "/",
	})
	svc.Attach(app, reg, win)

	// System tray: click toggles the settings window, right-click shows menu.
	tray := app.SystemTray.New()
	tray.SetIcon(trayIcon)
	tray.SetTooltip("PolyTools")
	trayMenu := application.NewMenu()
	trayMenu.Add("Open PolyTools").OnClick(func(*application.Context) { svc.ShowWindow() })
	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(*application.Context) { app.Quit() })
	tray.SetMenu(trayMenu)
	tray.OnClick(func() {
		if win.IsVisible() {
			svc.HideWindow()
		} else {
			svc.ShowWindow()
		}
	})

	// Closing the window hides to the tray instead of quitting.
	win.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		e.Cancel()
		win.Hide()
	})

	// Notify the frontend when a module's enabled state changes.
	reg.OnChanged(func(key string) {
		app.Event.Emit("modules:changed", key)
	})

	modules.RegisterAll(reg, app, win, tray, svc.ShowWindow)
	reg.ApplyPersisted()
	reg.Boot()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
