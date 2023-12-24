package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
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
	editor := Editor{
		LineWrapping: true,
		Colors: Colors{
			Background: editorBackground,
			Foreground: editorForeground,
			StatusBarBackground: editorStatusbarBackground,
			StatusBarForeground: editorStatusbarForeground,
		},
	}

	fontSize = 20
	font = rl.LoadFontEx("Consolas.ttf", int32(fontSize), nil)

	rl.SetTextLineSpacing(int(fontSize))
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	textEditorBuffer := &TextEditorBuffer{
		File:	"main.go",
		TabSize: 4,
	}

	textEditorBuffer.Initialize(BufferOptions{
		MaxHeight:	int32(rl.GetRenderHeight()),
		MaxWidth:	 int32(rl.GetRenderWidth()),
		Colors:	   editor.Colors,
		ZeroPosition: rl.Vector2{},
	})
	editor.Buffers = append(editor.Buffers, textEditorBuffer)

	for !rl.WindowShouldClose() {
		editor.HandleWindowResize()
		editor.HandleMouseEvents()
		editor.HandleKeyEvents()
		editor.Render()
	}

}
