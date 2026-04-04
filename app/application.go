package app

import (
	"io/fs"
	"runtime"

	"github.com/huanfeng/wind-ui/platform"
	pwin "github.com/huanfeng/wind-ui/platform/windows"
	"github.com/huanfeng/wind-ui/res"
	"github.com/huanfeng/wind-ui/theme"
)

// Application is the top-level entry point for a Wind UI application.
// It owns the platform, resource manager, layout inflater, theme, and windows.
type Application struct {
	plat      platform.Platform
	resources *res.ResourceManager
	theme     *theme.Theme
	inflater  *res.LayoutInflater
	windows   []platform.Window
}

// NewApplication creates and initializes a new Application.
// It locks the OS thread (required for Win32 message loop) and sets up the
// resource manager and layout inflater with built-in view factories.
func NewApplication() *Application {
	runtime.LockOSThread()
	app := &Application{}
	app.plat = pwin.NewPlatform()
	app.resources = res.NewResourceManager(nil)
	app.inflater = res.NewLayoutInflater(app.resources)
	res.RegisterBuiltinViews(app.inflater)
	return app
}

// Platform returns the underlying platform implementation.
func (a *Application) Platform() platform.Platform { return a.plat }

// Resources returns the resource manager.
func (a *Application) Resources() *res.ResourceManager { return a.resources }

// Inflater returns the layout inflater.
func (a *Application) Inflater() *res.LayoutInflater { return a.inflater }

// SetTheme sets the application theme.
func (a *Application) SetTheme(t *theme.Theme) { a.theme = t }

// Theme returns the current application theme.
func (a *Application) Theme() *theme.Theme { return a.theme }

// SetEmbeddedResources sets the embedded filesystem for the resource manager.
func (a *Application) SetEmbeddedResources(fsys fs.FS) {
	a.resources.SetEmbedded(fsys)
}

// CreateWindow creates a new platform window with the given options.
func (a *Application) CreateWindow(opts platform.WindowOptions) (platform.Window, error) {
	w, err := a.plat.CreateWindow(opts)
	if err != nil {
		return nil, err
	}
	a.windows = append(a.windows, w)
	return w, nil
}

// Run starts the platform main loop. This call blocks until Quit is called.
func (a *Application) Run() {
	a.plat.RunMainLoop()
}

// Quit signals the platform to exit the main loop.
func (a *Application) Quit() {
	a.plat.Quit()
}
