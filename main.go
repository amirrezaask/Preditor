package main

import (

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	font     rl.Font
	fontSize float32
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
	editor.Buffers = append(editor.Buffers, Buffer{
				Content: []byte(loremIpsum),
		FilePath: "test.txt",
	})
	editor.Windows = append(editor.Windows, Window{
		BufferIndex: 0,
		zeroLocation: rl.Vector2{
			X: 0, Y: 0,
		},
		Height: rl.GetRenderHeight(),
		Width:  rl.GetRenderWidth(),
		Cursor: Position{},
	})

	font = rl.LoadFontEx("FiraCode.ttf", int32(fontSize), nil)
	for !rl.WindowShouldClose() {
		buffer := &editor.Buffers[editor.Windows[editor.ActiveWindowIndex].BufferIndex]

		// execute any key command that should be executed
		key := MakeKey(buffer)
		// fmt.Printf("key: %+v\n", key)
		cmd := defaultKeymap[key]
		if cmd != nil {
			if err := cmd(&editor); err != nil {
				panic(err)
			}
		}

		// cmd = defaultKeymap[MakeMouseKey(buffer)]
		// if cmd != nil {
		// 	if err := cmd(editor); cmd != nil {
		// 		panic(err)
		// 	}
		// }

		// Render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		renderBufferOnWindow(&editor, &editor.Buffers[0], &editor.Windows[0])

		rl.EndDrawing()
	}

}
