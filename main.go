package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	font rl.Font
	fontSize float32
)

type Command func(Editor) error
type Variables map[string]any
type Key struct {
	Ctrl  bool
	Alt   bool
	Shift bool
	Super bool
	K     string
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

type Window struct {
	BufferIndex int
	zeroLocation rl.Vector2
	Height    int
	Width  int
}

func (e *Editor) RenderBufferInWindow(buffer *Buffer, window *Window) {
	for idx, line := range buffer.Content {
		renderAt := rl.Vector2{
			Y: float32(window.zeroLocation.Y) + measureTextSize(font, ' ', fontSize, 0).Y * float32(idx),
			X: float32(window.zeroLocation.X),
		}
		rl.DrawTextEx(font, string(line), renderAt, fontSize, 0, rl.White)
		cursorSize := measureTextSize(font, buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column], fontSize, 0) //TODO: can be cached
		cursorPosX := int32(buffer.Cursor.Column) * int32(cursorSize.X)
		cursorPosY := int32(buffer.Cursor.Line) * int32(cursorSize.Y)
		rl.DrawRectangleLines(cursorPosX, cursorPosY, int32(cursorSize.X), int32(cursorSize.Y), rl.Yellow)
	}
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
	Buffers           []Buffer
	GlobalKeymaps     []Keymap
	GlobalVariables   Variables
	Commands          Commands
	Windows           []Window
	ActiveWindowIndex int
	LineWrapping      bool
}

func (e Editor) CurrentBuffer() *Buffer {
	return &e.Buffers[e.Windows[e.ActiveWindowIndex].BufferIndex]
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
	editor := Editor{
		LineWrapping : true,
	}

	fontSize = 70
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	editor.Buffers = append(editor.Buffers, Buffer{
		Cursor:   Position{0, 0},
		Content:  [][]byte{[]byte("hello"), []byte("world")},
		FilePath: "test.txt",
	})
	editor.Windows = append(editor.Windows, Window{
		BufferIndex:   0,
		zeroLocation:  rl.Vector2{
			X: 0, Y: 0,
		},
		Height:        rl.GetRenderHeight(),
		Width:         rl.GetRenderWidth(),
	})

	font = rl.LoadFontEx("FiraCode.ttf", int32(fontSize), nil)
	for !rl.WindowShouldClose() {
		buffer := &editor.Buffers[editor.Windows[editor.ActiveWindowIndex].BufferIndex]

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

		editor.RenderBufferInWindow(&editor.Buffers[0], &editor.Windows[0])
		
		rl.EndDrawing()
	}

}


var charSizeCache = map[byte]rl.Vector2{} //TODO: if font size or font changes this is fucked
func measureTextSize(font rl.Font, s byte, size float32, spacing float32) rl.Vector2 {
	if charSize, exists := charSizeCache[s]; exists {
		return charSize
	}
	charSize := rl.MeasureTextEx(font, string(s), size, spacing)
	charSizeCache[s] = charSize
	return charSize
}
