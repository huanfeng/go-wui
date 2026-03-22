package core

// Visibility controls whether a node is drawn and takes up layout space.
type Visibility int

const (
	Visible   Visibility = iota // Drawn and takes space
	Invisible                   // Not drawn but takes space
	Gone                        // Not drawn and takes no space
)

// View is the public interface for all UI elements (widgets).
type View interface {
	Node() *Node
	SetId(id string)
	GetId() string
	SetVisibility(v Visibility)
	GetVisibility() Visibility
	SetEnabled(enabled bool)
	IsEnabled() bool
}

// Node is the internal tree element that backs every View.
type Node struct {
	parent   *Node
	children []*Node

	bounds  Rect
	padding Insets
	margin  Insets

	layout  Layout
	painter Painter
	handler Handler
	style   *Style

	id         string
	tag        string
	visibility Visibility
	enabled    bool
	dirty      bool
	childDirty bool

	// View association — the widget wrapping this node
	view View

	measuredSize Size
	data         map[string]interface{}
}

// NewNode creates a new Node with the given tag name.
func NewNode(tag string) *Node {
	return &Node{
		tag:        tag,
		visibility: Visible,
		enabled:    true,
		dirty:      true,
		data:       make(map[string]interface{}),
	}
}

// ---------- Tree operations ----------

// Parent returns the parent node, or nil if this is a root.
func (n *Node) Parent() *Node {
	return n.parent
}

// Children returns the list of child nodes.
func (n *Node) Children() []*Node {
	return n.children
}

// Tag returns the tag name of this node (e.g. "LinearLayout", "TextView").
func (n *Node) Tag() string {
	return n.tag
}

// AddChild appends a child node and sets its parent.
func (n *Node) AddChild(child *Node) {
	if child == nil {
		return
	}
	// Remove from previous parent if any
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}
	child.parent = n
	n.children = append(n.children, child)
	n.MarkDirty()
}

// RemoveChild removes a child node and clears its parent reference.
func (n *Node) RemoveChild(child *Node) {
	if child == nil {
		return
	}
	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			child.parent = nil
			n.MarkDirty()
			return
		}
	}
}

// ---------- ID ----------

// SetId sets the identifier for this node.
func (n *Node) SetId(id string) {
	n.id = id
}

// GetId returns the identifier of this node.
func (n *Node) GetId() string {
	return n.id
}

// Id is an alias for GetId.
func (n *Node) Id() string {
	return n.id
}

// ---------- Visibility ----------

// SetVisibility sets the visibility of this node.
func (n *Node) SetVisibility(v Visibility) {
	if n.visibility != v {
		n.visibility = v
		n.MarkDirty()
	}
}

// GetVisibility returns the visibility of this node.
func (n *Node) GetVisibility() Visibility {
	return n.visibility
}

// ---------- Enabled ----------

// SetEnabled sets whether this node is enabled for interaction.
func (n *Node) SetEnabled(b bool) {
	n.enabled = b
}

// IsEnabled reports whether this node is enabled.
func (n *Node) IsEnabled() bool {
	return n.enabled
}

// ---------- Component accessors ----------

// SetLayout sets the layout strategy for this node.
func (n *Node) SetLayout(l Layout) {
	n.layout = l
}

// GetLayout returns the layout strategy, or nil.
func (n *Node) GetLayout() Layout {
	return n.layout
}

// SetPainter sets the painter for this node.
func (n *Node) SetPainter(p Painter) {
	n.painter = p
}

// GetPainter returns the painter, or nil.
func (n *Node) GetPainter() Painter {
	return n.painter
}

// SetHandler sets the event handler for this node.
func (n *Node) SetHandler(h Handler) {
	n.handler = h
}

// GetHandler returns the event handler, or nil.
func (n *Node) GetHandler() Handler {
	return n.handler
}

// SetStyle sets the style for this node.
func (n *Node) SetStyle(s *Style) {
	n.style = s
}

// GetStyle returns the style, or nil.
func (n *Node) GetStyle() *Style {
	return n.style
}

// ---------- Geometry ----------

// Bounds returns the layout bounds of this node.
func (n *Node) Bounds() Rect {
	return n.bounds
}

// SetBounds sets the layout bounds of this node.
func (n *Node) SetBounds(r Rect) {
	n.bounds = r
}

// Padding returns the padding insets.
func (n *Node) Padding() Insets {
	return n.padding
}

// SetPadding sets the padding insets.
func (n *Node) SetPadding(insets Insets) {
	n.padding = insets
}

// Margin returns the margin insets.
func (n *Node) Margin() Insets {
	return n.margin
}

// SetMargin sets the margin insets.
func (n *Node) SetMargin(insets Insets) {
	n.margin = insets
}

// MeasuredSize returns the size computed during the measure pass.
func (n *Node) MeasuredSize() Size {
	return n.measuredSize
}

// SetMeasuredSize stores the size computed during the measure pass.
func (n *Node) SetMeasuredSize(s Size) {
	n.measuredSize = s
}

// ---------- Data store ----------

// SetData stores an arbitrary value by key.
func (n *Node) SetData(key string, value interface{}) {
	if n.data == nil {
		n.data = make(map[string]interface{})
	}
	n.data[key] = value
}

// GetData retrieves a value by key, or nil if not found.
func (n *Node) GetData(key string) interface{} {
	if n.data == nil {
		return nil
	}
	return n.data[key]
}

// GetDataString retrieves a string value by key, or "" if not found or not a string.
func (n *Node) GetDataString(key string) string {
	v := n.GetData(key)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// ---------- View association ----------

// SetView associates a View with this node.
func (n *Node) SetView(v View) {
	n.view = v
}

// GetView returns the associated View, or nil.
func (n *Node) GetView() View {
	return n.view
}

// ---------- Dirty tracking ----------

// MarkDirty marks this node as dirty and bubbles childDirty up to all ancestors.
func (n *Node) MarkDirty() {
	n.dirty = true
	// Bubble childDirty up to ancestors
	for p := n.parent; p != nil; p = p.parent {
		if p.childDirty {
			break // Already marked — ancestors above are also marked
		}
		p.childDirty = true
	}
}

// IsDirty reports whether this node itself is dirty.
func (n *Node) IsDirty() bool {
	return n.dirty
}

// IsChildDirty reports whether any descendant is dirty.
func (n *Node) IsChildDirty() bool {
	return n.childDirty
}

// ClearDirty clears the dirty and childDirty flags on this node.
func (n *Node) ClearDirty() {
	n.dirty = false
	n.childDirty = false
}

// ---------- Search ----------

// FindNodeById searches this node and its descendants for a node with the given id.
// Returns nil if not found.
func (n *Node) FindNodeById(id string) *Node {
	if n.id == id {
		return n
	}
	for _, child := range n.children {
		if found := child.FindNodeById(id); found != nil {
			return found
		}
	}
	return nil
}

// FindViewById searches this node and its descendants for a node with the given id
// that has an associated View. Returns nil if not found or if the node has no View.
func (n *Node) FindViewById(id string) View {
	node := n.FindNodeById(id)
	if node == nil {
		return nil
	}
	return node.view
}
