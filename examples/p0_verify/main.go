// p0_verify is a comprehensive demo that validates all P0 fixes:
//   - Margin support in layouts
//   - PaintNode unification (overlay rendering)
//   - Keyboard event dispatch + FocusManager
//
// Usage:
//
//	go run ./examples/p0_verify              # offscreen PNG (margin + layout verification)
//	go run ./examples/p0_verify --window     # live window (keyboard/focus/overlay testing)
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"

	"github.com/huanfeng/wind-ui/app"
	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/layout"
	"github.com/huanfeng/wind-ui/platform"
	"github.com/huanfeng/wind-ui/render/freetype"
	"github.com/huanfeng/wind-ui/render/gg"
	"github.com/huanfeng/wind-ui/widget"
)

func main() {
	windowed := false
	for _, arg := range os.Args[1:] {
		if arg == "--window" {
			windowed = true
		}
	}

	if windowed {
		runWindowed()
	} else {
		runOffscreen()
	}
}

func runWindowed() {
	application := app.NewApplication()
	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "Wind UI P0 Verification",
		Width:     700,
		Height:    650,
		Resizable: true,
	})
	if err != nil {
		fmt.Printf("Failed to create window: %v\n", err)
		os.Exit(1)
	}

	root := buildUI()
	window.SetContentView(root)

	// Wire up buttons for interactive testing
	wireButtons(root, window)

	window.Center()
	window.Show()
	fmt.Println("=== P0 Verification (Window Mode) ===")
	fmt.Println("- Tab/Shift+Tab: cycle focus between buttons")
	fmt.Println("- Click 'Show Dialog': verify overlay rendering")
	fmt.Println("- Check margins between elements")
	application.Run()
}

func runOffscreen() {
	tr := freetype.NewFreeTypeTextRenderer()
	defer tr.Close()
	tm := core.NewTextMeasurer(tr)

	width, height := 700, 650
	root := buildUI()
	root.SetData("textMeasurer", tm)

	// Layout
	wSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(width)}
	hSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(height)}
	layout.MeasureChild(root, wSpec, hSpec)
	if l := root.GetLayout(); l != nil {
		l.Arrange(root, core.Rect{Width: float64(width), Height: float64(height)})
	}
	root.SetBounds(core.Rect{Width: float64(width), Height: float64(height)})

	// Paint using unified core.PaintNode
	canvas := gg.NewGGCanvas(width, height, tr)
	core.PaintNode(root, canvas)
	img := canvas.Target()

	// Print margin verification report
	printMarginReport(root)

	// Save
	outDir := filepath.Join(getExampleDir(), "output")
	os.MkdirAll(outDir, 0o755)
	outPath := filepath.Join(outDir, "p0_verify.png")
	if err := savePNG(outPath, img); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nScreenshot saved: %s (%dx%d)\n", outPath, width, height)
}

func buildUI() *core.Node {
	root := core.NewNode("root")
	root.SetLayout(&layout.LinearLayout{Orientation: layout.Vertical, Spacing: 4})
	root.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})
	root.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 245, G: 245, B: 245, A: 255},
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
	})
	root.SetPainter(&bgPainter{})

	// ====== Section 1: Margin Verification ======
	addTitle(root, "1. Margin Verification")

	// Row with margin on children
	row1 := core.NewNode("row1")
	row1.SetLayout(&layout.LinearLayout{Orientation: layout.Horizontal, Spacing: 0})
	row1.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	root.AddChild(row1)

	// Three boxes with different margins
	colors := []color.RGBA{
		{R: 230, G: 100, B: 100, A: 255},
		{R: 100, G: 200, B: 100, A: 255},
		{R: 100, G: 100, B: 230, A: 255},
	}
	margins := []core.Insets{
		{Left: 0, Top: 0, Right: 0, Bottom: 0},
		{Left: 10, Top: 5, Right: 10, Bottom: 5},
		{Left: 20, Top: 10, Right: 20, Bottom: 10},
	}
	labels := []string{"No Margin", "M:10x5", "M:20x10"}

	for i := 0; i < 3; i++ {
		box := widget.NewButton(labels[i], nil)
		box.Node().SetMargin(margins[i])
		box.Node().SetStyle(&core.Style{
			Width:           core.Dimension{Unit: core.DimensionDp, Value: 120},
			Height:          core.Dimension{Unit: core.DimensionDp, Value: 40},
			BackgroundColor: colors[i],
			TextColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			FontSize:        12,
		})
		row1.AddChild(box.Node())
	}

	// Vertical margin test
	addTitle(root, "2. Vertical Margin")
	for i, label := range []string{"Top margin=0", "Top margin=8", "Top margin=16"} {
		tv := widget.NewTextView(label)
		tv.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
		tv.Node().GetStyle().BackgroundColor = color.RGBA{R: 200, G: 220, B: 255, A: 200}
		tv.Node().SetMargin(core.Insets{Top: float64(i * 8), Left: float64(i * 10)})
		root.AddChild(tv.Node())
	}

	// ====== Section 2: FrameLayout Margin ======
	addTitle(root, "3. FrameLayout Margin")
	frame := core.NewNode("frame")
	frame.SetLayout(&layout.FrameLayout{})
	frame.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionDp, Value: 60},
	})
	frame.SetPainter(&bgPainterColor{clr: color.RGBA{R: 220, G: 220, B: 220, A: 255}})
	root.AddChild(frame)

	// Child with margin in FrameLayout
	frameChild := widget.NewTextView("Margin: 20px left")
	frameChild.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
	frameChild.Node().GetStyle().BackgroundColor = color.RGBA{R: 255, G: 200, B: 150, A: 255}
	frameChild.Node().SetMargin(core.Insets{Left: 20, Top: 10})
	frame.AddChild(frameChild.Node())

	// ====== Section 3: Buttons for Keyboard/Focus Test ======
	addTitle(root, "4. Keyboard & Focus (use --window)")
	btnRow := core.NewNode("btnRow")
	btnRow.SetLayout(&layout.LinearLayout{Orientation: layout.Horizontal, Spacing: 8})
	btnRow.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	root.AddChild(btnRow)

	for _, label := range []string{"Button A", "Button B", "Button C"} {
		btn := widget.NewButton(label, nil)
		btnRow.AddChild(btn.Node())
	}

	// Dialog trigger button
	dialogBtn := widget.NewButton("Show Dialog", nil)
	dialogBtn.SetId("btn_dialog")
	root.AddChild(dialogBtn.Node())

	// ====== Section 4: ScrollView with Margin Children ======
	addTitle(root, "5. ScrollView + Margin")
	sv := widget.NewScrollView()
	sv.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionDp, Value: 120},
		BackgroundColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	root.AddChild(sv.Node())

	svContent := core.NewNode("svContent")
	svContent.SetLayout(&layout.LinearLayout{Orientation: layout.Vertical, Spacing: 4})
	svContent.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	sv.Node().AddChild(svContent)

	for i := 0; i < 10; i++ {
		item := widget.NewTextView(fmt.Sprintf("Scroll Item %d (margin=8)", i+1))
		item.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
		item.Node().GetStyle().BackgroundColor = color.RGBA{R: 240, G: 248, B: 255, A: 255}
		item.Node().SetMargin(core.Insets{Left: 8, Right: 8, Top: 2, Bottom: 2})
		svContent.AddChild(item.Node())
	}

	// ====== Section 5: GridLayout with Margin ======
	addTitle(root, "6. GridLayout + Margin")
	grid := core.NewNode("grid")
	grid.SetLayout(&layout.GridLayout{ColumnCount: 3, Spacing: 4})
	grid.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	grid.SetPainter(&bgPainterColor{clr: color.RGBA{R: 235, G: 235, B: 235, A: 255}})
	root.AddChild(grid)

	gridColors := []color.RGBA{
		{R: 255, G: 182, B: 193, A: 255},
		{R: 173, G: 216, B: 230, A: 255},
		{R: 144, G: 238, B: 144, A: 255},
		{R: 255, G: 218, B: 185, A: 255},
		{R: 221, G: 160, B: 221, A: 255},
		{R: 255, G: 255, B: 224, A: 255},
	}
	for i := 0; i < 6; i++ {
		cell := widget.NewTextView(fmt.Sprintf("Cell %d", i+1))
		cell.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
		cell.Node().GetStyle().BackgroundColor = gridColors[i]
		cell.Node().SetMargin(core.Insets{Left: 4, Top: 4, Right: 4, Bottom: 4})
		grid.AddChild(cell.Node())
	}

	return root
}

func wireButtons(root *core.Node, window platform.Window) {
	if v := root.FindViewById("btn_dialog"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(_ core.View) {
				widget.NewAlertDialogBuilder().
					SetTitle("Dialog Test").
					SetMessage("This dialog verifies overlay rendering via unified core.PaintNode. It should appear on top of all content.").
					SetPositiveButton("OK", nil).
					SetNegativeButton("Cancel", nil).
					Show(root)
				window.Invalidate()
			})
		}
	}
}

func addTitle(parent *core.Node, text string) {
	tv := widget.NewTextView(text)
	tv.Node().GetStyle().FontSize = 14
	tv.Node().GetStyle().TextColor = color.RGBA{R: 25, G: 118, B: 210, A: 255}
	tv.Node().SetMargin(core.Insets{Top: 6})
	parent.AddChild(tv.Node())
}

func printMarginReport(root *core.Node) {
	fmt.Println("\n=== Margin Verification Report ===")
	fmt.Printf("%-40s  %12s  %20s\n", "Node", "Bounds(XxY)", "Margin(L,T,R,B)")
	fmt.Println(repeatStr("-", 76))

	walkNodes(root, func(node *core.Node) {
		tag := node.Tag()
		text := node.GetDataString("text")
		m := node.Margin()
		if m.Left == 0 && m.Top == 0 && m.Right == 0 && m.Bottom == 0 {
			return
		}
		b := node.Bounds()
		label := tag
		if text != "" {
			label = fmt.Sprintf("[%s] %s", tag, truncate(text, 25))
		}
		fmt.Printf("%-40s  %5.0f, %5.0f  %4.0f, %4.0f, %4.0f, %4.0f\n",
			label, b.X, b.Y, m.Left, m.Top, m.Right, m.Bottom)
	})
}

func walkNodes(node *core.Node, fn func(*core.Node)) {
	fn(node)
	for _, child := range node.Children() {
		walkNodes(child, fn)
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func repeatStr(s string, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = '-'
	}
	return string(out)
}

func savePNG(path string, img *image.RGBA) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func getExampleDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

// bgPainter draws a solid background.
type bgPainter struct{}

func (p *bgPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
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
		canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, paint)
	}
}

// bgPainterColor draws a specific color background.
type bgPainterColor struct {
	clr color.RGBA
}

func (p *bgPainterColor) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *bgPainterColor) Paint(node *core.Node, canvas core.Canvas) {
	b := node.Bounds()
	paint := &core.Paint{Color: p.clr, DrawStyle: core.PaintFill}
	canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, paint)
}
