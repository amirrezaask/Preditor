package main

import (
	"github.com/amirrezaask/preditor"
)

func main() {
	setKeyBindings()

	// creates new instance of the editor
	editor, err := preditor.New()
	if err != nil {
		panic(err)
	}

	setKeyBindings()

	// start main loop
	editor.StartMainLoop()

}

// Sample
func setKeyBindings() {
	// preditor.GlobalKeymap.BindKey(preditor.Key{K: "\\", Alt: true}, func(c *preditor.Context) { preditor.VSplit(c) })
	// preditor.GlobalKeymap.BindKey(preditor.Key{K: "=", Alt: true}, func(c *preditor.Context) { preditor.HSplit(c) })
}
