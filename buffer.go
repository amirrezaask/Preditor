package preditor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/smacker/go-tree-sitter/golang"
	"go/format"
	"image/color"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.design/x/clipboard"

	"github.com/amirrezaask/preditor/byteutils"
	sitter "github.com/smacker/go-tree-sitter"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var EditorKeymap Keymap
var SearchTextBufferKeymap Keymap

func BufferOpenLocationInCurrentLine(c *Context) error {
	b, ok := c.ActiveDrawable().(*BufferView)
	if !ok {
		return nil
	}

	line := BufferGetCurrentLine(b)
	if line == nil || len(line) < 1 {
		return nil
	}

	segs := bytes.SplitN(line, []byte(":"), 4)
	if len(segs) < 2 {
		return nil

	}

	var targetWindow *Window
	for _, col := range c.Windows {
		for _, win := range col {
			if c.ActiveWindowIndex != win.ID {
				targetWindow = win
				break
			}
		}
	}

	filename := segs[0]
	var lineNum int
	var col int
	var err error
	switch len(segs) {
	case 3:
		//filename:line: text
		lineNum, err = strconv.Atoi(string(segs[1]))
		if err != nil {
		}
	case 4:
		//filename:line:col: text
		lineNum, err = strconv.Atoi(string(segs[1]))
		if err != nil {
		}
		col, err = strconv.Atoi(string(segs[2]))
		if err != nil {
		}

	}
	_ = SwitchOrOpenFileInWindow(c, c.Cfg, string(filename), &Position{Line: lineNum, Column: col}, targetWindow)

	c.ActiveWindowIndex = targetWindow.ID
	return nil
}

func RunCommandInBuffer(parent *Context, cfg *Config, bufferName string, command string) (*BufferView, error) {
	bufferView := NewBufferViewFromFilename(parent, cfg, bufferName)
	cwd := parent.getCWD()

	bufferView.Buffer.Readonly = true
	runCompileCommand := func() {
		bufferView.Buffer.Content = nil
		bufferView.Buffer.Content = append(bufferView.Buffer.Content, []byte(fmt.Sprintf("Command: %s\n", command))...)
		bufferView.Buffer.Content = append(bufferView.Buffer.Content, []byte(fmt.Sprintf("Dir: %s\n", cwd))...)
		go func() {
			segs := strings.Split(command, " ")
			var args []string
			bin := segs[0]
			if len(segs) > 1 {
				args = segs[1:]
			}
			cmd := exec.Command(bin, args...)
			cmd.Dir = cwd
			since := time.Now()
			output, err := cmd.CombinedOutput()
			if err != nil {
				bufferView.Buffer.Content = []byte(err.Error())
				bufferView.Buffer.Content = append(bufferView.Buffer.Content, '\n')
			}
			bufferView.Buffer.Content = append(bufferView.Buffer.Content, output...)
			bufferView.Buffer.Content = append(bufferView.Buffer.Content, []byte(fmt.Sprintf("Done in %s\n", time.Since(since)))...)

		}()

	}

	bufferView.keymaps[1].BindKey(Key{K: "g"}, MakeCommand(func(b *BufferView) error {
		runCompileCommand()

		return nil
	}))
	bufferView.keymaps[1].BindKey(Key{K: "<enter>"}, BufferOpenLocationInCurrentLine)

	runCompileCommand()
	return bufferView, nil
}

func NewGrepBuffer(parent *Context, cfg *Config, command string) (*BufferView, error) {
	return RunCommandInBuffer(parent, cfg, "*Grep", command)
}

func NewCompilationBuffer(parent *Context, cfg *Config, command string) (*BufferView, error) {
	return RunCommandInBuffer(parent, cfg, "*Compilation*", command)

}

func NewBufferViewFromFilename(parent *Context, cfg *Config, filename string) *BufferView {
	buffer := parent.GetBufferByFilename(filename)
	if buffer == nil {
		buffer = parent.OpenFileAsBuffer(filename)
	}

	return NewBufferView(parent, cfg, buffer)
}

type FileType struct {
	TabSize                  int
	BeforeSave               func(*BufferView) error
	AfterSave                func(*BufferView) error
	DefaultCompileCommand    string
	CommentLineBeginingChars []byte
	FindRootOfProject        func(currentFilePath string) (string, error)
	TSHighlightQuery         []byte
}

var FileTypes map[string]FileType

func init() {
	FileTypes = map[string]FileType{
		".go": {
			TabSize: 4,
			BeforeSave: func(e *BufferView) error {
				newBytes, err := format.Source(e.Buffer.Content)
				if err != nil {
					return err
				}

				e.Buffer.Content = newBytes
				return nil
			},
			TSHighlightQuery: []byte(`
[
  "break"
  "case"
  "chan"
  "const"
  "continue"
  "default"
  "defer"
  "else"
  "fallthrough"
  "for"
  "func"
  "go"
  "goto"
  "if"
  "import"
  "interface"
  "map"
  "package"
  "range"
  "return"
  "select"
  "struct"
  "switch"
  "type"
  "var"
] @keyword

(type_identifier) @type
(comment) @comment
[(interpreted_string_literal) (raw_string_literal)] @string
[(identifier)] @ident
(selector_expression operand: (_) @selector field: (_) @field)
(if_statement condition: (_) @if_condition)
`),
			DefaultCompileCommand: "go build -v ./...",
		},
	}
}

func NewBufferView(parent *Context, cfg *Config, buffer *Buffer) *BufferView {
	t := BufferView{cfg: cfg}
	t.parent = parent
	t.Buffer = buffer
	t.LastCompileCommand = buffer.fileType.DefaultCompileCommand
	t.keymaps = append([]Keymap{}, EditorKeymap, MakeInsertionKeys(func(c *Context, b byte) error {
		return BufferInsertChar(&t, b)
	}))
	t.ActionStack = NewStack[BufferAction](1000)
	t.Cursors = append(t.Cursors, Cursor{Point: 0, Mark: 0})
	t.replaceTabsWithSpaces()
	return &t
}

const (
	State_Clean = 0
	State_Dirty = 1
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

type ISearch struct {
	IsSearching               bool
	LastSearchString          string
	SearchString              string
	SearchMatches             [][]int
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool
}

type Buffer struct {
	File     string
	Content  []byte
	CRLF     bool
	State    int
	Readonly bool

	tokens      []WordToken
	oldTSTree   *sitter.Tree
	highlights  []highlight
	needParsing bool
	fileType    FileType
}

type BufferView struct {
	BaseDrawable
	Buffer                     *Buffer
	cfg                        *Config
	parent                     *Context
	maxLine                    int32
	maxColumn                  int32
	NoStatusbar                bool
	zeroLocation               rl.Vector2
	bufferLines                []BufferLine
	VisibleStart               int32
	MoveToPositionInNextRender *Position

	keymaps []Keymap

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

func (e *BufferView) String() string {
	return e.Buffer.File
}

func (e *BufferView) Keymaps() []Keymap {
	return e.keymaps
}

func (e *BufferView) IsSpecial() bool {
	return e.Buffer.File == "" || e.Buffer.File[0] == '*'
}

func (e *BufferView) AddBufferAction(a BufferAction) {
	a.Data = bytes.Clone(a.Data)
	e.ActionStack.Push(a)
}
func (e *BufferView) RevertLastBufferAction() {
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
		e.Buffer.Content = append(e.Buffer.Content[:last.Idx], append(last.Data, e.Buffer.Content[last.Idx:]...)...)
	}
	e.SetStateDirty()
}

func (e *BufferView) SetStateDirty() {
	e.Buffer.State = State_Dirty
	e.Buffer.needParsing = true
}

func (e *BufferView) SetStateClean() {
	e.Buffer.State = State_Clean
	e.Buffer.needParsing = true
}

func (e *BufferView) replaceTabsWithSpaces() {
	e.Buffer.Content = bytes.Replace(e.Buffer.Content, []byte("\t"), []byte(strings.Repeat(" ", e.Buffer.fileType.TabSize)), -1)
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

func (e *BufferView) generateWordTokens() []WordToken {
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
	for i := 0; i < len(e.Buffer.Content); i++ {
		c := e.Buffer.Content[i]
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
		End:   len(e.Buffer.Content) - 1,
		Type:  typ,
	})

	return tokens
}

func (e *BufferView) getBufferLineForIndex(i int) BufferLine {
	for _, line := range e.bufferLines {
		if line.startIndex <= i && line.endIndex >= i {
			return line
		}
	}

	if len(e.bufferLines) > 0 {
		lastLine := e.bufferLines[len(e.bufferLines)-1]
		lastLine.endIndex++
		return lastLine
	}
	return BufferLine{}
}

func (e *BufferView) BufferIndexToPosition(i int) Position {
	if len(e.bufferLines) == 0 {
		return Position{Line: 0, Column: i}
	}

	line := e.getBufferLineForIndex(i)
	return Position{
		Line:   line.Index,
		Column: i - line.startIndex,
	}

}

func (e *BufferView) PositionToBufferIndex(pos Position) int {
	if len(e.bufferLines) <= pos.Line {
		return len(e.Buffer.Content)
	}

	return e.bufferLines[pos.Line].startIndex + pos.Column
}

func (e *BufferView) Destroy() error {
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
func (e *BufferView) moveCursorTo(pos rl.Vector2) error {
	if len(e.Cursors) > 1 {
		RemoveAllCursorsButOne(e)
	}
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64((pos.Y - e.zeroLocation.Y) / charSize.Y))
	apprColumn := math.Floor(float64((pos.X - e.zeroLocation.X) / charSize.X))

	if e.cfg.LineNumbers {
		apprColumn -= float64(e.getLineNumbersMaxLength())
	}

	if len(e.bufferLines) < 1 {
		return nil
	}

	line := int(apprLine) + int(e.VisibleStart) - 1
	col := int(apprColumn)
	if line >= len(e.bufferLines) {
		line = len(e.bufferLines) - 1
	}

	if line < 0 {
		line = 0
	}

	// check if cursor should be moved back
	if col > e.bufferLines[line].Length {
		col = e.bufferLines[line].Length
	}
	if col < 0 {
		col = 0
	}

	e.Cursors[0].SetBoth(e.PositionToBufferIndex(Position{Line: line, Column: col}))

	return nil
}

func (e *BufferView) VisibleEnd() int32 {
	return e.VisibleStart + e.maxLine
}

func (e *BufferView) renderTextRange(zeroLocation rl.Vector2, idx1 int, idx2 int, maxH float64, maxW float64, color color.RGBA) {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	var start Position
	var end Position
	if idx1 > idx2 {
		start = e.BufferIndexToPosition(idx2)
		end = e.BufferIndexToPosition(idx1)
	} else if idx2 > idx1 {
		start = e.BufferIndexToPosition(idx1)
		end = e.BufferIndexToPosition(idx2)
	}
	for i := start.Line; i <= end.Line; i++ {
		if len(e.bufferLines) <= i {
			break
		}
		var thisLineEnd int
		var thisLineStart int
		line := e.bufferLines[i]
		if i == start.Line {
			thisLineStart = start.Column
		} else {
			thisLineStart = 0
		}

		if i < int(e.VisibleStart) {
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
		posY := int32(i-int(e.VisibleStart))*int32(charSize.Y) + int32(zeroLocation.Y)
		rl.DrawTextEx(e.parent.Font,
			string(e.Buffer.Content[thisLineStart+line.startIndex:thisLineEnd+line.startIndex]),
			rl.Vector2{
				X: float32(posX), Y: float32(posY),
			}, float32(e.parent.FontSize), 0, color)
	}
}

func (e *BufferView) AddBytesAtIndex(data []byte, idx int, addBufferAction bool) {
	if idx >= len(e.Buffer.Content) {
		e.Buffer.Content = append(e.Buffer.Content, data...)
	} else {
		e.Buffer.Content = append(e.Buffer.Content[:idx], append(data, e.Buffer.Content[idx:]...)...)
	}
	if addBufferAction {
		e.AddBufferAction(BufferAction{
			Type: BufferActionType_Insert,
			Idx:  idx,
			Data: data,
		})
	}
}

func (e *BufferView) RemoveRange(start int, end int, addBufferAction bool) {
	if start < 0 {
		start = 0
	}
	if end >= len(e.Buffer.Content) {
		end = len(e.Buffer.Content)
	}
	rangeData := e.Buffer.Content[start:end]
	if len(e.Buffer.Content) <= end {
		e.Buffer.Content = e.Buffer.Content[:start]
	} else {
		e.Buffer.Content = append(e.Buffer.Content[:start], e.Buffer.Content[end:]...)
	}
	if addBufferAction {

		e.AddBufferAction(BufferAction{
			Type: BufferActionType_Delete,
			Idx:  start,
			Data: rangeData,
		})
	}

}
func (e *BufferView) WordAtPoint(curIndex int) (int, int) {
	return -1, -1
}

func (e *BufferView) WordLeftEnd(cursor int) int {
	return -1
}

func (e *BufferView) WordRightStart(cursor int) int {
	return -1
}

func (e *BufferView) getLineNumbersMaxLength() int {
	return len(fmt.Sprint(len(e.bufferLines))) + 1
}

func BufferGetCurrentLine(e *BufferView) []byte {
	cur := e.Cursors[0]
	pos := e.BufferIndexToPosition(cur.Point)

	if pos.Line < len(e.bufferLines) {
		line := e.bufferLines[pos.Line]
		return e.Buffer.Content[line.startIndex:line.endIndex]
	}

	return nil
}

func (e *BufferView) highlightBetweenTwoIndexes(zeroLocation rl.Vector2, idx1 int, idx2 int, maxH float64, maxW float64, bg color.RGBA, fg color.RGBA) {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	var start Position
	var end Position
	if idx1 > idx2 {
		start = e.BufferIndexToPosition(idx2)
		end = e.BufferIndexToPosition(idx1)
	} else if idx2 > idx1 {
		start = e.BufferIndexToPosition(idx1)
		end = e.BufferIndexToPosition(idx2)
	}
	for i := start.Line; i <= end.Line; i++ {
		if len(e.bufferLines) <= i {
			break
		}
		var thisLineEnd int
		var thisLineStart int
		line := e.bufferLines[i]
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

		posY := int32(i-int(e.VisibleStart))*int32(charSize.Y) + int32(zeroLocation.Y)
		if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
			continue
		}
		rl.DrawTextEx(e.parent.Font, string(e.Buffer.Content[thisLineStart+line.startIndex:thisLineEnd+line.startIndex+1]),
			rl.Vector2{X: float32(posX), Y: float32(posY)}, float32(e.parent.FontSize), 0, fg)

		for j := thisLineStart; j <= thisLineEnd; j++ {
			posX := int32(j)*int32(charSize.X) + int32(zeroLocation.X)
			if e.cfg.LineNumbers {
				posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
			}
			posY := int32(i-int(e.VisibleStart))*int32(charSize.Y) + int32(zeroLocation.Y)
			if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
				continue
			}
			rl.DrawRectangle(posX, posY, int32(charSize.X), int32(charSize.Y), rl.Fade(bg, 0.5))
		}
	}

}

func (e *BufferView) convertBufferIndexToLineAndColumn(idx int) *Position {
	for lineIndex, line := range e.bufferLines {
		if line.startIndex <= idx && line.endIndex >= idx {
			return &Position{
				Line:   lineIndex,
				Column: idx - line.startIndex,
			}
		}
	}

	return nil
}

func matchPatternCaseInsensitive(data []byte, pattern []byte) [][]int {
	var matched [][]int
	var buf []byte
	start := -1
	for i, b := range data {

		if len(pattern) == len(buf) {
			matched = append(matched, []int{start, i - 1})
			buf = nil
			start = -1
		}
		idxToCheck := len(buf)
		if idxToCheck == 0 {
			start = i
		}
		if unicode.ToLower(rune(pattern[idxToCheck])) == unicode.ToLower(rune(b)) {
			buf = append(buf, b)
		} else {
			buf = nil
			start = -1
		}
	}

	return matched
}

func findNextMatch(data []byte, idx int, pattern []byte) []int {
	var buf []byte
	start := -1
	for i := idx; i < len(data); i++ {
		if len(pattern) == len(buf) {
			return []int{start, i - 1}
		}
		idxToCheck := len(buf)
		if idxToCheck == 0 {
			start = i
		}
		if unicode.ToLower(rune(pattern[idxToCheck])) == unicode.ToLower(rune(data[i])) {
			buf = append(buf, data[i])
		} else {
			buf = nil
			start = -1
		}
	}

	return nil
}

func matchPatternAsync(dst *[][]int, data []byte, pattern []byte) {
	go func() {
		*dst = matchPatternCaseInsensitive(data, pattern)
	}()
}

func (e *BufferView) findMatches(pattern string) {
	e.ISearch.SearchMatches = [][]int{}
	matchPatternAsync(&e.ISearch.SearchMatches, e.Buffer.Content, []byte(pattern))
}
func TSHighlights(cfg *Config, queryString []byte, prev *sitter.Tree, code []byte) ([]highlight, *sitter.Tree, error) {
	var highlights []highlight
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), prev, code)
	if err != nil {
		return nil, nil, err
	}
	query, err := sitter.NewQuery(queryString, golang.GetLanguage())
	if err != nil {
		return nil, tree, err
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, tree.RootNode())
	for {
		qm, exists := qc.NextMatch()
		if !exists {
			break
		}
		for _, capture := range qm.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			if c, exists := cfg.CurrentThemeColors().SyntaxColors[captureName]; exists {
				highlights = append(highlights, highlight{
					start: int(capture.Node.StartByte()),
					end:   int(capture.Node.EndByte()),
					Color: c.ToColorRGBA(),
				})
			}
		}
	}

	return highlights, tree, nil
}

func (e *BufferView) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	e.maxColumn = int32(maxW / float64(charSize.X))
	e.maxLine = int32(maxH / float64(charSize.Y))
	e.maxLine-- //reserve one line of screen for statusbar

	// generate visual lines that we are going to render
	e.Buffer.tokens = e.generateWordTokens()
	e.bufferLines = []BufferLine{}
	totalVisualLines := 0
	lineCharCounter := 0
	var actualLineIndex = 1
	var start int
	for idx, char := range e.Buffer.Content {
		lineCharCounter++
		if char == '\n' {
			line := BufferLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start + 1,
				ActualLine: actualLineIndex,
			}
			e.bufferLines = append(e.bufferLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1
		}
		if idx == len(e.Buffer.Content)-1 {
			// last index
			line := BufferLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx + 1,
				Length:     idx - start + 1,
				ActualLine: actualLineIndex,
			}
			e.bufferLines = append(e.bufferLines, line)
			totalVisualLines++
			actualLineIndex++
			lineCharCounter = 0
			start = idx + 1
		}
		if int32(lineCharCounter) > e.maxColumn-5 {
			line := BufferLine{
				Index:      totalVisualLines,
				startIndex: start,
				endIndex:   idx,
				Length:     idx - start + 1,
				ActualLine: actualLineIndex,
			}
			e.bufferLines = append(e.bufferLines, line)
			totalVisualLines++
			lineCharCounter = 0
			start = idx + 1
		}
	}

	if !e.NoStatusbar {
		var sections []string

		file := e.Buffer.File

		var state string
		if e.Buffer.State == State_Dirty {
			state = "U"
		} else {
			state = ""
		}
		sections = append(sections, fmt.Sprintf("%s %s", state, file))

		if len(e.Cursors) > 1 {
			sections = append(sections, fmt.Sprintf("%d#Cursors", len(e.Cursors)))
		} else {
			if e.Cursors[0].Start() == e.Cursors[0].End() {
				selStart := e.BufferIndexToPosition(e.Cursors[0].Start())
				if len(e.bufferLines) > selStart.Line {
					selLine := e.bufferLines[selStart.Line]
					sections = append(sections, fmt.Sprintf("L#%d C#%d", selLine.ActualLine, selStart.Column))
				} else {
					sections = append(sections, fmt.Sprintf("L#%d C#%d", selStart.Line, selStart.Column))
				}

			} else {
				selEnd := e.BufferIndexToPosition(e.Cursors[0].End())
				sections = append(sections, fmt.Sprintf("L#%d C#%d (Selected %d)", selEnd.Line, selEnd.Column, int(math.Abs(float64(e.Cursors[0].Start()-e.Cursors[0].End())))))
			}

		}
		if e.ISearch.IsSearching {
			sections = append(sections, fmt.Sprintf("ISearch: Match#%d Of %d", e.ISearch.CurrentMatch+1, len(e.ISearch.SearchMatches)+1))
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
		rl.DrawRectangleLines(int32(zeroLocation.X),
			int32(zeroLocation.Y),
			int32(maxW),
			int32(charSize.Y), e.cfg.CurrentThemeColors().Foreground.ToColorRGBA())

		rl.DrawTextEx(e.parent.Font,
			strings.Join(sections, " "),
			rl.Vector2{X: zeroLocation.X, Y: float32(zeroLocation.Y)},
			float32(e.parent.FontSize),
			0,
			fg)
		zeroLocation.Y += measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0).Y

	}
	if e.MoveToPositionInNextRender != nil {
		bufferIndex := e.PositionToBufferIndex(*e.MoveToPositionInNextRender)
		e.Cursors[0].SetBoth(bufferIndex)
		e.VisibleStart = int32(e.MoveToPositionInNextRender.Line) - e.maxLine/2
		e.MoveToPositionInNextRender = nil
	}
	if e.Buffer.needParsing {
		var err error
		e.Buffer.highlights, e.Buffer.oldTSTree, err = TSHighlights(e.cfg, e.Buffer.fileType.TSHighlightQuery, nil, e.Buffer.Content) //TODO: see how we can use old tree
		if err != nil {
			panic(err)
		}
		e.Buffer.needParsing = false
	}

	var visibleLines []BufferLine
	if e.VisibleStart < 0 {
		e.VisibleStart = 0
	}

	if e.VisibleEnd() > int32(len(e.bufferLines)) {
		visibleLines = e.bufferLines[e.VisibleStart:]
	} else {
		visibleLines = e.bufferLines[e.VisibleStart:e.VisibleEnd()]
	}

	for idx, line := range visibleLines {
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

	if e.cfg.EnableSyntaxHighlighting {
		for _, h := range e.Buffer.highlights {
			e.renderTextRange(zeroLocation, h.start, h.end, maxH, maxW, h.Color)
		}
	}

	if e.ISearch.IsSearching {
		if e.ISearch.SearchString != e.ISearch.LastSearchString && e.ISearch.SearchString != "" {
			e.findMatches(e.ISearch.SearchString)
		}
		for idx, match := range e.ISearch.SearchMatches {
			if idx == e.ISearch.CurrentMatch {
				matchStartLine := e.BufferIndexToPosition(match[0])
				matchEndLine := e.BufferIndexToPosition(match[0])
				if !(e.VisibleStart < int32(matchStartLine.Line) && e.VisibleEnd() > int32(matchEndLine.Line)) && !e.ISearch.MovedAwayFromCurrentMatch {
					// current match is not in view
					// move the view
					e.VisibleStart = int32(matchStartLine.Line) - e.maxLine/2
					if e.VisibleStart < 0 {
						e.VisibleStart = int32(matchStartLine.Line)
					}

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

	// render cursors
	if e.parent.ActiveDrawableID() == e.ID {
		for _, sel := range e.Cursors {
			if sel.Start() == sel.End() {
				cursor := e.BufferIndexToPosition(sel.Start())
				cursorView := Position{
					Line:   cursor.Line - int(e.VisibleStart),
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
					if len(e.Buffer.Content)-1 >= sel.Point {
						rl.DrawTextEx(e.parent.Font, string(e.Buffer.Content[sel.Point]), rl.Vector2{X: float32(posX), Y: float32(posY)}, float32(e.parent.FontSize), 0, e.cfg.CurrentThemeColors().Background.ToColorRGBA())
					}

				case CURSOR_SHAPE_LINE:
					rl.DrawRectangleLines(posX, posY, 2, int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
				}
				if e.cfg.CursorLineHighlight {
					rl.DrawRectangle(int32(zeroLocation.X), int32(cursorView.Line)*int32(charSize.Y)+int32(zeroLocation.Y), e.maxColumn*int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().CursorLineBackground.ToColorRGBA(), 0.15))
				}

				// highlight matching char
				if e.cfg.HighlightMatchingParen {
					matchingIdx := byteutils.FindMatching(e.Buffer.Content, e.Cursors[0].Point)
					if matchingIdx != -1 {
						idxPosition := e.BufferIndexToPosition(matchingIdx)
						idxPositionView := Position{
							Line:   idxPosition.Line - int(e.VisibleStart),
							Column: idxPosition.Column,
						}
						posX := int32(idxPositionView.Column)*int32(charSize.X) + int32(zeroLocation.X)
						if e.cfg.LineNumbers {
							posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
						}
						posY := int32(idxPositionView.Line)*int32(charSize.Y) + int32(zeroLocation.Y)

						rl.DrawRectangle(posX, posY, int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().HighlightMatching.ToColorRGBA(), 0.4))
					}
				}

			} else {
				e.highlightBetweenTwoIndexes(zeroLocation, sel.Start(), sel.End(), maxH, maxW, e.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), e.cfg.CurrentThemeColors().SelectionForeground.ToColorRGBA())
			}
		}

	}

	e.zeroLocation = zeroLocation

}

func (e *BufferView) isValidCursorPosition(newPosition Position) bool {
	if newPosition.Line < 0 {
		return false
	}
	if len(e.bufferLines) == 0 && newPosition.Line == 0 && newPosition.Column >= 0 && int32(newPosition.Column) < e.maxColumn-int32(len(fmt.Sprint(newPosition.Line)))-1 {
		return true
	}
	if newPosition.Line >= len(e.bufferLines) && (len(e.bufferLines) != 0) {
		return false
	}

	if newPosition.Column < 0 {
		return false
	}
	if newPosition.Column > e.bufferLines[newPosition.Line].Length+1 {
		return false
	}

	return true
}

func (e *BufferView) deleteSelectionsIfAnySelection() {
	if e.Buffer.Readonly {
		return
	}
	old := len(e.Buffer.Content)
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		if sel.Start() == sel.End() {
			continue
		}
		e.AddBufferAction(BufferAction{
			Type: BufferActionType_Delete,
			Idx:  sel.Start(),
			Data: e.Buffer.Content[sel.Start() : sel.End()+1],
		})
		e.Buffer.Content = append(e.Buffer.Content[:sel.Start()], e.Buffer.Content[sel.End()+1:]...)
		sel.Point = sel.Mark
		e.moveLeft(sel, old-1-len(e.Buffer.Content))
	}

}

func (e *BufferView) findIndexPositionInTokens(idx int) int {
	for i, t := range e.Buffer.tokens {
		if t.Start <= idx && idx < t.End {
			return i
		}
	}

	return -1
}

func (e *BufferView) findClosestLeftTokenToIndex(idx int) int {
	var closestTokenIndex int
	for i, t := range e.Buffer.tokens {
		if t.Start <= idx && t.End <= idx {
			if t.Start > e.Buffer.tokens[closestTokenIndex].Start {
				closestTokenIndex = i
			}
		}
	}

	return closestTokenIndex
}

func (e *BufferView) sortCursors() {
	sortme(e.Cursors, func(t1 Cursor, t2 Cursor) bool {
		return t1.Start() < t2.Start()
	})
}

func (e *BufferView) removeDuplicateSelectionsAndSort() {
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

func (e *BufferView) moveRight(s *Cursor, n int) {
	e.Cursors[0].Point = e.Cursors[0].Mark
	s.AddToBoth(n)
	if s.Start() > len(e.Buffer.Content) {
		s.SetBoth(len(e.Buffer.Content))
	}
	e.ScrollIfNeeded()

}
func (e *BufferView) moveLeft(s *Cursor, n int) {
	s.AddToBoth(-n)
	if s.Start() < 0 {
		s.SetBoth(0)
	}

	e.ScrollIfNeeded()
}
func (e *BufferView) addAnotherCursorAt(pos rl.Vector2) error {
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64(pos.Y / charSize.Y))
	apprColumn := math.Floor(float64(pos.X / charSize.X))

	if e.cfg.LineNumbers {
		apprColumn -= float64(e.getLineNumbersMaxLength())
	}

	if len(e.bufferLines) < 1 {
		return nil
	}

	line := int(apprLine) + int(e.VisibleStart)
	col := int(apprColumn)

	if line >= len(e.bufferLines) {
		line = len(e.bufferLines) - 1
	}

	if line < 0 {
		line = 0
	}

	// check if cursor should be moved back
	if col > e.bufferLines[line].Length {
		col = e.bufferLines[line].Length
	}
	if col < 0 {
		col = 0
	}
	idx := e.PositionToBufferIndex(Position{Line: line, Column: col})
	e.Cursors = append(e.Cursors, Cursor{Point: idx, Mark: idx})

	e.removeDuplicateSelectionsAndSort()
	return nil
}
func (e *BufferView) AnotherSelectionOnMatch() error {
	lastSel := e.Cursors[len(e.Cursors)-1]
	var thingToSearch []byte
	if lastSel.Point != lastSel.Mark {
		thingToSearch = e.Buffer.Content[lastSel.Start():lastSel.End()]
		next := findNextMatch(e.Buffer.Content, lastSel.End()+1, thingToSearch)
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
			token := e.Buffer.tokens[tokenPos]
			thingToSearch = e.Buffer.Content[token.Start:token.End]
			next := findNextMatch(e.Buffer.Content, lastSel.End()+1, thingToSearch)
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
func (e *BufferView) ScrollIfNeeded() {
	pos := e.BufferIndexToPosition(e.Cursors[0].End())
	if int32(pos.Line) <= e.VisibleStart {
		e.VisibleStart = int32(pos.Line) - e.maxLine/3

	}

	if pos.Line > int(e.VisibleEnd()) {
		e.VisibleStart += e.maxLine / 3
	}

	if int(e.VisibleEnd()) >= len(e.bufferLines) {
		e.VisibleStart = int32(len(e.bufferLines)-1) - e.maxLine
	}

	if e.VisibleStart < 0 {
		e.VisibleStart = 0
	}
}

func (e *BufferView) readFileFromDisk() error {
	bs, err := os.ReadFile(e.Buffer.File)
	if err != nil {
		return nil
	}

	e.Buffer.Content = bs
	e.replaceTabsWithSpaces()
	e.SetStateClean()
	return nil
}

// Things that change buffer content

func BufferInsertChar(e *BufferView, char byte) error {
	if e.Buffer.Readonly {
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

func DeleteCharBackward(e *BufferView) error {
	if e.Buffer.Readonly {
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

func DeleteCharForward(e *BufferView) error {
	if e.Buffer.Readonly {
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

func DeleteWordBackward(e *BufferView) {
	if e.Buffer.Readonly || len(e.Cursors) > 1 {
		return
	}
	e.deleteSelectionsIfAnySelection()

	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findClosestLeftTokenToIndex(cur.Point)
		if tokenPos == -1 {
			continue
		}
		start := e.Buffer.tokens[tokenPos].Start
		if start == cur.Point && tokenPos-1 >= 0 {
			start = e.Buffer.tokens[tokenPos-1].Start
		}
		old := len(e.Buffer.Content)
		e.RemoveRange(start, cur.Point, true)
		cur.SetBoth(cur.Point + (len(e.Buffer.Content) - old))
	}

	e.SetStateDirty()
}

func Indent(e *BufferView) error {
	e.removeDuplicateSelectionsAndSort()
	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], i*e.Buffer.fileType.TabSize)
		e.AddBytesAtIndex([]byte(strings.Repeat(" ", e.Buffer.fileType.TabSize)), e.Cursors[i].Point, true)
		e.moveRight(&e.Cursors[i], e.Buffer.fileType.TabSize)
	}
	e.SetStateDirty()

	return nil
}

func KillLine(e *BufferView) error {
	if e.Buffer.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	var lastChange int
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		old := len(e.Buffer.Content)
		e.moveLeft(cur, lastChange)
		line := e.getBufferLineForIndex(cur.Start())
		e.RemoveRange(cur.Point, line.endIndex, true)
		lastChange += -1 * (len(e.Buffer.Content) - old)
	}
	e.SetStateDirty()

	return nil
}
func Cut(e *BufferView) error {
	if e.Buffer.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	cur := &e.Cursors[0]
	if cur.Start() != cur.End() {
		// Copy selection
		WriteToClipboard(e.Buffer.Content[cur.Start() : cur.End()+1])
		fmt.Printf("Cutting '%s'\n", string(e.Buffer.Content[cur.Start():cur.End()+1]))
		e.RemoveRange(cur.Start(), cur.End()+1, true)
		cur.Mark = cur.Point
	} else {
		line := e.getBufferLineForIndex(cur.Start())
		fmt.Printf("Cutting '%s'\n", string(e.Buffer.Content[line.startIndex:line.endIndex+1]))
		WriteToClipboard(e.Buffer.Content[line.startIndex : line.endIndex+1])
		e.RemoveRange(line.startIndex, line.endIndex+1, true)
	}
	e.SetStateDirty()

	return nil
}

func Paste(e *BufferView) error {
	if e.Buffer.Readonly || len(e.Cursors) > 1 {
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

func InteractiveGotoLine(e *BufferView) error {
	doneHook := func(userInput string, c *Context) error {
		number, err := strconv.Atoi(userInput)
		if err != nil {
			return nil
		}

		for _, line := range e.bufferLines {
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

func ScrollUp(e *BufferView, n int) error {
	if e.VisibleStart <= 0 {
		return nil
	}
	e.VisibleStart += int32(-1 * n)

	if e.VisibleStart < 0 {
		e.VisibleStart = 0
	}

	return nil

}

func ScrollToTop(e *BufferView) error {
	e.VisibleStart = 0
	e.Cursors[0].SetBoth(0)

	return nil
}

func ScrollToBottom(e *BufferView) error {
	e.VisibleStart = int32(len(e.bufferLines) - 1 - int(e.maxLine))
	e.Cursors[0].SetBoth(e.bufferLines[len(e.bufferLines)-1].startIndex)

	return nil
}

func ScrollDown(e *BufferView, n int) error {
	if int(e.VisibleEnd()) >= len(e.bufferLines) {
		return nil
	}
	e.VisibleStart += int32(n)
	if int(e.VisibleEnd()) >= len(e.bufferLines) {
		e.VisibleStart = int32(len(e.bufferLines)-1) - e.maxLine
	}

	return nil

}

// @Point

func PointLeft(e *BufferView, n int) error {
	for i := range e.Cursors {
		e.moveLeft(&e.Cursors[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func PointRight(e *BufferView, n int) error {
	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], n)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func PointUp(e *BufferView) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Point)
		prevLineIndex := currentLine.Index - 1
		if prevLineIndex < 0 {
			return nil
		}

		prevLine := e.bufferLines[prevLineIndex]
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

func PointDown(e *BufferView) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Point)
		nextLineIndex := currentLine.Index + 1
		if nextLineIndex >= len(e.bufferLines) {
			return nil
		}

		nextLine := e.bufferLines[nextLineIndex]
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

func CentralizePoint(e *BufferView) error {
	cur := e.Cursors[0]
	pos := e.convertBufferIndexToLineAndColumn(cur.Start())
	e.VisibleStart = int32(pos.Line) - (e.maxLine / 2)
	if e.VisibleStart < 0 {
		e.VisibleStart = 0
	}
	return nil
}

func PointToBeginningOfLine(e *BufferView) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].Start())
		e.Cursors[i].SetBoth(line.startIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil

}

func PointToEndOfLine(e *BufferView) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].Start())
		e.Cursors[i].SetBoth(line.endIndex)
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PointToMatchingChar(e *BufferView) error {
	for i := range e.Cursors {
		matching := byteutils.FindMatching(e.Buffer.Content, e.Cursors[i].Point)
		if matching != -1 {
			e.Cursors[i].SetBoth(matching)
		}
	}

	e.removeDuplicateSelectionsAndSort()
	return nil

}

// @Mark

func MarkRight(e *BufferView, n int) error {
	for i := range e.Cursors {
		sel := &e.Cursors[i]
		sel.Mark += n
		if sel.Mark >= len(e.Buffer.Content) {
			sel.Mark = len(e.Buffer.Content)
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func MarkLeft(e *BufferView, n int) error {
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

func MarkUp(e *BufferView, n int) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Mark)
		nextLineIndex := currentLine.Index - n
		if nextLineIndex >= len(e.bufferLines) || nextLineIndex < 0 {
			return nil
		}

		nextLine := e.bufferLines[nextLineIndex]
		newcol := nextLine.startIndex
		e.Cursors[i].Mark = newcol
		e.ScrollIfNeeded()
	}

	return nil
}

func MarkDown(e *BufferView, n int) error {
	for i := range e.Cursors {
		currentLine := e.getBufferLineForIndex(e.Cursors[i].Mark)
		nextLineIndex := currentLine.Index + n
		if nextLineIndex >= len(e.bufferLines) {
			return nil
		}

		nextLine := e.bufferLines[nextLineIndex]
		newcol := nextLine.startIndex
		e.Cursors[i].Mark = newcol
		e.ScrollIfNeeded()
	}

	return nil
}

func MarkPreviousWord(e *BufferView) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findIndexPositionInTokens(cur.Mark)
		if tokenPos != -1 && tokenPos-1 >= 0 {
			e.Cursors[i].Mark = e.Buffer.tokens[tokenPos-1].Start
		}
		e.ScrollIfNeeded()
	}

	return nil
}

func MarkNextWord(e *BufferView) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findIndexPositionInTokens(cur.Mark)
		if tokenPos != -1 && tokenPos != len(e.Buffer.tokens)-1 {
			e.Cursors[i].Mark = e.Buffer.tokens[tokenPos+1].Start
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func MarkToEndOfLine(e *BufferView) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].End())
		e.Cursors[i].Mark = line.endIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func MarkToBeginningOfLine(e *BufferView) error {
	for i := range e.Cursors {
		line := e.getBufferLineForIndex(e.Cursors[i].End())
		e.Cursors[i].Mark = line.startIndex
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}
func MarkToMatchingChar(e *BufferView) error {
	for i := range e.Cursors {
		matching := byteutils.FindMatching(e.Buffer.Content, e.Cursors[i].Point)
		if matching != -1 {
			e.Cursors[i].Mark = matching
		}
	}

	e.removeDuplicateSelectionsAndSort()
	return nil

}

// @Cursors

func RemoveAllCursorsButOne(p *BufferView) error {
	p.Cursors = p.Cursors[:1]
	p.Cursors[0].Point = p.Cursors[0].Mark

	return nil
}

func AddCursorNextLine(e *BufferView) error {
	pos := e.BufferIndexToPosition(e.Cursors[len(e.Cursors)-1].Start())
	pos.Line++
	if e.isValidCursorPosition(pos) {
		newidx := e.PositionToBufferIndex(pos)
		e.Cursors = append(e.Cursors, Cursor{Point: newidx, Mark: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func AddCursorPreviousLine(e *BufferView) error {
	pos := e.BufferIndexToPosition(e.Cursors[len(e.Cursors)-1].Start())
	pos.Line--
	if e.isValidCursorPosition(pos) {
		newidx := e.PositionToBufferIndex(pos)
		e.Cursors = append(e.Cursors, Cursor{Point: newidx, Mark: newidx})
	}
	e.removeDuplicateSelectionsAndSort()

	return nil
}

func PointForwardWord(e *BufferView, n int) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		cur.SetBoth(cur.Point)
		tokenPos := e.findIndexPositionInTokens(cur.Mark)
		if tokenPos != -1 && tokenPos != len(e.Buffer.tokens)-1 {
			cur.SetBoth(e.Buffer.tokens[tokenPos+1].Start)
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func PointBackwardWord(e *BufferView, n int) error {
	for i := range e.Cursors {
		cur := &e.Cursors[i]
		cur.SetBoth(cur.Point)
		tokenPos := e.findIndexPositionInTokens(cur.Point)
		if tokenPos != -1 && tokenPos != 0 {
			cur.SetBoth(e.Buffer.tokens[tokenPos-1].Start)
		}
		e.ScrollIfNeeded()

	}

	return nil
}

func Write(e *BufferView) error {
	if e.Buffer.Readonly && e.IsSpecial() {
		return nil
	}

	if e.Buffer.fileType.TabSize != 0 {
		e.Buffer.Content = bytes.Replace(e.Buffer.Content, []byte(strings.Repeat(" ", e.Buffer.fileType.TabSize)), []byte("\t"), -1)
	}

	if e.Buffer.fileType.BeforeSave != nil {
		_ = e.Buffer.fileType.BeforeSave(e)
	}

	if e.Buffer.CRLF {
		e.Buffer.Content = bytes.Replace(e.Buffer.Content, []byte("\n"), []byte("\r\n"), -1)
	}

	if err := os.WriteFile(e.Buffer.File, e.Buffer.Content, 0644); err != nil {
		return err
	}
	e.SetStateClean()
	e.replaceTabsWithSpaces()
	if e.Buffer.CRLF {
		e.Buffer.Content = bytes.Replace(e.Buffer.Content, []byte("\r\n"), []byte("\n"), -1)
	}
	if e.Buffer.fileType.AfterSave != nil {
		_ = e.Buffer.fileType.AfterSave(e)

	}

	return nil
}

func Copy(e *BufferView) error {
	if len(e.Cursors) > 1 {
		return nil
	}
	cur := e.Cursors[0]
	if cur.Start() != cur.End() {
		// Copy selection
		end := cur.End() + 1
		if end >= len(e.Buffer.Content) {
			end = len(e.Buffer.Content) - 1
		}
		WriteToClipboard(e.Buffer.Content[cur.Start():end])
	} else {
		line := e.getBufferLineForIndex(cur.Start())
		WriteToClipboard(e.Buffer.Content[line.startIndex : line.endIndex+1])
	}

	return nil
}

func CompileAskForCommand(a *BufferView) error {
	a.parent.SetPrompt("Compile", nil, func(userInput string, c *Context) error {
		a.LastCompileCommand = userInput
		if err := a.parent.OpenCompilationBufferInBuildWindow(userInput); err != nil {
			return err
		}

		return nil
	}, nil, a.LastCompileCommand)

	return nil
}

func CompileNoAsk(a *BufferView) error {
	if a.LastCompileCommand == "" {
		return CompileAskForCommand(a)
	}

	if err := a.parent.OpenCompilationBufferInBuildWindow(a.LastCompileCommand); err != nil {
		return err
	}

	return nil
}

func GrepAsk(a *BufferView) error {
	a.parent.SetPrompt("Grep", nil, func(userInput string, c *Context) error {
		if err := a.parent.OpenGrepBufferInSensibleSplit(fmt.Sprintf("rg --vimgrep %s", userInput)); err != nil {
			return err
		}

		return nil
	}, nil, "")

	return nil
}

func ISearchDeleteBackward(e *BufferView) error {
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

func ISearchActivate(e *BufferView) error {
	e.ISearch.IsSearching = true
	e.ISearch.SearchString = ""
	e.keymaps = append(e.keymaps, SearchTextBufferKeymap, MakeInsertionKeys(func(c *Context, b byte) error {
		e.ISearch.SearchString += string(b)
		return nil
	}))
	return nil
}

func ISearchExit(editor *BufferView) error {
	editor.keymaps = editor.keymaps[:len(editor.keymaps)-2]
	editor.ISearch.IsSearching = false
	editor.ISearch.SearchMatches = nil
	editor.ISearch.CurrentMatch = 0
	editor.ISearch.MovedAwayFromCurrentMatch = false
	return nil
}

func ISearchNextMatch(editor *BufferView) error {
	editor.ISearch.CurrentMatch++
	if editor.ISearch.CurrentMatch >= len(editor.ISearch.SearchMatches) {
		editor.ISearch.CurrentMatch = 0
	}
	editor.ISearch.MovedAwayFromCurrentMatch = false
	return nil
}

func ISearchPreviousMatch(editor *BufferView) error {
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
