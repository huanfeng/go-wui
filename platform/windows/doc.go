// Package windows implements the Wind UI platform interfaces for Microsoft Windows.
//
// It provides:
//   - Win32 window creation and message loop (CreateWindowEx, GetMessage/DispatchMessage)
//   - DirectWrite text rendering via CGO (with FreeType fallback)
//   - DPI detection and per-monitor DPI scaling
//   - Native EDIT control bridge for text input (EditText widget)
//   - Clipboard support via Win32 API
//   - Double-buffered rendering using DIB sections
//
// This package requires Windows and CGO (for DirectWrite integration).
package windows
