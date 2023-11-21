package main

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Command func(Editor) error
type Variables map[string]any
type Key struct {
	Ctrl bool
	Alt bool
	Shift bool
	Super bool
	K string
}
type Keymap map[Key]Command
type Commands map[string]Command
type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("Line: %d Column:%d\n", p.Line, p.Column)
}

type View struct {
	BufferIndex int
	MaxLines int
	MaxColumns int
}
type Buffer struct {
	Cursor    Position
	Content   [][]byte
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
func (e Editor) CurrentBuffer() *Buffer {
	return &e.Buffers[e.Views[e.ActiveViewIndex].BufferIndex]
}

func (buffer *Buffer) InsertCharAtCursor(char byte) error {
	buffer.Content[buffer.Cursor.Line] = append(buffer.Content[buffer.Cursor.Line][0:buffer.Cursor.Column+1], buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column:]...)
	buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column] = char
	buffer.Cursor.Column = buffer.Cursor.Column + 1

	return nil
}

func main() {
	// basic setup
	rl.InitWindow(1920, 1080, "core editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// create editor
	editor := Editor{}

	const textSize = 70
	rl.SetTextLineSpacing(textSize)
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	editor.Buffers = append(editor.Buffers, Buffer{
		Cursor:   Position{0, 0},
		Content:  [][]byte{[]byte("hello"), []byte("world")},
		FilePath: "test.txt",
	})

	editor.Views = append(editor.Views, View{BufferIndex: 0})

	font := rl.LoadFontEx("FiraCode.ttf", 100, nil)
	for !rl.WindowShouldClose() {
		buffer := &editor.Buffers[editor.Views[editor.ActiveViewIndex].BufferIndex]

		// execute any command that should be executed
		cmd := defaultKeymap[MakeKey(buffer)]
		if cmd != nil {
			if err := cmd(editor); err != nil {
				panic(err)
			}
		}

		// Render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		var sb strings.Builder
		for _, line := range buffer.Content {
			_, _ = sb.Write(line)
			_, _ = sb.WriteString("\n")
		}

		// render text
		rl.DrawTextEx(font, sb.String(), rl.Vector2{X: 0, Y: 0}, textSize, 0, rl.White)
		// render cursor
		// we should find cursor size, for that we should measure string size of cursor
		cursorSize := rl.MeasureTextEx(font, string(buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column]), textSize, 0) //TODO: can be cached
		cursorPosX := int32(buffer.Cursor.Column) * int32(cursorSize.X)
		cursorPosY := int32(buffer.Cursor.Line) * int32(cursorSize.Y)
		rl.DrawRectangleLines(cursorPosX, cursorPosY, int32(cursorSize.X), int32(cursorSize.Y), rl.Yellow)
		rl.EndDrawing()
	}

}
