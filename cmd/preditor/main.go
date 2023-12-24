package main

import (
	"github.com/amirrezaask/preditor"
)

func main() {
	editor, err := preditor.New()
	if err != nil {
		panic(err)
	}

	// start main loop
	editor.StartMainLoop()

}
