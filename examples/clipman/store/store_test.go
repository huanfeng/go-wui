package store

import (
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestAddAndGetAll(t *testing.T) {
	s := tempStore(t)

	s.Add("hello", "notepad")
	s.Add("world", "chrome")

	all := s.GetAll()
	if len(all) != 2 {
		t.Fatalf("got %d entries, want 2", len(all))
	}
	// Newest first
	if all[0].Text != "world" {
		t.Errorf("first entry text = %q, want %q", all[0].Text, "world")
	}
	if all[1].Text != "hello" {
		t.Errorf("second entry text = %q, want %q", all[1].Text, "hello")
	}
}

func TestDeduplication(t *testing.T) {
	s := tempStore(t)

	s.Add("hello", "notepad")
	s.Add("world", "chrome")
	s.Add("hello", "vscode") // duplicate text, should move to top

	all := s.GetAll()
	if len(all) != 2 {
		t.Fatalf("got %d entries, want 2 (dedup)", len(all))
	}
	if all[0].Text != "hello" {
		t.Errorf("first entry = %q, want %q", all[0].Text, "hello")
	}
	if all[0].Source != "vscode" {
		t.Errorf("source = %q, want %q", all[0].Source, "vscode")
	}
	if all[0].UsedCount != 1 {
		t.Errorf("used_count = %d, want 1", all[0].UsedCount)
	}
}

func TestDelete(t *testing.T) {
	s := tempStore(t)

	s.Add("hello", "notepad")
	entry := s.Add("world", "chrome")

	if !s.Delete(entry.ID) {
		t.Fatal("Delete returned false")
	}
	if s.Count() != 1 {
		t.Fatalf("count = %d, want 1", s.Count())
	}
}

func TestPinAndGetPinned(t *testing.T) {
	s := tempStore(t)

	e1 := s.Add("hello", "notepad")
	s.Add("world", "chrome")

	s.Pin(e1.ID, true)

	pinned := s.GetPinned()
	if len(pinned) != 1 {
		t.Fatalf("got %d pinned, want 1", len(pinned))
	}
	if pinned[0].Text != "hello" {
		t.Errorf("pinned text = %q, want %q", pinned[0].Text, "hello")
	}
}

func TestCategory(t *testing.T) {
	s := tempStore(t)

	e1 := s.Add("code snippet", "vscode")
	e2 := s.Add("url link", "chrome")

	s.SetCategory(e1.ID, "code")
	s.SetCategory(e2.ID, "links")

	cats := s.Categories()
	if len(cats) != 2 {
		t.Fatalf("got %d categories, want 2", len(cats))
	}

	byCode := s.GetByCategory("code")
	if len(byCode) != 1 {
		t.Fatalf("got %d entries for 'code', want 1", len(byCode))
	}
}

func TestSearch(t *testing.T) {
	s := tempStore(t)

	s.Add("hello world", "notepad")
	s.Add("foo bar", "chrome")
	s.Add("hello again", "vscode")

	results := s.Search("hello")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestGetRecent(t *testing.T) {
	s := tempStore(t)

	s.Add("a", "app")
	s.Add("b", "app")
	s.Add("c", "app")

	recent := s.GetRecent(2)
	if len(recent) != 2 {
		t.Fatalf("got %d, want 2", len(recent))
	}
	if recent[0].Text != "c" {
		t.Errorf("first = %q, want %q", recent[0].Text, "c")
	}
}

func TestUse(t *testing.T) {
	s := tempStore(t)

	s.Add("a", "app")
	e := s.Add("b", "app")
	s.Add("c", "app")

	// Use "b" — should move to top
	used := s.Use(e.ID)
	if used == nil {
		t.Fatal("Use returned nil")
	}

	all := s.GetAll()
	if all[0].Text != "b" {
		t.Errorf("first = %q, want %q after Use", all[0].Text, "b")
	}
	if all[0].UsedCount != 1 {
		t.Errorf("used_count = %d, want 1", all[0].UsedCount)
	}
}

func TestClearAll(t *testing.T) {
	s := tempStore(t)

	e := s.Add("keep me", "app")
	s.Pin(e.ID, true)
	s.Add("delete me", "app")

	s.ClearAll()

	all := s.GetAll()
	if len(all) != 1 {
		t.Fatalf("got %d, want 1 (only pinned)", len(all))
	}
	if all[0].Text != "keep me" {
		t.Errorf("surviving entry = %q, want %q", all[0].Text, "keep me")
	}
}

func TestCleanup(t *testing.T) {
	s := tempStore(t)
	s.config.MaxEntries = 3

	s.Add("a", "app")
	s.Add("b", "app")
	s.Add("c", "app")
	s.Add("d", "app") // should trigger cleanup, "a" removed

	if s.Count() != 3 {
		t.Fatalf("count = %d, want 3 after cleanup", s.Count())
	}

	all := s.GetAll()
	for _, e := range all {
		if e.Text == "a" {
			t.Error("oldest entry 'a' should have been cleaned up")
		}
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()

	// Create store and add entries
	s1, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	e := s1.Add("persist me", "app")
	s1.Pin(e.ID, true)
	s1.SetCategory(e.ID, "test")
	s1.SaveNow()

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "history.json")); err != nil {
		t.Fatalf("history.json not found: %v", err)
	}

	// Create new store from same directory — should restore
	s2, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	all := s2.GetAll()
	if len(all) != 1 {
		t.Fatalf("restored %d entries, want 1", len(all))
	}
	if all[0].Text != "persist me" {
		t.Errorf("text = %q, want %q", all[0].Text, "persist me")
	}
	if !all[0].Pinned {
		t.Error("pinned not restored")
	}
	if all[0].Category != "test" {
		t.Errorf("category = %q, want %q", all[0].Category, "test")
	}
}

func TestConfigPersistence(t *testing.T) {
	dir := t.TempDir()

	s1, _ := New(dir)
	cfg := s1.GetConfig()
	cfg.MaxEntries = 100
	s1.SetConfig(cfg)

	s2, _ := New(dir)
	if s2.GetConfig().MaxEntries != 100 {
		t.Errorf("max_entries = %d, want 100", s2.GetConfig().MaxEntries)
	}
}
