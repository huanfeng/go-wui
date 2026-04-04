package widget

import "github.com/huanfeng/wind-ui/core"

// BaseView provides the shared View interface implementation.
// Concrete widgets embed this.
type BaseView struct {
	node *core.Node
}

func (b *BaseView) Node() *core.Node                { return b.node }
func (b *BaseView) SetId(id string)                  { b.node.SetId(id) }
func (b *BaseView) GetId() string                    { return b.node.GetId() }
func (b *BaseView) SetVisibility(v core.Visibility)  { b.node.SetVisibility(v) }
func (b *BaseView) GetVisibility() core.Visibility   { return b.node.GetVisibility() }
func (b *BaseView) SetEnabled(enabled bool)           { b.node.SetEnabled(enabled) }
func (b *BaseView) IsEnabled() bool                   { return b.node.IsEnabled() }

// initNode creates a node, associates the view, and returns it.
func initNode(tag string, view core.View) *core.Node {
	n := core.NewNode(tag)
	n.SetView(view)
	return n
}

// getDPIScale walks up the node tree to find the DPI scale factor stored on
// an ancestor (typically the root node). Returns 1.0 if no scale is set.
func getDPIScale(node *core.Node) float64 {
	for n := node; n != nil; n = n.Parent() {
		if s, ok := n.GetData("dpiScale").(float64); ok && s > 0 {
			return s
		}
	}
	return 1.0
}
