package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// basic setup
	rl.InitWindow(1920, 1080, "core editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// create editor
	editor := Editor{
		LineWrapping: true,
	}

	fontSize = 20
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	textEditorBuffer := &TextEditorBuffer{
		Content: []byte(loremIpsum),
		File: "test.txt",
	}

	textEditorBuffer.Initialize(BufferOptions{
		MaxHeight:    int32(rl.GetRenderHeight()),
		MaxWidth:     int32(rl.GetRenderWidth()),
		ZeroPosition: rl.Vector2{
			
		},
	})
	editor.Buffers = append(editor.Buffers, textEditorBuffer)

	font = rl.LoadFontEx("Consolas.ttf", int32(fontSize), nil)
	for !rl.WindowShouldClose() {
		
		key := getKey()
		if !key.IsEmpty() {
			cmd := defaultKeymap[key]
			if cmd != nil {
				cmd(&editor)
			}
		}
		
		// Render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		editor.ActiveBuffer().Render()

		rl.EndDrawing()
	}

}
