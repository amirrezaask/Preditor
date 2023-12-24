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
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func SwitchOrOpenFileInTextBuffer(parent *Context, cfg *Config, filename string, startingPos *Position) error {
	for _, buf := range parent.Drawables {
		switch t := buf.(type) {
		case *Buffer:
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

	tb, err := NewBuffer(parent, cfg, filename)
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

func NewBuffer(parent *Context, cfg *Config, filename string) (*Buffer, error) {
	t := Buffer{cfg: cfg}
	t.parent = parent
	t.File = filename
	t.keymaps = append([]Keymap{}, EditorKeymap, MakeInsertionKeys(func(c *Context, b byte) error {
		return InsertChar(&t, b)
	}))
	t.UndoStack = NewStack[EditorAction](1000)
	t.Cursors = append(t.Cursors, Cursor{Point: 0, Mark: 0})
	var err error
	if t.File != "" {
		if _, err = os.Stat(t.File); err == nil {
			t.Content, err = os.ReadFile(t.File)
			if err != nil {
				return nil, err
			}
		}

		fileType, exists := FileTypes[path.Ext(t.File)]
		if exists {
			t.fileType = fileType
			t.LastCompileCommand = fileType.DefaultCompileCommand
			t.needParsing = true
		}
	}
	t.replaceTabsWithSpaces()
	return &t, nil

}

const (
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
	Lines     []BufferLine
}

type ISearch struct {
	IsSearching               bool
	LastSearchString          string
	SearchString              string
	SearchMatches             [][]int
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool
}

type Buffer struct {
	BaseDrawable
	cfg          *Config
	parent       *Context
	File         string
	Content      []byte
	State        int
	Readonly     bool
	maxLine      int32
	maxColumn    int32
	NoStatusbar  bool
	zeroLocation rl.Vector2

	Tokens []WordToken

	oldTSTree   *sitter.Tree
	highlights  []highlight
	needParsing bool
	fileType    FileType

	keymaps []Keymap

	View View

	// Cursor
	Cursors []Cursor

	// Searching
	ISearch ISearch

	LastCompileCommand string

	UndoStack Stack[EditorAction]
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

func (e *Buffer) String() string {
	return fmt.Sprintf("%s", e.File)
}

func (e *Buffer) Keymaps() []Keymap {
	return e.keymaps
}

func (e *Buffer) IsSpecial() bool {
	return e.File == "" || e.File[0] == '*'
}

func (e *Buffer) AddUndoAction(a EditorAction) {
	a.Selections = e.Cursors
	a.Data = bytes.Clone(a.Data)
	e.UndoStack.Push(a)
}
func (e *Buffer) PopAndReverseLastAction() {
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

func (e *Buffer) SetStateDirty() {
	e.State = State_Dirty
	e.needParsing = true
	e.calculateVisualLines()
}

func (e *Buffer) SetStateClean() {
	e.State = State_Clean
	e.needParsing = true
}

func (e *Buffer) replaceTabsWithSpaces() {
	e.Content = bytes.Replace(e.Content, []byte("\t"), []byte(strings.Repeat(" ", e.fileType.TabSize)), -1)
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isWhitespace(b byte) bool {
	return b == '\n' || b == '\r' || b == ' '
}

func isLetterOrDigit(b byte) bool {
	return isLetter(b) || isDigit(b)
}

func isSymbol(c byte) bool {
	return (c >= 33 && c <= 47) || (c >= 58 && c <= 64) || (c >= 91 && c <= 96) || (c >= 123 && c <= 126)
}

type WordToken struct {
	Start int
	End   int
	Type  int
}

func (e *Buffer) generateWordTokens() []WordToken {
	const (
		WORD_LEXER_TOKEN_TYPE_WORD       = 1
		WORD_LEXER_TOKEN_TYPE_SYMBOL     = 2
		WORD_LEXER_TOKEN_TYPE_WHITESPACE = 3
	)

	const (
		wordLexer_insideWord        = 1
		wordLexer_insideWhitespaces = 2
	)

	state := 0
	point := 0

	var tokens []WordToken
	for i := 0; i < len(e.Content); i++ {
		c := e.Content[i]
		switch {
		case isLetterOrDigit(c):
			switch state {
			case wordLexer_insideWord, 0:
				state = wordLexer_insideWord
				continue
			case wordLexer_insideWhitespaces:
				tokens = append(tokens, WordToken{
					Start: point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WHITESPACE,
				})
				state = wordLexer_insideWord
				point = i

			}
		case isWhitespace(c):
			switch state {
			case wordLexer_insideWhitespaces, 0:
				state = wordLexer_insideWhitespaces
				continue
			case wordLexer_insideWord:
				tokens = append(tokens, WordToken{
					Start: point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WORD,
				})
				state = wordLexer_insideWhitespaces
				point = i
			}
		default:
			switch state {
			case wordLexer_insideWord:
				tokens = append(tokens, WordToken{
					Start: point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WORD,
				})
				point = i
				state = 0

			case wordLexer_insideWhitespaces:
				tokens = append(tokens, WordToken{
					Start: point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WHITESPACE,
				})
				point = i
				state = 0
			}

			tokens = append(tokens, WordToken{
				Start: point,
				End:   i + 1,
				Type:  WORD_LEXER_TOKEN_TYPE_SYMBOL,
			})
			point = i + 1
		}
	}
	var typ int
	if state == wordLexer_insideWord {
		typ = WORD_LEXER_TOKEN_TYPE_WORD
	} else if state == wordLexer_insideWhitespaces {
		typ = WORD_LEXER_TOKEN_TYPE_WHITESPACE
	} else {
		typ = WORD_LEXER_TOKEN_TYPE_SYMBOL
	}

	tokens = append(tokens, WordToken{
		Start: point,
		End:   len(e.Content) - 1,
		Type:  typ,
	})

	return tokens
}

func (e *Buffer) updateMaxLineAndColumn(maxH float64, maxW float64) {
	oldMaxLine := e.maxLine
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	e.maxColumn = int32(maxW / float64(charSize.X))
	e.maxLine = int32(maxH / float64(charSize.Y))

	// reserve one line forr status bar
	e.maxLine--

	diff := e.maxLine - oldMaxLine
	e.View.EndLine += diff
}

func (e *Buffer) getBufferLineForIndex(i int) BufferLine {
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
	return BufferLine{}
}

func (e *Buffer) getIndexPosition(i int) Position {
	if len(e.View.Lines) == 0 {
		return Position{Line: 0, Column: i}
	}

	line := e.getBufferLineForIndex(i)
	return Position{
		Line:   line.Index,
		Column: i - line.startIndex,
	}

}

func (e *Buffer) positionToBufferIndex(pos Position) int {
	if len(e.View.Lines) <= pos.Line {
		return len(e.Content)
	}

	return e.View.Lines[pos.Line].startIndex + pos.Column
}

func (e *Buffer) Destroy() error {
	return nil
}

type highlight struct {
	start int
	end   int
	Color color.RGBA
}

type BufferLine struct {
	Index      int
	startIndex int
	endIndex   int
	ActualLine int
	Length     int
}

func sortme[T any](slice []T, pred func(t1 T, t2 T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return pred(slice[i], slice[j])
	})
}
func (e *Buffer) moveCursorTo(pos rl.Vector2) error {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64((pos.Y - e.zeroLocation.Y) / charSize.Y))
	apprColumn := math.Floor(float64((pos.X - e.zeroLocation.X) / charSize.X))

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
func (e *Buffer) calculateVisualLines() {
	e.Tokens = e.generateWordTokens()
	e.View.Lines = []BufferLine{}
	totalVisualLines := 0
	lineCharCounter := 0
	var actualLineIndex = 1
	var start int
	if e.View.EndLine == 0 {
		e.View.EndLine = e.maxLine
	}

	for idx, char := range e.Content {
		lineCharCounter++
		if char == '\n' {
			line := BufferLine{
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
			line := BufferLine{
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
			line := BufferLine{
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

func (e *Buffer) renderCursors(zeroLocation rl.Vector2, maxH float64, maxW float64) {

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
			posY := int32(cursorView.Line)*int32(charSize.Y) + int32(zeroLocation.Y)

			if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
				continue
			}
			switch e.cfg.CursorShape {
			case CURSOR_SHAPE_OUTLINE:
				rl.DrawRectangleLines(posX, posY, int32(charSize.X), int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
			case CURSOR_SHAPE_BLOCK:
				rl.DrawRectangle(posX, posY, int32(charSize.X), int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
				if len(e.Content)-1 >= sel.Point {
					rl.DrawTextEx(e.parent.Font, string(e.Content[sel.Point]), rl.Vector2{X: float32(posX), Y: float32(posY)}, float32(e.parent.FontSize), 0, e.cfg.CurrentThemeColors().Background.ToColorRGBA())
				}

			case CURSOR_SHAPE_LINE:
				rl.DrawRectangleLines(posX, posY, 2, int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
			}
			if e.cfg.CursorLineHighlight {
				rl.DrawRectangle(int32(zeroLocation.X), int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), e.maxColumn*int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().CursorLineBackground.ToColorRGBA(), 0.15))
			}

		} else {
			e.highlightBetweenTwoIndexes(zeroLocation, sel.Start(), sel.End(), maxH, maxW, e.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), e.cfg.CurrentThemeColors().SelectionForeground.ToColorRGBA())
		}

	}

}

func (e *Buffer) renderStatusbar(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if e.NoStatusbar {
		return
	}
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
	//render status bar
	rl.DrawRectangle(
		int32(zeroLocation.X),
		int32(zeroLocation.Y),
		int32(maxW),
		int32(charSize.Y),
		e.cfg.CurrentThemeColors().StatusBarBackground.ToColorRGBA(),
	)
	rl.DrawRectangleLines(int32(zeroLocation.X),
		int32(zeroLocation.Y),
		int32(maxW),
		int32(charSize.Y), e.cfg.CurrentThemeColors().Foreground.ToColorRGBA())

	sections = append(sections, fmt.Sprintf("%s %s", state, file))
	rl.DrawTextEx(e.parent.Font,
		strings.Join(sections, " "),
		rl.Vector2{X: zeroLocation.X, Y: float32(zeroLocation.Y)},
		float32(e.parent.FontSize),
		0,
		e.cfg.CurrentThemeColors().StatusBarForeground.ToColorRGBA())

}

func (e *Buffer) renderTextRange(zeroLocation rl.Vector2, idx1 int, idx2 int, maxH float64, maxW float64, color color.RGBA) {
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

		if i > int(e.View.EndLine) || i < int(e.View.StartLine) {
			break
		}

		if i < end.Line {
			thisLineEnd = line.Length - 1
		} else {
			thisLineEnd = end.Column
		}
		posX := int32(thisLineStart)*int32(charSize.X) + int32(zeroLocation.X)
		if e.cfg.LineNumbers {
			if len(e.View.Lines) > i {
				posX += int32((len(fmt.Sprint(e.View.Lines[i].ActualLine)) + 1) * int(charSize.X))
			} else {
				posX += int32(charSize.X)

			}
		}
		posY := int32(i-int(e.View.StartLine))*int32(charSize.Y) + int32(zeroLocation.Y)
		rl.DrawTextEx(e.parent.Font,
			string(e.Content[thisLineStart+line.startIndex:thisLineEnd+line.startIndex]),
			rl.Vector2{
				X: float32(posX), Y: float32(posY),
			}, float32(e.parent.FontSize), 0, color)
	}
}

func (e *Buffer) highlightBetweenTwoIndexes(zeroLocation rl.Vector2, idx1 int, idx2 int, maxH float64, maxW float64, bg color.RGBA, fg color.RGBA) {
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
		posX := int32(thisLineStart)*int32(charSize.X) + int32(zeroLocation.X)
		posY := int32(i-int(e.View.StartLine))*int32(charSize.Y) + int32(zeroLocation.Y)
		if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
			continue
		}
		rl.DrawTextEx(e.parent.Font, string(e.Content[thisLineStart+line.startIndex:thisLineEnd+line.startIndex+1]),
			rl.Vector2{X: float32(posX), Y: float32(posY)}, float32(e.parent.FontSize), 0, fg)

		for j := thisLineStart; j <= thisLineEnd; j++ {
			posX := int32(j)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				if len(e.View.Lines) > i {
					posX += int32((len(fmt.Sprint(e.View.Lines[i].ActualLine)) + 1) * int(charSize.X))
				} else {
					posX += int32(charSize.X)

				}
			}
			posY := int32(i-int(e.View.StartLine))*int32(charSize.Y) + int32(zeroLocation.Y)
			if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
				continue
			}
			rl.DrawRectangle(posX, posY, int32(charSize.X), int32(charSize.Y), rl.Fade(bg, 0.5))
		}
	}

}

func (e *Buffer) renderText(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	var visibleLines []BufferLine
	if e.View.EndLine < 0 {
		return
	}
	if e.View.EndLine > int32(len(e.View.Lines)) {
		visibleLines = e.View.Lines[e.View.StartLine:]
	} else {
		visibleLines = e.View.Lines[e.View.StartLine:e.View.EndLine]
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	if e.needParsing {
		var err error
		e.highlights, e.oldTSTree, err = TSHighlights(e.cfg, e.fileType.TSHighlightQuery, nil, e.Content) //TODO: see how we can use old tree
		if err != nil {
			panic(err)
		}
		e.needParsing = false
	}
	for idx, line := range visibleLines {
		if e.shouldBufferLineBeRendered(line) {
			if e.cfg.LineNumbers {
				rl.DrawTextEx(e.parent.Font,
					fmt.Sprintf("%d", line.ActualLine),
					rl.Vector2{X: zeroLocation.X, Y: zeroLocation.Y + float32(idx)*charSize.Y},
					float32(e.parent.FontSize),
					0,
					e.cfg.CurrentThemeColors().LineNumbersForeground.ToColorRGBA())
			}
			e.renderTextRange(zeroLocation, line.startIndex, line.endIndex, maxH, maxW, e.cfg.CurrentThemeColors().Foreground.ToColorRGBA())
		}
	}

	if e.cfg.EnableSyntaxHighlighting {
		for _, h := range e.highlights {
			e.renderTextRange(zeroLocation, h.start, h.end, maxH, maxW, h.Color)
		}
	}
}
func (e *Buffer) convertBufferIndexToLineAndColumn(idx int) *Position {
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
func (e *Buffer) findMatches(pattern string) error {
	e.ISearch.SearchMatches = [][]int{}
	matchPatternAsync(&e.ISearch.SearchMatches, e.Content, []byte(pattern))
	return nil
}

func (e *Buffer) findMatchesAndHighlight(pattern string, zeroLocation rl.Vector2, maxH float64, maxW float64) error {
	if pattern != e.ISearch.LastSearchString && pattern != "" {
		if err := e.findMatches(pattern); err != nil {
			return err
		}
	}
	for idx, match := range e.ISearch.SearchMatches {
		if idx == e.ISearch.CurrentMatch {
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
		e.highlightBetweenTwoIndexes(zeroLocation, match[0], match[1], maxH, maxW, e.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), e.cfg.CurrentThemeColors().SelectionForeground.ToColorRGBA())
	}
	e.ISearch.LastSearchString = pattern

	return nil
}
func (e *Buffer) renderSearch(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if !e.ISearch.IsSearching {
		return
	}
	e.findMatchesAndHighlight(e.ISearch.SearchString, zeroLocation, maxH, maxW)
	if len(e.ISearch.SearchMatches) > 0 {
		e.Cursors = e.Cursors[:1]
		e.Cursors[0].Point = e.ISearch.SearchMatches[e.ISearch.CurrentMatch][0]
		e.Cursors[0].Mark = e.ISearch.SearchMatches[e.ISearch.CurrentMatch][0]
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts.ToColorRGBA())
	rl.DrawTextEx(e.parent.Font, fmt.Sprintf("ISearch: %s", e.ISearch.SearchString), rl.Vector2{
		X: zeroLocation.X,
		Y: zeroLocation.Y,
	}, float32(e.parent.FontSize), 0, rl.White)
}

func (e *Buffer) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	e.zeroLocation = zeroLocation
	e.updateMaxLineAndColumn(maxH, maxW)
	e.calculateVisualLines()
	e.renderStatusbar(zeroLocation, maxH, maxW)
	zeroLocation.Y += measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0).Y
	e.renderText(zeroLocation, maxH, maxW)
	e.renderSearch(zeroLocation, maxH, maxW)
	e.renderCursors(zeroLocation, maxH, maxW)
}

func (e *Buffer) shouldBufferLineBeRendered(line BufferLine) bool {
	if e.View.StartLine <= int32(line.Index) && line.Index <= int(e.View.EndLine) {
		return true
	}

	return false
}

func (e *Buffer) isValidCursorPosition(newPosition Position) bool {
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

func (e *Buffer) deleteSelectionsIfAnySelection() {
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
		e.moveLeft(sel, old-1-len(e.Content))
	}

}

func (e *Buffer) indexOfFirstNonLetter(bs []byte) int {

	for idx, b := range bs {
		if !unicode.IsLetter(rune(b)) {
			return idx
		}
	}

	return -1
}

func (e *Buffer) findIndexPositionInTokens(idx int) int {
	for i, t := range e.Tokens {
		if t.Start <= idx && idx < t.End {
			return i
		}
	}

	return -1
}

func (e *Buffer) sortSelections() {
	sortme(e.Cursors, func(t1 Cursor, t2 Cursor) bool {
		return t1.Start() < t2.Start()
	})
}

func (e *Buffer) getLastCursor() *Cursor {
	return &e.Cursors[len(e.Cursors)-1]
}

func (e *Buffer) removeDuplicateSelectionsAndSort() {
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

func (e *Buffer) moveRight(s *Cursor, n int) {
	e.Cursors[0].Point = e.Cursors[0].Mark
	s.AddToBoth(n)
	if s.Start() > len(e.Content) {
		s.SetBoth(len(e.Content))
	}
	e.ScrollIfNeeded()

}
func (e *Buffer) moveLeft(s *Cursor, n int) {
	s.AddToBoth(-n)
	if s.Start() < 0 {
		s.SetBoth(0)
	}

	e.ScrollIfNeeded()
}
func (e *Buffer) addAnotherCursorAt(pos rl.Vector2) error {
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
func (e *Buffer) AnotherSelectionOnMatch() error {
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
		tokenPos := e.findIndexPositionInTokens(lastSel.Point)
		if tokenPos != -1 {
			token := e.Tokens[tokenPos]
			thingToSearch = e.Content[token.Start:token.End]
			next := findNextMatch(e.Content, lastSel.End()+1, thingToSearch)
			if len(next) == 0 {
				return nil
			}
			e.Cursors = append(e.Cursors, Cursor{
				Point: next[0],
				Mark:  next[0],
			})
		}

	}

	return nil
}
func (e *Buffer) ScrollIfNeeded() error {
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

func (e *Buffer) readFileFromDisk() error {
	bs, err := os.ReadFile(e.File)
	if err != nil {
		return nil
	}

	e.Content = bs
	e.replaceTabsWithSpaces()
	e.SetStateClean()
	return nil
}
