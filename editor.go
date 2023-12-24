package main

import (
	"fmt"
	"time"
	rl "github.com/gen2brain/raylib-go/raylib"

)

type BufferID int64
type Buffer struct {
	Content   []byte
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

func CurrentBuffer(e *Editor) *Buffer {
	return &e.Buffers[e.Windows[e.ActiveWindowIndex].BufferIndex]
}
func CurrentWindow(e *Editor) *Window {
	return &e.Windows[e.ActiveWindowIndex]
}

type Command func(*Editor) error
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

type WindowID int64
type Window struct {
	BufferIndex          int
	zeroLocation         rl.Vector2
	Height               int
	Width                int
	Cursor               Position
	VisualLines          []visualLine
	VisiblePartStartLine int
	VisiblePartEndLine   int
}

type visualLine struct {
	visualLineIndex int
	startIndex      int
	endIndex        int
	ActualLine      int
}

// we are considering fonts to mono spaced,
func RenderBufferInWindow(e *Editor, buffer *Buffer, window *Window) {
	//first scan through buffer.Contents
	// every new line adds a visual line
	// every time we reach windowMaxColumn we add visualLine
	window.VisualLines = []visualLine{}
	charSize := measureTextSize(font, ' ', fontSize, 0)
	totalVisualLines := 0
	lineCharCounter := 0
	var actualLineIndex int
	var start int
	windowMaxColumn := window.Width / int(charSize.X)
	windowMaxLine := window.Height / int(charSize.Y)
	loopStart := time.Now()
	if window.VisiblePartEndLine == 0 {
		// just initialized the editor so we set it to default value which is maximum line number that it can show
		window.VisiblePartEndLine = windowMaxLine
	}
	for idx, char := range buffer.Content {
		lineCharCounter++
		if char == '\n' {
			line := visualLine{
				visualLineIndex: totalVisualLines,
				startIndex:      start,
				endIndex:        idx,
				ActualLine:      actualLineIndex,
			}
			window.VisualLines = append(window.VisualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1

			if visualLineShouldBeRendered(e, window, line) {
				renderVisualLine(e, window, buffer, line)
			}
		}

		if lineCharCounter > windowMaxColumn {
			line := visualLine{
				visualLineIndex: totalVisualLines,
				startIndex:      start,
				endIndex:        idx,
				ActualLine:      actualLineIndex,
			}
			window.VisualLines = append(window.VisualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx
			if visualLineShouldBeRendered(e, window, line) {
				renderVisualLine(e, window, buffer, line)
			}
		}
	}

	fmt.Printf("Render buffer in window: Scan Loop took: %s\n", time.Since(loopStart))
	rl.DrawRectangleLines(int32(window.Cursor.Column)*int32(charSize.X), int32(window.Cursor.Line)*int32(charSize.Y), int32(charSize.X), int32(charSize.Y), rl.White)
}

func cursorToBufferIndex(e *Editor, window *Window, buffer *Buffer) int {
	return window.VisualLines[window.Cursor.Line].startIndex + window.Cursor.Column
}
func visualLineShouldBeRendered(e *Editor, window *Window, line visualLine) bool {
	if window.VisiblePartStartLine <= line.visualLineIndex && line.visualLineIndex <= window.VisiblePartEndLine {
		return true
	}

	return false
}
func renderVisualLine(e *Editor, window *Window, buffer *Buffer, line visualLine) {
	charSize := measureTextSize(font, ' ', fontSize, 0)
	rl.DrawTextEx(font,
		string(buffer.Content[line.startIndex:line.endIndex]),
		rl.Vector2{X: window.zeroLocation.X, Y: float32(line.visualLineIndex) * charSize.Y},
		fontSize,
		0,
		rl.White)

}
func isValidCursorPosition(e *Editor, window *Window, buffer *Buffer, newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if newPosition.Line >= len(window.VisualLines) {
		return false
	}
	if newPosition.Column < 0 {
		return false
	}
	a := window.VisualLines[newPosition.Line].endIndex - window.VisualLines[newPosition.Line].startIndex
	if newPosition.Column > a {
		return false
	}

	return true
}

func InsertCharAtCursor(e *Editor, char byte) error {
	idx := cursorToBufferIndex(e, CurrentWindow(e), CurrentBuffer(e))
	buffer := CurrentBuffer(e)
	buffer.Content = append(buffer.Content[:idx+1], buffer.Content[idx:]...)
	buffer.Content[idx] = char
	window := CurrentWindow(e)
	charSize := measureTextSize(font, ' ', fontSize, 0)

	if char == '\n' {
		window.Cursor.Column = 0
		window.Cursor.Line++
	} else {
		if window.Cursor.Column+1 < (window.Width / int(charSize.X)) {
			window.Cursor.Column = window.Cursor.Column + 1
		}
	}
	return nil
}
