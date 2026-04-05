# 渲染性能优化设计 — 方案 B：帧保留 + 局部重绘 + 动画帧循环

**日期**: 2026-04-05
**状态**: 已确认
**范围**: 脏区域传播、帧保留、选择性绘制、局部像素提交、统一动画帧循环

---

## 背景

当前渲染管线每次 `Invalidate()` 都执行全量重绘：

```
clear(全部像素) → Measure(全树) → Arrange(全树) → Paint(全树) → RGBA→BGRA(全量) → BitBlt(全量)
```

`Node` 已有 `dirty` / `childDirty` 标记但未用于优化。动画无集成帧驱动。

## 目标

- 静态 UI 的 hover 等小范围变化：只重绘变化区域
- 动画：统一帧循环驱动，无需控件自管 goroutine/ticker
- 像素提交：只拷贝和 BitBlt 脏矩形区域
- 无变化时跳过整个渲染，CPU 占用为零

---

## §1 脏区域传播（Android invalidateChild 模式）

脏矩形由子节点沿父链向上冒泡，每层做坐标偏移，最终汇聚到根节点。

### Node 层变化（core/node.go）

```go
type Node struct {
    // 已有
    dirty      bool
    childDirty bool

    // 新增：窗口级脏区域（仅根节点持有）
    dirtyRects []Rect
    dirtyFull  bool  // true = 退化为全量重绘
}

// InvalidateRect 标记局部坐标矩形为脏，沿父链冒泡
func (n *Node) InvalidateRect(localRect Rect) {
    n.dirty = true
    rect := localRect
    current := n
    for p := current.parent; p != nil; current, p = p, p.parent {
        cb := current.Bounds()
        rect.X += cb.X
        rect.Y += cb.Y
        p.childDirty = true
    }
    // current 是根节点，rect 是屏幕坐标
    current.addDirtyRect(rect)
}

// Invalidate 标记整个节点为脏
func (n *Node) Invalidate() {
    n.dirty = true
    n.InvalidateRect(Rect{Width: n.Bounds().Width, Height: n.Bounds().Height})
}
```

### 矩形合并策略

- `addDirtyRect`：与现有矩形做重叠测试，重叠则合并为包围矩形
- 上限 8 个矩形，超过退化为 `dirtyFull = true`
- `PopDirtyRegion()` 供 `render()` 取出并清空

### 兼容

现有 `MarkDirty()` 调用点迁移为 `Invalidate()`。

---

## §2 帧保留 + 选择性绘制

### 帧保留

不再每帧 `clear(img.Pix)` 清空整个画布。上一帧像素保留在 `cachedImage` 中，只清除脏矩形区域。

```
旧: NewGGCanvasForImage(img) → clear(全部) → Paint(全树)
新: NewGGCanvasRetained(img)  → ClearRect(脏区域) → Paint(仅脏区域相关节点)
```

GGCanvas 新增：
- `NewGGCanvasRetained(img, tr)` — 不清空像素，保留上帧
- `ClearRect(rect)` — 只清除指定矩形区域

### 选择性绘制

`PaintNodeDirty` 增加脏矩形列表参数，用节点绝对边界与脏矩形做相交测试：

```go
func PaintNodeDirty(node *Node, canvas Canvas, dirtyRects []Rect, ox, oy float64) {
    if !node.IsDirty() && !node.IsChildDirty() {
        return  // 完全干净，跳过
    }
    absRect := Rect{X: ox + b.X, Y: oy + b.Y, Width: b.Width, Height: b.Height}
    if !rectIntersectsAny(absRect, dirtyRects) {
        return  // 不在脏区域内，跳过
    }
    // Save → Translate → Paint → 递归 children → Restore → ClearDirty
}
```

### render() 整合

```go
func (w *win32Window) render() {
    dirtyRects, fullDirty := root.PopDirtyRegion()

    if !fullDirty && len(dirtyRects) == 0 && !w.needsFullRender {
        return  // 无变化，跳过
    }

    // Measure + Arrange 仍全量执行（布局可能联动）

    if fullDirty || sizeChanged {
        // 全量路径（当前行为）
        canvas := NewGGCanvasForImage(cachedImage, tr)
        PaintNode(root, canvas)
        w.present(canvas.Target())
    } else {
        // 局部路径
        canvas := NewGGCanvasRetained(cachedImage, tr)
        for _, r := range dirtyRects { canvas.ClearRect(r) }
        PaintNodeDirty(root, canvas, dirtyRects, 0, 0)
        w.presentDirty(canvas.Target(), dirtyRects)
    }
}
```

---

## §3 局部像素提交

### RGBA→BGRA 局部拷贝

```go
func copyRGBAtoBGRA(src *image.RGBA, dibBits unsafe.Pointer, dibW, x0, y0, x1, y1 int) {
    // 只遍历脏矩形区域的行，逐行做 R/B 交换
}
```

全量 `present()` 也复用此函数（x0=0, y0=0, x1=width, y1=height），统一代码路径。

### 局部 BitBlt

```go
func (w *win32Window) presentDirty(img *image.RGBA, dirtyRects []Rect) {
    w.ensureDIB(img)
    for _, r := range dirtyRects {
        copyRGBAtoBGRA(img, w.dibBits, w.dibWidth, x0, y0, x1, y1)
    }
    for _, r := range dirtyRects {
        procBitBlt.Call(hdc, x0, y0, rw, rh, w.dibMemDC, x0, y0, SRCCOPY)
    }
}
```

---

## §4 统一动画帧循环

### 设计思路（Android Choreographer / Web requestAnimationFrame）

窗口层集成帧调度器，有活跃动画时以 ~60fps 驱动，无动画时停止计时器。

### 窗口层新增

```go
type win32Window struct {
    // 新增
    animTimerID   uintptr
    animators     []*core.ValueAnimator
    lastFrameTime time.Time
    frameCallback func()
}

func (w *win32Window) StartAnimator(anim *core.ValueAnimator)  // 注册动画
func (w *win32Window) RequestFrame()                           // 请求下一帧
func (w *win32Window) ensureFrameTimer()                       // 启动 SetTimer 16ms
func (w *win32Window) stopFrameTimer()                         // KillTimer
```

### WM_TIMER 处理

```go
case WM_TIMER:
    elapsed := now.Sub(w.lastFrameTime)
    w.lastFrameTime = now
    // Tick 所有活跃动画，移除已完成的
    // 执行 frameCallback
    // 无活跃动画 → stopFrameTimer()
```

### 控件使用

```go
// 旧：控件自管 goroutine + ticker
// 新：
anim := &core.ValueAnimator{
    From: 0, To: 1, Duration: 200ms,
    OnUpdate: func(v float64) {
        sw.thumbPosition = v
        node.Invalidate()
    },
}
window.StartAnimator(anim)  // 自动驱动，结束自动停止
```

---

## 实施顺序

1. **§1 脏区域传播** — `core/node.go` 新增 `InvalidateRect` / `addDirtyRect` / `PopDirtyRegion`，迁移 `MarkDirty` 调用点
2. **§2 帧保留 + 选择性绘制** — `render/gg/canvas_gg.go` 新增 `NewGGCanvasRetained` / `ClearRect`，`app/render_loop.go` 新增 `PaintNodeDirty`
3. **§3 局部像素提交** — `window_windows.go` 新增 `presentDirty` / `ensureDIB` 重构
4. **§4 动画帧循环** — `window_windows.go` 新增 `StartAnimator` / `WM_TIMER` 处理
5. **§1→§4 在 render() 中整合** — 局部路径 + 全量路径共存
6. **迁移现有控件** — 将所有 `MarkDirty()` + 手动 `Invalidate()` 迁移为 `node.Invalidate()`

## 不在本次范围

- Direct2D GPU 加速（后续单独设计）
- Layer Cache / 位图缓存（当滚动成为瓶颈时再引入）
- Layout dirty 优化（Measure/Arrange 始终全量）
