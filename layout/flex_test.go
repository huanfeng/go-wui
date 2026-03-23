package layout

import (
	"testing"

	"gowui/core"
)

func TestFlexLayoutHorizontalNoWrap(t *testing.T) {
	parent := core.NewNode("Flex")
	fl := &FlexLayout{Orientation: Horizontal, Spacing: 10}
	parent.SetLayout(fl)

	for i := 0; i < 3; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 50, h: 30})
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 400}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 300}
	size := fl.Measure(parent, ws, hs)

	if size.Width != 400 {
		t.Errorf("expected width 400, got %f", size.Width)
	}
	// cross = max child height = 30
	if size.Height != 30 {
		t.Errorf("expected height 30, got %f", size.Height)
	}

	fl.Arrange(parent, core.Rect{Width: 400, Height: 30})
	children := parent.Children()
	// Children at x=0, 60, 120
	if children[0].Bounds().X != 0 {
		t.Errorf("child 0 X: expected 0, got %f", children[0].Bounds().X)
	}
	if children[1].Bounds().X != 60 {
		t.Errorf("child 1 X: expected 60, got %f", children[1].Bounds().X)
	}
	if children[2].Bounds().X != 120 {
		t.Errorf("child 2 X: expected 120, got %f", children[2].Bounds().X)
	}
}

func TestFlexLayoutWrap(t *testing.T) {
	parent := core.NewNode("Flex")
	fl := &FlexLayout{Orientation: Horizontal, Wrap: FlexWrapOn, Spacing: 10, LineSpacing: 5}
	parent.SetLayout(fl)

	// 4 items × 80 wide, container 200 wide → 2 per line
	for i := 0; i < 4; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 80, h: 30})
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500}
	size := fl.Measure(parent, ws, hs)

	// 2 lines × 30h + 1 lineSpacing × 5 = 65
	if size.Height != 65 {
		t.Errorf("expected height 65, got %f", size.Height)
	}

	fl.Arrange(parent, core.Rect{Width: 200, Height: 65})
	children := parent.Children()
	// Line 1: children[0] at y=0, children[1] at y=0
	// Line 2: children[2] at y=35, children[3] at y=35
	if children[2].Bounds().Y != 35 {
		t.Errorf("child 2 Y: expected 35, got %f", children[2].Bounds().Y)
	}
}

func TestFlexLayoutJustifyCenter(t *testing.T) {
	parent := core.NewNode("Flex")
	fl := &FlexLayout{Orientation: Horizontal, Justify: FlexJustifyCenter}
	parent.SetLayout(fl)

	child := core.NewNode("Item")
	child.SetPainter(&fixedPainter{w: 100, h: 30})
	parent.AddChild(child)

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 100}
	fl.Measure(parent, ws, hs)
	fl.Arrange(parent, core.Rect{Width: 300, Height: 100})

	b := parent.Children()[0].Bounds()
	// Centered: (300 - 100) / 2 = 100
	if b.X != 100 {
		t.Errorf("expected X=100 for centered item, got %f", b.X)
	}
}

func TestFlexLayoutJustifyEnd(t *testing.T) {
	parent := core.NewNode("Flex")
	fl := &FlexLayout{Orientation: Horizontal, Justify: FlexJustifyEnd}
	parent.SetLayout(fl)

	child := core.NewNode("Item")
	child.SetPainter(&fixedPainter{w: 80, h: 30})
	parent.AddChild(child)

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 100}
	fl.Measure(parent, ws, hs)
	fl.Arrange(parent, core.Rect{Width: 200, Height: 100})

	b := parent.Children()[0].Bounds()
	if b.X != 120 {
		t.Errorf("expected X=120 for end-justified item, got %f", b.X)
	}
}

func TestFlexLayoutAlignCenter(t *testing.T) {
	parent := core.NewNode("Flex")
	fl := &FlexLayout{Orientation: Horizontal, AlignItems: FlexAlignCenter}
	parent.SetLayout(fl)

	child1 := core.NewNode("Short")
	child1.SetPainter(&fixedPainter{w: 50, h: 20})
	parent.AddChild(child1)

	child2 := core.NewNode("Tall")
	child2.SetPainter(&fixedPainter{w: 50, h: 60})
	parent.AddChild(child2)

	ws := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 200}
	hs := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 200}
	fl.Measure(parent, ws, hs)
	fl.Arrange(parent, core.Rect{Width: 200, Height: 200})

	// Short child centered in 60-high line: y = (60-20)/2 = 20
	b := parent.Children()[0].Bounds()
	if b.Y != 20 {
		t.Errorf("expected Y=20 for cross-centered item, got %f", b.Y)
	}
}

func TestFlexLayoutVertical(t *testing.T) {
	parent := core.NewNode("Flex")
	fl := &FlexLayout{Orientation: Vertical, Spacing: 5}
	parent.SetLayout(fl)

	for i := 0; i < 3; i++ {
		child := core.NewNode("Item")
		child.SetPainter(&fixedPainter{w: 60, h: 40})
		parent.AddChild(child)
	}

	ws := core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 300}
	hs := core.MeasureSpec{Mode: core.MeasureModeExact, Size: 400}
	size := fl.Measure(parent, ws, hs)

	// cross = max width = 60
	if size.Width != 60 {
		t.Errorf("expected width 60, got %f", size.Width)
	}

	fl.Arrange(parent, core.Rect{Width: 60, Height: 400})
	children := parent.Children()
	if children[0].Bounds().Y != 0 {
		t.Errorf("child 0 Y: expected 0, got %f", children[0].Bounds().Y)
	}
	if children[1].Bounds().Y != 45 { // 40 + 5
		t.Errorf("child 1 Y: expected 45, got %f", children[1].Bounds().Y)
	}
}
