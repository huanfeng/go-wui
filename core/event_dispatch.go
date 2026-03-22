package core

// DispatchEvent dispatches a pointer event through the node tree using
// Android-style 3-phase dispatch:
//  1. Hit test: walk down the tree, find deepest node whose bounds contain the point.
//  2. Intercept: walk from root toward target; each ancestor may intercept.
//  3. Handle + Bubble: target handles first; if not consumed, bubble up to parent.
//
// hitPoint is in the coordinate space of the root node.
func DispatchEvent(root *Node, event Event, hitPoint Point) bool {
	// Phase 1: Build hit chain — root -> ... -> deepest hit node
	chain := buildHitChain(root, hitPoint, Point{})
	if len(chain) == 0 {
		return false
	}

	// Phase 2: Intercept (from root to parent of target)
	for i := 0; i < len(chain)-1; i++ {
		node := chain[i]
		if h := node.GetHandler(); h != nil {
			if h.OnInterceptEvent(node, event) {
				// Intercepted — this node becomes the target
				chain = chain[:i+1]
				break
			}
		}
	}

	// Phase 3: Handle + Bubble (from target up to root)
	for i := len(chain) - 1; i >= 0; i-- {
		node := chain[i]
		if h := node.GetHandler(); h != nil {
			if h.OnEvent(node, event) {
				return true // consumed
			}
		}
	}
	return false
}

// buildHitChain finds the path from root to the deepest node containing the point.
func buildHitChain(node *Node, point Point, offset Point) []*Node {
	if node.GetVisibility() != Visible {
		return nil
	}
	b := node.Bounds()
	localX := point.X - offset.X
	localY := point.Y - offset.Y
	if !b.Contains(localX, localY) {
		return nil
	}

	chain := []*Node{node}
	childOffset := Point{X: offset.X + b.X, Y: offset.Y + b.Y}
	// Check children in reverse order (topmost/last child first)
	for i := len(node.Children()) - 1; i >= 0; i-- {
		child := node.Children()[i]
		childChain := buildHitChain(child, point, childOffset)
		if len(childChain) > 0 {
			return append(chain, childChain...)
		}
	}
	return chain
}
