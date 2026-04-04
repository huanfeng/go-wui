// Package widget provides the built-in UI controls for Wind UI.
//
// Every widget wraps a *core.Node and implements the core.View interface.
// Widgets follow an Android-inspired naming convention and API style.
//
// # Basic Widgets
//
//   - View: generic container with background and border support.
//   - TextView: single or multi-line text display.
//   - ImageView: image display with scaling.
//   - Button: clickable control with pressed/disabled states.
//   - Divider: visual separator line.
//
// # Input Widgets
//
//   - EditText: text input bridged to the platform's native text control.
//   - CheckBox: checkable box with a label.
//   - RadioButton / RadioGroup: exclusive single-selection.
//   - Switch: toggle on/off control.
//   - SeekBar: horizontal slider for value selection.
//   - Spinner: dropdown selection.
//
// # Container Widgets
//
//   - ScrollView: scrollable container with drag, fling, and mouse wheel support.
//   - ViewPager: swipeable page container.
//   - RecyclerView: efficient list/grid with adapter-based view recycling.
//   - SplitPane: resizable split panel with draggable divider.
//
// # Navigation and Overlay Widgets
//
//   - Toolbar: application bar with title, subtitle, and action items.
//   - TabLayout: tabbed interface header.
//   - Menu / PopupMenu: context and dropdown menus.
//   - Dialog / AlertDialog: modal dialog windows.
//   - Toast: temporary floating notification with optional action button.
//   - TreeView: hierarchical data display with expand/collapse.
//
// Each widget provides a constructor (e.g. NewButton, NewTextView) and
// a custom Painter for rendering. Event handling is wired through the
// node's Handler.
package widget
