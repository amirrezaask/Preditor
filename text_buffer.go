package preditor

import (
	"bytes"
	"errors"
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

	"github.com/amirrezaask/preditor/byteutils"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
)

const (

	//this is a comment
	State_Clean = 1
	State_Dirty = 2
)

type Cursor struct {
	Point int
	Mark  int
}

func (r *Cursor) SetBoth(n int) {
	r.Point = n
	r.Mark = n
}
func (r *Cursor) AddToBoth(n int) {
	r.Point += n
	r.Mark += n
}
func (r *Cursor) AddToStart(n int) {
	if r.Point > r.Mark {
		r.Mark += n
	} else {
		r.Point += n
	}
}
func (r *Cursor) AddToEnd(n int) {
	if r.Point > r.Mark {
		r.Point += n
	} else {
		r.Mark += n
	}
}
func (r *Cursor) Start() int {
	if r.Point > r.Mark {
		return r.Mark
	} else {
		return r.Point
	}
}
func (r *Cursor) End() int {
	if r.Point > r.Mark {
		return r.Point
	} else {
		return r.Mark
	}
}

type View struct {
	StartLine int32
	EndLine   int32
	Lines     []visualLine
}

type ISearch struct {
	IsSearching               bool
	LastSearchString          string
	SearchString              *string
	SearchMatches             [][]int
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool
}

type Gotoline struct {
	IsGotoLine        bool
	GotoLineUserInput []byte
}

type Compilation struct {
	IsActive                  bool
	CompilationCommand        []byte
	CompilationOutputBufferID int
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
	maxLine        int32
	maxColumn      int32

	keymaps []Keymap

	HasSyntaxHighlights bool
	SyntaxHighlights    SyntaxHighlights

	TabSize int

	View View

	// Cursor
	Cursors []Cursor

	// Searching
	ISearch ISearch

	UndoStack Stack[EditorAction]

	//Gotoline
	GotoLine Gotoline

	//Compilation
	Compilation Compilation

	LastCursorBlink time.Time
}

const (
	EditorActionType_Insert = iota + 1
	EditorActionType_Delete
)

type EditorAction struct {
	Type       int
	Idx        int
	Selections []Cursor
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
	a.Selections = e.Cursors
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
	e.calculateVisualLines()
}

func (e *TextBuffer) SetStateClean() {
	e.State = State_Clean
}

func (e *TextBuffer) replaceTabsWithSpaces() {
	e.Content = bytes.Replace(e.Content, []byte("\t"), []byte(strings.Repeat(" ", e.TabSize)), -1)
}

func (e *TextBuffer) updateMaxLineAndColumn(maxH float64, maxW float64) {
	oldMaxLine := e.maxLine
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	e.maxColumn = int32(maxW / float64(charSize.X))
	e.maxLine = int32(maxH / float64(charSize.Y))

	// reserve one line forr status bar
	e.maxLine--

	diff := e.maxLine - oldMaxLine
	e.View.EndLine += diff
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
					t.Cursors = []Cursor{{Point: t.positionToBufferIndex(*startingPos), Mark: t.positionToBufferIndex(*startingPos)}}
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
		tb.Cursors = []Cursor{{Point: tb.positionToBufferIndex(*startingPos), Mark: tb.positionToBufferIndex(*startingPos)}}
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
	t.keymaps = append([]Keymap{}, EditorKeymap, MakeInsertionKeys[*TextBuffer](func(b byte) error {
		return t.InsertCharAtCursor(b)
	}))
	t.UndoStack = NewStack[EditorAction](1000)
	t.TabSize = t.cfg.TabSize
	t.Cursors = append(t.Cursors, Cursor{Point: 0, Mark: 0})
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
	for _, line := range e.View.Lines {
		if line.startIndex <= i && line.endIndex >= i {
			return line
		}
	}

	if len(e.View.Lines) > 0 {
		lastLine := e.View.Lines[len(e.View.Lines)-1]
		lastLine.endIndex++
		return lastLine
	}
	return visualLine{}
}

func (e *TextBuffer) getIndexPosition(i int) Position {
	if len(e.View.Lines) == 0 {
		return Position{Line: 0, Column: i}
	}

	line := e.getIndexVisualLine(i)
	return Position{
		Line:   line.Index,
		Column: i - line.startIndex,
	}

}

func (e *TextBuffer) positionToBufferIndex(pos Position) int {
	if len(e.View.Lines) <= pos.Line {
		return len(e.Content)
	}

	return e.View.Lines[pos.Line].startIndex + pos.Column
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
	for rx, color := range e.SyntaxHighlights {
		indexes := rx.FindAllStringIndex(string(bs), -1)
		for _, index := range indexes {
			highlights = append(highlights, highlight{
				start: index[0] + offset,
				end:   index[1] + offset - 1,
				Color: color,
			})
		}

	}
	return highlights
}

func sortme[T any](slice []T, pred func(t1 T, t2 T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return pred(slice[i], slice[j])
	})
}

func (e *TextBuffer) calculateVisualLines() {
	e.View.Lines = []visualLine{}
	totalVisualLines := 0
	lineCharCounter := 0
	var actualLineIndex int
	var start int
	if e.View.EndLine == 0 {
		e.View.EndLine = e.maxLine
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
			e.View.Lines = append(e.View.Lines, line)
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
			e.View.Lines = append(e.View.Lines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1
			continue
		}

		if int32(lineCharCounter) > e.maxColumn-int32(len(fmt.Sprint(totalVisualLines)))-1 {
			line := visualLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start + 1,
				ActualLine: actualLineIndex,
			}

			e.View.Lines = append(e.View.Lines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx + 1
			continue
		}
	}
}

func (e *TextBuffer) renderCursors(zeroLocation rl.Vector2, maxH float64, maxW float64) {

	//TODO:
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	for _, sel := range e.Cursors {
		if sel.Start() == sel.End() {
			cursor := e.getIndexPosition(sel.Start())
			cursorView := Position{
				Line:   cursor.Line - int(e.View.StartLine),
				Column: cursor.Column,
			}
			posX := int32(cursorView.Column)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				if len(e.View.Lines) > cursor.Line {
					posX += int32((len(fmt.Sprint(e.View.Lines[cursor.Line].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			switch e.cfg.CursorShape {
			case CURSOR_SHAPE_OUTLINE:
				rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), int32(charSize.X), int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor)
			case CURSOR_SHAPE_BLOCK:
				rl.DrawRectangle(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().Cursor, 0.5))
			case CURSOR_SHAPE_LINE:
				rl.DrawRectangleLines(posX, int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), 2, int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor)
			}
			if e.cfg.CursorLineHighlight {
				rl.DrawRectangle(int32(zeroLocation.X), int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), e.maxColumn*int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().CursorLineBackground, 0.2))
			}

		} else {
			e.highlightBetweenTwoIndexes(zeroLocation, sel.Start(), sel.End(), e.cfg.CurrentThemeColors().Selection)
		}

	}

}

func (e *TextBuffer) renderStatusbar(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)

	var sections []string
	if len(e.Cursors) > 1 {
		sections = append(sections, fmt.Sprintf("%d#Cursors", len(e.Cursors)))
	} else {
		if e.Cursors[0].Start() == e.Cursors[0].End() {
			selStart := e.getIndexPosition(e.Cursors[0].Start())
			if len(e.View.Lines) > selStart.Line {
				selLine := e.View.Lines[selStart.Line]
				sections = append(sections, fmt.Sprintf("Line#%d Col#%d", selLine.ActualLine, selStart.Column))
			} else {
				sections = append(sections, fmt.Sprintf("Line#%d Col#%d", selStart.Line, selStart.Column))
			}

		} else {
			selEnd := e.getIndexPosition(e.Cursors[0].End())
			sections = append(sections, fmt.Sprintf("Line#%d Col#%d (Selected %d)", selEnd.Line, selEnd.Column, int(math.Abs(float64(e.Cursors[0].Start()-e.Cursors[0].End())))))
		}

	}

	file := e.File

	var state string
	if e.State == State_Dirty {
		state = "U"
	} else {
		state = ""
	}

	var isActiveWindow string
	for _, col := range e.parent.Windows {
		for _, win := range col {
			if win.ID == e.parent.ActiveWindowIndex {
				if win.BufferID == e.ID {
					isActiveWindow = "@"
				}
			}
		}
	}

	//render status bar
	rl.DrawRectangle(
		int32(zeroLocation.X),
		int32(zeroLocation.Y),
		int32(maxW),
		int32(charSize.Y),
		e.cfg.CurrentThemeColors().StatusBarBackground,
	)
	sections = append(sections, fmt.Sprintf("%s %s", state, file))
	sections = append(sections, isActiveWindow)
	rl.DrawTextEx(e.parent.Font,
		strings.Join(sections, " "),
		rl.Vector2{X: zeroLocation.X, Y: float32(zeroLocation.Y)},
		float32(e.parent.FontSize),
		0,
		e.cfg.CurrentThemeColors().StatusBarForeground)

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
		if len(e.View.Lines) <= i {
			break
		}
		var thisLineEnd int
		var thisLineStart int
		line := e.View.Lines[i]
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
				if len(e.View.Lines) > i {
					posX += int32((len(fmt.Sprint(e.View.Lines[i].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			rl.DrawRectangle(posX, int32(i-int(e.View.StartLine))*int32(charSize.Y)+int32(zeroLocation.Y), int32(charSize.X), int32(charSize.Y), rl.Fade(color, 0.5))
		}
	}

}

func (e *TextBuffer) renderText(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	var visibleLines []visualLine
	if e.View.EndLine > int32(len(e.View.Lines)) {
		visibleLines = e.View.Lines[e.View.StartLine:]
	} else {
		visibleLines = e.View.Lines[e.View.StartLine:e.View.EndLine]
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	for idx, line := range visibleLines {
		if e.visualLineShouldBeRendered(line) {
			var lineNumberWidth int
			if e.cfg.LineNumbers {
				lineNumberWidth = (len(fmt.Sprint(line.ActualLine)) + 1) * int(charSize.X)
				rl.DrawTextEx(e.parent.Font,
					fmt.Sprintf("%d", line.ActualLine),
					rl.Vector2{X: zeroLocation.X, Y: zeroLocation.Y + float32(idx)*charSize.Y},
					float32(e.parent.FontSize),
					0,
					e.cfg.CurrentThemeColors().LineNumbersForeground)

			}
			rl.DrawTextEx(e.parent.Font,
				string(e.Content[line.startIndex:line.endIndex]),
				rl.Vector2{X: zeroLocation.X + float32(lineNumberWidth), Y: zeroLocation.Y + float32(idx)*charSize.Y},
				float32(e.parent.FontSize),
				0,
				e.cfg.CurrentThemeColors().Foreground)

			if e.cfg.EnableSyntaxHighlighting && e.HasSyntaxHighlights {
				highlights := e.calculateHighlights(e.Content[line.startIndex:line.endIndex], line.startIndex)
				for _, h := range highlights {
					rl.DrawTextEx(e.parent.Font,
						string(e.Content[h.start:h.end+1]),
						rl.Vector2{X: zeroLocation.X + float32(lineNumberWidth) + float32(h.start-line.startIndex)*charSize.X, Y: zeroLocation.Y + float32(idx)*charSize.Y},
						float32(e.parent.FontSize),
						0,
						h.Color)

				}
			}
		}
	}
}
func (e *TextBuffer) convertBufferIndexToLineAndColumn(idx int) *Position {
	for lineIndex, line := range e.View.Lines {
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
	e.ISearch.SearchMatches = [][]int{}
	matchPatternAsync(&e.ISearch.SearchMatches, e.Content, []byte(pattern))
	return nil
}

func (e *TextBuffer) findMatchesAndHighlight(pattern string, zeroLocation rl.Vector2) error {
	if pattern != e.ISearch.LastSearchString {
		if err := e.findMatches(pattern); err != nil {
			return err
		}
	}
	for idx, match := range e.ISearch.SearchMatches {
		c := e.cfg.CurrentThemeColors().Selection
		_ = c
		if idx == e.ISearch.CurrentMatch {
			c = rl.Fade(rl.Red, 0.5)
			matchStartLine := e.getIndexPosition(match[0])
			matchEndLine := e.getIndexPosition(match[0])
			if !(e.View.StartLine < int32(matchStartLine.Line) && e.View.EndLine > int32(matchEndLine.Line)) && !e.ISearch.MovedAwayFromCurrentMatch {
				// current match is not in view
				// move the view
				oldStart := e.View.StartLine
				e.View.StartLine = int32(matchStartLine.Line) - e.maxLine/2
				if e.View.StartLine < 0 {
					e.View.StartLine = int32(matchStartLine.Line)
				}

				diff := e.View.StartLine - oldStart
				e.View.EndLine += diff
			}
		}
		e.highlightBetweenTwoIndexes(zeroLocation, match[0], match[1], c)
	}
	e.ISearch.LastSearchString = pattern

	return nil
}
func (e *TextBuffer) renderSearch(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if e.ISearch.SearchString == nil || len(*e.ISearch.SearchString) < 1 {
		return
	}
	e.findMatchesAndHighlight(*e.ISearch.SearchString, zeroLocation)
	if len(e.ISearch.SearchMatches) > 0 {
		e.Cursors = e.Cursors[:1]
		e.Cursors[0].Point = e.ISearch.SearchMatches[e.ISearch.CurrentMatch][0]
		e.Cursors[0].Mark = e.ISearch.SearchMatches[e.ISearch.CurrentMatch][0]
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts)
	rl.DrawTextEx(e.parent.Font, fmt.Sprintf("ISearch: %s", *e.ISearch.SearchString), rl.Vector2{
		X: zeroLocation.X,
		Y: zeroLocation.Y,
	}, float32(e.parent.FontSize), 0, rl.White)
}

func (e *TextBuffer) renderGoto(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if !e.GotoLine.IsGotoLine {
		return
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts)
	rl.DrawTextEx(e.parent.Font, fmt.Sprintf("Goto: %s", string(e.GotoLine.GotoLineUserInput)), rl.Vector2{
		X: zeroLocation.X,
		Y: zeroLocation.Y,
	}, float32(e.parent.FontSize), 0, rl.White)
}

func (e *TextBuffer) renderCompilation(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if !e.Compilation.IsActive {
		return
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts)
	rl.DrawTextEx(e.parent.Font, fmt.Sprintf("Compile: %s", string(e.Compilation.CompilationCommand)), rl.Vector2{
		X: zeroLocation.X,
		Y: zeroLocation.Y,
	}, float32(e.parent.FontSize), 0, rl.White)
}

func (e *TextBuffer) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	e.updateMaxLineAndColumn(maxH, maxW)
	e.calculateVisualLines()

	e.renderStatusbar(zeroLocation, maxH, maxW)
	zeroLocation.Y += measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0).Y
	e.renderText(zeroLocation, maxH, maxW)
	e.renderSearch(zeroLocation, maxH, maxW)
	e.renderGoto(zeroLocation, maxH, maxW)
	e.renderCompilation(zeroLocation, maxH, maxW)
	e.renderCursors(zeroLocation, maxH, maxW)
}

func (e *TextBuffer) visualLineShouldBeRendered(line visualLine) bool {
	if e.View.StartLine <= int32(line.Index) && line.Index <= int(e.View.EndLine) {
		return true
	}

	return false
}

func (e *TextBuffer) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if len(e.View.Lines) == 0 && newPosition.Line == 0 && newPosition.Column >= 0 && int32(newPosition.Column) < e.maxColumn-int32(len(fmt.Sprint(newPosition.Line)))-1 {
		return true
	}
	if newPosition.Line >= len(e.View.Lines) && (len(e.View.Lines) != 0) {
		return false
	}

	if newPosition.Column < 0 {
		return false
	}
	if newPosition.Column > e.View.Lines[newPosition.Line].Length+1 {
		return false
	}

	return true
}

func (e *TextBuffer) deleteSelectionsIfAnySelection() {
	if e.Readonly {
		return
	}
	old := len(e.Content)
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		if sel.Start() == sel.End() {
			continue
		}
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  sel.Start(),
			Data: e.Content[sel.Start() : sel.End()+1],
		})
		e.Content = append(e.Content[:sel.Start()], e.Content[sel.End()+1:]...)
		sel.Point = sel.Mark
		e.MoveLeft(sel, old-1-len(e.Content))
	}

}

func (e *TextBuffer) sortSelections() {
	sortme(e.Cursors, func(t1 Cursor, t2 Cursor) bool {
		return t1.Start() < t2.Start()
	})
}
func (e *TextBuffer) GetLastSelection() *Cursor {
	return &e.Cursors[len(e.Cursors)-1]
}

func (e *TextBuffer) removeDuplicateSelectionsAndSort() {
	selections := map[string]struct{}{}
	for i := range e.Cursors {
		selections[fmt.Sprintf("%d:%d", e.Cursors[i].Start(), e.Cursors[i].End())] = struct{}{}
	}

	e.Cursors = nil
	for k := range selections {
		start, _ := strconv.Atoi(strings.Split(k, ":")[0])
		end, _ := strconv.Atoi(strings.Split(k, ":")[1])
		e.Cursors = append(e.Cursors, Cursor{Point: start, Mark: end})
	}

	e.sortSelections()
}

func (e *TextBuffer) InsertCharAtCursor(char byte) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Cursors {
		e.MoveRight(&e.Cursors[i], i*1)

		if e.Cursors[i].Start() >= len(e.Content) { // end of file, appending
			e.Content = append(e.Content, char)

		} else {
			e.Content = append(e.Content[:e.Cursors[i].Start()+1], e.Content[e.Cursors[i].End():]...)
			e.Content[e.Cursors[i].Start()] = char
		}
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Insert,
			Idx:  e.Cursors[i].Start(),
			Data: []byte{char},
		})
		e.MoveRight(&e.Cursors[i], 1)
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
	for i := range e.Cursors {
		e.MoveLeft(&e.Cursors[i], i*1)

		switch {
		case e.Cursors[i].Start() == 0:
			continue
		case e.Cursors[i].Start() < len(e.Content):
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Cursors[i].Start() - 1,
				Data: []byte{e.Content[e.Cursors[i].Start()-1]},
			})
			e.Content = append(e.Content[:e.Cursors[i].Start()-1], e.Content[e.Cursors[i].Start():]...)
		case e.Cursors[i].Start() == len(e.Content):
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  e.Cursors[i].Start() - 1,
				Data: []byte{e.Content[e.Cursors[i].Start()-1]},
			})
			e.Content = e.Content[:e.Cursors[i].Start()-1]
		}

		e.MoveLeft(&e.Cursors[i], 1)

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
	for i := range e.Cursors {
		if len(e.Content) > e.Cursors[i].Start()+1 {
			e.MoveLeft(&e.Cursors[i], i*1)
			//e.AddUndoAction(EditorAction{
			//	Type: EditorActionType_Delete,
			//	Idx:  e.Cursors[i].Start(),
			//	Data: []byte{e.Content[e.Cursors[i].Start()]},
			//})
			e.Content = append(e.Content[:e.Cursors[i].Start()], e.Content[e.Cursors[i].Start()+1:]...)
			e.SetStateDirty()
		}
	}

	return nil
}

func (e *TextBuffer) ScrollUp(n int) error {
	if e.View.StartLine <= 0 {
		return nil
	}
	e.View.EndLine += int32(-1 * n)
	e.View.StartLine += int32(-1 * n)

	diff := e.View.EndLine - e.View.StartLine

	if e.View.StartLine < 0 {
		e.View.StartLine = 0
		e.View.EndLine = diff
	}

	return nil

}

func (e *TextBuffer) CenteralizeCursor() error {
	cur := e.Cursors[0]
	pos := e.convertBufferIndexToLineAndColumn(cur.Start())
	e.View.StartLine = int32(pos.Line) - (e.maxLine / 2)
	e.View.EndLine = int32(pos.Line) + (e.maxLine / 2)
	if e.View.StartLine < 0 {
		e.View.StartLine = 0
		e.View.EndLine = e.maxLine
	}
	return nil
}

func (e *TextBuffer) ScrollToTop() error {
	e.View.StartLine = 0
	e.View.EndLine = e.maxLine
	e.Cursors[0].SetBoth(0)

	return nil
}

func (e *TextBuffer) ScrollToBottom() error {
	e.View.StartLine = int32(len(e.View.Lines) - 1 - int(e.maxLine))
	e.View.EndLine = int32(len(e.View.Lines) - 1)
	e.Cursors[0].SetBoth(e.View.Lines[len(e.View.Lines)-1].startIndex)

	return nil
}

func (e *TextBuffer) ScrollDown(n int) error {
	if int(e.View.EndLine) >= len(e.View.Lines) {
		return nil
	}
	e.View.EndLine += int32(n)
	e.View.StartLine += int32(n)
	diff := e.View.EndLine - e.View.StartLine
	if int(e.View.EndLine) >= len(e.View.Lines) {
		e.View.EndLine = int32(len(e.View.Lines) - 1)
		e.View.StartLine = e.View.EndLine - diff
	}

	return nil

}

// Move* functions change Point part of cursor

func (e *TextBuffer) MoveLeft(s *Cursor, n int) {
	s.AddToBoth(-n)
	if s.Start() < 0 {
		s.SetBoth(0)
	}
}

func (e *TextBuffer) MoveAllLeft(n int) error {
	for i := range e.Cursors {
		e.MoveLeft(&e.Cursors[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}
func (e *TextBuffer) MoveRight(s *Cursor, n int) {
	e.Cursors[0].Point = e.Cursors[0].Mark
	s.AddToBoth(n)
	if s.Start() > len(e.Content) {
		s.SetBoth(len(e.Content))
	}
}

func (e *TextBuffer) MoveAllRight(n int) error {
	for i := range e.Cursors {
		e.MoveRight(&e.Cursors[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func (e *TextBuffer) MoveUp() error {
	for i := range e.Cursors {
		currentLine := e.getIndexVisualLine(e.Cursors[i].Point)
		prevLineIndex := currentLine.Index - 1
		if prevLineIndex < 0 {
			return nil
		}

		prevLine := e.View.Lines[prevLineIndex]
		col := e.Cursors[i].Point - currentLine.startIndex
		newidx := prevLine.startIndex + col
		if newidx > prevLine.endIndex {
			newidx = prevLine.endIndex
		}
		e.Cursors[i].SetBoth(newidx)
		e.ScrollIfNeeded()
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func (e *TextBuffer) MoveDown() error {
	for i := range e.Cursors {
		currentLine := e.getIndexVisualLine(e.Cursors[i].Point)
		nextLineIndex := currentLine.Index + 1
		if nextLineIndex >= len(e.View.Lines) {
			return nil
		}

		nextLine := e.View.Lines[nextLineIndex]
		col := e.Cursors[i].Point - currentLine.startIndex
		newIndex := nextLine.startIndex + col
		if newIndex > nextLine.endIndex {
			newIndex = nextLine.endIndex
		}
		e.Cursors[i].SetBoth(newIndex)
		e.ScrollIfNeeded()

	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func SelectionsToRight(e *TextBuffer, n int) error {
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		sel.Mark += n
		if sel.Mark >= len(e.Content) {
			sel.Mark = len(e.Content)
		}
	}

	return nil
}
func SelectionsToLeft(e *TextBuffer, n int) error {
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		sel.Mark -= n
		if sel.Mark < 0 {
			sel.Mark = 0
		}
	}

	return nil
}
func SelectionsUp(e *TextBuffer, n int) error {
	for i := range e.Cursors {
		currentLine := e.getIndexVisualLine(e.Cursors[i].Mark)
		nextLineIndex := currentLine.Index - n
		if nextLineIndex >= len(e.View.Lines) || nextLineIndex < 0 {
			return nil
		}

		nextLine := e.View.Lines[nextLineIndex]
		newcol := nextLine.startIndex
		e.Cursors[i].Mark = newcol
		e.ScrollIfNeeded()
	}

	return nil
}
func SelectionsDown(e *TextBuffer, n int) error {
	for i := range e.Cursors {
		currentLine := e.getIndexVisualLine(e.Cursors[i].Mark)
		nextLineIndex := currentLine.Index + n
		if nextLineIndex >= len(e.View.Lines) {
			return nil
		}

		nextLine := e.View.Lines[nextLineIndex]
		newcol := nextLine.startIndex
		e.Cursors[i].Mark = newcol
		e.ScrollIfNeeded()
	}

	return nil
}

func AnotherSelectionHere(e *TextBuffer, pos rl.Vector2) error {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64(pos.Y / charSize.Y))
	apprColumn := math.Floor(float64(pos.X / charSize.X))

	if e.cfg.LineNumbers {
		apprColumn -= float64(len(fmt.Sprint(apprLine)) + 1)
	}

	if len(e.View.Lines) < 1 {
		return nil
	}

	line := int(apprLine) + int(e.View.StartLine)
	col := int(apprColumn)

	if line >= len(e.View.Lines) {
		line = len(e.View.Lines) - 1
	}

	if line < 0 {
		line = 0
	}

	// check if cursor should be moved back
	if col > e.View.Lines[line].Length {
		col = e.View.Lines[line].Length
	}
	if col < 0 {
		col = 0
	}
	idx := e.positionToBufferIndex(Position{Line: line, Column: col})
	e.Cursors = append(e.Cursors, Cursor{Point: idx, Mark: idx})

	e.removeDuplicateSelectionsAndSort()
	return nil
}

func (e *TextBuffer) AnotherSelectionOnMatch() error {
	lastSel := e.Cursors[len(e.Cursors)-1]
	var thingToSearch []byte
	if lastSel.Point != lastSel.Mark {
		thingToSearch = e.Content[lastSel.Start():lastSel.End()]
		next := findNextMatch(e.Content, lastSel.End()+1, thingToSearch)
		if len(next) == 0 {
			return nil
		}
		e.Cursors = append(e.Cursors, Cursor{
			Point: next[0],
			Mark:  next[1],
		})

	} else {
		start := byteutils.SeekPreviousNonLetter(e.Content, lastSel.Point)
		end := byteutils.SeekNextNonLetter(e.Content, lastSel.Point)
		e.Cursors[len(e.Cursors)-1].Point = start + 1
		e.Cursors[len(e.Cursors)-1].Mark = start + 1
		thingToSearch = e.Content[start+1 : end]
		next := findNextMatch(e.Content, lastSel.End()+1, thingToSearch)
		if len(next) == 0 {
			return nil
		}
		e.Cursors = append(e.Cursors, Cursor{
			Point: next[0],
			Mark:  next[0],
		})

	}

	return nil
}
func SelectionPreviousWord(e *TextBuffer) error {
	for i := range e.Cursors {
		previousWord := byteutils.SeekPreviousNonLetter(e.Content, e.Cursors[i].Mark)
		if previousWord < 0 {
			continue
		}
		e.Cursors[i].Mark = previousWord
	}

	return nil
}
func SelectionNextWord(e *TextBuffer) error {
	for i := range e.Cursors {
		nextWord := byteutils.SeekNextNonLetter(e.Content, e.Cursors[i].Mark)
		if nextWord > len(e.Content) {
			continue
		}
		e.Cursors[i].Mark = nextWord
	}

	return nil
}

func (e *TextBuffer) SelectionEndOfLine() error {
	for i := range e.Cursors {
		line := e.getIndexVisualLine(e.Cursors[i].End())
		e.Cursors[i].Mark = line.endIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}
func (e *TextBuffer) SelectionBeginningOfLine() error {
	for i := range e.Cursors {
		line := e.getIndexVisualLine(e.Cursors[i].End())
		e.Cursors[i].Mark = line.startIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func (e *TextBuffer) MoveToBeginningOfTheLine() error {
	for i := range e.Cursors {
		line := e.getIndexVisualLine(e.Cursors[i].Start())
		e.Cursors[i].SetBoth(line.startIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func (e *TextBuffer) MoveToEndOfTheLine() error {
	for i := range e.Cursors {
		line := e.getIndexVisualLine(e.Cursors[i].Start())
		e.Cursors[i].SetBoth(line.endIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherCursorNextLine(e *TextBuffer) error {
	pos := e.getIndexPosition(e.Cursors[len(e.Cursors)-1].Start())
	pos.Line++
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Cursors = append(e.Cursors, Cursor{Point: newidx, Mark: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PlaceAnotherCursorPreviousLine(e *TextBuffer) error {
	pos := e.getIndexPosition(e.Cursors[len(e.Cursors)-1].Start())
	pos.Line--
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Cursors = append(e.Cursors, Cursor{Point: newidx, Mark: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func (e *TextBuffer) indexOfFirstNonLetter(bs []byte) int {

	for idx, b := range bs {
		if !unicode.IsLetter(rune(b)) {
			return idx
		}
	}

	return -1
}

func (e *TextBuffer) MoveToNextWord() error {
	for i := range e.Cursors {
		newidx := byteutils.NextWordInBuffer(e.Content, e.Cursors[i].Start())
		if newidx == -1 {
			return nil
		}
		if newidx > len(e.Content) {
			newidx = len(e.Content)
		}
		e.Cursors[i].SetBoth(newidx)

	}
	return nil
}

func (e *TextBuffer) MoveToPreviousWord() error {
	for i := range e.Cursors {
		newidx := byteutils.PreviousWordInBuffer(e.Content, e.Cursors[i].Start())
		if newidx == -1 {
			return nil
		}
		if newidx < 0 {
			newidx = 0
		}
		e.Cursors[i].SetBoth(newidx)

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

	if len(e.View.Lines) < 1 {
		return nil
	}

	line := int(apprLine) + int(e.View.StartLine) - 1
	col := int(apprColumn)
	if line >= len(e.View.Lines) {
		line = len(e.View.Lines) - 1
	}

	if line < 0 {
		line = 0
	}

	// check if cursor should be moved back
	if col > e.View.Lines[line].Length {
		col = e.View.Lines[line].Length
	}
	if col < 0 {
		col = 0
	}

	e.Cursors[0].SetBoth(e.positionToBufferIndex(Position{Line: line, Column: col}))

	return nil
}

func (e *TextBuffer) ScrollIfNeeded() error {
	pos := e.getIndexPosition(e.Cursors[0].End())
	if int32(pos.Line) <= e.View.StartLine {
		e.View.StartLine = int32(pos.Line) - e.maxLine/3
		e.View.EndLine = e.View.StartLine + e.maxLine

	} else if int32(pos.Line) >= e.View.EndLine {
		e.View.EndLine = int32(pos.Line) + e.maxLine/3
		e.View.StartLine = e.View.EndLine - e.maxLine
	}

	if int(e.View.EndLine) >= len(e.View.Lines) {
		e.View.EndLine = int32(len(e.View.Lines) - 1)
		e.View.StartLine = e.View.EndLine - e.maxLine
	}

	if e.View.StartLine < 0 {
		e.View.StartLine = 0
		e.View.EndLine = e.maxLine
	}
	if e.View.EndLine < 0 {
		e.View.StartLine = 0
		e.View.EndLine = e.maxLine
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
			continue
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

func (e *TextBuffer) Indent() error {
	e.removeDuplicateSelectionsAndSort()

	for i := range e.Cursors {
		e.MoveRight(&e.Cursors[i], i*e.TabSize)
		if e.Cursors[i].Start() >= len(e.Content) { // end of file, appending
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Cursors[i].Start(),
				Data: []byte(strings.Repeat(" ", e.TabSize)),
			})
			e.Content = append(e.Content, []byte(strings.Repeat(" ", e.TabSize))...)
		} else {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Cursors[i].Start(),
				Data: []byte(strings.Repeat(" ", e.TabSize)),
			})
			e.Content = append(e.Content[:e.Cursors[i].Start()], append([]byte(strings.Repeat(" ", e.TabSize)), e.Content[e.Cursors[i].Start():]...)...)
		}
		e.MoveRight(&e.Cursors[i], e.TabSize)

	}
	e.SetStateDirty()

	return nil
}

func (e *TextBuffer) KillLine() error {
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	var lastChange int
	for i := range e.Cursors {
		cur := &e.Cursors[i]
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
	if e.Readonly || len(e.Cursors) > 1 {
		return
	}
	cur := e.Cursors[0]
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
	if len(e.Cursors) > 1 {
		return nil
	}
	cur := e.Cursors[0]
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
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	cur := &e.Cursors[0]
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
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	e.deleteSelectionsIfAnySelection()
	contentToPaste := getClipboardContent()
	cur := e.Cursors[0]
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

	Key{K: ".", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.AnotherSelectionOnMatch()
	}),
	Key{K: ",", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.ScrollToTop()

		return nil
	}),
	Key{K: "l", Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.CenteralizeCursor()

		return nil
	}),

	Key{K: ".", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.ScrollToBottom()

		return nil
	}),
	Key{K: "<right>", Shift: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsToRight(e, 1)

		return nil
	}),
	Key{K: "<right>", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionNextWord(e)

		return nil
	}),
	Key{K: "<left>", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionPreviousWord(e)

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
	Key{K: "a", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.SelectionBeginningOfLine()

		return nil
	}),
	Key{K: "e", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		e.SelectionEndOfLine()

		return nil
	}),
	Key{K: "n", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsDown(e, 1)

		return nil
	}),
	Key{K: "p", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsUp(e, 1)

		return nil
	}),
	Key{K: "f", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsToRight(e, 1)

		return nil
	}),
	Key{K: "b", Shift: true, Control: true}: MakeCommand(func(e *TextBuffer) error {
		SelectionsToLeft(e, 1)

		return nil
	}),
	Key{K: "<lmouse>-click", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return AnotherSelectionHere(e, rl.GetMousePosition())
	}),
	Key{K: "<lmouse>-hold", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return AnotherSelectionHere(e, rl.GetMousePosition())
	}),
	Key{K: "<up>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return PlaceAnotherCursorPreviousLine(e)
	}),

	Key{K: "<down>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return PlaceAnotherCursorNextLine(e)
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
		e.GotoLine.IsGotoLine = true

		return nil
	}),
	Key{K: "c", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.Copy()
	}),

	Key{K: "c", Alt: true}: MakeCommand(func(a *TextBuffer) error {
		a.Compilation.IsActive = true
		a.keymaps = append(a.keymaps, CompilationKeymap, MakeInsertionKeys[*TextBuffer](func(b byte) error {
			a.Compilation.CompilationCommand = append(a.Compilation.CompilationCommand, b)

			return nil
		}))
		return nil
	}),

	Key{K: "s", Control: true}: MakeCommand(func(a *TextBuffer) error {
		a.keymaps = append(a.keymaps, SearchTextBufferKeymap, MakeInsertionKeys[*TextBuffer](func(b byte) error {
			if a.ISearch.SearchString == nil {
				a.ISearch.SearchString = new(string)
			}

			*a.ISearch.SearchString += string(b)

			return nil
		}))
		return nil
	}),
	Key{K: "w", Control: true}: MakeCommand(func(a *TextBuffer) error {
		return a.Write()
	}),

	Key{K: "<esc>"}: MakeCommand(func(p *TextBuffer) error {
		p.Cursors = p.Cursors[:1]

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
		return e.MoveUp()
	}),

	Key{K: "n", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveDown()
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
		return e.MoveToNextWord()
	}),
	Key{K: "<left>"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveAllLeft(1)
	}),
	Key{K: "<left>", Control: true}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveToPreviousWord()
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
	Key{K: "<enter>"}: MakeCommand(func(e *TextBuffer) error {
		return insertChar(e, '\n')
	}),
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
	Key{K: "<delete>"}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharForward()
	}),

	Key{K: "<tab>"}: MakeCommand(func(e *TextBuffer) error { return e.Indent() }),
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

func insertCharAtGotoLineBuffer(editor *TextBuffer, char byte) error {
	editor.GotoLine.GotoLineUserInput = append(editor.GotoLine.GotoLineUserInput, char)

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
	if e.ISearch.SearchString == nil {
		return nil
	}
	s := []byte(*e.ISearch.SearchString)
	if len(s) < 1 {
		return nil
	}
	s = s[:len(s)-1]

	e.ISearch.SearchString = &[]string{string(s)}[0]

	return nil
}

func (e *TextBuffer) DeleteCharBackwardFromGotoLine() error {
	if len(e.GotoLine.GotoLineUserInput) < 1 {
		return nil
	}
	e.GotoLine.GotoLineUserInput = e.GotoLine.GotoLineUserInput[:len(e.GotoLine.GotoLineUserInput)-1]

	return nil
}

var SearchTextBufferKeymap = Keymap{
	Key{K: "<backspace>"}: MakeCommand(func(e *TextBuffer) error {
		return e.DeleteCharBackwardFromActiveSearch()
	}),
	Key{K: "<enter>"}: MakeCommand(func(editor *TextBuffer) error {
		editor.ISearch.CurrentMatch++
		if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
			editor.ISearch.CurrentMatch = 0
		}
		editor.ISearch.MovedAwayFromCurrentMatch = false
		return nil
	}),

	Key{K: "<enter>", Control: true}: MakeCommand(func(editor *TextBuffer) error {
		editor.ISearch.CurrentMatch--
		if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
			editor.ISearch.CurrentMatch = 0
		}
		if editor.ISearch.CurrentMatch < 0 {
			editor.ISearch.CurrentMatch = len(editor.ISearch.SearchMatches) - 1
		}
		editor.ISearch.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<esc>"}: MakeCommand(func(editor *TextBuffer) error {
		editor.keymaps = editor.keymaps[:len(editor.keymaps)-2]
		editor.ISearch.IsSearching = false
		editor.ISearch.LastSearchString = ""
		editor.ISearch.SearchString = nil
		editor.ISearch.SearchMatches = nil
		editor.ISearch.CurrentMatch = 0
		editor.ISearch.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<lmouse>-click"}: MakeCommand(func(e *TextBuffer) error {
		return e.MoveCursorTo(rl.GetMousePosition())
	}),
	Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *TextBuffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true
		return e.ScrollUp(20)

	}),
	Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *TextBuffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true

		return e.ScrollDown(20)
	}),

	Key{K: "<rmouse>-click"}: MakeCommand(func(editor *TextBuffer) error {
		editor.ISearch.CurrentMatch++
		if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
			editor.ISearch.CurrentMatch = 0
		}
		if editor.ISearch.CurrentMatch < 0 {
			editor.ISearch.CurrentMatch = len(editor.ISearch.SearchMatches) - 1
		}
		editor.ISearch.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<mmouse>-click"}: MakeCommand(func(editor *TextBuffer) error {
		editor.ISearch.CurrentMatch--
		if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
			editor.ISearch.CurrentMatch = 0
		}
		if editor.ISearch.CurrentMatch < 0 {
			editor.ISearch.CurrentMatch = len(editor.ISearch.SearchMatches) - 1
		}
		editor.ISearch.MovedAwayFromCurrentMatch = false
		return nil
	}),
	Key{K: "<pagedown>"}: MakeCommand(func(e *TextBuffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true
		return e.ScrollDown(1)
	}),
	Key{K: "<pageup>"}: MakeCommand(func(e *TextBuffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true

		return e.ScrollUp(1)
	}),
}

var GotoLineKeymap = Keymap{
	Key{K: "<backspace>"}: MakeCommand(func(e *TextBuffer) error {
		if len(e.GotoLine.GotoLineUserInput) < 1 {
			return nil
		}
		e.GotoLine.GotoLineUserInput = e.GotoLine.GotoLineUserInput[:len(e.GotoLine.GotoLineUserInput)-1]

		return nil
	}),
	Key{K: "<enter>"}: MakeCommand(func(editor *TextBuffer) error {
		number, err := strconv.Atoi(string(editor.GotoLine.GotoLineUserInput))
		if err != nil {
			return nil
		}

		for _, line := range editor.View.Lines {
			if line.Index == number {
				editor.GotoLine.IsGotoLine = false
				editor.GotoLine.GotoLineUserInput = nil
				editor.Cursors[0].SetBoth(line.startIndex)
				editor.ScrollIfNeeded()
			}
		}

		return nil
	}),

	Key{K: "<esc>"}: MakeCommand(func(editor *TextBuffer) error {
		editor.keymaps = editor.keymaps[:len(editor.keymaps)-1]
		editor.GotoLine.IsGotoLine = false
		editor.GotoLine.GotoLineUserInput = nil
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

var CompilationKeymap Keymap

func init() {
	CompilationKeymap = Keymap{
		Key{K: "<enter>"}: MakeCommand(func(e *TextBuffer) error {
			e.Compilation.IsActive = false
			e.parent.openCompilationBuffer(string(e.Compilation.CompilationCommand))
			e.keymaps = e.keymaps[:len(e.keymaps)-2]
			return nil
		}),

		Key{K: "<backspace>"}: MakeCommand(func(e *TextBuffer) error {
			e.Compilation.CompilationCommand = e.Compilation.CompilationCommand[:len(e.Compilation.CompilationCommand)-1]

			return nil
		}),
		Key{K: "<esc>"}: MakeCommand(func(e *TextBuffer) error {
			e.Compilation.IsActive = false
			e.keymaps = e.keymaps[:len(e.keymaps)-2]

			return nil
		}),
	}

}
