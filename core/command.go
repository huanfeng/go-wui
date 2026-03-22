package core

// KeyBinding associates a key code with modifier keys for shortcuts.
type KeyBinding struct {
	KeyCode  Key
	Modifier KeyModifier
}

// Command represents a named action that can be triggered by shortcut or menu.
type Command struct {
	ID       string
	Title    string
	Shortcut KeyBinding
	Icon     string
	Enabled  bool
	Handler  func()
}

// CommandManager tracks registered commands and dispatches them by ID or shortcut.
type CommandManager struct {
	commands map[string]*Command
}

// NewCommandManager creates an empty CommandManager.
func NewCommandManager() *CommandManager {
	return &CommandManager{commands: make(map[string]*Command)}
}

// Register adds a command to the manager.
func (cm *CommandManager) Register(cmd *Command) {
	cm.commands[cmd.ID] = cmd
}

// Execute runs the handler for the given command ID.
// Returns true if the command was found, enabled, and executed.
func (cm *CommandManager) Execute(id string) bool {
	if cmd, ok := cm.commands[id]; ok && cmd.Enabled && cmd.Handler != nil {
		cmd.Handler()
		return true
	}
	return false
}

// FindByShortcut returns the first enabled command matching the key combination, or nil.
func (cm *CommandManager) FindByShortcut(keyCode Key, modifier KeyModifier) *Command {
	for _, cmd := range cm.commands {
		if cmd.Shortcut.KeyCode == keyCode && cmd.Shortcut.Modifier == modifier && cmd.Enabled {
			return cmd
		}
	}
	return nil
}
