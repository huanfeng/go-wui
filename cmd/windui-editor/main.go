package main

import (
	"os"

	"github.com/huanfeng/wind-ui/editor"
)

func main() {
	// Usage: windui-editor [--screenshot output.png]
	var screenshot string
	for i, arg := range os.Args[1:] {
		if arg == "--screenshot" && i+1 < len(os.Args[1:]) {
			screenshot = os.Args[i+2]
		}
	}
	editor.Run(screenshot)
}
