package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var charSizeCache = map[byte]rl.Vector2{} //TODO: if font size or font changes this is fucked
func measureTextSize(font rl.Font, s byte, size float32, spacing float32) rl.Vector2 {
	if charSize, exists := charSizeCache[s]; exists {
		return charSize
	}
	charSize := rl.MeasureTextEx(font, string(s), size, spacing)
	charSizeCache[s] = charSize
	return charSize
}

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
	Index int
	startIndex      int
	endIndex        int
	ActualLine      int
}

// we are considering fonts to mono spaced,
func renderBufferOnWindow(e *Editor, buffer *Buffer, window *Window) {
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
	//loopStart := time.Now()
	if window.VisiblePartEndLine == 0 {
		// just initialized the editor so we set it to default value which is maximum line number that it can show
		window.VisiblePartEndLine = windowMaxLine
	}

	//	fmt.Printf("window visible part: %d %d\n", window.VisiblePartStartLine, window.VisiblePartEndLine)
	for idx, char := range buffer.Content {
		lineCharCounter++
		if char == '\n' {
			line := visualLine{
				Index: totalVisualLines,
				startIndex:      start,
				endIndex:        idx,
				ActualLine:      actualLineIndex,
			}
			window.VisualLines = append(window.VisualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1

		
		}

		if lineCharCounter > windowMaxColumn {
			line := visualLine{
				Index: totalVisualLines,
				startIndex:      start,
				endIndex:        idx,
				ActualLine:      actualLineIndex,
			}
			window.VisualLines = append(window.VisualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx
		
		}
	}
	//	fmt.Printf("Render buffer in window: Scan Loop took: %s\n", time.Since(loopStart))
	//	loopStart = time.Now()
	visibleView := windowVisibleLines(window)
	for idx, line := range visibleView {
		if visualLineShouldBeRendered(e, window, line) {
			renderVisualLine(e, window, buffer, line, idx)
		}
	}
	//fmt.Printf("Render buffer in window: render Loop took: %s\n", time.Since(loopStart))
	rl.DrawRectangleLines(int32(window.Cursor.Column)*int32(charSize.X), int32(window.Cursor.Line)*int32(charSize.Y), int32(charSize.X), int32(charSize.Y), rl.White)
}

func windowVisibleLines(window *Window) []visualLine {
	var visibleView []visualLine

	switch {
	case len(window.VisualLines) > windowMaxLines(font, int(fontSize), window):
		visibleView = window.VisualLines[window.VisiblePartStartLine: window.VisiblePartEndLine]
	default:
		visibleView = window.VisualLines
	}


	return visibleView

}

func cursorToBufferIndex(e *Editor, window *Window, buffer *Buffer) int {
	return windowVisibleLines(window)[window.Cursor.Line].startIndex + window.Cursor.Column
}

func fixCursorColumnIfNeeded(window *Window, newPosition *Position) {
	line := windowVisibleLines(window)[newPosition.Line]
	if newPosition.Column > (line.endIndex - line.startIndex) {
		newPosition.Column = line.endIndex - line.startIndex - 1
	}
	if newPosition.Column < 0 {
		newPosition.Column = 0
	}
}

func visualLineShouldBeRendered(e *Editor, window *Window, line visualLine) bool {
	if window.VisiblePartStartLine <= line.Index && line.Index <= window.VisiblePartEndLine {
		return true
	}

	return false
}

func windowMaxLines(font rl.Font, fontsize int, window *Window) int {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	return window.Height / int(charSize.Y)
}

func renderVisualLine(e *Editor, window *Window, buffer *Buffer, line visualLine, index int) {
	charSize := measureTextSize(font, ' ', fontSize, 0)
	rl.DrawTextEx(font,
		string(buffer.Content[line.startIndex:line.endIndex]),
		rl.Vector2{X: window.zeroLocation.X, Y: float32(index) * charSize.Y},
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Public API
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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

func ScrollUp(e *Editor) error {
	window := CurrentWindow(e)
	if window.VisiblePartStartLine <= 0 {
		return nil
	}
	window.VisiblePartEndLine--
	window.VisiblePartStartLine--
	if window.VisiblePartStartLine < 0 {
		window.VisiblePartStartLine = 0
	}
	fixCursorColumnIfNeeded(window, &window.Cursor)

	return nil

}
func ScrollDown(e *Editor) error {
	window := CurrentWindow(e)
	if window.VisiblePartEndLine >= len(window.VisualLines) {
		return nil
	}
	window.VisiblePartEndLine++
	window.VisiblePartStartLine++
	if window.VisiblePartEndLine >= len(window.VisualLines) {
		window.VisiblePartEndLine = len(window.VisualLines) - 1
	}
	fixCursorColumnIfNeeded(window, &window.Cursor)

	return nil
}

func CursorLeft(e *Editor) error {
	window := CurrentWindow(e)
	buffer := CurrentBuffer(e)
	newPosition := window.Cursor
	newPosition.Column--
	if isValidCursorPosition(e, window, buffer, newPosition) {
		window.Cursor.Column = window.Cursor.Column - 1
	}

	return nil

}
func CursorRight(e *Editor) error {
	window := CurrentWindow(e)
	buffer := CurrentBuffer(e)
	newPosition := window.Cursor
	newPosition.Column++

	if isValidCursorPosition(e, window, buffer, newPosition) {
		window.Cursor.Column = window.Cursor.Column + 1
	}
	return nil
}
func CursorUp(e *Editor) error {
	window := CurrentWindow(e)
	buffer := CurrentBuffer(e)
	newPosition := window.Cursor
	newPosition.Line--

	fixCursorColumnIfNeeded(window, &newPosition)
	if isValidCursorPosition(e, window, buffer, newPosition) {
		window.Cursor = newPosition
	}
	fmt.Println("up")
	fmt.Printf("%+v\n", window.Cursor)

	return nil
}
func CursorDown(e *Editor) error {
	window := CurrentWindow(e)
	buffer := CurrentBuffer(e)
	newPosition := window.Cursor
	newPosition.Line++
	fixCursorColumnIfNeeded(window, &newPosition)
	if isValidCursorPosition(e, window, buffer, newPosition) {
		window.Cursor = newPosition
	}
	fmt.Println("down")
	fmt.Printf("%+v\n", window.Cursor)
	return nil
}
