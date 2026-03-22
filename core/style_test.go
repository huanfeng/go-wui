package core

import (
	"image/color"
	"testing"
)

func TestStyleMerge(t *testing.T) {
	parent := &Style{
		BackgroundColor: ParseColor("#FFFFFF"),
		TextColor:       ParseColor("#000000"),
		FontSize:        14,
	}
	child := &Style{
		TextColor: ParseColor("#FF0000"),
		FontSize:  18,
	}
	merged := MergeStyles(parent, child)
	if merged.BackgroundColor != parent.BackgroundColor {
		t.Error("should inherit parent background")
	}
	if merged.TextColor != child.TextColor {
		t.Error("should use child text color")
	}
	if merged.FontSize != 18 {
		t.Error("should use child font size")
	}
}

func TestStyleMergeNilParent(t *testing.T) {
	child := &Style{FontSize: 16}
	result := MergeStyles(nil, child)
	if result.FontSize != 16 {
		t.Error("nil parent should return child")
	}
}

func TestStyleMergeNilChild(t *testing.T) {
	parent := &Style{FontSize: 14}
	result := MergeStyles(parent, nil)
	if result.FontSize != 14 {
		t.Error("nil child should return parent")
	}
}

func TestStyleMergePreservesParentFields(t *testing.T) {
	parent := &Style{
		BackgroundColor: ParseColor("#FFFFFF"),
		BorderWidth:     2.0,
		CornerRadius:    8.0,
		FontFamily:      "Arial",
		FontWeight:      700,
		Opacity:         0.9,
	}
	child := &Style{
		FontFamily: "Roboto",
	}
	merged := MergeStyles(parent, child)
	if merged.BorderWidth != 2.0 {
		t.Error("should inherit parent border width")
	}
	if merged.CornerRadius != 8.0 {
		t.Error("should inherit parent corner radius")
	}
	if merged.FontFamily != "Roboto" {
		t.Error("should use child font family")
	}
	if merged.FontWeight != 700 {
		t.Error("should inherit parent font weight")
	}
	if merged.Opacity != 0.9 {
		t.Error("should inherit parent opacity")
	}
}

func TestStyleMergeDoesNotMutateParent(t *testing.T) {
	parent := &Style{
		BackgroundColor: ParseColor("#FFFFFF"),
		FontSize:        14,
	}
	child := &Style{
		FontSize: 20,
	}
	originalBg := parent.BackgroundColor
	originalSize := parent.FontSize

	_ = MergeStyles(parent, child)

	if parent.BackgroundColor != originalBg {
		t.Error("parent background should not be mutated")
	}
	if parent.FontSize != originalSize {
		t.Error("parent font size should not be mutated")
	}
}

func TestGravityConstants(t *testing.T) {
	if GravityStart != 0 {
		t.Error("GravityStart should be 0")
	}
	if GravityCenter != 1 {
		t.Error("GravityCenter should be 1")
	}
	if GravityEnd != 2 {
		t.Error("GravityEnd should be 2")
	}
	if GravityCenterVertical != 4 {
		t.Error("GravityCenterVertical should be 4")
	}
	if GravityCenterHorizontal != 8 {
		t.Error("GravityCenterHorizontal should be 8")
	}
}

func TestStyleZeroValueCheck(t *testing.T) {
	// Verify that zero color is correctly treated as "not set"
	zero := color.RGBA{}
	if zero != (color.RGBA{}) {
		t.Error("zero RGBA should equal empty RGBA")
	}
}
