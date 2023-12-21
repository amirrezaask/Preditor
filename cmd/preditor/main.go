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

func setKeyBindings() {
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "\\", Alt: true}, func(c *preditor.Context) { preditor.VSplit(c) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "=", Alt: true}, func(c *preditor.Context) { preditor.HSplit(c) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: ";", Control: true}, func(c *preditor.Context) { preditor.Compile(c) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "q", Alt: true}, func(c *preditor.Context) { c.CloseWindow(c.ActiveWindowIndex) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "q", Alt: true, Shift: true}, preditor.Exit)
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "0", Control: true}, func(c *preditor.Context) { c.CloseWindow(c.ActiveWindowIndex) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "1", Control: true}, func(c *preditor.Context) { c.BuildWindowToggleState() })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "k", Alt: true}, func(c *preditor.Context) { c.KillDrawable(c.ActiveDrawableID()) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "t", Alt: true}, func(c *preditor.Context) { c.OpenThemesList() })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "o", Control: true}, func(c *preditor.Context) { c.OpenFileList() })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "b", Alt: true}, func(c *preditor.Context) { c.OpenBufferList() })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "<mouse-wheel-down>", Control: true}, func(c *preditor.Context) { c.DecreaseFontSize(2) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "<mouse-wheel-up>", Control: true}, func(c *preditor.Context) { c.IncreaseFontSize(2) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "=", Control: true}, func(c *preditor.Context) { c.IncreaseFontSize(2) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "-", Control: true}, func(c *preditor.Context) { c.DecreaseFontSize(2) })
	preditor.GlobalKeymap.BindKey(preditor.Key{K: "w", Alt: true}, func(c *preditor.Context) { c.OtherWindow() })
}
