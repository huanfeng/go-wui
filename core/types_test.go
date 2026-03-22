package core

import "testing"

func TestRectContains(t *testing.T) {
	r := Rect{X: 10, Y: 10, Width: 100, Height: 50}
	if !r.Contains(50, 30) {
		t.Error("point inside should be contained")
	}
	if r.Contains(5, 5) {
		t.Error("point outside should not be contained")
	}
}

func TestRectIntersect(t *testing.T) {
	a := Rect{X: 0, Y: 0, Width: 100, Height: 100}
	b := Rect{X: 50, Y: 50, Width: 100, Height: 100}
	inter := a.Intersect(b)
	if inter.Width != 50 || inter.Height != 50 {
		t.Errorf("unexpected: %v", inter)
	}
}

func TestInsetsApply(t *testing.T) {
	r := Rect{X: 0, Y: 0, Width: 100, Height: 100}
	insets := Insets{Left: 10, Top: 10, Right: 10, Bottom: 10}
	inner := r.ApplyInsets(insets)
	if inner.Width != 80 || inner.Height != 80 {
		t.Errorf("unexpected: %v", inner)
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input    string
		r, g, b, a uint8
	}{
		{"#FF5722", 0xFF, 0x57, 0x22, 0xFF},
		{"#80FF5722", 0xFF, 0x57, 0x22, 0x80},
	}
	for _, tt := range tests {
		c := ParseColor(tt.input)
		r, g, b, a := c.RGBA()
		if uint8(r>>8) != tt.r || uint8(g>>8) != tt.g || uint8(b>>8) != tt.b || uint8(a>>8) != tt.a {
			t.Errorf("ParseColor(%q) unexpected", tt.input)
		}
	}
}

func TestDimensionParse(t *testing.T) {
	tests := []struct {
		input string
		unit  DimensionUnit
		value float64
	}{
		{"200dp", DimensionDp, 200},
		{"100px", DimensionPx, 100},
		{"match_parent", DimensionMatchParent, 0},
		{"wrap_content", DimensionWrapContent, 0},
	}
	for _, tt := range tests {
		d := ParseDimension(tt.input)
		if d.Unit != tt.unit || d.Value != tt.value {
			t.Errorf("ParseDimension(%q) = %+v", tt.input, d)
		}
	}
}
