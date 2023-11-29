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

	err = loadFont(cfg.FontName, 20)
	if err != nil {
		panic(err)
	}

	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetExitKey(0)

	// initialize first editor
	textEditorBuffer, err := NewEditor(EditorOptions{
		Filename:           filename,
		LineNumbers:        true,
		TabSize:            4,
		MaxHeight:          int32(rl.GetRenderHeight()),
		MaxWidth:           int32(rl.GetRenderWidth()),
		Colors:             editor.Colors,
		CursorShape:        cfg.CursorShape,
		CursorBlinking:     cfg.CursorBlinking,
		SyntaxHighlighting: cfg.EnableSyntaxHighlighting,
	})
	if err != nil {
		panic(err)
	}

	editor.Windows = append(editor.Windows, textEditorBuffer)

	for !rl.WindowShouldClose() {
		editor.HandleWindowResize()
		editor.HandleMouseEvents()
		editor.HandleKeyEvents()
		editor.Render()
	}

}
