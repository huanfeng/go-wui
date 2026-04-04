package layout

import (
	"testing"

	"github.com/huanfeng/wind-ui/core"
)

// Helper: create leaf node with fixed-size mock painter
func newTestLeaf(w, h float64) *core.Node {
	n := core.NewNode("TestLeaf")
	n.SetPainter(&fixedSizePainter{width: w, height: h})
	n.SetStyle(&core.Style{})
	return n
}

type fixedSizePainter struct {
	width, height float64
}

func (p *fixedSizePainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
	w := p.width
	h := p.height
	if ws.Mode == core.MeasureModeExact {
		w = ws.Size
	}
	if hs.Mode == core.MeasureModeExact {
		h = hs.Size
	}
	return core.Size{Width: w, Height: h}
}

func (p *fixedSizePainter) Paint(node *core.Node, canvas core.Canvas) {}

func TestLinearVertical_WrapContent(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Vertical, Spacing: 10}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(100, 30)
	child2 := newTestLeaf(80, 40)
	parent.AddChild(child1)
	parent.AddChild(child2)

	size := ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500},
	)
	// Width = max(100, 80) = 100, Height = 30 + 10 + 40 = 80
	if size.Width != 100 {
		t.Errorf("width: got %v, want 100", size.Width)
	}
	if size.Height != 80 {
		t.Errorf("height: got %v, want 80", size.Height)
	}
}

func TestLinearVertical_Weight(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Vertical}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(100, 0)
	s1 := &core.Style{Weight: 1}
	s1.Height = core.Dimension{Unit: core.DimensionWeight}
	child1.SetStyle(s1)

	child2 := newTestLeaf(100, 0)
	s2 := &core.Style{Weight: 2}
	s2.Height = core.Dimension{Unit: core.DimensionWeight}
	child2.SetStyle(s2)

	parent.AddChild(child1)
	parent.AddChild(child2)

	ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)
	ll.Arrange(parent, core.Rect{Width: 100, Height: 300})

	if child1.Bounds().Height != 100 {
		t.Errorf("child1 height: got %v, want 100", child1.Bounds().Height)
	}
	if child2.Bounds().Height != 200 {
		t.Errorf("child2 height: got %v, want 200", child2.Bounds().Height)
	}
}

func TestLinearHorizontal_Basic(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Horizontal, Spacing: 5}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(60, 30)
	child2 := newTestLeaf(40, 20)
	parent.AddChild(child1)
	parent.AddChild(child2)

	size := ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 300},
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 100},
	)
	// Width = 60 + 5 + 40 = 105, Height = max(30, 20) = 30
	if size.Width != 105 {
		t.Errorf("width: got %v, want 105", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height: got %v, want 30", size.Height)
	}
}

func TestLinearVertical_Arrange_Positions(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Vertical, Spacing: 10}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(80, 30)
	child2 := newTestLeaf(80, 40)
	parent.AddChild(child1)
	parent.AddChild(child2)

	ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
	)
	ll.Arrange(parent, core.Rect{Width: 100, Height: 200})

	// child1 at Y=0, child2 at Y=30+10=40
	if child1.Bounds().Y != 0 {
		t.Errorf("child1 Y: got %v, want 0", child1.Bounds().Y)
	}
	if child2.Bounds().Y != 40 {
		t.Errorf("child2 Y: got %v, want 40", child2.Bounds().Y)
	}
}
