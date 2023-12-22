package main

import (
	"github.com/amirrezaask/preditor"
)

func main() {
	setKeyBindings()

	editor, err := preditor.New()
	if err != nil {
		panic(err)
	}

	// start main loop
	editor.StartMainLoop()

}


// Sample
func setKeyBindings() {
// 	preditor.GlobalKeymap.BindKey(preditor.Key{K: "\\", Alt: true}, func(c *preditor.Context) { preditor.VSplit(c) })
// 	preditor.GlobalKeymap.BindKey(preditor.Key{K: "=", Alt: true}, func(c *preditor.Context) { preditor.HSplit(c) })
}
