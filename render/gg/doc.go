// Package gg implements the core.Canvas interface using the fogleman/gg
// 2D graphics library.
//
// GGCanvas provides anti-aliased vector drawing (rectangles, rounded
// rectangles, circles, lines), image rendering, and text drawing
// (delegated to a core.TextRenderer). It maintains a manual clip rect
// stack with early-out optimization and tracks cumulative translation
// offsets for operations that bypass gg's built-in transform.
//
// Canvas buffers are cached and reused across frames to minimize
// memory allocation.
package gg
