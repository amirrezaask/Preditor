package preditor

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
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

type Selection struct {
	Start int
	End   int
}

type TextBuffer struct {
	BaseBuffer
	cfg            *Config
	parent         *Preditor
	File           string
	Content        []byte
	State          int
	BeforeSaveHook []func(*TextBuffer) error
	Readonly       bool

	keymaps []Keymap

	HasSyntaxHighlights bool
	SyntaxHighlights    *SyntaxHighlights

	TabSize int

	VisibleStart int32
	VisibleEnd   int32
	visualLines  []visualLine

	maxLine   int32
	maxColumn int32

	// Cursor
	Selections []Selection

	// Searching
	IsSearching               bool
	LastSearchString          string
	SearchString              *string
	SearchMatches             [][]int
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool

	UndoStack Stack[EditorAction]

	//Gotoline
	isGotoLine        bool
	GotoLineUserInput []byte

	LastCursorBlink time.Time
}

const (
	EditorActionType_Insert = iota + 1
	EditorActionType_Delete
)

type EditorAction struct {
	Type       int
	Idx        int
	Selections []Selection
	Data       []byte
}

func (e *TextBuffer) String() string {
	return fmt.Sprintf("%s", e.File)
}

func (e *TextBuffer) Keymaps() []Keymap {
	return e.keymaps
}

func (e *TextBuffer) IsSpecial() bool {
	return e.File != "" && e.File[0] != '*'
}

func (e *TextBuffer) AddUndoAction(a EditorAction) {
	a.Selections = e.Selections
	a.Data = bytes.Clone(a.Data)
	e.UndoStack.Push(a)
}

func (e *TextBuffer) PopAndReverseLastAction() {
	last, err := e.UndoStack.Pop()
	if err != nil {
		if err == EmptyStack {
			e.SetStateClean()
		}
		return
	}

	switch last.Type {
	case EditorActionType_Insert:
		e.Content = append(e.Content[:last.Idx], e.Content[last.Idx+len(last.Data):]...)
	case EditorActionType_Delete:
		e.Content = append(e.Content[:last.Idx], append(last.Data, e.Content[last.Idx:]...)...)
	}
	//e.bufferIndex = last.BufferIndex

	e.SetStateDirty()
}

func (e *TextBuffer) SetStateDirty() {
	e.State = State_Dirty
}

func (e *TextBuffer) SetStateClean() {
	e.State = State_Clean
}

func (e *TextBuffer) replaceTabsWithSpaces() {
	e.Content = bytes.Replace(e.Content, []byte("\t"), []byte(strings.Repeat(" ", e.TabSize)), -1)
}

func (e *TextBuffer) updateMaxLineAndColumn(maxH int32, maxW int32) {
	oldMaxLine := e.maxLine
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	e.maxColumn = maxW / int32(charSize.X)
	e.maxLine = maxH / int32(charSize.Y)

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

func SwitchOrOpenFileInTextBuffer(parent *Preditor, cfg *Config, filename string, startingPos *Position) error {
	for _, buf := range parent.Buffers {
		switch t := buf.(type) {
		case *TextBuffer:
			if t.File == filename {
				parent.MarkBufferAsActive(t.ID)
				if startingPos != nil {
					t.Selections = []Selection{{Start: t.positionToBufferIndex(*startingPos), End: t.positionToBufferIndex(*startingPos)}}
					t.ScrollIfNeeded()
				}
				return nil
			}
		}
	}

	tb, err := NewTextBuffer(parent, cfg, filename)
	if err != nil {
		return nil
	}

	if startingPos != nil {
		tb.Selections = []Selection{{Start: tb.positionToBufferIndex(*startingPos), End: tb.positionToBufferIndex(*startingPos)}}
		tb.ScrollIfNeeded()

	}
	parent.AddBuffer(tb)
	parent.MarkBufferAsActive(tb.ID)
	return nil
}

func NewTextBuffer(parent *Preditor, cfg *Config, filename string) (*TextBuffer, error) {
	t := TextBuffer{cfg: cfg}
	t.parent = parent
	t.File = filename
	t.keymaps = append([]Keymap{}, EditorKeymap)
	t.UndoStack = NewStack[EditorAction](1000)
	t.TabSize = t.cfg.TabSize
	t.Selections = append(t.Selections, Selection{Start: 0, End: 0})
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
	return &t, nil

}
func (e *TextBuffer) getIndexVisualLine(i int) visualLine {
	for _, line := range e.visualLines {
		if line.startIndex <= i && line.endIndex >= i {
			return line
		}
	}

	if len(e.visualLines) > 0 {
		lastLine := e.visualLines[len(e.visualLines)-1]
		lastLine.endIndex++
		return lastLine
	}
	return visualLine{}
}

func (e *TextBuffer) getIndexPosition(i int) Position {
	if len(e.visualLines) == 0 {
		return Position{Line: 0, Column: i}
	}

	line := e.getIndexVisualLine(i)
	return Position{
		Line:   line.Index,
		Column: i - line.startIndex,
	}

}

func (e *TextBuffer) positionToBufferIndex(pos Position) int {
	if len(e.visualLines) <= pos.Line {
		return len(e.Content)
	}

	return e.visualLines[pos.Line].startIndex + pos.Column
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
	if start == end {
		return hs
	}
	if len(hs) == 0 {
		missing = append(missing, highlight{
			start: start,
			end:   end - 1,
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
					end:   end - 1,
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
				Length:     idx - start + 1,
				ActualLine: actualLineIndex,
			}
			e.visualLines = append(e.visualLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1
		}
		if idx == len(e.Content)-1 {
			// last index
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx + 1,
				Length:     idx - start + 1,
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
				Length:     idx - start + 1,
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

func (e *TextBuffer) renderSelections(zeroLocation rl.Vector2, maxH int32, maxW int32) {

	//TODO:
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)

	for _, sel := range e.Selections {
		if sel.Start == sel.End {
			cursor := e.getIndexPosition(sel.Start)
			cursorView := Position{
				Line:   cursor.Line - int(e.VisibleStart),
				Column: cursor.Column,
			}
			posX := int32(cursorView.Column)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				if len(e.visualLines) > cursor.Line {
					posX += int32((len(fmt.Sprint(e.visualLines[cursor.Line].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			switch e.cfg.CursorShape {
			case CURSOR_SHAPE_OUTLINE:
				rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), int32(charSize.X), int32(charSize.Y), e.cfg.Colors.Cursor)
			case CURSOR_SHAPE_BLOCK:
				rl.DrawRectangle(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.Colors.Cursor, 0.5))
			case CURSOR_SHAPE_LINE:
				rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), 2, int32(charSize.Y), e.cfg.Colors.Cursor)
			}

			rl.DrawRectangle(0, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), e.maxColumn*int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.Colors.CursorLineBackground, 0.2))

		} else {
			e.highlightBetweenTwoIndexes(zeroLocation, sel.Start, sel.End, e.cfg.Colors.Selection)
		}

	}

}

func (e *TextBuffer) renderStatusBar(zeroLocation rl.Vector2, maxH int32, maxW int32) {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	//render status bar
	rl.DrawRectangle(
		int32(zeroLocation.X),
		e.maxLine*int32(charSize.Y),
		maxW,
		int32(charSize.Y),
		e.cfg.Colors.StatusBarBackground,
	)
	var sections []string
	if len(e.Selections) > 1 {
		sections = append(sections, fmt.Sprintf("%d#Selections", len(e.Selections)))
	} else {
		if e.Selections[0].Start == e.Selections[0].End {
			selStart := e.getIndexPosition(e.Selections[0].Start)
			sections = append(sections, fmt.Sprintf("Line#%d Col#%d", selStart.Line, selStart.Column))
		} else {
			selEnd := e.getIndexPosition(e.Selections[0].End)
			sections = append(sections, fmt.Sprintf("Line#%d Col#%d (Selected %d)", selEnd.Line, selEnd.Column, int(math.Abs(float64(e.Selections[0].Start-e.Selections[0].End)))))
		}

	}

	file := e.File

	var state string
	if e.State == State_Dirty {
		state = "**"
	} else {
		state = "--"
	}
	sections = append(sections, fmt.Sprintf("%s %s", state, file))

	var searchString string
	if e.SearchString != nil {
		searchString = fmt.Sprintf("Searching: \"%s\" %d of %d matches", *e.SearchString, e.CurrentMatch, len(e.SearchMatches)-1)
		sections = append(sections, searchString)
	}

	var gotoLine string
	if e.isGotoLine {
		gotoLine = fmt.Sprintf("Goto Line: %s", e.GotoLineUserInput)
		sections = append(sections, gotoLine)
	}

	rl.DrawTextEx(e.parent.Font,
		strings.Join(sections, " | "),
		rl.Vector2{X: zeroLocation.X, Y: float32(e.maxLine) * charSize.Y},
		float32(e.parent.FontSize),
		0,
		e.cfg.Colors.StatusBarForeground)
}

func (e *TextBuffer) highlightBetweenTwoIndexes(zeroLocation rl.Vector2, idx1 int, idx2 int, color color.RGBA) {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	var start Position
	var end Position
	if idx1 > idx2 {
		start = e.getIndexPosition(idx2)
		end = e.getIndexPosition(idx1)
	} else if idx2 > idx1 {
		start = e.getIndexPosition(idx1)
		end = e.getIndexPosition(idx2)
	}
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
			posX := int32(j)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				if len(e.visualLines) > i {
					posX += int32((len(fmt.Sprint(e.visualLines[i].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			rl.DrawRectangle(posX, int32(i-int(e.VisibleStart))*int32(charSize.Y)+int32(zeroLocation.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(color, 0.5))
		}
	}

}

func (e *TextBuffer) renderText(zeroLocation rl.Vector2, maxH int32, maxW int32) {
	e.calculateVisualLines()
	var visibleLines []visualLine
	if e.VisibleEnd > int32(len(e.visualLines)) {
		visibleLines = e.visualLines[e.VisibleStart:]
	} else {
		visibleLines = e.visualLines[e.VisibleStart:e.VisibleEnd]
	}
	for idx, line := range visibleLines {
		if e.visualLineShouldBeRendered(line) {
			charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
			var lineNumberWidth int
			if e.cfg.LineNumbers {
				lineNumberWidth = (len(fmt.Sprint(line.ActualLine)) + 1) * int(charSize.X)
				rl.DrawTextEx(e.parent.Font,
					fmt.Sprintf("%d", line.ActualLine),
					rl.Vector2{X: zeroLocation.X, Y: float32(idx) * charSize.Y},
					float32(e.parent.FontSize),
					0,
					e.cfg.Colors.LineNumbersForeground)

			}

			if e.cfg.EnableSyntaxHighlighting && e.HasSyntaxHighlights {
				highlights := e.fillInTheBlanks(e.calculateHighlights(e.Content[line.startIndex:line.endIndex], line.startIndex), line.startIndex, line.endIndex)
				for _, h := range highlights {
					rl.DrawTextEx(e.parent.Font,
						string(e.Content[h.start:h.end+1]),
						rl.Vector2{X: zeroLocation.X + float32(lineNumberWidth) + float32(h.start-line.startIndex)*charSize.X, Y: float32(idx) * charSize.Y},
						float32(e.parent.FontSize),
						0,
						h.Color)

				}
			} else {

				rl.DrawTextEx(e.parent.Font,
					string(e.Content[line.startIndex:line.endIndex]),
					rl.Vector2{X: zeroLocation.X + float32(lineNumberWidth), Y: float32(idx) * charSize.Y},
					float32(e.parent.FontSize),
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
func (e *TextBuffer) findMatches(pattern string) error {
	e.SearchMatches = [][]int{}
	matchPatternAsync(&e.SearchMatches, e.Content, []byte(pattern))
	return nil
}

func (e *TextBuffer) findMatchesAndHighlight(pattern string, zeroLocation rl.Vector2) error {
	if pattern != e.LastSearchString {
		if err := e.findMatches(pattern); err != nil {
			return err
		}
	}
	for idx, match := range e.SearchMatches {
		c := e.cfg.Colors.Selection
		_ = c
		if idx == e.CurrentMatch {
			c = rl.Fade(rl.Red, 0.5)
			matchStartLine := e.getIndexPosition(match[0])
			matchEndLine := e.getIndexPosition(match[0])
			if !(e.VisibleStart < int32(matchStartLine.Line) && e.VisibleEnd > int32(matchEndLine.Line)) && !e.MovedAwayFromCurrentMatch {
				// current match is not in view
				// move the view
				oldStart := e.VisibleStart
				e.VisibleStart = int32(matchStartLine.Line) - e.maxLine/2
				if e.VisibleStart < 0 {
					e.VisibleStart = int32(matchStartLine.Line)
				}

				diff := e.VisibleStart - oldStart
				e.VisibleEnd += diff
			}
		}
		e.highlightBetweenTwoIndexes(zeroLocation, match[0], match[1], c)
	}
	e.LastSearchString = pattern

	return nil
}
func (e *TextBuffer) renderSearchResults(zeroLocation rl.Vector2) {
	if e.SearchString == nil || len(*e.SearchString) < 1 {
		return
	}
	e.findMatchesAndHighlight(*e.SearchString, zeroLocation)
	if len(e.SearchMatches) > 0 {
		e.Selections = e.Selections[:1]
		e.Selections[0].Start = e.SearchMatches[e.CurrentMatch][0]
		e.Selections[0].End = e.SearchMatches[e.CurrentMatch][0]
	}
}

func (e *TextBuffer) Render(zeroLocation rl.Vector2, maxH int32, maxW int32) {
	e.updateMaxLineAndColumn(maxH, maxW)
	e.renderText(zeroLocation, maxH, maxW)
	e.renderSearchResults(zeroLocation)
	e.renderSelections(zeroLocation, maxH, maxW)
	e.renderStatusBar(zeroLocation, maxH, maxW)
}

func (e *TextBuffer) visualLineShouldBeRendered(line visualLine) bool {
	if e.VisibleStart <= int32(line.Index) && line.Index <= int(e.VisibleEnd) {
		return true
	}

	return false
}

func (e *TextBuffer) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if len(e.visualLines) == 0 && newPosition.Line == 0 && newPosition.Column >= 0 && int32(newPosition.Column) < e.maxColumn {
		return true
	}
	if newPosition.Line >= len(e.visualLines) && (len(e.visualLines) != 0) {
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

func (e *TextBuffer) deleteSelectionIfSelection() {
	if e.Readonly {
		return
	}
	for i, sel := range e.Selections {
		if sel.Start == sel.End {
			continue
		}

		if sel.Start > sel.End {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  sel.End,
				Data: e.Content[sel.End:sel.Start],
			})
			e.Content = append(e.Content[:sel.End], e.Content[sel.Start+1:]...)
		} else {
			e.Content = append(e.Content[:sel.Start], e.Content[sel.End+1:]...)
		}

		e.Selections[i].Start = e.Selections[i].End
	}

}

func (e *TextBuffer) sortSelections() {
	sortme(e.Selections, func(t1 Selection, t2 Selection) bool {
		return t1.Start < t2.Start
	})
}

func (e *TextBuffer) removeDuplicateSelectionsAndSort() {
	selections := map[string]struct{}{}
	for i := range e.Selections {
		selections[fmt.Sprintf("%d:%d", e.Selections[i].Start, e.Selections[i].End)] = struct{}{}
	}

	e.Selections = nil
	for k := range selections {
		start, _ := strconv.Atoi(strings.Split(k, ":")[0])
		end, _ := strconv.Atoi(strings.Split(k, ":")[1])
		e.Selections = append(e.Selections, Selection{Start: start, End: end})
	}

	e.sortSelections()
}

func (e *TextBuffer) InsertCharAtCursor(char byte) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionIfSelection()
	for i := range e.Selections {
		e.MoveRight(&e.Selections[i], i*1)

		if e.Selections[i].Start >= len(e.Content) { // end of file, appending
			e.Content = append(e.Content, char)

		} else {
			e.Content = append(e.Content[:e.Selections[i].Start+1], e.Content[e.Selections[i].Start:]...)
			e.Content[e.Selections[i].Start] = char
		}
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  e.Selections[i].Start,
			Data: []byte{char},
		})
		e.MoveRight(&e.Selections[i], 1)
	}
	e.SetStateDirty()
	e.ScrollIfNeeded()
	return nil
}

func (e *TextBuffer) DeleteCharBackward() error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()

	e.deleteSelectionIfSelection()
	for i := range e.Selections {
		e.MoveLeft(&e.Selections[i], i*1)

		switch {
		case e.Selections[i].Start == 0:
			continue
		case e.Selections[i].Start < len(e.Content):
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Selections[i].Start - 1,
				Data: []byte{e.Content[e.Selections[i].Start-1]},
			})
			e.Content = append(e.Content[:e.Selections[i].Start-1], e.Content[e.Selections[i].Start:]...)
		case e.Selections[i].Start == len(e.Content):
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Selections[i].Start - 1,
				Data: []byte{e.Content[e.Selections[i].Start-1]},
			})
			e.Content = e.Content[:e.Selections[i].Start-1]
		}

		e.MoveLeft(&e.Selections[i], 1)

	}

	e.SetStateDirty()
	return nil
}

func (e *TextBuffer) DeleteCharForward() error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()

	for i := range e.Selections {
		if len(e.Content) > e.Selections[i].Start {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Selections[i].Start,
				Data: []byte{e.Content[e.Selections[i].Start]},
			})
			e.Content = append(e.Content[:e.Selections[i].Start], e.Content[e.Selections[i].Start+1:]...)
			e.SetStateDirty()
			e.MoveLeft(&e.Selections[i], i)
		}
	}

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

func (e *TextBuffer) MoveLeft(s *Selection, n int) {
	s.Start -= n
	if s.Start < 0 {
		s.Start = 0
	}
	s.End = s.Start

}

func (e *TextBuffer) MoveAllLeft(n int) error {
	for i := range e.Selections {
		e.MoveLeft(&e.Selections[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}
func (e *TextBuffer) MoveRight(s *Selection, n int) {
	s.Start += n
	if s.Start > len(e.Content) {
		s.Start = len(e.Content)
	}
	line := e.getIndexVisualLine(s.Start)
	if s.Start-line.startIndex > int(e.maxColumn) {
		s.Start = int(e.maxColumn)
	}
	s.End = s.Start

}

func (e *TextBuffer) MoveAllRight(n int) error {
	for i := range e.Selections {
		e.MoveRight(&e.Selections[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func (e *TextBuffer) MoveUp() error {
	for i := range e.Selections {
		currentLine := e.getIndexVisualLine(e.Selections[i].Start)
		prevLineIndex := currentLine.Index - 1
		if prevLineIndex < 0 {
			return nil
		}

		prevLine := e.visualLines[prevLineIndex]
		col := e.Selections[i].Start - currentLine.startIndex
		newcol := prevLine.startIndex + col
		if newcol > prevLine.endIndex {
			newcol = prevLine.endIndex
		}
		e.Selections[i].Start = newcol
		e.Selections[i].End = newcol

		e.ScrollIfNeeded()
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func (e *TextBuffer) MoveDown() error {
	for i := range e.Selections {
		currentLine := e.getIndexVisualLine(e.Selections[i].Start)
		nextLineIndex := currentLine.Index + 1
		if nextLineIndex >= len(e.visualLines) {
			return nil
		}

		nextLine := e.visualLines[nextLineIndex]
		col := e.Selections[i].Start - currentLine.startIndex
		newcol := nextLine.startIndex + col
		if newcol > nextLine.endIndex {
			newcol = nextLine.endIndex
		}
		e.Selections[i].Start = newcol
		e.Selections[i].End = newcol
		e.ScrollIfNeeded()

	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherSelectionHere(e *TextBuffer, pos rl.Vector2) error {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64(pos.Y / charSize.Y))
	apprColumn := math.Floor(float64(pos.X / charSize.X))

	if e.cfg.LineNumbers {
		apprColumn -= float64(len(fmt.Sprint(apprLine)) + 1)
	}

	if len(e.visualLines) < 1 {
		return nil
	}

	line := int(apprLine) + int(e.VisibleStart)
	col := int(apprColumn)

	if line >= len(e.visualLines) {
		line = len(e.visualLines) - 1
	}

	if line < 0 {
		line = 0
	}

	// check if cursor should be moved back
	if col > e.visualLines[line].Length {
		col = e.visualLines[line].Length
	}
	if col < 0 {
		col = 0
	}
	idx := e.positionToBufferIndex(Position{Line: line, Column: col})
	e.Selections = append(e.Selections, Selection{Start: idx, End: idx})

	e.removeDuplicateSelectionsAndSort()
	return nil
}

func (e *TextBuffer) MoveToBeginningOfTheLine() error {
	for i := range e.Selections {
		line := e.getIndexVisualLine(e.Selections[i].Start)
		e.Selections[i].Start = line.startIndex
		e.Selections[i].End = line.startIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func (e *TextBuffer) MoveToEndOfTheLine() error {
	for i := range e.Selections {
		line := e.getIndexVisualLine(e.Selections[i].Start)
		e.Selections[i].Start = line.endIndex
		e.Selections[i].End = line.endIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherCursorNextLine(e *TextBuffer) error {
	pos := e.getIndexPosition(e.Selections[len(e.Selections)-1].Start)
	pos.Line++
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Selections = append(e.Selections, Selection{Start: newidx, End: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherCursorPreviousLine(e *TextBuffer) error {
	pos := e.getIndexPosition(e.Selections[len(e.Selections)-1].Start)
	pos.Line--
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Selections = append(e.Selections, Selection{Start: newidx, End: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}
func (e *TextBuffer) PreviousLine() error {
	return e.MoveUp()
}

func (e *TextBuffer) NextLine() error {
	return e.MoveDown()
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
	for i := range e.Selections {
		newidx := nextWordInBuffer(e.Content, e.Selections[i].Start)
		if newidx == -1 {
			return nil
		}
		if newidx > len(e.Content) {
			newidx = len(e.Content)
		}
		e.Selections[i].Start = newidx
		e.Selections[i].End = newidx

	}
	return nil
}

func (e *TextBuffer) PreviousWord() error {
	for i := range e.Selections {
		newidx := previousWordInBuffer(e.Content, e.Selections[i].Start)
		if newidx == -1 {
			return nil
		}
		if newidx < 0 {
			newidx = 0
		}
		e.Selections[i].Start = newidx
		e.Selections[i].End = newidx

	}
	return nil
}

func (e *TextBuffer) MoveCursorTo(pos rl.Vector2) error {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64(pos.Y / charSize.Y))
	apprColumn := math.Floor(float64(pos.X / charSize.X))

	if e.cfg.LineNumbers {
		apprColumn -= float64((len(fmt.Sprint(apprLine)) + 1))
	}

	if len(e.visualLines) < 1 {
		return nil
	}

	line := int(apprLine) + int(e.VisibleStart)
	col := int(apprColumn)

	if line >= len(e.visualLines) {
		line = len(e.visualLines) - 1
	}

	if line < 0 {
		line = 0
	}

	// check if cursor should be moved back
	if col > e.visualLines[line].Length {
		col = e.visualLines[line].Length
	}
	if col < 0 {
		col = 0
	}

	e.Selections[0].Start = e.positionToBufferIndex(Position{Line: line, Column: col})
	e.Selections[0].End = e.positionToBufferIndex(Position{Line: line, Column: col})

	return nil
}

func (e *TextBuffer) ScrollIfNeeded() error {
	pos := e.getIndexPosition(e.Selections[0].Start)
	if int32(pos.Line) <= e.VisibleStart {
		e.VisibleStart = int32(pos.Line) - e.maxLine/3
		e.VisibleEnd = e.VisibleStart + e.maxLine

	} else if int32(pos.Line) >= e.VisibleEnd {
		e.VisibleEnd = int32(pos.Line) + e.maxLine/3
		e.VisibleStart = e.VisibleEnd - e.maxLine
	}

	if int(e.VisibleEnd) >= len(e.visualLines) {
		e.VisibleEnd = int32(len(e.visualLines) - 1)
		e.VisibleStart = e.VisibleEnd - e.maxLine
	}

	if e.VisibleStart < 0 {
		e.VisibleStart = 0
		e.VisibleEnd = e.maxLine
	}
	if e.VisibleEnd < 0 {
		e.VisibleStart = 0
		e.VisibleEnd = e.maxLine
	}

	return nil
}

func (e *TextBuffer) Write() error {
	if e.Readonly && e.IsSpecial() {
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
	return nil
}

func (e *TextBuffer) Indent() error {
	e.removeDuplicateSelectionsAndSort()

	for i := range e.Selections {
		e.MoveRight(&e.Selections[i], i*e.TabSize)
		if e.Selections[i].Start >= len(e.Content) { // end of file, appending
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Selections[i].Start,
				Data: []byte(strings.Repeat(" ", e.TabSize)),
			})
			e.Content = append(e.Content, []byte(strings.Repeat(" ", e.TabSize))...)
		} else {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Selections[i].Start,
				Data: []byte(strings.Repeat(" ", e.TabSize)),
			})
			e.Content = append(e.Content[:e.Selections[i].Start], append([]byte(strings.Repeat(" ", e.TabSize)), e.Content[e.Selections[i].Start:]...)...)
		}
		e.MoveRight(&e.Selections[i], e.TabSize)

	}
	e.SetStateDirty()

	return nil
}

func (e *TextBuffer) KillLine() error {
	if e.Readonly {
		return nil
	}
	var lastChange int
	for i := range e.Selections {
		old := len(e.Content)
		cur := &e.Selections[i]
		e.MoveLeft(cur, -1*lastChange)
		line := e.getIndexVisualLine(cur.Start)
		writeToClipboard(e.Content[cur.Start:line.endIndex])
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  cur.Start,
			Data: e.Content[cur.Start:line.endIndex],
		})
		e.Content = append(e.Content[:cur.Start], e.Content[line.endIndex:]...)
		lastChange += len(e.Content) - old
	}
	e.SetStateDirty()

	return nil
}

func (e *TextBuffer) DeleteWordBackward() {
	if e.Readonly || len(e.Selections) > 1 {
		return
	}
	cur := e.Selections[0]
	previousWordEndIdx := previousWordInBuffer(e.Content, cur.Start)
	oldLen := len(e.Content)
	if len(e.Content) > cur.Start+1 {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  previousWordEndIdx + 1,
			Data: e.Content[previousWordEndIdx+1 : cur.Start],
		})
		e.Content = append(e.Content[:previousWordEndIdx+1], e.Content[cur.Start+1:]...)
	} else {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  previousWordEndIdx + 1,
			Data: e.Content[previousWordEndIdx+1:],
		})
		e.Content = e.Content[:previousWordEndIdx+1]
	}
	cur.Start += len(e.Content) - oldLen
	e.SetStateDirty()
}

func (e *TextBuffer) Copy() error {
	if len(e.Selections) > 1 {
		return nil
	}
	cur := e.Selections[0]
	if cur.Start != cur.End {
		// Copy selection
		switch {
		case cur.Start < cur.End:
			writeToClipboard(e.Content[cur.Start:cur.End])
		case cur.Start > cur.End:
			writeToClipboard(e.Content[cur.End:cur.Start])
		}
	} else {
		line := e.getIndexVisualLine(cur.Start)
		writeToClipboard(e.Content[line.startIndex : line.endIndex+1])
	}

	return nil
}

func (e *TextBuffer) Cut() error {
	if e.Readonly || len(e.Selections) > 1 {
		return nil
	}
	cur := &e.Selections[0]
	if cur.Start != cur.End {
		// Copy selection
		switch {
		case cur.Start < cur.End:
			writeToClipboard(e.Content[cur.Start:cur.End])
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  cur.Start,
				Data: e.Content[cur.Start:cur.End],
			})
			e.Content = append(e.Content[:cur.Start], e.Content[cur.End+1:]...)

		case cur.Start > cur.End:
			writeToClipboard(e.Content[cur.End:cur.Start])
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  cur.End,
				Data: e.Content[cur.End:cur.Start],
			})
			e.Content = append(e.Content[:cur.End], e.Content[cur.Start+1:]...)
		}
	} else {
		line := e.getIndexVisualLine(cur.Start)
		writeToClipboard(e.Content[line.startIndex : line.endIndex+1])
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  line.startIndex,
			Data: e.Content[line.startIndex:line.endIndex],
		})
		e.Content = append(e.Content[:line.startIndex], e.Content[line.endIndex+1:]...)
	}
	e.SetStateDirty()

	return nil
}
func (e *TextBuffer) Paste() error {
	if e.Readonly || len(e.Selections) > 1 {
		return nil
	}
	e.deleteSelectionIfSelection()
	contentToPaste := getClipboardContent()
	cur := e.Selections[0]
	if cur.Start >= len(e.Content) { // end of file, appending
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  cur.Start,
			Data: contentToPaste,
		})
		e.Content = append(e.Content, contentToPaste...)
	} else {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  cur.Start,
			Data: contentToPaste,
		})
		e.Content = append(e.Content[:cur.Start], append(contentToPaste, e.Content[cur.Start:]...)...)
	}

	e.SetStateDirty()

	e.MoveAllRight(len(contentToPaste))
	return nil
}

func makeCommand(f func(e *TextBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*TextBuffer))
	}
}

var EditorKeymap = Keymap{
	Key{K: "=", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.parent.IncreaseFontSize(10)

		return nil
	}),
	Key{K: "<lmouse>-click", Control: true}: makeCommand(func(e *TextBuffer) error {
		return PlaceAnotherSelectionHere(e, rl.GetMousePosition())
	}),
	Key{K: "<lmouse>-hold", Control: true}: makeCommand(func(e *TextBuffer) error {
		return PlaceAnotherSelectionHere(e, rl.GetMousePosition())
	}),
	Key{K: "<up>", Control: true}: makeCommand(func(e *TextBuffer) error {
		return PlaceAnotherCursorPreviousLine(e)
	}),

	Key{K: "<down>", Control: true}: makeCommand(func(e *TextBuffer) error {
		return PlaceAnotherCursorNextLine(e)
	}),
	Key{K: "-", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.parent.DecreaseFontSize(10)

		return nil
	}),
	Key{K: "r", Alt: true}: makeCommand(func(e *TextBuffer) error {
		return e.readFileFromDisk()
	}),
	Key{K: "/", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.PopAndReverseLastAction()
		return nil
	}),
	Key{K: "z", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.PopAndReverseLastAction()
		return nil
	}),
	Key{K: "f", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.MoveAllRight(1)
	}),
	Key{K: "x", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.Cut()
	}),
	Key{K: "v", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.Paste()
	}),
	Key{K: "k", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.KillLine()
	}),
	Key{K: "g", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.keymaps = append(e.keymaps, GotoLineKeymap)
		e.isGotoLine = true

		return nil
	}),
	Key{K: "c", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.Copy()
	}),

	Key{K: "s", Control: true}: makeCommand(func(a *TextBuffer) error {
		a.keymaps = append(a.keymaps, SearchTextBufferKeymap)
		return nil
	}),
	Key{K: "w", Control: true}: makeCommand(func(a *TextBuffer) error {
		return a.Write()
	}),

	Key{K: "<esc>"}: makeCommand(func(p *TextBuffer) error {
		p.Selections = p.Selections[:1]

		return nil
	}),

	// navigation
	Key{K: "<lmouse>-click"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),
	Key{K: "<lmouse>-hold"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),

	Key{K: "a", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.MoveToBeginningOfTheLine()
	}),
	Key{K: "e", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.MoveToEndOfTheLine()
	}),

	Key{K: "p", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.PreviousLine()
	}),

	Key{K: "n", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.NextLine()
	}),

	Key{K: "<up>"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveUp()
	}),
	Key{K: "<down>"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveDown()
	}),
	Key{K: "<right>"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveAllRight(1)
	}),
	Key{K: "<right>", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.NextWord()
	}),
	Key{K: "<left>"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveAllLeft(1)
	}),
	Key{K: "<left>", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.PreviousWord()
	}),

	Key{K: "b", Control: true}: makeCommand(func(e *TextBuffer) error {
		return e.MoveAllLeft(1)
	}),
	Key{K: "<home>"}: makeCommand(func(e *TextBuffer) error {
		return e.MoveToBeginningOfTheLine()
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
	Key{K: "<backspace>", Control: true}: makeCommand(func(e *TextBuffer) error {
		e.DeleteWordBackward()
		return nil
	}),
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

func insertCharAtGotoLineBuffer(editor *TextBuffer, char byte) error {
	editor.GotoLineUserInput = append(editor.GotoLineUserInput, char)

	return nil
}

func (e *TextBuffer) readFileFromDisk() error {
	bs, err := os.ReadFile(e.File)
	if err != nil {
		return nil
	}

	e.Content = bs
	e.replaceTabsWithSpaces()
	e.SetStateClean()
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

func (e *TextBuffer) DeleteCharBackwardFromGotoLine() error {
	if len(e.GotoLineUserInput) < 1 {
		return nil
	}
	e.GotoLineUserInput = e.GotoLineUserInput[:len(e.GotoLineUserInput)-1]

	return nil
}

var SearchTextBufferKeymap = Keymap{
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

var GotoLineKeymap = Keymap{
	Key{K: "<backspace>"}: makeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackwardFromActiveSearch()
	}),
	Key{K: "<enter>"}: makeCommand(func(editor *TextBuffer) error {
		number, err := strconv.Atoi(string(editor.GotoLineUserInput))
		if err != nil {
			return nil
		}

		for _, line := range editor.visualLines {
			if line.Index == number {
				if !(editor.VisibleStart < int32(line.Index)) || !(editor.VisibleEnd > int32(line.Index)) {
					editor.VisibleStart = int32(int32(line.Index) - editor.maxLine/2)
					if editor.VisibleStart < 0 {
						editor.VisibleStart = 0
					}

					editor.VisibleEnd = int32(int32(line.Index) + editor.maxLine/2)
					if editor.VisibleEnd > int32(len(editor.visualLines)) {
						editor.VisibleEnd = int32(len(editor.visualLines))
					}

					editor.isGotoLine = false
					editor.GotoLineUserInput = nil
					editor.Selections[0].Start = line.startIndex
				}
			}
		}

		return nil
	}),

	Key{K: "<esc>"}: makeCommand(func(editor *TextBuffer) error {
		editor.keymaps = editor.keymaps[:len(editor.keymaps)-1]
		editor.isGotoLine = false
		return nil
	}),

	Key{K: "0"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '0') }),
	Key{K: "1"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '1') }),
	Key{K: "2"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '2') }),
	Key{K: "3"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '3') }),
	Key{K: "4"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '4') }),
	Key{K: "5"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '5') }),
	Key{K: "6"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '6') }),
	Key{K: "7"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '7') }),
	Key{K: "8"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '8') }),
	Key{K: "9"}: makeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '9') }),
}
