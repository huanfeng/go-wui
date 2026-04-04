package platform

// TrayIcon represents a system tray (notification area) icon.
type TrayIcon interface {
	// SetIcon sets the tray icon from raw ICO data.
	SetIcon(iconData []byte)
	// SetTooltip sets the hover tooltip text.
	SetTooltip(text string)
	// SetMenu sets the right-click context menu.
	SetMenu(menu *TrayMenu)
	// ShowBalloon displays a balloon notification.
	ShowBalloon(title, message string)
	// SetOnClick registers a callback for single-click.
	SetOnClick(fn func())
	// SetOnDoubleClick registers a callback for double-click.
	SetOnDoubleClick(fn func())
	// Destroy removes the tray icon and releases resources.
	Destroy()
}

// TrayMenu defines a context menu for the tray icon.
type TrayMenu struct {
	Items []TrayMenuItem
}

// TrayMenuItem is a single entry in a tray context menu.
type TrayMenuItem struct {
	ID       string
	Title    string
	Checked  bool
	Disabled bool
	IsSeparator bool
	Children []TrayMenuItem // submenu items
	OnClick  func()
}
