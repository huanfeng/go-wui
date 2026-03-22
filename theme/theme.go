package theme

import (
	"encoding/xml"
	"strings"
)

// Theme holds semantic color/style attributes resolved from a named style.
type Theme struct {
	Name  string
	attrs map[string]string
}

// ResolveAttr returns the value of a theme attribute, or "" if not found.
func (t *Theme) ResolveAttr(key string) string {
	if t.attrs == nil {
		return ""
	}
	return t.attrs[key]
}

// StyleEntry is a named style with an optional explicit parent.
type StyleEntry struct {
	Name   string
	Parent string
	Items  map[string]string
}

// StyleRegistry stores named styles and resolves parent chains.
type StyleRegistry struct {
	styles map[string]*StyleEntry
}

// NewStyleRegistry creates an empty StyleRegistry.
func NewStyleRegistry() *StyleRegistry {
	return &StyleRegistry{styles: make(map[string]*StyleEntry)}
}

// Register adds a StyleEntry to the registry.
func (sr *StyleRegistry) Register(entry *StyleEntry) {
	sr.styles[entry.Name] = entry
}

// Resolve returns the fully merged attribute map for a named style,
// with child values overriding parent values.
func (sr *StyleRegistry) Resolve(name string) map[string]string {
	result := make(map[string]string)
	sr.resolveChain(name, result, 10) // max depth 10 to prevent loops
	return result
}

// resolveChain recursively resolves a style's parent chain.
func (sr *StyleRegistry) resolveChain(name string, result map[string]string, depth int) {
	if depth <= 0 {
		return
	}
	entry, ok := sr.styles[name]
	if !ok {
		return
	}

	// Determine parent name
	parentName := entry.Parent
	if parentName == "" {
		// Implicit inheritance via dot notation: "Widget.Button.Outlined" -> parent "Widget.Button"
		if idx := strings.LastIndex(name, "."); idx > 0 {
			parentName = name[:idx]
		}
	}

	// Resolve parent first (so child overrides parent values)
	if parentName != "" {
		sr.resolveChain(parentName, result, depth-1)
	}

	// Apply this style's items (overrides parent)
	for k, v := range entry.Items {
		result[k] = v
	}
}

// --- XML parsing ---

type xmlStyleResources struct {
	XMLName xml.Name       `xml:"resources"`
	Styles  []xmlStyleDef  `xml:"style"`
}

type xmlStyleDef struct {
	Name   string           `xml:"name,attr"`
	Parent string           `xml:"parent,attr"`
	Items  []xmlStyleItem   `xml:"item"`
}

type xmlStyleItem struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// LoadStylesFromXML parses style definitions from XML data into a StyleRegistry.
func LoadStylesFromXML(data []byte) (*StyleRegistry, error) {
	var res xmlStyleResources
	if err := xml.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	reg := NewStyleRegistry()
	for _, s := range res.Styles {
		items := make(map[string]string)
		for _, item := range s.Items {
			items[item.Name] = item.Value
		}
		reg.Register(&StyleEntry{
			Name:   s.Name,
			Parent: s.Parent,
			Items:  items,
		})
	}
	return reg, nil
}

// LoadThemeFromXML loads a named theme from XML style data.
// It resolves the full parent chain and returns a Theme with all merged attributes.
func LoadThemeFromXML(data []byte, themeName string) (*Theme, error) {
	reg, err := LoadStylesFromXML(data)
	if err != nil {
		return nil, err
	}
	attrs := reg.Resolve(themeName)
	return &Theme{Name: themeName, attrs: attrs}, nil
}
