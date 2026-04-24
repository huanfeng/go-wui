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

// TestLinearVertical_Margin 验证垂直 LinearLayout 中 margin 的测量效果。
func TestLinearVertical_Margin(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Vertical, Spacing: 0}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(100, 30)
	child1.SetMargin(core.Insets{Top: 5, Bottom: 10, Left: 8, Right: 8})

	child2 := newTestLeaf(80, 40)
	child2.SetMargin(core.Insets{Top: 4, Bottom: 6})

	parent.AddChild(child1)
	parent.AddChild(child2)

	size := ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 300},
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500},
	)
	// 高度 = (30 + 5 + 10) + (40 + 4 + 6) = 45 + 50 = 95
	// 宽度 = max(100+16, 80+0) = 116
	if size.Height != 95 {
		t.Errorf("height: got %v, want 95", size.Height)
	}
	if size.Width != 116 {
		t.Errorf("width: got %v, want 116", size.Width)
	}
}

// TestLinearVertical_Margin_Arrange 验证垂直 LinearLayout 中 margin 的排列效果。
func TestLinearVertical_Margin_Arrange(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Vertical, Spacing: 10}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(80, 30)
	child1.SetMargin(core.Insets{Top: 5, Bottom: 10})

	child2 := newTestLeaf(80, 40)
	child2.SetMargin(core.Insets{Top: 3, Bottom: 0})

	parent.AddChild(child1)
	parent.AddChild(child2)

	ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)
	ll.Arrange(parent, core.Rect{Width: 100, Height: 300})

	// child1: Y = 0 + margin.Top(5) = 5
	if child1.Bounds().Y != 5 {
		t.Errorf("child1 Y: got %v, want 5", child1.Bounds().Y)
	}
	// child2: curY after child1 = 5(margin.Top) + 30(height) + 10(margin.Bottom) + 10(spacing) = 55
	//         child2.Y = 55 + margin.Top(3) = 58
	if child2.Bounds().Y != 58 {
		t.Errorf("child2 Y: got %v, want 58", child2.Bounds().Y)
	}
}

// TestLinearHorizontal_Margin 验证水平 LinearLayout 中 margin 的测量和排列效果。
func TestLinearHorizontal_Margin(t *testing.T) {
	parent := core.NewNode("LinearLayout")
	ll := &LinearLayout{Orientation: Horizontal, Spacing: 5}
	parent.SetLayout(ll)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(60, 30)
	child1.SetMargin(core.Insets{Left: 4, Right: 6})

	child2 := newTestLeaf(40, 20)
	child2.SetMargin(core.Insets{Left: 2, Right: 3, Top: 5, Bottom: 5})

	parent.AddChild(child1)
	parent.AddChild(child2)

	size := ll.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 300},
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 100},
	)
	// 宽度 = (60+4+6) + spacing(5) + (40+2+3) = 70 + 5 + 45 = 120
	// 高度 = max(30+0, 20+10) = 30
	if size.Width != 120 {
		t.Errorf("width: got %v, want 120", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height: got %v, want 30", size.Height)
	}

	ll.Arrange(parent, core.Rect{Width: 120, Height: 100})

	// child1: X = 0 + margin.Left(4) = 4
	if child1.Bounds().X != 4 {
		t.Errorf("child1 X: got %v, want 4", child1.Bounds().X)
	}
	// child2: curX after child1 = 4(margin.Left) + 60(width) + 6(margin.Right) + 5(spacing) = 75
	//         child2.X = 75 + margin.Left(2) = 77
	if child2.Bounds().X != 77 {
		t.Errorf("child2 X: got %v, want 77", child2.Bounds().X)
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
