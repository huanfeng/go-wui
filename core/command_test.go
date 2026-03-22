package core

import "testing"

func TestCommandExecute(t *testing.T) {
	cm := NewCommandManager()
	executed := false
	cm.Register(&Command{
		ID:      "test.action",
		Enabled: true,
		Handler: func() { executed = true },
	})
	if !cm.Execute("test.action") {
		t.Error("should return true")
	}
	if !executed {
		t.Error("handler should have been called")
	}
}

func TestCommandDisabled(t *testing.T) {
	cm := NewCommandManager()
	cm.Register(&Command{
		ID:      "test.disabled",
		Enabled: false,
		Handler: func() { t.Error("should not be called") },
	})
	if cm.Execute("test.disabled") {
		t.Error("disabled command should return false")
	}
}

func TestCommandNotFound(t *testing.T) {
	cm := NewCommandManager()
	if cm.Execute("nonexistent") {
		t.Error("should return false for unknown command")
	}
}

func TestCommandFindByShortcut(t *testing.T) {
	cm := NewCommandManager()
	cm.Register(&Command{
		ID:       "test.copy",
		Enabled:  true,
		Shortcut: KeyBinding{KeyCode: Key(67), Modifier: ModCtrl}, // Ctrl+C
		Handler:  func() {},
	})
	cmd := cm.FindByShortcut(Key(67), ModCtrl)
	if cmd == nil {
		t.Error("should find command by shortcut")
	}
	if cmd.ID != "test.copy" {
		t.Errorf("expected test.copy, got %s", cmd.ID)
	}
}

func TestCommandFindByShortcutDisabled(t *testing.T) {
	cm := NewCommandManager()
	cm.Register(&Command{
		ID:       "test.disabled",
		Enabled:  false,
		Shortcut: KeyBinding{KeyCode: Key(67), Modifier: ModCtrl},
		Handler:  func() {},
	})
	cmd := cm.FindByShortcut(Key(67), ModCtrl)
	if cmd != nil {
		t.Error("should not find disabled command")
	}
}

func TestCommandNilHandler(t *testing.T) {
	cm := NewCommandManager()
	cm.Register(&Command{
		ID:      "test.nohandler",
		Enabled: true,
		Handler: nil,
	})
	if cm.Execute("test.nohandler") {
		t.Error("should return false when handler is nil")
	}
}
