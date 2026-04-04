package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/examples/clipman/store"
	"github.com/huanfeng/wind-ui/layout"
	"github.com/huanfeng/wind-ui/platform"
	"github.com/huanfeng/wind-ui/widget"
)

const quickPanelMaxItems = 15

type quickPanel struct {
	app     *ClipManApp
	window  platform.Window
	list    *widget.RecyclerView
	adapter *quickAdapter
	created bool
	visible bool

	previousHwnd uintptr
	dismissTimer *time.Ticker
	dismissDone  chan struct{}
}

type quickAdapter struct {
	panel   *quickPanel
	entries []store.ClipEntry
}

func (qa *quickAdapter) GetItemCount() int           { return len(qa.entries) }
func (qa *quickAdapter) GetItemViewType(pos int) int { return 0 }

func (qa *quickAdapter) CreateViewHolder(viewType int) *widget.ViewHolder {
	item := NewClipItemWidget()
	return &widget.ViewHolder{ItemView: item}
}

func (qa *quickAdapter) BindViewHolder(holder *widget.ViewHolder, position int) {
	if position >= len(qa.entries) {
		return
	}
	entry := qa.entries[position]
	if item, ok := holder.ItemView.(*ClipItemWidget); ok {
		preview := truncateText(entry.Text, 50)
		source := entry.Source
		if source == "" {
			source = ""
		}
		item.SetContent(preview, source, formatTime(entry.CreatedAt), entry.Pinned)
	}
}

func newQuickPanel(app *ClipManApp) *quickPanel {
	return &quickPanel{app: app}
}

func (qp *quickPanel) show() {
	if qp.visible {
		qp.hide()
		return
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	getFG := user32.NewProc("GetForegroundWindow")
	hwnd, _, _ := getFG.Call()
	qp.previousHwnd = hwnd

	if !qp.created {
		qp.createWindow()
	}

	qp.adapter.entries = qp.app.store.GetRecent(quickPanelMaxItems)
	qp.list.NotifyDataSetChanged()

	x, y := qp.getCaretScreenPos()
	qp.window.SetPosition(x, y)
	qp.window.Show()
	qp.visible = true

	qp.startDismissWatch()
}

func (qp *quickPanel) hide() {
	qp.stopDismissWatch()
	if qp.window != nil {
		qp.window.Hide()
	}
	qp.visible = false
	runtime.GC()
}

func (qp *quickPanel) getCaretScreenPos() (int, int) {
	user32 := syscall.NewLazyDLL("user32.dll")
	getWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")
	getGUIThreadInfo := user32.NewProc("GetGUIThreadInfo")
	clientToScreen := user32.NewProc("ClientToScreen")

	if qp.previousHwnd != 0 {
		tid, _, _ := getWindowThreadProcessId.Call(qp.previousHwnd, 0)
		if tid != 0 {
			type GUITHREADINFO struct {
				CbSize        uint32
				Flags         uint32
				HwndActive    uintptr
				HwndFocus     uintptr
				HwndCapture   uintptr
				HwndMenuOwner uintptr
				HwndMoveSize  uintptr
				HwndCaret     uintptr
				RcCaret       struct{ Left, Top, Right, Bottom int32 }
			}
			var gti GUITHREADINFO
			gti.CbSize = uint32(unsafe.Sizeof(gti))
			ret, _, _ := getGUIThreadInfo.Call(tid, uintptr(unsafe.Pointer(&gti)))
			if ret != 0 && gti.HwndCaret != 0 {
				pt := struct{ X, Y int32 }{gti.RcCaret.Left, gti.RcCaret.Bottom}
				clientToScreen.Call(gti.HwndCaret, uintptr(unsafe.Pointer(&pt)))
				return int(pt.X), int(pt.Y) + 4
			}
		}
	}

	getCursorPos := user32.NewProc("GetCursorPos")
	var pt struct{ X, Y int32 }
	getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X) - 160, int(pt.Y)
}

func (qp *quickPanel) createWindow() {
	w, err := qp.app.plat.CreateWindow(platform.WindowOptions{
		Title:      "ClipMan Quick",
		Width:      360,
		Height:     440,
		Frameless:  true,
		TopMost:    true,
		NoActivate: true,
	})
	if err != nil {
		fmt.Printf("quick panel create failed: %v\n", err)
		return
	}
	qp.window = w
	qp.created = true

	root := qp.buildUI()
	w.SetContentView(root)
}

func (qp *quickPanel) buildUI() *core.Node {
	root := newContainerNode(layout.Vertical, 0)
	root.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#F3F3F3"),
		BorderColor:     core.ParseColor("#E0E0E0"),
		BorderWidth:     1,
		CornerRadius:    12,
	})

	// Header: "剪贴板" title
	headerBar := newContainerNode(layout.Horizontal, 0)
	headerBar.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 36, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#F3F3F3"),
	})
	headerBar.SetPadding(core.Insets{Left: 16, Top: 8, Right: 16, Bottom: 4})

	headerTitle := widget.NewTextView("剪贴板")
	headerTitle.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		TextColor: core.ParseColor("#1A1A1A"),
		FontSize:  14,
	})
	headerBar.AddChild(headerTitle.Node())
	root.AddChild(headerBar)

	// Card list
	dpiScale := qp.app.plat.GetPrimaryScreen().DPI / 96.0
	if dpiScale < 1.0 {
		dpiScale = 1.0
	}
	qp.adapter = &quickAdapter{panel: qp}
	qp.list = widget.NewRecyclerView(68 * dpiScale) // card items, DPI-scaled
	qp.list.SetAdapter(qp.adapter)
	qp.list.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#F3F3F3"),
		Weight:          1,
	})

	qp.list.SetOnItemClickListener(func(position int) {
		if position >= 0 && position < len(qp.adapter.entries) {
			qp.selectEntry(qp.adapter.entries[position])
		}
	})
	qp.list.SetOnItemRightClickListener(func(position int, screenX, screenY int) {
		if position >= 0 && position < len(qp.adapter.entries) {
			entry := qp.adapter.entries[position]
			pinLabel := "★ 收藏"
			if entry.Pinned {
				pinLabel = "取消收藏"
			}
			items := []platform.TrayMenuItem{
				{Title: "粘贴", OnClick: func() { qp.selectEntry(entry) }},
				{Title: pinLabel, OnClick: func() {
					qp.app.store.Pin(entry.ID, !entry.Pinned)
					qp.adapter.entries = qp.app.store.GetRecent(quickPanelMaxItems)
					qp.list.NotifyDataSetChanged()
				}},
				{IsSeparator: true},
				{Title: "删除", OnClick: func() {
					qp.app.store.Delete(entry.ID)
					qp.adapter.entries = qp.app.store.GetRecent(quickPanelMaxItems)
					qp.list.NotifyDataSetChanged()
				}},
			}
			showNativePopupMenu(qp.window.NativeHandle(), items, screenX, screenY)
		}
	})
	root.AddChild(qp.list.Node())

	return root
}

func (qp *quickPanel) selectEntry(entry store.ClipEntry) {
	clip := qp.app.plat.GetClipboard()
	clip.SetText(entry.Text)
	qp.app.store.Use(entry.ID)
	qp.hide()

	if qp.previousHwnd != 0 {
		simulateCtrlV()
	}
	qp.app.updateTrayMenu()
}

func (qp *quickPanel) startDismissWatch() {
	qp.stopDismissWatch()
	qp.dismissDone = make(chan struct{})
	qp.dismissTimer = time.NewTicker(100 * time.Millisecond)

	go func() {
		user32 := syscall.NewLazyDLL("user32.dll")
		getAsyncKeyState := user32.NewProc("GetAsyncKeyState")
		getWindowRect := user32.NewProc("GetWindowRect")
		getCursorPos := user32.NewProc("GetCursorPos")

		for {
			select {
			case <-qp.dismissDone:
				return
			case <-qp.dismissTimer.C:
				if !qp.visible || qp.window == nil {
					return
				}

				// Escape key
				retEsc, _, _ := getAsyncKeyState.Call(0x1B)
				if retEsc&0x8000 != 0 {
					qp.app.plat.PostToMainThread(func() { qp.hide() })
					return
				}

				// Mouse button click outside
				retL, _, _ := getAsyncKeyState.Call(0x01)
				retR, _, _ := getAsyncKeyState.Call(0x02)
				if retL&0x8000 == 0 && retR&0x8000 == 0 {
					continue
				}

				hwnd := qp.window.NativeHandle()
				var wr struct{ Left, Top, Right, Bottom int32 }
				getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&wr)))

				var pt struct{ X, Y int32 }
				getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

				if pt.X < wr.Left || pt.X > wr.Right || pt.Y < wr.Top || pt.Y > wr.Bottom {
					qp.app.plat.PostToMainThread(func() { qp.hide() })
					return
				}
			}
		}
	}()
}

func (qp *quickPanel) stopDismissWatch() {
	if qp.dismissTimer != nil {
		qp.dismissTimer.Stop()
		qp.dismissTimer = nil
	}
	if qp.dismissDone != nil {
		close(qp.dismissDone)
		qp.dismissDone = nil
	}
}

func simulateCtrlV() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendInput := user32.NewProc("SendInput")

	const (
		INPUT_KEYBOARD  = 1
		KEYEVENTF_KEYUP = 0x0002
		VK_CONTROL      = 0x11
		VK_V            = 0x56
	)

	type KEYBDINPUT struct {
		Vk        uint16
		Scan      uint16
		Flags     uint32
		Time      uint32
		ExtraInfo uintptr
	}

	type INPUT struct {
		Type uint32
		Ki   KEYBDINPUT
		_    [8]byte
	}

	inputs := []INPUT{
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: VK_CONTROL}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: VK_V}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: VK_V, Flags: KEYEVENTF_KEYUP}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: VK_CONTROL, Flags: KEYEVENTF_KEYUP}},
	}

	sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
}
