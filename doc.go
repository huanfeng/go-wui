// Package gowui is a lightweight, native desktop UI framework for Go.
//
// GoWUI provides an Android-inspired widget and layout system for building
// small, memory-efficient Windows desktop applications. It features XML
// declarative layouts, a composable Node tree, high-quality text rendering
// (DirectWrite / FreeType), DPI awareness, and a rich set of built-in controls.
//
// # Architecture
//
// The framework uses a six-layer architecture:
//
//   - app: application entry point, resource loading, render loop
//   - widget: 30+ built-in controls (Button, TextView, EditText, RecyclerView, etc.)
//   - core: Node tree, event dispatch, layout protocol, Canvas/Paint abstractions
//   - render: Canvas implementation (gg) and text renderers (DirectWrite, FreeType)
//   - platform: OS abstraction interfaces (Window, Clipboard, NativeEditText)
//   - platform/windows: Win32 backend implementation
//
// # Quick Start
//
//	application := app.NewApplication()
//	resFS, _ := fs.Sub(resources, "res")
//	application.SetEmbeddedResources(resFS)
//	window, _ := application.CreateWindow(platform.WindowOptions{
//	    Title: "Hello GoWUI", Width: 400, Height: 300,
//	})
//	root := application.Inflater().Inflate("@layout/main")
//	window.SetContentView(root)
//	window.Show()
//	application.Run()
//
// See the examples/ directory for complete working applications.
package gowui
