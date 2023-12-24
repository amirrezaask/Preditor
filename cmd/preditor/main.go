package main

import (
	"flag"
	"github.com/amirrezaask/preditor"
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
	cfg, err := preditor.ReadConfig(configPath)
	if err != nil {
		panic(err)
	}

	// basic setup
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized)

	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(1920, 1080, "Preditor")

	preditor.InitFileTypes(cfg.Colors)

	if err := clipboard.Init(); err != nil {
		panic(err)
	}
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	// create editor
	editor := preditor.Preditor{
		LineNumbers:  true,
		LineWrapping: true,
		Colors:       cfg.Colors,
	}
	rl.SetTextLineSpacing(int(preditor.FontSize))
	rl.SetExitKey(0)

	err = preditor.LoadFont(cfg.FontName, float32(cfg.FontSize))
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
		editor.Buffers = append(editor.Buffers, preditor.NewFilePickerBuffer(&editor, cfg, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{}))
	} else {
		e, err := preditor.NewTextBuffer(&editor, cfg, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{})
		if err != nil {
			panic(err)
		}
		editor.Buffers = append(editor.Buffers, e)
	}

	for !rl.WindowShouldClose() {
		editor.HandleWindowResize()
		editor.HandleMouseEvents()
		editor.HandleKeyEvents()
		editor.Render()
	}

}
