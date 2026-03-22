package main

import (
	"embed"
	"fmt"
	"io/fs"

	"gowui/app"
	"gowui/core"
	"gowui/platform"
	"gowui/widget"
)

//go:embed res
var resources embed.FS

func main() {
	application := app.NewApplication()

	// The embed FS has paths like "res/layout/main.xml", but the ResourceManager
	// and LayoutInflater expect paths without the "res/" prefix (e.g. "layout/main.xml").
	// Use fs.Sub to strip the "res/" prefix.
	resFS, err := fs.Sub(resources, "res")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub FS: %v", err))
	}
	application.SetEmbeddedResources(resFS)

	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "Hello GoWUI",
		Width:     400,
		Height:    300,
		Resizable: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create window: %v", err))
	}

	root := application.Inflater().Inflate("@layout/main")
	if root == nil {
		panic("failed to inflate layout")
	}
	window.SetContentView(root)

	// Wire up button clicks via FindViewById
	if v := root.FindViewById("btn_ok"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(view core.View) {
				fmt.Println("OK clicked!")
				// Update title text to demonstrate dynamic UI
				if tv := root.FindViewById("title"); tv != nil {
					if title, ok := tv.(*widget.TextView); ok {
						title.SetText("Button Clicked!")
					}
				}
				window.Invalidate()
			})
		}
	}

	if v := root.FindViewById("btn_cancel"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(view core.View) {
				fmt.Println("Cancel clicked - closing window")
				window.Close()
			})
		}
	}

	window.Center() // Center before Show to avoid position flicker
	window.Show()
	application.Run()
}
