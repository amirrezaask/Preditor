package preditor

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/amirrezaask/preditor/byteutils"
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

type Range struct {
	Static int
	Moving int
}

func (r *Range) SetBoth(n int) {
	r.Static = n
	r.Moving = n
}
func (r *Range) AddToBoth(n int) {
	r.Static += n
	r.Moving += n
}
func (r *Range) AddToStart(n int) {
	if r.Static > r.Moving {
		r.Moving += n
	} else {
		r.Static += n
	}
}
func (r *Range) AddToEnd(n int) {
	if r.Static > r.Moving {
		r.Static += n
	} else {
		r.Moving += n
	}
}
func (r *Range) Start() int {
	if r.Static > r.Moving {
		return r.Moving
	} else {
		return r.Static
	}
}
func (r *Range) End() int {
	if r.Static > r.Moving {
		return r.Static
	} else {
		return r.Moving
	}
}

type TextBuffer struct {
	BaseBuffer
	cfg            *Config
	parent         *Context
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
	Ranges []Range

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
	Selections []Range
	Data       []byte
}

func (e *TextBuffer) String() string {
	return fmt.Sprintf("%s", e.File)
}

func (e *TextBuffer) Keymaps() []Keymap {
	return e.keymaps
}

func (e *TextBuffer) IsSpecial() bool {
	return e.File == "" || e.File[0] == '*'
}

func (e *TextBuffer) AddUndoAction(a EditorAction) {
	a.Selections = e.Ranges
	a.Data = bytes.Clone(a.Data)
	e.UndoStack.Push(a)
}

func (e *TextBuffer) PopAndReverseLastAction() {
	last, err := e.UndoStack.Pop()
	if err != nil {
		if errors.Is(err, EmptyStack) {
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

func SwitchOrOpenFileInTextBuffer(parent *Context, cfg *Config, filename string, startingPos *Position) error {
	for _, buf := range parent.Buffers {
		switch t := buf.(type) {
		case *TextBuffer:
			if t.File == filename {
				parent.MarkBufferAsActive(t.ID)
				if startingPos != nil {
					t.Ranges = []Range{{Static: t.positionToBufferIndex(*startingPos), Moving: t.positionToBufferIndex(*startingPos)}}
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
		tb.Ranges = []Range{{Static: tb.positionToBufferIndex(*startingPos), Moving: tb.positionToBufferIndex(*startingPos)}}
		tb.ScrollIfNeeded()

	}
	parent.AddBuffer(tb)
	parent.MarkBufferAsActive(tb.ID)
	return nil
}

func NewTextBuffer(parent *Context, cfg *Config, filename string) (*TextBuffer, error) {
	t := TextBuffer{cfg: cfg}
	t.parent = parent
	t.File = filename
	t.keymaps = append([]Keymap{}, EditorKeymap)
	t.UndoStack = NewStack[EditorAction](1000)
	t.TabSize = t.cfg.TabSize
	t.Ranges = append(t.Ranges, Range{Static: 0, Moving: 0})
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

	for _, sel := range e.Ranges {
		if sel.Start() == sel.End() {
			cursor := e.getIndexPosition(sel.Start())
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
			e.highlightBetweenTwoIndexes(zeroLocation, sel.Start(), sel.End(), e.cfg.Colors.Selection)
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
	if len(e.Ranges) > 1 {
		sections = append(sections, fmt.Sprintf("%d#Ranges", len(e.Ranges)))
	} else {
		if e.Ranges[0].Start() == e.Ranges[0].End() {
			selStart := e.getIndexPosition(e.Ranges[0].Start())
			sections = append(sections, fmt.Sprintf("Line#%d Col#%d", selStart.Line, selStart.Column))
		} else {
			selEnd := e.getIndexPosition(e.Ranges[0].End())
			sections = append(sections, fmt.Sprintf("Line#%d Col#%d (Selected %d)", selEnd.Line, selEnd.Column, int(math.Abs(float64(e.Ranges[0].Start()-e.Ranges[0].End())))))
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
		e.Ranges = e.Ranges[:1]
		e.Ranges[0].Static = e.SearchMatches[e.CurrentMatch][0]
		e.Ranges[0].Moving = e.SearchMatches[e.CurrentMatch][0]
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

func (e *TextBuffer) deleteSelectionsIfAnySelection() {
	if e.Readonly {
		return
	}
	old := len(e.Content)
	for i := range e.Ranges {
		sel := &e.Ranges[i]
		if sel.Start() == sel.End() {
			continue
		}
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  sel.Start(),
			Data: e.Content[sel.Start() : sel.End()+1],
		})
		e.Content = append(e.Content[:sel.Start()], e.Content[sel.End()+1:]...)
		sel.Static = sel.Moving
		e.MoveLeft(sel, old-1-len(e.Content))
	}

}

func (e *TextBuffer) sortSelections() {
	sortme(e.Ranges, func(t1 Range, t2 Range) bool {
		return t1.Start() < t2.Start()
	})
}

func (e *TextBuffer) removeDuplicateSelectionsAndSort() {
	selections := map[string]struct{}{}
	for i := range e.Ranges {
		selections[fmt.Sprintf("%d:%d", e.Ranges[i].Start(), e.Ranges[i].End())] = struct{}{}
	}

	e.Ranges = nil
	for k := range selections {
		start, _ := strconv.Atoi(strings.Split(k, ":")[0])
		end, _ := strconv.Atoi(strings.Split(k, ":")[1])
		e.Ranges = append(e.Ranges, Range{Static: start, Moving: end})
	}

	e.sortSelections()
}

func (e *TextBuffer) InsertCharAtCursor(char byte) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Ranges {
		e.MoveRight(&e.Ranges[i], i*1)

		if e.Ranges[i].Start() >= len(e.Content) { // end of file, appending
			e.Content = append(e.Content, char)

		} else {
			e.Content = append(e.Content[:e.Ranges[i].Start()+1], e.Content[e.Ranges[i].End():]...)
			e.Content[e.Ranges[i].Start()] = char
		}
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  e.Ranges[i].Start(),
			Data: []byte{char},
		})
		e.MoveRight(&e.Ranges[i], 1)
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

	e.deleteSelectionsIfAnySelection()
	for i := range e.Ranges {
		e.MoveLeft(&e.Ranges[i], i*1)

		switch {
		case e.Ranges[i].Start() == 0:
			continue
		case e.Ranges[i].Start() < len(e.Content):
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Ranges[i].Start() - 1,
				Data: []byte{e.Content[e.Ranges[i].Start()-1]},
			})
			e.Content = append(e.Content[:e.Ranges[i].Start()-1], e.Content[e.Ranges[i].Start():]...)
		case e.Ranges[i].Start() == len(e.Content):
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Ranges[i].Start() - 1,
				Data: []byte{e.Content[e.Ranges[i].Start()-1]},
			})
			e.Content = e.Content[:e.Ranges[i].Start()-1]
		}

		e.MoveLeft(&e.Ranges[i], 1)

	}

	e.SetStateDirty()
	return nil
}

func (e *TextBuffer) DeleteCharForward() error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Ranges {
		if len(e.Content) > e.Ranges[i].Start() {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Ranges[i].Start(),
				Data: []byte{e.Content[e.Ranges[i].Start()]},
			})
			e.Content = append(e.Content[:e.Ranges[i].Start()], e.Content[e.Ranges[i].Start()+1:]...)
			e.SetStateDirty()
			e.MoveLeft(&e.Ranges[i], i)
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

// Move* functions change Static part of cursor

func (e *TextBuffer) MoveLeft(s *Range, n int) {
	s.AddToBoth(-n)
	if s.Start() < 0 {
		s.SetBoth(0)
	}
}

func (e *TextBuffer) MoveAllLeft(n int) error {
	for i := range e.Ranges {
		e.MoveLeft(&e.Ranges[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}
func (e *TextBuffer) MoveRight(s *Range, n int) {
	e.Ranges[0].Static = e.Ranges[0].Moving
	s.AddToBoth(n)
	if s.Start() > len(e.Content) {
		s.SetBoth(len(e.Content))
	}
	line := e.getIndexVisualLine(s.Start())
	if s.Start()-line.startIndex > int(e.maxColumn) {
		s.SetBoth(int(e.maxColumn))
	}
}

func (e *TextBuffer) MoveAllRight(n int) error {
	for i := range e.Ranges {
		e.MoveRight(&e.Ranges[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func (e *TextBuffer) MoveUp() error {
	for i := range e.Ranges {
		currentLine := e.getIndexVisualLine(e.Ranges[i].Static)
		prevLineIndex := currentLine.Index - 1
		if prevLineIndex < 0 {
			return nil
		}

		prevLine := e.visualLines[prevLineIndex]
		col := e.Ranges[i].Static - currentLine.startIndex
		newcol := prevLine.startIndex + col
		if newcol > prevLine.endIndex {
			newcol = prevLine.endIndex
		}
		e.Ranges[i].SetBoth(newcol)

		e.ScrollIfNeeded()
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func (e *TextBuffer) MoveDown() error {
	for i := range e.Ranges {
		currentLine := e.getIndexVisualLine(e.Ranges[i].Static)
		nextLineIndex := currentLine.Index + 1
		if nextLineIndex >= len(e.visualLines) {
			return nil
		}

		nextLine := e.visualLines[nextLineIndex]
		col := e.Ranges[i].Static - currentLine.startIndex
		newcol := nextLine.startIndex + col
		if newcol > nextLine.endIndex {
			newcol = nextLine.endIndex
		}
		e.Ranges[i].SetBoth(newcol)
		e.ScrollIfNeeded()

	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func SelectionsToRight(e *TextBuffer, n int) error {
	for i := range e.Ranges {
		sel := &e.Ranges[i]
		sel.Moving += n
		if sel.Moving >= len(e.Content) {
			sel.Moving = len(e.Content)
		}
	}

	return nil
}
func SelectionsToLeft(e *TextBuffer, n int) error {
	for i := range e.Ranges {
		sel := &e.Ranges[i]
		sel.Moving -= n
		if sel.Moving < 0 {
			sel.Moving = 0
		}
	}

	return nil
}
func SelectionsUp(e *TextBuffer, n int) error {
	for i := range e.Ranges {
		sel := &e.Ranges[i]
		pos := e.getIndexPosition(sel.Moving)
		pos.Line -= n
		if !e.isValidCursorPosition(pos) {
			continue
		}
		newidx := e.positionToBufferIndex(pos)
		sel.Moving += newidx - sel.Moving
		if sel.Moving < 0 {
			sel.Moving = 0
		}
	}

	return nil
}
func SelectionsDown(e *TextBuffer, n int) error {
	for i := range e.Ranges {
		sel := &e.Ranges[i]
		pos := e.getIndexPosition(sel.Moving)
		pos.Line += n
		if !e.isValidCursorPosition(pos) {
			continue
		}
		newidx := e.positionToBufferIndex(pos)
		sel.Moving += newidx - sel.Moving
		if sel.Moving > len(e.Content) {
			sel.Moving = len(e.Content)
		}
	}

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
	e.Ranges = append(e.Ranges, Range{Static: idx, Moving: idx})

	e.removeDuplicateSelectionsAndSort()
	return nil
}

func (e *TextBuffer) PlaceSelectionOnNextMatch() error {
	lastSel := e.Ranges[len(e.Ranges)-1]
	next := findNextMatch(e.Content[lastSel.End():], e.Content[lastSel.Start():lastSel.End()+1])

	if len(next) == 0 {
		return nil
	}

	e.Ranges = append(e.Ranges, Range{
		Static: next[0],
		Moving: next[1],
	})

	return nil
}

func (e *TextBuffer) MoveToBeginningOfTheLine() error {
	for i := range e.Ranges {
		line := e.getIndexVisualLine(e.Ranges[i].Start())
		e.Ranges[i].SetBoth(line.startIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func (e *TextBuffer) MoveToEndOfTheLine() error {
	for i := range e.Ranges {
		line := e.getIndexVisualLine(e.Ranges[i].Start())
		e.Ranges[i].SetBoth(line.endIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherCursorNextLine(e *TextBuffer) error {
	pos := e.getIndexPosition(e.Ranges[len(e.Ranges)-1].Start())
	pos.Line++
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Ranges = append(e.Ranges, Range{Static: newidx, Moving: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherCursorPreviousLine(e *TextBuffer) error {
	pos := e.getIndexPosition(e.Ranges[len(e.Ranges)-1].Start())
	pos.Line--
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Ranges = append(e.Ranges, Range{Static: newidx, Moving: newidx})
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
	for i := range e.Ranges {
		newidx := byteutils.NextWordInBuffer(e.Content, e.Ranges[i].Start())
		if newidx == -1 {
			return nil
		}
		if newidx > len(e.Content) {
			newidx = len(e.Content)
		}
		e.Ranges[i].SetBoth(newidx)

	}
	return nil
}

func (e *TextBuffer) PreviousWord() error {
	for i := range e.Ranges {
		newidx := byteutils.PreviousWordInBuffer(e.Content, e.Ranges[i].Start())
		if newidx == -1 {
			return nil
		}
		if newidx < 0 {
			newidx = 0
		}
		e.Ranges[i].SetBoth(newidx)

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

	e.Ranges[0].SetBoth(e.positionToBufferIndex(Position{Line: line, Column: col}))

	return nil
}

func (e *TextBuffer) ScrollIfNeeded() error {
	pos := e.getIndexPosition(e.Ranges[0].Start())
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

	for i := range e.Ranges {
		e.MoveRight(&e.Ranges[i], i*e.TabSize)
		if e.Ranges[i].Start() >= len(e.Content) { // end of file, appending
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Ranges[i].Start(),
				Data: []byte(strings.Repeat(" ", e.TabSize)),
			})
			e.Content = append(e.Content, []byte(strings.Repeat(" ", e.TabSize))...)
		} else {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Ranges[i].Start(),
				Data: []byte(strings.Repeat(" ", e.TabSize)),
			})
			e.Content = append(e.Content[:e.Ranges[i].Start()], append([]byte(strings.Repeat(" ", e.TabSize)), e.Content[e.Ranges[i].Start():]...)...)
		}
		e.MoveRight(&e.Ranges[i], e.TabSize)

	}
	e.SetStateDirty()

	return nil
}

func (e *TextBuffer) KillLine() error {
	if e.Readonly || len(e.Ranges) > 1 {
		return nil
	}
	var lastChange int
	for i := range e.Ranges {
		cur := &e.Ranges[i]
		old := len(e.Content)
		e.MoveLeft(cur, lastChange)
		line := e.getIndexVisualLine(cur.Start())
		writeToClipboard(e.Content[cur.Start():line.endIndex])
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  cur.Start(),
			Data: e.Content[cur.Start():line.endIndex],
		})
		e.Content = append(e.Content[:cur.Start()], e.Content[line.endIndex:]...)
		lastChange += -1 * (len(e.Content) - old)
	}
	e.SetStateDirty()

	return nil
}

func (e *TextBuffer) DeleteWordBackward() {
	if e.Readonly || len(e.Ranges) > 1 {
		return
	}
	cur := e.Ranges[0]
	previousWordEndIdx := byteutils.PreviousWordInBuffer(e.Content, cur.Start())
	oldLen := len(e.Content)
	if len(e.Content) > cur.Start()+1 {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  previousWordEndIdx + 1,
			Data: e.Content[previousWordEndIdx+1 : cur.Start()],
		})
		e.Content = append(e.Content[:previousWordEndIdx+1], e.Content[cur.Start()+1:]...)
	} else {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  previousWordEndIdx + 1,
			Data: e.Content[previousWordEndIdx+1:],
		})
		e.Content = e.Content[:previousWordEndIdx+1]
	}
	cur.AddToStart(len(e.Content) - oldLen)
	e.SetStateDirty()
}

func (e *TextBuffer) Copy() error {
	if len(e.Ranges) > 1 {
		return nil
	}
	cur := e.Ranges[0]
	if cur.Start() != cur.End() {
		// Copy selection
		writeToClipboard(e.Content[cur.Start():cur.End()])
	} else {
		line := e.getIndexVisualLine(cur.Start())
		writeToClipboard(e.Content[line.startIndex : line.endIndex+1])
	}

	return nil
}

func (e *TextBuffer) Cut() error {
	if e.Readonly || len(e.Ranges) > 1 {
		return nil
	}
	cur := &e.Ranges[0]
	if cur.Start() != cur.End() {
		// Copy selection
		writeToClipboard(e.Content[cur.Start():cur.End()])
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  cur.Start(),
			Data: e.Content[cur.Start():cur.End()],
		})
		e.Content = append(e.Content[:cur.Start()], e.Content[cur.End()+1:]...)
	} else {
		line := e.getIndexVisualLine(cur.Start())
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
	if e.Readonly || len(e.Ranges) > 1 {
		return nil
	}
	e.deleteSelectionsIfAnySelection()
	contentToPaste := getClipboardContent()
	cur := e.Ranges[0]
	if cur.Start() >= len(e.Content) { // end of file, appending
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  cur.Start(),
			Data: contentToPaste,
		})
		e.Content = append(e.Content, contentToPaste...)
	} else {
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  cur.Start(),
			Data: contentToPaste,
		})
		e.Content = append(e.Content[:cur.Start()], append(contentToPaste, e.Content[cur.Start():]...)...)
	}

	e.SetStateDirty()

	e.MoveAllRight(len(contentToPaste))
	return nil
}

var EditorKeymap = Keymap{
	Key{K: "=", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.parent.IncreaseFontSize(10)

		return nil
	}),

	Key{K: ".", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.PlaceSelectionOnNextMatch()
	}),
	Key{K: "<right>", Shift: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsToRight(e, 1)

		return nil
	}),
	Key{K: "<left>", Shift: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsToLeft(e, 1)

		return nil
	}),
	Key{K: "<up>", Shift: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsUp(e, 1)

		return nil
	}),
	Key{K: "<down>", Shift: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsDown(e, 1)

		return nil
	}),
	Key{K: "<lmouse>-click", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return PlaceAnotherSelectionHere(e, rl.GetMousePosition())
	}),
	Key{K: "<lmouse>-hold", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return PlaceAnotherSelectionHere(e, rl.GetMousePosition())
	}),
	Key{K: "<up>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return PlaceAnotherCursorPreviousLine(e)
	}),

	Key{K: "<down>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return PlaceAnotherCursorNextLine(e)
	}),
	Key{K: "-", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.parent.DecreaseFontSize(10)

		return nil
	}),
	Key{K: "r", Alt: true}: MakeCommand(func(e *TextBuffer) error {
		return e.readFileFromDisk()
	}),
	Key{K: "/", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.PopAndReverseLastAction()
		return nil
	}),
	Key{K: "z", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.PopAndReverseLastAction()
		return nil
	}),
	Key{K: "f", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveAllRight(1)
	}),
	Key{K: "x", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.Cut()
	}),
	Key{K: "v", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.Paste()
	}),
	Key{K: "k", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.KillLine()
	}),
	Key{K: "g", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.keymaps = append(e.keymaps, GotoLineKeymap)
		e.isGotoLine = true

		return nil
	}),
	Key{K: "c", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.Copy()
	}),

	Key{K: "s", Control: true}: MakeCommand(func(a *TextBuffer) error {
		a.keymaps = append(a.keymaps, SearchTextBufferKeymap)
		return nil
	}),
	Key{K: "w", Control: true}: MakeCommand(func(a *TextBuffer) error {
		return a.Write()
	}),

	Key{K: "<esc>"}: MakeCommand(func(p *TextBuffer) error {
		p.Ranges = p.Ranges[:1]

		return nil
	}),

	// navigation
	Key{K: "<lmouse>-click"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),

	Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(5)
	}),

	Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(5)
	}),

	Key{K: "<lmouse>-hold"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),

	Key{K: "a", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveToBeginningOfTheLine()
	}),
	Key{K: "e", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveToEndOfTheLine()
	}),

	Key{K: "p", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.PreviousLine()
	}),

	Key{K: "n", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.NextLine()
	}),

	Key{K: "<up>"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveUp()
	}),
	Key{K: "<down>"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveDown()
	}),
	Key{K: "<right>"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveAllRight(1)
	}),
	Key{K: "<right>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.NextWord()
	}),
	Key{K: "<left>"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveAllLeft(1)
	}),
	Key{K: "<left>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.PreviousWord()
	}),

	Key{K: "b", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveAllLeft(1)
	}),
	Key{K: "<home>"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveToBeginningOfTheLine()
	}),
	Key{K: "<pagedown>"}: MakeCommand(func(e *TextBuffer) error {
		return e.ScrollDown(1)
	}),
	Key{K: "<pageup>"}: MakeCommand(func(e *TextBuffer) error {
		return e.ScrollUp(1)
	}),

	//insertion
	Key{K: "<enter>"}: MakeCommand(func(e *TextBuffer) error { return insertChar(e, '\n') }),
	Key{K: "<space>"}: MakeCommand(func(e *TextBuffer) error { return insertChar(e, ' ') }),
	Key{K: "<backspace>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.DeleteWordBackward()
		return nil
	}),
	Key{K: "<backspace>"}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackward()
	}),
	Key{K: "d", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),
	Key{K: "d", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),
	Key{K: "<delete>"}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),
	Key{K: "a"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'a') }),
	Key{K: "b"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'b') }),
	Key{K: "c"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'c') }),
	Key{K: "d"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'd') }),
	Key{K: "e"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'e') }),
	Key{K: "f"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'f') }),
	Key{K: "g"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'g') }),
	Key{K: "h"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'h') }),
	Key{K: "i"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'i') }),
	Key{K: "j"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'j') }),
	Key{K: "k"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'k') }),
	Key{K: "l"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'l') }),
	Key{K: "m"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'm') }),
	Key{K: "n"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'n') }),
	Key{K: "o"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'o') }),
	Key{K: "p"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'p') }),
	Key{K: "q"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'q') }),
	Key{K: "r"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'r') }),
	Key{K: "s"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 's') }),
	Key{K: "t"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 't') }),
	Key{K: "u"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'u') }),
	Key{K: "v"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'v') }),
	Key{K: "w"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'w') }),
	Key{K: "x"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'x') }),
	Key{K: "y"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'y') }),
	Key{K: "z"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'z') }),
	Key{K: "0"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '0') }),
	Key{K: "1"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '1') }),
	Key{K: "2"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '2') }),
	Key{K: "3"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '3') }),
	Key{K: "4"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '4') }),
	Key{K: "5"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '5') }),
	Key{K: "6"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '6') }),
	Key{K: "7"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '7') }),
	Key{K: "8"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '8') }),
	Key{K: "9"}:               MakeCommand(func(e *TextBuffer) error { return insertChar(e, '9') }),
	Key{K: "\\"}:              MakeCommand(func(e *TextBuffer) error { return insertChar(e, '\\') }),
	Key{K: "\\", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertChar(e, '|') }),

	Key{K: "0", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, ')') }),
	Key{K: "1", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '!') }),
	Key{K: "2", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '@') }),
	Key{K: "3", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '#') }),
	Key{K: "4", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '$') }),
	Key{K: "5", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '%') }),
	Key{K: "6", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '^') }),
	Key{K: "7", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '&') }),
	Key{K: "8", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '*') }),
	Key{K: "9", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '(') }),
	Key{K: "a", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'A') }),
	Key{K: "b", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'B') }),
	Key{K: "c", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'C') }),
	Key{K: "d", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'D') }),
	Key{K: "e", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'E') }),
	Key{K: "f", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'F') }),
	Key{K: "g", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'G') }),
	Key{K: "h", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'H') }),
	Key{K: "i", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'I') }),
	Key{K: "j", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'J') }),
	Key{K: "k", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'K') }),
	Key{K: "l", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'L') }),
	Key{K: "m", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'M') }),
	Key{K: "n", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'N') }),
	Key{K: "o", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'O') }),
	Key{K: "p", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'P') }),
	Key{K: "q", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'Q') }),
	Key{K: "r", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'R') }),
	Key{K: "s", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'S') }),
	Key{K: "t", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'T') }),
	Key{K: "u", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'U') }),
	Key{K: "v", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'V') }),
	Key{K: "w", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'W') }),
	Key{K: "x", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'X') }),
	Key{K: "y", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'Y') }),
	Key{K: "z", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, 'Z') }),
	Key{K: "["}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '[') }),
	Key{K: "]"}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, ']') }),
	Key{K: "[", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '{') }),
	Key{K: "]", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '}') }),
	Key{K: ";"}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, ';') }),
	Key{K: ";", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, ':') }),
	Key{K: "'"}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '\'') }),
	Key{K: "'", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '"') }),
	Key{K: "\""}:                   MakeCommand(func(e *TextBuffer) error { return insertChar(e, '"') }),
	Key{K: ","}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, ',') }),
	Key{K: "."}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '.') }),
	Key{K: ",", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '<') }),
	Key{K: ".", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '>') }),
	Key{K: "/"}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '/') }),
	Key{K: "/", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '?') }),
	Key{K: "-"}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '-') }),
	Key{K: "="}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '=') }),
	Key{K: "-", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '_') }),
	Key{K: "=", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '+') }),
	Key{K: "`"}:                    MakeCommand(func(e *TextBuffer) error { return insertChar(e, '`') }),
	Key{K: "`", Shift: true}:       MakeCommand(func(e *TextBuffer) error { return insertChar(e, '~') }),
	Key{K: "<space>", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertChar(e, ' ') }),
	Key{K: "<tab>"}:                MakeCommand(func(e *TextBuffer) error { return e.Indent() }),
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
	Key{K: "<space>"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ' ') }),
	Key{K: "<backspace>"}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackwardFromActiveSearch()
	}),
	Key{K: "<enter>"}: MakeCommand(func(editor *TextBuffer) error {
		editor.CurrentMatch++
		if editor.CurrentMatch >= len(editor.SearchMatches) {
			editor.CurrentMatch = 0
		}
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),

	Key{K: "<enter>", Control: true}: MakeCommand(func(editor *TextBuffer) error {
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
	Key{K: "<esc>"}: MakeCommand(func(editor *TextBuffer) error {
		editor.keymaps = editor.keymaps[:len(editor.keymaps)-1]
		editor.IsSearching = false
		editor.LastSearchString = ""
		editor.SearchString = nil
		editor.SearchMatches = nil
		editor.CurrentMatch = 0
		editor.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<lmouse>-click"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),
	Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true
		return e.ScrollUp(20)

	}),
	Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true

		return e.ScrollDown(20)
	}),

	Key{K: "<rmouse>-click"}: MakeCommand(func(editor *TextBuffer) error {
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
	Key{K: "<mmouse>-click"}: MakeCommand(func(editor *TextBuffer) error {
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
	Key{K: "<pagedown>"}: MakeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true
		return e.ScrollDown(1)
	}),
	Key{K: "<pageup>"}: MakeCommand(func(e *TextBuffer) error {
		e.MovedAwayFromCurrentMatch = true

		return e.ScrollUp(1)
	}),

	Key{K: "a"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'a') }),
	Key{K: "b"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'b') }),
	Key{K: "c"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'c') }),
	Key{K: "d"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'd') }),
	Key{K: "e"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'e') }),
	Key{K: "f"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'f') }),
	Key{K: "g"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'g') }),
	Key{K: "h"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'h') }),
	Key{K: "i"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'i') }),
	Key{K: "j"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'j') }),
	Key{K: "k"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'k') }),
	Key{K: "l"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'l') }),
	Key{K: "m"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'm') }),
	Key{K: "n"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'n') }),
	Key{K: "o"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'o') }),
	Key{K: "p"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'p') }),
	Key{K: "q"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'q') }),
	Key{K: "r"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'r') }),
	Key{K: "s"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 's') }),
	Key{K: "t"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 't') }),
	Key{K: "u"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'u') }),
	Key{K: "v"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'v') }),
	Key{K: "w"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'w') }),
	Key{K: "x"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'x') }),
	Key{K: "y"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'y') }),
	Key{K: "z"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'z') }),
	Key{K: "0"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '0') }),
	Key{K: "1"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '1') }),
	Key{K: "2"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '2') }),
	Key{K: "3"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '3') }),
	Key{K: "4"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '4') }),
	Key{K: "5"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '5') }),
	Key{K: "6"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '6') }),
	Key{K: "7"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '7') }),
	Key{K: "8"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '8') }),
	Key{K: "9"}:               MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '9') }),
	Key{K: "\\"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '\\') }),
	Key{K: "\\", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '|') }),

	Key{K: "0", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ')') }),
	Key{K: "1", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '!') }),
	Key{K: "2", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '@') }),
	Key{K: "3", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '#') }),
	Key{K: "4", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '$') }),
	Key{K: "5", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '%') }),
	Key{K: "6", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '^') }),
	Key{K: "7", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '&') }),
	Key{K: "8", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '*') }),
	Key{K: "9", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '(') }),
	Key{K: "a", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'A') }),
	Key{K: "b", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'B') }),
	Key{K: "c", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'C') }),
	Key{K: "d", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'D') }),
	Key{K: "e", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'E') }),
	Key{K: "f", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'F') }),
	Key{K: "g", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'G') }),
	Key{K: "h", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'H') }),
	Key{K: "i", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'I') }),
	Key{K: "j", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'J') }),
	Key{K: "k", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'K') }),
	Key{K: "l", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'L') }),
	Key{K: "m", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'M') }),
	Key{K: "n", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'N') }),
	Key{K: "o", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'O') }),
	Key{K: "p", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'P') }),
	Key{K: "q", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'Q') }),
	Key{K: "r", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'R') }),
	Key{K: "s", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'S') }),
	Key{K: "t", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'T') }),
	Key{K: "u", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'U') }),
	Key{K: "v", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'V') }),
	Key{K: "w", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'W') }),
	Key{K: "x", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'X') }),
	Key{K: "y", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'Y') }),
	Key{K: "z", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, 'Z') }),
	Key{K: "["}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '[') }),
	Key{K: "]"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ']') }),
	Key{K: "[", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '{') }),
	Key{K: "]", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '}') }),
	Key{K: ";"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ';') }),
	Key{K: ";", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ':') }),
	Key{K: "'"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '\'') }),
	Key{K: "'", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '"') }),
	Key{K: "\""}:             MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '"') }),
	Key{K: ","}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, ',') }),
	Key{K: "."}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '.') }),
	Key{K: ",", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '<') }),
	Key{K: ".", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '>') }),
	Key{K: "/"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '/') }),
	Key{K: "/", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '?') }),
	Key{K: "-"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '-') }),
	Key{K: "="}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '=') }),
	Key{K: "-", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '_') }),
	Key{K: "=", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '+') }),
	Key{K: "`"}:              MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '`') }),
	Key{K: "`", Shift: true}: MakeCommand(func(e *TextBuffer) error { return insertCharAtSearchString(e, '~') }),
}

var GotoLineKeymap = Keymap{
	Key{K: "<backspace>"}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackwardFromActiveSearch()
	}),
	Key{K: "<enter>"}: MakeCommand(func(editor *TextBuffer) error {
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
					editor.Ranges[0].SetBoth(line.startIndex)
				}
			}
		}

		return nil
	}),

	Key{K: "<esc>"}: MakeCommand(func(editor *TextBuffer) error {
		editor.keymaps = editor.keymaps[:len(editor.keymaps)-1]
		editor.isGotoLine = false
		return nil
	}),

	Key{K: "0"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '0') }),
	Key{K: "1"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '1') }),
	Key{K: "2"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '2') }),
	Key{K: "3"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '3') }),
	Key{K: "4"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '4') }),
	Key{K: "5"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '5') }),
	Key{K: "6"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '6') }),
	Key{K: "7"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '7') }),
	Key{K: "8"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '8') }),
	Key{K: "9"}: MakeCommand(func(e *TextBuffer) error { return insertCharAtGotoLineBuffer(e, '9') }),
}
