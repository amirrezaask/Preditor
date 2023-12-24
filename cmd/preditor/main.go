package main

import (
	"github.com/amirrezaask/preditor"
)

func main() {
	editor, err := preditor.New()
	if err != nil {
		panic(err)
	}

	// editor.AddWindowInANewColumn(&preditor.Window{
	//	 DrawableID: editor.MessageBufferID,
	// })

	// start main loop
	editor.StartMainLoop()

}
