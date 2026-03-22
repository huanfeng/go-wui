package core

// Handler processes events on a node using typed Event interface.
type Handler interface {
	OnDispatchEvent(node *Node, event Event) bool
	OnInterceptEvent(node *Node, event Event) bool
	OnEvent(node *Node, event Event) bool
}

// DefaultHandler is a no-op handler that can be embedded by widgets.
type DefaultHandler struct{}

func (h *DefaultHandler) OnDispatchEvent(node *Node, event Event) bool  { return false }
func (h *DefaultHandler) OnInterceptEvent(node *Node, event Event) bool { return false }
func (h *DefaultHandler) OnEvent(node *Node, event Event) bool          { return false }
