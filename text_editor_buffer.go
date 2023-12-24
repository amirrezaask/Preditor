package main

import (
	"bytes"
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
	charSize := measureTextSize(font, ' ', fontSize, 0)
	t.maxColumn = t.MaxWidth / int32(charSize.X)
	t.maxLine = t.MaxHeight / int32(charSize.Y)
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
	t.Content, err = os.ReadFile(t.File)
	if err != nil {
		return err
	}
	t.replaceTabsWithSpaces()
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
	t.maxColumn = t.MaxWidth / int32(charSize.X)
	t.maxLine = t.MaxHeight / int32(charSize.Y)
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
			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1

		}
		if idx == len(t.Content) - 1 {
			// last index
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start - 1,
				ActualLine: actualLineIndex,
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
			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx

		}
	}
	// fmt.Printf("Render buffer in window: Scan Loop took: %s\n", time.Since(loopStart))
	// loopStart = time.Now()
	visibleView := t.visibleLines()
	for idx, line := range visibleView {
		if t.visualLineShouldBeRendered(line) {
			t.renderVisualLine(line, idx)
		}
	}
	// fmt.Printf("Rendering buffer: render Loop took: %s\n", time.Since(loopStart))
	rl.DrawRectangleLines(int32(t.Cursor.Column)*int32(charSize.X), int32(t.Cursor.Line)*int32(charSize.Y), int32(charSize.X), int32(charSize.Y), rl.White)
}

func (t *TextEditorBuffer) visibleLines() []visualLine {
	var visibleView []visualLine

	switch {
	case len(t.visualLines) > int(t.maxLine):
		visibleView = t.visualLines[t.VisibleStart:t.VisibleEnd]
	default:
		visibleView = t.visualLines
	}

	return visibleView

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
	return t.visibleLines()[t.Cursor.Line].startIndex + t.Cursor.Column
}

func (t *TextEditorBuffer) fixCursorColumnIfNeeded(newPosition *Position) {
	line := t.visibleLines()[newPosition.Line]
	if newPosition.Column > (line.Length) {
		newPosition.Column = line.Length
	}
	if newPosition.Column < 0 {
		newPosition.Column = 0
	}
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
	t.Content = append(t.Content[:idx+1], t.Content[idx:]...)
	t.Content[idx] = char
	charSize := measureTextSize(font, ' ', fontSize, 0)

	if char == '\n' {
		t.Cursor.Column = 0
		t.Cursor.Line++
	} else {
		if int32(t.Cursor.Column+1) < (t.MaxWidth / int32(charSize.X)) {
			t.Cursor.Column = t.Cursor.Column + 1
		}
	}
	return nil
}

func (t *TextEditorBuffer) DeleteCharBackward() error {
	idx := t.cursorToBufferIndex()
	if idx < 0 {
		return nil
	}
	if t.Cursor.Column == 0 {
		if err := t.PreviousLine(); err != nil {
			return err
		}
	}
	t.Content = append(t.Content[:idx-1], t.Content[idx:]...)
	t.Cursor.Column--
	t.fixCursorColumnIfNeeded(&t.Cursor)
	return nil

}

func (t *TextEditorBuffer) DeleteCharForeward() error {
	idx := t.cursorToBufferIndex()
	if idx < 0 || t.Cursor.Column < 0 {
		return nil
	}
	t.Content = append(t.Content[:idx], t.Content[idx+1:]...)
	t.fixCursorColumnIfNeeded(&t.Cursor)
	return nil
}

func (t *TextEditorBuffer) ScrollUp(n int) error {
	if t.VisibleStart <= 0 {
		return nil
	}
	t.VisibleEnd+=int32(-1*n)
	t.VisibleStart+=int32(-1*n)
	if t.VisibleStart < 0 {
		t.VisibleStart = 0
	}
	t.fixCursorColumnIfNeeded(&t.Cursor)

	return nil

}

func (t *TextEditorBuffer) ScrollDown(n int) error {
	if int(t.VisibleEnd) >= len(t.visualLines) {
		return nil
	}
	t.VisibleEnd++
	t.VisibleStart++
	if int(t.VisibleEnd) >= len(t.visualLines) {
		t.VisibleEnd = int32(len(t.visualLines) - 1)
	}
	t.fixCursorColumnIfNeeded(&t.Cursor)

	return nil

}

func (t *TextEditorBuffer) CursorLeft() error {
	newPosition := t.Cursor
	newPosition.Column--
	if t.Cursor.Column <= 0 {
		if newPosition.Line > 0 {
			newPosition.Line--
			lineColumns := t.visualLines[newPosition.Line].Length + 1
			if lineColumns <= 0 {
				lineColumns = 0
			}
			newPosition.Column = lineColumns
		}

	}

	if t.isValidCursorPosition(newPosition) {
		t.Cursor = newPosition
	}

	return nil

}

func (t *TextEditorBuffer) CursorRight() error {
	newPosition := t.Cursor
	newPosition.Column++
	lineColumns := t.visualLines[t.Cursor.Line].Length

	if newPosition.Column > lineColumns + 1 {
		newPosition.Line++
		newPosition.Column = 0
	}

	if t.isValidCursorPosition(newPosition) {
		t.Cursor = newPosition
	}
	return nil

}

func (t *TextEditorBuffer) CursorUp() error {
	newPosition := t.Cursor
	newPosition.Line--

	t.fixCursorColumnIfNeeded(&newPosition)
	if t.isValidCursorPosition(newPosition) {
		t.Cursor = newPosition
	}

	return nil

}

func (t *TextEditorBuffer) CursorDown() error {
	newPosition := t.Cursor
	newPosition.Line++
	t.fixCursorColumnIfNeeded(&newPosition)
	if t.isValidCursorPosition(newPosition) {
		t.Cursor = newPosition
	}
	return nil

}

func (t *TextEditorBuffer) BeginingOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = 0
	t.fixCursorColumnIfNeeded(&newPosition)
	if t.isValidCursorPosition(newPosition) {
		t.Cursor = newPosition
	}
	return nil

}

func (t *TextEditorBuffer) EndOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = t.visualLines[t.Cursor.Line].Length
	t.fixCursorColumnIfNeeded(&newPosition)
	if t.isValidCursorPosition(newPosition) {
		t.Cursor = newPosition
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

	t.Cursor.Line = int(apprLine)
	t.Cursor.Column = int(apprColumn)

	t.fixCursorColumnIfNeeded(&t.Cursor)

	return nil
}

func (t *TextEditorBuffer) Write() error {
	t.Content = bytes.Replace(t.Content, []byte("    "), []byte("\t"), -1)
	if err := os.WriteFile(t.File, t.Content, 0644); err != nil {
		return err
	}
	t.replaceTabsWithSpaces()
	return nil
}
