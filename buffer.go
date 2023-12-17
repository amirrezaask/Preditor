package preditor

import (
	"bytes"
	"errors"
	"fmt"
	"golang.design/x/clipboard"
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
	t.ActionStack = NewStack[BufferAction](1000)
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

	ActionStack Stack[BufferAction]
}

const (
	BufferActionType_Insert = iota + 1
	BufferActionType_Delete
)

type BufferAction struct {
	Type int
	Idx  int
	Data []byte
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

func (e *Buffer) AddBufferAction(a BufferAction) {
	a.Data = bytes.Clone(a.Data)
	e.ActionStack.Push(a)
}
func (e *Buffer) RevertLastBufferAction() {
	last, err := e.ActionStack.Pop()
	if err != nil {
		if errors.Is(err, EmptyStack) {
			e.SetStateClean()
		}
		return
	}
	switch last.Type {
	case BufferActionType_Insert:
		e.RemoveRange(last.Idx, last.Idx+len(last.Data), false)
	case BufferActionType_Delete:
		e.Content = append(e.Content[:last.Idx], append(last.Data, e.Content[last.Idx:]...)...)
	}
	e.SetStateDirty()
}

func (e *Buffer) SetStateDirty() {
	e.State = State_Dirty
	e.needParsing = true
	e.generateBufferLines()
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
	if len(e.Cursors) > 1 {
		RemoveAllCursorsButOne(e)
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64((pos.Y - e.zeroLocation.Y) / charSize.Y))
	apprColumn := math.Floor(float64((pos.X - e.zeroLocation.X) / charSize.X))

	if e.cfg.LineNumbers {
		apprColumn -= float64(e.getLineNumbersMaxLength())
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
func (e *Buffer) generateBufferLines() {
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
		} else if idx == len(e.Content)-1 {
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
		} else if int32(lineCharCounter) > e.maxColumn-5 {
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
		}
	}
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
			posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
		}
		posY := int32(i-int(e.View.StartLine))*int32(charSize.Y) + int32(zeroLocation.Y)
		rl.DrawTextEx(e.parent.Font,
			string(e.Content[thisLineStart+line.startIndex:thisLineEnd+line.startIndex]),
			rl.Vector2{
				X: float32(posX), Y: float32(posY),
			}, float32(e.parent.FontSize), 0, color)
	}
}

func (e *Buffer) AddBytesAtIndex(data []byte, idx int, addBufferAction bool) {
	if idx >= len(e.Content) {
		e.Content = append(e.Content, data...)
	} else {
		e.Content = append(e.Content[:idx], append(data, e.Content[idx:]...)...)
	}
	if addBufferAction {
		e.AddBufferAction(BufferAction{
			Type: BufferActionType_Insert,
			Idx:  idx,
			Data: data,
		})
	}
}

func (e *Buffer) RemoveRange(start int, end int, addBufferAction bool) {
	if start < 0 {
		start = 0
	}
	if end >= len(e.Content) {
		end = len(e.Content)
	}
	rangeData := e.Content[start:end]
	if len(e.Content) <= end {
		e.Content = e.Content[:start]
	} else {
		e.Content = append(e.Content[:start], e.Content[end:]...)
	}
	if addBufferAction {

		e.AddBufferAction(BufferAction{
			Type: BufferActionType_Delete,
			Idx:  start,
			Data: rangeData,
		})
	}

}

func (e *Buffer) getLineNumbersMaxLength() int {
	return len(fmt.Sprint(len(e.View.Lines))) + 1
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
		if e.cfg.LineNumbers {
			posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
		}

		posY := int32(i-int(e.View.StartLine))*int32(charSize.Y) + int32(zeroLocation.Y)
		if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
			continue
		}
		rl.DrawTextEx(e.parent.Font, string(e.Content[thisLineStart+line.startIndex:thisLineEnd+line.startIndex+1]),
			rl.Vector2{X: float32(posX), Y: float32(posY)}, float32(e.parent.FontSize), 0, fg)

		for j := thisLineStart; j <= thisLineEnd; j++ {
			posX := int32(j)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
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
func (e *Buffer) findMatches(pattern string) {
	e.ISearch.SearchMatches = [][]int{}
	matchPatternAsync(&e.ISearch.SearchMatches, e.Content, []byte(pattern))
}

func (e *Buffer) findMatchesAndHighlight(pattern string, zeroLocation rl.Vector2, maxH float64, maxW float64) error {

	return nil
}

func (e *Buffer) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	e.zeroLocation = zeroLocation
	oldMaxLine := e.maxLine
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	e.maxColumn = int32(maxW / float64(charSize.X))
	e.maxLine = int32(maxH / float64(charSize.Y))
	e.maxLine-- //reserve one line of screen for statusbar
	diff := e.maxLine - oldMaxLine
	e.View.EndLine += diff
	e.generateBufferLines()

	if !e.NoStatusbar {
		var sections []string

		file := e.File

		var state string
		if e.State == State_Dirty {
			state = "U"
		} else {
			state = ""
		}
		sections = append(sections, fmt.Sprintf("%s %s", state, file))

		if len(e.Cursors) > 1 {
			sections = append(sections, fmt.Sprintf("%d#Cursors", len(e.Cursors)))
		} else {
			if e.Cursors[0].Start() == e.Cursors[0].End() {
				selStart := e.getIndexPosition(e.Cursors[0].Start())
				if len(e.View.Lines) > selStart.Line {
					selLine := e.View.Lines[selStart.Line]
					sections = append(sections, fmt.Sprintf("L#%d C#%d", selLine.ActualLine, selStart.Column))
				} else {
					sections = append(sections, fmt.Sprintf("L#%d C#%d", selStart.Line, selStart.Column))
				}

			} else {
				selEnd := e.getIndexPosition(e.Cursors[0].End())
				sections = append(sections, fmt.Sprintf("L#%d C#%d (Selected %d)", selEnd.Line, selEnd.Column, int(math.Abs(float64(e.Cursors[0].Start()-e.Cursors[0].End())))))
			}

		}

		bg := e.cfg.CurrentThemeColors().StatusBarBackground.ToColorRGBA()
		fg := e.cfg.CurrentThemeColors().StatusBarForeground.ToColorRGBA()
		if win := e.parent.ActiveWindow(); win != nil && win.DrawableID == e.ID {
			bg = e.cfg.CurrentThemeColors().ActiveStatusBarBackground.ToColorRGBA()
			fg = e.cfg.CurrentThemeColors().ActiveStatusBarForeground.ToColorRGBA()
		}
		rl.DrawRectangle(
			int32(zeroLocation.X),
			int32(zeroLocation.Y),
			int32(maxW),
			int32(charSize.Y),
			bg,
		)
		rl.DrawTextEx(e.parent.Font,
			strings.Join(sections, " "),
			rl.Vector2{X: zeroLocation.X, Y: float32(zeroLocation.Y)},
			float32(e.parent.FontSize),
			0,
			fg)
		zeroLocation.Y += measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0).Y

	}

	var visibleLines []BufferLine
	if e.View.EndLine < 0 {
		return
	}
	if e.View.EndLine > int32(len(e.View.Lines)) {
		visibleLines = e.View.Lines[e.View.StartLine:]
	} else {
		visibleLines = e.View.Lines[e.View.StartLine:e.View.EndLine]
	}
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

	if e.ISearch.IsSearching {
		if e.ISearch.SearchString != e.ISearch.LastSearchString && e.ISearch.SearchString != "" {
			e.findMatches(e.ISearch.SearchString)
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
		e.ISearch.LastSearchString = e.ISearch.SearchString
		if len(e.ISearch.SearchMatches) > 0 {
			e.Cursors = e.Cursors[:1]
			e.Cursors[0].Point = e.ISearch.SearchMatches[e.ISearch.CurrentMatch][0]
			e.Cursors[0].Mark = e.ISearch.SearchMatches[e.ISearch.CurrentMatch][0]
		}

		rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts.ToColorRGBA())
		rl.DrawTextEx(e.parent.Font, fmt.Sprintf("ISearch: %s", e.ISearch.SearchString), rl.Vector2{
			X: zeroLocation.X,
			Y: zeroLocation.Y,
		}, float32(e.parent.FontSize), 0, rl.White)
	}

	for _, sel := range e.Cursors {
		if sel.Start() == sel.End() {
			cursor := e.getIndexPosition(sel.Start())
			cursorView := Position{
				Line:   cursor.Line - int(e.View.StartLine),
				Column: cursor.Column,
			}
			posX := int32(cursorView.Column)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
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
		e.AddBufferAction(BufferAction{
			Type: BufferActionType_Delete,
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

func (e *Buffer) findClosestLeftTokenToIndex(idx int) int {
	var closestTokenIndex int
	for i, t := range e.Tokens {
		if t.Start <= idx && t.End <= idx {
			if t.Start > e.Tokens[closestTokenIndex].Start {
				closestTokenIndex = i
			}
		}
	}

	return closestTokenIndex
}

func (e *Buffer) sortCursors() {
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

	e.sortCursors()
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
		apprColumn -= float64(e.getLineNumbersMaxLength())
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

// Things that change buffer content

func InsertChar(e *Buffer, char byte) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], i*1)
		e.AddBytesAtIndex([]byte{char}, e.Cursors[i].Point, true)
		e.moveRight(&e.Cursors[i], 1)
	}
	e.SetStateDirty()
	e.ScrollIfNeeded()
	return nil
}

func DeleteCharBackward(e *Buffer) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Cursors {
		e.moveLeft(&e.Cursors[i], i*1)
		e.RemoveRange(e.Cursors[i].Point-1, e.Cursors[i].Point, true)
		e.moveLeft(&e.Cursors[i], 1)
	}
	e.SetStateDirty()
	return nil
}

func DeleteCharForward(e *Buffer) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Cursors {
		e.moveLeft(&e.Cursors[i], i*1)
		e.RemoveRange(e.Cursors[i].Point, e.Cursors[i].Point+1, true)
	}

	e.SetStateDirty()
	return nil
}

func DeleteWordBackward(e *Buffer) {
	if e.Readonly || len(e.Cursors) > 1 {
		return
	}
	e.deleteSelectionsIfAnySelection()

	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findClosestLeftTokenToIndex(cur.Point)
		if tokenPos == -1 {
			continue
		}
		start := e.Tokens[tokenPos].Start
		if start == cur.Point && tokenPos-1 >= 0 {
			start = e.Tokens[tokenPos-1].Start
		}
		old := len(e.Content)
		e.RemoveRange(start, cur.Point, true)
		cur.SetBoth(cur.Point + (len(e.Content) - old))
	}

	e.SetStateDirty()
}

func Indent(e *Buffer) error {
	e.removeDuplicateSelectionsAndSort()
	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], i*e.fileType.TabSize)
		e.AddBytesAtIndex([]byte(strings.Repeat(" ", e.fileType.TabSize)), e.Cursors[i].Point, true)
		e.moveRight(&e.Cursors[i], e.fileType.TabSize)
	}
	e.SetStateDirty()

	return nil
}

func KillLine(e *Buffer) error {
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	var lastChange int
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		old := len(e.Content)
		e.moveLeft(cur, lastChange)
		line := e.getBufferLineForIndex(cur.Start())
		e.RemoveRange(cur.Point, line.endIndex, true)
		lastChange += -1 * (len(e.Content) - old)
	}
	e.SetStateDirty()

	return nil
}
func Cut(e *Buffer) error {
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	cur := &e.Cursors[0]
	if cur.Start() != cur.End() {
		// Copy selection
		WriteToClipboard(e.Content[cur.Start():cur.End()])
		e.RemoveRange(cur.Start(), cur.End(), true)
		cur.Mark = cur.Point
	} else {
		line := e.getBufferLineForIndex(cur.Start())
		WriteToClipboard(e.Content[line.startIndex : line.endIndex+1])
		e.RemoveRange(line.startIndex, line.endIndex, true)
	}
	e.SetStateDirty()

	return nil
}

func Paste(e *Buffer) error {
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	e.deleteSelectionsIfAnySelection()
	contentToPaste := GetClipboardContent()
	cur := e.Cursors[0]
	e.AddBytesAtIndex(contentToPaste, cur.Start(), true)
	e.SetStateDirty()
	PointRight(e, len(contentToPaste))
	return nil
}

func InteractiveGotoLine(e *Buffer) error {
	doneHook := func(userInput string, c *Context) error {
		number, err := strconv.Atoi(userInput)
		if err != nil {
			return nil
		}

		for _, line := range e.View.Lines {
			if line.ActualLine == number {
				e.Cursors[0].SetBoth(line.startIndex)
				e.ScrollIfNeeded()
			}
		}

		return nil
	}
	e.parent.SetPrompt("Goto", nil, doneHook, nil, "")

	return nil
}

// @Scroll

func ScrollUp(e *Buffer, n int) error {
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

func ScrollToTop(e *Buffer) error {
	e.View.StartLine = 0
	e.View.EndLine = e.maxLine
	e.Cursors[0].SetBoth(0)

	return nil
}

func ScrollToBottom(e *Buffer) error {
	e.View.StartLine = int32(len(e.View.Lines) - 1 - int(e.maxLine))
	e.View.EndLine = int32(len(e.View.Lines) - 1)
	e.Cursors[0].SetBoth(e.View.Lines[len(e.View.Lines)-1].startIndex)

	return nil
}

func ScrollDown(e *Buffer, n int) error {
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

// @Point

func PointLeft(e *Buffer, n int) error {
	for i := range e.Cursors {
		e.moveLeft(&e.Cursors[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func PointRight(e *Buffer, n int) error {
	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func PointUp(e *Buffer) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Point)
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

func PointDown(e *Buffer) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Point)
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

func CentralizePoint(e *Buffer) error {
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

func PointToBeginningOfLine(e *Buffer) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].Start())
		e.Cursors[i].SetBoth(line.startIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func PointToEndOfLine(e *Buffer) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].Start())
		e.Cursors[i].SetBoth(line.endIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

// @Mark

func MarkRight(e *Buffer, n int) error {
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		sel.Mark += n
		if sel.Mark >= len(e.Content) {
			sel.Mark = len(e.Content)
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func MarkLeft(e *Buffer, n int) error {
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		sel.Mark -= n
		if sel.Mark < 0 {
			sel.Mark = 0
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func MarkUp(e *Buffer, n int) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Mark)
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

func MarkDown(e *Buffer, n int) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Mark)
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

func MarkPreviousWord(e *Buffer) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findIndexPositionInTokens(cur.Mark)
		if tokenPos != -1 && tokenPos-1 >= 0 {
			e.Cursors[i].Mark = e.Tokens[tokenPos-1].Start
		}
		e.ScrollIfNeeded()
	}

	return nil
}

func MarkNextWord(e *Buffer) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findIndexPositionInTokens(cur.Mark)
		if tokenPos != -1 && tokenPos != len(e.Tokens)-1 {
			e.Cursors[i].Mark = e.Tokens[tokenPos+1].Start
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func MarkToEndOfLine(e *Buffer) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].End())
		e.Cursors[i].Mark = line.endIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func MarkToBeginningOfLine(e *Buffer) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].End())
		e.Cursors[i].Mark = line.startIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

// @Cursors

func RemoveAllCursorsButOne(p *Buffer) error {
	p.Cursors = p.Cursors[:1]
	p.Cursors[0].Point = p.Cursors[0].Mark

	return nil
}

func AddCursorNextLine(e *Buffer) error {
	pos := e.getIndexPosition(e.Cursors[len(e.Cursors)-1].Start())
	pos.Line++
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Cursors = append(e.Cursors, Cursor{Point: newidx, Mark: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func AddCursorPreviousLine(e *Buffer) error {
	pos := e.getIndexPosition(e.Cursors[len(e.Cursors)-1].Start())
	pos.Line--
	if e.isValidCursorPosition(pos) {
		newidx := e.positionToBufferIndex(pos)
		e.Cursors = append(e.Cursors, Cursor{Point: newidx, Mark: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PointForwardWord(e *Buffer, n int) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		cur.SetBoth(cur.Point)
		tokenPos := e.findIndexPositionInTokens(cur.Mark)
		if tokenPos != -1 && tokenPos != len(e.Tokens)-1 {
			cur.SetBoth(e.Tokens[tokenPos+1].Start)
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func PointBackwardWord(e *Buffer, n int) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		cur.SetBoth(cur.Point)
		tokenPos := e.findIndexPositionInTokens(cur.Point)
		if tokenPos != -1 && tokenPos != 0 {
			cur.SetBoth(e.Tokens[tokenPos-1].Start)
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func Write(e *Buffer) error {
	if e.Readonly && e.IsSpecial() {
		return nil
	}

	if e.fileType.TabSize != 0 {
		e.Content = bytes.Replace(e.Content, []byte(strings.Repeat(" ", e.fileType.TabSize)), []byte("\t"), -1)
	}

	if e.fileType.BeforeSave != nil {
		_ = e.fileType.BeforeSave(e)
	}

	if err := os.WriteFile(e.File, e.Content, 0644); err != nil {
		return err
	}
	e.SetStateClean()
	e.replaceTabsWithSpaces()
	e.generateBufferLines()
	if e.fileType.AfterSave != nil {
		_ = e.fileType.AfterSave(e)

	}

	return nil
}

func Copy(e *Buffer) error {
	if len(e.Cursors) > 1 {
		return nil
	}
	cur := e.Cursors[0]
	if cur.Start() != cur.End() {
		// Copy selection
		end := cur.End() + 1
		if end >= len(e.Content) {
			end = len(e.Content) - 1
		}
		WriteToClipboard(e.Content[cur.Start():end])
	} else {
		line := e.getBufferLineForIndex(cur.Start())
		WriteToClipboard(e.Content[line.startIndex : line.endIndex+1])
	}

	return nil
}

func CompileAskForCommand(a *Buffer) error {
	a.parent.SetPrompt("Compile", nil, func(userInput string, c *Context) error {
		a.LastCompileCommand = userInput
		if err := a.parent.OpenCompilationBufferInBuildWindow(userInput); err != nil {
			return err
		}

		return nil
	}, nil, a.LastCompileCommand)

	return nil
}

func CompileNoAsk(a *Buffer) error {
	if a.LastCompileCommand == "" {
		return CompileAskForCommand(a)
	}

	if err := a.parent.OpenCompilationBufferInBuildWindow(a.LastCompileCommand); err != nil {
		return err
	}

	return nil
}

func GrepAsk(a *Buffer) error {
	a.parent.SetPrompt("Grep", nil, func(userInput string, c *Context) error {
		if err := a.parent.OpenGrepBufferInSensibleSplit(fmt.Sprintf("rg --vimgrep %s", userInput)); err != nil {
			return err
		}

		return nil
	}, nil, "")

	return nil
}

func ISearchDeleteBackward(e *Buffer) error {
	if e.ISearch.SearchString == "" {
		return nil
	}
	s := []byte(e.ISearch.SearchString)
	if len(s) < 1 {
		return nil
	}
	s = s[:len(s)-1]

	e.ISearch.SearchString = string(s)

	return nil
}

func ISearchActivate(e *Buffer) error {
	e.ISearch.IsSearching = true
	e.ISearch.SearchString = ""
	e.keymaps = append(e.keymaps, SearchTextBufferKeymap, MakeInsertionKeys(func(c *Context, b byte) error {
		e.ISearch.SearchString += string(b)
		return nil
	}))
	return nil
}

func ISearchExit(editor *Buffer) error {
	editor.keymaps = editor.keymaps[:len(editor.keymaps)-2]
	editor.ISearch.IsSearching = false
	editor.ISearch.SearchMatches = nil
	editor.ISearch.CurrentMatch = 0
	editor.ISearch.MovedAwayFromCurrentMatch = false
	return nil
}

func ISearchNextMatch(editor *Buffer) error {
	editor.ISearch.CurrentMatch++
	if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
		editor.ISearch.CurrentMatch = 0
	}
	editor.ISearch.MovedAwayFromCurrentMatch = false
	return nil
}

func ISearchPreviousMatch(editor *Buffer) error {
	editor.ISearch.CurrentMatch--
	if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
		editor.ISearch.CurrentMatch = 0
	}
	if editor.ISearch.CurrentMatch < 0 {
		editor.ISearch.CurrentMatch = len(editor.ISearch.SearchMatches) - 1
	}
	editor.ISearch.MovedAwayFromCurrentMatch = false
	return nil

}

func GetClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func WriteToClipboard(bs []byte) {
	clipboard.Write(clipboard.FmtText, bytes.Clone(bs))
}
