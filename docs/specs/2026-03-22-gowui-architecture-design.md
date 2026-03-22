# GoWUI Architecture Design

> Go 原生轻量 UI 框架架构设计文档
>
> 日期: 2026-03-22
> 状态: Draft

---

## 1. 概述

### 1.1 项目定位

GoWUI 是一个基于 Go 语言的原生轻量 UI 框架库，目标是方便开发小巧、内存占用优秀的桌面工具和应用程序。框架借鉴了 WindInput（Go + gg + DirectWrite 实现的输入法服务应用）的设计理念，在极低内存占用下实现高质量的 UI 渲染。

### 1.2 核心特性

- **轻量级 Node 组合模型** — 组合优于继承，Node 本身极轻量
- **Android 风格 XML UI 描述** — 熟悉的 Layout/View 体系和命名约定
- **多后端文本渲染** — DirectWrite（Windows）/ FreeType（跨平台），系统字体零内存开销
- **跨平台架构** — 接口层平台无关，Backend 层按平台实现
- **资源打包系统** — go:embed 内嵌 + .gwpack 资源包 + 外部目录三层叠加
- **主题与样式继承** — Theme/Style 体系，支持明暗主题切换

### 1.3 目标应用场景

初期聚焦桌面小工具场景（系统托盘、浮动窗口、设置面板），架构可扩展到更复杂的桌面应用。

### 1.4 跨平台策略

- **Windows 优先**，完善后再扩展其他平台
- **macOS 次之，Linux 最后**
- 接口从一开始设计为跨平台抽象

### 1.5 与 WindInput 的关系

完全独立的项目。WindInput 继续使用自己的 UI 代码，GoWUI 借鉴其设计理念和架构经验。

---

## 2. 整体架构分层

```
┌─────────────────────────────────────────────────┐
│              Application Layer                   │  ← 用户应用代码
│         (XML 加载 / Go API 构建 UI)              │
├─────────────────────────────────────────────────┤
│              Widget Layer                        │  ← 控件库
│    Button, TextView, ImageView, CheckBox...      │
├─────────────────────────────────────────────────┤
│              Core Layer                          │  ← 框架核心
│  Node Tree │ Layout Engine │ Style │ Event Bus   │
├─────────────────────────────────────────────────┤
│              Render Layer                        │  ← 渲染抽象
│    Canvas API │ Painter │ TextRenderer           │
├─────────────────────────────────────────────────┤
│              Platform Layer                      │  ← 平台抽象
│  Window │ Input │ Clipboard │ DPI │ NativeEdit   │
├─────────────────────────────────────────────────┤
│              Backend Layer                       │  ← 具体实现
│  gg │ DirectWrite │ FreeType │ Win32 │ (未来...)  │
└─────────────────────────────────────────────────┘
```

| 层 | 职责 | 平台依赖 |
|---|------|---------|
| **Backend** | 具体的图形/文本/窗口 API 调用 | 完全平台相关 |
| **Platform** | 统一的平台抽象接口 | 接口平台无关，实现平台相关 |
| **Render** | Canvas 绘图抽象、Painter 协议、文本渲染接口 | 平台无关 |
| **Core** | Node 树、布局计算、样式解析、事件分发 | 完全平台无关 |
| **Widget** | 内置控件 | 平台无关（EditText 除外） |
| **Application** | XML 加载、资源管理、应用生命周期 | 平台无关 |

**设计原则：**

- Core 层以上完全不知道操作系统的存在
- 跨平台扩展只需在 Backend + Platform 层增加新实现
- Widget 层只有 EditText 依赖 Platform 层的原生控件

---

## 3. Node 核心与组合模型

Node 是 UI 树的基本单元，通过组合而非继承获得能力。

### 3.0 Node 与 View 的关系

在 GoWUI 中，**Node** 是内部树节点类型，**View** 是面向用户的 API 概念（对齐 Android 命名）。每个 Widget（如 TextView、Button）是一个包装了 `*Node` 的具体类型，提供类型安全的访问器方法。

```go
// View 是所有 UI 元素的公共接口
type View interface {
    Node() *Node             // 获取底层 Node
    SetId(id string)
    GetId() string
    SetVisibility(v Visibility)
    SetEnabled(enabled bool)
    SetOnClickListener(fn func(v View))
    // ... 通用 View 方法
}

// TextView 是具体的 Widget 类型，包装 Node
type TextView struct {
    node *Node
}

func (t *TextView) Node() *Node          { return t.node }
func (t *TextView) SetText(text string)  { /* 更新 node 属性并标记 dirty */ }
func (t *TextView) GetText() string      { /* ... */ }

// FindViewById 返回 View 接口，可以类型断言为具体 Widget
func (n *Node) FindViewById(id string) View
```

这样既保持了 Node 内部的轻量组合模型，又为使用者提供了类型安全的 Widget API。

### 3.1 Node 结构

```go
type Node struct {
    // 树结构
    parent   *Node
    children []*Node

    // 几何信息（layout 阶段填充）
    bounds   Rect
    padding  Insets
    margin   Insets

    // 组合组件（按需挂载，nil = 不需要 = 零开销）
    layout   Layout
    painter  Painter
    handler  Handler
    style    *Style

    // 状态
    id       string
    tag      string    // 节点类型标签，如 "Button", "TextView"
    visible  bool
    enabled  bool
    dirty      bool
    childDirty bool      // 子树中有 dirty 节点（向上冒泡标记）
}
```

### 3.2 四个可插拔组件接口

```go
// Layout — 布局策略，决定子节点如何排列
type Layout interface {
    Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
    Arrange(node *Node, bounds Rect)
}

// Painter — 绘制逻辑，决定节点如何呈现
// 同时负责叶子节点的内容测量（如文本宽高、图片尺寸）
type Painter interface {
    // Measure 测量内容的固有尺寸（叶子节点必须实现）
    Measure(node *Node, widthSpec, heightSpec MeasureSpec) Size
    // Paint 绘制节点内容
    Paint(node *Node, canvas Canvas)
}

// Handler — 事件处理
type Handler interface {
    OnDispatchEvent(node *Node, event Event) bool
    OnInterceptEvent(node *Node, event Event) bool
    OnEvent(node *Node, event Event) bool
}

// Style — 样式属性集合（数据结构，非接口）
type Style struct {
    BackgroundColor color.Color
    BorderColor     color.Color
    BorderWidth     float64
    CornerRadius    float64
    FontSize        float64
    FontFamily      string
    FontWeight      int
    TextColor       color.Color
    Opacity         float64
}
```

### 3.3 组合示例

**Button:**

```
Node (tag="Button")
 ├─ Layout:  nil（叶子节点，由 Painter.Measure 测量内容尺寸）
 ├─ Painter: ButtonPainter（测量文字尺寸 + 绘制背景+边框+圆角+文字）
 ├─ Handler: ButtonHandler（处理 click/hover/pressed 状态）
 └─ Style:   从主题继承 + XML 内联覆盖
```

**LinearLayout 容器:**

```
Node (tag="LinearLayout")
 ├─ Layout:  LinearLayout{orientation: Vertical}
 ├─ Painter: BackgroundPainter（仅绘制背景色/边框）
 ├─ Handler: nil（容器通常不直接处理事件）
 ├─ Style:   从主题继承
 └─ Children: [Button, TextView, ImageView, ...]
```

### 3.4 设计优势

- Node 本身非常轻量（几个指针 + 几何数据 + 标志位），远比 Android View 的数百个字段精简
- nil 组件零开销
- 新增控件 = 实现 View 接口 + Painter（含 Measure）+ 可选 Handler，不需要继承基类
- 新增布局 = 实现 Layout 接口

---

## 4. 布局系统

采用 Android 风格的两阶段流程：Measure → Arrange。

### 4.1 测量规格

```go
type MeasureMode int

const (
    MeasureModeExact   MeasureMode = iota  // 精确尺寸（width="200dp" 或 match_parent 由父容器决定确切大小）
    MeasureModeAtMost                       // 最大不超过（wrap_content，子节点自行决定尺寸但不超过此值）
    MeasureModeUnbound                      // 不限制（ScrollView 子节点，可以无限大）
)

type MeasureSpec struct {
    Mode MeasureMode
    Size float64
}
```

### 4.2 布局流程

```
Measure 阶段（自底向上汇总尺寸）
 ├─ 叶子节点（有 Painter 无 Layout）：调用 Painter.Measure() 测量内容固有尺寸
 └─ 容器节点（有 Layout）：Layout.Measure() 递归测量子节点，汇总确定自身尺寸

Arrange 阶段（自顶向下分配位置）
 ├─ 根节点获得窗口尺寸作为 bounds
 └─ 容器节点：Layout.Arrange() 按策略给每个子节点分配位置和尺寸
```

测量分发逻辑：框架在遍历 Node 树时，若 `node.layout != nil` 则调用 `Layout.Measure()`，否则调用 `Painter.Measure()` 获取叶子节点的固有尺寸。若 layout 和 painter 均为 nil，节点尺寸为 0x0（可用作纯间距占位节点）。

### 4.3 初期布局类型

- **LinearLayout** — 线性排列（vertical/horizontal），支持 weight、gravity、spacing
- **FrameLayout** — 帧布局，子节点叠加，通过 gravity 定位
- **ScrollLayout** — 滚动容器的布局策略（单方向），由 ScrollView 控件内部使用

### 4.4 尺寸描述

```go
type DimensionUnit int

const (
    DimensionPx          DimensionUnit = iota  // 像素值
    DimensionDp                                 // 密度无关像素
    DimensionMatchParent                        // 填满父容器
    DimensionWrapContent                        // 包裹内容
    DimensionWeight                             // 权重分配
)
```

### 4.5 dp 单位与 DPI 适配

- 基准密度 = 96 DPI（Windows 100% 缩放），1dp = 1px @ 96DPI
- 150% 缩放时 1dp = 1.5px，自动换算
- 纯像素 px 值也支持

### 4.6 脏标记优化

节点属性变化 → 标记 dirty → 向上冒泡 childDirty → 下一帧只从最高 dirty 祖先开始重新 measure/arrange，未变化的子树跳过。

---

## 5. 渲染管线与文本后端

### 5.1 Canvas 接口

```go
type Canvas interface {
    // 基础图形
    DrawRect(rect Rect, paint *Paint)
    DrawRoundRect(rect Rect, radius float64, paint *Paint)
    DrawCircle(cx, cy, radius float64, paint *Paint)
    DrawLine(x1, y1, x2, y2 float64, paint *Paint)
    DrawImage(img *ImageResource, dst Rect)

    // 文本
    DrawText(text string, x, y float64, paint *Paint)
    MeasureText(text string, paint *Paint) Size

    // 状态管理
    Save()
    Restore()
    Translate(dx, dy float64)
    ClipRect(rect Rect)

    // 输出
    Target() *image.RGBA
}
```

### 5.2 Paint（画笔）

```go
type Paint struct {
    Color       color.Color
    DrawStyle   PaintStyle    // Fill / Stroke / FillAndStroke
    StrokeWidth float64
    FontSize    float64
    FontFamily  string
    FontWeight  int
    AntiAlias   bool
}
```

注意：使用 `DrawStyle` 而非 `Style`，避免与 Node 的 `*Style`（样式属性集）产生命名混淆。

### 5.3 默认实现：GGCanvas

使用 `gogpu/gg` 实现 Canvas 接口。Canvas 内部持有 `TextRenderer` 实例，`DrawText`/`MeasureText` 调用委托给 `TextRenderer` 处理。

渲染流程：

1. Node Tree 深度优先遍历
2. 对每个可见节点：Save → Translate → ClipRect → Painter.Paint → 递归 children → Restore
3. 遍历完成后：canvas.Target() 获取 `*image.RGBA`
4. RGBA → BGRA 通道转换（Windows，仅脏区域转换以降低开销）
5. UpdateLayeredWindow / BitBlt 提交到屏幕

### 5.3.1 渲染调度（Render Scheduling）

GoWUI 采用**按需渲染**（invalidation-driven）模式，而非固定帧率：

```
节点属性变化 / 动画 tick
  → Window.Invalidate()
    → 向平台投递重绘请求（Windows: PostMessage(WM_APP_PAINT)）
      → 同一帧内多次 Invalidate 合并为一次重绘
        → 消息循环处理重绘请求
          → measure → arrange → paint → present
```

- **无动画时**：无 Invalidate 则不渲染，CPU 占用为零
- **动画进行时**：Animator 通过定时器持续调用 Invalidate，驱动每帧更新
- **脏区优化**：InvalidateRect 标记局部区域，paint 阶段可跳过未变化的子树

### 5.3.2 动画系统（Animation）

提供基础的 ValueAnimator 驱动属性动画和滚动惯性：

```go
// ValueAnimator 驱动数值从 start 到 end 的平滑过渡
type ValueAnimator struct {
    Duration     time.Duration
    Interpolator Interpolator    // Linear / AccelerateDecelerate / Decelerate
    OnUpdate     func(value float64)
    OnEnd        func()
}

// Interpolator 插值器接口
type Interpolator interface {
    GetInterpolation(fraction float64) float64  // input 0.0~1.0 → output 0.0~1.0
}

// 内置插值器
type LinearInterpolator struct{}
type AccelerateDecelerateInterpolator struct{}
type DecelerateInterpolator struct{}
```

使用场景：
- **按钮按压反馈**：背景色从 normal → pressed，200ms 过渡
- **ScrollView 惯性滚动**：OnFling 触发 DecelerateInterpolator 驱动 scrollOffset
- **ProgressBar 动画**：持续旋转/进度平滑过渡
- **主题切换**：可选渐变过渡（标记全树 dirty，合并为单帧重绘）

Animator 通过平台定时器（Windows: SetTimer 或 PostMessage 循环）驱动，每 tick 调用 OnUpdate → 修改节点属性 → Invalidate → 下一帧渲染。Phase 1 实现基础 ValueAnimator 即可，高级动画（如 ObjectAnimator、AnimatorSet）留到后续阶段。

### 5.4 TextRenderer 接口

```go
type TextRenderer interface {
    SetFont(fontFamily string, weight int, size float64)
    MeasureText(text string) Size
    DrawText(canvas Canvas, text string, x, y float64, paint *Paint)
    CreateTextLayout(text string, paint *Paint, maxWidth float64) *TextLayoutResult
    Close()
}
```

### 5.5 文本后端

| 平台 | 首选 | 回退 |
|------|------|------|
| Windows | DirectWrite | FreeType |
| macOS（未来） | CoreText | FreeType |
| Linux（未来） | FreeType | — |

DirectWrite 通过 C++ shim DLL 调用，系统字体渲染，零字体内存开销。FreeType 为跨平台兜底（纯 Go 实现，通过 gogpu/gg text）。

### 5.6 文本排版（TextLayout）

```go
type TextLayout struct {
    Text        string
    MaxWidth    float64
    MaxLines    int
    Ellipsize   Ellipsize       // None / End / Middle
    Alignment   TextAlignment   // Start / Center / End
    LineSpacing float64         // 行间距倍数（默认 1.2）
}

type TextLayoutResult struct {
    Lines     []TextLine
    TotalSize Size
}

type TextLine struct {
    Text     string
    Offset   Point
    Width    float64
    Baseline float64
}
```

排版能力：自动换行、maxLines 限制、ellipsize 截断、文本对齐、行间距、字体回退（segmentByFont）。

DirectWrite 后端利用 IDWriteTextLayout 原生排版；FreeType 后端自行实现贪心换行算法。

---

## 6. 事件系统

### 6.1 事件基础接口与类型

```go
// Event 基础事件接口，所有具体事件类型实现此接口
type Event interface {
    Type() EventType
    Target() *Node       // 事件目标节点
    Consumed() bool      // 是否已被消费
    Consume()            // 标记为已消费
}

// Handler 通过类型断言获取具体事件：
//   func (h *MyHandler) OnEvent(node *Node, event Event) bool {
//       switch e := event.(type) {
//       case *MotionEvent: // 处理指针事件
//       case *KeyEvent:    // 处理键盘事件
//       }
//   }
```

```go
type EventType int

const (
    EventMotion        EventType = iota  // 指针事件（鼠标+触摸统一）
    EventClick
    EventLongClick
    EventScroll
    EventKeyDown
    EventKeyUp
    EventFocusChanged
    EventMenuCommand
    EventShortcut
)
```

### 6.2 统一指针模型（鼠标+触摸）

```go
type PointerSource int

const (
    PointerMouse PointerSource = iota
    PointerTouch
    PointerPen
)

type MotionEvent struct {
    Action   MotionAction   // ActionDown / ActionMove / ActionUp / ActionCancel / ActionHover*
    Source   PointerSource
    X, Y     float64        // 相对于目标节点
    RawX     float64        // 屏幕绝对坐标
    RawY     float64
    Pointers []Pointer      // 多点触控
    Button   MouseButton    // 仅 PointerMouse
    Modifier KeyModifier
    Pressure float32        // 触摸/笔压感
}

type Pointer struct {
    ID       int
    X, Y     float64
    Pressure float32
}
```

### 6.3 事件分发（三阶段，对齐 Android）

```
Phase 1: dispatchEvent（自顶向下）— 找到命中的子节点链
Phase 2: onInterceptEvent（拦截机会）— 容器可拦截（如 ScrollView 拦截滑动）
Phase 3: onEvent（自底向上冒泡）— 目标节点先处理，未消费则向上冒泡
```

### 6.4 手势识别（GestureDetector）

```go
type GestureDetector struct {
    OnClick     func(e MotionEvent)
    OnLongClick func(e MotionEvent)
    OnScroll    func(e MotionEvent, distX, distY float64)
    OnFling     func(e MotionEvent, velocityX, velocityY float64)
    OnScale     func(detector *ScaleGestureDetector)
}
```

### 6.5 触摸与鼠标行为差异

| 行为 | 鼠标 | 触摸 |
|------|------|------|
| Hover | HoverEnter/Move/Exit | 不支持 |
| 右键菜单 | Button == Right | LongClick 触发 |
| 滚动 | 滚轮 EventScroll | 手指拖拽 OnScroll |
| 精确点击 | 像素级 | 扩大触摸热区（最小 48dp） |

### 6.6 统一命令系统（Command）

桌面端菜单、工具栏按钮、键盘快捷键统一为 Command：

```go
type Command struct {
    ID       string
    Title    string
    Shortcut KeyBinding
    Icon     string
    Enabled  bool
    Handler  func()
}
```

触发路径：菜单点击 / 工具栏按钮 / 键盘快捷键 → CommandManager.Execute(commandID) → command.Handler()

### 6.7 焦点管理

- Tab / Shift+Tab 在可获焦节点间导航
- `focusable="true"` XML 属性标记
- 键盘事件优先级：CommandManager 快捷键 → 焦点节点 dispatchEvent → Tab 导航

---

## 7. XML UI 描述与解析

### 7.1 XML 结构

```xml
<?xml version="1.0" encoding="utf-8"?>
<LinearLayout
    xmlns:app="https://gowui.dev/schema"
    width="match_parent"
    height="match_parent"
    orientation="vertical"
    padding="16dp">

    <TextView
        id="title"
        width="match_parent"
        height="wrap_content"
        text="@string/app_title"
        textSize="24dp"
        textColor="@color/textPrimary"
        gravity="center" />

    <ImageView
        width="200dp"
        height="200dp"
        src="@drawable/logo"
        scaleType="fitCenter" />

    <Button
        width="match_parent"
        height="48dp"
        text="@string/confirm"
        style="@style/PrimaryButton" />

    <include layout="@layout/footer" />
</LinearLayout>
```

### 7.2 资源引用约定（对齐 Android）

| 语法 | 含义 |
|------|------|
| `@string/key` | 字符串资源 |
| `@color/key` | 颜色资源 |
| `@drawable/key` | 图片资源 |
| `@style/key` | 样式引用 |
| `@layout/key` | 子布局引用 |
| `@dimen/key` | 尺寸资源 |
| `?attr/key` | 当前主题属性引用 |
| `#RRGGBB` / `#AARRGGBB` | 直接颜色值 |

### 7.3 LayoutInflater

```go
type LayoutInflater struct {
    resourceManager *ResourceManager
    viewRegistry    map[string]ViewFactory  // tag → 工厂函数
}

// ViewFactory 创建 Node 并配置组件，框架负责后续的 View 包装
type ViewFactory func(attrs AttributeSet) *Node
```

### 7.4 使用方式

```go
// 方式 1：XML 加载
root := inflater.Inflate("@layout/main")
window.SetContentView(root)

// ID 查找 — FindViewById 返回 View 接口，类型断言为具体 Widget
title := root.FindViewById("title").(*widget.TextView)
title.SetText("动态标题")

// 方式 2：Go 代码构建 — New* 函数返回具体 Widget 类型
layout := widget.NewLinearLayout(gowui.Vertical)
layout.AddView(widget.NewTextView("Hello GoWUI"))
layout.AddView(widget.NewButton("Click Me", func() {
    fmt.Println("clicked!")
}))
window.SetContentView(layout.Node())
```

注意：`SetContentView` 接收 `*Node`，Widget 类型通过 `.Node()` 方法获取底层 Node。`AddView` 接收 `View` 接口，所有 Widget 都实现了该接口。

### 7.5 自定义控件注册

```go
inflater.RegisterView("MyCustomView", func(attrs AttributeSet) *Node {
    node := gowui.NewNode("MyCustomView")
    node.SetPainter(&MyCustomPainter{})
    return node
})
```

---

## 8. 主题与样式系统

### 8.1 三层优先级（从低到高）

```
Theme（全局主题）→ Style（命名样式）→ Inline Attributes（XML 内联）
```

### 8.2 Theme 定义

```xml
<resources>
    <style name="Theme.GoWUI.Light">
        <item name="colorPrimary">#1976D2</item>
        <item name="colorPrimaryDark">#1565C0</item>
        <item name="colorAccent">#FF5722</item>
        <item name="colorBackground">#FFFFFF</item>
        <item name="colorSurface">#F5F5F5</item>
        <item name="textColorPrimary">#212121</item>
        <item name="textColorSecondary">#757575</item>
        <item name="buttonStyle">@style/Widget.Button</item>
        <item name="textViewStyle">@style/Widget.TextView</item>
    </style>
</resources>
```

### 8.3 Style 定义（支持继承）

```xml
<resources>
    <style name="Widget.Button">
        <item name="background">@color/colorPrimary</item>
        <item name="textColor">#FFFFFF</item>
        <item name="textSize">16dp</item>
        <item name="cornerRadius">4dp</item>
    </style>

    <!-- parent 显式继承 -->
    <style name="Widget.Button.Outlined" parent="Widget.Button">
        <item name="background">#00000000</item>
        <item name="borderColor">@color/colorPrimary</item>
    </style>

    <!-- 点号隐式继承 -->
    <style name="Widget.Button.Text">
        <item name="background">#00000000</item>
    </style>
</resources>
```

### 8.4 样式解析链

1. 检查 XML 内联属性
2. 检查 style 引用 → 沿 parent 链向上
3. 检查 Theme 的控件默认样式
4. 框架硬编码默认值

### 8.5 ?attr/ 语义化引用

`?attr/textColorPrimary` 引用当前主题中该属性的值，使同一套 XML 布局自动适配明暗主题。

### 8.6 运行时主题切换

```go
app.SetTheme(gowui.LoadTheme("@style/Theme.GoWUI.Dark"))
```

触发全树样式重新解析 + 重绘。实现上将所有节点标记 dirty，合并为单帧批量重绘。对于初期目标场景（小工具窗口，节点数通常 < 100），这个开销可以忽略。

---

## 9. 资源管理与打包

### 9.1 资源目录结构（对齐 Android）

```
res/
 ├─ values/                    默认语言
 │   ├─ strings.xml
 │   ├─ colors.xml
 │   ├─ dimens.xml
 │   └─ styles.xml
 ├─ values-zh/                 中文覆盖
 ├─ values-ja/                 日文覆盖
 ├─ layout/                    布局文件
 ├─ drawable/                  图片资源
 ├─ drawable-hdpi/             高 DPI 图片
 └─ font/                      自定义字体
```

### 9.2 字符串格式（完全对齐 Android）

```xml
<resources>
    <string name="greeting">Hello, %1$s! You have %2$d messages.</string>
    <string name="apostrophe">It\'s a test</string>
    <string-array name="planets">
        <item>Mercury</item>
        <item>Venus</item>
    </string-array>
    <plurals name="unread_messages">
        <item quantity="one">%d unread message</item>
        <item quantity="other">%d unread messages</item>
    </plurals>
</resources>
```

### 9.3 三层加载优先级（从高到低）

1. **外部散文件目录** — 开发调试、用户自定义，修改后实时生效
2. **外部资源包 .gwpack** — 主题包分发、插件，支持多个包叠加
3. **go:embed 内嵌资源** — 单 exe 分发，始终可用

### 9.4 资源包格式（.gwpack）

实质是 ZIP，包含 manifest.json + res/ 目录。

```json
{
    "name": "Ocean Dark Theme",
    "version": "1.0.0",
    "author": "Author Name",
    "description": "深海暗色主题",
    "minGoWUIVersion": "0.1.0",
    "type": "theme"
}
```

### 9.5 ResourceManager

```go
type ResourceManager struct {
    embedded    fs.FS
    packs       []*ResourcePack
    overrideDir string
    locale      string
    cache       map[string]any
}

func (r *ResourceManager) GetString(key string) string
func (r *ResourceManager) GetColor(key string) color.Color
func (r *ResourceManager) GetDrawable(key string) *ImageResource
func (r *ResourceManager) LoadPack(path string) error
func (r *ResourceManager) LoadPackFromMemory(name string, data []byte) error
func (r *ResourceManager) SetLocale(locale string)
```

### 9.6 语言匹配策略

locale = "zh-CN" 时：values-zh-rCN/ → values-zh/ → values/

### 9.7 go:embed 集成

```go
//go:embed res
var embeddedRes embed.FS

//go:embed themes/dark.gwpack
var darkThemePack []byte

func main() {
    app := gowui.NewApplication()
    app.Resources().SetEmbedded(embeddedRes)
    app.Resources().LoadPackFromMemory("dark", darkThemePack)
}
```

---

## 10. 窗口管理与平台抽象

### 10.1 Platform 接口

```go
type Platform interface {
    OS() OSType
    CreateWindow(opts WindowOptions) (Window, error)
    RunMainLoop()
    PostToMainThread(fn func())
    Quit()
    GetClipboard() Clipboard
    GetScreens() []Screen
    GetPrimaryScreen() Screen
    GetSystemLocale() string
    GetSystemTheme() ThemeMode
    CreateTextRenderer() TextRenderer
    CreateNativeEditText(parent Window) NativeEditText
    ShowMessageDialog(opts MessageDialogOptions) DialogResult
    ShowFileDialog(opts FileDialogOptions) (string, error)
}
```

### 10.2 Window 接口

```go
type WindowOptions struct {
    Title       string
    Width       int        // dp
    Height      int
    Resizable   bool
    Frameless   bool       // 无边框（自绘标题栏）
    TopMost     bool
    Transparent bool       // 透明背景（Layered Window）
    Icon        *ImageResource
}

type Window interface {
    SetContentView(root *Node)
    SetTitle(title string)
    Show()
    Hide()
    Close()
    Minimize()
    Maximize()
    Center()
    SetSize(width, height int)
    SetPosition(x, y int)
    IsVisible() bool
    GetDPI() float64
    SetOnClose(fn func() bool)
    SetOnResize(fn func(w, h int))
    SetOnDPIChanged(fn func(dpi float64))
    NativeHandle() uintptr
    Invalidate()
    InvalidateRect(rect Rect)
}
```

### 10.3 三种窗口模式

- **标准窗口** (Frameless=false) — 系统原生标题栏 + GoWUI 内容区
- **无边框窗口** (Frameless=true) — 全自绘，含标题栏
- **透明窗口** (Transparent=true) — 背景透明，适合浮动面板

### 10.4 NativeEditText（原生文本输入控件）

```go
type NativeEditText interface {
    AttachToNode(node *Node)
    Detach()
    GetText() string
    SetText(text string)
    SetPlaceholder(text string)
    SetFont(family string, size float64, weight int)
    SetMultiLine(multiLine bool)
    SetInputType(inputType InputType)
    SetOnTextChanged(fn func(text string))
    SetOnSubmit(fn func(text string))
    Focus()
    ClearFocus()
}
```

文本输入使用系统原生控件处理，避免 IME、光标、选区、剪贴板等复杂问题。

### 10.5 线程模型

- **主线程（UI 线程）**：runtime.LockOSThread()，运行平台消息循环
- **Worker Goroutines**：不能直接操作 UI，通过 PostToMainThread() 安全更新
- 类比 Android 的 runOnUiThread()

### 10.6 Windows 实现

- CreateWindow → Win32 CreateWindowEx
- RunMainLoop → GetMessage/DispatchMessage
- PostToMainThread → PostMessage 自定义消息
- TextRenderer → DWriteTextRenderer（首选）/ FreeTypeTextRenderer（回退）
- NativeEditText → Win32 EDIT 控件子窗口
- Clipboard → Win32 Clipboard API
- Screens → EnumDisplayMonitors

---

## 11. 控件体系与分阶段计划

### Phase 1 — 基础可用（MVP）

| 控件 | 说明 |
|------|------|
| View | 基础视图（背景、边框、圆角） |
| TextView | 文本显示（单行/多行、ellipsize） |
| ImageView | 图片显示（scaleType） |
| Button | 按钮（click/hover/pressed 状态） |
| LinearLayout | 线性布局（horizontal/vertical、weight） |
| FrameLayout | 帧布局（叠加、gravity） |
| Window | 窗口容器（标准/无边框/透明） |

**交付标准：** 能用 XML 描述一个带文本和按钮的窗口，点击有响应。

### Phase 2 — 交互完善

| 控件 | 说明 |
|------|------|
| EditText | 文本输入（原生控件桥接） |
| CheckBox | 复选框 |
| RadioButton / RadioGroup | 单选 |
| Switch | 开关 |
| ProgressBar | 进度条（linear/circular） |
| ScrollView / HorizontalScrollView | 滚动容器 |
| Divider | 分隔线 |

**交付标准：** 能做设置面板、简单对话框、带输入的小工具。

### Phase 3 — 列表与导航

| 控件 | 说明 |
|------|------|
| RecyclerView | 高性能列表（View 复用） |
| TabLayout | 标签栏 |
| ViewPager | 页面切换 |
| Toolbar | 工具栏 |
| Menu / PopupMenu | 菜单 |
| Dialog / AlertDialog | 对话框 |

**交付标准：** 列表、多页面、工具栏菜单，可开发实用工具。

### Phase 4 — 高级控件

| 控件 | 说明 |
|------|------|
| GridLayout | 网格布局 |
| FlexLayout | 弹性布局 |
| Spinner | 下拉选择 |
| SeekBar | 滑块 |
| Toast / Snackbar | 提示 |
| TreeView | 树形控件 |
| SplitPane | 分栏面板 |

**交付标准：** 复杂布局、高级交互，可开发中等规模桌面应用。

---

## 12. Go 包结构

```
gowui/
 ├─ app/                    应用生命周期（Application, 入口）
 │
 ├─ core/                   核心（平台无关）
 │   ├─ node.go             Node 结构体、树操作、View 接口定义
 │   ├─ layout.go           Layout 接口、MeasureSpec（接口定义，非具体实现）
 │   ├─ style.go            Style 结构体
 │   ├─ event.go            Event 接口、EventType、MotionEvent、KeyEvent 等
 │   ├─ event_dispatch.go   事件分发引擎
 │   ├─ command.go          Command 系统、CommandManager
 │   └─ anim.go             ValueAnimator、Interpolator
 │
 ├─ render/                 渲染抽象（接口定义）
 │   ├─ canvas.go           Canvas 接口
 │   ├─ paint.go            Paint 结构体
 │   └─ text.go             TextRenderer 接口、TextLayout
 │
 ├─ render/gg/              gg Canvas 具体实现
 │   └─ canvas_gg.go        GGCanvas（内部持有 TextRenderer）
 │
 ├─ render/freetype/        FreeType TextRenderer 具体实现
 │   └─ text_freetype.go
 │
 ├─ platform/               平台抽象接口（接口定义，非具体实现）
 │   ├─ platform.go         Platform 接口
 │   ├─ window.go           Window 接口、WindowOptions
 │   └─ clipboard.go        Clipboard 接口
 │
 ├─ platform/windows/       Windows 平台具体实现
 │   ├─ platform_windows.go
 │   ├─ window_windows.go
 │   ├─ dwrite_text.go      DirectWrite TextRenderer
 │   └─ native_edit.go      Win32 EDIT 控件桥接
 │
 ├─ widget/                 内置控件（View 接口的具体实现）
 │   ├─ view.go             BaseView（公共 View 实现）
 │   ├─ textview.go         TextView
 │   ├─ imageview.go        ImageView
 │   ├─ button.go           Button
 │   └─ ...
 │
 ├─ layout/                 布局具体实现（Layout 接口的具体实现）
 │   ├─ linear.go           LinearLayout
 │   ├─ frame.go            FrameLayout
 │   └─ scroll.go           ScrollLayout（由 ScrollView 内部使用）
 │
 ├─ res/                    资源管理
 │   ├─ manager.go          ResourceManager
 │   ├─ inflater.go         LayoutInflater
 │   ├─ pack.go             .gwpack 加载
 │   └─ values.go           strings/colors/dimens 解析
 │
 └─ theme/                  主题系统
     ├─ theme.go            Theme 加载与管理
     └─ style_resolver.go   样式解析链（Theme → Style → Inline）
```

注意包的分层：`core/layout.go` 定义 Layout **接口**和 MeasureSpec；`layout/` 包含 LinearLayout、FrameLayout 等**具体实现**。同理 `render/text.go` 定义 TextRenderer **接口**；`render/freetype/` 和 `platform/windows/dwrite_text.go` 是**具体实现**。

---

## 13. 技术选型总结

| 维度 | 选择 | 理由 |
|------|------|------|
| 语言 | Go | 已验证（WindInput），编译为单 exe，跨平台 |
| 图形库 | gogpu/gg（默认后端） | 已验证，跨平台，API 简洁，接口可插拔 |
| 文本渲染(Win) | DirectWrite | 系统字体零内存，渲染质量与平台一致 |
| 文本渲染(跨平台) | FreeType (gg/text) | 纯 Go 实现，任何平台可用 |
| UI 描述 | XML | 对齐 Android 约定，工具链可复用 |
| 资源格式 | Android 兼容（strings.xml 等） | 工具链复用，开发者熟悉 |
| 布局模型 | Android 风格 measure/arrange | 成熟、直觉、开发者熟悉 |
| 事件模型 | dispatch/intercept/event 三阶段 + Command | Android 事件分发 + 桌面命令统一 |
| 窗口 | Win32 API（Windows） | 原生性能，支持 Layered Window |
| 资源打包 | .gwpack (ZIP) | Go 标准库支持，随机访问，embed 兼容 |
