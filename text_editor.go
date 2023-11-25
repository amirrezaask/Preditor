package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	State_Clean = 1
	State_Dirty = 2
)

type TextEditor struct {
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
	State        int
}

func (t *TextEditor) replaceTabsWithSpaces() {
	t.Content = bytes.Replace(t.Content, []byte("\t"), []byte(strings.Repeat(" ", 4)), -1)
}

func (t *TextEditor) SetMaxWidth(w int32) {
	t.MaxWidth = w
	t.updateMaxLineAndColumn()
}
func (t *TextEditor) SetMaxHeight(h int32) {
	t.MaxHeight = h
	t.updateMaxLineAndColumn()
}
func (t *TextEditor) updateMaxLineAndColumn() {
	oldMaxLine := t.maxLine
	charSize := measureTextSize(font, ' ', fontSize, 0)
	t.maxColumn = t.MaxWidth / int32(charSize.X)
	t.maxLine = t.MaxHeight / int32(charSize.Y)

	// reserve one line for status bar
	t.maxLine--

	diff := t.maxLine - oldMaxLine
	t.VisibleEnd += diff
}
func (t *TextEditor) Type() string {
	return "text_editor_buffer"
}

func (t *TextEditor) Initialize(opts BufferOptions) error {
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
	}
	t.replaceTabsWithSpaces()
	t.updateMaxLineAndColumn()
	return nil
}

func (t *TextEditor) Destroy() error {
	return nil
}

type visualLine struct {
	Index      int
	startIndex int
	endIndex   int
	ActualLine int
	Length     int
}

func (t *TextEditor) calculateVisualLines() {
	t.visualLines = []visualLine{}
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
				Length:     idx - start + 1,
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
				Length:     idx - start + 1,
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
				Length:     idx - start + 1,
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

}

func (t *TextEditor) renderCursor() {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	// render cursor
	// fmt.Printf("Rendering buffer: render Loop took: %s\n", time.Since(loopStart))
	cursorView := Position{
		Line:   t.Cursor.Line - int(t.VisibleStart),
		Column: t.Cursor.Column,
	}
	rl.DrawRectangleLines(int32(cursorView.Column)*int32(charSize.X)+int32(t.ZeroPosition.X), int32(cursorView.Line)*int32(charSize.Y)+int32(t.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), rl.White)
}

func (t *TextEditor) Render() {

	t.calculateVisualLines()

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

	t.renderCursor()

}

func (t *TextEditor) visualLineShouldBeRendered(line visualLine) bool {
	if t.VisibleStart <= int32(line.Index) && line.Index <= int(t.VisibleEnd) {
		return true
	}

	return false
}

func (t *TextEditor) renderVisualLine(line visualLine, index int) {
	charSize := measureTextSize(font, ' ', fontSize, 0)
	rl.DrawTextEx(font,
		string(t.Content[line.startIndex:line.endIndex+1]),
		rl.Vector2{X: t.ZeroPosition.X, Y: float32(index) * charSize.Y},
		fontSize,
		0,
		t.Colors.Foreground)

}

func (t *TextEditor) cursorToBufferIndex() int {
	if t.Cursor.Line >= len(t.visualLines) {
		return 0
	}
	return t.visualLines[t.Cursor.Line].startIndex + t.Cursor.Column
}

func (t *TextEditor) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if newPosition.Line >= len(t.visualLines) {
		return false
	}
	if newPosition.Column < 0 {
		return false
	}
	a := t.visualLines[newPosition.Line].Length
	if newPosition.Column > a+1 {
		return false
	}

	return true
}

func (t *TextEditor) InsertCharAtCursor(char byte) error {
	idx := t.cursorToBufferIndex()
	if idx >= len(t.Content) { // end of file, appending
		t.Content = append(t.Content, char)
	} else {
		t.Content = append(t.Content[:idx+1], t.Content[idx:]...)
		t.Content[idx] = char
	}
	t.State = State_Dirty

	t.calculateVisualLines()
	if char == '\n' {
		t.CursorDown()
		t.BeginingOfTheLine()
	} else {
		t.CursorRight()
	}
	return nil

}

func (t *TextEditor) DeleteCharBackward() error {
	idx := t.cursorToBufferIndex()
	if idx <= 0 {
		return nil
	}
	if len(t.Content) <= idx {
		t.Content = t.Content[:idx]
	} else {
		t.Content = append(t.Content[:idx-1], t.Content[idx:]...)
	}
	t.State = State_Dirty
	t.calculateVisualLines()
	t.CursorLeft()

	return nil

}

func (t *TextEditor) DeleteCharForward() error {
	idx := t.cursorToBufferIndex()
	if idx < 0 || t.Cursor.Column < 0 {
		return nil
	}
	t.Content = append(t.Content[:idx], t.Content[idx+1:]...)
	t.State = State_Dirty

	t.calculateVisualLines()
	return nil
}

func (t *TextEditor) ScrollUp(n int) error {
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

func (t *TextEditor) ScrollDown(n int) error {
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

func (t *TextEditor) CursorLeft() error {
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

func (t *TextEditor) CursorRight() error {
	newPosition := t.Cursor
	newPosition.Column++
	if t.Cursor.Line == len(t.visualLines) {
		return nil
	}
	lineColumns := t.visualLines[t.Cursor.Line].Length
	if newPosition.Column > lineColumns+1 {
		newPosition.Line++
		newPosition.Column = 0
	}

	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}
	return nil

}

func (t *TextEditor) CursorUp() error {
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

func (t *TextEditor) CursorDown() error {
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

func (t *TextEditor) BeginingOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = 0
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *TextEditor) EndOfTheLine() error {
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

func (t *TextEditor) PreviousLine() error {
	return t.CursorUp()
}

func (t *TextEditor) NextLine() error {
	return t.CursorDown()
}

func (t *TextEditor) MoveCursorTo(pos rl.Vector2) error {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	apprColumn := pos.X / charSize.X
	apprLine := pos.Y / charSize.Y

	if len(t.visualLines) < 1 {
		return nil
	}

	t.Cursor.Line = int(apprLine) + int(t.VisibleStart)
	t.Cursor.Column = int(apprColumn)

	if t.Cursor.Line >= len(t.visualLines) {
		t.Cursor.Line = len(t.visualLines) - 1
	}

	// check if cursor should be moved back
	if t.Cursor.Column > t.visualLines[t.Cursor.Line].Length {
		t.Cursor.Column = t.visualLines[t.Cursor.Line].Length
	}

	fmt.Printf("moving cursor to: %+v\n", t.Cursor)

	return nil
}

func (t *TextEditor) MoveCursorToPositionAndScrollIfNeeded(pos Position) error {
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

func (t *TextEditor) Write() error {
	if t.File == "" {
		return nil
	}
	t.Content = bytes.Replace(t.Content, []byte("    "), []byte("\t"), -1)
	if err := os.WriteFile(t.File, t.Content, 0644); err != nil {
		return err
	}
	t.State = State_Clean
	t.replaceTabsWithSpaces()
	return nil
}

func (t *TextEditor) GetMaxHeight() int32 { return t.MaxHeight }
func (t *TextEditor) GetMaxWidth() int32  { return t.MaxWidth }
