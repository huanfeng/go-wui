// EditText Demo — Compatibility & Feature Testing
//
// Tests: single-line, multi-line, password, number input, inside ScrollView,
// dynamic resize, multiple EditTexts, placeholder text.
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
		Title:     "GoWUI EditText Demo",
		Width:     480,
		Height:    600,
		Resizable: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create window: %v", err))
	}

	root := buildUI(application, window)
	window.SetContentView(root)
	window.Center()
	window.Show()

	// Attach native edits AFTER first render (DPI scaling applied)
	attachNativeEdits(root, application, window)

	application.Run()
}

func buildUI(app *app.Application, window platform.Window) *core.Node {
	rootLayout := &layout.LinearLayout{Orientation: layout.Vertical}
	root := core.NewNode("LinearLayout")
	root.SetLayout(rootLayout)
	root.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#F5F5F5"),
	})

	// Toolbar
	toolbar := widget.NewToolbar("EditText Demo")
	toolbar.SetSubtitle("Compatibility Testing")
	toolbar.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 50, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#1976D2"),
		TextColor:       core.ParseColor("#FFFFFF"),
		FontSize:        18,
	})

	// Status bar
	statusTV := widget.NewTextView("Type in any field below")
	statusTV.SetId("status")
	statusTV.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Value: 22, Unit: core.DimensionDp},
		FontSize:  11,
		TextColor: core.ParseColor("#FF5722"),
	})
	statusTV.Node().GetStyle().Gravity = core.GravityCenter

	// === Fixed section (outside ScrollView) ===
	fixedSection := buildFixedSection()

	// Divider
	divider := widget.NewDivider()
	divider.Node().SetStyle(&core.Style{
		Width:  core.Dimension{Unit: core.DimensionMatchParent},
		Height: core.Dimension{Value: 1, Unit: core.DimensionDp},
	})

	// === Scrollable section ===
	scrollView := widget.NewScrollView()
	scrollView.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionWeight},
		Weight:          1,
		BackgroundColor: core.ParseColor("#FFFFFF"),
	})

	scrollContent := buildScrollContent()
	scrollView.Node().AddChild(scrollContent)

	root.AddChild(toolbar.Node())
	root.AddChild(statusTV.Node())
	root.AddChild(fixedSection)
	root.AddChild(divider.Node())
	root.AddChild(scrollView.Node())

	return root
}

// buildFixedSection creates EditText fields outside ScrollView (always visible).
func buildFixedSection() *core.Node {
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 8}
	section := core.NewNode("LinearLayout")
	section.SetLayout(ll)
	section.SetPainter(&containerBgPainter{}) // needed to draw BackgroundColor
	section.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		BackgroundColor: core.ParseColor("#F5F5F5"),
	})
	section.SetPadding(core.Insets{Left: 16, Top: 12, Right: 16, Bottom: 12})

	section.AddChild(newSectionLabel("Fixed Section (outside ScrollView)").Node())

	// Single-line: Name
	section.AddChild(newFieldLabel("Single-line — Name").Node())
	et1 := widget.NewEditText("Enter your name")
	et1.SetId("et_name")
	et1.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	section.AddChild(et1.Node())

	// Single-line: Email
	section.AddChild(newFieldLabel("Single-line — Email").Node())
	et2 := widget.NewEditText("Enter your email")
	et2.SetId("et_email")
	et2.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	section.AddChild(et2.Node())

	return section
}

// buildScrollContent creates EditText fields inside ScrollView for testing scrolling.
func buildScrollContent() *core.Node {
	ll := &layout.LinearLayout{Orientation: layout.Vertical, Spacing: 8}
	content := core.NewNode("LinearLayout")
	content.SetLayout(ll)
	content.SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Unit: core.DimensionWrapContent},
		BackgroundColor: core.ParseColor("#FFFFFF"),
	})
	content.SetPadding(core.Insets{Left: 16, Top: 12, Right: 16, Bottom: 16})

	content.AddChild(newSectionLabel("Scrollable Section (inside ScrollView)").Node())

	// Password input
	content.AddChild(newFieldLabel("Password Input").Node())
	etPass := widget.NewEditText("Enter password")
	etPass.SetId("et_password")
	etPass.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	content.AddChild(etPass.Node())

	// Number input
	content.AddChild(newFieldLabel("Number Input").Node())
	etNum := widget.NewEditText("Enter a number")
	etNum.SetId("et_number")
	etNum.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	content.AddChild(etNum.Node())

	// Multi-line
	content.AddChild(newFieldLabel("Multi-line — Notes").Node())
	etMulti := widget.NewEditText("Enter notes...")
	etMulti.SetId("et_multiline")
	etMulti.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 100, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	content.AddChild(etMulti.Node())

	// Additional fields to force scrolling
	content.AddChild(newFieldLabel("Address Line 1").Node())
	etAddr1 := widget.NewEditText("Street address")
	etAddr1.SetId("et_addr1")
	etAddr1.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	content.AddChild(etAddr1.Node())

	content.AddChild(newFieldLabel("Address Line 2").Node())
	etAddr2 := widget.NewEditText("City, State, ZIP")
	etAddr2.SetId("et_addr2")
	etAddr2.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	content.AddChild(etAddr2.Node())

	content.AddChild(newFieldLabel("Phone").Node())
	etPhone := widget.NewEditText("+86 xxx xxxx xxxx")
	etPhone.SetId("et_phone")
	etPhone.Node().SetStyle(&core.Style{
		Width:        core.Dimension{Unit: core.DimensionMatchParent},
		Height:       core.Dimension{Value: 36, Unit: core.DimensionDp},
		FontSize:     14,
		BorderWidth:  1,
		CornerRadius: 4,
	})
	content.AddChild(etPhone.Node())

	// Submit button
	btn := widget.NewButton("Submit All", nil)
	btn.SetId("btn_submit")
	btn.Node().SetStyle(&core.Style{
		Width:           core.Dimension{Unit: core.DimensionMatchParent},
		Height:          core.Dimension{Value: 40, Unit: core.DimensionDp},
		BackgroundColor: core.ParseColor("#1976D2"),
		TextColor:       core.ParseColor("#FFFFFF"),
		CornerRadius:    4,
		FontSize:        14,
	})
	content.AddChild(btn.Node())

	// Footer text
	footer := widget.NewTextView("Scroll up/down to test EditText inside ScrollView")
	footer.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  11,
		TextColor: core.ParseColor("#9E9E9E"),
	})
	footer.Node().GetStyle().Gravity = core.GravityCenter
	content.AddChild(footer.Node())

	return content
}

func attachNativeEdits(root *core.Node, application *app.Application, window platform.Window) {
	dpiScale := window.GetDPI() / 96.0

	type editConfig struct {
		id          string
		placeholder string
		inputType   platform.InputType
		multiLine   bool
	}

	configs := []editConfig{
		// Fixed section
		{id: "et_name", placeholder: "Enter your name"},
		{id: "et_email", placeholder: "Enter your email"},
		// Scroll section
		{id: "et_password", placeholder: "Enter password", inputType: platform.InputTypePassword},
		{id: "et_number", placeholder: "Enter a number", inputType: platform.InputTypeNumber},
		{id: "et_multiline", placeholder: "Enter notes here...", multiLine: true},
		{id: "et_addr1", placeholder: "Street address"},
		{id: "et_addr2", placeholder: "City, State, ZIP"},
		{id: "et_phone", placeholder: "+86 xxx xxxx xxxx"},
	}

	var statusTV *widget.TextView
	if v := root.FindViewById("status"); v != nil {
		statusTV, _ = v.(*widget.TextView)
	}

	for _, cfg := range configs {
		if v := root.FindViewById(cfg.id); v != nil {
			ne := application.Platform().CreateNativeEditText(window)
			if ne == nil {
				continue
			}
			ne.SetFont("Segoe UI", 14*dpiScale, 400)
			ne.SetPlaceholder(cfg.placeholder)
			if cfg.inputType != platform.InputTypeText {
				ne.SetInputType(cfg.inputType)
			}
			if cfg.multiLine {
				ne.SetMultiLine(true)
			}
			ne.AttachToNode(v.Node())

			// Wire text change to status
			id := cfg.id
			ne.SetOnTextChanged(func(text string) {
				if statusTV != nil {
					statusTV.SetText(fmt.Sprintf("%s: %s", id, text))
					window.Invalidate()
				}
			})
		}
	}

	// Submit button
	if v := root.FindViewById("btn_submit"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(_ core.View) {
				if statusTV != nil {
					statusTV.SetText("Submit clicked! Check console for values.")
					window.Invalidate()
				}
				// Print all field values
				for _, cfg := range configs {
					if fv := root.FindViewById(cfg.id); fv != nil {
						if ne, ok := fv.Node().GetData("nativeEdit").(interface{ GetText() string }); ok {
							fmt.Printf("  %s = %q\n", cfg.id, ne.GetText())
						}
					}
				}
			})
		}
	}
}

func newSectionLabel(text string) *widget.TextView {
	tv := widget.NewTextView(text)
	tv.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  14,
		TextColor: core.ParseColor("#1976D2"),
	})
	return tv
}

func newFieldLabel(text string) *widget.TextView {
	tv := widget.NewTextView(text)
	tv.Node().SetStyle(&core.Style{
		Width:     core.Dimension{Unit: core.DimensionMatchParent},
		Height:    core.Dimension{Unit: core.DimensionWrapContent},
		FontSize:  12,
		TextColor: core.ParseColor("#757575"),
	})
	return tv
}

// containerBgPainter draws background for programmatic container nodes.
// Needed because core.NewNode doesn't set a painter — without one,
// BackgroundColor in the style is never rendered.
type containerBgPainter struct{}

func (p *containerBgPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w, h := 0.0, 0.0
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *containerBgPainter) Paint(node *core.Node, canvas core.Canvas) {
	s := node.GetStyle()
	if s == nil || s.BackgroundColor.A == 0 {
		return
	}
	b := node.Bounds()
	paint := &core.Paint{Color: s.BackgroundColor, DrawStyle: core.PaintFill}
	canvas.DrawRect(core.Rect{Width: b.Width, Height: b.Height}, paint)
}
