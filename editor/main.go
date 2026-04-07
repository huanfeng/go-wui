package editor

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/huanfeng/wind-ui/app"
	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/layout"
	"github.com/huanfeng/wind-ui/platform"
	"github.com/huanfeng/wind-ui/render/gg"
	"github.com/huanfeng/wind-ui/widget"
)


// captureScreenshot renders the editor UI offscreen and saves as PNG.
func (ed *Editor) captureScreenshot() {
	if ed.screenshotPath == "" || ed.uiRoot == nil {
		return
	}
	width, height := 1280, 800
	if s := ed.window.GetSize(); s.Width > 0 && s.Height > 0 {
		width, height = int(s.Width), int(s.Height)
	}

	tr := ed.app.Platform().CreateTextRenderer()
	canvas := gg.NewGGCanvas(width, height, tr)
	wSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(width)}
	hSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(height)}
	layout.MeasureChild(ed.uiRoot, wSpec, hSpec)
	if l := ed.uiRoot.GetLayout(); l != nil {
		l.Arrange(ed.uiRoot, core.Rect{Width: float64(width), Height: float64(height)})
	}
	ed.uiRoot.SetBounds(core.Rect{Width: float64(width), Height: float64(height)})
	app.PaintNode(ed.uiRoot, canvas)

	img := canvas.Target()
	if err := savePNG(ed.screenshotPath, img); err != nil {
		fmt.Printf("Screenshot error: %v\n", err)
	} else {
		fmt.Printf("Screenshot saved: %s (%dx%d)\n", ed.screenshotPath, width, height)
	}
	ed.window.Close()
}

func savePNG(path string, img *image.RGBA) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// --- Container/node helpers ---

// containerPainter draws the background for layout containers.
type containerPainter struct{}

func (p *containerPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *containerPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil {
		return
	}
	b := node.Bounds()
	localRect := core.Rect{Width: b.Width, Height: b.Height}
	if s.BackgroundColor.A > 0 {
		paint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(localRect, paint)
		}
	}
	if s.BorderWidth > 0 && s.BorderColor.A > 0 {
		paint := &core.Paint{Color: s.BorderColor, DrawStyle: core.PaintStroke, StrokeWidth: s.BorderWidth}
		if s.CornerRadius > 0 {
			canvas.DrawRoundRect(localRect, s.CornerRadius, paint)
		} else {
			canvas.DrawRect(localRect, paint)
		}
	}
}

var sharedContainerPainter = &containerPainter{}

// newLinearLayout creates a properly initialized LinearLayout node.
// Callers should use SetStyle to configure dimensions and colors.
func newLinearLayout(orient layout.Orientation) *core.Node {
	n := core.NewNode("LinearLayout")
	n.SetLayout(&layout.LinearLayout{Orientation: orient})
	n.SetPainter(sharedContainerPainter)
	return n
}

// newFrameLayout creates a properly initialized FrameLayout node.
func newFrameLayout() *core.Node {
	n := core.NewNode("FrameLayout")
	n.SetLayout(&layout.FrameLayout{})
	n.SetPainter(sharedContainerPainter)
	return n
}

// newView creates a View node with background painter.
func newView() *core.Node {
	n := core.NewNode("View")
	n.SetPainter(sharedContainerPainter)
	return n
}

// nodeView wraps a *core.Node as a core.View for use with SplitPane etc.
type nodeView struct{ node *core.Node }

func (nv *nodeView) Node() *core.Node                { return nv.node }
func (nv *nodeView) SetId(id string)                  { nv.node.SetId(id) }
func (nv *nodeView) GetId() string                    { return nv.node.GetId() }
func (nv *nodeView) SetVisibility(v core.Visibility)  { nv.node.SetVisibility(v) }
func (nv *nodeView) GetVisibility() core.Visibility   { return nv.node.GetVisibility() }
func (nv *nodeView) SetEnabled(b bool)                { nv.node.SetEnabled(b) }
func (nv *nodeView) IsEnabled() bool                  { return nv.node.IsEnabled() }

func wrapNode(n *core.Node) core.View {
	v := &nodeView{node: n}
	n.SetView(v)
	return v
}

// Editor is the main UI editor application.
type Editor struct {
	app    *app.Application
	window platform.Window
	doc    *Document

	// UI state
	selectedNode   *EditorNode
	screenshotPath string // if set, save PNG after first render and exit
	uiRoot         *core.Node

	// Panels
	hierarchyTree *widget.TreeView
	propsScroll   *widget.ScrollView
	statusLabel   *widget.TextView
	previewRoot   *core.Node
}

// Run launches the editor. If screenshotPath is non-empty, the editor
// renders one frame, saves a PNG screenshot, and exits without entering
// the event loop — useful for automated visual debugging.
func Run(screenshotPath ...string) {
	application := app.NewApplication()

	ed := &Editor{
		app: application,
		doc: NewDocument(),
	}

	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "Wind UI Editor",
		Width:     1280,
		Height:    800,
		Resizable: true,
	})
	if err != nil {
		fmt.Printf("Failed to create window: %v\n", err)
		return
	}
	ed.window = window

	ed.uiRoot = ed.buildUI()
	window.SetContentView(ed.uiRoot)
	ed.refreshAll()

	// Screenshot mode: capture after first render via animation timer.
	if len(screenshotPath) > 0 && screenshotPath[0] != "" {
		ed.screenshotPath = screenshotPath[0]
		anim := &core.ValueAnimator{
			From: 0, To: 1,
			Duration: 200 * time.Millisecond,
			OnEnd: func() {
				ed.captureScreenshot()
			},
		}
		window.StartAnimator(anim)
	}

	window.Center()
	window.Show()
	application.Run()
}

func (ed *Editor) buildUI() *core.Node {
	main := newLinearLayout(layout.Vertical)
	main.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: color.RGBA{R: 240, G: 240, B: 240, A: 255},
	})

	main.AddChild(ed.buildToolbar())

	content := ed.buildContent()
	cs := content.Node().GetStyle()
	cs.Height = core.Dimension{Unit: core.DimensionWeight, Value: 1}
	cs.Weight = 1
	main.AddChild(content.Node())

	main.AddChild(ed.buildStatusBar())
	return main
}

func (ed *Editor) buildToolbar() *core.Node {
	bar := newLinearLayout(layout.Horizontal)
	bar.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionDp, Value: 40},
		BackgroundColor: color.RGBA{R: 50, G: 50, B: 60, A: 255},
	})
	bar.SetPadding(core.Insets{Left: 8, Right: 8, Top: 4, Bottom: 4})

	addBtn := func(label string, onClick func()) {
		btn := widget.NewButton(label, func(v core.View) { onClick() })
		btn.Node().SetStyle(&core.Style{
			Width:           core.Dimension{Unit: core.DimensionDp, Value: 70},
			Height:          core.Dimension{Unit: core.DimensionDp, Value: 32},
			FontSize:        12,
			BackgroundColor: color.RGBA{R: 70, G: 70, B: 85, A: 255},
		})
		btn.Node().SetMargin(core.Insets{Right: 4})
		bar.AddChild(btn.Node())
	}

	addBtn("New", ed.onNew)
	addBtn("Open", ed.onOpen)
	addBtn("Save", ed.onSave)

	spacer := newView()
	spacer.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionWeight, Value: 1},
		Height: core.Dimension{Unit: core.DimensionDp, Value: 1},
			Weight: 1,
	})
	bar.AddChild(spacer)

	addBtn("Preview", ed.onPreview)
	return bar
}

func (ed *Editor) buildContent() *widget.SplitPane {
	// Outer: palette | rest
	outerSplit := widget.NewSplitPane(layout.Horizontal, 0.15)
	outerSplit.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionMatchParent},
	})

	outerSplit.SetFirstPane(wrapNode(ed.buildPalette()))

	// Inner: canvas | right panel
	innerSplit := widget.NewSplitPane(layout.Horizontal, 0.7)
	innerSplit.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionMatchParent},
	})

	innerSplit.SetFirstPane(wrapNode(ed.buildCanvas()))

	// Right: hierarchy (top) | properties (bottom)
	rightSplit := widget.NewSplitPane(layout.Vertical, 0.5)
	rightSplit.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionMatchParent},
	})
	rightSplit.SetFirstPane(wrapNode(ed.buildHierarchy()))
	rightSplit.SetSecondPane(wrapNode(ed.buildPropsPanel()))

	innerSplit.SetSecondPane(rightSplit)
	outerSplit.SetSecondPane(innerSplit)
	return outerSplit
}

func (ed *Editor) buildPalette() *core.Node {
	panel := newLinearLayout(layout.Vertical)
	panel.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: color.RGBA{R: 245, G: 245, B: 248, A: 255},
	})

	title := widget.NewTextView("Widgets")
	title.Node().SetStyle(&core.Style{
		Width:    core.Dimension{Unit: core.DimensionMatchParent},
		Height:   core.Dimension{Unit: core.DimensionDp, Value: 28},
		FontSize: 12,
	})
	title.Node().SetPadding(core.Insets{Left: 8, Top: 6})
	panel.AddChild(title.Node())

	scroll := widget.NewScrollView()
	scroll.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWeight, Value: 1},
		Weight: 1,
	})

	list := newLinearLayout(layout.Vertical)
	list.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})

	for _, cat := range GetPaletteCategories() {
		header := widget.NewTextView("— " + cat.Name + " —")
		header.Node().SetStyle(&core.Style{
			Width:    core.Dimension{Unit: core.DimensionMatchParent},
			Height:   core.Dimension{Unit: core.DimensionDp, Value: 24},
			FontSize: 11,
		})
		header.Node().SetPadding(core.Insets{Left: 6, Top: 6})
		list.AddChild(header.Node())

		for _, item := range cat.Widgets {
			tag := item.Tag
			btn := widget.NewButton(item.DisplayName, func(v core.View) {
				ed.addNodeToSelection(tag)
			})
			btn.Node().SetStyle(&core.Style{
				Width:           core.Dimension{Unit: core.DimensionMatchParent},
				Height:          core.Dimension{Unit: core.DimensionDp, Value: 30},
				FontSize:        12,
				BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
			})
			btn.Node().SetMargin(core.Insets{Left: 4, Right: 4, Top: 2, Bottom: 2})
			list.AddChild(btn.Node())
		}
	}

	scroll.Node().AddChild(list)
	panel.AddChild(scroll.Node())
	return panel
}

func (ed *Editor) buildCanvas() *core.Node {
	frame := newFrameLayout()
	frame.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: color.RGBA{R: 220, G: 220, B: 225, A: 255},
	})
	frame.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})

	ed.previewRoot = newLinearLayout(layout.Vertical)
	ed.previewRoot.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	frame.AddChild(ed.previewRoot)
	return frame
}

func (ed *Editor) buildHierarchy() *core.Node {
	panel := newLinearLayout(layout.Vertical)
	panel.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: color.RGBA{R: 248, G: 248, B: 250, A: 255},
	})

	title := widget.NewTextView("Hierarchy")
	title.Node().SetStyle(&core.Style{
		Width:    core.Dimension{Unit: core.DimensionMatchParent},
		Height:   core.Dimension{Unit: core.DimensionDp, Value: 24},
		FontSize: 12,
	})
	title.Node().SetPadding(core.Insets{Left: 8, Top: 4})
	panel.AddChild(title.Node())

	ed.hierarchyTree = widget.NewTreeView()
	ed.hierarchyTree.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWeight, Value: 1},
		Weight: 1,
	})
	ed.hierarchyTree.SetOnNodeSelectedListener(func(tnode *widget.TreeNode) {
		if enode, ok := tnode.Data.(*EditorNode); ok {
			ed.selectNode(enode)
		}
	})
	panel.AddChild(ed.hierarchyTree.Node())
	return panel
}

func (ed *Editor) buildPropsPanel() *core.Node {
	panel := newLinearLayout(layout.Vertical)
	panel.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: color.RGBA{R: 248, G: 248, B: 250, A: 255},
	})

	title := widget.NewTextView("Properties")
	title.Node().SetStyle(&core.Style{
		Width:    core.Dimension{Unit: core.DimensionMatchParent},
		Height:   core.Dimension{Unit: core.DimensionDp, Value: 24},
		FontSize: 12,
	})
	title.Node().SetPadding(core.Insets{Left: 8, Top: 4})
	panel.AddChild(title.Node())

	ed.propsScroll = widget.NewScrollView()
	ed.propsScroll.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWeight, Value: 1},
		Weight: 1,
	})

	placeholder := widget.NewTextView("Select a node to edit properties")
	placeholder.Node().SetStyle(&core.Style{
		Width:    core.Dimension{Unit: core.DimensionMatchParent},
		Height:   core.Dimension{Unit: core.DimensionWrapContent},
		FontSize: 11,
	})
	placeholder.Node().SetPadding(core.Insets{Left: 8, Top: 12})
	ed.propsScroll.Node().AddChild(placeholder.Node())

	panel.AddChild(ed.propsScroll.Node())
	return panel
}

func (ed *Editor) buildStatusBar() *core.Node {
	bar := newLinearLayout(layout.Horizontal)
	bar.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionDp, Value: 24},
		BackgroundColor: color.RGBA{R: 50, G: 50, B: 60, A: 255},
	})
	bar.SetPadding(core.Insets{Left: 8, Top: 2})

	ed.statusLabel = widget.NewTextView("Ready")
	ed.statusLabel.Node().SetStyle(&core.Style{
		Width:    core.Dimension{Unit: core.DimensionMatchParent},
		Height:   core.Dimension{Unit: core.DimensionMatchParent},
		FontSize: 11,
	})
	bar.AddChild(ed.statusLabel.Node())
	return bar
}

// --- Actions ---

func (ed *Editor) onNew() {
	ed.doc = NewDocument()
	ed.selectedNode = nil
	ed.refreshAll()
	ed.setStatus("New document created")
}

func (ed *Editor) onOpen() {
	ed.setStatus("Open: not yet implemented")
}

func (ed *Editor) onSave() {
	if ed.doc.FilePath == "" {
		ed.setStatus("Save: not yet implemented (no file path)")
		return
	}
	if err := SaveXML(ed.doc.FilePath, ed.doc.Root); err != nil {
		ed.setStatus("Save error: " + err.Error())
		return
	}
	ed.doc.Modified = false
	ed.setStatus("Saved: " + ed.doc.FilePath)
}

func (ed *Editor) onPreview() {
	ed.setStatus("Preview: not yet implemented")
}

func (ed *Editor) addNodeToSelection(tag string) {
	newNode := NewEditorNode(tag)
	newNode.SetAttr("width", "wrap_content")
	newNode.SetAttr("height", "wrap_content")

	switch tag {
	case "LinearLayout", "FrameLayout", "ScrollView", "GridLayout", "FlexLayout":
		newNode.SetAttr("width", "match_parent")
		newNode.SetAttr("height", "wrap_content")
	}
	if tag == "LinearLayout" {
		newNode.SetAttr("orientation", "vertical")
	}
	if tag == "TextView" {
		newNode.SetAttr("text", "Text")
	}
	if tag == "Button" {
		newNode.SetAttr("text", "Button")
	}

	target := ed.doc.Root
	if ed.selectedNode != nil && ed.selectedNode.IsContainer() {
		target = ed.selectedNode
	} else if ed.selectedNode != nil && ed.selectedNode.Parent != nil {
		target = ed.selectedNode.Parent
	}
	target.AddChild(newNode)
	ed.doc.Modified = true
	ed.selectedNode = newNode
	ed.refreshAll()
	ed.setStatus(fmt.Sprintf("Added %s to %s", tag, target.DisplayName()))
}

func (ed *Editor) selectNode(node *EditorNode) {
	ed.selectedNode = node
	ed.refreshProperties()
	ed.refreshCanvas()
	ed.setStatus("Selected: " + node.DisplayName())
}

// --- Refresh ---

func (ed *Editor) refreshAll() {
	ed.refreshHierarchy()
	ed.refreshCanvas()
	ed.refreshProperties()
	ed.updateTitle()
}

func (ed *Editor) refreshHierarchy() {
	if ed.hierarchyTree == nil || ed.doc.Root == nil {
		return
	}
	root := ed.buildTreeNode(ed.doc.Root)
	ed.hierarchyTree.SetRoots([]*widget.TreeNode{root})
}

func (ed *Editor) buildTreeNode(enode *EditorNode) *widget.TreeNode {
	tn := &widget.TreeNode{
		Text:     enode.DisplayName(),
		Data:     enode,
		Expanded: true,
	}
	for _, child := range enode.Children {
		tn.AddChild(ed.buildTreeNode(child))
	}
	return tn
}

func (ed *Editor) refreshCanvas() {
	if ed.previewRoot == nil || ed.doc.Root == nil {
		return
	}
	for _, child := range ed.previewRoot.Children() {
		ed.previewRoot.RemoveChild(child)
	}
	preview := ed.buildPreviewNode(ed.doc.Root)
	if preview != nil {
		ed.previewRoot.AddChild(preview)
	}
	ed.previewRoot.Invalidate()
}

func (ed *Editor) buildPreviewNode(enode *EditorNode) *core.Node {
	// Create properly initialized node based on tag.
	var node *core.Node
	orient := layout.Vertical
	if enode.Attrs["orientation"] == "horizontal" {
		orient = layout.Horizontal
	}
	switch enode.Tag {
	case "LinearLayout":
		node = newLinearLayout(orient)
	case "FrameLayout":
		node = newFrameLayout()
	default:
		if enode.IsContainer() {
			node = newLinearLayout(orient)
		} else {
			node = newView()
		}
	}
	s := node.GetStyle()
	if s == nil {
		s = &core.Style{}
		node.SetStyle(s)
	}
	s.Width = core.Dimension{Unit: core.DimensionMatchParent}
	s.Height = core.Dimension{Unit: core.DimensionWrapContent}

	if w, ok := enode.Attrs["width"]; ok {
		s.Width = parseDimension(w)
	}
	if h, ok := enode.Attrs["height"]; ok {
		s.Height = parseDimension(h)
	}
	if bg, ok := enode.Attrs["background"]; ok {
		s.BackgroundColor = parseColor(bg)
	}
	if enode == ed.selectedNode {
		s.BorderWidth = 2
		s.BorderColor = color.RGBA{R: 33, G: 150, B: 243, A: 255}
	}

	if enode.IsContainer() {
		if s.BackgroundColor == (color.RGBA{}) {
			s.BackgroundColor = color.RGBA{R: 250, G: 250, B: 255, A: 255}
		}
		node.SetPadding(core.Insets{Left: 4, Top: 20, Right: 4, Bottom: 4})
		for _, child := range enode.Children {
			if cn := ed.buildPreviewNode(child); cn != nil {
				node.AddChild(cn)
			}
		}
	} else {
		text := "[" + enode.Tag + "]"
		if t, ok := enode.Attrs["text"]; ok {
			text = t
		}
		tv := widget.NewTextView(text)
		tv.Node().SetStyle(&core.Style{
			Width:    core.Dimension{Unit: core.DimensionMatchParent},
			Height:   core.Dimension{Unit: core.DimensionWrapContent},
			FontSize: 12,
		})
		tv.Node().SetPadding(core.Insets{Left: 4, Top: 2, Bottom: 2})
		node.AddChild(tv.Node())
	}

	node.SetHandler(&canvasClickHandler{editor: ed, enode: enode})
	return node
}

func (ed *Editor) refreshProperties() {
	if ed.propsScroll == nil {
		return
	}

	// Clear old content
	for _, child := range ed.propsScroll.Node().Children() {
		ed.propsScroll.Node().RemoveChild(child)
	}

	if ed.selectedNode == nil {
		tv := widget.NewTextView("Select a node to edit")
		tv.Node().SetStyle(&core.Style{
			Width:    core.Dimension{Unit: core.DimensionMatchParent},
			Height:   core.Dimension{Unit: core.DimensionWrapContent},
			FontSize: 11,
		})
		tv.Node().SetPadding(core.Insets{Left: 8, Top: 12})
		ed.propsScroll.Node().AddChild(tv.Node())
		return
	}

	form := newLinearLayout(layout.Vertical)
	form.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	form.SetPadding(core.Insets{Left: 4, Right: 4, Top: 4, Bottom: 4})

	tagLabel := widget.NewTextView(ed.selectedNode.Tag)
	tagLabel.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionDp, Value: 22},
		FontSize:        13,
		BackgroundColor: color.RGBA{R: 230, G: 230, B: 240, A: 255},
	})
	tagLabel.Node().SetPadding(core.Insets{Left: 4, Top: 2})
	form.AddChild(tagLabel.Node())

	props := GetPropertiesForTag(ed.selectedNode.Tag)
	currentGroup := ""
	for _, prop := range props {
		if prop.Group != currentGroup {
			currentGroup = prop.Group
			gl := widget.NewTextView(currentGroup)
			gl.Node().SetStyle(&core.Style{
				Width:    core.Dimension{Unit: core.DimensionMatchParent},
				Height:   core.Dimension{Unit: core.DimensionDp, Value: 20},
				FontSize: 10,
			})
			gl.Node().SetPadding(core.Insets{Left: 2, Top: 6})
			form.AddChild(gl.Node())
		}

		row := newLinearLayout(layout.Horizontal)
		row.SetStyle(&core.Style{
			Width:  core.Dimension{Unit: core.DimensionMatchParent},
			Height: core.Dimension{Unit: core.DimensionDp, Value: 26},
		})

		label := widget.NewTextView(prop.Name)
		label.Node().SetStyle(&core.Style{
			Width:    core.Dimension{Unit: core.DimensionDp, Value: 90},
			Height:   core.Dimension{Unit: core.DimensionMatchParent},
			FontSize: 11,
		})
		label.Node().SetPadding(core.Insets{Left: 4, Top: 4})
		row.AddChild(label.Node())

		val, _ := ed.selectedNode.GetAttr(prop.Name)
		if val == "" {
			val = prop.Default
		}
		valTV := widget.NewTextView(val)
		valTV.Node().SetStyle(&core.Style{
			Width:           core.Dimension{Unit: core.DimensionWeight, Value: 1},
			Height:          core.Dimension{Unit: core.DimensionMatchParent},
			FontSize:        11,
			BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Weight: 1,
		})
		valTV.Node().SetPadding(core.Insets{Left: 4, Top: 4})
		row.AddChild(valTV.Node())

		form.AddChild(row)
	}

	ed.propsScroll.Node().AddChild(form)
}

func (ed *Editor) setStatus(msg string) {
	if ed.statusLabel != nil {
		ed.statusLabel.SetText(msg)
		ed.statusLabel.Node().Invalidate()
	}
}

func (ed *Editor) updateTitle() {
	title := "Wind UI Editor"
	if ed.doc.FilePath != "" {
		title += " — " + ed.doc.FilePath
	}
	if ed.doc.Modified {
		title += " *"
	}
	if ed.window != nil {
		ed.window.SetTitle(title)
	}
}

// --- Canvas click handler ---

type canvasClickHandler struct {
	core.DefaultHandler
	editor *Editor
	enode  *EditorNode
}

func (h *canvasClickHandler) OnEvent(node *core.Node, event core.Event) bool {
	if event.Type() == core.EventMotion {
		if me, ok := event.(*core.MotionEvent); ok && me.Action == core.ActionDown {
			h.editor.selectNode(h.enode)
			return true
		}
	}
	return false
}

// --- Helpers ---

func parseDimension(s string) core.Dimension {
	switch s {
	case "match_parent":
		return core.Dimension{Unit: core.DimensionMatchParent}
	case "wrap_content":
		return core.Dimension{Unit: core.DimensionWrapContent}
	}
	s = strings.TrimSuffix(s, "dp")
	var v float64
	fmt.Sscanf(s, "%f", &v)
	if v > 0 {
		return core.Dimension{Unit: core.DimensionDp, Value: v}
	}
	return core.Dimension{Unit: core.DimensionWrapContent}
}

func parseColor(s string) color.RGBA {
	s = strings.TrimPrefix(s, "#")
	var r, g, b, a uint8 = 0, 0, 0, 255
	switch len(s) {
	case 6:
		fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	case 8:
		fmt.Sscanf(s, "%02x%02x%02x%02x", &a, &r, &g, &b)
	}
	return color.RGBA{R: r, G: g, B: b, A: a}
}
