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

type TextBuffer struct {
	cfg                       *Config
	parent                    *Preditor
	File                      string
	Content                   []byte
	keymaps                   []Keymap
	HasSyntaxHighlights       bool
	SyntaxHighlights          *SyntaxHighlights
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
	State                     int
	HasSelection              bool
	SelectionStart            *Position
	IsSearching               bool
	LastSearchString          string
	SearchString              *string
	SearchMatches             [][2]Position
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool
	LastCursorBlink           time.Time
	BeforeSaveHook            []func(*TextBuffer) error
	UndoStack                 Stack[EditorAction]
}

const (
	EditorActionType_Insert = iota + 1
	EditorActionType_Delete
)

type EditorAction struct {
	Type int
	Idx  int
	Data []byte
}

func (e *TextBuffer) Keymaps() []Keymap {
	return e.keymaps
}
func (e *TextBuffer) AddUndoAction(a EditorAction) {
	e.UndoStack.Push(a)
}

func (e *TextBuffer) PopAndReverseLastAction() {
	last, err := e.UndoStack.Pop()
	if err != nil {
		return
	}

	switch last.Type {
	case EditorActionType_Insert:
		e.Content = append(e.Content[:last.Idx], e.Content[last.Idx+1:]...)
	case EditorActionType_Delete:
		e.Content = append(e.Content[:last.Idx], append(last.Data, e.Content[last.Idx:]...)...)
	}

	e.SetStateDirty()
}

func (e *TextBuffer) SetStateDirty() {
	e.calculateVisualLines()
	e.State = State_Dirty
}

func (e *TextBuffer) SetStateClean() {
	e.State = State_Clean
}

func (e *TextBuffer) replaceTabsWithSpaces() {
	e.Content = bytes.Replace(e.Content, []byte("\t"), []byte(strings.Repeat(" ", e.TabSize)), -1)
}

func (e *TextBuffer) SetMaxWidth(w int32) {
	e.MaxWidth = w
	e.updateMaxLineAndColumn()
}
func (e *TextBuffer) SetMaxHeight(h int32) {
	e.MaxHeight = h
	e.updateMaxLineAndColumn()
}
func (e *TextBuffer) updateMaxLineAndColumn() {
	oldMaxLine := e.maxLine
	charSize := measureTextSize(font, ' ', fontSize, 0)
	e.maxColumn = e.MaxWidth / int32(charSize.X)
	e.maxLine = e.MaxHeight / int32(charSize.Y)

	// reserve one line for status bar
	e.maxLine--

	diff := e.maxLine - oldMaxLine
	e.VisibleEnd += diff
}
func (e *TextBuffer) Type() string {
	return "text_editor_buffer"
}

const (
	CURSOR_SHAPE_BLOCK   = 1
	CURSOR_SHAPE_OUTLINE = 2
	CURSOR_SHAPE_LINE    = 3
)

func NewTextBuffer(parent *Preditor, cfg *Config, filename string, maxH int32, maxW int32, zeroPosition rl.Vector2) (*TextBuffer, error) {
	t := TextBuffer{cfg: cfg}
	t.parent = parent
	t.File = filename
	t.MaxHeight = maxH
	t.MaxWidth = maxW
	t.ZeroPosition = zeroPosition
	t.keymaps = append([]Keymap{}, editorKeymap)
	var err error
	if t.File != "" {
		if _, err = os.Stat(t.File); err == nil {
			t.Content, err = os.ReadFile(t.File)
			if err != nil {
				return nil, err
			}
		}

		fileType, exists := fileTypeMappings[path.Ext(t.File)]
		if exists {
			t.BeforeSaveHook = append(t.BeforeSaveHook, fileType.BeforeSave)
			t.SyntaxHighlights = fileType.SyntaxHighlights
			t.HasSyntaxHighlights = fileType.SyntaxHighlights != nil
			t.TabSize = fileType.TabSize
		}
	}

	t.replaceTabsWithSpaces()
	t.updateMaxLineAndColumn()
	t.calculateVisualLines()
	return &t, nil

}

func (e *TextBuffer) Destroy() error {
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
}

func (e *TextBuffer) calculateHighlights(bs []byte, offset int) []highlight {
	if !e.HasSyntaxHighlights {
		return nil
	}
	var highlights []highlight
	//keywords
	indexes := e.SyntaxHighlights.Keywords.Regex.FindAllStringIndex(string(bs), -1)
	for _, index := range indexes {
		highlights = append(highlights, highlight{
			start: index[0] + offset,
			end:   index[1] + offset - 1,
			Color: e.SyntaxHighlights.Keywords.Color,
		})
	}
	//types
	indexesTypes := e.SyntaxHighlights.Types.Regex.FindAllStringIndex(string(bs), -1)
	for _, index := range indexesTypes {
		highlights = append(highlights, highlight{
			start: index[0] + offset,
			end:   index[1] + offset - 1,
			Color: e.SyntaxHighlights.Types.Color,
		})
	}
	return highlights
}
func sortme[T any](slice []T, pred func(t1 T, t2 T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return pred(slice[i], slice[j])
	})
}
func (e *TextBuffer) fillInTheBlanks(hs []highlight, start, end int) []highlight {
	var missing []highlight
	sortme[highlight](hs, func(t1, t2 highlight) bool {
		return t1.start < t2.start
	})
	if len(hs) == 0 {
		missing = append(missing, highlight{
			start: start,
			end:   end,
			Color: e.cfg.Colors.Foreground,
		})
	} else {
		for i, h := range hs {
			if i == 0 {
				if h.start != start {
					missing = append(missing, highlight{
						start: start,
						end:   h.start - 1,
						Color: e.cfg.Colors.Foreground,
					})
				}
			}
			if i == len(hs)-1 && h.end != end {
				missing = append(missing, highlight{
					start: h.end + 1,
					end:   end,
					Color: e.cfg.Colors.Foreground,
				})
			}
			if i+1 < len(hs) && hs[i+1].start-h.end != 1 {
				missing = append(missing, highlight{
					start: h.end + 1,
					end:   hs[i+1].start - 1,
					Color: e.cfg.Colors.Foreground,
				})
			}
		}
	}

	hs = append(hs, missing...)
	sortme[highlight](hs, func(t1, t2 highlight) bool {
		return t1.start < t2.start
	})

	return hs
}

func (e *TextBuffer) calculateVisualLines() {
	e.visualLines = []visualLine{}
	totalVisualLines := 0
	lineCharCounter := 0
	var actualLineIndex int
	var start int
	if e.VisibleEnd == 0 {
		e.VisibleEnd = e.maxLine
	}

	for idx, char := range e.Content {
		lineCharCounter++
		if char == '\n' {
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start,
				ActualLine: actualLineIndex,
			}
			e.visualLines = append(e.visualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1
			continue
		}
		if idx >= len(e.Content)-1 {
			// last index
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start,
				ActualLine: actualLineIndex,
			}
			e.visualLines = append(e.visualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1
			continue
		}

		if int32(lineCharCounter) > e.maxColumn {
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start,
				ActualLine: actualLineIndex,
			}

			e.visualLines = append(e.visualLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx
			continue
		}
	}
}

func (e *TextBuffer) renderCursor() {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	if e.cfg.CursorBlinking && time.Since(e.LastCursorBlink).Milliseconds() < 1000 {
		return
	}
	cursorView := Position{
		Line:   e.Cursor.Line - int(e.VisibleStart),
		Column: e.Cursor.Column,
	}
	posX := int32(cursorView.Column)*int32(charSize.X) + int32(e.ZeroPosition.X)
	if e.cfg.LineNumbers {
		if len(e.visualLines) > e.Cursor.Line {
			posX += int32((len(fmt.Sprint(e.visualLines[e.Cursor.Line].ActualLine)) + 1) * int(charSize.X))
		} else {
			posX += int32(charSize.X)

		}
	}
	switch e.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(e.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), e.cfg.Colors.Cursor)
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(e.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.Colors.Cursor, 0.6))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(e.ZeroPosition.Y), 2, int32(charSize.Y), e.cfg.Colors.Cursor)
	}

	rl.DrawRectangle(0, int32(cursorView.Line)*int32(charSize.Y)+int32(e.ZeroPosition.Y), e.maxColumn*int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.Colors.CursorLineBackground, 0.2))

	e.LastCursorBlink = time.Now()
}

func (e *TextBuffer) renderStatusBar() {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	//render status bar
	rl.DrawRectangle(
		int32(e.ZeroPosition.X),
		e.maxLine*int32(charSize.Y),
		e.MaxWidth,
		int32(charSize.Y),
		e.cfg.Colors.StatusBarBackground,
	)
	file := e.File
	if file == "" {
		file = "[scratch]"
	}
	var state string
	if e.State == State_Dirty {
		state = "**"
	} else {
		state = "--"
	}
	var line int
	if len(e.visualLines) > e.Cursor.Line {
		line = e.visualLines[e.Cursor.Line].ActualLine
	} else {
		line = 0
	}
	var searchString string
	if e.SearchString != nil {
		searchString = fmt.Sprintf("Searching: \"%s\" %d of %d matches", *e.SearchString, e.CurrentMatch, len(e.SearchMatches)-1)
	}

	rl.DrawTextEx(font,
		fmt.Sprintf("%s %s %d:%d %s", state, file, line, e.Cursor.Column, searchString),
		rl.Vector2{X: e.ZeroPosition.X, Y: float32(e.maxLine) * charSize.Y},
		fontSize,
		0,
		e.cfg.Colors.StatusBarForeground)
}

func (e *TextBuffer) highilightBetweenTwoPositions(start Position, end Position, color color.RGBA) {
	charSize := measureTextSize(font, ' ', fontSize, 0)

	for i := start.Line; i <= end.Line; i++ {
		if len(e.visualLines) <= i {
			break
		}
		var thisLineEnd int
		var thisLineStart int
		line := e.visualLines[i]
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
			posX := int32(j)*int32(charSize.X) + int32(e.ZeroPosition.X)
			if e.cfg.LineNumbers {
				if len(e.visualLines) > i {
					posX += int32((len(fmt.Sprint(e.visualLines[i].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			rl.DrawRectangle(posX, int32(i-int(e.VisibleStart))*int32(charSize.Y)+int32(e.ZeroPosition.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(color, 0.5))
		}
	}

}

func (e *TextBuffer) renderSelection() {
	if e.SelectionStart == nil {
		return
	}

	var startLine int
	var startColumn int
	var endLine int
	var endColumn int
	switch {
	case e.SelectionStart.Line < e.Cursor.Line:
		startLine = e.SelectionStart.Line
		startColumn = e.SelectionStart.Column
		endLine = e.Cursor.Line
		endColumn = e.Cursor.Column
	case e.Cursor.Line < e.SelectionStart.Line:
		startLine = e.Cursor.Line
		startColumn = e.Cursor.Column
		endLine = e.SelectionStart.Line
		endColumn = e.SelectionStart.Column
	case e.Cursor.Line == e.SelectionStart.Line:
		startLine = e.Cursor.Line
		endLine = e.Cursor.Line
		if e.SelectionStart.Column > e.Cursor.Column {
			startColumn = e.Cursor.Column
			endColumn = e.SelectionStart.Column
		} else {
			startColumn = e.SelectionStart.Column
			endColumn = e.Cursor.Column
		}

	}

	e.highilightBetweenTwoPositions(Position{
		Line:   startLine,
		Column: startColumn,
	}, Position{
		Line:   endLine,
		Column: endColumn,
	}, e.cfg.Colors.Selection)

}

func (e *TextBuffer) renderText() {
	var visibleLines []visualLine
	if e.VisibleEnd > int32(len(e.visualLines)) {
		visibleLines = e.visualLines[e.VisibleStart:]
	} else {
		visibleLines = e.visualLines[e.VisibleStart:e.VisibleEnd]
	}
	for idx, line := range visibleLines {
		if e.visualLineShouldBeRendered(line) {
			charSize := measureTextSize(font, ' ', fontSize, 0)
			var lineNumberWidth int
			if e.cfg.LineNumbers {
				lineNumberWidth = (len(fmt.Sprint(line.ActualLine)) + 1) * int(charSize.X)
				rl.DrawTextEx(font,
					fmt.Sprintf("%d", line.ActualLine),
					rl.Vector2{X: e.ZeroPosition.X, Y: float32(idx) * charSize.Y},
					fontSize,
					0,
					e.cfg.Colors.LineNumbersForeground)

			}

			if e.cfg.EnableSyntaxHighlighting && e.HasSyntaxHighlights {
				highlights := e.fillInTheBlanks(e.calculateHighlights(e.Content[line.startIndex:line.endIndex], line.startIndex), line.startIndex, line.endIndex)

				for _, h := range highlights {

					rl.DrawTextEx(font,
						string(e.Content[h.start:h.end+1]),
						rl.Vector2{X: e.ZeroPosition.X + float32(lineNumberWidth) + float32(h.start-line.startIndex)*charSize.X, Y: float32(idx) * charSize.Y},
						fontSize,
						0,
						h.Color)
				}
			} else {

				rl.DrawTextEx(font,
					string(e.Content[line.startIndex:line.endIndex+1]),
					rl.Vector2{X: e.ZeroPosition.X + float32(lineNumberWidth), Y: float32(idx) * charSize.Y},
					fontSize,
					0,
					e.cfg.Colors.Foreground)
			}
		}
	}
}
func (e *TextBuffer) convertBufferIndexToLineAndColumn(idx int) *Position {
	for lineIndex, line := range e.visualLines {
		if line.startIndex <= idx && line.endIndex >= idx {
			return &Position{
				Line:   lineIndex,
				Column: idx - line.startIndex,
			}
		}
	}

	return nil
}
func (e *TextBuffer) findMatchesRegex(pattern string) error {
	e.SearchMatches = [][2]Position{}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matches := re.FindAllStringIndex(string(e.Content), -1)
	for _, match := range matches {
		matchStart := e.convertBufferIndexToLineAndColumn(match[0])
		matchEnd := e.convertBufferIndexToLineAndColumn(match[1])
		if matchStart == nil || matchEnd == nil {
			continue
		}
		matchEnd.Column--
		e.SearchMatches = append(e.SearchMatches, [2]Position{*matchStart, *matchEnd})
	}

	return nil
}

func (e *TextBuffer) findMatchesAndHighlight(pattern string) error {
	if pattern != e.LastSearchString {
		if err := e.findMatchesRegex(pattern); err != nil {
			return err
		}
	}
	for idx, match := range e.SearchMatches {
		c := e.cfg.Colors.Selection
		if idx == e.CurrentMatch {
			c = rl.Fade(rl.Red, 0.5)
			if !(e.VisibleStart < int32(match[0].Line) && e.VisibleEnd > int32(match[1].Line)) && !e.MovedAwayFromCurrentMatch {
				// current match is not in view
				// move the view
				oldStart := e.VisibleStart
				e.VisibleStart = int32(match[0].Line) - e.maxLine/2
				if e.VisibleStart < 0 {
					e.VisibleStart = int32(match[0].Line)
				}

				diff := e.VisibleStart - oldStart
				e.VisibleEnd += diff
			}
		}
		e.highilightBetweenTwoPositions(match[0], match[1], c)
	}
	e.LastSearchString = pattern

	return nil
}
func (e *TextBuffer) renderSearchResults() {
	if e.SearchString == nil || len(*e.SearchString) < 1 {
		return
	}
	e.findMatchesAndHighlight(*e.SearchString)
}

func (e *TextBuffer) Render() {
	e.renderText()
	e.renderSearchResults()
	e.renderCursor()
	e.renderStatusBar()
	e.renderSelection()
}

func (e *TextBuffer) visualLineShouldBeRendered(line visualLine) bool {
	if e.VisibleStart <= int32(line.Index) && line.Index <= int(e.VisibleEnd) {
		return true
	}

	return false
}

func (e *TextBuffer) positionToBufferIndex(pos Position) int {
	if pos.Line >= len(e.visualLines) {
		return 0
	}
	return e.visualLines[pos.Line].startIndex + pos.Column
}

func (e *TextBuffer) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if newPosition.Line >= len(e.visualLines) {
		return false
	}
	if newPosition.Column < 0 {
		return false
	}
	if newPosition.Column > e.visualLines[newPosition.Line].Length+1 {
		return false
	}

	return true
}

func (e *TextBuffer) InsertCharAtCursor(char byte) error {
	var idx int
	if e.HasSelection {
		var startLine int
		var startColumn int
		var endLine int
		var endColumn int
		switch {
		case e.SelectionStart.Line < e.Cursor.Line:
			startLine = e.SelectionStart.Line
			startColumn = e.SelectionStart.Column
			endLine = e.Cursor.Line
			endColumn = e.Cursor.Column
		case e.Cursor.Line < e.SelectionStart.Line:
			startLine = e.Cursor.Line
			startColumn = e.Cursor.Column
			endLine = e.SelectionStart.Line
			endColumn = e.SelectionStart.Column
		case e.Cursor.Line == e.SelectionStart.Line:
			startLine = e.Cursor.Line
			endLine = e.Cursor.Line
			if e.SelectionStart.Column > e.Cursor.Column {
				startColumn = e.Cursor.Column
				endColumn = e.SelectionStart.Column
			} else {
				startColumn = e.SelectionStart.Column
				endColumn = e.Cursor.Column
			}
			e.HasSelection = false
			e.SelectionStart = nil
		}
		idx = e.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		startIndex := e.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		endIndex := e.positionToBufferIndex(Position{Line: endLine, Column: endColumn})

		e.Content = append(e.Content[:startIndex], e.Content[endIndex+1:]...)
		e.Cursor = Position{Line: startLine, Column: startColumn}
	} else {
		idx = e.positionToBufferIndex(e.Cursor)
	}
	e.AddUndoAction(EditorAction{
		Type: EditorActionType_Insert,
		Idx:  idx,
		Data: []byte{char},
	})
	if idx >= len(e.Content) { // end of file, appending
		e.Content = append(e.Content, char)
	} else {
		e.Content = append(e.Content[:idx+1], e.Content[idx:]...)
		e.Content[idx] = char
	}
	e.SetStateDirty()

	if char == '\n' {
		e.CursorDown()
		e.BeginingOfTheLine()
	} else {
		e.CursorRight(1)
	}
	return nil
}

func (e *TextBuffer) DeleteCharBackward() error {
	if e.HasSelection {
		var startLine int
		var startColumn int
		var endLine int
		var endColumn int
		switch {
		case e.SelectionStart.Line < e.Cursor.Line:
			startLine = e.SelectionStart.Line
			startColumn = e.SelectionStart.Column
			endLine = e.Cursor.Line
			endColumn = e.Cursor.Column
		case e.Cursor.Line < e.SelectionStart.Line:
			startLine = e.Cursor.Line
			startColumn = e.Cursor.Column
			endLine = e.SelectionStart.Line
			endColumn = e.SelectionStart.Column
		case e.Cursor.Line == e.SelectionStart.Line:
			startLine = e.Cursor.Line
			endLine = e.Cursor.Line
			if e.SelectionStart.Column > e.Cursor.Column {
				startColumn = e.Cursor.Column
				endColumn = e.SelectionStart.Column
			} else {
				startColumn = e.SelectionStart.Column
				endColumn = e.Cursor.Column
			}
			e.HasSelection = false
			e.SelectionStart = nil
		}
		startIndex := e.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		endIndex := e.positionToBufferIndex(Position{Line: endLine, Column: endColumn})
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Data: e.Content[startIndex:endIndex],
		})
		e.Content = append(e.Content[:startIndex], e.Content[endIndex+1:]...)
		e.Cursor = Position{Line: startLine, Column: startColumn}
		e.SetStateDirty()
		e.HasSelection = false
		return nil
	}

	idx := e.positionToBufferIndex(e.Cursor)
	if idx <= 0 {
		return nil
	}
	if len(e.Content) <= idx {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Data: []byte{e.Content[idx]},
		})
		e.Content = e.Content[:idx]
	} else {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Data: []byte{e.Content[idx]},
		})
		e.Content = append(e.Content[:idx-1], e.Content[idx:]...)
	}
	e.SetStateDirty()
	e.CursorLeft()

	return nil

}

func (e *TextBuffer) DeleteCharForward() error {
	idx := e.positionToBufferIndex(e.Cursor)
	if idx < 0 || e.Cursor.Column < 0 || len(e.Content) <= idx {
		return nil
	}

	e.AddUndoAction(EditorAction{
		Type: EditorActionType_Delete,
		Data: []byte{e.Content[idx]},
		Idx:  idx,
	})
	e.Content = append(e.Content[:idx], e.Content[idx+1:]...)
	e.SetStateDirty()

	return nil
}

func (e *TextBuffer) ScrollUp(n int) error {
	if e.VisibleStart <= 0 {
		return nil
	}
	e.VisibleEnd += int32(-1 * n)
	e.VisibleStart += int32(-1 * n)

	diff := e.VisibleEnd - e.VisibleStart

	if e.VisibleStart < 0 {
		e.VisibleStart = 0
		e.VisibleEnd = diff
	}

	return nil

}

func (e *TextBuffer) ScrollDown(n int) error {
	if int(e.VisibleEnd) >= len(e.visualLines) {
		return nil
	}
	e.VisibleEnd += int32(n)
	e.VisibleStart += int32(n)
	diff := e.VisibleEnd - e.VisibleStart
	if int(e.VisibleEnd) >= len(e.visualLines) {
		e.VisibleEnd = int32(len(e.visualLines) - 1)
		e.VisibleStart = e.VisibleEnd - diff
	}

	return nil

}

func (e *TextBuffer) CursorLeft() error {
	newPosition := e.Cursor
	newPosition.Column--
	if e.Cursor.Column <= 0 {
		if newPosition.Line > 0 {
			newPosition.Line--
			lineColumns := e.visualLines[newPosition.Line].Length
			if lineColumns <= 0 {
				lineColumns = 0
			}
			newPosition.Column = lineColumns
		}

	}

	if e.isValidCursorPosition(newPosition) {
		e.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}

	return nil

}

func (e *TextBuffer) CursorRight(n int) error {
	newPosition := e.Cursor
	newPosition.Column += n
	if e.Cursor.Line == len(e.visualLines) {
		return nil
	}

	if int32(newPosition.Column) > e.maxColumn {
		newPosition.Line++
		newPosition.Column = 0
	}

	if e.isValidCursorPosition(newPosition) {
		e.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}
	return nil

}

func (e *TextBuffer) CursorUp() error {
	newPosition := e.Cursor
	newPosition.Line--

	if newPosition.Line < 0 {
		newPosition.Line = 0
	}

	if newPosition.Column > e.visualLines[newPosition.Line].Length {
		newPosition.Column = e.visualLines[newPosition.Line].Length
	}
	if e.isValidCursorPosition(newPosition) {
		e.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}

	return nil

}

func (e *TextBuffer) CursorDown() error {
	newPosition := e.Cursor
	newPosition.Line++

	// check if cursor should be moved back
	if newPosition.Line < len(e.visualLines) {
		if newPosition.Column > e.visualLines[newPosition.Line].Length {
			newPosition.Column = e.visualLines[newPosition.Line].Length
		}
	} else {
		newPosition.Column = 0
	}

	if e.isValidCursorPosition(newPosition) {
		e.MoveCursorToPositionAndScrollIfNeeded(newPosition)

	}
	return nil

}

func (e *TextBuffer) BeginingOfTheLine() error {
	newPosition := e.Cursor
	newPosition.Column = 0
	if e.isValidCursorPosition(newPosition) {
		e.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (e *TextBuffer) EndOfTheLine() error {
	newPosition := e.Cursor
	newPosition.Column = e.visualLines[e.Cursor.Line].Length
	if e.isValidCursorPosition(newPosition) {
		e.MoveCursorToPositionAndScrollIfNeeded(newPosition)
	}
	return nil

}

func (e *TextBuffer) PreviousLine() error {
	return e.CursorUp()
}

func (e *TextBuffer) NextLine() error {
	return e.CursorDown()
}
func (e *TextBuffer) indexOfFirstNonLetter(bs []byte) int {

	for idx, b := range bs {
		if !unicode.IsLetter(rune(b)) {
			return idx
		}
	}

	return -1
}

func (e *TextBuffer) NextWord() error {
	currentidx := e.positionToBufferIndex(e.Cursor)
	newidx := nextWordInBuffer(e.Content, currentidx)
	pos := e.convertBufferIndexToLineAndColumn(newidx)

	if e.isValidCursorPosition(*pos) {
		return e.MoveCursorToPositionAndScrollIfNeeded(*pos)
	}
	return nil
}

func (e *TextBuffer) PreviousWord() error {
	currentidx := e.positionToBufferIndex(e.Cursor)
	newidx := previousWordInBuffer(e.Content, currentidx)
	pos := e.convertBufferIndexToLineAndColumn(newidx)

	if e.isValidCursorPosition(*pos) {
		return e.MoveCursorToPositionAndScrollIfNeeded(*pos)
	}

	return nil
}

func (e *TextBuffer) MoveCursorTo(pos rl.Vector2) error {
	charSize := measureTextSize(font, ' ', fontSize, 0)
	oldPos := e.Cursor

	apprLine := pos.Y / charSize.Y
	apprColumn := pos.X / charSize.X
	if e.cfg.LineNumbers {
		var line int
		if len(e.visualLines) > e.Cursor.Line {
			line = e.visualLines[e.Cursor.Line].ActualLine
		}
		apprColumn -= float32((len(fmt.Sprint(line)) + 1))

	}

	if len(e.visualLines) < 1 {
		return nil
	}

	e.Cursor.Line = int(apprLine) + int(e.VisibleStart)
	e.Cursor.Column = int(apprColumn)

	if e.Cursor.Line >= len(e.visualLines) {
		e.Cursor.Line = len(e.visualLines) - 1
	}

	if e.Cursor.Line < 0 {
		e.Cursor.Line = 0
	}

	// check if cursor should be moved back
	if e.Cursor.Column > e.visualLines[e.Cursor.Line].Length {
		e.Cursor.Column = e.visualLines[e.Cursor.Line].Length
	}
	if !e.isValidCursorPosition(e.Cursor) {
		e.Cursor = oldPos
		return nil
	}

	return nil
}

func (e *TextBuffer) MoveCursorToPositionAndScrollIfNeeded(pos Position) error {
	if !e.isValidCursorPosition(pos) {
		return nil
	}
	e.Cursor = pos
	if e.Cursor.Line == int(e.VisibleStart-1) {
		jump := int(e.maxLine / 2)
		e.ScrollUp(jump)
	}

	if e.Cursor.Line == int(e.VisibleEnd)+1 {
		jump := int(e.maxLine / 2)
		e.ScrollDown(jump)
	}

	return nil
}

func (e *TextBuffer) Write() error {
	if e.File == "" {
		return nil
	}
	if e.TabSize != 0 {
		e.Content = bytes.Replace(e.Content, []byte(strings.Repeat(" ", e.TabSize)), []byte("\t"), -1)
	}

	for _, hook := range e.BeforeSaveHook {
		if err := hook(e); err != nil {
			return err
		}
	}

	if err := os.WriteFile(e.File, e.Content, 0644); err != nil {
		return err
	}
	e.SetStateClean()
	e.replaceTabsWithSpaces()
	e.calculateVisualLines()
	return nil
}

func (e *TextBuffer) GetMaxHeight() int32 { return e.MaxHeight }
func (e *TextBuffer) GetMaxWidth() int32  { return e.MaxWidth }
func (e *TextBuffer) Indent() error {
	idx := e.positionToBufferIndex(e.Cursor)
	if idx >= len(e.Content) { // end of file, appending
		e.Content = append(e.Content, []byte(strings.Repeat(" ", e.TabSize))...)
	} else {
		e.Content = append(e.Content[:idx], append([]byte(strings.Repeat(" ", e.TabSize)), e.Content[idx:]...)...)
	}

	e.SetStateDirty()
	return nil
}

func (e *TextBuffer) copy() error {
	if e.HasSelection {
		// copy selection
		selection := e.positionToBufferIndex(*e.SelectionStart)
		cursor := e.positionToBufferIndex(e.Cursor)
		switch {
		case selection < cursor:
			writeToClipboard(e.Content[selection:cursor])
		case selection > cursor:
			writeToClipboard(e.Content[cursor:selection])
		case cursor == selection:
			return nil
		}
	} else {
		writeToClipboard(e.Content[e.visualLines[e.Cursor.Line].startIndex:e.visualLines[e.Cursor.Line].endIndex])
	}

	return nil
}
func (e *TextBuffer) killLine() error {
	var startIndex int
	var endIndex int
	if e.HasSelection {
		// copy selection
		selection := e.positionToBufferIndex(*e.SelectionStart)
		cursor := e.positionToBufferIndex(e.Cursor)
		switch {
		case selection < cursor:
			writeToClipboard(e.Content[selection:cursor])
			startIndex = selection
			endIndex = cursor
		case selection > cursor:
			writeToClipboard(e.Content[cursor:selection])
			startIndex = cursor
			endIndex = selection
		case cursor == selection:
			return nil
		}
		e.HasSelection = false
		e.SelectionStart = nil
	} else {
		writeToClipboard(e.Content[e.visualLines[e.Cursor.Line].startIndex:e.visualLines[e.Cursor.Line].endIndex])
		startIndex = e.Cursor.Column + e.visualLines[e.Cursor.Line].startIndex
		endIndex = e.visualLines[e.Cursor.Line].endIndex
	}
	e.AddUndoAction(EditorAction{
		Type: EditorActionType_Delete,
		Data: e.Content[startIndex:endIndex],
	})
	e.Content = append(e.Content[:startIndex], e.Content[endIndex+1:]...)

	e.SetStateDirty()

	return nil
}
func (e *TextBuffer) cut() error {
	var startIndex int
	var endIndex int
	if e.HasSelection {
		// copy selection
		selection := e.positionToBufferIndex(*e.SelectionStart)
		cursor := e.positionToBufferIndex(e.Cursor)
		switch {
		case selection < cursor:
			writeToClipboard(e.Content[selection:cursor])
			startIndex = selection
			endIndex = cursor
		case selection > cursor:
			writeToClipboard(e.Content[cursor:selection])
			startIndex = cursor
			endIndex = selection
		case cursor == selection:
			return nil
		}
		e.HasSelection = false
		e.SelectionStart = nil
	} else {
		writeToClipboard(e.Content[e.visualLines[e.Cursor.Line].startIndex:e.visualLines[e.Cursor.Line].endIndex])
		startIndex = e.visualLines[e.Cursor.Line].startIndex
		endIndex = e.visualLines[e.Cursor.Line].endIndex
	}
	e.AddUndoAction(EditorAction{
		Type: EditorActionType_Delete,
		Data: e.Content[startIndex:endIndex],
	})
	e.Content = append(e.Content[:startIndex], e.Content[endIndex+1:]...)

	e.SetStateDirty()

	return nil
}
func (e *TextBuffer) paste() error {
	var idx int
	if e.HasSelection {
		var startLine int
		var startColumn int
		var endLine int
		var endColumn int
		switch {
		case e.SelectionStart.Line < e.Cursor.Line:
			startLine = e.SelectionStart.Line
			startColumn = e.SelectionStart.Column
			endLine = e.Cursor.Line
			endColumn = e.Cursor.Column
		case e.Cursor.Line < e.SelectionStart.Line:
			startLine = e.Cursor.Line
			startColumn = e.Cursor.Column
			endLine = e.SelectionStart.Line
			endColumn = e.SelectionStart.Column
		case e.Cursor.Line == e.SelectionStart.Line:
			startLine = e.Cursor.Line
			endLine = e.Cursor.Line
			if e.SelectionStart.Column > e.Cursor.Column {
				startColumn = e.Cursor.Column
				endColumn = e.SelectionStart.Column
			} else {
				startColumn = e.SelectionStart.Column
				endColumn = e.Cursor.Column
			}
			e.HasSelection = false
			e.SelectionStart = nil
		}
		idx = e.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		startIndex := e.positionToBufferIndex(Position{Line: startLine, Column: startColumn})
		endIndex := e.positionToBufferIndex(Position{Line: endLine, Column: endColumn})

		e.Content = append(e.Content[:startIndex], e.Content[endIndex+1:]...)
		e.Cursor = Position{Line: startLine, Column: startColumn}
	} else {
		idx = e.positionToBufferIndex(e.Cursor)
	}
	contentToPaste := getClipboardContent()
	if idx >= len(e.Content) { // end of file, appending
		e.Content = append(e.Content, contentToPaste...)
	} else {
		e.Content = append(e.Content[:idx], append(contentToPaste, e.Content[idx:]...)...)
	}
	e.SetStateDirty()
	e.HasSelection = false
	e.SelectionStart = nil

	e.CursorRight(len(contentToPaste))
	return nil
}

func (e *TextBuffer) openFileBuffer() {
	dir := path.Dir(e.File)
	ofb := NewOpenFileBuffer(e.parent, e.cfg, dir, e.MaxHeight, e.MaxWidth, e.ZeroPosition)

	e.parent.Windows = append(e.parent.Windows, ofb)
	e.parent.ActiveWindowIndex = len(e.parent.Windows) - 1
}

func makeCommand(f func(e *TextBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveWindow().(*TextBuffer))
	}
}

var editorKeymap = Keymap{
	Key{K: "/", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.PopAndReverseLastAction()
		return nil
	}),
	Key{K: "f", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.CursorRight(1)
	}),
	Key{K: "x", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.cut()
	}),
	Key{K: "v", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.paste()
	}),
	Key{K: "k", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.killLine()
	}),

	Key{K: "c", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.copy()
	}),

	Key{K: "s", Control: true}: makeCommand(func(a *TextBuffer) error {
		a.keymaps = append(a.keymaps, searchModeKeymap)
		return nil
	}),
	Key{K: "x", Alt: true}: makeCommand(func(a *TextBuffer) error {
		return a.Write()
	}),
	Key{K: "o", Control: true}: makeCommand(func(a *TextBuffer) error {
		a.openFileBuffer()

		return nil
	}),
	Key{K: "<esc>"}: makeCommand(func(p *TextBuffer) error {
		if p.HasSelection {
			p.HasSelection = !p.HasSelection
			p.SelectionStart = nil
		}

		return nil
	}),

	//selection
	Key{K: "<space>", Control: true}: makeCommand(func(editor *TextBuffer) error {
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
	}),

	// navigation
	Key{K: "<lmouse>-click"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),
	Key{K: "<lmouse>-hold"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),

	Key{K: "<lmouse>-click", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(20)
	}),

	Key{K: "<lmouse>-hold", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(20)
	}),
	Key{K: "<rmouse>-click", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(20)
	}),

	Key{K: "<rmouse>-hold", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(20)
	}),
	Key{K: "<mouse-wheel-up>"}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(20)

	}),
	Key{K: "<mouse-wheel-down>"}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(20)
	}),

	Key{K: "<rmouse>-click"}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(20)
	}),
	Key{K: "<mmouse>-click"}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(20)
	}),

	Key{K: "a", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.BeginingOfTheLine()
	}),
	Key{K: "e", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.EndOfTheLine()
	}),

	Key{K: "p", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.PreviousLine()
	}),

	Key{K: "n", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.NextLine()
	}),

	Key{K: "<up>"}: makeCommand(func(e *TextBuffer) error {
		return e.CursorUp()
	}),
	Key{K: "<down>"}: makeCommand(func(e *TextBuffer) error {
		return e.CursorDown()
	}),
	Key{K: "<right>"}: makeCommand(func(e *TextBuffer) error {
		return e.CursorRight(1)
	}),
	Key{K: "<right>", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.NextWord()
	}),
	Key{K: "<left>"}: makeCommand(func(e *TextBuffer) error {
		return e.CursorLeft()
	}),
	Key{K: "<left>", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.PreviousWord()
	}),

	Key{K: "b", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.CursorLeft()
	}),
	Key{K: "<home>"}: makeCommand(func(e *TextBuffer) error {
		return e.BeginingOfTheLine()
	}),
	Key{K: "<pagedown>"}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(1)
	}),
	Key{K: "<pageup>"}: makeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(1)
	}),

	//insertion
	Key{K: "<enter>"}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '\n') }),
	Key{K: "<space>"}: makeCommand(func(e *TextBuffer) error { return insertChar(e, ' ') }),
	Key{K: "<backspace>"}: makeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackward()
	}),
	Key{K: "d", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),
	Key{K: "d", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),
	Key{K: "<delete>"}: makeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),
	Key{K: "a"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'a') }),
	Key{K: "b"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'b') }),
	Key{K: "c"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'c') }),
	Key{K: "d"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'd') }),
	Key{K: "e"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'e') }),
	Key{K: "f"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'f') }),
	Key{K: "g"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'g') }),
	Key{K: "h"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'h') }),
	Key{K: "i"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'i') }),
	Key{K: "j"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'j') }),
	Key{K: "k"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'k') }),
	Key{K: "l"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'l') }),
	Key{K: "m"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'm') }),
	Key{K: "n"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'n') }),
	Key{K: "o"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'o') }),
	Key{K: "p"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'p') }),
	Key{K: "q"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'q') }),
	Key{K: "r"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'r') }),
	Key{K: "s"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 's') }),
	Key{K: "t"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 't') }),
	Key{K: "u"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'u') }),
	Key{K: "v"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'v') }),
	Key{K: "w"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'w') }),
	Key{K: "x"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'x') }),
	Key{K: "y"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'y') }),
	Key{K: "z"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, 'z') }),
	Key{K: "0"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '0') }),
	Key{K: "1"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '1') }),
	Key{K: "2"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '2') }),
	Key{K: "3"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '3') }),
	Key{K: "4"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '4') }),
	Key{K: "5"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '5') }),
	Key{K: "6"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '6') }),
	Key{K: "7"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '7') }),
	Key{K: "8"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '8') }),
	Key{K: "9"}:               makeCommand(func(e *TextBuffer) error { return insertChar(e, '9') }),
	Key{K: "\\"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '\\') }),
	Key{K: "\\", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '|') }),

	Key{K: "0", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, ')') }),
	Key{K: "1", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '!') }),
	Key{K: "2", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '@') }),
	Key{K: "3", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '#') }),
	Key{K: "4", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '$') }),
	Key{K: "5", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '%') }),
	Key{K: "6", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '^') }),
	Key{K: "7", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '&') }),
	Key{K: "8", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '*') }),
	Key{K: "9", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '(') }),
	Key{K: "a", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'A') }),
	Key{K: "b", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'B') }),
	Key{K: "c", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'C') }),
	Key{K: "d", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'D') }),
	Key{K: "e", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'E') }),
	Key{K: "f", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'F') }),
	Key{K: "g", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'G') }),
	Key{K: "h", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'H') }),
	Key{K: "i", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'I') }),
	Key{K: "j", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'J') }),
	Key{K: "k", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'K') }),
	Key{K: "l", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'L') }),
	Key{K: "m", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'M') }),
	Key{K: "n", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'N') }),
	Key{K: "o", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'O') }),
	Key{K: "p", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'P') }),
	Key{K: "q", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'Q') }),
	Key{K: "r", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'R') }),
	Key{K: "s", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'S') }),
	Key{K: "t", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'T') }),
	Key{K: "u", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'U') }),
	Key{K: "v", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'V') }),
	Key{K: "w", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'W') }),
	Key{K: "x", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'X') }),
	Key{K: "y", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'Y') }),
	Key{K: "z", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, 'Z') }),
	Key{K: "["}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '[') }),
	Key{K: "]"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, ']') }),
	Key{K: "[", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '{') }),
	Key{K: "]", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '}') }),
	Key{K: ";"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, ';') }),
	Key{K: ";", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, ':') }),
	Key{K: "'"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '\'') }),
	Key{K: "'", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '"') }),
	Key{K: "\""}:             makeCommand(func(e *TextBuffer) error { return insertChar(e, '"') }),
	Key{K: ","}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, ',') }),
	Key{K: "."}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '.') }),
	Key{K: ",", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '<') }),
	Key{K: ".", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '>') }),
	Key{K: "/"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '/') }),
	Key{K: "/", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '?') }),
	Key{K: "-"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '-') }),
	Key{K: "="}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '=') }),
	Key{K: "-", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '_') }),
	Key{K: "=", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '+') }),
	Key{K: "`"}:              makeCommand(func(e *TextBuffer) error { return insertChar(e, '`') }),
	Key{K: "`", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertChar(e, '~') }),
	Key{K: "<tab>"}:          makeCommand(func(e *TextBuffer) error { return e.Indent() }),
}

func insertChar(editor *TextBuffer, char byte) error {
	return editor.InsertCharAtCursor(char)
}

func getClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func writeToClipboard(bs []byte) {
	clipboard.Write(clipboard.FmtText, bytes.Clone(bs))
}

func insertCharAtSearchString(editor *TextBuffer, char byte) error {
	if editor.SearchString == nil {
		editor.SearchString = new(string)
	}

	*editor.SearchString += string(char)

	return nil
}

func (e *TextBuffer) DeleteCharBackwardFromActiveSearch() error {
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
	Key{K: "<space>"}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ' ') }),
	Key{K: "<backspace>"}: makeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackwardFromActiveSearch()
	}),
	Key{K: "<enter>"}: makeCommand(func(editor *TextBuffer) error {
		editor.CurrentMatch++
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),

	Key{K: "<enter>", Control: true}: makeCommand(func(editor *TextBuffer) error {
		editor.CurrentMatch--
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		if editor.CurrentMatch < 0 {
			editor.CurrentMatch = len(editor.SearchMatches) - 1
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<esc>"}: makeCommand(func(editor *TextBuffer) error {
		editor.keymaps = editor.keymaps[:len(editor.keymaps)-1]
		fmt.Println("exiting search mode")
		editor.IsSearching = false
		editor.LastSearchString = ""
		editor.SearchString = nil
		editor.SearchMatches = nil
		editor.CurrentMatch = 0
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<lmouse>-click"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),
	Key{K: "<mouse-wheel-up>"}: makeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true
		return e.ScrollUp(20)

	}),
	Key{K: "<mouse-wheel-down>"}: makeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true

		return e.ScrollDown(20)
	}),

	Key{K: "<rmouse>-click"}: makeCommand(func(editor *TextBuffer) error {
		editor.CurrentMatch++
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		if editor.CurrentMatch < 0 {
			editor.CurrentMatch = len(editor.SearchMatches) - 1
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<mmouse>-click"}: makeCommand(func(editor *TextBuffer) error {
		editor.CurrentMatch--
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		if editor.CurrentMatch < 0 {
			editor.CurrentMatch = len(editor.SearchMatches) - 1
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<pagedown>"}: makeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true
		return e.ScrollDown(1)
	}),
	Key{K: "<pageup>"}: makeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true

		return e.ScrollUp(1)
	}),

	Key{K: "a"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'a') }),
	Key{K: "b"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'b') }),
	Key{K: "c"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'c') }),
	Key{K: "d"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'd') }),
	Key{K: "e"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'e') }),
	Key{K: "f"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'f') }),
	Key{K: "g"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'g') }),
	Key{K: "h"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'h') }),
	Key{K: "i"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'i') }),
	Key{K: "j"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'j') }),
	Key{K: "k"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'k') }),
	Key{K: "l"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'l') }),
	Key{K: "m"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'm') }),
	Key{K: "n"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'n') }),
	Key{K: "o"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'o') }),
	Key{K: "p"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'p') }),
	Key{K: "q"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'q') }),
	Key{K: "r"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'r') }),
	Key{K: "s"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 's') }),
	Key{K: "t"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 't') }),
	Key{K: "u"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'u') }),
	Key{K: "v"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'v') }),
	Key{K: "w"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'w') }),
	Key{K: "x"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'x') }),
	Key{K: "y"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'y') }),
	Key{K: "z"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'z') }),
	Key{K: "0"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '0') }),
	Key{K: "1"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '1') }),
	Key{K: "2"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '2') }),
	Key{K: "3"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '3') }),
	Key{K: "4"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '4') }),
	Key{K: "5"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '5') }),
	Key{K: "6"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '6') }),
	Key{K: "7"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '7') }),
	Key{K: "8"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '8') }),
	Key{K: "9"}:               makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '9') }),
	Key{K: "\\"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '\\') }),
	Key{K: "\\", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '|') }),

	Key{K: "0", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ')') }),
	Key{K: "1", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '!') }),
	Key{K: "2", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '@') }),
	Key{K: "3", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '#') }),
	Key{K: "4", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '$') }),
	Key{K: "5", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '%') }),
	Key{K: "6", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '^') }),
	Key{K: "7", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '&') }),
	Key{K: "8", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '*') }),
	Key{K: "9", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '(') }),
	Key{K: "a", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'A') }),
	Key{K: "b", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'B') }),
	Key{K: "c", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'C') }),
	Key{K: "d", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'D') }),
	Key{K: "e", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'E') }),
	Key{K: "f", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'F') }),
	Key{K: "g", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'G') }),
	Key{K: "h", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'H') }),
	Key{K: "i", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'I') }),
	Key{K: "j", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'J') }),
	Key{K: "k", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'K') }),
	Key{K: "l", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'L') }),
	Key{K: "m", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'M') }),
	Key{K: "n", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'N') }),
	Key{K: "o", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'O') }),
	Key{K: "p", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'P') }),
	Key{K: "q", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'Q') }),
	Key{K: "r", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'R') }),
	Key{K: "s", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'S') }),
	Key{K: "t", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'T') }),
	Key{K: "u", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'U') }),
	Key{K: "v", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'V') }),
	Key{K: "w", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'W') }),
	Key{K: "x", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'X') }),
	Key{K: "y", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'Y') }),
	Key{K: "z", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'Z') }),
	Key{K: "["}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '[') }),
	Key{K: "]"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ']') }),
	Key{K: "[", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '{') }),
	Key{K: "]", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '}') }),
	Key{K: ";"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ';') }),
	Key{K: ";", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ':') }),
	Key{K: "'"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '\'') }),
	Key{K: "'", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '"') }),
	Key{K: "\""}:             makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '"') }),
	Key{K: ","}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ',') }),
	Key{K: "."}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '.') }),
	Key{K: ",", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '<') }),
	Key{K: ".", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '>') }),
	Key{K: "/"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '/') }),
	Key{K: "/", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '?') }),
	Key{K: "-"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '-') }),
	Key{K: "="}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '=') }),
	Key{K: "-", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '_') }),
	Key{K: "=", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '+') }),
	Key{K: "`"}:              makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '`') }),
	Key{K: "`", Shift: true}: makeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '~') }),
}
