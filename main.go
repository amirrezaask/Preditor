package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	font     rl.Font
	fontSize float32
)
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
func (e Editor) CurrentWindow() *Window {
	return &e.Windows[e.ActiveWindowIndex]
}

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
	BufferIndex  int
	zeroLocation rl.Vector2
	Height       int
	Width        int
	Cursor       Position
}
type visualLine struct {
	visualLineIndex int
	startIndex      int
	endIndex        int
	ActualLine      int
}

/*


*/



// we are considering fonts to mono spaced,
func (e *Editor) RenderBufferInWindow(buffer *Buffer, window *Window) {
	var windowLines []visualLine // actual line -> [] visual line ()
	cursorSize := measureTextSize(font, buffer.Content[buffer.Cursor.Line][buffer.Cursor.Column], fontSize, 0)
	windowMaxColumn := float32(window.Width) / cursorSize.X
	var renderAt rl.Vector2

	// calculate visual lines
	var totalVisualLines int
	for idx, line := range buffer.Content {
		if float32(len(line)) > windowMaxColumn {
			// line should be wrapped(splitted into multiple visual lines)
			var start int
			end := int(windowMaxColumn)
			var visualLines []visualLine
			for start < end {
				visualLines = append(visualLines, visualLine{
					visualLineIndex: totalVisualLines,
					ActualLine:      idx,
					startIndex:      start,
					endIndex:        end,
				})
				totalVisualLines++

				start = end + 1
				end += int(windowMaxColumn)
				if end > len(line) {
					end = len(line)
				}
			}

			windowLines = append(windowLines, visualLines...)

		} else {
			windowLines = append(windowLines, []visualLine{
				{ActualLine: idx, startIndex: 0, endIndex: len(line), visualLineIndex: totalVisualLines},
			}...)
			totalVisualLines++

		}
	}

	for _, line := range windowLines {
		fmt.Printf("%+v\n", line)
		renderAt.Y = float32(line.visualLineIndex) * cursorSize.Y
		rl.DrawTextEx(font, string(buffer.Content[line.ActualLine][line.startIndex:line.endIndex]), renderAt, fontSize, 0, rl.White)
	}

	for _, visualLine := range windowLines {
		if visualLine.ActualLine == buffer.Cursor.Line {
			if visualLine.startIndex <= buffer.Cursor.Column && visualLine.endIndex >= buffer.Cursor.Column {
				// this is the visual line we need the cursor to be rendered\
				column := int32(buffer.Cursor.Column)
				if float32(buffer.Cursor.Column) >= windowMaxColumn {
					column = int32(buffer.Cursor.Column) - int32(windowMaxColumn) - 1
				}
				rl.DrawRectangleLines(int32(column)*int32(cursorSize.X), int32(visualLine.visualLineIndex)*int32(cursorSize.Y), int32(cursorSize.X), int32(cursorSize.Y), rl.White)
			}
		}

	}
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
		LineWrapping: true,
	}

	fontSize = 70
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	editor.Buffers = append(editor.Buffers, Buffer{
		Cursor:   Position{0, 0},
		Content:  [][]byte{[]byte("helloooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo"), []byte("world")},
		FilePath: "test.txt",
	})
	editor.Windows = append(editor.Windows, Window{
		BufferIndex: 0,
		zeroLocation: rl.Vector2{
			X: 0, Y: 0,
		},
		Height: rl.GetRenderHeight(),
		Width:  rl.GetRenderWidth(),
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
