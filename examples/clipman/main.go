package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/huanfeng/wind-ui/examples/clipman/store"
	"github.com/huanfeng/wind-ui/platform"
	pwin "github.com/huanfeng/wind-ui/platform/windows"
)

const (
	maxTrayRecentItems = 10
	hotkeyIDQuickPanel = "clipman_quick"
)

// ClipManApp is the main application controller.
type ClipManApp struct {
	plat    platform.Platform
	store   *store.Store
	tray    platform.TrayIcon
	clipMon platform.ClipboardMonitor
	hotkeys platform.HotkeyManager

	// Windows
	managerWindow platform.Window
	managerUI     *managerUI
	quickPanel    *quickPanel
}

func main() {
	runtime.LockOSThread()

	app, err := newClipManApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start ClipMan: %v\n", err)
		os.Exit(1)
	}
	defer app.close()

	app.run()
}

func newClipManApp() (*ClipManApp, error) {
	plat := pwin.NewPlatform()

	s, err := store.New("")
	if err != nil {
		return nil, fmt.Errorf("init store: %w", err)
	}

	app := &ClipManApp{
		plat:  plat,
		store: s,
	}

	app.quickPanel = newQuickPanel(app)

	app.setupClipboardMonitor()
	app.setupTray()
	app.setupHotkeys()

	return app, nil
}

func (a *ClipManApp) setupClipboardMonitor() {
	a.clipMon = a.plat.CreateClipboardMonitor()
	a.clipMon.SetOnClipboardChanged(func(content platform.ClipboardContent) {
		if content.Text == "" {
			return
		}
		a.store.Add(content.Text, content.Source)
		a.updateTrayMenu()
	})
	if err := a.clipMon.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "clipboard monitor start failed: %v\n", err)
	}
}

func (a *ClipManApp) setupTray() {
	a.tray = a.plat.CreateTrayIcon()
	a.tray.SetTooltip("ClipMan - 剪贴板管理器")

	a.tray.SetOnClick(func() {
		a.showManagerWindow()
	})
	a.tray.SetOnDoubleClick(func() {
		a.showManagerWindow()
	})

	a.updateTrayMenu()
}

func (a *ClipManApp) setupHotkeys() {
	a.hotkeys = a.plat.CreateHotkeyManager()

	cfg := a.store.GetConfig()
	err := a.hotkeys.Register(
		hotkeyIDQuickPanel,
		platform.KeyModifier(cfg.HotkeyMod),
		cfg.HotkeyKey,
		func() {
			a.quickPanel.show()
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hotkey register failed: %v\n", err)
	}
}

func (a *ClipManApp) updateTrayMenu() {
	recent := a.store.GetRecent(maxTrayRecentItems)

	var items []platform.TrayMenuItem

	// Recent clipboard entries
	for _, entry := range recent {
		e := entry // capture
		label := truncateText(e.Text, 40)
		if e.Pinned {
			label = "★ " + label
		}
		items = append(items, platform.TrayMenuItem{
			ID:    e.ID,
			Title: label,
			OnClick: func() {
				a.store.Use(e.ID)
				// Copy to clipboard
				clip := a.plat.GetClipboard()
				clip.SetText(e.Text)
			},
		})
	}

	if len(items) > 0 {
		items = append(items, platform.TrayMenuItem{IsSeparator: true})
	}

	// Action items
	items = append(items,
		platform.TrayMenuItem{
			ID:    "show",
			Title: "打开管理窗口",
			OnClick: func() {
				a.showManagerWindow()
			},
		},
		platform.TrayMenuItem{
			ID:    "clear",
			Title: "清空历史",
			OnClick: func() {
				a.store.ClearAll()
				a.updateTrayMenu()
			},
		},
		platform.TrayMenuItem{IsSeparator: true},
		platform.TrayMenuItem{
			ID:    "exit",
			Title: "退出",
			OnClick: func() {
				a.plat.Quit()
			},
		},
	)

	a.tray.SetMenu(&platform.TrayMenu{Items: items})
}

func (a *ClipManApp) showManagerWindow() {
	if a.managerWindow != nil && a.managerWindow.IsVisible() {
		// Already visible, just bring to front
		return
	}

	if a.managerWindow == nil {
		w, err := a.plat.CreateWindow(platform.WindowOptions{
			Title:     "ClipMan - 剪贴板管理器",
			Width:     800,
			Height:    600,
			MinWidth:  400,
			MinHeight: 300,
			Resizable: true,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "create manager window failed: %v\n", err)
			return
		}
		a.managerWindow = w

		a.managerUI = newManagerUI(a, w)

		w.SetOnClose(func() bool {
			w.Hide()
			// GC when window is hidden to free rendering buffers
			runtime.GC()
			return false // don't destroy, just hide
		})
	}

	a.managerWindow.Center()
	a.managerWindow.Show()
}

func (a *ClipManApp) run() {
	a.plat.RunMainLoop()
}

func (a *ClipManApp) close() {
	if a.clipMon != nil {
		a.clipMon.Stop()
	}
	if a.hotkeys != nil {
		a.hotkeys.UnregisterAll()
	}
	if a.tray != nil {
		a.tray.Destroy()
	}
	if a.store != nil {
		a.store.Close()
	}
}

// truncateText shortens text to maxLen characters, replacing newlines with spaces.
func truncateText(text string, maxLen int) string {
	// Replace newlines with spaces for display
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)

	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen-1]) + "…"
	}
	return text
}
