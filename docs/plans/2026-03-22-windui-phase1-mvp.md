# Wind UI Phase 1 (MVP) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Wind UI framework skeleton from Backend to Application layer, delivering a working MVP where an XML-described window with text and buttons renders on screen and responds to clicks.

**Architecture:** Six-layer architecture (Backend → Platform → Render → Core → Widget → Application). Node-based composition model with pluggable Layout/Painter/Handler/Style. Android-style XML UI description with resource management.

**Tech Stack:** Go 1.22+, gogpu/gg (graphics), DirectWrite via C++ shim DLL (Windows text), Win32 API (windowing), encoding/xml (XML parsing)

**Spec:** `docs/specs/2026-03-22-windui-architecture-design.md`

---

## File Structure

**Import Cycle 解决方案：** Canvas 接口、Paint 结构体、TextRenderer 接口均定义在 `core/` 包中（与 Node、Painter、Handler 同包），避免 `core/` ↔ `render/` 的循环依赖。`render/gg/` 和 `platform/windows/` 仅实现这些接口。

```
windui/
 ├─ go.mod                     module windui (所有内部 import 使用 windui/core, windui/widget 等)
 ├─ go.sum
 │
 ├─ core/                      所有接口和基础类型（零外部依赖，平台无关）
 │   ├─ types.go              Rect, Size, Point, Insets, Dimension, Color utilities
 │   ├─ types_test.go
 │   ├─ node.go               Node struct, View interface, tree operations, FindViewById
 │   ├─ node_test.go
 │   ├─ canvas.go             Canvas interface（在 core/ 中定义，避免 import cycle）
 │   ├─ paint.go              Paint struct, PaintStyle（在 core/ 中定义）
 │   ├─ text.go               TextRenderer interface, TextLayout, TextLayoutResult
 │   ├─ image.go              ImageResource struct
 │   ├─ layout.go             Layout interface, MeasureSpec, MeasureMode
 │   ├─ painter.go            Painter interface（使用同包 Canvas 类型，无 import cycle）
 │   ├─ handler.go            Handler interface（使用同包 Event 类型，无 import cycle）
 │   ├─ style.go              Style struct, style resolution
 │   ├─ style_test.go
 │   ├─ event.go              Event interface, EventType, MotionEvent, KeyEvent
 │   ├─ event_dispatch.go     Event dispatch engine (3-phase)
 │   ├─ event_dispatch_test.go
 │   ├─ command.go            Command, CommandManager
 │   ├─ command_test.go
 │   ├─ focus.go              FocusManager
 │   ├─ anim.go               ValueAnimator, Interpolators
 │   └─ anim_test.go
 │
 ├─ render/gg/                 GGCanvas 实现（依赖 core/ 和 gogpu/gg）
 │   ├─ canvas_gg.go
 │   └─ canvas_gg_test.go
 │
 ├─ render/freetype/           FreeType TextRenderer 实现（依赖 core/ 和 gg/text）
 │   ├─ text_freetype.go
 │   └─ text_freetype_test.go
 │
 ├─ platform/                  平台抽象接口（依赖 core/）
 │   ├─ platform.go           Platform interface, OSType
 │   ├─ window.go             Window interface, WindowOptions
 │   └─ clipboard.go          Clipboard interface
 │
 ├─ platform/windows/          Windows 实现（依赖 core/, platform/, golang.org/x/sys）
 │   ├─ platform_windows.go   WindowsPlatform implementation
 │   ├─ window_windows.go     Win32 Window implementation
 │   ├─ dwrite_text.go        DirectWrite TextRenderer
 │   ├─ native_edit.go        Win32 EDIT control bridge (stub for Phase 1)
 │   ├─ msgloop.go            Message loop + render scheduling
 │   ├─ dpi.go                DPI detection and scaling
 │   └─ dpi_test.go           DPI 换算、消息转换的单元测试
 │
 ├─ layout/                    布局实现（依赖 core/）
 │   ├─ linear.go             LinearLayout implementation
 │   ├─ linear_test.go
 │   ├─ frame.go              FrameLayout implementation
 │   └─ frame_test.go
 │
 ├─ widget/                    内置控件（依赖 core/）
 │   ├─ base.go               BaseView (shared View interface impl)
 │   ├─ view.go               View widget (background/border)
 │   ├─ textview.go           TextView widget
 │   ├─ textview_test.go
 │   ├─ imageview.go          ImageView widget
 │   ├─ button.go             Button widget (states: normal/hover/pressed)
 │   └─ button_test.go
 │
 ├─ res/                       资源管理（依赖 core/, theme/）
 │   ├─ manager.go            ResourceManager
 │   ├─ manager_test.go
 │   ├─ inflater.go           LayoutInflater + registerBuiltinViews
 │   ├─ inflater_test.go
 │   ├─ values.go             strings.xml / colors.xml / dimens.xml parsing
 │   ├─ values_test.go
 │   └─ pack.go               .gwpack loading (stub, full impl in Phase 2)
 │
 ├─ theme/                     主题系统（依赖 core/）
 │   ├─ theme.go              Theme loading, style resolution chain
 │   └─ theme_test.go
 │
 ├─ app/                       应用层（依赖所有上层包）
 │   ├─ application.go        Application: Platform, ResourceManager, Theme, Inflater, Window list
 │   └─ render_loop.go        Render scheduling, dirty tracking, frame dispatch
 │
 └─ examples/
     └─ hello/
         ├─ main.go           Hello World demo app
         └─ res/
             ├─ layout/
             │   └─ main.xml
             └─ values/
                 ├─ strings.xml
                 ├─ colors.xml
                 └─ styles.xml
```

---

## Task 1: Project Bootstrap & Core Types

**Files:**
- Create: `go.mod`, `core/types.go`, `core/types_test.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd D:\Develop\workspace\go_dev\go_wui
go mod init windui
```

使用短模块路径 `windui`，所有内部 import 使用 `windui/core`、`windui/widget` 等。如果后续要发布到 GitHub，可以改为 `github.com/<user>/windui`。

- [ ] **Step 2: Write tests for core types**

Create `core/types_test.go` — test Rect, Size, Point, Insets, Dimension, color parsing:

```go
// core/types_test.go
package core

import "testing"

func TestRectContains(t *testing.T) {
    r := Rect{X: 10, Y: 10, Width: 100, Height: 50}
    if !r.Contains(50, 30) { t.Error("point inside should be contained") }
    if r.Contains(5, 5) { t.Error("point outside should not be contained") }
}

func TestRectIntersect(t *testing.T) {
    a := Rect{X: 0, Y: 0, Width: 100, Height: 100}
    b := Rect{X: 50, Y: 50, Width: 100, Height: 100}
    inter := a.Intersect(b)
    if inter.Width != 50 || inter.Height != 50 { t.Errorf("unexpected: %v", inter) }
}

func TestInsetsApply(t *testing.T) {
    r := Rect{X: 0, Y: 0, Width: 100, Height: 100}
    insets := Insets{Left: 10, Top: 10, Right: 10, Bottom: 10}
    inner := r.ApplyInsets(insets)
    if inner.Width != 80 || inner.Height != 80 { t.Errorf("unexpected: %v", inner) }
}

func TestParseColor(t *testing.T) {
    tests := []struct{ input string; r, g, b, a uint8 }{
        {"#FF5722", 0xFF, 0x57, 0x22, 0xFF},
        {"#80FF5722", 0xFF, 0x57, 0x22, 0x80},      // #AARRGGBB
    }
    for _, tt := range tests {
        c := ParseColor(tt.input)
        r, g, b, a := c.RGBA()
        // Compare as 8-bit values
        if uint8(r>>8) != tt.r || uint8(g>>8) != tt.g || uint8(b>>8) != tt.b || uint8(a>>8) != tt.a {
            t.Errorf("ParseColor(%q) unexpected", tt.input)
        }
    }
}

func TestDimensionParse(t *testing.T) {
    tests := []struct{ input string; unit DimensionUnit; value float64 }{
        {"200dp", DimensionDp, 200},
        {"100px", DimensionPx, 100},
        {"match_parent", DimensionMatchParent, 0},
        {"wrap_content", DimensionWrapContent, 0},
    }
    for _, tt := range tests {
        d := ParseDimension(tt.input)
        if d.Unit != tt.unit || d.Value != tt.value {
            t.Errorf("ParseDimension(%q) = %+v", tt.input, d)
        }
    }
}
```

- [ ] **Step 3: Run tests, verify they fail**

```bash
cd D:\Develop\workspace\go_dev\go_wui && go test ./core/ -v
```

Expected: compilation error (types not defined yet).

- [ ] **Step 4: Implement core types**

Create `core/types.go`:

```go
// core/types.go
package core

import (
    "fmt"
    "image/color"
    "strconv"
    "strings"
)

// Rect represents a rectangle with position and size.
type Rect struct {
    X, Y          float64
    Width, Height float64
}

func (r Rect) Contains(x, y float64) bool {
    return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

func (r Rect) Intersect(other Rect) Rect {
    x := max(r.X, other.X)
    y := max(r.Y, other.Y)
    right := min(r.X+r.Width, other.X+other.Width)
    bottom := min(r.Y+r.Height, other.Y+other.Height)
    if right <= x || bottom <= y {
        return Rect{}
    }
    return Rect{X: x, Y: y, Width: right - x, Height: bottom - y}
}

func (r Rect) ApplyInsets(insets Insets) Rect {
    return Rect{
        X:      r.X + insets.Left,
        Y:      r.Y + insets.Top,
        Width:  r.Width - insets.Left - insets.Right,
        Height: r.Height - insets.Top - insets.Bottom,
    }
}

type Size struct {
    Width, Height float64
}

type Point struct {
    X, Y float64
}

type Insets struct {
    Left, Top, Right, Bottom float64
}

// Dimension represents a size value with unit.
type DimensionUnit int

const (
    DimensionPx          DimensionUnit = iota
    DimensionDp
    DimensionMatchParent
    DimensionWrapContent
    DimensionWeight
)

type Dimension struct {
    Value float64
    Unit  DimensionUnit
}

func ParseDimension(s string) Dimension {
    switch s {
    case "match_parent":
        return Dimension{Unit: DimensionMatchParent}
    case "wrap_content":
        return Dimension{Unit: DimensionWrapContent}
    }
    if strings.HasSuffix(s, "dp") {
        v, _ := strconv.ParseFloat(s[:len(s)-2], 64)
        return Dimension{Value: v, Unit: DimensionDp}
    }
    if strings.HasSuffix(s, "px") {
        v, _ := strconv.ParseFloat(s[:len(s)-2], 64)
        return Dimension{Value: v, Unit: DimensionPx}
    }
    v, _ := strconv.ParseFloat(s, 64)
    return Dimension{Value: v, Unit: DimensionDp}
}

// ParseColor parses #RRGGBB or #AARRGGBB hex color strings.
func ParseColor(s string) color.RGBA {
    s = strings.TrimPrefix(s, "#")
    switch len(s) {
    case 6: // RRGGBB
        r, _ := strconv.ParseUint(s[0:2], 16, 8)
        g, _ := strconv.ParseUint(s[2:4], 16, 8)
        b, _ := strconv.ParseUint(s[4:6], 16, 8)
        return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}
    case 8: // AARRGGBB
        a, _ := strconv.ParseUint(s[0:2], 16, 8)
        r, _ := strconv.ParseUint(s[2:4], 16, 8)
        g, _ := strconv.ParseUint(s[4:6], 16, 8)
        b, _ := strconv.ParseUint(s[6:8], 16, 8)
        return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
    }
    return color.RGBA{}
}

func (d Dimension) String() string {
    switch d.Unit {
    case DimensionMatchParent:
        return "match_parent"
    case DimensionWrapContent:
        return "wrap_content"
    case DimensionDp:
        return fmt.Sprintf("%.0fdp", d.Value)
    case DimensionPx:
        return fmt.Sprintf("%.0fpx", d.Value)
    default:
        return fmt.Sprintf("%.0f", d.Value)
    }
}
```

- [ ] **Step 5: Run tests, verify they pass**

```bash
cd D:\Develop\workspace\go_dev\go_wui && go test ./core/ -v
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add core/types.go core/types_test.go go.mod
git commit -m "feat(core): add fundamental types — Rect, Size, Point, Insets, Dimension, color parsing"
```

---

## Task 2: Node Tree & View Interface

**Files:**
- Create: `core/node.go`, `core/node_test.go`, `core/layout.go`, `core/painter.go`, `core/handler.go`

- [ ] **Step 1: Write Node tree tests**

```go
// core/node_test.go
package core

import "testing"

func TestNodeAddChild(t *testing.T) {
    parent := NewNode("LinearLayout")
    child1 := NewNode("TextView")
    child2 := NewNode("Button")
    parent.AddChild(child1)
    parent.AddChild(child2)

    if len(parent.Children()) != 2 { t.Fatalf("expected 2 children") }
    if child1.Parent() != parent { t.Error("child1 parent mismatch") }
    if child2.Parent() != parent { t.Error("child2 parent mismatch") }
}

func TestNodeRemoveChild(t *testing.T) {
    parent := NewNode("LinearLayout")
    child := NewNode("TextView")
    parent.AddChild(child)
    parent.RemoveChild(child)

    if len(parent.Children()) != 0 { t.Error("expected 0 children") }
    if child.Parent() != nil { t.Error("parent should be nil after remove") }
}

func TestNodeFindById(t *testing.T) {
    root := NewNode("LinearLayout")
    child := NewNode("TextView")
    child.SetId("title")
    grandchild := NewNode("Button")
    grandchild.SetId("btn_ok")
    root.AddChild(child)
    child.AddChild(grandchild)

    found := root.FindNodeById("title")
    if found != child { t.Error("should find child by id") }

    found2 := root.FindNodeById("btn_ok")
    if found2 != grandchild { t.Error("should find grandchild by id") }

    notFound := root.FindNodeById("nonexistent")
    if notFound != nil { t.Error("should return nil for missing id") }
}

func TestNodeDirtyBubble(t *testing.T) {
    root := NewNode("Root")
    child := NewNode("Child")
    grandchild := NewNode("Leaf")
    root.AddChild(child)
    child.AddChild(grandchild)

    root.ClearDirty()
    child.ClearDirty()
    grandchild.ClearDirty()

    grandchild.MarkDirty()

    if !grandchild.IsDirty() { t.Error("grandchild should be dirty") }
    if !child.IsChildDirty() { t.Error("child should have childDirty") }
    if !root.IsChildDirty() { t.Error("root should have childDirty") }
}
```

- [ ] **Step 2: Run tests, verify they fail**

```bash
go test ./core/ -run TestNode -v
```

- [ ] **Step 3: Implement Node, interfaces, and View**

Create `core/layout.go`:

```go
package core

// MeasureMode determines how a dimension constraint should be interpreted.
type MeasureMode int

const (
    MeasureModeExact   MeasureMode = iota // Exact size (dp value or match_parent resolved by parent)
    MeasureModeAtMost                      // Up to this size (wrap_content)
    MeasureModeUnbound                     // No constraint (ScrollView children)
)

type MeasureSpec struct {
    Mode MeasureMode
    Size float64
}

// Layout defines a strategy for measuring and arranging child nodes.
type Layout interface {
    Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
    Arrange(node *Node, bounds Rect)
}
```

Create `core/painter.go`:

```go
package core

// Painter handles measurement and rendering of node content.
// Canvas is defined in the same package (core/canvas.go), so no import cycle.
type Painter interface {
    Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
    Paint(node *Node, canvas Canvas)
}
```

Create `core/handler.go`:

```go
package core

// Handler processes events on a node.
// Event is defined in the same package (core/event.go), so no import cycle.
type Handler interface {
    OnDispatchEvent(node *Node, event Event) bool
    OnInterceptEvent(node *Node, event Event) bool
    OnEvent(node *Node, event Event) bool
}

// DefaultHandler provides no-op defaults for Handler.
type DefaultHandler struct{}

func (h *DefaultHandler) OnDispatchEvent(node *Node, event Event) bool  { return false }
func (h *DefaultHandler) OnInterceptEvent(node *Node, event Event) bool { return false }
func (h *DefaultHandler) OnEvent(node *Node, event Event) bool          { return false }
```

Create `core/node.go`:

```go
package core

// Visibility constants
type Visibility int

const (
    Visible   Visibility = iota
    Invisible            // Takes up space but not drawn
    Gone                 // No space, not drawn
)

// View is the public interface for all UI elements.
type View interface {
    Node() *Node
    SetId(id string)
    GetId() string
    SetVisibility(v Visibility)
    GetVisibility() Visibility
    SetEnabled(enabled bool)
    IsEnabled() bool
}

// Node is the internal tree element backing every View.
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

    // measured size (filled during measure phase)
    measuredSize Size
    // View association — the widget wrapping this node (set by widget constructors)
    view View

    // custom data store for widget-specific state
    data map[string]interface{}
}

func NewNode(tag string) *Node {
    return &Node{
        tag:        tag,
        visibility: Visible,
        enabled:    true,
        dirty:      true,
        data:       make(map[string]interface{}),
    }
}

// Tree operations

func (n *Node) Parent() *Node        { return n.parent }
func (n *Node) Children() []*Node    { return n.children }
func (n *Node) Tag() string          { return n.tag }

func (n *Node) AddChild(child *Node) {
    if child.parent != nil {
        child.parent.RemoveChild(child)
    }
    child.parent = n
    n.children = append(n.children, child)
    n.MarkDirty()
}

func (n *Node) RemoveChild(child *Node) {
    for i, c := range n.children {
        if c == child {
            n.children = append(n.children[:i], n.children[i+1:]...)
            child.parent = nil
            n.MarkDirty()
            return
        }
    }
}

func (n *Node) SetId(id string)   { n.id = id }
func (n *Node) GetId() string     { return n.id }
func (n *Node) Id() string        { return n.id }

func (n *Node) SetVisibility(v Visibility) { n.visibility = v; n.MarkDirty() }
func (n *Node) GetVisibility() Visibility  { return n.visibility }

func (n *Node) SetEnabled(enabled bool) { n.enabled = enabled }
func (n *Node) IsEnabled() bool         { return n.enabled }

// Component accessors

func (n *Node) SetLayout(l Layout)   { n.layout = l; n.MarkDirty() }
func (n *Node) GetLayout() Layout    { return n.layout }
func (n *Node) SetPainter(p Painter) { n.painter = p; n.MarkDirty() }
func (n *Node) GetPainter() Painter  { return n.painter }
func (n *Node) SetHandler(h Handler) { n.handler = h }
func (n *Node) GetHandler() Handler  { return n.handler }
func (n *Node) SetStyle(s *Style)    { n.style = s; n.MarkDirty() }
func (n *Node) GetStyle() *Style     { return n.style }

// Geometry

func (n *Node) Bounds() Rect          { return n.bounds }
func (n *Node) SetBounds(r Rect)      { n.bounds = r }
func (n *Node) Padding() Insets       { return n.padding }
func (n *Node) SetPadding(i Insets)   { n.padding = i }
func (n *Node) Margin() Insets        { return n.margin }
func (n *Node) SetMargin(i Insets)    { n.margin = i }
func (n *Node) MeasuredSize() Size    { return n.measuredSize }
func (n *Node) SetMeasuredSize(s Size) { n.measuredSize = s }

// Data store (widget-specific state like text content, image src, etc.)

func (n *Node) SetData(key string, value interface{}) { n.data[key] = value }
func (n *Node) GetData(key string) interface{}         { return n.data[key] }
func (n *Node) GetDataString(key string) string {
    if v, ok := n.data[key].(string); ok { return v }
    return ""
}

// Dirty tracking

func (n *Node) MarkDirty() {
    n.dirty = true
    // Bubble childDirty up to ancestors
    for p := n.parent; p != nil; p = p.parent {
        if p.childDirty { break } // already marked
        p.childDirty = true
    }
}

func (n *Node) IsDirty() bool      { return n.dirty }
func (n *Node) IsChildDirty() bool { return n.childDirty }

func (n *Node) ClearDirty() {
    n.dirty = false
    n.childDirty = false
}

// View association — set by widget constructors, enables FindViewById
func (n *Node) SetView(v View) { n.view = v }
func (n *Node) GetView() View  { return n.view }

// FindNodeById searches the subtree for a node with the given id.
func (n *Node) FindNodeById(id string) *Node {
    if n.id == id { return n }
    for _, child := range n.children {
        if found := child.FindNodeById(id); found != nil {
            return found
        }
    }
    return nil
}

// FindViewById searches the subtree and returns the View wrapping the found node.
// Returns nil if not found or if the node has no associated View.
func (n *Node) FindViewById(id string) View {
    node := n.FindNodeById(id)
    if node == nil { return nil }
    return node.view
}
```

- [ ] **Step 4: Run tests, verify they pass**

```bash
go test ./core/ -run TestNode -v
```

- [ ] **Step 5: Commit**

```bash
git add core/
git commit -m "feat(core): add Node tree, View interface, Layout/Painter/Handler interfaces"
```

---

## Task 3: Style System

**Files:**
- Create: `core/style.go`, `core/style_test.go`

- [ ] **Step 1: Write style resolution tests**

Test style merging (parent → child override), default values, and attribute access:

```go
// core/style_test.go
package core

import (
    "image/color"
    "testing"
)

func TestStyleMerge(t *testing.T) {
    parent := &Style{
        BackgroundColor: ParseColor("#FFFFFF"),
        TextColor:       ParseColor("#000000"),
        FontSize:        14,
    }
    child := &Style{
        TextColor: ParseColor("#FF0000"), // override
        FontSize:  18,                    // override
    }
    merged := MergeStyles(parent, child)
    if merged.BackgroundColor != parent.BackgroundColor {
        t.Error("should inherit parent background")
    }
    if merged.TextColor != child.TextColor {
        t.Error("should use child text color")
    }
    if merged.FontSize != 18 { t.Error("should use child font size") }
}

func TestStyleIsZero(t *testing.T) {
    s := &Style{}
    if s.FontSize != 0 { t.Error("default should be zero") }
    if s.BackgroundColor != (color.RGBA{}) { t.Error("default color should be zero") }
}
```

- [ ] **Step 2: Implement Style**

```go
// core/style.go
package core

import "image/color"

// Style holds visual properties for a node.
type Style struct {
    BackgroundColor color.RGBA
    BorderColor     color.RGBA
    BorderWidth     float64
    CornerRadius    float64
    FontSize        float64
    FontFamily      string
    FontWeight      int
    TextColor       color.RGBA
    Opacity         float64

    // Width/Height dimensions from XML
    Width  Dimension
    Height Dimension
    Weight float64 // layout_weight for LinearLayout

    // Gravity
    Gravity     Gravity
    TextGravity Gravity
}

type Gravity int

const (
    GravityStart        Gravity = 0
    GravityCenter       Gravity = 1
    GravityEnd          Gravity = 2
    GravityCenterVertical Gravity = 4
    GravityCenterHorizontal Gravity = 8
)

// MergeStyles creates a new style with child values overriding parent.
// Zero values in child are treated as "not set" and inherited from parent.
func MergeStyles(parent, child *Style) *Style {
    if parent == nil { return child }
    if child == nil  { return parent }
    merged := *parent
    if child.BackgroundColor != (color.RGBA{}) { merged.BackgroundColor = child.BackgroundColor }
    if child.BorderColor != (color.RGBA{})     { merged.BorderColor = child.BorderColor }
    if child.BorderWidth != 0     { merged.BorderWidth = child.BorderWidth }
    if child.CornerRadius != 0    { merged.CornerRadius = child.CornerRadius }
    if child.FontSize != 0        { merged.FontSize = child.FontSize }
    if child.FontFamily != ""     { merged.FontFamily = child.FontFamily }
    if child.FontWeight != 0      { merged.FontWeight = child.FontWeight }
    if child.TextColor != (color.RGBA{}) { merged.TextColor = child.TextColor }
    if child.Opacity != 0         { merged.Opacity = child.Opacity }
    return &merged
}
```

- [ ] **Step 3: Run tests, verify they pass**

```bash
go test ./core/ -run TestStyle -v
```

- [ ] **Step 4: Commit**

```bash
git add core/style.go core/style_test.go
git commit -m "feat(core): add Style struct with merge/inheritance support"
```

---

## Task 4: Event System

**Files:**
- Create: `core/event.go`, `core/event_dispatch.go`, `core/event_dispatch_test.go`, `core/command.go`, `core/command_test.go`, `core/focus.go`

- [ ] **Step 1: Write event dispatch tests**

```go
// core/event_dispatch_test.go
package core

import "testing"

func TestEventDispatch_BubbleUp(t *testing.T) {
    root := NewNode("Root")
    child := NewNode("Child")
    leaf := NewNode("Leaf")
    root.AddChild(child)
    child.AddChild(leaf)

    var received []string
    child.SetHandler(&testHandler{onEvent: func(n *Node, e Event) bool {
        received = append(received, "child")
        return false // not consumed, bubble up
    }})
    root.SetHandler(&testHandler{onEvent: func(n *Node, e Event) bool {
        received = append(received, "root")
        return true
    }})

    event := &MotionEvent{action: ActionDown, x: 5, y: 5}
    leaf.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})
    child.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})
    root.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})

    DispatchEvent(root, event, Point{X: 5, Y: 5})

    if len(received) != 2 || received[0] != "child" || received[1] != "root" {
        t.Errorf("expected bubble child→root, got %v", received)
    }
}

func TestEventDispatch_Consumed(t *testing.T) {
    root := NewNode("Root")
    child := NewNode("Child")
    root.AddChild(child)

    rootCalled := false
    child.SetHandler(&testHandler{onEvent: func(n *Node, e Event) bool {
        return true // consumed
    }})
    root.SetHandler(&testHandler{onEvent: func(n *Node, e Event) bool {
        rootCalled = true
        return false
    }})

    event := &MotionEvent{action: ActionDown, x: 5, y: 5}
    child.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})
    root.SetBounds(Rect{X: 0, Y: 0, Width: 100, Height: 100})

    DispatchEvent(root, event, Point{X: 5, Y: 5})

    if rootCalled { t.Error("root should not be called when child consumed event") }
}

// testHandler is a test helper
type testHandler struct {
    DefaultHandler
    onEvent func(*Node, Event) bool
}

func (h *testHandler) OnEvent(node *Node, event Event) bool {
    if h.onEvent != nil {
        return h.onEvent(node, event)
    }
    return false
}
```

- [ ] **Step 2: Implement Event types and dispatch**

Create `core/event.go` with Event interface, MotionEvent, KeyEvent, EventType.
Create `core/event_dispatch.go` with the 3-phase dispatch (dispatch → intercept → handle with bubble).
Create `core/command.go` with Command struct and CommandManager.
Create `core/focus.go` with FocusManager (basic Tab navigation).

Implementation details: follow the spec Section 6. Event dispatch does hit-testing on children (check bounds.Contains), walks down to find deepest hit node, then bubbles up calling OnEvent until consumed.

- [ ] **Step 3: Run tests, verify they pass**

```bash
go test ./core/ -run TestEvent -v
go test ./core/ -run TestCommand -v
```

- [ ] **Step 4: Commit**

```bash
git add core/event.go core/event_dispatch.go core/event_dispatch_test.go core/command.go core/command_test.go core/focus.go
git commit -m "feat(core): add event system with 3-phase dispatch, Command manager, and focus navigation"
```

---

## Task 5: Canvas, Paint, TextRenderer 接口（在 core/ 中定义）

**Files:**
- Create: `core/canvas.go`, `core/paint.go`, `core/text.go`, `core/image.go`

**关键设计决策：** Canvas、Paint、TextRenderer 接口全部定义在 `core/` 包中，与 Painter/Handler 同包。这从根本上消除了 `core/` ↔ `render/` 的循环依赖。`render/gg/` 和 `platform/windows/` 仅作为这些接口的实现包。

- [ ] **Step 1: Define Canvas interface**

`core/canvas.go`: Canvas interface — DrawRect, DrawRoundRect, DrawCircle, DrawLine, DrawImage, DrawText, MeasureText, Save, Restore, Translate, ClipRect, Target. 所有参数使用同包类型（Rect, Paint, ImageResource）。

- [ ] **Step 2: Define Paint struct**

`core/paint.go`: Paint struct with DrawStyle, Color, StrokeWidth, FontSize, FontFamily, FontWeight, AntiAlias. PaintStyle enum (PaintFill, PaintStroke, PaintFillAndStroke).

- [ ] **Step 3: Define TextRenderer interface**

`core/text.go`: TextRenderer interface (SetFont, MeasureText, DrawText, CreateTextLayout, Close). TextLayoutResult, TextLine, TextLayout structs.

- [ ] **Step 4: Define ImageResource**

`core/image.go`: ImageResource struct (Image *image.RGBA, Width, Height int, Name string).

- [ ] **Step 5: Commit**

```bash
git add core/canvas.go core/paint.go core/text.go core/image.go
git commit -m "feat(core): add Canvas, Paint, TextRenderer, ImageResource interfaces and types

All rendering interfaces live in core/ to avoid import cycles with Painter/Handler."
```

---

## Task 6: Platform Abstractions

**Files:**
- Create: `platform/platform.go`, `platform/window.go`, `platform/clipboard.go`

- [ ] **Step 1: Define Platform, Window, Clipboard interfaces**

Pure interface definitions per spec Section 10. Include WindowOptions, Screen, ThemeMode, OSType, NativeEditText, DialogResult types.

- [ ] **Step 2: Commit**

```bash
git add platform/
git commit -m "feat(platform): add Platform, Window, Clipboard, NativeEditText interface definitions"
```

---

## Task 7: GG Canvas Implementation

**Files:**
- Create: `render/gg/canvas_gg.go`, `render/gg/canvas_gg_test.go`
- Modify: `go.mod` (add gg dependency)

- [ ] **Step 1: Add gg dependency**

```bash
cd D:\Develop\workspace\go_dev\go_wui && go get github.com/gogpu/gg@v0.35.3
```

- [ ] **Step 2: Write GGCanvas tests**

Test basic drawing operations produce non-zero pixels:

```go
// render/gg/canvas_gg_test.go
func TestGGCanvasDrawRect(t *testing.T) {
    c := NewGGCanvas(100, 100, nil) // nil TextRenderer for now
    paint := &core.Paint{Color: color.RGBA{R: 255, A: 255}, DrawStyle: core.PaintFill}
    c.DrawRect(core.Rect{X: 10, Y: 10, Width: 50, Height: 50}, paint)
    img := c.Target()
    // Check pixel at (25, 25) is red
    r, _, _, a := img.At(25, 25).RGBA()
    if r == 0 || a == 0 { t.Error("expected non-zero red pixel") }
}

func TestGGCanvasSaveRestore(t *testing.T) {
    c := NewGGCanvas(100, 100, nil)
    c.Save()
    c.Translate(50, 50)
    c.Restore()
    // After restore, translate should be undone — drawing at (0,0) should work at original position
    paint := &core.Paint{Color: color.RGBA{G: 255, A: 255}, DrawStyle: core.PaintFill}
    c.DrawRect(core.Rect{X: 0, Y: 0, Width: 10, Height: 10}, paint)
    img := c.Target()
    _, g, _, _ := img.At(5, 5).RGBA()
    if g == 0 { t.Error("expected green pixel at original position") }
}
```

- [ ] **Step 3: Implement GGCanvas**

Wrap `gg.Context` — implement all Canvas methods. Use state stack ([]gg.Matrix or manual save/restore of gg context) for Save/Restore. DrawText/MeasureText delegate to the injected TextRenderer.

- [ ] **Step 4: Run tests, verify they pass**

```bash
go test ./render/gg/ -v
```

- [ ] **Step 5: Commit**

```bash
git add render/gg/ go.mod go.sum
git commit -m "feat(render): implement GGCanvas backed by gogpu/gg"
```

---

## Task 8: FreeType TextRenderer

**Files:**
- Create: `render/freetype/text_freetype.go`, `render/freetype/text_freetype_test.go`

- [ ] **Step 1: Write FreeType TextRenderer tests**

```go
// render/freetype/text_freetype_test.go
package freetype

import (
    "testing"
    "windui/core"
)

func TestFreeTypeMeasureText_NonZero(t *testing.T) {
    tr := NewFreeTypeTextRenderer()
    defer tr.Close()
    // Use a system font that's likely available
    tr.SetFont("sans-serif", 400, 16)
    size := tr.MeasureText("Hello")
    if size.Width <= 0 || size.Height <= 0 {
        t.Errorf("MeasureText should return positive size, got %v", size)
    }
}

func TestFreeTypeMeasureText_Empty(t *testing.T) {
    tr := NewFreeTypeTextRenderer()
    defer tr.Close()
    tr.SetFont("sans-serif", 400, 16)
    size := tr.MeasureText("")
    if size.Width != 0 { t.Error("empty string should have zero width") }
}

func TestFreeTypeCreateTextLayout_LineBreak(t *testing.T) {
    tr := NewFreeTypeTextRenderer()
    defer tr.Close()
    tr.SetFont("sans-serif", 400, 16)
    paint := &core.Paint{FontSize: 16, FontFamily: "sans-serif"}
    // Use a narrow maxWidth to force line breaking
    result := tr.CreateTextLayout("Hello World this is a long text", paint, 50)
    if len(result.Lines) < 2 {
        t.Errorf("expected multiple lines for narrow width, got %d", len(result.Lines))
    }
}
```

- [ ] **Step 2: Run tests, verify they fail**

```bash
go test ./render/freetype/ -v
```

- [ ] **Step 3: Implement FreeTypeTextRenderer**

Wrap `gogpu/gg/text` for font loading, text measurement, and drawing. Implement:
- `SetFont`: load font via `gg/text.FontSource`, with LRU font cache
- `MeasureText`: use `gg.Context.MeasureString`
- `DrawText`: draw onto the Canvas target image
- `CreateTextLayout`: implement basic greedy line-breaking algorithm (accumulate word widths, break when exceeding maxWidth)

Font fallback: iterate through fallback font list, check `face.HasGlyph(r)` per rune, segment text by font — same approach as WindInput's `segmentByFont` but implemented independently.

- [ ] **Step 4: Run tests, verify they pass**

```bash
go test ./render/freetype/ -v
```

- [ ] **Step 5: Commit**

```bash
git add render/freetype/
git commit -m "feat(render): implement FreeType TextRenderer with line-breaking and font fallback"
```

---

## Task 9: Windows Platform — Basic Window

**Files:**
- Create: `platform/windows/platform_windows.go`, `platform/windows/window_windows.go`, `platform/windows/msgloop.go`, `platform/windows/dpi.go`
- Modify: `go.mod` (add golang.org/x/sys)

- [ ] **Step 1: Add sys dependency**

```bash
go get golang.org/x/sys@latest
```

- [ ] **Step 2: Implement WindowsPlatform**

`platform_windows.go`:
- `CreateWindow`: call Win32 `CreateWindowEx` — support standard (WS_OVERLAPPEDWINDOW), frameless (WS_POPUP), transparent (WS_EX_LAYERED)
- `RunMainLoop`: GetMessage/DispatchMessage loop
- `PostToMainThread`: PostMessage with WM_APP custom message
- `Quit`: PostQuitMessage
- `GetScreens`: EnumDisplayMonitors
- `GetSystemTheme`: read registry `HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize\AppsUseLightTheme`

`window_windows.go`:
- Win32 Window wrapper implementing platform.Window
- WndProc callback: handle WM_PAINT, WM_SIZE, WM_CLOSE, WM_DESTROY, WM_MOUSEMOVE, WM_LBUTTONDOWN/UP, WM_KEYDOWN/UP, WM_MOUSEWHEEL, WM_DPICHANGED
- Convert Win32 messages to Wind UI Events (MotionEvent, KeyEvent)
- Transparent mode: RGBA → BGRA + UpdateLayeredWindow

`msgloop.go`:
- Render scheduling: Invalidate → PostMessage(WM_APP_PAINT) → coalesced per frame
- Dirty check → measure → arrange → paint → present

`dpi.go`:
- GetDpiForWindow / GetDpiForMonitor for DPI-aware scaling
- dp → px conversion utility

- [ ] **Step 3: Manual test — create a blank window**

Create a temporary `cmd/test_window/main.go`:

```go
package main

import (
    "runtime"
    "windui/platform/windows"
    "windui/platform"
)

func main() {
    runtime.LockOSThread()
    p := windows.NewPlatform()
    w, _ := p.CreateWindow(platform.WindowOptions{
        Title:  "Wind UI Test",
        Width:  400,
        Height: 300,
    })
    w.Show()
    p.RunMainLoop()
}
```

```bash
go run ./cmd/test_window/
```

Expected: a 400x300 window appears with title "Wind UI Test", can be moved/resized/closed.

- [ ] **Step 4: Write DPI and event conversion unit tests**

```go
// platform/windows/dpi_test.go
package windows

import "testing"

func TestDpToPx(t *testing.T) {
    tests := []struct{ dp, dpi, expected float64 }{
        {16, 96, 16},    // 100% scaling
        {16, 144, 24},   // 150% scaling
        {16, 192, 32},   // 200% scaling
    }
    for _, tt := range tests {
        px := DpToPx(tt.dp, tt.dpi)
        if px != tt.expected {
            t.Errorf("DpToPx(%v, %v) = %v, want %v", tt.dp, tt.dpi, px, tt.expected)
        }
    }
}

func TestPxToDp(t *testing.T) {
    px := PxToDp(24, 144) // 150% scaling
    if px != 16 { t.Errorf("PxToDp(24, 144) = %v, want 16", px) }
}
```

```bash
go test ./platform/windows/ -run TestDp -v
```

- [ ] **Step 5: Commit**

```bash
git add platform/windows/ go.mod go.sum
git commit -m "feat(platform/windows): Win32 window creation, message loop, DPI, render scheduling"
```

---

## Task 10: DirectWrite TextRenderer

**Files:**
- Create: `platform/windows/dwrite_text.go`

- [ ] **Step 1: Implement DWriteTextRenderer**

Port the approach from WindInput's `dwrite_text.go`:
- Load `wind_dwrite.dll` (or a new `windui_dwrite.dll`) via syscall.NewLazyDLL
- Implement TextRenderer interface: SetFont, MeasureText, DrawText, CreateTextLayout
- CreateTextLayout: use IDWriteTextLayout for native line-breaking and measurement
- Fallback: if DLL not found, return error so platform can fall back to FreeType

Note: The C++ shim DLL needs to be built separately. For Phase 1, if the DLL is not available, FreeType is used automatically.

- [ ] **Step 2: Commit**

```bash
git add platform/windows/dwrite_text.go
git commit -m "feat(platform/windows): DirectWrite TextRenderer via C++ shim DLL"
```

---

## Task 11: Layout Implementations

**Files:**
- Create: `layout/linear.go`, `layout/linear_test.go`, `layout/frame.go`, `layout/frame_test.go`

- [ ] **Step 1: Write LinearLayout tests**

```go
// layout/linear_test.go
package layout

import (
    "testing"
    "windui/core"
)

func TestLinearVertical_WrapContent(t *testing.T) {
    parent := core.NewNode("LinearLayout")
    ll := &LinearLayout{Orientation: Vertical, Spacing: 10}
    parent.SetLayout(ll)

    child1 := newTestLeaf(100, 30)  // helper: creates node with mock painter returning fixed size
    child2 := newTestLeaf(80, 40)
    parent.AddChild(child1)
    parent.AddChild(child2)

    size := ll.Measure(parent,
        core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 200},
        core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 500},
    )

    // Width = max(100, 80) = 100, Height = 30 + 10 + 40 = 80
    if size.Width != 100 { t.Errorf("width: got %v, want 100", size.Width) }
    if size.Height != 80 { t.Errorf("height: got %v, want 80", size.Height) }
}

func TestLinearVertical_Weight(t *testing.T) {
    parent := core.NewNode("LinearLayout")
    ll := &LinearLayout{Orientation: Vertical}
    parent.SetLayout(ll)

    child1 := newTestLeaf(100, 0)
    child1.GetStyle().Weight = 1
    child1.GetStyle().Height = core.Dimension{Unit: core.DimensionWeight}
    child2 := newTestLeaf(100, 0)
    child2.GetStyle().Weight = 2
    child2.GetStyle().Height = core.Dimension{Unit: core.DimensionWeight}
    parent.AddChild(child1)
    parent.AddChild(child2)

    ll.Measure(parent,
        core.MeasureSpec{Mode: core.MeasureModeExact, Size: 100},
        core.MeasureSpec{Mode: core.MeasureModeExact, Size: 300},
    )
    ll.Arrange(parent, core.Rect{Width: 100, Height: 300})

    // child1 gets 1/3 = 100, child2 gets 2/3 = 200
    if child1.Bounds().Height != 100 { t.Errorf("child1 height: got %v, want 100", child1.Bounds().Height) }
    if child2.Bounds().Height != 200 { t.Errorf("child2 height: got %v, want 200", child2.Bounds().Height) }
}

func TestLinearHorizontal_Basic(t *testing.T) {
    parent := core.NewNode("LinearLayout")
    ll := &LinearLayout{Orientation: Horizontal, Spacing: 5}
    parent.SetLayout(ll)

    child1 := newTestLeaf(60, 30)
    child2 := newTestLeaf(40, 20)
    parent.AddChild(child1)
    parent.AddChild(child2)

    size := ll.Measure(parent,
        core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 300},
        core.MeasureSpec{Mode: core.MeasureModeAtMost, Size: 100},
    )

    // Width = 60 + 5 + 40 = 105, Height = max(30, 20) = 30
    if size.Width != 105 { t.Errorf("width: got %v, want 105", size.Width) }
    if size.Height != 30 { t.Errorf("height: got %v, want 30", size.Height) }
}

// newTestLeaf creates a node with a mock painter that returns a fixed size.
func newTestLeaf(w, h float64) *core.Node {
    n := core.NewNode("TestLeaf")
    n.SetPainter(&fixedSizePainter{width: w, height: h})
    n.SetStyle(&core.Style{})
    return n
}

type fixedSizePainter struct {
    width, height float64
}

func (p *fixedSizePainter) Measure(node *core.Node, ws, hs core.MeasureSpec) core.Size {
    return core.Size{Width: p.width, Height: p.height}
}

func (p *fixedSizePainter) Paint(node *core.Node, canvas interface{}) {}
```

- [ ] **Step 2: Run tests, verify they fail**

```bash
go test ./layout/ -v
```

- [ ] **Step 3: Implement LinearLayout**

`layout/linear.go`:
- Measure: iterate children, sum sizes along orientation axis, max across cross axis, handle weight distribution
- Arrange: position children sequentially with spacing, apply gravity for cross-axis alignment

- [ ] **Step 4: Write and run FrameLayout tests**

`layout/frame_test.go`: test children overlapping, gravity positioning (center, end).

- [ ] **Step 5: Implement FrameLayout**

`layout/frame.go`:
- Measure: max of all children widths/heights
- Arrange: position each child according to its gravity within the frame bounds

- [ ] **Step 6: Run all layout tests**

```bash
go test ./layout/ -v
```

- [ ] **Step 7: Commit**

```bash
git add layout/
git commit -m "feat(layout): implement LinearLayout (vertical/horizontal, weight, gravity) and FrameLayout"
```

---

## Task 12: Widget Implementations

**Files:**
- Create: `widget/base.go`, `widget/view.go`, `widget/textview.go`, `widget/imageview.go`, `widget/button.go`, `widget/textview_test.go`, `widget/button_test.go`

- [ ] **Step 1: Implement BaseView**

`widget/base.go`: shared View interface implementation wrapping *Node. Provides SetId, GetId, SetVisibility, IsEnabled, Node() accessors.

- [ ] **Step 2: Implement View widget**

`widget/view.go`: BackgroundPainter that draws background color, border, corner radius from style. Measure returns (0, 0) for pure container use.

- [ ] **Step 3: Implement TextView**

`widget/textview.go`:
- TextViewPainter: Measure uses TextRenderer.MeasureText/CreateTextLayout for size calculation. Paint draws text with color, alignment.
- SetText/GetText via Node data store.
- XML attrs: text, textSize, textColor, maxLines, ellipsize, gravity, lineSpacingMultiplier.

- [ ] **Step 4: Write TextView tests**

```go
// widget/textview_test.go
func TestTextViewSetText(t *testing.T) {
    tv := NewTextView("")
    tv.SetText("Hello")
    if tv.GetText() != "Hello" { t.Error("text mismatch") }
}
```

- [ ] **Step 5: Implement ImageView**

`widget/imageview.go`:
- ImageViewPainter: Measure returns image dimensions (scaled by scaleType). Paint draws image via Canvas.DrawImage.
- XML attrs: src, scaleType (fitCenter, centerCrop, fitXY).

- [ ] **Step 6: Implement Button**

`widget/button.go`:
- ButtonPainter: draws background with state-dependent colors (normal/hover/pressed). Draws text centered.
- ButtonHandler: tracks pointer state (hover, pressed), fires OnClickListener. Uses GestureDetector internally.
- SetText/GetText, SetOnClickListener.

- [ ] **Step 7: Write Button tests**

```go
// widget/button_test.go
func TestButtonClickListener(t *testing.T) {
    btn := NewButton("OK", nil)
    clicked := false
    btn.SetOnClickListener(func(v core.View) { clicked = true })

    // Simulate click: ActionDown then ActionUp within bounds
    btn.Node().SetBounds(core.Rect{X: 0, Y: 0, Width: 100, Height: 48})
    handler := btn.Node().GetHandler()

    downEvent := &core.MotionEvent{Action: core.ActionDown, X: 50, Y: 24, Source: core.PointerMouse}
    handler.OnEvent(btn.Node(), downEvent)

    upEvent := &core.MotionEvent{Action: core.ActionUp, X: 50, Y: 24, Source: core.PointerMouse}
    handler.OnEvent(btn.Node(), upEvent)

    if !clicked { t.Error("click listener not called after down+up") }
}
```

- [ ] **Step 8: Run all widget tests**

```bash
go test ./widget/ -v
```

- [ ] **Step 9: Commit**

```bash
git add widget/
git commit -m "feat(widget): implement View, TextView, ImageView, Button widgets"
```

---

## Task 13: Resource Manager & Value Parsing

**Files:**
- Create: `res/manager.go`, `res/manager_test.go`, `res/values.go`, `res/values_test.go`

- [ ] **Step 1: Write resource value parsing tests**

```go
// res/values_test.go
func TestParseStringsXML(t *testing.T) {
    xml := `<resources>
        <string name="app_title">My App</string>
        <string name="greeting">Hello, %1$s!</string>
    </resources>`
    strings, err := ParseStringsXML([]byte(xml))
    if err != nil { t.Fatal(err) }
    if strings["app_title"] != "My App" { t.Error("app_title mismatch") }
    if strings["greeting"] != "Hello, %1$s!" { t.Error("greeting mismatch") }
}

func TestParseColorsXML(t *testing.T) {
    xml := `<resources>
        <color name="primary">#1976D2</color>
        <color name="background">#FFFFFF</color>
    </resources>`
    colors, err := ParseColorsXML([]byte(xml))
    if err != nil { t.Fatal(err) }
    if colors["primary"] != "#1976D2" { t.Error("primary mismatch") }
}
```

- [ ] **Step 2: Implement value parsing**

`res/values.go`: Parse strings.xml, colors.xml, dimens.xml using encoding/xml. Handle string-array and plurals. Android-compatible format.

- [ ] **Step 3: Write ResourceManager tests**

```go
// res/manager_test.go
func TestResourceManagerGetString(t *testing.T) {
    // Use testing/fstest.MapFS to simulate embedded resources
    fs := fstest.MapFS{
        "values/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="ok">OK</string></resources>`)},
        "values-zh/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="ok">确定</string></resources>`)},
    }
    rm := NewResourceManager(fs)
    rm.SetLocale("en")
    if rm.GetString("ok") != "OK" { t.Error("expected English") }
    rm.SetLocale("zh")
    if rm.GetString("ok") != "确定" { t.Error("expected Chinese") }
}
```

- [ ] **Step 4: Implement ResourceManager**

`res/manager.go`: Three-layer loading (overrideDir → packs → embedded fs.FS). Locale fallback chain (zh-rCN → zh → default). Caching. GetString, GetColor, GetDimension, GetDrawable, GetStringFormatted.

- [ ] **Step 5: Run tests**

```bash
go test ./res/ -v
```

- [ ] **Step 6: Commit**

```bash
git add res/
git commit -m "feat(res): implement ResourceManager with value parsing, locale fallback, and three-layer loading"
```

---

## Task 14: LayoutInflater & XML Loading

**Files:**
- Create: `res/inflater.go`, `res/inflater_test.go`

- [ ] **Step 1: Write LayoutInflater tests**

```go
// res/inflater_test.go
func TestInflateLinearLayout(t *testing.T) {
    xml := `<LinearLayout width="match_parent" height="match_parent" orientation="vertical" padding="16dp">
        <TextView id="title" width="match_parent" height="wrap_content" text="Hello" textSize="18dp" />
        <Button id="btn" width="match_parent" height="48dp" text="Click" />
    </LinearLayout>`

    inflater := NewLayoutInflater(nil) // nil ResourceManager for inline values
    registerBuiltinViews(inflater)
    root, err := inflater.InflateFromString(xml)
    if err != nil { t.Fatal(err) }
    if root.Tag() != "LinearLayout" { t.Error("root should be LinearLayout") }
    if len(root.Children()) != 2 { t.Fatalf("expected 2 children, got %d", len(root.Children())) }
    if root.Children()[0].Id() != "title" { t.Error("first child should be 'title'") }
    if root.Children()[1].Id() != "btn" { t.Error("second child should be 'btn'") }
}

func TestInflateResourceRef(t *testing.T) {
    fs := fstest.MapFS{
        "values/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="hello">Hello World</string></resources>`)},
    }
    rm := NewResourceManager(fs)
    inflater := NewLayoutInflater(rm)
    registerBuiltinViews(inflater)

    xml := `<TextView width="wrap_content" height="wrap_content" text="@string/hello" />`
    node, _ := inflater.InflateFromString(xml)
    if node.GetDataString("text") != "Hello World" { t.Error("resource ref not resolved") }
}
```

- [ ] **Step 2: Implement LayoutInflater**

`res/inflater.go`:
- Parse XML using encoding/xml Decoder
- ViewFactory registry (map[string]ViewFactory)
- `registerBuiltinViews`: register LinearLayout, FrameLayout, TextView, ImageView, Button, View
- Attribute parsing: resolve @string/, @color/, @dimen/ references via ResourceManager
- Build Node tree recursively
- Handle `<include layout="@layout/name" />`

- [ ] **Step 3: Run tests**

```bash
go test ./res/ -run TestInflate -v
```

- [ ] **Step 4: Commit**

```bash
git add res/inflater.go res/inflater_test.go
git commit -m "feat(res): implement LayoutInflater with XML parsing, resource ref resolution, and built-in view registry"
```

---

## Task 15: Theme System

**Files:**
- Create: `theme/theme.go`, `theme/theme_test.go`

- [ ] **Step 1: Write theme tests**

```go
// theme/theme_test.go
func TestThemeResolveAttr(t *testing.T) {
    themeXML := `<resources><style name="Theme.Wind UI.Light">
        <item name="colorPrimary">#1976D2</item>
        <item name="textColorPrimary">#212121</item>
    </style></resources>`
    theme, _ := LoadThemeFromXML([]byte(themeXML), "Theme.Wind UI.Light")
    c := theme.ResolveAttr("colorPrimary")
    if c != "#1976D2" { t.Errorf("got %s", c) }
}

func TestStyleInheritance(t *testing.T) {
    stylesXML := `<resources>
        <style name="Widget.Button"><item name="textSize">16dp</item><item name="textColor">#FFFFFF</item></style>
        <style name="Widget.Button.Outlined" parent="Widget.Button"><item name="textColor">#1976D2</item></style>
    </resources>`
    reg, _ := LoadStylesFromXML([]byte(stylesXML))
    resolved := reg.Resolve("Widget.Button.Outlined")
    if resolved["textSize"] != "16dp" { t.Error("should inherit textSize from parent") }
    if resolved["textColor"] != "#1976D2" { t.Error("should override textColor") }
}
```

- [ ] **Step 2: Implement Theme**

`theme/theme.go`:
- Parse themes.xml and styles.xml
- Theme struct: holds semantic color/style attributes
- StyleRegistry: stores named styles, resolves parent chain (explicit parent= and implicit dot notation)
- ResolveAttr: lookup ?attr/key in current theme
- Style resolution chain: inline → @style/ → theme default → hardcoded fallback
- Built-in Theme.Wind UI.Light and Theme.Wind UI.Dark

- [ ] **Step 3: Run tests**

```bash
go test ./theme/ -v
```

- [ ] **Step 4: Commit**

```bash
git add theme/
git commit -m "feat(theme): implement Theme loading, style inheritance, and ?attr/ resolution"
```

---

## Task 16: Application & Render Loop

**Files:**
- Create: `app/application.go`, `app/render_loop.go`

- [ ] **Step 1: Implement Application**

`app/application.go`:
- Application struct: holds Platform, ResourceManager, Theme, LayoutInflater, Window list
- NewApplication(): detect platform (Windows for now), initialize ResourceManager, create LayoutInflater with built-in view registry (registerBuiltinViews)
- SetTheme, Resources(), Inflater(), CreateWindow
- Run(): delegates to platform.RunMainLoop()

```go
// Application 核心结构
type Application struct {
    platform    platform.Platform
    resources   *res.ResourceManager
    theme       *theme.Theme
    inflater    *res.LayoutInflater
    windows     []platform.Window
}

func NewApplication() *Application {
    app := &Application{}
    app.platform = windows.NewPlatform()
    app.resources = res.NewResourceManager(nil)
    app.inflater = res.NewLayoutInflater(app.resources)
    res.RegisterBuiltinViews(app.inflater) // 注册 LinearLayout, FrameLayout, TextView, ImageView, Button, View
    return app
}

func (a *Application) Resources() *res.ResourceManager    { return a.resources }
func (a *Application) Inflater() *res.LayoutInflater       { return a.inflater }
func (a *Application) SetTheme(t *theme.Theme)             { a.theme = t }
```

`app/render_loop.go`:
- RenderLoop: connects dirty tracking → measure → arrange → paint → present
- MeasureNode: recursive — if node has Layout call Layout.Measure, else call Painter.Measure
- ArrangeNode: recursive — call Layout.Arrange
- PaintNode: recursive depth-first — Canvas.Save/Translate/ClipRect, Painter.Paint, recurse children, Restore
- RGBA→BGRA conversion: only on Windows, only for dirty regions
- Trigger: called by Window on WM_APP_PAINT (invalidation-driven)

```go
func PaintNode(node *core.Node, canvas core.Canvas) {
    if node.GetVisibility() != core.Visible { return }
    canvas.Save()
    b := node.Bounds()
    canvas.Translate(b.X, b.Y)
    // Painter draws content (background, text, image, etc.)
    if p := node.GetPainter(); p != nil {
        p.Paint(node, canvas)
    }
    // Recurse children
    for _, child := range node.Children() {
        PaintNode(child, canvas)
    }
    canvas.Restore()
}
```

- [ ] **Step 2: Commit**

```bash
git add app/
git commit -m "feat(app): implement Application lifecycle and render loop (measure → arrange → paint → present)"
```

---

## Task 17: Integration — Hello World Demo

**Files:**
- Create: `examples/hello/main.go`, `examples/hello/res/layout/main.xml`, `examples/hello/res/values/strings.xml`, `examples/hello/res/values/colors.xml`, `examples/hello/res/values/styles.xml`

- [ ] **Step 1: Create resource files**

`examples/hello/res/layout/main.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<LinearLayout
    width="match_parent"
    height="match_parent"
    orientation="vertical"
    padding="24dp"
    background="@color/background">

    <TextView
        id="title"
        width="match_parent"
        height="wrap_content"
        text="@string/app_title"
        textSize="28dp"
        textColor="@color/textPrimary"
        gravity="center" />

    <TextView
        width="match_parent"
        height="wrap_content"
        text="@string/description"
        textSize="16dp"
        textColor="@color/textSecondary"
        gravity="center" />

    <LinearLayout
        width="match_parent"
        height="wrap_content"
        orientation="horizontal"
        spacing="12dp"
        gravity="center">

        <Button
            id="btn_cancel"
            width="0"
            height="48dp"
            weight="1"
            text="@string/cancel"
            style="@style/Widget.Button.Outlined" />

        <Button
            id="btn_ok"
            width="0"
            height="48dp"
            weight="1"
            text="@string/confirm"
            style="@style/Widget.Button" />
    </LinearLayout>
</LinearLayout>
```

`examples/hello/res/values/strings.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_title">Hello Wind UI!</string>
    <string name="description">A lightweight native UI framework for Go</string>
    <string name="confirm">OK</string>
    <string name="cancel">Cancel</string>
</resources>
```

`examples/hello/res/values/colors.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <color name="primary">#1976D2</color>
    <color name="background">#FFFFFF</color>
    <color name="textPrimary">#212121</color>
    <color name="textSecondary">#757575</color>
</resources>
```

`examples/hello/res/values/styles.xml`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <style name="Widget.Button">
        <item name="background">#1976D2</item>
        <item name="textColor">#FFFFFF</item>
        <item name="textSize">16dp</item>
        <item name="cornerRadius">4dp</item>
    </style>
    <style name="Widget.Button.Outlined" parent="Widget.Button">
        <item name="background">#00000000</item>
        <item name="textColor">#1976D2</item>
        <item name="borderColor">#1976D2</item>
        <item name="borderWidth">1dp</item>
    </style>
</resources>
```

- [ ] **Step 2: Create main.go**

```go
// examples/hello/main.go
package main

import (
    "embed"
    "fmt"
    "runtime"

    "windui/app"
    "windui/core"
    "windui/platform"
    "windui/widget"
)

//go:embed res
var resources embed.FS

func main() {
    runtime.LockOSThread()

    application := app.NewApplication()
    application.Resources().SetEmbedded(resources)

    window, err := application.CreateWindow(platform.WindowOptions{
        Title:  "Hello Wind UI",
        Width:  400,
        Height: 300,
    })
    if err != nil {
        panic(err)
    }

    root := application.Inflater().Inflate("@layout/main")
    window.SetContentView(root)

    // Wire up button clicks via FindViewById → type assertion to *widget.Button
    if v := root.FindViewById("btn_ok"); v != nil {
        if btn, ok := v.(*widget.Button); ok {
            btn.SetOnClickListener(func(view core.View) {
                fmt.Println("OK clicked!")
                // Update title text to demonstrate dynamic UI update
                if tv := root.FindViewById("title"); tv != nil {
                    if title, ok := tv.(*widget.TextView); ok {
                        title.SetText("Button Clicked!")
                    }
                }
            })
        }
    }
    if v := root.FindViewById("btn_cancel"); v != nil {
        if btn, ok := v.(*widget.Button); ok {
            btn.SetOnClickListener(func(view core.View) {
                fmt.Println("Cancel clicked!")
                window.Close()
            })
        }
    }

    window.Show()
    window.Center()
    application.Run()
}
```

- [ ] **Step 3: Build and run**

```bash
cd D:\Develop\workspace\go_dev\go_wui && go run ./examples/hello/
```

Expected: A 400x300 window appears with:
- "Hello Wind UI!" title text centered
- Description text below
- Two buttons side-by-side at bottom: "Cancel" (outlined) and "OK" (filled blue)
- Clicking buttons triggers visual feedback (hover/pressed states)

- [ ] **Step 4: Run all tests**

```bash
go test ./... -v
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add examples/ app/
git commit -m "feat: add Hello World demo app — Phase 1 MVP complete

Demonstrates XML layout, resource loading, theme/style system,
LinearLayout with weight distribution, and Button click handling."
```

---

## Task 18: Animation Basics

**Files:**
- Create: `core/anim.go`, `core/anim_test.go`

- [ ] **Step 1: Write animation tests**

```go
// core/anim_test.go
func TestLinearInterpolator(t *testing.T) {
    li := &LinearInterpolator{}
    if li.GetInterpolation(0.5) != 0.5 { t.Error("linear 0.5 should be 0.5") }
}

func TestValueAnimator(t *testing.T) {
    var values []float64
    anim := &ValueAnimator{
        From:     0,
        To:       100,
        Duration: 100 * time.Millisecond,
        Interp:   &LinearInterpolator{},
        OnUpdate: func(v float64) { values = append(values, v) },
    }
    // Simulate ticks
    anim.Start()
    anim.Tick(50 * time.Millisecond)  // 50% → 50
    anim.Tick(100 * time.Millisecond) // 100% → 100, finished
    if len(values) < 2 { t.Fatal("expected at least 2 updates") }
    if values[len(values)-1] != 100 { t.Error("final value should be 100") }
    if !anim.IsFinished() { t.Error("should be finished") }
}
```

- [ ] **Step 2: Implement ValueAnimator and Interpolators**

`core/anim.go`: ValueAnimator (From, To, Duration, Interpolator, OnUpdate, OnEnd), Start, Tick, IsFinished. LinearInterpolator, AccelerateDecelerateInterpolator, DecelerateInterpolator.

- [ ] **Step 3: Run tests**

```bash
go test ./core/ -run TestAnim -v
go test ./core/ -run TestLinear -v
go test ./core/ -run TestValue -v
```

- [ ] **Step 4: Commit**

```bash
git add core/anim.go core/anim_test.go
git commit -m "feat(core): add ValueAnimator with Linear/AccelerateDecelerate/Decelerate interpolators"
```

---

## Summary

| Task | Description | Key Deliverable |
|------|-------------|-----------------|
| 1 | Core Types | Rect, Size, Dimension, color parsing |
| 2 | Node Tree | Node struct, View interface, FindViewById, tree ops |
| 3 | Style System | Style merge/inheritance |
| 4 | Event System | 3-phase dispatch, Command, Focus |
| 5 | Canvas/Paint/Text 接口 | 在 core/ 中定义，避免 import cycle |
| 6 | Platform Abstractions | Platform, Window interfaces |
| 7 | GG Canvas | gogpu/gg Canvas implementation |
| 8 | FreeType Text | Cross-platform text rendering + tests |
| 9 | Windows Platform | Win32 window, message loop, DPI + tests |
| 10 | DirectWrite Text | Windows-native text rendering |
| 11 | Layouts | LinearLayout, FrameLayout |
| 12 | Widgets | View, TextView, ImageView, Button |
| 13 | Resources | ResourceManager, value parsing |
| 14 | XML Inflater | LayoutInflater, XML → Node tree, RegisterBuiltinViews |
| 15 | Theme | Theme/Style loading and resolution |
| 16 | Application | App lifecycle, Inflater(), render loop |
| 17 | Hello World | Integration demo with working button clicks |
| 18 | Animation | ValueAnimator, Interpolators |

**Phase 1 完成标志：** `examples/hello/` 运行后显示一个包含标题文本、描述文本和两个按钮的窗口，按钮可点击有视觉反馈。所有测试通过。
