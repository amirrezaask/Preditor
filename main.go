package main

import (
	"flag"
	"os"
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
)

func main() {
	var fontname string
	var configPath string
	flag.StringVar(&fontname, "font", "Consolas", "")
	flag.StringVar(&configPath, "cfg", path.Join(os.Getenv("HOME"), ".core"), "path to config file, defaults to: ~/.core.cfg")
	flag.Parse()
	// basic setup
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized)
	rl.InitWindow(1920, 1080, "editor")

	if err := clipboard.Init(); err != nil {
		panic(err)
	}
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	editorBackground, _ := parseHexColor("#333333")
	editorForeground, _ := parseHexColor("#ffffff")
	editorStatusbarBackground, _ := parseHexColor("#d3b58d")
	editorStatusbarForeground, _ := parseHexColor("#000000")

	// create editor
	editor := Application{
		LineNumbers:  true,
		LineWrapping: true,
		Colors: Colors{
			Background:            editorBackground,
			Foreground:            editorForeground,
			StatusBarBackground:   editorStatusbarBackground,
			StatusBarForeground:   editorStatusbarForeground,
			LineNumbersForeground: rl.White,
		},
	}

	var err error
	err = loadFont(fontname, 20)
	if err != nil {
		panic(err)
	}

	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}
	rl.SetTextLineSpacing(int(fontSize))

	// initialize first editor
	textEditorBuffer, err := NewEditorBuffer(EditorBufferOptions{
		Filename:       filename,
		LineNumbers:    true,
		TabSize:        4,
		MaxHeight:      int32(rl.GetRenderHeight()),
		MaxWidth:       int32(rl.GetRenderWidth()),
		Colors:         editor.Colors,
		CursorBlinking: false,
	})
	if err != nil {
		panic(err)
	}

	editor.Editors = append(editor.Editors, textEditorBuffer)

	for !rl.WindowShouldClose() {
		editor.HandleWindowResize()
		editor.HandleMouseEvents()
		editor.HandleKeyEvents()
		editor.Render()
	}

}
