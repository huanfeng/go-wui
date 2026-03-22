package widget

import "testing"

func TestTextViewSetText(t *testing.T) {
	tv := NewTextView("Hello")
	if tv.GetText() != "Hello" {
		t.Error("initial text mismatch")
	}
	tv.SetText("World")
	if tv.GetText() != "World" {
		t.Error("updated text mismatch")
	}
}

func TestTextViewViewInterface(t *testing.T) {
	tv := NewTextView("Test")
	tv.SetId("mytext")
	if tv.GetId() != "mytext" {
		t.Error("id mismatch")
	}
	if tv.Node() == nil {
		t.Error("node should not be nil")
	}
	if tv.Node().GetView() != tv {
		t.Error("node's view should point back to TextView")
	}
}
