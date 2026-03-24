package widget

import (
	"testing"

	"github.com/huanfeng/go-wui/core"
	"github.com/huanfeng/go-wui/layout"
)

func TestNewSplitPane(t *testing.T) {
	sp := NewSplitPane(layout.Horizontal, 0.5)
	if sp.Node().Tag() != "SplitPane" {
		t.Errorf("expected tag 'SplitPane', got %q", sp.Node().Tag())
	}
	if sp.GetRatio() != 0.5 {
		t.Errorf("expected ratio 0.5, got %f", sp.GetRatio())
	}
	if sp.GetOrientation() != layout.Horizontal {
		t.Error("expected horizontal orientation")
	}
}

func TestSplitPaneRatioClamp(t *testing.T) {
	sp := NewSplitPane(layout.Vertical, 0.0)
	if sp.GetRatio() != 0.1 {
		t.Errorf("expected clamped to 0.1, got %f", sp.GetRatio())
	}

	sp2 := NewSplitPane(layout.Vertical, 1.0)
	if sp2.GetRatio() != 0.9 {
		t.Errorf("expected clamped to 0.9, got %f", sp2.GetRatio())
	}
}

func TestSplitPaneSetRatio(t *testing.T) {
	sp := NewSplitPane(layout.Horizontal, 0.5)
	sp.SetRatio(0.3)
	if sp.GetRatio() != 0.3 {
		t.Errorf("expected 0.3, got %f", sp.GetRatio())
	}

	sp.SetRatio(0.0)
	if sp.GetRatio() != 0.1 {
		t.Errorf("expected clamped to 0.1, got %f", sp.GetRatio())
	}

	sp.SetRatio(1.0)
	if sp.GetRatio() != 0.9 {
		t.Errorf("expected clamped to 0.9, got %f", sp.GetRatio())
	}
}

func TestSplitPaneSetPanes(t *testing.T) {
	sp := NewSplitPane(layout.Horizontal, 0.5)

	tv1 := NewTextView("Left")
	tv2 := NewTextView("Right")

	sp.SetFirstPane(tv1)
	sp.SetSecondPane(tv2)

	if sp.GetFirstPane() != tv1 {
		t.Error("expected first pane to be tv1")
	}
	if sp.GetSecondPane() != tv2 {
		t.Error("expected second pane to be tv2")
	}
	if len(sp.Node().Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(sp.Node().Children()))
	}
}

func TestSplitPaneReplacePanes(t *testing.T) {
	sp := NewSplitPane(layout.Vertical, 0.5)

	tv1 := NewTextView("Top")
	tv2 := NewTextView("Bottom")
	sp.SetFirstPane(tv1)
	sp.SetSecondPane(tv2)

	// Replace first pane
	tv3 := NewTextView("New Top")
	sp.SetFirstPane(tv3)

	if sp.GetFirstPane() != tv3 {
		t.Error("expected first pane to be tv3")
	}
	if len(sp.Node().Children()) != 2 {
		t.Errorf("expected 2 children after replace, got %d", len(sp.Node().Children()))
	}
}

func TestSplitPaneDividerRect(t *testing.T) {
	sp := NewSplitPane(layout.Horizontal, 0.5)
	b := core.Rect{X: 0, Y: 0, Width: 400, Height: 300}
	dr := sp.dividerRect(b, 1.0)

	// Divider at 200, width 6
	if dr.X != 197 { // 200 - 3
		t.Errorf("expected divider X=197, got %f", dr.X)
	}
	if dr.Width != 6 {
		t.Errorf("expected divider width 6, got %f", dr.Width)
	}
	if dr.Height != 300 {
		t.Errorf("expected divider height 300, got %f", dr.Height)
	}
}

func TestSplitPaneVerticalDivider(t *testing.T) {
	sp := NewSplitPane(layout.Vertical, 0.4)
	b := core.Rect{X: 0, Y: 0, Width: 300, Height: 500}
	dr := sp.dividerRect(b, 1.0)

	// Divider at 200 (0.4 * 500), height 6
	if dr.Y != 197 { // 200 - 3
		t.Errorf("expected divider Y=197, got %f", dr.Y)
	}
	if dr.Height != 6 {
		t.Errorf("expected divider height 6, got %f", dr.Height)
	}
}
