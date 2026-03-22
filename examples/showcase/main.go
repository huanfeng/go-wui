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

	resFS, err := fs.Sub(resources, "res")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub FS: %v", err))
	}
	application.SetEmbeddedResources(resFS)

	window, err := application.CreateWindow(platform.WindowOptions{
		Title:     "GoWUI Widget Showcase",
		Width:     500,
		Height:    600,
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

	// Wire up interactive widgets
	wireWidgets(root, window)

	window.Center()
	window.Show()

	// Attach native edits after first render (DPI scaling applied)
	attachNativeEdit := func(id, placeholder string) {
		if v := root.FindViewById(id); v != nil {
			nativeEdit := application.Platform().CreateNativeEditText(window)
			if nativeEdit != nil {
				nativeEdit.AttachToNode(v.Node())
				nativeEdit.SetFont("Segoe UI", 14, 400)
				nativeEdit.SetPlaceholder(placeholder)
			}
		}
	}
	attachNativeEdit("et_name", "Enter your name")
	attachNativeEdit("et_email", "Enter your email")

	application.Run()
}

func wireWidgets(root *core.Node, window platform.Window) {
	// Status text — updated by widget interactions
	var statusTV *widget.TextView
	if v := root.FindViewById("status"); v != nil {
		statusTV, _ = v.(*widget.TextView)
	}
	updateStatus := func(msg string) {
		if statusTV != nil {
			statusTV.SetText(msg)
			window.Invalidate()
		}
		fmt.Println(msg)
	}

	// Buttons
	if v := root.FindViewById("btn_primary"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(_ core.View) {
				updateStatus("Primary button clicked!")
			})
		}
	}
	if v := root.FindViewById("btn_secondary"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(_ core.View) {
				updateStatus("Secondary button clicked!")
			})
		}
	}

	// CheckBox
	if v := root.FindViewById("cb_agree"); v != nil {
		if cb, ok := v.(*widget.CheckBox); ok {
			cb.SetOnCheckedChanged(func(checked bool) {
				if checked {
					updateStatus("CheckBox: Agreed")
				} else {
					updateStatus("CheckBox: Not agreed")
				}
			})
		}
	}

	// Switch
	if v := root.FindViewById("sw_dark"); v != nil {
		if sw, ok := v.(*widget.Switch); ok {
			sw.SetOnChanged(func(on bool) {
				if on {
					updateStatus("Switch: Dark mode ON")
				} else {
					updateStatus("Switch: Dark mode OFF")
				}
			})
		}
	}

	// RadioGroup — register inflated RadioButton children, then set callback
	if v := root.FindViewById("rg_size"); v != nil {
		if rg, ok := v.(*widget.RadioGroup); ok {
			// Children were added by inflater as plain child nodes.
			// We need to register them as managed RadioButtons in the group.
			for _, child := range rg.Node().Children() {
				if rbView := child.GetView(); rbView != nil {
					if rb, ok := rbView.(*widget.RadioButton); ok {
						rg.RegisterButton(rb)
					}
				}
			}
			rg.SetOnChanged(func(idx int) {
				sizes := []string{"Small", "Medium", "Large"}
				if idx >= 0 && idx < len(sizes) {
					updateStatus("Radio: " + sizes[idx] + " selected")
				}
			})
		}
	}

	// ProgressBar — increment on button click
	if v := root.FindViewById("pb_demo"); v != nil {
		if pb, ok := v.(*widget.ProgressBar); ok {
			pb.SetProgress(0.3)
			if btnV := root.FindViewById("btn_progress"); btnV != nil {
				if btn, ok := btnV.(*widget.Button); ok {
					btn.SetOnClickListener(func(_ core.View) {
						p := pb.GetProgress() + 0.1
						if p > 1.0 {
							p = 0.0
						}
						pb.SetProgress(p)
						updateStatus(fmt.Sprintf("Progress: %.0f%%", p*100))
						window.Invalidate()
					})
				}
			}
		}
	}

	// Submit button — read values from native edits
	if v := root.FindViewById("btn_submit"); v != nil {
		if btn, ok := v.(*widget.Button); ok {
			btn.SetOnClickListener(func(_ core.View) {
				updateStatus("Submit clicked!")
			})
		}
	}
}
