package main

import (
	"flag"

	"github.com/flopp/go-findfont"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	var fontname string
	flag.StringVar(&fontname, "font", "Consolas", "")
	flag.Parse()
	// basic setup
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized)
	rl.InitWindow(1920, 1080, "editor")

	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	editorBackground, _ := parseHexColor("#062329")
	editorForeground, _ := parseHexColor("#d3b58d")
	editorStatusbarBackground, _ := parseHexColor("#d3b58d")
	editorStatusbarForeground, _ := parseHexColor("#000000")

	// create editor
	editor := Application{
		LineWrapping: true,
		Colors: Colors{
			Background:          editorBackground,
			Foreground:          editorForeground,
			StatusBarBackground: editorStatusbarBackground,
			StatusBarForeground: editorStatusbarForeground,
		},
	}

	fontPath, err := findfont.Find(fontname + ".ttf")
	if err != nil {
		panic(err)
	}

	fontSize = 20
	font = rl.LoadFontEx(fontPath, int32(fontSize), nil)
	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	textEditorBuffer := &TextEditor{
		File:    filename,
		TabSize: 4,
	}

	textEditorBuffer.Initialize(BufferOptions{
		MaxHeight:    int32(rl.GetRenderHeight()),
		MaxWidth:     int32(rl.GetRenderWidth()),
		Colors:       editor.Colors,
		ZeroPosition: rl.Vector2{},
	})
	editor.Editors = append(editor.Editors, textEditorBuffer)

	for !rl.WindowShouldClose() {
		editor.HandleWindowResize()
		editor.HandleMouseEvents()
		editor.HandleKeyEvents()
		editor.Render()
	}

}
