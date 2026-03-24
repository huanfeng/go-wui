package res

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io/fs"
	"strconv"
	"strings"

	"github.com/huanfeng/go-wui/core"
)

// ViewFactory creates a Node from an AttributeSet parsed from XML.
type ViewFactory func(attrs *AttributeSet) *core.Node

// AttributeSet provides typed access to XML attributes, with resource reference resolution.
type AttributeSet struct {
	attrs map[string]string
	res   *ResourceManager
}

// NewAttributeSet creates an AttributeSet from raw attribute key-value pairs.
func NewAttributeSet(attrs map[string]string, res *ResourceManager) *AttributeSet {
	return &AttributeSet{attrs: attrs, res: res}
}

// GetString returns the string value for the given attribute name.
// Resource references (@string/key etc.) are resolved if a ResourceManager is available.
func (a *AttributeSet) GetString(name string) string {
	v, ok := a.attrs[name]
	if !ok {
		return ""
	}
	if a.res != nil {
		return a.res.ResolveRef(v)
	}
	return v
}

// GetDimension parses the attribute as a Dimension.
func (a *AttributeSet) GetDimension(name string) core.Dimension {
	v := a.GetString(name)
	if v == "" {
		return core.Dimension{}
	}
	return core.ParseDimension(v)
}

// GetColor parses the attribute as an RGBA color.
func (a *AttributeSet) GetColor(name string) color.RGBA {
	v := a.GetString(name)
	if v == "" {
		return color.RGBA{}
	}
	return core.ParseColor(v)
}

// GetFloat parses the attribute as a float64.
func (a *AttributeSet) GetFloat(name string) float64 {
	v := a.GetString(name)
	if v == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

// GetBool parses the attribute as a boolean.
func (a *AttributeSet) GetBool(name string) bool {
	v := a.GetString(name)
	return v == "true"
}

// GetInt parses the attribute as an int.
func (a *AttributeSet) GetInt(name string) int {
	v := a.GetString(name)
	if v == "" {
		return 0
	}
	i, _ := strconv.Atoi(v)
	return i
}

// LayoutInflater parses XML layout files into Node trees.
type LayoutInflater struct {
	resourceManager *ResourceManager
	viewRegistry    map[string]ViewFactory
}

// NewLayoutInflater creates a new LayoutInflater with the given ResourceManager.
func NewLayoutInflater(rm *ResourceManager) *LayoutInflater {
	return &LayoutInflater{
		resourceManager: rm,
		viewRegistry:    make(map[string]ViewFactory),
	}
}

// RegisterView registers a ViewFactory for a given XML tag name.
func (li *LayoutInflater) RegisterView(tag string, factory ViewFactory) {
	li.viewRegistry[tag] = factory
}

// Inflate loads a layout from a resource reference like "@layout/main".
func (li *LayoutInflater) Inflate(ref string) *core.Node {
	name := strings.TrimPrefix(ref, "@layout/")
	if li.resourceManager == nil || li.resourceManager.Embedded() == nil {
		return nil
	}
	data, err := fs.ReadFile(li.resourceManager.Embedded(), "layout/"+name+".xml")
	if err != nil {
		return nil
	}
	node, _ := li.InflateFromBytes(data)
	return node
}

// InflateFromString parses an XML layout string into a Node tree.
func (li *LayoutInflater) InflateFromString(xmlStr string) (*core.Node, error) {
	return li.InflateFromBytes([]byte(xmlStr))
}

// InflateFromBytes parses XML layout bytes into a Node tree.
func (li *LayoutInflater) InflateFromBytes(data []byte) (*core.Node, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	node, err := li.inflateElement(decoder)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// inflateElement recursively reads XML tokens and builds the node tree.
func (li *LayoutInflater) inflateElement(decoder *xml.Decoder) (*core.Node, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			return li.inflateStartElement(decoder, t)
		case xml.EndElement:
			// Unexpected end element at top level
			return nil, nil
		}
		// Skip CharData, Comments, ProcInst, Directive
	}
}

// inflateStartElement processes a start element and its children.
func (li *LayoutInflater) inflateStartElement(decoder *xml.Decoder, start xml.StartElement) (*core.Node, error) {
	tag := start.Name.Local

	// Handle <include> tag
	if tag == "include" {
		return li.inflateInclude(start)
	}

	// Build attribute map
	attrs := make(map[string]string)
	for _, attr := range start.Attr {
		// Strip namespace prefix if present (e.g. "android:text" -> "text")
		name := attr.Name.Local
		attrs[name] = attr.Value
	}

	attrSet := NewAttributeSet(attrs, li.resourceManager)

	// Look up factory for this tag
	var node *core.Node
	factory, ok := li.viewRegistry[tag]
	if ok {
		node = factory(attrSet)
	} else {
		// Unknown tag: create a plain node
		node = core.NewNode(tag)
		node.SetStyle(&core.Style{})
		applyCommonAttrs(node, attrSet)
	}

	if node == nil {
		// Factory returned nil — skip children
		li.skipElement(decoder)
		return nil, fmt.Errorf("factory for %q returned nil", tag)
	}

	// Process children until we hit the matching end element
	for {
		token, err := decoder.Token()
		if err != nil {
			return node, nil // EOF is ok for self-closing tags
		}

		switch t := token.(type) {
		case xml.StartElement:
			child, err := li.inflateStartElement(decoder, t)
			if err != nil {
				continue
			}
			if child != nil {
				node.AddChild(child)
			}
		case xml.EndElement:
			return node, nil
		}
	}
}

// inflateInclude handles <include layout="@layout/name" /> tags.
func (li *LayoutInflater) inflateInclude(start xml.StartElement) (*core.Node, error) {
	var layoutRef string
	for _, attr := range start.Attr {
		if attr.Name.Local == "layout" {
			layoutRef = attr.Value
			break
		}
	}
	if layoutRef == "" {
		return nil, fmt.Errorf("include missing layout attribute")
	}
	node := li.Inflate(layoutRef)
	return node, nil
}

// skipElement skips tokens until the matching end element is consumed.
func (li *LayoutInflater) skipElement(decoder *xml.Decoder) {
	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			return
		}
		switch token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
}

// applyCommonAttrs applies standard attributes (id, width, height, padding, margin,
// background, visibility) to a node.
func applyCommonAttrs(node *core.Node, attrs *AttributeSet) {
	// ID
	if id := attrs.GetString("id"); id != "" {
		node.SetId(id)
	}

	// Style setup
	style := node.GetStyle()
	if style == nil {
		style = &core.Style{}
		node.SetStyle(style)
	}

	// Width
	if w := attrs.GetString("width"); w != "" {
		style.Width = core.ParseDimension(w)
	}

	// Height
	if h := attrs.GetString("height"); h != "" {
		style.Height = core.ParseDimension(h)
	}

	// Weight — also mark the appropriate dimension as DimensionWeight so that
	// LinearLayout distributes remaining space to this child.
	if wt := attrs.GetFloat("layout_weight"); wt > 0 {
		style.Weight = wt
		// In a horizontal LinearLayout the width is distributed; in vertical the height.
		// We mark whichever axis is set to 0 (or explicitly 0dp) as weight-based.
		if style.Width.Value == 0 && style.Width.Unit != core.DimensionMatchParent && style.Width.Unit != core.DimensionWrapContent {
			style.Width = core.Dimension{Value: wt, Unit: core.DimensionWeight}
		}
		if style.Height.Value == 0 && style.Height.Unit != core.DimensionMatchParent && style.Height.Unit != core.DimensionWrapContent {
			style.Height = core.Dimension{Value: wt, Unit: core.DimensionWeight}
		}
	}

	// Padding (uniform)
	if p := attrs.GetDimension("padding"); p.Value > 0 {
		node.SetPadding(core.Insets{
			Left:   p.Value,
			Top:    p.Value,
			Right:  p.Value,
			Bottom: p.Value,
		})
	}

	// Individual padding sides
	pad := node.Padding()
	if v := attrs.GetDimension("paddingLeft"); v.Value > 0 {
		pad.Left = v.Value
	}
	if v := attrs.GetDimension("paddingTop"); v.Value > 0 {
		pad.Top = v.Value
	}
	if v := attrs.GetDimension("paddingRight"); v.Value > 0 {
		pad.Right = v.Value
	}
	if v := attrs.GetDimension("paddingBottom"); v.Value > 0 {
		pad.Bottom = v.Value
	}
	node.SetPadding(pad)

	// Margin (uniform)
	if m := attrs.GetDimension("margin"); m.Value > 0 {
		node.SetMargin(core.Insets{
			Left:   m.Value,
			Top:    m.Value,
			Right:  m.Value,
			Bottom: m.Value,
		})
	}

	// Individual margin sides
	margin := node.Margin()
	if v := attrs.GetDimension("marginLeft"); v.Value > 0 {
		margin.Left = v.Value
	}
	if v := attrs.GetDimension("marginTop"); v.Value > 0 {
		margin.Top = v.Value
	}
	if v := attrs.GetDimension("marginRight"); v.Value > 0 {
		margin.Right = v.Value
	}
	if v := attrs.GetDimension("marginBottom"); v.Value > 0 {
		margin.Bottom = v.Value
	}
	node.SetMargin(margin)

	// Background color
	if bg := attrs.GetColor("background"); bg.A > 0 {
		style.BackgroundColor = bg
	}

	// Visibility
	if vis := attrs.GetString("visibility"); vis != "" {
		switch vis {
		case "invisible":
			node.SetVisibility(core.Invisible)
		case "gone":
			node.SetVisibility(core.Gone)
		default:
			node.SetVisibility(core.Visible)
		}
	}
}
