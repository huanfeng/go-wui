// Package layout provides the built-in layout engines for GoWUI.
//
// Each layout implements the core.Layout interface (Measure + Arrange) and
// optionally core.DPIScalable for dp-to-px conversion on high-DPI displays.
//
// Available layouts:
//
//   - LinearLayout: arranges children in a single row or column, with optional
//     weight-based size distribution, cross-axis gravity, and spacing.
//   - FrameLayout: stacks children on top of each other with gravity-based
//     positioning, useful for overlays and simple containers.
//   - FlexLayout: flexbox-inspired layout supporting wrap, justify-content,
//     align-items, and line spacing.
//   - GridLayout: arranges children in a rows × columns grid with uniform
//     cell sizing and configurable spacing.
//   - ScrollLayout: manages a single child that may exceed the viewport,
//     tracking scroll offset for ScrollView.
//
// All dimension values specified in dp units are automatically scaled by the
// current display DPI factor.
package layout
