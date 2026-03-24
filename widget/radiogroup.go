package widget

import (
	"github.com/huanfeng/go-wui/core"
	"github.com/huanfeng/go-wui/layout"
)

// RadioGroup manages mutual exclusion of RadioButtons.
// It acts as a vertical LinearLayout container for RadioButton children.
type RadioGroup struct {
	BaseView
	buttons    []*RadioButton
	selectedId int // index of selected button, -1 for none
	onChanged  func(index int)
}

// NewRadioGroup creates a new RadioGroup with a vertical LinearLayout.
func NewRadioGroup() *RadioGroup {
	rg := &RadioGroup{
		selectedId: -1,
	}
	rg.node = initNode("RadioGroup", rg)
	rg.node.SetStyle(&core.Style{})
	rg.node.SetLayout(&layout.LinearLayout{
		Orientation: layout.Vertical,
		Spacing:     8,
	})
	return rg
}

// AddButton adds a RadioButton to the group.
// It registers the button, sets the group back-reference, and adds it as a child node.
func (rg *RadioGroup) AddButton(rb *RadioButton) {
	rb.group = rg
	rg.buttons = append(rg.buttons, rb)
	rg.node.AddChild(rb.Node())
}

// RegisterButton registers an existing child node as a managed RadioButton
// without re-adding it as a child (it's already a child from XML inflation).
func (rg *RadioGroup) RegisterButton(rb *RadioButton) {
	rb.group = rg
	rg.buttons = append(rg.buttons, rb)
}

// GetSelectedIndex returns the index of the currently selected button, or -1 if none.
func (rg *RadioGroup) GetSelectedIndex() int {
	return rg.selectedId
}

// SetSelectedIndex selects the button at the given index.
// Pass -1 to deselect all. Out-of-range indices are ignored.
func (rg *RadioGroup) SetSelectedIndex(index int) {
	if index < -1 || index >= len(rg.buttons) {
		return
	}

	// Deselect all buttons
	for _, b := range rg.buttons {
		b.SetSelected(false)
	}

	rg.selectedId = index

	// Select the target button
	if index >= 0 {
		rg.buttons[index].SetSelected(true)
	}

	if rg.onChanged != nil {
		rg.onChanged(index)
	}
}

// SetOnChanged sets the callback invoked when the selected button changes.
// The callback receives the index of the newly selected button.
func (rg *RadioGroup) SetOnChanged(fn func(index int)) {
	rg.onChanged = fn
}

// selectButton is called internally by a RadioButton's click handler.
// It deselects all buttons, selects the given one, and fires the onChanged callback.
func (rg *RadioGroup) selectButton(rb *RadioButton) {
	idx := -1
	for i, b := range rg.buttons {
		if b == rb {
			idx = i
		}
		b.SetSelected(false)
	}

	if idx >= 0 {
		rb.SetSelected(true)
		rg.selectedId = idx
		if rb.onChanged != nil {
			rb.onChanged(true)
		}
		if rg.onChanged != nil {
			rg.onChanged(idx)
		}
	}
}
