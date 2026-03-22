package res

import (
	"testing"
	"testing/fstest"
)

func TestResourceManagerGetString(t *testing.T) {
	f := fstest.MapFS{
		"values/strings.xml":    &fstest.MapFile{Data: []byte(`<resources><string name="ok">OK</string></resources>`)},
		"values-zh/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="ok">确定</string></resources>`)},
	}
	rm := NewResourceManager(f)
	rm.SetLocale("en")
	if rm.GetString("ok") != "OK" {
		t.Error("expected English")
	}
	rm.SetLocale("zh")
	if rm.GetString("ok") != "确定" {
		t.Error("expected Chinese")
	}
}

func TestResourceManagerGetStringFallback(t *testing.T) {
	f := fstest.MapFS{
		"values/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="ok">OK</string></resources>`)},
	}
	rm := NewResourceManager(f)
	// Key not found should return the key itself
	if rm.GetString("missing") != "missing" {
		t.Error("expected fallback to key")
	}
}

func TestResourceManagerResolveRef(t *testing.T) {
	f := fstest.MapFS{
		"values/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="title">Hello</string></resources>`)},
		"values/colors.xml":  &fstest.MapFile{Data: []byte(`<resources><color name="primary">#1976D2</color></resources>`)},
		"values/dimens.xml":  &fstest.MapFile{Data: []byte(`<resources><dimen name="padding">16dp</dimen></resources>`)},
	}
	rm := NewResourceManager(f)
	if rm.ResolveRef("@string/title") != "Hello" {
		t.Error("string ref")
	}
	if rm.ResolveRef("@color/primary") != "#1976D2" {
		t.Error("color ref")
	}
	if rm.ResolveRef("@dimen/padding") != "16dp" {
		t.Error("dimen ref")
	}
	if rm.ResolveRef("plain text") != "plain text" {
		t.Error("non-ref should pass through")
	}
}

func TestResourceManagerGetColor(t *testing.T) {
	f := fstest.MapFS{
		"values/colors.xml": &fstest.MapFile{Data: []byte(`<resources><color name="red">#FF0000</color></resources>`)},
	}
	rm := NewResourceManager(f)
	c := rm.GetColor("red")
	if c.R != 0xFF || c.G != 0x00 || c.B != 0x00 || c.A != 0xFF {
		t.Errorf("unexpected color: %+v", c)
	}
}

func TestResourceManagerGetDimension(t *testing.T) {
	f := fstest.MapFS{
		"values/dimens.xml": &fstest.MapFile{Data: []byte(`<resources><dimen name="margin">8dp</dimen></resources>`)},
	}
	rm := NewResourceManager(f)
	if rm.GetDimension("margin") != "8dp" {
		t.Error("expected 8dp")
	}
	if rm.GetDimension("missing") != "" {
		t.Error("expected empty for missing key")
	}
}

func TestResourceManagerGetStringArray(t *testing.T) {
	f := fstest.MapFS{
		"values/arrays.xml": &fstest.MapFile{Data: []byte(`<resources><string-array name="items"><item>A</item><item>B</item></string-array></resources>`)},
	}
	rm := NewResourceManager(f)
	arr := rm.GetStringArray("items")
	if len(arr) != 2 {
		t.Fatalf("expected 2 items, got %d", len(arr))
	}
	if arr[0] != "A" || arr[1] != "B" {
		t.Error("unexpected array values")
	}
}
