package main

import (
	"flag"
	"github.com/amirrezaask/preditor"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "cfg", path.Join(os.Getenv("HOME"), ".preditor"), "path to config file, defaults to: ~/.preditor")
	flag.Parse()

	// read config file
	cfg, err := preditor.ReadConfig(configPath)
	if err != nil {
		panic(err)
	}

	// create editor
	editor, err := preditor.New(cfg)
	if err != nil {
		panic(err)
	}

	// handle command line argument
	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	} else {
		filename, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	err = preditor.SwitchOrOpenFileInTextBuffer(editor, cfg, filename, editor.OSWindowHeight, editor.OSWindowWidth, rl.Vector2{}, nil)
	if err != nil {
		panic(err)
	}

	// start main loop
	editor.StartMainLoop()

}
