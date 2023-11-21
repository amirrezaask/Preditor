package main

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Command func(Editor) error
type Variables map[string]any
type Keymap map[string]Command
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
		var modifiers []string
		switch {
		case rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl):
			modifiers = append(modifiers, "C")
		case rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt):
			modifiers = append(modifiers, "A")
		case rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift):
			modifiers = append(modifiers, "S")
		}
		handleKeyEvent(KeyEvent{MODS:modifiers, Key: rl.GetKeyPressed()}, buffer)
		// fmt.Println("buffer cursor", buffer.Cursor)

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


type KeyEvent struct {
	MODS []string
	Key int32
}


func handleKeyEvent(k KeyEvent, buffer *Buffer) {
	switch k.Key {
	case 0:

	case rl.KeyUp:
		if buffer.Cursor.Line-1 >= 0 {
			buffer.Cursor.Line = buffer.Cursor.Line - 1
		}
	case rl.KeyDown:
		if buffer.Cursor.Line+1 < len(buffer.Content) {
			buffer.Cursor.Line = buffer.Cursor.Line + 1
		}
	case rl.KeyRight:
		if buffer.Cursor.Column+1 < len(buffer.Content[buffer.Cursor.Line]) {
			buffer.Cursor.Column = buffer.Cursor.Column + 1
		}
	case rl.KeyLeft:
		if buffer.Cursor.Column-1 >= 0 {
			buffer.Cursor.Column = buffer.Cursor.Column - 1
		}
	default:
		buffer.Content[buffer.Cursor.Line] = append(buffer.Content[buffer.Cursor.Line][0:buffer.Cursor.Column+1], buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column:]...)
		buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column] = byte(k.Key)
		buffer.Cursor.Column = buffer.Cursor.Column + 1
	}
	
}




func isKeyModifier(k int32) bool {
	return (k >= 256 && k <= 348) ||
		(k == 32) ||
		(k >= 91 && k <= 96)
}
