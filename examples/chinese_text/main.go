// chinese_text verifies Chinese text rendering, measurement, and word wrapping.
//
// It can run in two modes:
//   - Offscreen (default): renders to PNG using FreeType (basicfont, ASCII-only fallback)
//   - Windowed (--window): opens a real window using DirectWrite for full CJK support
//
// Usage:
//
//	go run ./examples/chinese_text              # offscreen PNG
//	go run ./examples/chinese_text --window     # live window with DirectWrite
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
	// Check for --window mode
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

// runWindowed opens a real window with DirectWrite text rendering.
func runWindowed() {
	application := app.NewApplication()
	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "Wind UI Chinese Text Test",
		Width:     700,
		Height:    700,
		Resizable: true,
	})
	if err != nil {
		fmt.Printf("Failed to create window: %v\n", err)
		os.Exit(1)
	}

	root := buildUI()
	window.SetContentView(root)
	window.Center()
	window.Show()
	application.Run()
}

// runOffscreen renders to PNG using FreeType for cross-platform testing.
func runOffscreen() {
	tr := freetype.NewFreeTypeTextRenderer()
	defer tr.Close()
	tm := core.NewTextMeasurer(tr)

	width, height := 700, 700
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

	// Paint
	canvas := gg.NewGGCanvas(width, height, tr)
	app.PaintNode(root, canvas)
	img := canvas.Target()

	// Print measurement report
	printReport(root)

	// Save
	outDir := filepath.Join(getExampleDir(), "output")
	os.MkdirAll(outDir, 0o755)
	outPath := filepath.Join(outDir, "chinese_text.png")
	if err := savePNG(outPath, img); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nScreenshot saved: %s (%dx%d)\n", outPath, width, height)
	fmt.Println("\nNote: FreeType basicfont only supports ASCII.")
	fmt.Println("Run with --window flag for full CJK rendering via DirectWrite.")
}

func buildUI() *core.Node {
	root := core.NewNode("root")
	root.SetLayout(&layout.LinearLayout{Orientation: layout.Vertical, Spacing: 6})
	root.SetPadding(core.Insets{Left: 16, Top: 16, Right: 16, Bottom: 16})
	root.SetStyle(&core.Style{
		BackgroundColor: color.RGBA{R: 250, G: 250, B: 250, A: 255},
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
	})
	root.SetPainter(&bgPainter{})

	// ====== Section 1: Pure Chinese text ======
	addSectionTitle(root, "1. Chinese Text (Pure)")
	chineseTexts := []string{
		"Hello World",
		"Hello, World! This is Wind UI.",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"Mixed: Hello World 123",
	}
	for _, text := range chineseTexts {
		addTextWithBg(root, text, 14, color.RGBA{R: 255, G: 255, B: 255, A: 220})
	}

	// ====== Section 2: Different font sizes ======
	addSectionTitle(root, "2. Font Size Comparison")
	sizes := []float64{10, 12, 14, 16, 20, 24}
	for _, sz := range sizes {
		label := fmt.Sprintf("[%.0fpx] Wind UI Framework Test", sz)
		tv := widget.NewTextView(label)
		tv.Node().GetStyle().FontSize = sz
		tv.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
		tv.Node().GetStyle().BackgroundColor = color.RGBA{R: 230, G: 240, B: 255, A: 200}
		root.AddChild(tv.Node())
	}

	// ====== Section 3: Buttons with text ======
	addSectionTitle(root, "3. Buttons")
	btnRow := core.NewNode("btnRow")
	btnRow.SetLayout(&layout.LinearLayout{Orientation: layout.Horizontal, Spacing: 8})
	btnRow.SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Unit: core.DimensionWrapContent},
	})
	root.AddChild(btnRow)

	for _, label := range []string{"OK", "Cancel", "Submit", "Settings"} {
		btn := widget.NewButton(label, nil)
		btnRow.AddChild(btn.Node())
	}

	// ====== Section 4: CheckBox / RadioButton ======
	addSectionTitle(root, "4. CheckBox & RadioButton")
	for _, label := range []string{"Agree to terms", "Enable feature", "Remember me"} {
		cb := widget.NewCheckBox(label)
		root.AddChild(cb.Node())
	}
	for _, label := range []string{"Option A", "Option B", "Option C"} {
		rb := widget.NewRadioButton(label)
		root.AddChild(rb.Node())
	}

	// ====== Section 5: Long text / wrapping test ======
	addSectionTitle(root, "5. Long Text (measurement accuracy)")
	longTexts := []string{
		"This is a very long text line to test whether the text measurement system correctly handles extended ASCII content without overflow or truncation issues in the Wind UI framework.",
		"Short.",
		"A B C D E F G H I J K L M N O P Q R S T U V W X Y Z 0 1 2 3 4 5 6 7 8 9",
	}
	for _, text := range longTexts {
		addTextWithBg(root, text, 13, color.RGBA{R: 255, G: 255, B: 240, A: 220})
	}

	return root
}

func addSectionTitle(parent *core.Node, text string) {
	tv := widget.NewTextView(text)
	tv.Node().GetStyle().FontSize = 15
	tv.Node().GetStyle().TextColor = color.RGBA{R: 25, G: 118, B: 210, A: 255}
	parent.AddChild(tv.Node())
}

func addTextWithBg(parent *core.Node, text string, fontSize float64, bg color.RGBA) {
	tv := widget.NewTextView(text)
	tv.Node().GetStyle().FontSize = fontSize
	tv.Node().GetStyle().TextColor = color.RGBA{R: 33, G: 33, B: 33, A: 255}
	tv.Node().GetStyle().BackgroundColor = bg
	parent.AddChild(tv.Node())
}

func printReport(root *core.Node) {
	fmt.Println("\n=== Chinese Text Measurement Report ===")
	fmt.Printf("%-55s  %12s\n", "Text", "Size(WxH)")
	fmt.Println(repeatStr("-", 72))

	walkNodes(root, func(node *core.Node) {
		text := node.GetDataString("text")
		if text == "" {
			return
		}
		sz := node.MeasuredSize()
		label := truncate(text, 50)
		fmt.Printf("%-55s  %5.1f x %5.1f\n", label, sz.Width, sz.Height)
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
		out[i] = s[0]
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
