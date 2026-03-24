// Phase 4 Demo — Advanced Controls
//
// Demonstrates: GridLayout, FlexLayout, Spinner, SeekBar, Toast/Snackbar, TreeView, SplitPane
// Uses TabLayout + ViewPager for page switching, each page shows one control group.
package main

import (
	"fmt"

	"github.com/huanfeng/go-wui/app"
	"github.com/huanfeng/go-wui/core"
	"github.com/huanfeng/go-wui/layout"
	"github.com/huanfeng/go-wui/platform"
	"github.com/huanfeng/go-wui/widget"
)

var globalWindow platform.Window

func main() {
	application := app.NewApplication()

	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "GoWUI Phase 4 — Advanced Controls",
		Width:     500,
		Height:    560,
		Resizable: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create window: %v", err))
	}
	globalWindow = window

	root := buildUI(window)
	window.SetContentView(root)
	window.Center()
	window.Show()
	application.Run()
}

func buildUI(window platform.Window) *core.Node {
	rootLayout := &layout.LinearLayout{Orientation: layout.Vertical}
	root := core.NewNode("LinearLayout")
	root.SetLayout(rootLayout)
	root.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#FAFAFA"),
	})

	// Toolbar
	toolbar := widget.NewToolbar("Phase 4 Demo")
	toolbar.SetSubtitle("Advanced Controls")
	toolbar.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 50, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#388E3C"),
		TextColor:       core.ParseColor("#FFFFFF"),
		FontSize:        18,
	})

	// Status
	statusTV := widget.NewTextView("Interact with controls below")
	statusTV.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Value: 22, Unit: core.DimensionDp},
		FontSize:  11,
		TextColor: core.ParseColor("#FF5722"),
	})
	statusTV.Node().GetStyle().Gravity = core.GravityCenter

	updateStatus := func(msg string) {
		statusTV.SetText(msg)
		window.Invalidate()
	}

	// TabLayout
	tabLayout := widget.NewTabLayout()
	tabLayout.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 40, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#2E7D32"),
		TextColor:       core.ParseColor("#FFFFFF"),
		FontSize:        12,
	})

	// ScrollView as content area (one child swapped by tab)
	scrollView := widget.NewScrollView()
	scrollView.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionWeight},
		Weight:          1,
		BackgroundColor: core.ParseColor("#FFFFFF"),
	})

	// Build all pages
	pages := []struct {
		title string
		build func(func(string)) *core.Node
	}{
		{"Grid", buildGridPage},
		{"Flex", buildFlexPage},
		{"Inputs", func(us func(string)) *core.Node { return buildInputsPage(us, window) }},
		{"Tree", buildTreePage},
	}

	for _, p := range pages {
		tabLayout.AddTab(widget.Tab{Text: p.title})
	}

	// Build initial page
	currentPage := pages[0].build(updateStatus)
	scrollView.Node().AddChild(currentPage)

	tabLayout.SetOnTabSelectedListener(func(index int) {
		// Remove old content
		for _, child := range scrollView.Node().Children() {
			scrollView.Node().RemoveChild(child)
		}
		// Add new content
		if index < len(pages) {
			newPage := pages[index].build(updateStatus)
			scrollView.Node().AddChild(newPage)
			scrollView.ScrollTo(0, 0)
			updateStatus(fmt.Sprintf("Tab: %s", pages[index].title))
		}
	})

	root.AddChild(toolbar.Node())
	root.AddChild(statusTV.Node())
	root.AddChild(tabLayout.Node())
	root.AddChild(scrollView.Node())

	return root
}

// ---------- Grid Page ----------

func buildGridPage(updateStatus func(string)) *core.Node {
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 12}
	page := core.NewNode("LinearLayout")
	page.SetLayout(ll)
	page.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	page.SetPainter(&bgPainter{})
	page.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})

	title := newLabel("GridLayout — 3 columns, 8dp spacing")
	page.AddChild(title.Node())

	// Grid
	gridNode := core.NewNode("GridLayout")
	gridNode.SetLayout(&layout.GridLayout{ColumnCount: 3, Spacing: 8})
	gridNode.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	gridNode.SetPainter(&bgPainter{})

	colors := []string{"#E3F2FD", "#E8F5E9", "#FFF3E0", "#FCE4EC", "#F3E5F5", "#E0F7FA", "#FFF9C4", "#D7CCC8", "#CFD8DC"}
	for i := 0; i < 9; i++ {
		cell := widget.NewTextView(fmt.Sprintf("Cell %d", i+1))
		cell.Node().SetStyle(&core.Style{
			Width:           core.Dimension{Unit: core.DimensionMatchParent},
			Height:          core.Dimension{Value: 48, Unit: core.DimensionDp},
			FontSize:        12,
			TextColor:       core.ParseColor("#333333"),
			BackgroundColor: core.ParseColor(colors[i]),
			CornerRadius:    4,
		})
		cell.Node().GetStyle().Gravity = core.GravityCenter
		gridNode.AddChild(cell.Node())
	}

	page.AddChild(gridNode)

	// Description
	desc := newLabel("GridLayout arranges children in a uniform grid. All cells have equal width and the tallest cell determines row height.")
	page.AddChild(desc.Node())

	return page
}

// ---------- Flex Page ----------

func buildFlexPage(updateStatus func(string)) *core.Node {
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 16}
	page := core.NewNode("LinearLayout")
	page.SetLayout(ll)
	page.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	page.SetPainter(&bgPainter{})
	page.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})

	title := newLabel("FlexLayout — wrap, centered")
	page.AddChild(title.Node())

	// Flex container with chip-like buttons
	flexNode := core.NewNode("FlexLayout")
	flexNode.SetLayout(&layout.FlexLayout{
		Orientation: layout.Horizontal,
		Wrap:        layout.FlexWrapOn,
		Spacing:     8,
		LineSpacing: 8,
		Justify:     layout.FlexJustifyStart,
		AlignItems:  layout.FlexAlignCenter,
	})
	flexNode.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionWrapContent},
		BackgroundColor: core.ParseColor("#F5F5F5"),
		CornerRadius:    8,
	})
	flexNode.SetPainter(&bgPainter{})
	flexNode.SetPadding(core.Insets{Left: 8, Top: 8, Right: 8, Bottom: 8})

	tags := []string{"Go", "Rust", "TypeScript", "Python", "Java", "Kotlin", "Swift", "C++", "Ruby", "Dart"}
	colors := []string{"#1976D2", "#E64A19", "#0288D1", "#388E3C", "#D32F2F", "#7B1FA2", "#F57C00", "#455A64", "#C62828", "#00838F"}
	for i, tag := range tags {
		btn := widget.NewButton(tag, nil)
		btn.Node().SetStyle(&core.Style{
			Width:           core.Dimension{Unit: core.DimensionWrapContent},
			Height:          core.Dimension{Value: 30, Unit: core.DimensionDp},
			BackgroundColor: core.ParseColor(colors[i]),
			TextColor:       core.ParseColor("#FFFFFF"),
			CornerRadius:    15,
			FontSize:        14,
		})
		btn.SetOnClickListener(func(v core.View) {
			updateStatus(fmt.Sprintf("Flex: clicked '%s'", tag))
		})
		flexNode.AddChild(btn.Node())
	}

	page.AddChild(flexNode)

	// Second flex: justify space-between
	title2 := newLabel("FlexLayout — space-between")
	page.AddChild(title2.Node())

	flex2 := core.NewNode("FlexLayout")
	flex2.SetLayout(&layout.FlexLayout{
		Orientation: layout.Horizontal,
		Justify:     layout.FlexJustifySpaceBetween,
		AlignItems:  layout.FlexAlignCenter,
	})
	flex2.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 40, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#ECEFF1"),
		CornerRadius:    4,
	})
	flex2.SetPainter(&bgPainter{})

	for _, label := range []string{"Start", "Middle", "End"} {
		tv := widget.NewTextView(label)
		tv.Node().SetStyle(&core.Style{
			Width:     core.Dimension{Unit: core.DimensionWrapContent},
			Height:    core.Dimension{Unit: core.DimensionWrapContent},
			FontSize:  13,
			TextColor: core.ParseColor("#37474F"),
		})
		flex2.AddChild(tv.Node())
	}
	page.AddChild(flex2)

	desc := newLabel("FlexLayout supports wrap, justify (start/center/end/space-between/space-around), and cross-axis alignment.")
	page.AddChild(desc.Node())

	return page
}

// ---------- Inputs Page ----------

func buildInputsPage(updateStatus func(string), window platform.Window) *core.Node {
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 12}
	page := core.NewNode("LinearLayout")
	page.SetLayout(ll)
	page.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	page.SetPainter(&bgPainter{})
	page.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})

	// Spinner
	page.AddChild(newLabel("Spinner").Node())
	spinner := widget.NewSpinner([]string{"Option A", "Option B", "Option C", "Option D"})
	spinner.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 38, Unit: core.DimensionDp},
		FontSize:        14,
		TextColor:       core.ParseColor("#212121"),
		BorderColor:     core.ParseColor("#BDBDBD"),
		BorderWidth:     1,
		CornerRadius:    4,
		BackgroundColor: core.ParseColor("#FFFFFF"),
	})
	spinner.SetOnItemSelectedListener(func(idx int, item string) {
		updateStatus(fmt.Sprintf("Spinner: %s", item))
	})
	page.AddChild(spinner.Node())

	// SeekBar
	page.AddChild(newLabel("SeekBar").Node())
	seekBar := widget.NewSeekBar(0.4)
	seekBar.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Value: 28, Unit: core.DimensionDp},
	})

	seekValue := widget.NewTextView("40%")
	seekValue.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  12,
		TextColor: core.ParseColor("#333333"),
	})
	seekValue.Node().GetStyle().Gravity = core.GravityCenter

	seekBar.SetOnProgressChangedListener(func(p float64) {
		seekValue.SetText(fmt.Sprintf("%.0f%%", p*100))
		updateStatus(fmt.Sprintf("SeekBar: %.0f%%", p*100))
	})
	page.AddChild(seekBar.Node())
	page.AddChild(seekValue.Node())

	// Toast button
	toastBtn := widget.NewButton("Show Toast", nil)
	toastBtn.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 38, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#FF9800"),
		TextColor:       core.ParseColor("#FFFFFF"),
		CornerRadius:    4,
		FontSize:        13,
	})
	toastBtn.SetOnClickListener(func(_ core.View) {
		widget.ShowToast(page, "This is a Toast message!", widget.ToastShort)
		window.Invalidate()
	})
	page.AddChild(toastBtn.Node())

	// Snackbar button
	snackBtn := widget.NewButton("Show Snackbar", nil)
	snackBtn.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 38, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#795548"),
		TextColor:       core.ParseColor("#FFFFFF"),
		CornerRadius:    4,
		FontSize:        13,
	})
	snackBtn.SetOnClickListener(func(_ core.View) {
		widget.NewSnackbar(page, "Item deleted", "UNDO", func() {
			updateStatus("Snackbar: Undo!")
			window.Invalidate()
		})
		window.Invalidate()
	})
	page.AddChild(snackBtn.Node())

	return page
}

// ---------- Tree Page ----------

func buildTreePage(updateStatus func(string)) *core.Node {
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 8}
	page := core.NewNode("LinearLayout")
	page.SetLayout(ll)
	page.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	page.SetPainter(&bgPainter{})
	page.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})

	title := newLabel("TreeView — expand/collapse")
	page.AddChild(title.Node())

	tv := widget.NewTreeView()
	tv.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 380, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#FAFAFA"),
		TextColor:       core.ParseColor("#333333"),
		FontSize:        14,
		BorderColor:     core.ParseColor("#E0E0E0"),
		BorderWidth:     1,
		CornerRadius:    4,
	})

	// Build tree
	src := &widget.TreeNode{Text: "src", Expanded: true}
	src.AddChild(&widget.TreeNode{Text: "main.go"})
	pkg := &widget.TreeNode{Text: "pkg", Expanded: true}
	pkg.AddChild(&widget.TreeNode{Text: "widget"})
	pkg.AddChild(&widget.TreeNode{Text: "layout"})
	pkg.AddChild(&widget.TreeNode{Text: "core"})
	src.AddChild(pkg)
	src.AddChild(&widget.TreeNode{Text: "go.mod"})

	tests := &widget.TreeNode{Text: "tests", Expanded: false}
	tests.AddChild(&widget.TreeNode{Text: "widget_test.go"})
	tests.AddChild(&widget.TreeNode{Text: "layout_test.go"})

	docs := &widget.TreeNode{Text: "docs", Expanded: false}
	docs.AddChild(&widget.TreeNode{Text: "README.md"})
	docs.AddChild(&widget.TreeNode{Text: "CHANGELOG.md"})

	tv.SetRoots([]*widget.TreeNode{src, tests, docs})
	tv.SetOnNodeSelectedListener(func(n *widget.TreeNode) {
		updateStatus(fmt.Sprintf("TreeView: '%s'", n.Text))
	})
	page.AddChild(tv.Node())

	desc := newLabel("Click nodes to select. Click non-leaf nodes to expand/collapse. Supports scroll for large trees.")
	page.AddChild(desc.Node())

	return page
}

// ---------- Helpers ----------

func newLabel(text string) *widget.TextView {
	tv := widget.NewTextView(text)
	tv.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  13,
		TextColor: core.ParseColor("#757575"),
	})
	return tv
}

// bgPainter draws background for container nodes that have a Layout.
type bgPainter struct{}

func (p *bgPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	// Delegate to layout
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *bgPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	if s.BackgroundColor.A > 0 {
		paint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(core.Rect{Width: b.Width, Height: b.Height}, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, paint)
		}
	}
	if s.BorderWidth > 0 && s.BorderColor.A > 0 {
		paint := &core.Paint{Color: s.BorderColor, DrawStyle: core.PaintStroke, StrokeWidth: s.BorderWidth}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(core.Rect{Width: b.Width, Height: b.Height}, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, paint)
		}
	}
}
