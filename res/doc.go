// Package res provides the resource management system for Wind UI.
//
// It handles loading, caching, and resolving resources from multiple sources:
// embedded filesystems (go:embed), override directories, and locale-specific variants
//
// # ResourceManager
//
// ResourceManager loads and caches parsed resource values (strings, colors,
// dimensions, arrays) with locale-based fallback (e.g. values-zh overrides
// values for Chinese locale).
//
// # LayoutInflater
//
// LayoutInflater parses Android-style XML layout files into a live node tree.
// It supports resource references (@string/name, @color/name, @dimen/name),
// dimension units (dp, px), and size keywords (match_parent, wrap_content).
//
// Custom view factories can be registered to extend the inflater with
// application-specific widgets.
//
// # XML Resource Files
//
// Standard resource file formats are supported:
//   - strings.xml: string name-value pairs
//   - colors.xml: color definitions (#RRGGBB, #AARRGGBB)
//   - dimens.xml: dimension values with units
//   - arrays.xml: string and integer arrays
package res
