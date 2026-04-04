// Package platform defines the platform abstraction interfaces for Wind UI.
//
// These interfaces isolate the framework from operating system specifics,
// allowing the core, layout, widget, and render layers to remain fully
// platform-independent.
//
// Key interfaces:
//
//   - Platform: creates windows, queries screens, detects system theme,
//     and provides platform services (clipboard, native edit controls).
//   - Window: manages a native OS window with content view, geometry,
//     visibility, DPI, and input event callbacks.
//   - Clipboard: text copy/paste operations.
//   - NativeEditText: bridges to the platform's native text input control.
//
// Platform implementations live in sub-packages (e.g. platform/windows
// for Win32). To add a new platform, implement these interfaces and
// register the platform in the app package.
package platform
