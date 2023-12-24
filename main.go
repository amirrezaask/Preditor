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
		LineNumbers: true,
		LineWrapping: true,
		Colors: Colors{
			Background:		  editorBackground,
			Foreground:		  editorForeground,
			StatusBarBackground: editorStatusbarBackground,
			StatusBarForeground: editorStatusbarForeground,
			LineNumbersForeground: rl.White,
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


	// initialize first editor
	textEditorBuffer, err := NewTextEditor(TextEditorOptions{
		Filename:	filename,
		LineNumbers: true,
		TabSize: 4,
		MaxHeight:	int32(rl.GetRenderHeight()),
		MaxWidth:	 int32(rl.GetRenderWidth()),
		Colors:	   editor.Colors,
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
