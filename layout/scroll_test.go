package layout

import (
	"testing"

	"github.com/huanfeng/go-wui/core"
)

func TestScrollLayoutMeasure_Vertical(t *testing.T) {
	parent := core.NewNode("ScrollView")
	sl := &ScrollLayout{Direction: Vertical}
	parent.SetLayout(sl)
	parent.SetStyle(&core.Style{})

	// Child taller than parent viewport
	child := newTestLeaf(200, 1000)
	parent.AddChild(child)

	size := sl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)
	// ScrollView itself is 200x300 (viewport size)
	if size.Width != 200 || size.Height != 300 {
		t.Errorf("viewport size: got %v x %v, want 200 x 300", size.Width, size.Height)
	}
	// Child should be measured at full 1000 height (unbound)
	if child.MeasuredSize().Height != 1000 {
		t.Errorf("child height: got %v, want 1000", child.MeasuredSize().Height)
	}
}

func TestScrollLayoutMeasure_Horizontal(t *testing.T) {
	parent := core.NewNode("HorizontalScrollView")
	sl := &ScrollLayout{Direction: Horizontal}
	parent.SetLayout(sl)
	parent.SetStyle(&core.Style{})

	// Child wider than parent viewport
	child := newTestLeaf(1500, 100)
	parent.AddChild(child)

	size := sl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 400},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
	)
	if size.Width != 400 || size.Height != 100 {
		t.Errorf("viewport size: got %v x %v, want 400 x 100", size.Width, size.Height)
	}
	if child.MeasuredSize().Width != 1500 {
		t.Errorf("child width: got %v, want 1500", child.MeasuredSize().Width)
	}
}

func TestScrollLayoutArrange_Vertical(t *testing.T) {
	parent := core.NewNode("ScrollView")
	sl := &ScrollLayout{Direction: Vertical, OffsetY: 50}
	parent.SetLayout(sl)
	parent.SetStyle(&core.Style{})

	child := newTestLeaf(200, 1000)
	parent.AddChild(child)

	sl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)
	sl.Arrange(parent, core.Rect{Width: 200, Height: 300})

	// Child should be positioned at Y = -50 due to scroll offset
	if child.Bounds().Y != -50 {
		t.Errorf("child Y: got %v, want -50", child.Bounds().Y)
	}
}

func TestScrollLayoutArrange_Horizontal(t *testing.T) {
	parent := core.NewNode("HorizontalScrollView")
	sl := &ScrollLayout{Direction: Horizontal, OffsetX: 100}
	parent.SetLayout(sl)
	parent.SetStyle(&core.Style{})

	child := newTestLeaf(1500, 100)
	parent.AddChild(child)

	sl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 400},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
	)
	sl.Arrange(parent, core.Rect{Width: 400, Height: 100})

	// Child should be positioned at X = -100 due to scroll offset
	if child.Bounds().X != -100 {
		t.Errorf("child X: got %v, want -100", child.Bounds().X)
	}
}

func TestScrollLayoutChildSize(t *testing.T) {
	parent := core.NewNode("ScrollView")
	sl := &ScrollLayout{Direction: Vertical}
	parent.SetLayout(sl)
	parent.SetStyle(&core.Style{})

	child := newTestLeaf(200, 800)
	parent.AddChild(child)

	sl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	cs := sl.ChildSize()
	if cs.Height != 800 {
		t.Errorf("child size height: got %v, want 800", cs.Height)
	}
}

func TestScrollLayoutNoChildren(t *testing.T) {
	parent := core.NewNode("ScrollView")
	sl := &ScrollLayout{Direction: Vertical}
	parent.SetLayout(sl)
	parent.SetStyle(&core.Style{})

	size := sl.Measure(parent,
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)
	if size.Width != 200 || size.Height != 300 {
		t.Errorf("empty scroll viewport: got %v x %v, want 200 x 300", size.Width, size.Height)
	}

	// Arrange should not panic with no children
	sl.Arrange(parent, core.Rect{Width: 200, Height: 300})
}
