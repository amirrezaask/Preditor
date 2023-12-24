package main

import (
	"flag"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
	"os"
	"path"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "cfg", path.Join(os.Getenv("HOME"), ".preditor"), "path to config file, defaults to: ~/.preditor")
	flag.Parse()

	// read config file
	cfg, err := readConfig(configPath)
	if err != nil {
		panic(err)
	}

	// basic setup
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized)

	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(1920, 1080, "Preditor")

	initFileTypes(cfg.Colors)

	if err := clipboard.Init(); err != nil {
		panic(err)
	}
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	// create editor
	editor := Preditor{
		LineNumbers:  true,
		LineWrapping: true,
		Colors:       cfg.Colors,
	}
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetExitKey(0)

	err = loadFont(cfg.FontName, 20)
	if err != nil {
		panic(err)
	}

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
		editor.Windows = append(editor.Windows, NewOpenFileBuffer(&editor, cfg, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{}))
	} else {
		e, err := NewEditor(cfg, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{})
		if err != nil {
			panic(err)
		}
		editor.Windows = append(editor.Windows, e)
	}

	for !rl.WindowShouldClose() {
		editor.HandleWindowResize()
		editor.HandleMouseEvents()
		editor.HandleKeyEvents()
		editor.Render()
	}

}
