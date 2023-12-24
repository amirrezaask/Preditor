package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
)

const (
	State_Clean = 1
	State_Dirty = 2
)

type EditorBuffer struct {
	File              string
	Content           []byte
	Keymap            Keymap
	Variables         Variables
	Commands          Commands
	MaxHeight         int32
	MaxWidth          int32
	ZeroPosition      rl.Vector2
	TabSize           int
	VisibleStart      int32
	VisibleEnd        int32
	visualLines       []visualLine
	Cursor            Position
	maxLine           int32
	maxColumn         int32
	Colors            Colors
	State             int
	CursorBlinking    bool
	RenderLineNumbers bool
	HasSelection      bool
	SelectionStart    Position
	SelectionEnd      Position
}

func (t *EditorBuffer) replaceTabsWithSpaces() {
	t.Content = bytes.Replace(t.Content, []byte("\t"), []byte(strings.Repeat(" ", 4)), -1)
}

func (t *EditorBuffer) SetMaxWidth(w int32) {
	t.MaxWidth = w
	t.updateMaxLineAndColumn()
}
func (t *EditorBuffer) SetMaxHeight(h int32) {
	t.MaxHeight = h
	t.updateMaxLineAndColumn()
}
func (t *EditorBuffer) updateMaxLineAndColumn() {
	oldMaxLine := t.maxLine
	charSize := measureTextSize(font, ' ', fontSize, 0)
	t.maxColumn = t.MaxWidth / int32(charSize.X)
	t.maxLine = t.MaxHeight / int32(charSize.Y)

	// reserve one line for status bar
	t.maxLine--

	diff := t.maxLine - oldMaxLine
	t.VisibleEnd += diff
}
func (t *EditorBuffer) Type() string {
	return "text_editor_buffer"
}

type EditorBufferOptions struct {
	MaxHeight      int32
	MaxWidth       int32
	ZeroPosition   rl.Vector2
	Colors         Colors
	Filename       string
	LineNumbers    bool
	TabSize        int
	CursorBlinking bool
}

func NewEditorBuffer(opts EditorBufferOptions) (*EditorBuffer, error) {
	t := EditorBuffer{}
	t.File = opts.Filename
	t.RenderLineNumbers = opts.LineNumbers
	t.TabSize = opts.TabSize
	t.MaxHeight = opts.MaxHeight
	t.MaxWidth = opts.MaxWidth
	t.ZeroPosition = opts.ZeroPosition
	t.Colors = opts.Colors
	t.Keymap = editorBufferKeymap
	var err error
	if t.File != "" {
		t.Content, err = os.ReadFile(t.File)
		if err != nil {
			return nil, err
		}
	}
	t.replaceTabsWithSpaces()
	t.updateMaxLineAndColumn()
	return &t, nil

}

func (t *EditorBuffer) Destroy() error {
	return nil
}

type visualLine struct {
	Index      int
	startIndex int
	endIndex   int
	ActualLine int
	Length     int
}

func (t *EditorBuffer) calculateVisualLines() {
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
				Length:     idx - start,
				ActualLine: actualLineIndex,
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
				Length:     idx - start,
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
				Length:     idx - start,
				ActualLine: actualLineIndex,
			}

			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx

		}
	}

}

func (t *EditorBuffer) renderCursor() {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	// render cursor
	// fmt.Printf("Rendering buffer: render Loop took: %s\n", time.Since(loopStart))
	if t.CursorBlinking && (time.Now().Unix())%2 == 0 {
		return
	}
	cursorView := Position{
		Line:   t.Cursor.Line - int(t.VisibleStart),
		Column: t.Cursor.Column,
	}
	posX := int32(cursorView.Column)*int32(charSize.X) + int32(t.ZeroPosition.X)
	if t.RenderLineNumbers {
		if len(t.visualLines) > t.Cursor.Line {
			posX += int32((len(fmt.Sprint(t.visualLines[t.Cursor.Line].ActualLine)) + 1) * int(charSize.X))
		} else {
			posX += int32(charSize.X)

		}
	}
	rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(t.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), rl.White)
}

func (t *EditorBuffer) renderStatusBar() {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	//render status bar
	rl.DrawRectangle(
		int32(t.ZeroPosition.X),
		t.maxLine*int32(charSize.Y),
		t.MaxWidth,
		int32(charSize.Y),
		t.Colors.StatusBarBackground,
	)
	file := t.File
	if file == "" {
		file = "[scratch]"
	}
	var state string
	if t.State == State_Dirty {
		state = "**"
	} else {
		state = "--"
	}
	var line int
	if len(t.visualLines) > t.Cursor.Line {
		line = t.visualLines[t.Cursor.Line].ActualLine
	} else {
		line = 0
	}

	rl.DrawTextEx(font,
		fmt.Sprintf("%s %s %d:%d", state, file, line, t.Cursor.Column),
		rl.Vector2{X: t.ZeroPosition.X, Y: float32(t.maxLine) * charSize.Y},
		fontSize,
		0,
		t.Colors.StatusBarForeground)
}

func (t *EditorBuffer) Render() {

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
	t.renderStatusBar()
}

func (t *EditorBuffer) visualLineShouldBeRendered(line visualLine) bool {
	if t.VisibleStart <= int32(line.Index) && line.Index <= int(t.VisibleEnd) {
		return true
	}

	return false
}

func (t *EditorBuffer) renderVisualLine(line visualLine, index int) {
	charSize := measureTextSize(font, ' ', fontSize, 0)
	var lineNumberWidth int
	if t.RenderLineNumbers {
		lineNumberWidth = (len(fmt.Sprint(line.ActualLine)) + 1) * int(charSize.X)
		rl.DrawTextEx(font,
			fmt.Sprintf("%d", line.ActualLine),
			rl.Vector2{X: t.ZeroPosition.X, Y: float32(index) * charSize.Y},
			fontSize,
			0,
			t.Colors.LineNumbersForeground)

	}

	rl.DrawTextEx(font,
		string(t.Content[line.startIndex:line.endIndex+1]),
		rl.Vector2{X: t.ZeroPosition.X + float32(lineNumberWidth), Y: float32(index) * charSize.Y},
		fontSize,
		0,
		t.Colors.Foreground)

}

func (t *EditorBuffer) cursorToBufferIndex() int {
	if t.Cursor.Line >= len(t.visualLines) {
		return 0
	}
	return t.visualLines[t.Cursor.Line].startIndex + t.Cursor.Column
}

func (t *EditorBuffer) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if newPosition.Line >= len(t.visualLines) {
		return false
	}
	if newPosition.Column < 0 {
		return false
	}
	if newPosition.Column > t.visualLines[newPosition.Line].Length {
		return false
	}

	return true
}

func (t *EditorBuffer) InsertCharAtCursor(char byte) error {
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
		t.CursorRight(1)
	}
	return nil

}

func (t *EditorBuffer) DeleteCharBackward() error {
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

func (t *EditorBuffer) DeleteCharForward() error {
	idx := t.cursorToBufferIndex()
	if idx < 0 || t.Cursor.Column < 0 {
		return nil
	}
	t.Content = append(t.Content[:idx], t.Content[idx+1:]...)
	t.State = State_Dirty

	t.calculateVisualLines()
	return nil
}

func (t *EditorBuffer) ScrollUp(n int) error {
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

func (t *EditorBuffer) ScrollDown(n int) error {
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

func (t *EditorBuffer) CursorLeft() error {
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

func (t *EditorBuffer) CursorRight(n int) error {
	newPosition := t.Cursor
	newPosition.Column += n
	if t.Cursor.Line == len(t.visualLines) {
		return nil
	}
	lineColumns := t.visualLines[t.Cursor.Line].Length
	if newPosition.Column > lineColumns {
		newPosition.Line++
		newPosition.Column = 0
	}

	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}
	return nil

}

func (t *EditorBuffer) CursorUp() error {
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

func (t *EditorBuffer) CursorDown() error {
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

func (t *EditorBuffer) BeginingOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = 0
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *EditorBuffer) EndOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = t.visualLines[t.Cursor.Line].Length
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *EditorBuffer) PreviousLine() error {
	return t.CursorUp()
}

func (t *EditorBuffer) NextLine() error {
	return t.CursorDown()
}

func (t *EditorBuffer) MoveCursorTo(pos rl.Vector2) error {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	apprLine := pos.Y / charSize.Y
	apprColumn := pos.X / charSize.X
	if t.RenderLineNumbers {
		var line int
		if len(t.visualLines) > t.Cursor.Line {
			line = t.visualLines[t.Cursor.Line].ActualLine
		}
		apprColumn -= float32((len(fmt.Sprint(line)) + 1))

	}

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

func (t *EditorBuffer) MoveCursorToPositionAndScrollIfNeeded(pos Position) error {
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

func (t *EditorBuffer) Write() error {
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

func (t *EditorBuffer) GetMaxHeight() int32 { return t.MaxHeight }
func (t *EditorBuffer) GetMaxWidth() int32  { return t.MaxWidth }
func (t *EditorBuffer) Indent() error {
	idx := t.cursorToBufferIndex()
	if idx >= len(t.Content) { // end of file, appending
		t.Content = append(t.Content, []byte(strings.Repeat(" ", t.TabSize))...)
	} else {
		t.Content = append(t.Content[:idx], append([]byte(strings.Repeat(" ", t.TabSize)), t.Content[idx:]...)...)
		t.calculateVisualLines()
		t.CursorRight(t.TabSize)
	}

	t.State = State_Dirty

	return nil
}

var editorBufferKeymap = Keymap{

	Key{K: "s", Control: true}: func(e *Application) error {
		return e.ActiveEditor().Write()
	},
	// navigation
	Key{K: "<lmouse>-click"}: func(e *Application) error {
		return e.ActiveEditor().MoveCursorTo(rl.GetMousePosition())
	},
	Key{K: "<mouse-wheel-up>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollUp(10)

	},
	Key{K: "<mouse-wheel-down>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollDown(10)
	},

	Key{K: "<rmouse>-click"}: func(e *Application) error {
		return e.ActiveEditor().ScrollDown(10)
	},
	Key{K: "<mmouse>-click"}: func(e *Application) error {
		return e.ActiveEditor().ScrollUp(10)
	},

	Key{K: "a", Control: true}: func(e *Application) error {
		return e.ActiveEditor().BeginingOfTheLine()
	},
	Key{K: "e", Control: true}: func(e *Application) error {
		return e.ActiveEditor().EndOfTheLine()
	},

	Key{K: "p", Control: true}: func(e *Application) error {
		return e.ActiveEditor().PreviousLine()
	},

	Key{K: "n", Control: true}: func(e *Application) error {
		return e.ActiveEditor().NextLine()
	},

	Key{K: "<up>"}: func(e *Application) error {
		return e.ActiveEditor().CursorUp()
	},
	Key{K: "<down>"}: func(e *Application) error {
		return e.ActiveEditor().CursorDown()
	},
	Key{K: "<right>"}: func(e *Application) error {
		return e.ActiveEditor().CursorRight(1)
	},
	Key{K: "f", Control: true}: func(e *Application) error {
		return e.ActiveEditor().CursorRight(1)
	},
	Key{K: "<left>"}: func(e *Application) error {
		return e.ActiveEditor().CursorLeft()
	},

	Key{K: "b", Control: true}: func(e *Application) error {
		return e.ActiveEditor().CursorLeft()
	},
	Key{K: "<home>"}: func(e *Application) error {
		return e.ActiveEditor().BeginingOfTheLine()
	},
	Key{K: "<pagedown>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollDown(1)
	},
	Key{K: "<pageup>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollUp(1)
	},

	//insertion
	Key{K: "<enter>"}: func(e *Application) error { return insertCharAtCursor(e, '\n') },
	Key{K: "<space>"}: func(e *Application) error { return insertCharAtCursor(e, ' ') },
	Key{K: "<backspace>"}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharBackward()
	},
	Key{K: "d", Control: true}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "d", Control: true}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "<delete>"}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "a"}:               func(e *Application) error { return insertCharAtCursor(e, 'a') },
	Key{K: "b"}:               func(e *Application) error { return insertCharAtCursor(e, 'b') },
	Key{K: "c"}:               func(e *Application) error { return insertCharAtCursor(e, 'c') },
	Key{K: "d"}:               func(e *Application) error { return insertCharAtCursor(e, 'd') },
	Key{K: "e"}:               func(e *Application) error { return insertCharAtCursor(e, 'e') },
	Key{K: "f"}:               func(e *Application) error { return insertCharAtCursor(e, 'f') },
	Key{K: "g"}:               func(e *Application) error { return insertCharAtCursor(e, 'g') },
	Key{K: "h"}:               func(e *Application) error { return insertCharAtCursor(e, 'h') },
	Key{K: "i"}:               func(e *Application) error { return insertCharAtCursor(e, 'i') },
	Key{K: "j"}:               func(e *Application) error { return insertCharAtCursor(e, 'j') },
	Key{K: "k"}:               func(e *Application) error { return insertCharAtCursor(e, 'k') },
	Key{K: "l"}:               func(e *Application) error { return insertCharAtCursor(e, 'l') },
	Key{K: "m"}:               func(e *Application) error { return insertCharAtCursor(e, 'm') },
	Key{K: "n"}:               func(e *Application) error { return insertCharAtCursor(e, 'n') },
	Key{K: "o"}:               func(e *Application) error { return insertCharAtCursor(e, 'o') },
	Key{K: "p"}:               func(e *Application) error { return insertCharAtCursor(e, 'p') },
	Key{K: "q"}:               func(e *Application) error { return insertCharAtCursor(e, 'q') },
	Key{K: "r"}:               func(e *Application) error { return insertCharAtCursor(e, 'r') },
	Key{K: "s"}:               func(e *Application) error { return insertCharAtCursor(e, 's') },
	Key{K: "t"}:               func(e *Application) error { return insertCharAtCursor(e, 't') },
	Key{K: "u"}:               func(e *Application) error { return insertCharAtCursor(e, 'u') },
	Key{K: "v"}:               func(e *Application) error { return insertCharAtCursor(e, 'v') },
	Key{K: "w"}:               func(e *Application) error { return insertCharAtCursor(e, 'w') },
	Key{K: "x"}:               func(e *Application) error { return insertCharAtCursor(e, 'x') },
	Key{K: "y"}:               func(e *Application) error { return insertCharAtCursor(e, 'y') },
	Key{K: "z"}:               func(e *Application) error { return insertCharAtCursor(e, 'z') },
	Key{K: "0"}:               func(e *Application) error { return insertCharAtCursor(e, '0') },
	Key{K: "1"}:               func(e *Application) error { return insertCharAtCursor(e, '1') },
	Key{K: "2"}:               func(e *Application) error { return insertCharAtCursor(e, '2') },
	Key{K: "3"}:               func(e *Application) error { return insertCharAtCursor(e, '3') },
	Key{K: "4"}:               func(e *Application) error { return insertCharAtCursor(e, '4') },
	Key{K: "5"}:               func(e *Application) error { return insertCharAtCursor(e, '5') },
	Key{K: "6"}:               func(e *Application) error { return insertCharAtCursor(e, '6') },
	Key{K: "7"}:               func(e *Application) error { return insertCharAtCursor(e, '7') },
	Key{K: "8"}:               func(e *Application) error { return insertCharAtCursor(e, '8') },
	Key{K: "9"}:               func(e *Application) error { return insertCharAtCursor(e, '9') },
	Key{K: "\\"}:              func(e *Application) error { return insertCharAtCursor(e, '\\') },
	Key{K: "\\", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '|') },

	Key{K: "0", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, ')') },
	Key{K: "1", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '!') },
	Key{K: "2", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '@') },
	Key{K: "3", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '#') },
	Key{K: "4", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '$') },
	Key{K: "5", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '%') },
	Key{K: "6", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '^') },
	Key{K: "7", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '&') },
	Key{K: "8", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '*') },
	Key{K: "9", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '(') },
	Key{K: "a", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'A') },
	Key{K: "b", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'B') },
	Key{K: "c", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'C') },
	Key{K: "d", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'D') },
	Key{K: "e", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'E') },
	Key{K: "f", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'F') },
	Key{K: "g", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'G') },
	Key{K: "h", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'H') },
	Key{K: "i", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'I') },
	Key{K: "j", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'J') },
	Key{K: "k", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'K') },
	Key{K: "l", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'L') },
	Key{K: "m", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'M') },
	Key{K: "n", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'N') },
	Key{K: "o", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'O') },
	Key{K: "p", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'P') },
	Key{K: "q", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'Q') },
	Key{K: "r", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'R') },
	Key{K: "s", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'S') },
	Key{K: "t", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'T') },
	Key{K: "u", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'U') },
	Key{K: "v", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'V') },
	Key{K: "w", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'W') },
	Key{K: "x", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'X') },
	Key{K: "y", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'Y') },
	Key{K: "z", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'Z') },
	Key{K: "["}:              func(e *Application) error { return insertCharAtCursor(e, '[') },
	Key{K: "]"}:              func(e *Application) error { return insertCharAtCursor(e, ']') },
	Key{K: "{", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '{') },
	Key{K: "}", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '}') },
	Key{K: ";"}:              func(e *Application) error { return insertCharAtCursor(e, ';') },
	Key{K: ";", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, ':') },
	Key{K: "'"}:              func(e *Application) error { return insertCharAtCursor(e, '\'') },
	Key{K: "'", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '"') },
	Key{K: "\""}:             func(e *Application) error { return insertCharAtCursor(e, '"') },
	Key{K: ","}:              func(e *Application) error { return insertCharAtCursor(e, ',') },
	Key{K: "."}:              func(e *Application) error { return insertCharAtCursor(e, '.') },
	Key{K: ",", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '<') },
	Key{K: ".", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '>') },
	Key{K: "/"}:              func(e *Application) error { return insertCharAtCursor(e, '/') },
	Key{K: "/", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '?') },
	Key{K: "-"}:              func(e *Application) error { return insertCharAtCursor(e, '-') },
	Key{K: "="}:              func(e *Application) error { return insertCharAtCursor(e, '=') },
	Key{K: "-", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '_') },
	Key{K: "=", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '+') },
	Key{K: "`"}:              func(e *Application) error { return insertCharAtCursor(e, '`') },
	Key{K: "`", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '~') },
	Key{K: "<tab>"}:          func(e *Application) error { return e.ActiveEditor().Indent() },
}

func insertCharAtCursor(e *Application, char byte) error {
	return e.ActiveEditor().InsertCharAtCursor(char)
}

func getClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func writeToClipboard(bs []byte) {
	<-clipboard.Write(clipboard.FmtText, bs)
}
