package widget

import "testing"

func TestEditTextSetText(t *testing.T) {
	et := NewEditText("Enter name")
	if et.GetText() != "" {
		t.Error("should start empty")
	}
	et.SetText("Hello")
	if et.GetText() != "Hello" {
		t.Error("text mismatch")
	}
}

func TestEditTextHint(t *testing.T) {
	et := NewEditText("Placeholder")
	if et.Node().GetDataString("hint") != "Placeholder" {
		t.Error("hint mismatch")
	}
	if et.GetHint() != "Placeholder" {
		t.Error("GetHint mismatch")
	}
}

func TestEditTextViewInterface(t *testing.T) {
	et := NewEditText("Test")
	et.SetId("et_test")
	if et.GetId() != "et_test" {
		t.Error("id mismatch")
	}
	if et.Node() == nil {
		t.Error("node should not be nil")
	}
	if et.Node().GetView() != et {
		t.Error("node's view should point back to EditText")
	}
}

func TestEditTextSetHint(t *testing.T) {
	et := NewEditText("initial")
	et.SetHint("updated")
	if et.GetHint() != "updated" {
		t.Error("hint should be updated")
	}
	if et.Node().GetDataString("hint") != "updated" {
		t.Error("node data hint should be updated")
	}
}

func TestEditTextNodeTag(t *testing.T) {
	et := NewEditText("")
	if et.Node().Tag() != "EditText" {
		t.Errorf("expected tag EditText, got %s", et.Node().Tag())
	}
}
