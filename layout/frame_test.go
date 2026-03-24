package layout

import (
	"testing"

	"github.com/huanfeng/go-wui/core"
)

func TestFrameLayout_Measure(t *testing.T) {
	parent := core.NewNode("FrameLayout")
	fl := &FrameLayout{}
	parent.SetLayout(fl)
	parent.SetStyle(&core.Style{})

	child1 := newTestLeaf(100, 50)
	child2 := newTestLeaf(80, 70)
	parent.AddChild(child1)
	parent.AddChild(child2)

	size := fl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 200},
	)
	// Width = max(100, 80) = 100, Height = max(50, 70) = 70
	if size.Width != 100 {
		t.Errorf("width: got %v, want 100", size.Width)
	}
	if size.Height != 70 {
		t.Errorf("height: got %v, want 70", size.Height)
	}
}

func TestFrameLayout_Arrange_GravityCenter(t *testing.T) {
	parent := core.NewNode("FrameLayout")
	fl := &FrameLayout{}
	parent.SetLayout(fl)
	parent.SetStyle(&core.Style{})

	child := newTestLeaf(50, 30)
	child.SetStyle(&core.Style{Gravity: core.GravityCenter})
	parent.AddChild(child)

	fl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
	)
	fl.Arrange(parent, core.Rect{Width: 200, Height: 100})

	// Centered: X = (200-50)/2 = 75, Y = (100-30)/2 = 35
	b := child.Bounds()
	if b.X != 75 {
		t.Errorf("X: got %v, want 75", b.X)
	}
	if b.Y != 35 {
		t.Errorf("Y: got %v, want 35", b.Y)
	}
}
