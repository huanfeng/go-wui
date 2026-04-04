package main

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/examples/clipman/store"
	"github.com/huanfeng/wind-ui/layout"
	"github.com/huanfeng/wind-ui/platform"
	"github.com/huanfeng/wind-ui/widget"
)

// Win11 Fluent theme
var (
	colorBg         = "#F3F3F3"
	colorSurface    = "#FFFFFF"
	colorSidebar    = "#F3F3F3"
	colorToolbar    = "#F3F3F3"
	colorStatusBar  = "#F3F3F3"
	colorDivider    = "#E5E5E5"
	colorTextPri    = "#1A1A1A"
	colorTextSec    = "#888888"
	colorTextLight  = "#FFFFFF"
	colorBtnNormal  = "#F3F3F3"
	colorBtnActive  = "#D4E8FC"
	colorAccent     = "#0078D4"
	colorStatusText = "#888888"
)

type managerUI struct {
	app       *ClipManApp
	window    platform.Window
	clipList  *widget.RecyclerView
	statusBar *widget.TextView
	sidebar   *core.Node

	currentFilter string
	searchQuery   string
	adapter       *clipAdapter
	dpiScale      float64
}

type clipAdapter struct {
	ui      *managerUI
	entries []store.ClipEntry
}

func (ca *clipAdapter) GetItemCount() int           { return len(ca.entries) }
func (ca *clipAdapter) GetItemViewType(pos int) int { return 0 }

func (ca *clipAdapter) CreateViewHolder(viewType int) *widget.ViewHolder {
	item := NewClipItemWidget()
	return &widget.ViewHolder{ItemView: item}
}

func (ca *clipAdapter) BindViewHolder(holder *widget.ViewHolder, position int) {
	if position >= len(ca.entries) {
		return
	}
	entry := ca.entries[position]
	if item, ok := holder.ItemView.(*ClipItemWidget); ok {
		preview := truncateText(entry.Text, 60)
		source := entry.Source
		if source == "" {
			source = "未知来源"
		}
		item.SetContent(preview, source, formatTime(entry.CreatedAt), entry.Pinned)
	}
}

func newContainerNode(orientation layout.Orientation, spacing float64) *core.Node {
	v := widget.NewView()
	v.Node().SetLayout(&layout.LinearLayout{Orientation: orientation, Spacing: spacing})
	return v.Node()
}

func newManagerUI(app *ClipManApp, window platform.Window) *managerUI {
	ui := &managerUI{
		app:           app,
		window:        window,
		currentFilter: "all",
		dpiScale:      window.GetDPI() / 96.0,
	}
	if ui.dpiScale < 1.0 {
		ui.dpiScale = 1.0
	}

	root := ui.buildRootLayout()
	window.SetContentView(root)

	// Attach NativeEditText for search
	if v := root.FindViewById("search_input"); v != nil {
		ne := app.plat.CreateNativeEditText(window)
		if ne != nil {
			dpiScale := window.GetDPI() / 96.0
			ne.AttachToNode(v.Node())
			ne.SetFont("Microsoft YaHei UI", 13*dpiScale, 400)
			ne.SetPlaceholder("搜索...")
			ne.SetOnTextChanged(func(text string) {
				ui.searchQuery = text
				ui.refreshList()
			})
		}
	}

	ui.refreshList()
	return ui
}

func (ui *managerUI) buildRootLayout() *core.Node {
	root := newContainerNode(layout.Vertical, 0)
	root.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor(colorBg),
	})

	// === Header: "剪贴板" title + "全部清除" ===
	header := newContainerNode(layout.Horizontal, 0)
	header.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 48, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor(colorBg),
	})
	header.SetPadding(core.Insets{Left: 16, Top: 8, Right: 8, Bottom: 8})

	title := widget.NewTextView("剪贴板")
	title.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionMatchParent},
		TextColor: core.ParseColor(colorTextPri),
		FontSize:  16,
		Weight:    1,
		Gravity:   core.GravityCenterVertical,
	})
	header.AddChild(title.Node())

	clearBtn := widget.NewButton("全部清除", func(v core.View) {
		ui.app.store.ClearAll()
		ui.refreshList()
		ui.app.updateTrayMenu()
	})
	clearBtn.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Value: 80, Unit: core.DimensionDp},
		Height:          core.Dimension{Value: 32, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#E8E8E8"),
		TextColor:       core.ParseColor(colorTextPri),
		FontSize:        12,
		CornerRadius:    4,
	})
	header.AddChild(clearBtn.Node())
	root.AddChild(header)

	// === Search bar ===
	searchBar := newContainerNode(layout.Horizontal, 0)
	searchBar.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 36, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor(colorBg),
	})
	searchBar.SetPadding(core.Insets{Left: 12, Top: 0, Right: 12, Bottom: 4})

	searchInput := widget.NewEditText("搜索...")
	searchInput.Node().SetId("search_input")
	searchInput.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor(colorSurface),
		TextColor:       core.ParseColor(colorTextPri),
		FontSize:        13,
		BorderWidth:     1,
		BorderColor:     core.ParseColor(colorDivider),
		CornerRadius:    6,
	})
	searchBar.AddChild(searchInput.Node())
	root.AddChild(searchBar)

	// === Main content: sidebar + list ===
	content := newContainerNode(layout.Horizontal, 0)
	content.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		Weight:          1,
		BackgroundColor: core.ParseColor(colorBg),
	})

	ui.sidebar = ui.buildSidebar()
	content.AddChild(ui.sidebar)

	// RecyclerView with Win11 card items (scale itemHeight by DPI)
	clipList := widget.NewRecyclerView(72 * ui.dpiScale)
	ui.clipList = clipList
	clipList.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor(colorBg),
		Weight:          1,
	})

	ui.adapter = &clipAdapter{ui: ui}
	clipList.SetAdapter(ui.adapter)
	clipList.SetOnItemClickListener(func(position int) {
		if position >= 0 && position < len(ui.adapter.entries) {
			entry := ui.adapter.entries[position]
			clip := ui.app.plat.GetClipboard()
			clip.SetText(entry.Text)
			ui.app.store.Use(entry.ID)
			ui.refreshList()
			ui.setStatus(fmt.Sprintf("已复制: %s", truncateText(entry.Text, 30)))
		}
	})
	clipList.SetOnItemRightClickListener(func(position int, screenX, screenY int) {
		if position >= 0 && position < len(ui.adapter.entries) {
			ui.showItemContextMenu(position, screenX, screenY)
		}
	})
	content.AddChild(clipList.Node())
	root.AddChild(content)

	// === Status bar (subtle) ===
	statusBar := newContainerNode(layout.Horizontal, 0)
	statusBar.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 22, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor(colorStatusBar),
	})
	statusBar.SetPadding(core.Insets{Left: 16, Top: 2, Right: 16, Bottom: 2})

	statusText := widget.NewTextView("就绪")
	ui.statusBar = statusText
	statusText.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		TextColor: core.ParseColor(colorStatusText),
		FontSize:  11,
	})
	statusBar.AddChild(statusText.Node())
	root.AddChild(statusBar)

	return root
}

func (ui *managerUI) buildSidebar() *core.Node {
	sidebar := newContainerNode(layout.Vertical, 2)
	sidebar.SetStyle(&core.Style{
		Width:           core.Dimension{Value: 140, Unit: core.DimensionDp},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor(colorSidebar),
	})
	sidebar.SetPadding(core.Insets{Left: 4, Top: 4, Right: 4, Bottom: 4})

	sidebar.AddChild(ui.createSidebarButton("全部", "all").Node())
	sidebar.AddChild(ui.createSidebarButton("★ 已收藏", "pinned").Node())

	sideDiv := widget.NewView()
	sideDiv.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 1, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor(colorDivider),
	})
	sideDiv.Node().SetMargin(core.Insets{Top: 4, Bottom: 4})
	sidebar.AddChild(sideDiv.Node())

	return sidebar
}

func (ui *managerUI) createSidebarButton(title, filter string) *widget.Button {
	btn := widget.NewButton(title, func(v core.View) {
		ui.currentFilter = filter
		ui.refreshList()
		ui.updateSidebarHighlight()
	})

	bg := core.ParseColor(colorBtnNormal)
	if filter == ui.currentFilter {
		bg = core.ParseColor(colorBtnActive)
	}

	btn.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 32, Unit: core.DimensionDp},
		BackgroundColor: bg,
		TextColor:       core.ParseColor(colorTextPri),
		FontSize:        13,
		CornerRadius:    4,
	})
	btn.Node().SetId("cat_" + filter)
	return btn
}

func (ui *managerUI) refreshList() {
	var entries []store.ClipEntry
	switch ui.currentFilter {
	case "all":
		if ui.searchQuery != "" {
			entries = ui.app.store.Search(ui.searchQuery)
		} else {
			entries = ui.app.store.GetAll()
		}
	case "pinned":
		entries = ui.app.store.GetPinned()
		if ui.searchQuery != "" {
			entries = filterByQuery(entries, ui.searchQuery)
		}
	default:
		entries = ui.app.store.GetByCategory(ui.currentFilter)
		if ui.searchQuery != "" {
			entries = filterByQuery(entries, ui.searchQuery)
		}
	}
	ui.adapter.entries = entries
	if ui.clipList != nil {
		ui.clipList.NotifyDataSetChanged()
	}
	ui.setStatus(fmt.Sprintf("共 %d 条记录", len(entries)))
}

func (ui *managerUI) refreshSidebar() {
	if ui.sidebar == nil {
		return
	}
	children := ui.sidebar.Children()
	for len(children) > 3 {
		ui.sidebar.RemoveChild(children[len(children)-1])
		children = ui.sidebar.Children()
	}
	for _, cat := range ui.app.store.Categories() {
		ui.sidebar.AddChild(ui.createSidebarButton(cat, cat).Node())
	}
}

func (ui *managerUI) updateSidebarHighlight() {
	if ui.sidebar == nil {
		return
	}
	for _, child := range ui.sidebar.Children() {
		id := child.GetId()
		if !strings.HasPrefix(id, "cat_") {
			continue
		}
		style := child.GetStyle()
		if style == nil {
			continue
		}
		if strings.TrimPrefix(id, "cat_") == ui.currentFilter {
			style.BackgroundColor = core.ParseColor(colorBtnActive)
		} else {
			style.BackgroundColor = core.ParseColor(colorBtnNormal)
		}
		child.MarkDirty()
	}
}

func (ui *managerUI) setStatus(text string) {
	if ui.statusBar != nil {
		ui.statusBar.SetText(text)
	}
}

func (ui *managerUI) showItemContextMenu(position int, screenX, screenY int) {
	entry := ui.adapter.entries[position]

	// Use platform tray menu infrastructure for the popup
	// Build menu items
	pinLabel := "★ 收藏"
	if entry.Pinned {
		pinLabel = "取消收藏"
	}

	menuItems := []platform.TrayMenuItem{
		{
			ID:    "copy",
			Title: "复制",
			OnClick: func() {
				clip := ui.app.plat.GetClipboard()
				clip.SetText(entry.Text)
				ui.app.store.Use(entry.ID)
				ui.refreshList()
				ui.setStatus(fmt.Sprintf("已复制: %s", truncateText(entry.Text, 30)))
			},
		},
		{
			ID:    "pin",
			Title: pinLabel,
			OnClick: func() {
				ui.app.store.Pin(entry.ID, !entry.Pinned)
				ui.refreshList()
				ui.app.updateTrayMenu()
			},
		},
		{IsSeparator: true},
		{
			ID:    "delete",
			Title: "删除",
			OnClick: func() {
				ui.app.store.Delete(entry.ID)
				ui.refreshList()
				ui.app.updateTrayMenu()
				ui.setStatus("已删除")
			},
		},
	}

	showNativePopupMenu(ui.window.NativeHandle(), menuItems, screenX, screenY)
}

// showNativePopupMenu displays a Win32 popup menu at the given screen coordinates.
func showNativePopupMenu(ownerHwnd uintptr, items []platform.TrayMenuItem, clientX, clientY int) {
	user32 := syscall.NewLazyDLL("user32.dll")
	createPopupMenu := user32.NewProc("CreatePopupMenu")
	appendMenuW := user32.NewProc("AppendMenuW")
	trackPopupMenu := user32.NewProc("TrackPopupMenu")
	destroyMenu := user32.NewProc("DestroyMenu")
	setForegroundWindow := user32.NewProc("SetForegroundWindow")
	clientToScreen := user32.NewProc("ClientToScreen")

	// Convert client coords to screen coords
	pt := struct{ X, Y int32 }{int32(clientX), int32(clientY)}
	clientToScreen.Call(ownerHwnd, uintptr(unsafe.Pointer(&pt)))
	screenX, screenY := int(pt.X), int(pt.Y)

	hMenu, _, _ := createPopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer destroyMenu.Call(hMenu)

	handlers := make(map[uint32]func())
	var nextID uint32 = 2000

	for _, item := range items {
		if item.IsSeparator {
			appendMenuW.Call(hMenu, 0x00000800, 0, 0) // MF_SEPARATOR
			continue
		}
		id := nextID
		nextID++
		if item.OnClick != nil {
			handlers[id] = item.OnClick
		}
		titlePtr, _ := syscall.UTF16PtrFromString(item.Title)
		appendMenuW.Call(hMenu, 0, uintptr(id), uintptr(unsafe.Pointer(titlePtr)))
	}

	setForegroundWindow.Call(ownerHwnd)

	cmd, _, _ := trackPopupMenu.Call(
		hMenu,
		0x0100, // TPM_RETURNCMD
		uintptr(screenX), uintptr(screenY),
		0, ownerHwnd, 0,
	)

	if cmd != 0 {
		if handler, ok := handlers[uint32(cmd)]; ok {
			handler()
		}
	}
}

func filterByQuery(entries []store.ClipEntry, query string) []store.ClipEntry {
	query = strings.ToLower(query)
	var result []store.ClipEntry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Text), query) {
			result = append(result, e)
		}
	}
	return result
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	switch {
	case diff < time.Minute:
		return "刚刚"
	case diff < time.Hour:
		return fmt.Sprintf("%d 分钟前", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d 小时前", int(diff.Hours()))
	default:
		return t.Format("01-02 15:04")
	}
}
