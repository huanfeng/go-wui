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

// getDPIScale 返回节点的 DPI 缩放系数。
// 优先直接读取当前节点已存储的 dpiScale 值（O(1)），
// 仅当当前节点未设置时才向上遍历父链，并将找到的值缓存到当前节点，避免重复遍历。
func getDPIScale(node *core.Node) float64 {
	if s, ok := node.GetData("dpiScale").(float64); ok && s > 0 {
		return s
	}
	// 向上查找并缓存到当前节点，加速后续调用
	for n := node.Parent(); n != nil; n = n.Parent() {
		if s, ok := n.GetData("dpiScale").(float64); ok && s > 0 {
			node.SetData("dpiScale", s)
			return s
		}
	}
	return 1.0
}
