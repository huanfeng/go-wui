package widget

import (
	"fmt"
	"testing"

	"github.com/huanfeng/go-wui/core"
)

// testPagerAdapter is a simple adapter for testing.
type testPagerAdapter struct {
	count int
}

func (a *testPagerAdapter) GetCount() int { return a.count }
func (a *testPagerAdapter) CreatePage(index int) core.View {
	tv := NewTextView(fmt.Sprintf("Page %d", index))
	return tv
}
func (a *testPagerAdapter) GetPageTitle(index int) string {
	return fmt.Sprintf("Tab %d", index)
}

func TestNewViewPager(t *testing.T) {
	vp := NewViewPager()
	if vp.Node().Tag() != "ViewPager" {
		t.Errorf("expected tag 'ViewPager', got %q", vp.Node().Tag())
	}
	if vp.GetCurrentPage() != 0 {
		t.Errorf("expected current page 0, got %d", vp.GetCurrentPage())
	}
	if vp.GetPageCount() != 0 {
		t.Errorf("expected 0 pages, got %d", vp.GetPageCount())
	}
}

func TestViewPagerSetAdapter(t *testing.T) {
	vp := NewViewPager()
	adapter := &testPagerAdapter{count: 3}
	vp.SetAdapter(adapter)

	if vp.GetPageCount() != 3 {
		t.Errorf("expected 3 pages, got %d", vp.GetPageCount())
	}
	if len(vp.node.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(vp.node.Children()))
	}
	// First page visible, others gone
	if vp.pages[0].Node().GetVisibility() != core.Visible {
		t.Error("expected page 0 visible")
	}
	if vp.pages[1].Node().GetVisibility() != core.Gone {
		t.Error("expected page 1 gone")
	}
}

func TestViewPagerSetCurrentPage(t *testing.T) {
	vp := NewViewPager()
	adapter := &testPagerAdapter{count: 3}
	vp.SetAdapter(adapter)

	changedTo := -1
	vp.SetOnPageChangedListener(func(index int) {
		changedTo = index
	})

	vp.SetCurrentPage(2)
	if vp.GetCurrentPage() != 2 {
		t.Errorf("expected current page 2, got %d", vp.GetCurrentPage())
	}
	if changedTo != 2 {
		t.Errorf("expected listener called with 2, got %d", changedTo)
	}
	if vp.pages[2].Node().GetVisibility() != core.Visible {
		t.Error("expected page 2 visible")
	}
	if vp.pages[0].Node().GetVisibility() != core.Gone {
		t.Error("expected page 0 gone")
	}

	// Same page — no callback
	changedTo = -1
	vp.SetCurrentPage(2)
	if changedTo != -1 {
		t.Error("expected listener NOT called for same page")
	}

	// Out of range — no change
	vp.SetCurrentPage(10)
	if vp.GetCurrentPage() != 2 {
		t.Errorf("expected page still 2, got %d", vp.GetCurrentPage())
	}
}

func TestViewPagerSetupWithTabLayout(t *testing.T) {
	vp := NewViewPager()
	tl := NewTabLayout()

	adapter := &testPagerAdapter{count: 3}
	vp.SetAdapter(adapter)
	vp.SetupWithTabLayout(tl)

	// TabLayout should have tabs synced
	if tl.GetTabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tl.GetTabCount())
	}
	if tl.tabs[0].Text != "Tab 0" {
		t.Errorf("expected tab text 'Tab 0', got %q", tl.tabs[0].Text)
	}

	// Selecting tab should change page
	tl.SetSelectedTab(1)
	if vp.GetCurrentPage() != 1 {
		t.Errorf("expected page 1 after tab select, got %d", vp.GetCurrentPage())
	}

	// Changing page should update tab
	vp.SetCurrentPage(2)
	if tl.GetSelectedTab() != 2 {
		t.Errorf("expected tab 2 after page change, got %d", tl.GetSelectedTab())
	}
}
