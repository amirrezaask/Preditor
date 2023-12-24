package main

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Command func(Editor) error
type Variables map[string]any
type Keymap map[string]Command
type Commands map[string]Command
type Vector2 [2]int32

type View struct {
	BufferIndex int
}
type Buffer struct {
	Cursor    Vector2
	Content   []string
	FilePath  string
	Keymaps   []Keymap
	Variables Variables
	Commands  Commands
}

type Editor struct {
	Buffers         []Buffer
	GlobalKeymaps   []Keymap
	GlobalVariables Variables
	Commands        Commands
	Views           []View
	ActiveViewIndex int
}

func main() {
	// basic setup
	rl.InitWindow(1920, 1080, "core editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(100)

	// create editor
	editor := Editor{}

	const textSize = 70
	rl.SetTextLineSpacing(textSize)
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	editor.Buffers = append(editor.Buffers, Buffer{
		Cursor:   Vector2{0, 0},
		Content:  []string{"hello", "world"},
		FilePath: "test.txt",
	})

	editor.Views = append(editor.Views, View{BufferIndex: 0})

	font := rl.LoadFontEx("FiraCode.ttf", 100, nil)
	// gui loop
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		var sb strings.Builder
		buffer := editor.Buffers[editor.Views[editor.ActiveViewIndex].BufferIndex]
		for _, line := range buffer.Content {
			_, _ = sb.WriteString(line)
			_, _ = sb.WriteString("\n")
		}
		// render text
		rl.DrawTextEx(font, sb.String(), rl.Vector2{X: 0, Y: 0}, textSize, 0, rl.White)
		// render cursor
		// we should find cursor size, for that we should measure string size of cursor
		cursorV := rl.MeasureTextEx(font, string(buffer.Content[buffer.Cursor[0]][buffer.Cursor[1]]), textSize, 0)
		rl.DrawRectangleLines(buffer.Cursor[0], buffer.Cursor[1], int32(cursorV.X), int32(cursorV.Y), rl.Red)
		rl.EndDrawing()
	}

}
