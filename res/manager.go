package res

import (
	"image/color"
	"io/fs"
	"strings"

	"github.com/huanfeng/go-wui/core"
)

// ResourceManager loads and caches resource values from embedded or override filesystems.
// It supports locale-based fallback (e.g. values-zh overrides values).
type ResourceManager struct {
	embedded    fs.FS   // go:embed base
	overrideDir string  // external directory override (reserved for future use)
	locale      string  // current locale, e.g. "zh", "en"

	// Cached parsed values
	strings map[string]string
	colors  map[string]string
	dimens  map[string]string
	arrays  map[string][]string
	loaded  bool
}

// NewResourceManager creates a new ResourceManager backed by the given filesystem.
func NewResourceManager(embedded fs.FS) *ResourceManager {
	return &ResourceManager{
		embedded: embedded,
		strings:  make(map[string]string),
		colors:   make(map[string]string),
		dimens:   make(map[string]string),
		arrays:   make(map[string][]string),
	}
}

// SetEmbedded replaces the backing filesystem and invalidates the cache.
func (rm *ResourceManager) SetEmbedded(f fs.FS) {
	rm.embedded = f
	rm.loaded = false
}

// SetLocale sets the current locale and invalidates the cache.
func (rm *ResourceManager) SetLocale(locale string) {
	rm.locale = locale
	rm.loaded = false
}

// SetOverrideDir sets an external override directory (reserved for future use).
func (rm *ResourceManager) SetOverrideDir(dir string) {
	rm.overrideDir = dir
	rm.loaded = false
}

// Embedded returns the backing filesystem.
func (rm *ResourceManager) Embedded() fs.FS {
	return rm.embedded
}

// ensureLoaded lazily loads all resource values on first access.
func (rm *ResourceManager) ensureLoaded() {
	if rm.loaded {
		return
	}
	rm.loaded = true

	// Reset maps
	rm.strings = make(map[string]string)
	rm.colors = make(map[string]string)
	rm.dimens = make(map[string]string)
	rm.arrays = make(map[string][]string)

	if rm.embedded == nil {
		return
	}

	// Load default values first
	rm.loadValuesDir("values")

	// Load locale-specific values to override defaults
	if rm.locale != "" {
		rm.loadValuesDir("values-" + rm.locale)
	}
}

// loadValuesDir loads all resource files from a values directory.
func (rm *ResourceManager) loadValuesDir(dir string) {
	if rm.embedded == nil {
		return
	}

	// Try loading strings.xml
	if data, err := fs.ReadFile(rm.embedded, dir+"/strings.xml"); err == nil {
		if parsed, err := ParseStringsXML(data); err == nil {
			for k, v := range parsed {
				rm.strings[k] = v
			}
		}
	}

	// Try loading colors.xml
	if data, err := fs.ReadFile(rm.embedded, dir+"/colors.xml"); err == nil {
		if parsed, err := ParseColorsXML(data); err == nil {
			for k, v := range parsed {
				rm.colors[k] = v
			}
		}
	}

	// Try loading dimens.xml
	if data, err := fs.ReadFile(rm.embedded, dir+"/dimens.xml"); err == nil {
		if parsed, err := ParseDimensXML(data); err == nil {
			for k, v := range parsed {
				rm.dimens[k] = v
			}
		}
	}

	// Try loading arrays.xml
	if data, err := fs.ReadFile(rm.embedded, dir+"/arrays.xml"); err == nil {
		if parsed, err := ParseStringArraysXML(data); err == nil {
			for k, v := range parsed {
				rm.arrays[k] = v
			}
		}
	}
}

// GetString returns the string resource for the given key.
// Falls back to the key itself if not found.
func (rm *ResourceManager) GetString(key string) string {
	rm.ensureLoaded()
	if v, ok := rm.strings[key]; ok {
		return v
	}
	return key
}

// GetColor returns the parsed color for the given key.
// Returns transparent black if not found.
func (rm *ResourceManager) GetColor(key string) color.RGBA {
	rm.ensureLoaded()
	if v, ok := rm.colors[key]; ok {
		return core.ParseColor(v)
	}
	return color.RGBA{}
}

// GetDimension returns the raw dimension string for the given key.
func (rm *ResourceManager) GetDimension(key string) string {
	rm.ensureLoaded()
	if v, ok := rm.dimens[key]; ok {
		return v
	}
	return ""
}

// GetStringArray returns the string array for the given key.
func (rm *ResourceManager) GetStringArray(key string) []string {
	rm.ensureLoaded()
	if v, ok := rm.arrays[key]; ok {
		return v
	}
	return nil
}

// ResolveRef resolves resource references like @string/key, @color/key, @dimen/key.
// Non-reference strings are returned as-is.
func (rm *ResourceManager) ResolveRef(ref string) string {
	if !strings.HasPrefix(ref, "@") {
		return ref
	}

	rm.ensureLoaded()

	parts := strings.SplitN(ref[1:], "/", 2)
	if len(parts) != 2 {
		return ref
	}

	resType := parts[0]
	resKey := parts[1]

	switch resType {
	case "string":
		if v, ok := rm.strings[resKey]; ok {
			return v
		}
	case "color":
		if v, ok := rm.colors[resKey]; ok {
			return v
		}
	case "dimen":
		if v, ok := rm.dimens[resKey]; ok {
			return v
		}
	}

	return ref
}
