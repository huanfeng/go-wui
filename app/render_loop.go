package app

import "github.com/huanfeng/wind-ui/core"

// PaintNode 递归绘制节点树到画布上（完整重绘路径）。
// 实现已统一到 core.PaintNode，此处保留为兼容性包装。
func PaintNode(node *core.Node, canvas core.Canvas) {
	core.PaintNode(node, canvas)
}

