// Package core provides the foundational types and interfaces for the GoWUI framework.
//
// It defines the Node tree structure, View interface, layout engine protocols,
// event dispatch system, Canvas/Paint drawing abstractions, style management,
// animation support, focus tracking, and command/shortcut handling.
//
// Core is platform-independent and sits at the center of the framework's
// six-layer architecture. All other packages (layout, widget, render, platform)
// depend on core but core depends on nothing else.
//
// # Node and View
//
// Node is the internal tree element backing every UI component. Each widget
// wraps a *Node and implements the View interface, which provides type-safe
// accessors following Android-style naming conventions (SetId, FindViewById,
// SetVisibility, etc.).
//
// Capabilities are attached to a Node through composition:
//   - Layout: measures and arranges child nodes (Measure + Arrange)
//   - Painter: measures the node itself and paints its visual content
//   - Handler: processes input events (touch, keyboard, focus)
//   - Style: holds visual properties (background, border, font, colors)
//
// # Event System
//
// Events follow a three-phase dispatch model: Capture → Target → Bubble,
// consistent with Android and Web event propagation. Supported event types
// include pointer input (mouse/touch/pen), keyboard input, focus changes,
// scroll, and application-level commands.
//
// # Layout Protocol
//
// The layout system uses a two-pass approach: Measure (determine desired size
// given constraints) and Arrange (assign final bounds to children). MeasureSpec
// carries a mode (Exact, AtMost, Unbound) and a size value, mirroring
// Android's MeasureSpec design.
//
// # Animation
//
// ValueAnimator drives float64 values from a start to an end value over a
// duration, with pluggable interpolators (Linear, AccelerateDecelerate,
// Decelerate).
package core
