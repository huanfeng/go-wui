// Package theme provides the theming and style system for Wind UI.
//
// A Theme holds semantic color and style attributes that can be switched
// at runtime (e.g. light/dark mode). StyleRegistry stores named styles
// with parent chain resolution, allowing style inheritance similar to
// Android's theme system.
//
// Widgets resolve their visual properties (background color, text color,
// border, corner radius, font size, etc.) through the style system,
// enabling consistent appearance across the application.
package theme
