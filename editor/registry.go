package editor

// PropertyType defines the editor control type for a property.
type PropertyType int

const (
	PropString    PropertyType = iota // free-text input
	PropDimension                     // "24dp", "match_parent", "wrap_content"
	PropColor                         // "#RRGGBB" or "#AARRGGBB"
	PropEnum                          // dropdown with fixed options
	PropBool                          // "true" / "false"
	PropInt                           // integer
	PropFloat                         // float
)

// PropertyDef describes a single editable property of a widget.
type PropertyDef struct {
	Name    string       // XML attribute name
	Type    PropertyType // editor control type
	Options []string     // for PropEnum
	Default string       // default value shown in editor
	Group   string       // grouping label ("Layout", "Appearance", "Text", etc.)
}

// containerTags is the set of tags that accept child nodes.
var containerTags = map[string]bool{
	"LinearLayout":       true,
	"FrameLayout":        true,
	"ScrollView":         true,
	"HorizontalScrollView": true,
	"GridLayout":         true,
	"FlexLayout":         true,
	"RadioGroup":         true,
	"View":               true,
}

// IsContainerTag reports whether the tag can hold children.
func IsContainerTag(tag string) bool {
	return containerTags[tag]
}

// Common properties shared by all widgets.
var commonProps = []PropertyDef{
	{Name: "id", Type: PropString, Group: "Common"},
	{Name: "width", Type: PropDimension, Default: "wrap_content", Group: "Layout"},
	{Name: "height", Type: PropDimension, Default: "wrap_content", Group: "Layout"},
	{Name: "layout_weight", Type: PropFloat, Group: "Layout"},
	{Name: "padding", Type: PropDimension, Group: "Layout"},
	{Name: "paddingLeft", Type: PropDimension, Group: "Layout"},
	{Name: "paddingTop", Type: PropDimension, Group: "Layout"},
	{Name: "paddingRight", Type: PropDimension, Group: "Layout"},
	{Name: "paddingBottom", Type: PropDimension, Group: "Layout"},
	{Name: "margin", Type: PropDimension, Group: "Layout"},
	{Name: "marginLeft", Type: PropDimension, Group: "Layout"},
	{Name: "marginTop", Type: PropDimension, Group: "Layout"},
	{Name: "marginRight", Type: PropDimension, Group: "Layout"},
	{Name: "marginBottom", Type: PropDimension, Group: "Layout"},
	{Name: "visibility", Type: PropEnum, Options: []string{"visible", "invisible", "gone"}, Default: "visible", Group: "Common"},
	{Name: "background", Type: PropColor, Group: "Appearance"},
}

// widgetProps maps widget tag → additional properties beyond commonProps.
var widgetProps = map[string][]PropertyDef{
	"LinearLayout": {
		{Name: "orientation", Type: PropEnum, Options: []string{"vertical", "horizontal"}, Default: "vertical", Group: "Layout"},
		{Name: "spacing", Type: PropDimension, Group: "Layout"},
		{Name: "gravity", Type: PropEnum, Options: []string{"left", "center", "right", "top", "bottom"}, Group: "Layout"},
	},
	"FrameLayout": {},
	"ScrollView": {
		{Name: "orientation", Type: PropEnum, Options: []string{"vertical", "horizontal"}, Default: "vertical", Group: "Layout"},
	},
	"GridLayout": {
		{Name: "columnCount", Type: PropInt, Default: "2", Group: "Layout"},
		{Name: "spacing", Type: PropDimension, Group: "Layout"},
	},
	"FlexLayout": {
		{Name: "orientation", Type: PropEnum, Options: []string{"horizontal", "vertical"}, Default: "horizontal", Group: "Layout"},
		{Name: "spacing", Type: PropDimension, Group: "Layout"},
		{Name: "wrap", Type: PropBool, Default: "false", Group: "Layout"},
	},
	"SplitPane": {
		{Name: "orientation", Type: PropEnum, Options: []string{"horizontal", "vertical"}, Default: "horizontal", Group: "Layout"},
		{Name: "ratio", Type: PropFloat, Default: "0.5", Group: "Layout"},
	},
	"TextView": {
		{Name: "text", Type: PropString, Group: "Text"},
		{Name: "textSize", Type: PropDimension, Default: "14dp", Group: "Text"},
		{Name: "textColor", Type: PropColor, Default: "#000000", Group: "Text"},
		{Name: "gravity", Type: PropEnum, Options: []string{"left", "center", "right"}, Group: "Text"},
	},
	"Button": {
		{Name: "text", Type: PropString, Group: "Text"},
		{Name: "textSize", Type: PropDimension, Default: "14dp", Group: "Text"},
		{Name: "textColor", Type: PropColor, Group: "Text"},
	},
	"EditText": {
		{Name: "text", Type: PropString, Group: "Text"},
		{Name: "hint", Type: PropString, Group: "Text"},
		{Name: "textSize", Type: PropDimension, Default: "14dp", Group: "Text"},
	},
	"ImageView": {
		{Name: "src", Type: PropString, Group: "Content"},
	},
	"CheckBox": {
		{Name: "text", Type: PropString, Group: "Text"},
		{Name: "checked", Type: PropBool, Default: "false", Group: "State"},
	},
	"RadioButton": {
		{Name: "text", Type: PropString, Group: "Text"},
		{Name: "checked", Type: PropBool, Default: "false", Group: "State"},
	},
	"RadioGroup": {
		{Name: "orientation", Type: PropEnum, Options: []string{"vertical", "horizontal"}, Default: "vertical", Group: "Layout"},
	},
	"Switch": {
		{Name: "checked", Type: PropBool, Default: "false", Group: "State"},
	},
	"ProgressBar": {
		{Name: "progress", Type: PropFloat, Default: "0", Group: "State"},
		{Name: "indeterminate", Type: PropBool, Default: "false", Group: "State"},
	},
	"SeekBar": {
		{Name: "progress", Type: PropFloat, Default: "0", Group: "State"},
	},
	"Spinner": {},
	"Divider": {
		{Name: "background", Type: PropColor, Default: "#CCCCCC", Group: "Appearance"},
	},
	"Toolbar": {
		{Name: "title", Type: PropString, Group: "Text"},
		{Name: "subtitle", Type: PropString, Group: "Text"},
		{Name: "textSize", Type: PropDimension, Group: "Text"},
		{Name: "textColor", Type: PropColor, Group: "Text"},
		{Name: "backgroundColor", Type: PropColor, Group: "Appearance"},
	},
	"TabLayout": {
		{Name: "backgroundColor", Type: PropColor, Group: "Appearance"},
		{Name: "textColor", Type: PropColor, Group: "Text"},
	},
	"RecyclerView": {
		{Name: "itemHeight", Type: PropDimension, Group: "Layout"},
	},
	"TreeView": {},
	"View": {},
}

// GetPropertiesForTag returns all editable properties for a widget tag
// (common + widget-specific), in display order.
func GetPropertiesForTag(tag string) []PropertyDef {
	result := make([]PropertyDef, len(commonProps))
	copy(result, commonProps)
	if extra, ok := widgetProps[tag]; ok {
		result = append(result, extra...)
	}
	return result
}

// PaletteCategory groups widgets for the palette panel.
type PaletteCategory struct {
	Name    string
	Widgets []PaletteItem
}

// PaletteItem represents a widget in the palette.
type PaletteItem struct {
	Tag         string // XML tag name
	DisplayName string // shown in palette
}

// GetPaletteCategories returns the categorized widget list for the palette.
func GetPaletteCategories() []PaletteCategory {
	return []PaletteCategory{
		{
			Name: "Container",
			Widgets: []PaletteItem{
				{"LinearLayout", "LinearLayout"},
				{"FrameLayout", "FrameLayout"},
				{"ScrollView", "ScrollView"},
				{"GridLayout", "GridLayout"},
				{"FlexLayout", "FlexLayout"},
				{"SplitPane", "SplitPane"},
			},
		},
		{
			Name: "Basic",
			Widgets: []PaletteItem{
				{"View", "View"},
				{"TextView", "TextView"},
				{"ImageView", "ImageView"},
				{"Button", "Button"},
				{"Divider", "Divider"},
			},
		},
		{
			Name: "Input",
			Widgets: []PaletteItem{
				{"EditText", "EditText"},
				{"CheckBox", "CheckBox"},
				{"RadioButton", "RadioButton"},
				{"RadioGroup", "RadioGroup"},
				{"Switch", "Switch"},
				{"SeekBar", "SeekBar"},
				{"Spinner", "Spinner"},
			},
		},
		{
			Name: "Advanced",
			Widgets: []PaletteItem{
				{"Toolbar", "Toolbar"},
				{"TabLayout", "TabLayout"},
				{"RecyclerView", "RecyclerView"},
				{"ProgressBar", "ProgressBar"},
				{"TreeView", "TreeView"},
			},
		},
	}
}
