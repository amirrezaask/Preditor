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
	}

	stat, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}

	if stat.IsDir() {
		editor.Windows = append(editor.Windows, NewFilePicker(&editor, filename, int32(rl.GetRenderHeight()), int32(rl.GetRenderWidth()), rl.Vector2{}, EditorOptions{
			LineNumbers:        true,
			TabSize:            4,
			MaxHeight:          int32(rl.GetRenderHeight()),
			MaxWidth:           int32(rl.GetRenderWidth()),
			Colors:             editor.Colors,
			CursorShape:        cfg.CursorShape,
			CursorBlinking:     cfg.CursorBlinking,
			SyntaxHighlighting: cfg.EnableSyntaxHighlighting,
		}))
	} else {
		e, err := NewEditor(EditorOptions{
			MaxHeight:          int32(rl.GetRenderHeight()),
			MaxWidth:           int32(rl.GetRenderWidth()),
			ZeroPosition:       rl.Vector2{},
			Colors:             editor.Colors,
			Filename:           filename,
			LineNumbers:        true,
			TabSize:            4,
			CursorBlinking:     cfg.CursorBlinking,
			CursorShape:        cfg.CursorShape,
			SyntaxHighlighting: cfg.EnableSyntaxHighlighting,
		})
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
