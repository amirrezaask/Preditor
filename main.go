package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Command 
type Variables map[string]any
type Keymap map[string]Command

type Buffer struct {
	Content []string
	FilePath string
	Keymaps []Keymap
	Variables []Variables
}

type Editor struct {
	Buffers []Buffer
}

func main() {
	// basic setup
	rl.InitWindow(1920, 1080, "core editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(100)

	font := rl.LoadFontEx("FiraCode.ttf", 100, nil)
	// gui loop
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawTextEx(font, "Core Editor Test!", rl.Vector2{X:10, Y: 10}, 100, 0, rl.White)
		rl.EndDrawing()
	}
	
}
