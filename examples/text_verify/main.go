// text_verify is an offscreen rendering tool that validates text measurement
// accuracy. It builds a widget tree, measures/arranges with a real TextRenderer,
// renders to an image, and saves it as PNG for visual inspection.
//
// Usage: go run ./examples/text_verify
// Output: examples/text_verify/output/screenshot.png
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
	"github.com/huanfeng/wind-ui/render/freetype"
	"github.com/huanfeng/wind-ui/render/gg"
	"github.com/huanfeng/wind-ui/widget"
)

func main() {
	tr := freetype.NewFreeTypeTextRenderer()
	defer tr.Close()
	tm := core.NewTextMeasurer(tr)

	width, height := 700, 600

	root := buildUI(tm)

	// Layout pass
	wSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(width)}
	hSpec := core.MeasureSpec{Mode: core.MeasureModeExact, Size: float64(height)}
	layout.MeasureChild(root, wSpec, hSpec)
	if l := root.GetLayout(); l != nil {
		l.Arrange(root, core.Rect{Width: float64(width), Height: float64(height)})
	}
	root.SetBounds(core.Rect{Width: float64(width), Height: float64(height)})

	// Paint pass
	canvas := gg.NewGGCanvas(width, height, tr)
	app.PaintNode(root, canvas)
	img := canvas.Target()

	// Print measurement report
	printMeasurementReport(root, tr)

	// Save PNG
	outDir := filepath.Join(getExampleDir(), "output")
	os.MkdirAll(outDir, 0o755)
	outPath := filepath.Join(outDir, "screenshot.png")
	if err := savePNG(outPath, img); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nScreenshot saved: %s (%dx%d)\n", outPath, width, height)
}

func buildUI(tm core.TextMeasurer) *core.Node {
	root := core.NewNode("root")
	root.SetLayout(&layout.LinearLayout{Orientation: layout.Vertical, Spacing: 6})
	root.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})
	root.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 245, G: 245, B: 245, A: 255},
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
	})
	root.SetPainter(&bgPainter{})
	root.SetData("textMeasurer", tm)

	// --- Section 1: Title ---
	addSection(root, "=== Text Measurement Verification ===", 18,
		color.RGBA{R: 25, G: 118, B: 210, A: 255})

	// --- Section 2: Various text lengths ---
	addSection(root, "-- TextView (varying lengths) --", 13,
		color.RGBA{R: 100, G: 100, B: 100, A: 255})

	texts := []string{
		"A",
		"Hello",
		"Hello, World!",
		"The quick brown fox jumps over the lazy dog",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz",
		"0123456789 !@#$%^&*() special chars test",
	}
	for _, text := range texts {
		tv := widget.NewTextView(text)
		tv.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
		tv.Node().GetStyle().BackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 200}
		root.AddChild(tv.Node())
	}

	// --- Section 3: Buttons ---
	addSection(root, "-- Buttons --", 13,
		color.RGBA{R: 100, G: 100, B: 100, A: 255})

	btnRow := core.NewNode("btnRow")
	btnRow.SetLayout(&layout.LinearLayout{Orientation: layout.Horizontal, Spacing: 8})
	btnRow.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	root.AddChild(btnRow)

	for _, label := range []string{"OK", "Cancel", "Submit", "A Long Label"} {
		btn := widget.NewButton(label, nil)
		btnRow.AddChild(btn.Node())
	}

	// --- Section 4: CheckBoxes ---
	addSection(root, "-- CheckBox --", 13,
		color.RGBA{R: 100, G: 100, B: 100, A: 255})

	for _, label := range []string{"Option Alpha", "Option Bravo Charlie", "X"} {
		cb := widget.NewCheckBox(label)
		root.AddChild(cb.Node())
	}

	// --- Section 5: RadioButtons ---
	addSection(root, "-- RadioButton --", 13,
		color.RGBA{R: 100, G: 100, B: 100, A: 255})

	for _, label := range []string{"Small", "Medium", "Extra Large Size"} {
		rb := widget.NewRadioButton(label)
		root.AddChild(rb.Node())
	}

	return root
}

func addSection(parent *core.Node, text string, fontSize float64, clr color.RGBA) {
	tv := widget.NewTextView(text)
	tv.Node().GetStyle().FontSize = fontSize
	tv.Node().GetStyle().TextColor = clr
	parent.AddChild(tv.Node())
}

// printMeasurementReport prints the measured vs rendered sizes for key nodes.
func printMeasurementReport(root *core.Node, tr core.TextRenderer) {
	fmt.Println("\n=== Measurement Report ===")
	fmt.Printf("%-50s  %12s  %12s  %s\n", "Widget", "Layout(WxH)", "Canvas(WxH)", "Match?")
	fmt.Println(repeatStr("-", 90))

	canvas := gg.NewGGCanvas(1, 1, tr)
	walkNodes(root, func(node *core.Node) {
		text := node.GetDataString("text")
		if text == "" {
			return
		}
		tag := node.Tag()
		s := node.GetStyle()
		fontSize := 14.0
		if s != nil && s.FontSize > 0 {
			fontSize = s.FontSize
		}

		layoutSize := node.MeasuredSize()
		paint := &core.Paint{FontSize: fontSize}
		canvasSize := canvas.MeasureText(text, paint)

		match := "OK"
		if layoutSize.Width < canvasSize.Width-2 {
			match = "OVERFLOW!"
		}

		label := fmt.Sprintf("[%s] %s", tag, truncate(text, 38))
		fmt.Printf("%-50s  %5.1f x %5.1f  %5.1f x %5.1f  %s\n",
			label,
			layoutSize.Width, layoutSize.Height,
			canvasSize.Width, canvasSize.Height,
			match)
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
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
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

// bgPainter draws a solid background for the root.
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
