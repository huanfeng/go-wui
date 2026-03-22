package core

import "testing"

func TestEventDispatch_BubbleUp(t *testing.T) {
	root := NewNode("Root")
	child := NewNode("Child")
	leaf := NewNode("Leaf")
	root.AddChild(child)
	child.AddChild(leaf)

	root.SetBounds(Rect{X: 0, Y: 0, Width: 200, Height: 200})
	child.SetBounds(Rect{X: 0, Y: 0, Width: 200, Height: 200})
	leaf.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})

	var received []string
	child.SetHandler(&testHandler{onEvent: func(n *Node, e interface{}) bool {
		received = append(received, "child")
		return false
	}})
	root.SetHandler(&testHandler{onEvent: func(n *Node, e interface{}) bool {
		received = append(received, "root")
		return true
	}})

	event := NewMotionEvent(ActionDown, 50, 50)
	DispatchEvent(root, event, Point{X: 50, Y: 50})

	if len(received) != 2 || received[0] != "child" || received[1] != "root" {
		t.Errorf("expected [child root], got %v", received)
	}
}

func TestEventDispatch_Consumed(t *testing.T) {
	root := NewNode("Root")
	child := NewNode("Child")
	root.AddChild(child)

	root.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})
	child.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})

	rootCalled := false
	child.SetHandler(&testHandler{onEvent: func(n *Node, e interface{}) bool {
		return true // consumed
	}})
	root.SetHandler(&testHandler{onEvent: func(n *Node, e interface{}) bool {
		rootCalled = true
		return false
	}})

	event := NewMotionEvent(ActionDown, 50, 50)
	DispatchEvent(root, event, Point{X: 50, Y: 50})
	if rootCalled {
		t.Error("root should not be called when child consumed event")
	}
}

func TestEventDispatch_Intercept(t *testing.T) {
	root := NewNode("Root")
	child := NewNode("Child")
	leaf := NewNode("Leaf")
	root.AddChild(child)
	child.AddChild(leaf)

	root.SetBounds(Rect{X: 0, Y: 0, Width: 200, Height: 200})
	child.SetBounds(Rect{X: 0, Y: 0, Width: 200, Height: 200})
	leaf.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})

	leafCalled := false
	leaf.SetHandler(&testHandler{onEvent: func(n *Node, e interface{}) bool {
		leafCalled = true
		return true
	}})

	var interceptedBy string
	child.SetHandler(&interceptHandler{
		onIntercept: func(n *Node, e interface{}) bool {
			return true // intercept
		},
		onEvent: func(n *Node, e interface{}) bool {
			interceptedBy = "child"
			return true
		},
	})

	event := NewMotionEvent(ActionDown, 50, 50)
	DispatchEvent(root, event, Point{X: 50, Y: 50})

	if leafCalled {
		t.Error("leaf should not be called when parent intercepts")
	}
	if interceptedBy != "child" {
		t.Errorf("expected child to handle after intercept, got %q", interceptedBy)
	}
}

func TestEventDispatch_MissedHit(t *testing.T) {
	root := NewNode("Root")
	root.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})

	event := NewMotionEvent(ActionDown, 200, 200)
	result := DispatchEvent(root, event, Point{X: 200, Y: 200})
	if result {
		t.Error("event outside bounds should not be dispatched")
	}
}

func TestEventDispatch_InvisibleNode(t *testing.T) {
	root := NewNode("Root")
	root.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})
	root.SetVisibility(Invisible)

	event := NewMotionEvent(ActionDown, 50, 50)
	result := DispatchEvent(root, event, Point{X: 50, Y: 50})
	if result {
		t.Error("invisible node should not receive events")
	}
}

func TestMotionEvent_Fields(t *testing.T) {
	e := NewMotionEvent(ActionDown, 10, 20)
	if e.Type() != EventMotion {
		t.Error("type should be EventMotion")
	}
	if e.Action != ActionDown {
		t.Error("action should be ActionDown")
	}
	if e.X != 10 || e.Y != 20 {
		t.Error("coordinates should match")
	}
	if e.Source != PointerMouse {
		t.Error("default source should be PointerMouse")
	}
	if e.Pressure != 1.0 {
		t.Error("default pressure should be 1.0")
	}
	if e.IsConsumed() {
		t.Error("should not be consumed initially")
	}
	e.Consume()
	if !e.IsConsumed() {
		t.Error("should be consumed after Consume()")
	}
}

type testHandler struct {
	DefaultHandler
	onEvent func(*Node, interface{}) bool
}

func (h *testHandler) OnEvent(node *Node, event interface{}) bool {
	if h.onEvent != nil {
		return h.onEvent(node, event)
	}
	return false
}

type interceptHandler struct {
	DefaultHandler
	onIntercept func(*Node, interface{}) bool
	onEvent     func(*Node, interface{}) bool
}

func (h *interceptHandler) OnInterceptEvent(node *Node, event interface{}) bool {
	if h.onIntercept != nil {
		return h.onIntercept(node, event)
	}
	return false
}

func (h *interceptHandler) OnEvent(node *Node, event interface{}) bool {
	if h.onEvent != nil {
		return h.onEvent(node, event)
	}
	return false
}
