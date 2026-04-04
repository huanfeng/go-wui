package widget

import (
	"testing"

	"github.com/huanfeng/wind-ui/core"
	"github.com/huanfeng/wind-ui/layout"
)

// testLeafPainter is a simple painter that reports a fixed intrinsic size.
type testLeafPainter struct {
	width, height float64
}

func (p *testLeafPainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
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

func (p *testLeafPainter) Paint(node *core.Node, canvas core.Canvas) {}

func newScrollTestChild(w, h float64) *core.Node {
	n := core.NewNode("TestChild")
	n.SetPainter(&testLeafPainter{width: w, height: h})
	n.SetStyle(&core.Style{})
	return n
}

func TestScrollViewScrollTo(t *testing.T) {
	sv := NewScrollView()
	// Add a tall child and measure so clamping has bounds
	child := newScrollTestChild(200, 1000)
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	sv.ScrollTo(0, 100)
	if sv.GetScrollY() != 100 {
		t.Errorf("expected ScrollY 100, got %v", sv.GetScrollY())
	}
}

func TestScrollViewClampNegative(t *testing.T) {
	sv := NewScrollView()
	child := newScrollTestChild(200, 1000)
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	sv.ScrollTo(0, -50) // negative should clamp to 0
	if sv.GetScrollY() != 0 {
		t.Errorf("should clamp to 0, got %v", sv.GetScrollY())
	}
}

func TestScrollViewClampMax(t *testing.T) {
	sv := NewScrollView()
	child := newScrollTestChild(200, 1000)
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	// Max scroll = childHeight(1000) - viewportHeight(300) = 700
	sv.ScrollTo(0, 900)
	if sv.GetScrollY() != 700 {
		t.Errorf("should clamp to max 700, got %v", sv.GetScrollY())
	}
}

func TestScrollViewDirection(t *testing.T) {
	sv := NewScrollView()
	if sv.Direction() != layout.Vertical {
		t.Error("NewScrollView should be vertical")
	}

	hsv := NewHorizontalScrollView()
	if hsv.Direction() != layout.Horizontal {
		t.Error("NewHorizontalScrollView should be horizontal")
	}
}

func TestHorizontalScrollViewScrollTo(t *testing.T) {
	sv := NewHorizontalScrollView()
	child := newScrollTestChild(1500, 100)
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 400},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
	)

	sv.ScrollTo(200, 0)
	if sv.GetScrollX() != 200 {
		t.Errorf("expected ScrollX 200, got %v", sv.GetScrollX())
	}
}

func TestHorizontalScrollViewClamp(t *testing.T) {
	sv := NewHorizontalScrollView()
	child := newScrollTestChild(1500, 100)
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 400},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
	)

	// Max scroll = 1500 - 400 = 1100
	sv.ScrollTo(2000, 0)
	if sv.GetScrollX() != 1100 {
		t.Errorf("should clamp to max 1100, got %v", sv.GetScrollX())
	}

	sv.ScrollTo(-100, 0)
	if sv.GetScrollX() != 0 {
		t.Errorf("should clamp to 0, got %v", sv.GetScrollX())
	}
}

func TestScrollViewNoChildNoScroll(t *testing.T) {
	sv := NewScrollView()
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	sv.ScrollTo(0, 100) // no child, should clamp to 0
	if sv.GetScrollY() != 0 {
		t.Errorf("no child: expected 0, got %v", sv.GetScrollY())
	}
}

func TestScrollViewChildSmallerThanViewport(t *testing.T) {
	sv := NewScrollView()
	child := newScrollTestChild(200, 100) // smaller than viewport
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	sv.ScrollTo(0, 50) // can't scroll when child fits
	if sv.GetScrollY() != 0 {
		t.Errorf("child fits in viewport: expected 0, got %v", sv.GetScrollY())
	}
}

func TestScrollViewPaintsChildrenFlag(t *testing.T) {
	sv := NewScrollView()
	if sv.Node().GetData("paintsChildren") == nil {
		t.Error("ScrollView should set paintsChildren data flag")
	}
}

func TestScrollViewScrollEvent(t *testing.T) {
	sv := NewScrollView()
	child := newScrollTestChild(200, 1000)
	sv.Node().AddChild(child)
	layout.MeasureChild(sv.Node(),
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200},
		core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
	)

	// Simulate scroll wheel event (deltaY negative = scroll down)
	handler := sv.Node().GetHandler()
	se := core.NewScrollEvent(100, 100, 0, -2) // 2 notches down
	consumed := handler.OnEvent(sv.Node(), se)
	if !consumed {
		t.Error("scroll event should be consumed")
	}
	// DeltaY = -2, so offset should increase by 2 * 48 = 96
	if sv.GetScrollY() != 96 {
		t.Errorf("after scroll wheel: expected 96, got %v", sv.GetScrollY())
	}
}

func TestScrollViewGetScrollInitialZero(t *testing.T) {
	sv := NewScrollView()
	if sv.GetScrollX() != 0 {
		t.Errorf("initial ScrollX should be 0, got %v", sv.GetScrollX())
	}
	if sv.GetScrollY() != 0 {
		t.Errorf("initial ScrollY should be 0, got %v", sv.GetScrollY())
	}
}
