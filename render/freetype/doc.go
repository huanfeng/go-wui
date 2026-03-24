// Package freetype implements the core.TextRenderer interface using the
// golang/freetype library.
//
// FreetypeTextRenderer provides cross-platform font loading, text
// measurement, and text drawing. It serves as the fallback text renderer
// on platforms where DirectWrite or other native text APIs are not
// available.
package freetype
