package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type TextEditorBuffer struct {
	File         string
	Content      []byte
	Keymaps      []Keymap
	Variables    Variables
	Commands     Commands
	MaxHeight    int32
	MaxWidth     int32
	ZeroPosition rl.Vector2
	TabSize      int
	VisibleStart int32
	VisibleEnd   int32
	visualLines  []visualLine
	Cursor       Position
	maxLine      int32
	maxColumn    int32
	Colors       Colors
}

func (t *TextEditorBuffer) replaceTabsWithSpaces() {
	t.Content = bytes.Replace(t.Content, []byte("\t"), []byte(strings.Repeat(" ", 4)), -1)
}

func (t *TextEditorBuffer) SetMaxWidth(w int32) {
	t.MaxWidth = w
	t.updateMaxLineAndColumn()
}
func (t *TextEditorBuffer) SetMaxHeight(h int32) {
	t.MaxHeight = h
	t.updateMaxLineAndColumn()
}
func (t *TextEditorBuffer) updateMaxLineAndColumn() {
	oldMaxLine := t.maxLine
	charSize := measureTextSize(font, ' ', fontSize, 0)
	t.maxColumn = t.MaxWidth / int32(charSize.X)
	t.maxLine = t.MaxHeight / int32(charSize.Y)

	// reserve one line for status bar
	t.maxLine--

	diff := t.maxLine - oldMaxLine
	t.VisibleEnd += diff
}
func (t *TextEditorBuffer) Type() string {
	return "text_editor_buffer"
}

func (t *TextEditorBuffer) Initialize(opts BufferOptions) error {
	t.MaxHeight = opts.MaxHeight
	t.MaxWidth = opts.MaxWidth
	t.ZeroPosition = opts.ZeroPosition
	t.Colors = opts.Colors
	var err error
	if t.File != "" {
		t.Content, err = os.ReadFile(t.File)
		if err != nil {
			return err
		}
	} else {
		t.Content = make([]byte, 1000)
	}
	t.replaceTabsWithSpaces()
	t.updateMaxLineAndColumn()
	return nil
}

func (t *TextEditorBuffer) Destroy() error {
	return nil
}

type visualLine struct {
	Index      int
	startIndex int
	endIndex   int
	ActualLine int
	Length     int
}

func (t *TextEditorBuffer) Render() {
	t.visualLines = []visualLine{}
	charSize := measureTextSize(font, ' ', fontSize, 0)
	totalVisualLines := 0
	lineCharCounter := 0
	var actualLineIndex int
	var start int
	if t.VisibleEnd == 0 {
		t.VisibleEnd = t.maxLine
	}
	for idx, char := range t.Content {
		lineCharCounter++
		if char == '\n' {
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start - 1,
				ActualLine: actualLineIndex,
			}
			if line.Length < 0 {
				line.Length = 0
			}

			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1

		}
		if idx >= len(t.Content)-1 {
			// last index
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start - 1,
				ActualLine: actualLineIndex,
			}
			if line.Length < 0 {
				line.Length = 0
			}
			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1

		}

		if int32(lineCharCounter) > t.maxColumn {
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start - 1,
				ActualLine: actualLineIndex,
			}
			if line.Length < 0 {
				line.Length = 0
			}

			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx

		}
	}
	// fmt.Printf("Render buffer in window: Scan Loop took: %s\n", time.Since(loopStart))
	// loopStart = time.Now()
	var visibleLines []visualLine
	if t.VisibleEnd > int32(len(t.visualLines)) {
		visibleLines = t.visualLines[t.VisibleStart:]
	} else {
		visibleLines = t.visualLines[t.VisibleStart:t.VisibleEnd]
	}
	for idx, line := range visibleLines {
		if t.visualLineShouldBeRendered(line) {
			t.renderVisualLine(line, idx)
		}
	}
	// fmt.Printf("Rendering buffer: render Loop took: %s\n", time.Since(loopStart))
	cursorView := Position{
		Line:   t.Cursor.Line - int(t.VisibleStart),
		Column: t.Cursor.Column,
	}
	rl.DrawRectangleLines(int32(cursorView.Column)*int32(charSize.X), int32(cursorView.Line)*int32(charSize.Y), int32(charSize.X), int32(charSize.Y), rl.White)

	//render status bar
	rl.DrawRectangle(
		int32(t.ZeroPosition.X),
		int32(float32(t.maxLine)*charSize.Y),
		t.MaxWidth,
		int32(charSize.Y),
		t.Colors.StatusBarBackground,
	)
	file := t.File
	if file == "" {
		file = "*scratch*"
	}
	rl.DrawTextEx(font,
		string(fmt.Sprintf("%s %d:%d", file, t.Cursor.Line, t.Cursor.Column)),
		rl.Vector2{X: t.ZeroPosition.X, Y: float32(t.maxLine) * charSize.Y},
		fontSize,
		0,
		t.Colors.StatusBarForeground)

}

func (t *TextEditorBuffer) visualLineShouldBeRendered(line visualLine) bool {
	if t.VisibleStart <= int32(line.Index) && line.Index <= int(t.VisibleEnd) {
		return true
	}

	return false
}

func (t *TextEditorBuffer) renderVisualLine(line visualLine, index int) {
	charSize := measureTextSize(font, ' ', fontSize, 0)
	rl.DrawTextEx(font,
		string(t.Content[line.startIndex:line.endIndex+1]),
		rl.Vector2{X: t.ZeroPosition.X, Y: float32(index) * charSize.Y},
		fontSize,
		0,
		t.Colors.Foreground)

}

func (t *TextEditorBuffer) cursorToBufferIndex() int {
	if t.Cursor.Line >= len(t.visualLines) {
		return 0
	}
	return t.visualLines[t.Cursor.Line].startIndex + t.Cursor.Column
}

func (t *TextEditorBuffer) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if newPosition.Line >= len(t.visualLines) {
		return false
	}
	if newPosition.Column < 0 {
		return false
	}
	a := t.visualLines[newPosition.Line].endIndex - t.visualLines[newPosition.Line].startIndex
	if newPosition.Column > a+1 {
		return false
	}

	return true
}

func (t *TextEditorBuffer) InsertCharAtCursor(char byte) error {
	idx := t.cursorToBufferIndex()
	if idx >= len(t.Content) { // end of file, appending
		t.Content = append(t.Content, char)
	} else {
		t.Content = append(t.Content[:idx+1], t.Content[idx:]...)
		t.Content[idx] = char
	}
	if char == '\n' {
		t.CursorDown()
		t.BeginingOfTheLine()
	} else {
		t.CursorRight()
	}
	return nil

}

func (t *TextEditorBuffer) DeleteCharBackward() error {
	idx := t.cursorToBufferIndex()
	if idx <= 0 {
		return nil
	}
	if len(t.Content) <= idx {
		t.Content = t.Content[:idx]
	} else {
		t.Content = append(t.Content[:idx-1], t.Content[idx:]...)
	}
	t.CursorLeft()

	return nil

}

func (t *TextEditorBuffer) DeleteCharForeward() error {
	idx := t.cursorToBufferIndex()
	if idx < 0 || t.Cursor.Column < 0 {
		return nil
	}
	t.Content = append(t.Content[:idx], t.Content[idx+1:]...)
	return nil
}

func (t *TextEditorBuffer) ScrollUp(n int) error {
	if t.VisibleStart <= 0 {
		return nil
	}
	t.VisibleEnd += int32(-1 * n)
	t.VisibleStart += int32(-1 * n)

	diff := t.VisibleEnd - t.VisibleStart

	if t.VisibleStart < 0 {
		t.VisibleStart = 0
		t.VisibleEnd = diff
	}

	return nil

}

func (t *TextEditorBuffer) ScrollDown(n int) error {
	if int(t.VisibleEnd) >= len(t.visualLines) {
		return nil
	}
	t.VisibleEnd += int32(n)
	t.VisibleStart += int32(n)
	diff := t.VisibleEnd - t.VisibleStart
	if int(t.VisibleEnd) >= len(t.visualLines) {
		t.VisibleEnd = int32(len(t.visualLines) - 1)
		t.VisibleStart = t.VisibleEnd - diff
	}

	return nil

}

func (t *TextEditorBuffer) CursorLeft() error {
	newPosition := t.Cursor
	newPosition.Column--
	if t.Cursor.Column <= 0 {
		if newPosition.Line > 0 {
			newPosition.Line--
			lineColumns := t.visualLines[newPosition.Line].Length
			if lineColumns <= 0 {
				lineColumns = 0
			}
			newPosition.Column = lineColumns
		}

	}

	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}

	return nil

}

func (t *TextEditorBuffer) CursorRight() error {
	newPosition := t.Cursor
	newPosition.Column++
	lineColumns := t.visualLines[t.Cursor.Line].Length + 1

	if newPosition.Column > lineColumns+1 {
		fmt.Printf("new position :%+v\n", newPosition)
		fmt.Printf("line columns: %+v\n", lineColumns)
		newPosition.Line++
		newPosition.Column = 0
	}

	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}
	return nil

}

func (t *TextEditorBuffer) CursorUp() error {
	newPosition := t.Cursor
	newPosition.Line--

	if newPosition.Line < 0 {
		newPosition.Line = 0
	}

	if newPosition.Column > t.visualLines[newPosition.Line].Length {
		newPosition.Column = t.visualLines[newPosition.Line].Length
	}
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}

	return nil

}

func (t *TextEditorBuffer) CursorDown() error {
	newPosition := t.Cursor
	newPosition.Line++

	// check if cursor should be moved back
	if newPosition.Line < len(t.visualLines) {
		if newPosition.Column > t.visualLines[newPosition.Line].Length {
			newPosition.Column = t.visualLines[newPosition.Line].Length
		}
	} else {
		newPosition.Column = 0
	}

	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}
	return nil

}

func (t *TextEditorBuffer) BeginingOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = 0
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *TextEditorBuffer) EndOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = t.visualLines[t.Cursor.Line].Length
	if int32(newPosition.Column) < t.maxColumn {
		newPosition.Column++
	}
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *TextEditorBuffer) PreviousLine() error {
	return t.CursorUp()
}

func (t *TextEditorBuffer) NextLine() error {
	return t.CursorDown()
}

func (t *TextEditorBuffer) MoveCursorTo(pos rl.Vector2) error {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	apprColumn := pos.X / charSize.X
	apprLine := pos.Y / charSize.Y

	t.Cursor.Line = int(apprLine) + int(t.VisibleStart)
	t.Cursor.Column = int(apprColumn)

	if t.Cursor.Line > len(t.visualLines) {
		t.Cursor.Line = len(t.visualLines) - 1
	}

	// check if cursor should be moved back
	if t.Cursor.Column > t.visualLines[t.Cursor.Line].Length {
		t.Cursor.Column = t.visualLines[t.Cursor.Line].Length
	}

	fmt.Printf("moving cursor to: %+v\n", t.Cursor)

	return nil
}

func (t *TextEditorBuffer) MoveCursorToPositionAndScrollIfNeeded(pos Position) error {
	t.Cursor = pos

	if t.Cursor.Line == int(t.VisibleStart-1) {
		jump := int(t.maxLine / 2)
		t.ScrollUp(jump)
	}

	if t.Cursor.Line == int(t.VisibleEnd)+1 {
		jump := int(t.maxLine / 2)
		t.ScrollDown(jump)
	}

	return nil
}

func (t *TextEditorBuffer) Write() error {
	if t.File == "" {
		return nil
	}
	t.Content = bytes.Replace(t.Content, []byte("    "), []byte("\t"), -1)
	if err := os.WriteFile(t.File, t.Content, 0644); err != nil {
		return err
	}
	t.replaceTabsWithSpaces()
	return nil
}

func (t *TextEditorBuffer) GetMaxHeight() int32 { return t.MaxHeight }
func (t *TextEditorBuffer) GetMaxWidth() int32  { return t.MaxWidth }
