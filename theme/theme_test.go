package theme

import "testing"

func TestThemeResolveAttr(t *testing.T) {
	xmlData := []byte(`<resources><style name="Theme.GoWUI.Light">
		<item name="colorPrimary">#1976D2</item>
		<item name="textColorPrimary">#212121</item>
	</style></resources>`)
	theme, err := LoadThemeFromXML(xmlData, "Theme.GoWUI.Light")
	if err != nil {
		t.Fatal(err)
	}
	if theme.ResolveAttr("colorPrimary") != "#1976D2" {
		t.Error("colorPrimary")
	}
	if theme.ResolveAttr("textColorPrimary") != "#212121" {
		t.Error("textColorPrimary")
	}
}

func TestStyleInheritance(t *testing.T) {
	xmlData := []byte(`<resources>
		<style name="Widget.Button"><item name="textSize">16dp</item><item name="textColor">#FFFFFF</item></style>
		<style name="Widget.Button.Outlined" parent="Widget.Button"><item name="textColor">#1976D2</item></style>
	</resources>`)
	reg, err := LoadStylesFromXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	resolved := reg.Resolve("Widget.Button.Outlined")
	if resolved["textSize"] != "16dp" {
		t.Error("should inherit textSize")
	}
	if resolved["textColor"] != "#1976D2" {
		t.Error("should override textColor")
	}
}

func TestStyleImplicitInheritance(t *testing.T) {
	xmlData := []byte(`<resources>
		<style name="Widget.Button"><item name="textSize">16dp</item></style>
		<style name="Widget.Button.Text"><item name="background">#00000000</item></style>
	</resources>`)
	reg, err := LoadStylesFromXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	resolved := reg.Resolve("Widget.Button.Text")
	if resolved["textSize"] != "16dp" {
		t.Error("should inherit via dot notation")
	}
	if resolved["background"] != "#00000000" {
		t.Error("should have own background")
	}
}

func TestStyleRegistryResolveUnknown(t *testing.T) {
	reg := NewStyleRegistry()
	resolved := reg.Resolve("NonExistent")
	if len(resolved) != 0 {
		t.Error("should return empty map for unknown style")
	}
}

func TestStyleDeepChain(t *testing.T) {
	xmlData := []byte(`<resources>
		<style name="Base"><item name="fontSize">12dp</item></style>
		<style name="Base.Medium"><item name="fontWeight">500</item></style>
		<style name="Base.Medium.Large"><item name="fontSize">24dp</item></style>
	</resources>`)
	reg, err := LoadStylesFromXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	resolved := reg.Resolve("Base.Medium.Large")
	if resolved["fontSize"] != "24dp" {
		t.Error("should override fontSize from Base")
	}
	if resolved["fontWeight"] != "500" {
		t.Error("should inherit fontWeight from Base.Medium")
	}
}

func TestThemeEmptyAttr(t *testing.T) {
	xmlData := []byte(`<resources><style name="Empty"></style></resources>`)
	theme, err := LoadThemeFromXML(xmlData, "Empty")
	if err != nil {
		t.Fatal(err)
	}
	if theme.ResolveAttr("anything") != "" {
		t.Error("should return empty for missing attr")
	}
}
