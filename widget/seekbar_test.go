package widget

import (
	"image/color"
	"testing"
)

func TestNewSeekBar(t *testing.T) {
	sb := NewSeekBar(0.5)
	if sb.Node().Tag() != "SeekBar" {
		t.Errorf("expected tag 'SeekBar', got %q", sb.Node().Tag())
	}
	if sb.GetProgress() != 0.5 {
		t.Errorf("expected progress 0.5, got %f", sb.GetProgress())
	}
}

func TestSeekBarClamp(t *testing.T) {
	sb := NewSeekBar(-0.5)
	if sb.GetProgress() != 0 {
		t.Errorf("expected clamped to 0, got %f", sb.GetProgress())
	}

	sb2 := NewSeekBar(1.5)
	if sb2.GetProgress() != 1 {
		t.Errorf("expected clamped to 1, got %f", sb2.GetProgress())
	}
}

func TestSeekBarSetProgress(t *testing.T) {
	sb := NewSeekBar(0)
	sb.SetProgress(0.75)
	if sb.GetProgress() != 0.75 {
		t.Errorf("expected 0.75, got %f", sb.GetProgress())
	}

	sb.SetProgress(-1)
	if sb.GetProgress() != 0 {
		t.Errorf("expected clamped to 0, got %f", sb.GetProgress())
	}

	sb.SetProgress(2)
	if sb.GetProgress() != 1 {
		t.Errorf("expected clamped to 1, got %f", sb.GetProgress())
	}
}

func TestSeekBarOnChanged(t *testing.T) {
	sb := NewSeekBar(0)
	changedTo := -1.0
	sb.SetOnProgressChangedListener(func(p float64) {
		changedTo = p
	})

	sb.onChanged(0.6)
	if changedTo != 0.6 {
		t.Errorf("expected 0.6, got %f", changedTo)
	}
}

func TestSeekBarColors(t *testing.T) {
	sb := NewSeekBar(0.5)
	red := color.RGBA{R: 255, A: 255}
	gray := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	sb.SetTrackColor(gray)
	sb.SetThumbColor(red)

	if sb.trackColor.R != 200 {
		t.Errorf("expected track R=200, got %d", sb.trackColor.R)
	}
	if sb.thumbColor.R != 255 {
		t.Errorf("expected thumb R=255, got %d", sb.thumbColor.R)
	}
}
