package main

import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
)

const (
	State_Clean = 1
	State_Dirty = 2
)

type Editor struct {
	File                      string
	Content                   []byte
	Keymaps                   []Keymap
	ColorGroups               map[*regexp.Regexp]color.RGBA
	Variables                 Variables
	Commands                  Commands
	MaxHeight                 int32
	MaxWidth                  int32
	ZeroPosition              rl.Vector2
	TabSize                   int
	VisibleStart              int32
	VisibleEnd                int32
	visualLines               []visualLine
	Cursor                    Position
	maxLine                   int32
	maxColumn                 int32
	Colors                    Colors
	State                     int
	CursorBlinking            bool
	RenderLineNumbers         bool
	HasSelection              bool
	SelectionStart            *Position
	IsSearching               bool
	LastSearchString          string
	SearchString              *string
	SearchMatches             [][2]Position
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool
	CursorShape               int
	LastCursorBlink           time.Time
	BeforeSaveHook            []func(*Editor) error
}

func (t *Editor) replaceTabsWithSpaces() {
	t.Content = bytes.Replace(t.Content, []byte("\t"), []byte(strings.Repeat(" ", 4)), -1)
}

func (t *Editor) SetMaxWidth(w int32) {
	t.MaxWidth = w
	t.updateMaxLineAndColumn()
}
func (t *Editor) SetMaxHeight(h int32) {
	t.MaxHeight = h
	t.updateMaxLineAndColumn()
}
func (t *Editor) updateMaxLineAndColumn() {
	oldMaxLine := t.maxLine
	charSize := measureTextSize(font, ' ', fontSize, 0)
	t.maxColumn = t.MaxWidth / int32(charSize.X)
	t.maxLine = t.MaxHeight / int32(charSize.Y)

	// reserve one line for status bar
	t.maxLine--

	diff := t.maxLine - oldMaxLine
	t.VisibleEnd += diff
}
func (t *Editor) Type() string {
	return "text_editor_buffer"
}

const (
	CURSOR_SHAPE_BLOCK   = 1
	CURSOR_SHAPE_OUTLINE = 2
	CURSOR_SHAPE_LINE    = 3
)

type EditorOptions struct {
	MaxHeight      int32
	MaxWidth       int32
	ZeroPosition   rl.Vector2
	Colors         Colors
	Filename       string
	LineNumbers    bool
	TabSize        int
	CursorBlinking bool
	CursorShape    int
}

func NewEditor(opts EditorOptions) (*Editor, error) {
	t := Editor{}
	t.File = opts.Filename
	t.RenderLineNumbers = opts.LineNumbers
	t.TabSize = opts.TabSize
	t.MaxHeight = opts.MaxHeight
	t.MaxWidth = opts.MaxWidth
	t.ZeroPosition = opts.ZeroPosition
	t.Colors = opts.Colors
	t.Keymaps = append([]Keymap{}, editorKeymap)
	t.CursorShape = opts.CursorShape
	t.CursorBlinking = opts.CursorBlinking
	var err error
	if t.File != "" {
		t.Content, err = os.ReadFile(t.File)
		if err != nil {
			return nil, err
		}

		fileType, exists := fileTypeMappings[path.Ext(t.File)]
		if exists {
			t.BeforeSaveHook = append(t.BeforeSaveHook, fileType.BeforeSave)
			t.ColorGroups = fileType.ColorGroups
		}
	}
	t.replaceTabsWithSpaces()
	t.updateMaxLineAndColumn()
	return &t, nil

}

func (t *Editor) Destroy() error {
	return nil
}

type highlight struct {
	start int
	end   int
	Color color.RGBA
}

type visualLine struct {
	Index      int
	startIndex int
	endIndex   int
	ActualLine int
	Length     int
	Highlights []highlight
}

func (t *Editor) calculateHighlights(bs []byte, offset int) []highlight {
	var highlights []highlight
	for re, c := range t.ColorGroups {
		indexes := re.FindAllStringIndex(string(bs), -1)
		for _, index := range indexes {
			highlights = append(highlights, highlight{
				start: index[0] + offset,
				end:   index[1] + offset,
				Color: c,
			})
		}
	}

	return highlights
}
func Sort[T any](slice []T, pred func(t1 T, t2 T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return pred(slice[i], slice[j])
	})
}
func (t *Editor) fillInTheBlanks(hs []highlight, start, end int) []highlight {
	var missing []highlight
	if len(hs) == 0 {
		missing = append(missing, highlight{
			start: start,
			end:   end,
			Color: t.Colors.Foreground,
		})
	} else {
		for i, h := range hs {
			if i+1 < len(hs) && hs[i+1].start-h.end != 1 {
				missing = append(missing, highlight{
					start: h.end + 1,
					end:   hs[i+1].start - 1,
					Color: t.Colors.Foreground,
				})
			}
		}
	}

	hs = append(hs, missing...)
	Sort[highlight](hs, func(t1, t2 highlight) bool {
		return t1.start < t2.start
	})

	return hs
}

func (t *Editor) calculateVisualLines() {
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
				Highlights: t.fillInTheBlanks(t.calculateHighlights(t.Content[start:idx], start), start, idx),
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
				Highlights: t.fillInTheBlanks(t.calculateHighlights(t.Content[start:], start), start, idx),
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
				Highlights: t.fillInTheBlanks(t.calculateHighlights(t.Content[start:idx], start), start, idx),
			}

			t.visualLines = append(t.visualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx

		}
	}
}

func (t *Editor) renderCursor() {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	if t.CursorBlinking && time.Since(t.LastCursorBlink).Milliseconds() < 1000 {
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
	switch t.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(t.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), t.Colors.Cursor)
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(t.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(t.Colors.Cursor, 0.6))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(t.ZeroPosition.Y), 2, int32(charSize.Y), t.Colors.Cursor)
	}

	t.LastCursorBlink = time.Now()
}

func (t *Editor) renderStatusBar() {
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
	var searchString string
	if t.SearchString != nil {
		searchString = fmt.Sprintf("Searching: \"%s\" %d of %d matches", *t.SearchString, t.CurrentMatch, len(t.SearchMatches)-1)
	}

	rl.DrawTextEx(font,
		fmt.Sprintf("%s %s %d:%d %s", state, file, line, t.Cursor.Column, searchString),
		rl.Vector2{X: t.ZeroPosition.X, Y: float32(t.maxLine) * charSize.Y},
		fontSize,
		0,
		t.Colors.StatusBarForeground)
}

func (t *Editor) highilightBetweenTwoPositions(start Position, end Position, color color.RGBA) {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	for i := start.Line; i <= end.Line; i++ {
		if len(t.visualLines) <= i {
			break
		}
		var thisLineEnd int
		var thisLineStart int
		line := t.visualLines[i]
		if i == start.Line {
			thisLineStart = start.Column
		} else {
			thisLineStart = 0
		}

		if i < end.Line {
			thisLineEnd = line.Length - 1
		} else {
			thisLineEnd = end.Column
		}
		for j := thisLineStart; j <= thisLineEnd; j++ {
			posX := int32(j)*int32(charSize.X) + int32(t.ZeroPosition.X)
			if t.RenderLineNumbers {
				if len(t.visualLines) > i {
					posX += int32((len(fmt.Sprint(t.visualLines[i].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			rl.DrawRectangle(posX, int32(i-int(t.VisibleStart))*int32(charSize.Y)+int32(t.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(color, 0.5))
		}
	}

}

func (t *Editor) renderSelection() {
	if t.SelectionStart == nil {
		return
	}

	var startLine int
	var startColumn int
	var endLine int
	var endColumn int
	switch {
	case t.SelectionStart.Line < t.Cursor.Line:
		startLine = t.SelectionStart.Line
		startColumn = t.SelectionStart.Column
		endLine = t.Cursor.Line
		endColumn = t.Cursor.Column
	case t.Cursor.Line < t.SelectionStart.Line:
		startLine = t.Cursor.Line
		startColumn = t.Cursor.Column
		endLine = t.SelectionStart.Line
		endColumn = t.SelectionStart.Column
	case t.Cursor.Line == t.SelectionStart.Line:
		startLine = t.Cursor.Line
		endLine = t.Cursor.Line
		if t.SelectionStart.Column > t.Cursor.Column {
			startColumn = t.Cursor.Column
			endColumn = t.SelectionStart.Column
		} else {
			startColumn = t.SelectionStart.Column
			endColumn = t.Cursor.Column
		}

	}

	t.highilightBetweenTwoPositions(Position{
		Line:   startLine,
		Column: startColumn,
	}, Position{
		Line:   endLine,
		Column: endColumn,
	}, t.Colors.Selection)

}

func (t *Editor) renderText() {
	var visibleLines []visualLine
	if t.VisibleEnd > int32(len(t.visualLines)) {
		visibleLines = t.visualLines[t.VisibleStart:]
	} else {
		visibleLines = t.visualLines[t.VisibleStart:t.VisibleEnd]
	}
	for idx, line := range visibleLines {
		if t.visualLineShouldBeRendered(line) {
			charSize := measureTextSize(font, ' ', fontSize, 0)
			var lineNumberWidth int
			if t.RenderLineNumbers {
				lineNumberWidth = (len(fmt.Sprint(line.ActualLine)) + 1) * int(charSize.X)
				rl.DrawTextEx(font,
					fmt.Sprintf("%d", line.ActualLine),
					rl.Vector2{X: t.ZeroPosition.X, Y: float32(idx) * charSize.Y},
					fontSize,
					0,
					t.Colors.LineNumbersForeground)

			}

			rl.DrawTextEx(font,
				string(t.Content[line.startIndex:line.endIndex+1]),
				rl.Vector2{X: t.ZeroPosition.X + float32(lineNumberWidth), Y: float32(idx) * charSize.Y},
				fontSize,
				0,
				t.Colors.Foreground)
		}
	}
}
func (t *Editor) convertBufferIndexToLineAndColumn(idx int) *Position {
	for lineIndex, line := range t.visualLines {
		if line.startIndex <= idx && line.endIndex >= idx {
			return &Position{
				Line:   lineIndex,
				Column: idx - line.startIndex,
			}
		}
	}

	return nil
}
func (t *Editor) findMatchesRegex(pattern string) error {
	t.SearchMatches = [][2]Position{}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matches := re.FindAllStringIndex(string(t.Content), -1)
	for _, match := range matches {
		matchStart := t.convertBufferIndexToLineAndColumn(match[0])
		matchEnd := t.convertBufferIndexToLineAndColumn(match[1])
		if matchStart == nil || matchEnd == nil {
			continue
		}
		matchEnd.Column--
		t.SearchMatches = append(t.SearchMatches, [2]Position{*matchStart, *matchEnd})
	}

	return nil
}

func (t *Editor) findMatchesAndHighlight(pattern string) error {
	if pattern != t.LastSearchString {
		if err := t.findMatchesRegex(pattern); err != nil {
			return err
		}
	}
	for idx, match := range t.SearchMatches {
		c := t.Colors.Selection
		if idx == t.CurrentMatch {
			c = rl.Fade(rl.Red, 0.5)
			if !(t.VisibleStart < int32(match[0].Line) && t.VisibleEnd > int32(match[1].Line)) && !t.MovedAwayFromCurrentMatch {
				// current match is not in view
				// move the view
				oldStart := t.VisibleStart
				t.VisibleStart = int32(match[0].Line) - t.maxLine/2
				if t.VisibleStart < 0 {
					t.VisibleStart = int32(match[0].Line)
				}

				diff := t.VisibleStart - oldStart
				t.VisibleEnd += diff
			}
		}
		t.highilightBetweenTwoPositions(match[0], match[1], c)
	}
	t.LastSearchString = pattern

	return nil
}
func (t *Editor) renderSearchResults() {
	if t.SearchString == nil || len(*t.SearchString) < 1 {
		return
	}
	t.findMatchesAndHighlight(*t.SearchString)
}

func (t *Editor) Render() {
	t.calculateVisualLines()
	t.renderText()
	t.renderSearchResults()
	t.renderCursor()
	t.renderStatusBar()
	t.renderSelection()
}

func (t *Editor) visualLineShouldBeRendered(line visualLine) bool {
	if t.VisibleStart <= int32(line.Index) && line.Index <= int(t.VisibleEnd) {
		return true
	}

	return false
}

func (t *Editor) positionToBufferIndex(pos Position) int {
	if pos.Line >= len(t.visualLines) {
		return 0
	}
	return t.visualLines[pos.Line].startIndex + pos.Column
}

func (t *Editor) isValidCursorPosition(newPosition Position) bool {
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

func (t *Editor) InsertCharAtCursor(char byte) error {
	var idx int
	if t.HasSelection {
		var startLine int
		var startColumn int
		var endLine int
		var endColumn int
		switch {
		case t.SelectionStart.Line < t.Cursor.Line:
			startLine = t.SelectionStart.Line
			startColumn = t.SelectionStart.Column
			endLine = t.Cursor.Line
			endColumn = t.Cursor.Column
		case t.Cursor.Line < t.SelectionStart.Line:
			startLine = t.Cursor.Line
			startColumn = t.Cursor.Column
			endLine = t.SelectionStart.Line
			endColumn = t.SelectionStart.Column
		case t.Cursor.Line == t.SelectionStart.Line:
			startLine = t.Cursor.Line
			endLine = t.Cursor.Line
			if t.SelectionStart.Column > t.Cursor.Column {
				startColumn = t.Cursor.Column
				endColumn = t.SelectionStart.Column
			} else {
				startColumn = t.SelectionStart.Column
				endColumn = t.Cursor.Column
			}
			t.HasSelection = false
			t.SelectionStart = nil
		}
		idx = t.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		startIndex := t.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		endIndex := t.positionToBufferIndex(Position{Line: endLine, Column: endColumn})

		t.Content = append(t.Content[:startIndex], t.Content[endIndex+1:]...)
		t.Cursor = Position{Line: startLine, Column: startColumn}
		t.calculateVisualLines()
	} else {
		idx = t.positionToBufferIndex(t.Cursor)
	}
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

func (t *Editor) DeleteCharBackward() error {
	if t.HasSelection {
		var startLine int
		var startColumn int
		var endLine int
		var endColumn int
		switch {
		case t.SelectionStart.Line < t.Cursor.Line:
			startLine = t.SelectionStart.Line
			startColumn = t.SelectionStart.Column
			endLine = t.Cursor.Line
			endColumn = t.Cursor.Column
		case t.Cursor.Line < t.SelectionStart.Line:
			startLine = t.Cursor.Line
			startColumn = t.Cursor.Column
			endLine = t.SelectionStart.Line
			endColumn = t.SelectionStart.Column
		case t.Cursor.Line == t.SelectionStart.Line:
			startLine = t.Cursor.Line
			endLine = t.Cursor.Line
			if t.SelectionStart.Column > t.Cursor.Column {
				startColumn = t.Cursor.Column
				endColumn = t.SelectionStart.Column
			} else {
				startColumn = t.SelectionStart.Column
				endColumn = t.Cursor.Column
			}
			t.HasSelection = false
			t.SelectionStart = nil
		}
		startIndex := t.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		endIndex := t.positionToBufferIndex(Position{Line: endLine, Column: endColumn})

		t.Content = append(t.Content[:startIndex], t.Content[endIndex+1:]...)
		t.Cursor = Position{Line: startLine, Column: startColumn}
		t.State = State_Dirty
		t.HasSelection = false
		return nil
	}

	idx := t.positionToBufferIndex(t.Cursor)
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

func (t *Editor) DeleteCharForward() error {
	idx := t.positionToBufferIndex(t.Cursor)
	if idx < 0 || t.Cursor.Column < 0 {
		return nil
	}
	t.Content = append(t.Content[:idx], t.Content[idx+1:]...)
	t.State = State_Dirty

	return nil
}

func (t *Editor) ScrollUp(n int) error {
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

func (t *Editor) ScrollDown(n int) error {
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

func (t *Editor) CursorLeft() error {
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

func (t *Editor) CursorRight(n int) error {
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

func (t *Editor) CursorUp() error {
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

func (t *Editor) CursorDown() error {
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

func (t *Editor) BeginingOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = 0
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *Editor) EndOfTheLine() error {
	newPosition := t.Cursor
	newPosition.Column = t.visualLines[t.Cursor.Line].Length
	if t.isValidCursorPosition(newPosition) {
		t.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (t *Editor) PreviousLine() error {
	return t.CursorUp()
}

func (t *Editor) NextLine() error {
	return t.CursorDown()
}
func (t *Editor) indexOfFirstNonLetter(bs []byte) int {
	for idx, b := range bs {
		if !unicode.IsLetter(rune(b)) {
			return idx
		}
	}

	return -1
}

func (t *Editor) NextWord() error {
	currentidx := t.positionToBufferIndex(t.Cursor)
	if len(t.Content) <= currentidx+1 {
		return nil
	}
	jump := t.indexOfFirstNonLetter(t.Content[currentidx+1:])
	if jump == -1 {
		return nil
	}
	if jump == 0 {
		jump = 1
	}
	pos := t.convertBufferIndexToLineAndColumn(jump + currentidx)

	Printlnf(pos)
	if t.isValidCursorPosition(*pos) {
		return t.MoveCursorToPositionAndScrollIfNeeded(*pos)
	}
	return nil
}

func (t *Editor) PreviousWord() error {
	currentidx := t.positionToBufferIndex(t.Cursor)
	if len(t.Content) <= currentidx+1 {
		return nil
	}
	jump := t.indexOfFirstNonLetter(t.Content[:currentidx])
	if jump == -1 {
		return nil
	}
	if jump == 0 {
		jump = 1
	}
	pos := t.convertBufferIndexToLineAndColumn(currentidx - jump)

	Printlnf(pos)
	if t.isValidCursorPosition(*pos) {
		return t.MoveCursorToPositionAndScrollIfNeeded(*pos)
	}
	return nil
}

func (t *Editor) MoveCursorTo(pos rl.Vector2) error {
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

func (t *Editor) MoveCursorToPositionAndScrollIfNeeded(pos Position) error {
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

func (t *Editor) Write() error {
	if t.File == "" {
		return nil
	}

	for _, hook := range t.BeforeSaveHook {
		if err := hook(t); err != nil {
			return err
		}
	}

	t.Content = bytes.Replace(t.Content, []byte("	"), []byte("\t"), -1)
	if err := os.WriteFile(t.File, t.Content, 0644); err != nil {
		return err
	}
	t.State = State_Clean
	t.replaceTabsWithSpaces()
	return nil
}

func (t *Editor) GetMaxHeight() int32 { return t.MaxHeight }
func (t *Editor) GetMaxWidth() int32  { return t.MaxWidth }
func (t *Editor) Indent() error {
	idx := t.positionToBufferIndex(t.Cursor)
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

func (t *Editor) copy() error {
	if t.HasSelection {
		// copy selection
		selection := t.positionToBufferIndex(*t.SelectionStart)
		cursor := t.positionToBufferIndex(t.Cursor)
		switch {
		case selection < cursor:
			writeToClipboard(t.Content[selection:cursor])
		case selection > cursor:
			writeToClipboard(t.Content[cursor:selection])
		case cursor == selection:
			return nil
		}
	} else {
		writeToClipboard(t.Content[t.visualLines[t.Cursor.Line].startIndex:t.visualLines[t.Cursor.Line].endIndex])
	}

	return nil
}

func (t *Editor) cut() error {
	var startIndex int
	var endIndex int
	if t.HasSelection {
		// copy selection
		selection := t.positionToBufferIndex(*t.SelectionStart)
		cursor := t.positionToBufferIndex(t.Cursor)
		switch {
		case selection < cursor:
			writeToClipboard(t.Content[selection:cursor])
			startIndex = selection
			endIndex = cursor
		case selection > cursor:
			writeToClipboard(t.Content[cursor:selection])
			startIndex = cursor
			endIndex = selection
		case cursor == selection:
			return nil
		}
		t.HasSelection = false
		t.SelectionStart = nil
	} else {
		writeToClipboard(t.Content[t.visualLines[t.Cursor.Line].startIndex:t.visualLines[t.Cursor.Line].endIndex])
		startIndex = t.visualLines[t.Cursor.Line].startIndex
		endIndex = t.visualLines[t.Cursor.Line].endIndex
	}

	t.Content = append(t.Content[:startIndex], t.Content[endIndex+1:]...)

	t.State = State_Dirty

	return nil
}
func (t *Editor) paste() error {
	var idx int
	if t.HasSelection {
		var startLine int
		var startColumn int
		var endLine int
		var endColumn int
		switch {
		case t.SelectionStart.Line < t.Cursor.Line:
			startLine = t.SelectionStart.Line
			startColumn = t.SelectionStart.Column
			endLine = t.Cursor.Line
			endColumn = t.Cursor.Column
		case t.Cursor.Line < t.SelectionStart.Line:
			startLine = t.Cursor.Line
			startColumn = t.Cursor.Column
			endLine = t.SelectionStart.Line
			endColumn = t.SelectionStart.Column
		case t.Cursor.Line == t.SelectionStart.Line:
			startLine = t.Cursor.Line
			endLine = t.Cursor.Line
			if t.SelectionStart.Column > t.Cursor.Column {
				startColumn = t.Cursor.Column
				endColumn = t.SelectionStart.Column
			} else {
				startColumn = t.SelectionStart.Column
				endColumn = t.Cursor.Column
			}
			t.HasSelection = false
			t.SelectionStart = nil
		}
		idx = t.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		startIndex := t.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		endIndex := t.positionToBufferIndex(Position{Line: endLine, Column: endColumn})

		t.Content = append(t.Content[:startIndex], t.Content[endIndex+1:]...)
		t.Cursor = Position{Line: startLine, Column: startColumn}
		t.calculateVisualLines()
	} else {
		idx = t.positionToBufferIndex(t.Cursor)
	}
	contentToPaste := getClipboardContent()
	fmt.Printf("content to paste: '%s'\n", contentToPaste)
	if idx >= len(t.Content) { // end of file, appending
		t.Content = append(t.Content, contentToPaste...)
	} else {
		t.Content = append(t.Content[:idx], append(contentToPaste, t.Content[idx:]...)...)
	}
	t.State = State_Dirty
	t.HasSelection = false
	t.SelectionStart = nil
	t.calculateVisualLines()

	t.CursorRight(len(contentToPaste))
	return nil
}

var editorKeymap = Keymap{
	Key{K: "f", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().CursorRight(1)
	},
	Key{K: "s", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().Write()
	},
	Key{K: "c", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().copy()
	},
	Key{K: "v", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().paste()
	},
	Key{K: "x", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().cut()
	},
	Key{K: "f", Alt: true}: func(a *Preditor) error {
		a.ActiveEditor().Keymaps = append(a.ActiveEditor().Keymaps, searchModeKeymap)
		return nil
	},
	Key{K: "<esc>"}: func(p *Preditor) error {
		editor := p.ActiveEditor()
		if editor.HasSelection {
			editor.HasSelection = !editor.HasSelection
			editor.SelectionStart = nil
		}

		return nil
	},

	//selection
	Key{K: "<space>", Control: true}: func(e *Preditor) error {
		editor := e.ActiveEditor()
		if editor.HasSelection {
			editor.HasSelection = !editor.HasSelection
			editor.SelectionStart = nil

		} else {
			editor.HasSelection = !editor.HasSelection
			editor.SelectionStart = &Position{
				Line:   editor.Cursor.Line,
				Column: editor.Cursor.Column,
			}

		}
		return nil
	},

	// navigation
	Key{K: "<lmouse>-click"}: func(e *Preditor) error {
		return e.ActiveEditor().MoveCursorTo(rl.GetMousePosition())
	},
	Key{K: "<mouse-wheel-up>"}: func(e *Preditor) error {
		return e.ActiveEditor().ScrollUp(10)

	},
	Key{K: "<mouse-wheel-down>"}: func(e *Preditor) error {
		return e.ActiveEditor().ScrollDown(10)
	},

	Key{K: "<rmouse>-click"}: func(e *Preditor) error {
		return e.ActiveEditor().ScrollDown(10)
	},
	Key{K: "<mmouse>-click"}: func(e *Preditor) error {
		return e.ActiveEditor().ScrollUp(10)
	},

	Key{K: "a", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().BeginingOfTheLine()
	},
	Key{K: "e", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().EndOfTheLine()
	},

	Key{K: "p", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().PreviousLine()
	},

	Key{K: "n", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().NextLine()
	},

	Key{K: "<up>"}: func(e *Preditor) error {
		return e.ActiveEditor().CursorUp()
	},
	Key{K: "<down>"}: func(e *Preditor) error {
		return e.ActiveEditor().CursorDown()
	},
	Key{K: "<right>"}: func(e *Preditor) error {
		return e.ActiveEditor().CursorRight(1)
	},
	Key{K: "<right>", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().NextWord()
	},
	Key{K: "<left>"}: func(e *Preditor) error {
		return e.ActiveEditor().CursorLeft()
	},
	Key{K: "<left>", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().PreviousWord()
	},

	Key{K: "b", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().CursorLeft()
	},
	Key{K: "<home>"}: func(e *Preditor) error {
		return e.ActiveEditor().BeginingOfTheLine()
	},
	Key{K: "<pagedown>"}: func(e *Preditor) error {
		return e.ActiveEditor().ScrollDown(1)
	},
	Key{K: "<pageup>"}: func(e *Preditor) error {
		return e.ActiveEditor().ScrollUp(1)
	},

	//insertion
	Key{K: "<enter>"}: func(e *Preditor) error { return insertChar(e, '\n') },
	Key{K: "<space>"}: func(e *Preditor) error { return insertChar(e, ' ') },
	Key{K: "<backspace>"}: func(e *Preditor) error {
		return e.ActiveEditor().DeleteCharBackward()
	},
	Key{K: "d", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "d", Control: true}: func(e *Preditor) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "<delete>"}: func(e *Preditor) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "a"}:               func(e *Preditor) error { return insertChar(e, 'a') },
	Key{K: "b"}:               func(e *Preditor) error { return insertChar(e, 'b') },
	Key{K: "c"}:               func(e *Preditor) error { return insertChar(e, 'c') },
	Key{K: "d"}:               func(e *Preditor) error { return insertChar(e, 'd') },
	Key{K: "e"}:               func(e *Preditor) error { return insertChar(e, 'e') },
	Key{K: "f"}:               func(e *Preditor) error { return insertChar(e, 'f') },
	Key{K: "g"}:               func(e *Preditor) error { return insertChar(e, 'g') },
	Key{K: "h"}:               func(e *Preditor) error { return insertChar(e, 'h') },
	Key{K: "i"}:               func(e *Preditor) error { return insertChar(e, 'i') },
	Key{K: "j"}:               func(e *Preditor) error { return insertChar(e, 'j') },
	Key{K: "k"}:               func(e *Preditor) error { return insertChar(e, 'k') },
	Key{K: "l"}:               func(e *Preditor) error { return insertChar(e, 'l') },
	Key{K: "m"}:               func(e *Preditor) error { return insertChar(e, 'm') },
	Key{K: "n"}:               func(e *Preditor) error { return insertChar(e, 'n') },
	Key{K: "o"}:               func(e *Preditor) error { return insertChar(e, 'o') },
	Key{K: "p"}:               func(e *Preditor) error { return insertChar(e, 'p') },
	Key{K: "q"}:               func(e *Preditor) error { return insertChar(e, 'q') },
	Key{K: "r"}:               func(e *Preditor) error { return insertChar(e, 'r') },
	Key{K: "s"}:               func(e *Preditor) error { return insertChar(e, 's') },
	Key{K: "t"}:               func(e *Preditor) error { return insertChar(e, 't') },
	Key{K: "u"}:               func(e *Preditor) error { return insertChar(e, 'u') },
	Key{K: "v"}:               func(e *Preditor) error { return insertChar(e, 'v') },
	Key{K: "w"}:               func(e *Preditor) error { return insertChar(e, 'w') },
	Key{K: "x"}:               func(e *Preditor) error { return insertChar(e, 'x') },
	Key{K: "y"}:               func(e *Preditor) error { return insertChar(e, 'y') },
	Key{K: "z"}:               func(e *Preditor) error { return insertChar(e, 'z') },
	Key{K: "0"}:               func(e *Preditor) error { return insertChar(e, '0') },
	Key{K: "1"}:               func(e *Preditor) error { return insertChar(e, '1') },
	Key{K: "2"}:               func(e *Preditor) error { return insertChar(e, '2') },
	Key{K: "3"}:               func(e *Preditor) error { return insertChar(e, '3') },
	Key{K: "4"}:               func(e *Preditor) error { return insertChar(e, '4') },
	Key{K: "5"}:               func(e *Preditor) error { return insertChar(e, '5') },
	Key{K: "6"}:               func(e *Preditor) error { return insertChar(e, '6') },
	Key{K: "7"}:               func(e *Preditor) error { return insertChar(e, '7') },
	Key{K: "8"}:               func(e *Preditor) error { return insertChar(e, '8') },
	Key{K: "9"}:               func(e *Preditor) error { return insertChar(e, '9') },
	Key{K: "\\"}:              func(e *Preditor) error { return insertChar(e, '\\') },
	Key{K: "\\", Shift: true}: func(e *Preditor) error { return insertChar(e, '|') },

	Key{K: "0", Shift: true}: func(e *Preditor) error { return insertChar(e, ')') },
	Key{K: "1", Shift: true}: func(e *Preditor) error { return insertChar(e, '!') },
	Key{K: "2", Shift: true}: func(e *Preditor) error { return insertChar(e, '@') },
	Key{K: "3", Shift: true}: func(e *Preditor) error { return insertChar(e, '#') },
	Key{K: "4", Shift: true}: func(e *Preditor) error { return insertChar(e, '$') },
	Key{K: "5", Shift: true}: func(e *Preditor) error { return insertChar(e, '%') },
	Key{K: "6", Shift: true}: func(e *Preditor) error { return insertChar(e, '^') },
	Key{K: "7", Shift: true}: func(e *Preditor) error { return insertChar(e, '&') },
	Key{K: "8", Shift: true}: func(e *Preditor) error { return insertChar(e, '*') },
	Key{K: "9", Shift: true}: func(e *Preditor) error { return insertChar(e, '(') },
	Key{K: "a", Shift: true}: func(e *Preditor) error { return insertChar(e, 'A') },
	Key{K: "b", Shift: true}: func(e *Preditor) error { return insertChar(e, 'B') },
	Key{K: "c", Shift: true}: func(e *Preditor) error { return insertChar(e, 'C') },
	Key{K: "d", Shift: true}: func(e *Preditor) error { return insertChar(e, 'D') },
	Key{K: "e", Shift: true}: func(e *Preditor) error { return insertChar(e, 'E') },
	Key{K: "f", Shift: true}: func(e *Preditor) error { return insertChar(e, 'F') },
	Key{K: "g", Shift: true}: func(e *Preditor) error { return insertChar(e, 'G') },
	Key{K: "h", Shift: true}: func(e *Preditor) error { return insertChar(e, 'H') },
	Key{K: "i", Shift: true}: func(e *Preditor) error { return insertChar(e, 'I') },
	Key{K: "j", Shift: true}: func(e *Preditor) error { return insertChar(e, 'J') },
	Key{K: "k", Shift: true}: func(e *Preditor) error { return insertChar(e, 'K') },
	Key{K: "l", Shift: true}: func(e *Preditor) error { return insertChar(e, 'L') },
	Key{K: "m", Shift: true}: func(e *Preditor) error { return insertChar(e, 'M') },
	Key{K: "n", Shift: true}: func(e *Preditor) error { return insertChar(e, 'N') },
	Key{K: "o", Shift: true}: func(e *Preditor) error { return insertChar(e, 'O') },
	Key{K: "p", Shift: true}: func(e *Preditor) error { return insertChar(e, 'P') },
	Key{K: "q", Shift: true}: func(e *Preditor) error { return insertChar(e, 'Q') },
	Key{K: "r", Shift: true}: func(e *Preditor) error { return insertChar(e, 'R') },
	Key{K: "s", Shift: true}: func(e *Preditor) error { return insertChar(e, 'S') },
	Key{K: "t", Shift: true}: func(e *Preditor) error { return insertChar(e, 'T') },
	Key{K: "u", Shift: true}: func(e *Preditor) error { return insertChar(e, 'U') },
	Key{K: "v", Shift: true}: func(e *Preditor) error { return insertChar(e, 'V') },
	Key{K: "w", Shift: true}: func(e *Preditor) error { return insertChar(e, 'W') },
	Key{K: "x", Shift: true}: func(e *Preditor) error { return insertChar(e, 'X') },
	Key{K: "y", Shift: true}: func(e *Preditor) error { return insertChar(e, 'Y') },
	Key{K: "z", Shift: true}: func(e *Preditor) error { return insertChar(e, 'Z') },
	Key{K: "["}:              func(e *Preditor) error { return insertChar(e, '[') },
	Key{K: "]"}:              func(e *Preditor) error { return insertChar(e, ']') },
	Key{K: "[", Shift: true}: func(e *Preditor) error { return insertChar(e, '{') },
	Key{K: "]", Shift: true}: func(e *Preditor) error { return insertChar(e, '}') },
	Key{K: ";"}:              func(e *Preditor) error { return insertChar(e, ';') },
	Key{K: ";", Shift: true}: func(e *Preditor) error { return insertChar(e, ':') },
	Key{K: "'"}:              func(e *Preditor) error { return insertChar(e, '\'') },
	Key{K: "'", Shift: true}: func(e *Preditor) error { return insertChar(e, '"') },
	Key{K: "\""}:             func(e *Preditor) error { return insertChar(e, '"') },
	Key{K: ","}:              func(e *Preditor) error { return insertChar(e, ',') },
	Key{K: "."}:              func(e *Preditor) error { return insertChar(e, '.') },
	Key{K: ",", Shift: true}: func(e *Preditor) error { return insertChar(e, '<') },
	Key{K: ".", Shift: true}: func(e *Preditor) error { return insertChar(e, '>') },
	Key{K: "/"}:              func(e *Preditor) error { return insertChar(e, '/') },
	Key{K: "/", Shift: true}: func(e *Preditor) error { return insertChar(e, '?') },
	Key{K: "-"}:              func(e *Preditor) error { return insertChar(e, '-') },
	Key{K: "="}:              func(e *Preditor) error { return insertChar(e, '=') },
	Key{K: "-", Shift: true}: func(e *Preditor) error { return insertChar(e, '_') },
	Key{K: "=", Shift: true}: func(e *Preditor) error { return insertChar(e, '+') },
	Key{K: "`"}:              func(e *Preditor) error { return insertChar(e, '`') },
	Key{K: "`", Shift: true}: func(e *Preditor) error { return insertChar(e, '~') },
	Key{K: "<tab>"}:          func(e *Preditor) error { return e.ActiveEditor().Indent() },
}

func insertChar(e *Preditor, char byte) error {
	return e.ActiveEditor().InsertCharAtCursor(char)
}

func getClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func writeToClipboard(bs []byte) {
	clipboard.Write(clipboard.FmtText, bytes.Clone(bs))
}

func insertCharAtSearchString(e *Preditor, char byte) error {

	editor := e.ActiveEditor()
	if editor.SearchString == nil {
		editor.SearchString = new(string)
	}

	*editor.SearchString += string(char)

	return nil
}

func (e *Editor) DeleteCharBackwardFromActiveSearch() error {
	if e.SearchString == nil {
		return nil
	}
	s := []byte(*e.SearchString)
	if len(s) < 1 {
		return nil
	}
	s = s[:len(s)-1]

	e.SearchString = &[]string{string(s)}[0]

	return nil
}

var searchModeKeymap = Keymap{
	Key{K: "<space>"}: func(e *Preditor) error { return insertCharAtSearchString(e, ' ') },
	Key{K: "<backspace>"}: func(e *Preditor) error {
		return e.ActiveEditor().DeleteCharBackwardFromActiveSearch()
	},
	Key{K: "<enter>"}: func(a *Preditor) error {
		editor := a.ActiveEditor()

		editor.CurrentMatch++
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	},

	Key{K: "<enter>", Control: true}: func(a *Preditor) error {
		editor := a.ActiveEditor()

		editor.CurrentMatch--
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		if editor.CurrentMatch < 0 {
			editor.CurrentMatch = len(editor.SearchMatches) - 1
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	},
	Key{K: "<esc>"}: func(a *Preditor) error {
		a.ActiveEditor().Keymaps = a.ActiveEditor().Keymaps[:len(a.ActiveEditor().Keymaps)-1]
		fmt.Println("exiting search mode")
		editor := a.ActiveEditor()
		editor.IsSearching = false
		editor.LastSearchString = ""
		editor.SearchString = nil
		editor.SearchMatches = nil
		editor.CurrentMatch = 0
		editor.MovedAwayFromCurrentMatch = false
		return nil
	},
	Key{K: "<lmouse>-click"}: func(e *Preditor) error {
		return e.ActiveEditor().MoveCursorTo(rl.GetMousePosition())
	},
	Key{K: "<mouse-wheel-up>"}: func(e *Preditor) error {
		e.ActiveEditor().MovedAwayFromCurrentMatch = true
		return e.ActiveEditor().ScrollUp(10)

	},
	Key{K: "<mouse-wheel-down>"}: func(e *Preditor) error {
		e.ActiveEditor().MovedAwayFromCurrentMatch = true

		return e.ActiveEditor().ScrollDown(10)
	},

	Key{K: "<rmouse>-click"}: func(a *Preditor) error {
		editor := a.ActiveEditor()

		editor.CurrentMatch++
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		if editor.CurrentMatch < 0 {
			editor.CurrentMatch = len(editor.SearchMatches) - 1
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	},
	Key{K: "<mmouse>-click"}: func(a *Preditor) error {
		editor := a.ActiveEditor()

		editor.CurrentMatch--
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		if editor.CurrentMatch < 0 {
			editor.CurrentMatch = len(editor.SearchMatches) - 1
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	},
	Key{K: "<pagedown>"}: func(e *Preditor) error {
		e.ActiveEditor().MovedAwayFromCurrentMatch = true
		return e.ActiveEditor().ScrollDown(1)
	},
	Key{K: "<pageup>"}: func(e *Preditor) error {
		e.ActiveEditor().MovedAwayFromCurrentMatch = true

		return e.ActiveEditor().ScrollUp(1)
	},

	Key{K: "a"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'a') },
	Key{K: "b"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'b') },
	Key{K: "c"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'c') },
	Key{K: "d"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'd') },
	Key{K: "e"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'e') },
	Key{K: "f"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'f') },
	Key{K: "g"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'g') },
	Key{K: "h"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'h') },
	Key{K: "i"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'i') },
	Key{K: "j"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'j') },
	Key{K: "k"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'k') },
	Key{K: "l"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'l') },
	Key{K: "m"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'm') },
	Key{K: "n"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'n') },
	Key{K: "o"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'o') },
	Key{K: "p"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'p') },
	Key{K: "q"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'q') },
	Key{K: "r"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'r') },
	Key{K: "s"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 's') },
	Key{K: "t"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 't') },
	Key{K: "u"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'u') },
	Key{K: "v"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'v') },
	Key{K: "w"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'w') },
	Key{K: "x"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'x') },
	Key{K: "y"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'y') },
	Key{K: "z"}:               func(e *Preditor) error { return insertCharAtSearchString(e, 'z') },
	Key{K: "0"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '0') },
	Key{K: "1"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '1') },
	Key{K: "2"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '2') },
	Key{K: "3"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '3') },
	Key{K: "4"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '4') },
	Key{K: "5"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '5') },
	Key{K: "6"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '6') },
	Key{K: "7"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '7') },
	Key{K: "8"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '8') },
	Key{K: "9"}:               func(e *Preditor) error { return insertCharAtSearchString(e, '9') },
	Key{K: "\\"}:              func(e *Preditor) error { return insertCharAtSearchString(e, '\\') },
	Key{K: "\\", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '|') },

	Key{K: "0", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, ')') },
	Key{K: "1", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '!') },
	Key{K: "2", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '@') },
	Key{K: "3", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '#') },
	Key{K: "4", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '$') },
	Key{K: "5", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '%') },
	Key{K: "6", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '^') },
	Key{K: "7", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '&') },
	Key{K: "8", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '*') },
	Key{K: "9", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '(') },
	Key{K: "a", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'A') },
	Key{K: "b", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'B') },
	Key{K: "c", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'C') },
	Key{K: "d", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'D') },
	Key{K: "e", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'E') },
	Key{K: "f", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'F') },
	Key{K: "g", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'G') },
	Key{K: "h", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'H') },
	Key{K: "i", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'I') },
	Key{K: "j", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'J') },
	Key{K: "k", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'K') },
	Key{K: "l", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'L') },
	Key{K: "m", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'M') },
	Key{K: "n", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'N') },
	Key{K: "o", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'O') },
	Key{K: "p", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'P') },
	Key{K: "q", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'Q') },
	Key{K: "r", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'R') },
	Key{K: "s", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'S') },
	Key{K: "t", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'T') },
	Key{K: "u", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'U') },
	Key{K: "v", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'V') },
	Key{K: "w", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'W') },
	Key{K: "x", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'X') },
	Key{K: "y", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'Y') },
	Key{K: "z", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, 'Z') },
	Key{K: "["}:              func(e *Preditor) error { return insertCharAtSearchString(e, '[') },
	Key{K: "]"}:              func(e *Preditor) error { return insertCharAtSearchString(e, ']') },
	Key{K: "[", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '{') },
	Key{K: "]", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '}') },
	Key{K: ";"}:              func(e *Preditor) error { return insertCharAtSearchString(e, ';') },
	Key{K: ";", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, ':') },
	Key{K: "'"}:              func(e *Preditor) error { return insertCharAtSearchString(e, '\'') },
	Key{K: "'", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '"') },
	Key{K: "\""}:             func(e *Preditor) error { return insertCharAtSearchString(e, '"') },
	Key{K: ","}:              func(e *Preditor) error { return insertCharAtSearchString(e, ',') },
	Key{K: "."}:              func(e *Preditor) error { return insertCharAtSearchString(e, '.') },
	Key{K: ",", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '<') },
	Key{K: ".", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '>') },
	Key{K: "/"}:              func(e *Preditor) error { return insertCharAtSearchString(e, '/') },
	Key{K: "/", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '?') },
	Key{K: "-"}:              func(e *Preditor) error { return insertCharAtSearchString(e, '-') },
	Key{K: "="}:              func(e *Preditor) error { return insertCharAtSearchString(e, '=') },
	Key{K: "-", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '_') },
	Key{K: "=", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '+') },
	Key{K: "`"}:              func(e *Preditor) error { return insertCharAtSearchString(e, '`') },
	Key{K: "`", Shift: true}: func(e *Preditor) error { return insertCharAtSearchString(e, '~') },
}
