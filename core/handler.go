package core

// Handler processes events on a node.
// Event interface is defined in core/event.go (Task 4). Using interface{} temporarily.
type Handler interface {
	OnDispatchEvent(node *Node, event interface{}) bool
	OnInterceptEvent(node *Node, event interface{}) bool
	OnEvent(node *Node, event interface{}) bool
}

// DefaultHandler is a no-op handler that can be embedded by widgets.
type DefaultHandler struct{}

func (h *DefaultHandler) OnDispatchEvent(node *Node, event interface{}) bool  { return false }
func (h *DefaultHandler) OnInterceptEvent(node *Node, event interface{}) bool { return false }
func (h *DefaultHandler) OnEvent(node *Node, event interface{}) bool          { return false }
