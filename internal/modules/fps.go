//go:build windows

package modules

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	etw "github.com/0xrawsec/golang-etw/etw"
	"github.com/wailsapp/wails/v3/pkg/application"
	"polytools/internal/core"
	"polytools/internal/win32"
)

// --- FPS Overlay: counts DXGI Present events (PresentMon-style ETW) for the
// foreground process and renders the rate in a click-through overlay. ---

func newFpsOverlay(app *application.App) *core.BaseModule {
	m := core.NewModule(core.Info{
		Key:         "fps-overlay",
		Name:        "FPS Overlay",
		Description: "Show the frame rate of the current game as an overlay.",
		Icon:        "Gauge",
		Category:    core.CategoryGaming,
		Available:   true,
		Settings: []core.SettingField{
			{Key: "hotkey", Label: "Toggle overlay", Type: core.SettingShortcut, Value: "super+shift+f"},
			{
				Key: "displayStyle", Label: "Display style", Type: core.SettingSelect, Value: "number",
				Options: []core.SelectOption{
					{Value: "number", Label: "Number only"},
					{Value: "graph", Label: "Graph only"},
					{Value: "number-graph", Label: "Number + graph"},
				},
			},
			{
				Key: "position", Label: "Position", Type: core.SettingSelect, Value: "top-left",
				Options: []core.SelectOption{
					{Value: "top-left", Label: "Top left"}, {Value: "top-right", Label: "Top right"},
					{Value: "bottom-left", Label: "Bottom left"}, {Value: "bottom-right", Label: "Bottom right"},
				},
			},
			{Key: "showFrametime", Label: "Show frame time", Type: core.SettingToggle, Value: true},
		},
	})

	var (
		mu      sync.Mutex
		running bool
		overlay *win32.Overlay
		cancel  context.CancelFunc
		session *etw.RealTimeSession
		// ring buffer of present-event timestamps for the tracked process
		stamps   []time.Time
		targetPid uint32
		history  []int // recent fps values for graph mode
	)

	trackForeground := func() uint32 {
		hwnd := win32.ForegroundWindow()
		if hwnd == 0 {
			return 0
		}
		var pid uint32
		win32.WindowPid(hwnd, &pid)
		return pid
	}

	fpsNow := func() int {
		mu.Lock()
		defer mu.Unlock()
		cut := time.Now().Add(-1 * time.Second)
		keep := stamps[:0]
		for _, ts := range stamps {
			if ts.After(cut) {
				keep = append(keep, ts)
			}
		}
		stamps = keep
		return len(stamps)
	}

	render := func(fps int) {
		mu.Lock()
		o := overlay
		history = append(history, fps)
		if len(history) > 60 {
			history = history[1:]
		}
		hist := append([]int(nil), history...)
		mu.Unlock()
		if o == nil {
			return
		}
		style := m.SettingString("displayStyle")
		o.Fill(12, 12, 12, 170)
		y := int32(8)
		if style != "graph" {
			o.Text(fmt.Sprintf("%d FPS", fps), 10, y, 22)
			y += 24
			if m.SettingBool("showFrametime") && fps > 0 {
				o.Text(fmt.Sprintf("%.1f ms", 1000.0/float64(fps)), 10, y, 13)
			}
		}
		if style != "number" {
			// mini bar graph along the bottom
			max := 1
			for _, v := range hist {
				if v > max {
					max = v
				}
			}
			gw := int32(180) / int32(len(hist))
			if gw < 2 {
				gw = 2
			}
			for i, v := range hist {
				bh := int32(float64(v) / float64(max) * 20)
				if bh < 1 {
					bh = 1
				}
				o.FillRect(int32(i)*gw, 56-bh, gw-1, bh, 30, 180, 60, 255)
			}
		}
		o.Present()
	}

	startPipeline := func() {
		ctx, cn := context.WithCancel(context.Background())
		cancel = cn

		session = etw.NewRealTimeSession("PolyToolsFPS")
		prov := etw.MustParseProvider("Microsoft-Windows-DXGI")
		if err := session.EnableProvider(prov); err != nil {
			log.Printf("[fps] enable DXGI provider: %v", err)
		}
		if err := session.Start(); err != nil {
			log.Printf("[fps] session start: %v", err)
		}

		consumer := etw.NewRealTimeConsumer(ctx)
		consumer.FromSessions(session)
		consumer.EventCallback = func(e *etw.Event) error {
			if e.System.Execution.ProcessID != targetPid {
				return nil
			}
			task := strings.ToLower(e.System.Task.Name)
			op := strings.ToLower(e.System.Opcode.Name)
			if strings.Contains(task, "present") &&
				(op == "stop" || op == "" || strings.Contains(task, "present")) {
				mu.Lock()
				stamps = append(stamps, time.Now())
				mu.Unlock()
			}
			return nil
		}
		if err := consumer.Start(); err != nil {
			log.Printf("[fps] consumer start: %v", err)
		}
		go func() {
			for e := range consumer.Events {
				_ = e // EventCallback already handled
			}
		}()

		// Refocus tracking + render loop (~4 updates/sec).
		go func() {
			tick := time.NewTicker(250 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-tick.C:
					if p := trackForeground(); p != 0 {
						targetPid = p
					}
					render(fpsNow())
				}
			}
		}()
	}

	stopPipeline := func() {
		if cancel != nil {
			cancel()
			cancel = nil
		}
		if session != nil {
			_ = session.Stop()
			session = nil
		}
		mu.Lock()
		if overlay != nil {
			overlay.Destroy()
			overlay = nil
		}
		stamps = nil
		history = nil
		mu.Unlock()
	}

	positionOverlay := func() (int32, int32) {
		mr := win32.MonitorRect(win32.ForegroundWindow())
		const w, h, pad = 190, 60, 12
		switch m.SettingString("position") {
		case "top-right":
			return mr.Right - w - pad, mr.Top + pad
		case "bottom-left":
			return mr.Left + pad, mr.Bottom - h - pad
		case "bottom-right":
			return mr.Right - w - pad, mr.Bottom - h - pad
		default:
			return mr.Left + pad, mr.Top + pad
		}
	}

	hk := newHotkeyCtl(app, func() string { return m.SettingString("hotkey") }, func() {
		mu.Lock()
		on := running
		mu.Unlock()
		if on {
			stopPipeline()
			mu.Lock()
			running = false
			mu.Unlock()
			return
		}
		x, y := positionOverlay()
		o, err := win32.NewOverlay(x, y, 190, 60)
		if err != nil {
			log.Printf("[fps] overlay: %v", err)
			return
		}
		mu.Lock()
		overlay = o
		running = true
		mu.Unlock()
		startPipeline()
	})

	return m.WithHandlers(hk.register, func() error {
		stopPipeline()
		return hk.unregister()
	}).WithSettingHandler(func(key string, _ any) error {
		if key != "hotkey" || !m.Info().Enabled {
			return nil
		}
		return hk.rebind()
	})
}
