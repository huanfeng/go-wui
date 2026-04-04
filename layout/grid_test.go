package layout

import (
	"testing"

	"github.com/huanfeng/wind-ui/core"
)

func TestGridLayoutMeasure(t *testing.T) {
	parent := core.NewNode("Grid")
	gl := &GridLayout{ColumnCount: 3, Spacing: 8}
	parent.SetLayout(gl)

	// Add 6 children (2 rows × 3 cols)
	for i := 0; i < 6; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 50, h: 30})
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500}
	size := gl.Measure(parent, ws, hs)

	if size.Width != 300 {
		t.Errorf("expected width 300, got %f", size.Width)
	}
	// 2 rows × 30h + 1 spacing × 8 = 68
	if size.Height != 68 {
		t.Errorf("expected height 68, got %f", size.Height)
	}
}

func TestGridLayoutArrange(t *testing.T) {
	parent := core.NewNode("Grid")
	gl := &GridLayout{ColumnCount: 2, Spacing: 10}
	parent.SetLayout(gl)

	for i := 0; i < 4; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 50, h: 40})
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500}
	gl.Measure(parent, ws, hs)
	gl.Arrange(parent, core.Rect{Width: 200, Height: 500})

	children := parent.Children()
	// Cell width = (200 - 10) / 2 = 95
	b0 := children[0].Bounds()
	if b0.X != 0 || b0.Y != 0 {
		t.Errorf("child 0: expected (0,0), got (%f,%f)", b0.X, b0.Y)
	}
	if b0.Width != 95 {
		t.Errorf("child 0: expected width 95, got %f", b0.Width)
	}

	b1 := children[1].Bounds()
	if b1.X != 105 { // 95 + 10 spacing
		t.Errorf("child 1: expected X=105, got %f", b1.X)
	}

	b2 := children[2].Bounds()
	if b2.Y != 50 { // 40 + 10 spacing
		t.Errorf("child 2: expected Y=50, got %f", b2.Y)
	}
}

func TestGridLayoutSingleColumn(t *testing.T) {
	parent := core.NewNode("Grid")
	gl := &GridLayout{ColumnCount: 1}
	parent.SetLayout(gl)

	for i := 0; i < 3; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 100, h: 30})
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500}
	size := gl.Measure(parent, ws, hs)

	// 3 rows × 30h = 90
	if size.Height != 90 {
		t.Errorf("expected height 90, got %f", size.Height)
	}
}

func TestGridLayoutGoneChildren(t *testing.T) {
	parent := core.NewNode("Grid")
	gl := &GridLayout{ColumnCount: 2}
	parent.SetLayout(gl)

	for i := 0; i < 4; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 50, h: 30})
		if i == 1 {
			child.SetVisibility(core.Gone)
		}
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500}
	size := gl.Measure(parent, ws, hs)

	// 3 visible children → 2 rows
	if size.Height != 60 {
		t.Errorf("expected height 60, got %f", size.Height)
	}
}

// fixedPainter returns a fixed size for testing.
type fixedPainter struct {
	w, h float64
}

func (p *fixedPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w := p.w
	h := p.h
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *fixedPainter) Paint(node *core.Node, canvas core.Canvas) {}
