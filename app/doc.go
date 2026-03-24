// Package app provides the application entry point and render loop for GoWUI.
//
// Application is the top-level object that owns the platform backend,
// resource manager, layout inflater, theme, and window list. A typical
// program creates an Application, loads embedded resources, creates a
// window, inflates an XML layout, and enters the message loop:
//
//	application := app.NewApplication()
//	resFS, _ := fs.Sub(resources, "res")
//	application.SetEmbeddedResources(resFS)
//	window, _ := application.CreateWindow(platform.WindowOptions{...})
//	root := application.Inflater().Inflate("@layout/main")
//	window.SetContentView(root)
//	window.Show()
//	application.Run()
//
// The render loop (PaintNode) recursively walks the node tree, applying
// translations and delegating to each node's Painter. It respects
// visibility, the paintsChildren flag (for self-painting containers like
// ScrollView), and the overlay drawing order.
package app
