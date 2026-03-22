package windows

import "testing"

func TestDpToPx(t *testing.T) {
	tests := []struct {
		dp, dpi, expected float64
	}{
		{16, 96, 16},
		{16, 144, 24},
		{16, 192, 32},
		{0, 96, 0},
		{100, 96, 100},
	}
	for _, tt := range tests {
		if got := DpToPx(tt.dp, tt.dpi); got != tt.expected {
			t.Errorf("DpToPx(%v, %v) = %v, want %v", tt.dp, tt.dpi, got, tt.expected)
		}
	}
}

func TestPxToDp(t *testing.T) {
	tests := []struct {
		px, dpi, expected float64
	}{
		{16, 96, 16},
		{24, 144, 16},
		{32, 192, 16},
		{0, 96, 0},
	}
	for _, tt := range tests {
		if got := PxToDp(tt.px, tt.dpi); got != tt.expected {
			t.Errorf("PxToDp(%v, %v) = %v, want %v", tt.px, tt.dpi, got, tt.expected)
		}
	}
}
