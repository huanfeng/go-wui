package widget

import (
	"testing"
)

func TestProgressBarSetProgress(t *testing.T) {
	pb := NewProgressBar()
	if pb.GetProgress() != 0 {
		t.Error("should start at 0")
	}
	pb.SetProgress(0.5)
	if pb.GetProgress() != 0.5 {
		t.Errorf("expected 0.5, got %v", pb.GetProgress())
	}
}

func TestProgressBarClamp(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(1.5)
	if pb.GetProgress() != 1.0 {
		t.Error("should clamp to 1.0")
	}
	pb.SetProgress(-0.5)
	if pb.GetProgress() != 0.0 {
		t.Error("should clamp to 0.0")
	}
}

func TestProgressBarIndeterminate(t *testing.T) {
	pb := NewProgressBar()
	if pb.IsIndeterminate() {
		t.Error("should not be indeterminate by default")
	}
	pb.SetIndeterminate(true)
	if !pb.IsIndeterminate() {
		t.Error("should be indeterminate after SetIndeterminate(true)")
	}
	pb.SetIndeterminate(false)
	if pb.IsIndeterminate() {
		t.Error("should not be indeterminate after SetIndeterminate(false)")
	}
}

func TestProgressBarProgressAtBoundaries(t *testing.T) {
	pb := NewProgressBar()

	pb.SetProgress(0.0)
	if pb.GetProgress() != 0.0 {
		t.Errorf("expected 0.0, got %v", pb.GetProgress())
	}

	pb.SetProgress(1.0)
	if pb.GetProgress() != 1.0 {
		t.Errorf("expected 1.0, got %v", pb.GetProgress())
	}
}
