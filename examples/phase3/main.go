// Phase 3 Demo — List & Navigation Controls
//
// Demonstrates: Toolbar, TabLayout, ViewPager, RecyclerView, PopupMenu, AlertDialog
// All controls are built programmatically (no XML) to keep the demo self-contained.
package main

import (
	"fmt"

	"gowui/app"
	"gowui/core"
	"gowui/layout"
	"gowui/platform"
	"gowui/widget"
)

func main() {
	application := app.NewApplication()

	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "GoWUI Phase 3 — List & Navigation",
		Width:     480,
		Height:    640,
		Resizable: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create window: %v", err))
	}

	root := buildUI(window)
	window.SetContentView(root)
	window.Center()
	window.Show()
	application.Run()
}

func buildUI(window platform.Window) *core.Node {
	// Root vertical LinearLayout
	rootLayout := &layout.LinearLayout{Orientation: layout.Vertical}
	root := core.NewNode("LinearLayout")
	root.SetLayout(rootLayout)
	root.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#FAFAFA"),
	})

	// Status text (updated by interactions)
	statusTV := widget.NewTextView("Interact with Phase 3 controls")
	statusTV.SetId("status")
	statusTV.Node().SetStyle(&core.Style{
		Width:    core.Dimension{Unit: core.DimensionMatchParent},
		Height:   core.Dimension{Value: 28, Unit: core.DimensionDp},
		FontSize: 12,
		TextColor: core.ParseColor("#FF5722"),
	})
	statusTV.Node().GetStyle().Gravity = core.GravityCenter

	updateStatus := func(msg string) {
		statusTV.SetText(msg)
		window.Invalidate()
		fmt.Println(msg)
	}

	// --- 1. Toolbar ---
	toolbar := widget.NewToolbar("Phase 3 Demo")
	toolbar.SetSubtitle("List & Navigation")
	toolbar.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 56, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#1976D2"),
		TextColor:       core.ParseColor("#FFFFFF"),
		FontSize:        20,
	})
	toolbar.SetNavigationOnClickListener(func() {
		updateStatus("Toolbar: Navigation clicked")
	})
	toolbar.AddAction(widget.ActionItem{
		ID: "search", Title: "Search",
		OnClick: func() { updateStatus("Toolbar: Search") },
	})
	toolbar.AddAction(widget.ActionItem{
		ID: "menu", Title: "Menu",
		OnClick: func() {
			menu := widget.NewMenu()
			menu.AddItem("settings", "Settings", func() { updateStatus("Menu: Settings") })
			menu.AddItem("about", "About", func() { updateStatus("Menu: About") })
			menu.AddItem("help", "Help", func() { updateStatus("Menu: Help") })
			pm := widget.NewPopupMenu(menu)
			pm.SetOnDismissListener(func() { window.Invalidate() })
			// Anchor to toolbar's right side, below toolbar
			tbPos := toolbar.Node().AbsolutePosition()
			tbBounds := toolbar.Node().Bounds()
			pm.ShowAtPosition(toolbar.Node(), tbPos.X+tbBounds.Width-160, tbPos.Y+tbBounds.Height)
			window.Invalidate()
		},
	})

	// --- 2. TabLayout ---
	tabLayout := widget.NewTabLayout()
	tabLayout.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 44, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#1565C0"),
		TextColor:       core.ParseColor("#FFFFFF"),
		FontSize:        13,
	})
	tabLayout.AddTab(widget.Tab{Text: "RecyclerView"})
	tabLayout.AddTab(widget.Tab{Text: "ViewPager"})
	tabLayout.AddTab(widget.Tab{Text: "Dialogs"})

	// --- 3. Content area (changes with tab) ---
	// We use a ViewPager linked with the TabLayout
	viewPager := widget.NewViewPager()
	viewPager.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionWeight},
		Weight:          1,
		BackgroundColor: core.ParseColor("#FAFAFA"),
	})

	viewPager.SetAdapter(&phase3PagerAdapter{
		window:       window,
		updateStatus: updateStatus,
	})
	viewPager.SetupWithTabLayout(tabLayout)
	viewPager.SetOnPageChangedListener(func(idx int) {
		pages := []string{"RecyclerView", "ViewPager", "Dialogs"}
		if idx < len(pages) {
			updateStatus(fmt.Sprintf("Tab: %s", pages[idx]))
		}
	})

	// --- Assemble tree ---
	root.AddChild(toolbar.Node())
	root.AddChild(statusTV.Node())
	root.AddChild(tabLayout.Node())
	root.AddChild(viewPager.Node())

	return root
}

// ---------- PagerAdapter ----------

type phase3PagerAdapter struct {
	window       platform.Window
	updateStatus func(string)
}

func (a *phase3PagerAdapter) GetCount() int { return 3 }

func (a *phase3PagerAdapter) GetPageTitle(index int) string {
	return []string{"RecyclerView", "ViewPager", "Dialogs"}[index]
}

func (a *phase3PagerAdapter) CreatePage(index int) core.View {
	switch index {
	case 0:
		return a.createRecyclerViewPage()
	case 1:
		return a.createViewPagerInfoPage()
	case 2:
		return a.createDialogPage()
	default:
		return widget.NewTextView("Unknown page")
	}
}

// Page 0: RecyclerView demo
func (a *phase3PagerAdapter) createRecyclerViewPage() core.View {
	rv := widget.NewRecyclerView(44)
	rv.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#FFFFFF"),
	})
	rv.SetAdapter(&demoRecyclerAdapter{count: 50})
	rv.SetOnItemClickListener(func(pos int) {
		a.updateStatus(fmt.Sprintf("RecyclerView: Item %d clicked", pos))
	})
	return rv
}

// Page 1: ViewPager info
func (a *phase3PagerAdapter) createViewPagerInfoPage() core.View {
	tv := widget.NewTextView("ViewPager Demo - Switch tabs above to navigate between pages. ViewPager + TabLayout are linked for synchronized tab/page switching.")
	tv.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		FontSize:        15,
		TextColor:       core.ParseColor("#333333"),
		BackgroundColor: core.ParseColor("#FAFAFA"),
	})
	tv.Node().GetStyle().Gravity = core.GravityCenter
	tv.Node().SetPadding(core.Insets{Left: 24, Top: 24, Right: 24, Bottom: 24})
	return tv
}

// Page 2: Dialog/PopupMenu demo
func (a *phase3PagerAdapter) createDialogPage() core.View {
	container := widget.NewView()
	containerNode := container.Node()
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 16}
	containerNode.SetLayout(ll)
	containerNode.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#FAFAFA"),
	})
	containerNode.SetPadding(core.Insets{Left: 24, Top: 24, Right: 24, Bottom: 24})

	// Title
	title := widget.NewTextView("Dialog & PopupMenu Demo")
	title.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  16,
		TextColor: core.ParseColor("#212121"),
	})

	// AlertDialog button
	btnDialog := widget.NewButton("Show AlertDialog", nil)
	btnDialog.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 44, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#1976D2"),
		TextColor:       core.ParseColor("#FFFFFF"),
		CornerRadius:    4,
		FontSize:        14,
	})
	btnDialog.SetOnClickListener(func(_ core.View) {
		widget.NewAlertDialogBuilder().
			SetTitle("Confirm Action").
			SetMessage("This is an AlertDialog from Phase 3. It supports title, message, and up to three buttons.").
			SetPositiveButton("OK", func() {
				a.updateStatus("Dialog: OK clicked")
			}).
			SetNegativeButton("Cancel", func() {
				a.updateStatus("Dialog: Cancel clicked")
			}).
			SetNeutralButton("Help", func() {
				a.updateStatus("Dialog: Help clicked")
			}).
			SetOnDismissListener(func() {
				a.window.Invalidate()
			}).
			Show(containerNode)
		a.window.Invalidate()
	})

	// PopupMenu button
	btnMenu := widget.NewButton("Show PopupMenu", nil)
	btnMenu.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 44, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#388E3C"),
		TextColor:       core.ParseColor("#FFFFFF"),
		CornerRadius:    4,
		FontSize:        14,
	})
	btnMenu.SetOnClickListener(func(_ core.View) {
		menu := widget.NewMenu()
		menu.AddItem("cut", "Cut", func() { a.updateStatus("PopupMenu: Cut") })
		menu.AddItem("copy", "Copy", func() { a.updateStatus("PopupMenu: Copy") })
		menu.AddItem("paste", "Paste", func() { a.updateStatus("PopupMenu: Paste") })
		menu.AddItem("selectall", "Select All", func() { a.updateStatus("PopupMenu: Select All") })

		pm := widget.NewPopupMenu(menu)
		pm.SetOnDismissListener(func() { a.window.Invalidate() })
		b := btnMenu.Node().Bounds()
		pos := btnMenu.Node().AbsolutePosition()
		pm.ShowAtPosition(btnMenu.Node(), pos.X, pos.Y+b.Height)
		a.window.Invalidate()
	})

	// Info text
	info := widget.NewTextView("PopupMenu anchors to the button position.\nAlertDialog shows centered with a modal backdrop.\nClick outside to dismiss.")
	info.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  12,
		TextColor: core.ParseColor("#757575"),
	})

	containerNode.AddChild(title.Node())
	containerNode.AddChild(btnDialog.Node())
	containerNode.AddChild(btnMenu.Node())
	containerNode.AddChild(info.Node())

	return container
}

// ---------- RecyclerView Adapter ----------

type demoRecyclerAdapter struct {
	count int
}

func (a *demoRecyclerAdapter) GetItemCount() int           { return a.count }
func (a *demoRecyclerAdapter) GetItemViewType(pos int) int { return 0 }

func (a *demoRecyclerAdapter) CreateViewHolder(viewType int) *widget.ViewHolder {
	tv := widget.NewTextView("")
	tv.Node().SetStyle(&core.Style{
		FontSize:  16,
		TextColor: core.ParseColor("#212121"),
	})
	tv.Node().SetPadding(core.Insets{Left: 16, Top: 12, Right: 16, Bottom: 12})
	return &widget.ViewHolder{ItemView: tv}
}

func (a *demoRecyclerAdapter) BindViewHolder(holder *widget.ViewHolder, position int) {
	if tv, ok := holder.ItemView.(*widget.TextView); ok {
		tv.SetText(fmt.Sprintf("Item #%d — RecyclerView with view recycling", position+1))
	}
}
