# GoWUI

Go 原生轻量桌面 UI 框架 — 用 Go 构建小巧、高效的 Windows 桌面应用。

[![Go Reference](https://pkg.go.dev/badge/github.com/huanfeng/go-wui.svg)](https://pkg.go.dev/github.com/huanfeng/go-wui)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 设计目标

- **轻量级** — 组合优于继承的 Node 模型，极低内存占用
- **Android 风格 API** — 熟悉的 LinearLayout / FrameLayout / View 体系，XML 声明式 UI
- **高质量文本渲染** — DirectWrite (Windows) / FreeType (跨平台) 双后端
- **跨平台架构** — 接口层平台无关，当前优先实现 Windows
- **资源系统** — `go:embed` 内嵌资源 + XML 布局 + 多语言 + 主题样式

## 当前状态

> **Alpha 阶段** — API 可能发生变化，欢迎试用和反馈。

| Phase | 内容 | 状态 |
|-------|------|------|
| Phase 1 — MVP 基础 | Node 树、LinearLayout、FrameLayout、TextView、Button、ImageView、XML 布局、资源系统、主题 | ✅ 完成 |
| Phase 2 — 交互完善 | EditText、CheckBox、RadioButton、Switch、ProgressBar、ScrollView、DPI 感知 | ✅ 完成 |
| Phase 3 — 列表与导航 | Toolbar、TabLayout、ViewPager、RecyclerView、Menu、Dialog、Overlay 系统 | ✅ 完成 |
| Phase 4 — 高级控件 | GridLayout、FlexLayout、Spinner、SeekBar、Toast、TreeView、SplitPane | ✅ 完成 |

**平台支持：**

| 平台 | 状态 |
|------|------|
| Windows | ✅ 完整支持 (Win32 + DirectWrite) |
| macOS | 🔲 计划中 |
| Linux | 🔲 计划中 |

## 快速开始

### 安装

```bash
go get github.com/huanfeng/go-wui@latest
```

### Hello World

创建一个简单的窗口应用：

**main.go**

```go
package main

import (
    "embed"
    "fmt"
    "io/fs"

    "github.com/huanfeng/go-wui/app"
    "github.com/huanfeng/go-wui/core"
    "github.com/huanfeng/go-wui/platform"
    "github.com/huanfeng/go-wui/widget"
)

//go:embed res
var resources embed.FS

func main() {
    application := app.NewApplication()

    resFS, _ := fs.Sub(resources, "res")
    application.SetEmbeddedResources(resFS)

    window, _ := application.CreateWindow(platform.WindowOptions{
        Title:     "Hello GoWUI",
        Width:     400,
        Height:    300,
        Resizable: true,
    })

    root := application.Inflater().Inflate("@layout/main")
    window.SetContentView(root)

    if v := root.FindViewById("btn_ok"); v != nil {
        if btn, ok := v.(*widget.Button); ok {
            btn.SetOnClickListener(func(view core.View) {
                fmt.Println("Clicked!")
            })
        }
    }

    window.Center()
    window.Show()
    application.Run()
}
```

**res/layout/main.xml**

```xml
<?xml version="1.0" encoding="utf-8"?>
<LinearLayout
    width="match_parent"
    height="match_parent"
    orientation="vertical"
    padding="24dp"
    spacing="16dp"
    gravity="center">

    <TextView
        width="match_parent"
        height="wrap_content"
        text="Hello, GoWUI!"
        textSize="28dp"
        gravity="center" />

    <Button
        id="btn_ok"
        width="200dp"
        height="48dp"
        text="Click Me" />
</LinearLayout>
```

**res/values/strings.xml**

```xml
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <string name="app_title">Hello GoWUI</string>
</resources>
```

### 运行

```bash
go run .
```

> **注意：** 当前仅支持 Windows 平台。需要 Go 1.22+ 和 CGO 环境（DirectWrite 文本渲染需要）。

## 架构

GoWUI 采用六层架构，自底向上：

```
┌─────────────────────────────────────────────────┐
│              Application Layer                   │  ← 用户应用代码
│         (XML 加载 / Go API 构建 UI)              │
├─────────────────────────────────────────────────┤
│              Widget Layer                        │  ← 控件库
│    Button, TextView, EditText, RecyclerView...   │
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
│  gg (2D) │ DirectWrite │ FreeType │ Win32        │
└─────────────────────────────────────────────────┘
```

### 目录结构

```
go-wui/
├── core/           # 框架核心：Node 树、事件系统、布局接口、动画
├── layout/         # 布局引擎：LinearLayout, FrameLayout, FlexLayout, GridLayout, ScrollLayout
├── widget/         # 控件库：30+ 控件
├── render/
│   ├── gg/         # 基于 fogleman/gg 的 Canvas 实现
│   └── freetype/   # FreeType 文本渲染器
├── platform/
│   ├── platform.go # 平台抽象接口
│   ├── window.go   # 窗口抽象接口
│   └── windows/    # Win32 实现 (窗口、消息循环、DPI、DirectWrite、原生 EditText)
├── res/            # 资源系统：XML 布局 Inflater、值解析、资源管理器
├── theme/          # 主题系统：语义化颜色/样式、明暗主题
├── app/            # 应用入口：Application、渲染循环
└── examples/       # 示例应用
    ├── hello/      # 最小示例
    ├── showcase/   # 控件展示
    ├── edittext/   # EditText 功能演示
    ├── phase3/     # 列表与导航演示
    └── phase4/     # 高级控件演示
```

### 核心设计

**Node + 组合模型**

每个 UI 元素由 `Node` 树节点 + 组合的能力构成：

| 组件 | 职责 |
|------|------|
| `Node` | 树结构、bounds、padding/margin、子节点管理 |
| `Layout` | 测量 (Measure) 和排列 (Arrange) 子节点 |
| `Painter` | 测量控件大小和绘制内容 |
| `Handler` | 处理输入事件 |
| `Style` | 视觉属性（背景色、边框、字体大小等） |

**事件分发**

三阶段事件传递：**Capture → Target → Bubble**，与 Android/Web 事件模型一致。

**XML 布局**

借鉴 Android 的 XML 布局语法，支持资源引用：

- `@string/name` — 字符串资源
- `@color/name` — 颜色资源
- `@dimen/name` — 尺寸资源
- `@layout/name` — 布局文件引用
- `match_parent` / `wrap_content` — 尺寸模式
- `dp` 单位 — DPI 感知尺寸

## 控件一览

### 基础控件

| 控件 | 说明 |
|------|------|
| `View` | 基础容器，支持背景/边框 |
| `TextView` | 文本显示 |
| `ImageView` | 图片显示 |
| `Button` | 按钮，支持点击状态 |
| `Divider` | 分隔线 |

### 输入控件

| 控件 | 说明 |
|------|------|
| `EditText` | 文本输入（桥接平台原生控件） |
| `CheckBox` | 复选框 |
| `RadioButton` / `RadioGroup` | 单选按钮/互斥选择组 |
| `Switch` | 开关切换 |
| `SeekBar` | 滑块 |
| `Spinner` | 下拉选择 |

### 容器控件

| 控件 | 说明 |
|------|------|
| `ScrollView` | 垂直滚动容器 |
| `ViewPager` | 页面切换容器 |
| `RecyclerView` | 高效列表（Adapter 模式） |
| `SplitPane` | 可拖拽分隔面板 |

### 导航与弹出

| 控件 | 说明 |
|------|------|
| `Toolbar` | 应用栏 |
| `TabLayout` | 标签页 |
| `Menu` / `PopupMenu` | 上下文菜单 |
| `Dialog` / `AlertDialog` | 对话框 |
| `Toast` | 浮动提示（支持 action 按钮） |
| `TreeView` | 树形数据展示 |

### 布局

| 布局 | 说明 |
|------|------|
| `LinearLayout` | 线性排列（水平/垂直），支持 weight 分配 |
| `FrameLayout` | 层叠布局 |
| `FlexLayout` | Flexbox 模型（wrap、justify、alignItems） |
| `GridLayout` | 网格布局 |
| `ScrollLayout` | 滚动布局 |

## 已知问题与限制

- **仅支持 Windows** — macOS 和 Linux 平台层尚未实现
- **需要 CGO** — DirectWrite 文本渲染依赖 CGO
- **EditText 依赖原生控件** — 使用 Win32 EDIT 控件桥接，而非纯自绘
- **无 GPU 加速** — 当前使用 CPU 软件渲染 (`fogleman/gg`)
- **API 不稳定** — Alpha 阶段，接口可能随版本变化
- **资源包 (.gwpack) 未实现** — 当前仅支持 `go:embed` 内嵌资源
- **无可访问性支持** — 尚未实现 UI Automation / Accessibility 接口

## 依赖

| 依赖 | 用途 |
|------|------|
| [fogleman/gg](https://github.com/fogleman/gg) | 2D 矢量渲染引擎 |
| [golang/freetype](https://github.com/golang/freetype) | 跨平台字体渲染 |
| [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) | 图像处理扩展 |

## 构建示例

```bash
# 克隆仓库
git clone https://github.com/huanfeng/go-wui.git
cd go-wui

# 运行 Hello World 示例
go run ./examples/hello/

# 运行控件展示
go run ./examples/showcase/

# 运行 Phase 4 高级控件演示
go run ./examples/phase4/

# 运行所有测试
go test ./...
```

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## License

本项目采用 [MIT License](LICENSE) 开源。
