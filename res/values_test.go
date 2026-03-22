package res

import "testing"

func TestParseStringsXML(t *testing.T) {
	xmlData := []byte(`<resources>
		<string name="app_title">My App</string>
		<string name="greeting">Hello, %1$s!</string>
	</resources>`)
	strings, err := ParseStringsXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if strings["app_title"] != "My App" {
		t.Error("app_title mismatch")
	}
	if strings["greeting"] != "Hello, %1$s!" {
		t.Error("greeting mismatch")
	}
}

func TestParseColorsXML(t *testing.T) {
	xmlData := []byte(`<resources>
		<color name="primary">#1976D2</color>
		<color name="background">#FFFFFF</color>
	</resources>`)
	colors, err := ParseColorsXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if colors["primary"] != "#1976D2" {
		t.Error("primary mismatch")
	}
}

func TestParseDimensXML(t *testing.T) {
	xmlData := []byte(`<resources>
		<dimen name="text_size">16dp</dimen>
		<dimen name="padding">8dp</dimen>
	</resources>`)
	dimens, err := ParseDimensXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if dimens["text_size"] != "16dp" {
		t.Error("text_size mismatch")
	}
	if dimens["padding"] != "8dp" {
		t.Error("padding mismatch")
	}
}

func TestParseStringArraysXML(t *testing.T) {
	xmlData := []byte(`<resources>
		<string-array name="fruits"><item>Apple</item><item>Banana</item></string-array>
	</resources>`)
	arrays, err := ParseStringArraysXML(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if len(arrays["fruits"]) != 2 {
		t.Error("expected 2 items")
	}
	if arrays["fruits"][0] != "Apple" {
		t.Error("first item should be Apple")
	}
	if arrays["fruits"][1] != "Banana" {
		t.Error("second item should be Banana")
	}
}
