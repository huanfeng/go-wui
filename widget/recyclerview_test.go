package widget

import (
	"fmt"
	"testing"

	"gowui/core"
)

// testRecyclerAdapter is a simple adapter for testing.
type testRecyclerAdapter struct {
	count       int
	createCalls int
	bindCalls   int
}

func (a *testRecyclerAdapter) GetItemCount() int         { return a.count }
func (a *testRecyclerAdapter) GetItemViewType(pos int) int { return 0 }
func (a *testRecyclerAdapter) CreateViewHolder(viewType int) *ViewHolder {
	a.createCalls++
	tv := NewTextView("")
	return &ViewHolder{ItemView: tv}
}
func (a *testRecyclerAdapter) BindViewHolder(holder *ViewHolder, position int) {
	a.bindCalls++
	if tv, ok := holder.ItemView.(*TextView); ok {
		tv.SetText(fmt.Sprintf("Item %d", position))
	}
}

func TestNewRecyclerView(t *testing.T) {
	rv := NewRecyclerView(48)
	if rv.Node().Tag() != "RecyclerView" {
		t.Errorf("expected tag 'RecyclerView', got %q", rv.Node().Tag())
	}
	if rv.GetItemHeight() != 48 {
		t.Errorf("expected item height 48, got %f", rv.GetItemHeight())
	}
	if rv.GetScrollY() != 0 {
		t.Errorf("expected scroll 0, got %f", rv.GetScrollY())
	}
}

func TestRecyclerViewDefaultItemHeight(t *testing.T) {
	rv := NewRecyclerView(0) // should default to 48
	if rv.GetItemHeight() != 48 {
		t.Errorf("expected default item height 48, got %f", rv.GetItemHeight())
	}
}

func TestRecyclerViewSetAdapter(t *testing.T) {
	rv := NewRecyclerView(40)
	adapter := &testRecyclerAdapter{count: 100}
	rv.SetAdapter(adapter)

	if rv.GetAdapter() != adapter {
		t.Error("expected adapter to be set")
	}
}

func TestRecyclerViewTotalContentHeight(t *testing.T) {
	rv := NewRecyclerView(40)
	adapter := &testRecyclerAdapter{count: 50}
	rv.SetAdapter(adapter)

	expected := 50.0 * 40.0
	if rv.totalContentHeight() != expected {
		t.Errorf("expected total height %f, got %f", expected, rv.totalContentHeight())
	}
}

func TestRecyclerViewScrollToPosition(t *testing.T) {
	rv := NewRecyclerView(40)
	adapter := &testRecyclerAdapter{count: 100}
	rv.SetAdapter(adapter)
	rv.node.SetMeasuredSize(core.Size{Width: 300, Height: 400})

	rv.ScrollToPosition(10)
	if rv.GetScrollY() != 400 { // 10 * 40
		t.Errorf("expected scroll 400, got %f", rv.GetScrollY())
	}
}

func TestRecyclerViewClampScroll(t *testing.T) {
	rv := NewRecyclerView(40)
	adapter := &testRecyclerAdapter{count: 5}
	rv.SetAdapter(adapter)
	rv.node.SetMeasuredSize(core.Size{Width: 300, Height: 400})

	// Total content = 5*40 = 200, viewport = 400 → no scroll possible
	rv.scrollY = 100
	rv.clampScroll()
	if rv.scrollY != 0 {
		t.Errorf("expected scroll clamped to 0, got %f", rv.scrollY)
	}

	// Negative scroll
	rv.scrollY = -50
	rv.clampScroll()
	if rv.scrollY != 0 {
		t.Errorf("expected scroll clamped to 0, got %f", rv.scrollY)
	}
}

func TestRecyclerViewRecyclePool(t *testing.T) {
	pool := newRecyclerPool()

	tv := NewTextView("test")
	vh := &ViewHolder{ItemView: tv, viewType: 0, position: 5, bound: true}

	pool.put(vh)
	if vh.bound {
		t.Error("expected bound reset after put")
	}

	got := pool.get(0)
	if got != vh {
		t.Error("expected to get same holder from pool")
	}

	// Pool empty now
	if pool.get(0) != nil {
		t.Error("expected nil from empty pool")
	}
}

func TestRecyclerViewNotifyDataSetChanged(t *testing.T) {
	rv := NewRecyclerView(40)
	adapter := &testRecyclerAdapter{count: 10}
	rv.SetAdapter(adapter)

	// Simulate some active holders
	vh := adapter.CreateViewHolder(0)
	rv.activeHolders[0] = vh

	rv.NotifyDataSetChanged()
	if len(rv.activeHolders) != 0 {
		t.Errorf("expected 0 active holders after notify, got %d", len(rv.activeHolders))
	}
}

func TestRecyclerViewPositionAtY(t *testing.T) {
	rv := NewRecyclerView(40)
	adapter := &testRecyclerAdapter{count: 10}
	rv.SetAdapter(adapter)

	pos := rv.positionAtY(80) // scrollY=0, y=80 → position 2
	if pos != 2 {
		t.Errorf("expected position 2, got %d", pos)
	}

	rv.scrollY = 40
	pos = rv.positionAtY(80) // scrollY=40, y=80 → (80+40)/40 = 3
	if pos != 3 {
		t.Errorf("expected position 3, got %d", pos)
	}
}

func TestRecyclerViewOnItemClick(t *testing.T) {
	rv := NewRecyclerView(40)
	clickedPos := -1
	rv.SetOnItemClickListener(func(pos int) {
		clickedPos = pos
	})

	rv.onItemClick(5)
	if clickedPos != 5 {
		t.Errorf("expected clicked position 5, got %d", clickedPos)
	}
}

func TestRecyclerViewSetItemHeight(t *testing.T) {
	rv := NewRecyclerView(40)
	rv.SetItemHeight(60)
	if rv.GetItemHeight() != 60 {
		t.Errorf("expected item height 60, got %f", rv.GetItemHeight())
	}
}
