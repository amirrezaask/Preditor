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

	stat, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}

	if stat.IsDir() {
		editor.Buffers = append(editor.Buffers, preditor.NewFilePickerBuffer(editor, cfg, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{}))
	} else {
		err := preditor.SwitchOrOpenFileInTextBuffer(editor, cfg, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{}, nil)
		if err != nil {
			panic(err)
		}
	}

	// start main loop
	editor.StartMainLoop()

}
