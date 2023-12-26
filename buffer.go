package preditor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/smacker/go-tree-sitter/golang"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
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

var BufferKeymap = Keymap{}
var SearchKeymap = Keymap{}
var QueryReplaceKeymap = Keymap{}
var CompileKeymap = Keymap{}

func BufferOpenLocationInCurrentLine(c *Context) {
	b, ok := c.ActiveDrawable().(*BufferView)
	if !ok {
		return
	}

	line := BufferGetCurrentLine(b)
	if line == nil || len(line) < 1 {
		return
	}

	segs := bytes.SplitN(line, []byte(":"), 4)
	if len(segs) < 2 {
		return

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
	_ = SwitchOrOpenFileInWindow(c, c.Cfg, filepath.Join(c.getCWD(), string(filename)), &Position{Line: lineNum, Column: col}, targetWindow)

	c.ActiveWindowIndex = targetWindow.ID
	return
}

func RunCommandInBuffer(parent *Context, cfg *Config, bufferName string, command string) (*BufferView, error) {
	cwd := parent.getCWD()
	bufferView := NewBufferViewFromFilename(parent, cfg, fmt.Sprintf("%s@%s", bufferName, cwd))

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
				bufferView.Buffer.Content = append(bufferView.Buffer.Content, []byte(err.Error())...)
				bufferView.Buffer.Content = append(bufferView.Buffer.Content, '\n')
			}
			if bytes.Contains(output, []byte("\r")) {
				output = bytes.Replace(output, []byte("\r"), []byte(""), -1)
			}
			bufferView.Buffer.Content = append(bufferView.Buffer.Content, output...)
			bufferView.Buffer.Content = append(bufferView.Buffer.Content, []byte(fmt.Sprintf("Done in %s\n", time.Since(since)))...)

		}()

	}

	thisKeymap := CompileKeymap.Clone()
	thisKeymap.BindKey(Key{K: "g"}, MakeCommand(func(b *BufferView) {
		runCompileCommand()
	}))

	bufferView.keymaps.Push(thisKeymap)

	runCompileCommand()
	return bufferView, nil
}

func NewGrepBuffer(parent *Context, cfg *Config, command string) (*BufferView, error) {
	return RunCommandInBuffer(parent, cfg, "*Grep", command)
}

func NewCompilationBuffer(parent *Context, cfg *Config, command string) (*BufferView, error) {
	return RunCommandInBuffer(parent, cfg, "*Compilation", command)
}

func NewBufferViewFromFilename(parent *Context, cfg *Config, filename string) *BufferView {
	buffer := parent.GetBufferByFilename(filename)
	if buffer == nil {
		buffer = parent.OpenFileAsBuffer(filename)
	}

	//check if we have a buffer view for this file that is not shown in any window
OUTER:
	for _, d := range parent.Drawables {
		bufView, ok := d.(*BufferView)
		if !ok {
			continue
		}
		if bufView.Buffer == buffer {
			for _, col := range parent.Windows {
				for _, win := range col {
					if win.DrawableID == bufView.ID {
						continue OUTER
					}
				}
			}

			// bufView is not shown in any window so it's suitable for us to return it
			return bufView
		}

	}

	return NewBufferView(parent, cfg, buffer)
}

type FileType struct {
	Name                     string
	TSLanguage               *sitter.Language
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
		".go":  GoFileType,
		".php": PHPFileType,
	}
}

func NewBufferView(parent *Context, cfg *Config, buffer *Buffer) *BufferView {
	t := BufferView{cfg: cfg}
	t.parent = parent
	t.Buffer = buffer
	t.LastCompileCommand = buffer.fileType.DefaultCompileCommand
	t.keymaps = NewStack[Keymap](5)
	t.keymaps.Push(BufferKeymap)
	t.ActionStack = NewStack[BufferAction](1000)
	t.Cursor = Cursor{Point: 0, Mark: 0}
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

type Search struct {
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

	oldTSTree   *sitter.Tree
	highlights  []highlight
	needParsing bool
	fileType    FileType
}

type QueryReplace struct {
	IsQueryReplace            bool
	SearchString              string
	ReplaceString             string
	SearchMatches             [][]int
	CurrentMatch              int
	MovedAwayFromCurrentMatch bool
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
	textZeroLocation           rl.Vector2
	bufferLines                []BufferLine
	VisibleStart               int32
	MoveToPositionInNextRender *Position
	OldBufferContentLen        int

	keymaps *Stack[Keymap]

	// Cursor
	Cursor         Cursor
	lastCursorTime time.Time
	showCursors    bool

	// Searching
	Search Search

	QueryReplace QueryReplace

	LastCompileCommand string

	ActionStack *Stack[BufferAction]
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
	return fmt.Sprintf("%s:%d", e.Buffer.File, e.Cursor.Point)
}

func (e *BufferView) Keymaps() []Keymap {
	return e.keymaps.data
}

func (e *BufferView) IsSpecial() bool {
	return e.Buffer.File == "" || e.Buffer.File[0] == '*'
}

func (e *BufferView) AddBufferAction(a BufferAction) {
	a.Data = bytes.Clone(a.Data)
	e.ActionStack.Push(a)
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
	tabSize := e.Buffer.fileType.TabSize
	if e.Buffer.fileType.TabSize == 0 {
		tabSize = 4
	}
	e.Buffer.Content = bytes.Replace(e.Buffer.Content, []byte("\t"), []byte(strings.Repeat(" ", tabSize)), -1)
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
	if len(e.bufferLines) < 1 {
		return nil
	}

	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	apprLine := math.Floor(float64((pos.Y - e.textZeroLocation.Y) / charSize.Y))
	apprColumn := math.Floor(float64((pos.X - e.textZeroLocation.X) / charSize.X))

	line := int(apprLine) + int(e.VisibleStart)
	col := int(apprColumn)
	if line >= len(e.bufferLines) {
		line = len(e.bufferLines) - 1
	}

	col -= e.getLineNumbersMaxLength()

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

	e.Cursor.SetBoth(e.PositionToBufferIndex(Position{Line: line, Column: col}))

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
	rangeData := bytes.Clone(e.Buffer.Content[start:end])
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

func (e *BufferView) getLineNumbersMaxLength() int {
	return len(fmt.Sprint(len(e.bufferLines))) + 1
}

func BufferGetCurrentLine(e *BufferView) []byte {
	pos := e.BufferIndexToPosition(e.Cursor.Point)

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
	if len(data) == 0 || len(pattern) == 0 {
		return nil
	}
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

func TSHighlights(fileType *FileType, cfg *Config, queryString []byte, prev *sitter.Tree, code []byte) ([]highlight, *sitter.Tree, error) {
	var highlights []highlight
	parser := sitter.NewParser()
	parser.SetLanguage(fileType.TSLanguage)

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

func safeSlice[T any](s []T, start int, end int) []T {
	if len(s) == 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	if end >= len(s) {
		end = len(s) - 1
	}

	return s[start:end]
}

type bufferRenderContext struct {
	textZeroLocation rl.Vector2
	zeroLocation     rl.Vector2
}

func (e *BufferView) calcRenderState() {
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
		if e.maxColumn != 0 && int32(lineCharCounter) > e.maxColumn-5 {
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

}

func (e *BufferView) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	oldMaxLine := e.maxLine
	oldMaxColumn := e.maxColumn
	oldBufferContentLen := e.OldBufferContentLen
	charSize := measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0)
	e.maxColumn = int32(maxW / float64(charSize.X))
	e.maxLine = int32(maxH / float64(charSize.Y))
	e.maxLine-- //reserve one line of screen for statusbar
	textZeroLocation := zeroLocation
	if e.Search.IsSearching || e.QueryReplace.IsQueryReplace {
		textZeroLocation.Y += charSize.Y
	}
	if e.Buffer.needParsing || len(e.Buffer.Content) != oldBufferContentLen || (e.maxLine != oldMaxLine) || (e.maxColumn != oldMaxColumn) {
		e.calcRenderState()
	}
	e.OldBufferContentLen = len(e.Buffer.Content)
	if !e.NoStatusbar && !e.parent.GlobalNoStatusbar {
		var sections []string

		file := e.Buffer.File

		var state string
		if e.Buffer.State == State_Dirty {
			state = "U"
		} else {
			state = ""
		}
		sections = append(sections, fmt.Sprintf("%s %s", state, file))

		if e.Cursor.Start() == e.Cursor.End() {
			selStart := e.BufferIndexToPosition(e.Cursor.Start())
			if len(e.bufferLines) > selStart.Line {
				selLine := e.bufferLines[selStart.Line]
				sections = append(sections, fmt.Sprintf("L#%d C#%d", selLine.ActualLine, selStart.Column))
			} else {
				sections = append(sections, fmt.Sprintf("L#%d C#%d", selStart.Line, selStart.Column))
			}

		} else {
			selEnd := e.BufferIndexToPosition(e.Cursor.End())
			sections = append(sections, fmt.Sprintf("L#%d C#%d (Selected %d)", selEnd.Line, selEnd.Column, int(math.Abs(float64(e.Cursor.Start()-e.Cursor.End())))))
		}

		if e.Search.IsSearching {
			sections = append(sections, fmt.Sprintf("Search: Match#%d Of %d", e.Search.CurrentMatch+1, len(e.Search.SearchMatches)+1))
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

		textZeroLocation.Y += measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0).Y
		zeroLocation.Y += measureTextSize(e.parent.Font, ' ', e.parent.FontSize, 0).Y

	}

	if e.MoveToPositionInNextRender != nil {
		var bufferIndex int
		for _, line := range e.bufferLines {
			if line.ActualLine == e.MoveToPositionInNextRender.Line {
				bufferIndex = line.startIndex + e.MoveToPositionInNextRender.Column
				e.VisibleStart = int32(e.MoveToPositionInNextRender.Line) - e.maxLine/2
			}
		}
		e.Cursor.SetBoth(bufferIndex)
		e.MoveToPositionInNextRender = nil
	}
	if e.Buffer.needParsing {
		var err error
		e.Buffer.highlights, e.Buffer.oldTSTree, err = TSHighlights(&e.Buffer.fileType, e.cfg, e.Buffer.fileType.TSHighlightQuery, nil, e.Buffer.Content) //TODO: see how we can use old tree
		if err != nil {
			panic(err)
		}
		e.Buffer.needParsing = false
	}

	if e.Search.IsSearching {
		for idx, match := range e.Search.SearchMatches {
			if idx == e.Search.CurrentMatch {
				matchStartLine := e.BufferIndexToPosition(match[0])
				matchEndLine := e.BufferIndexToPosition(match[0])
				if !(e.VisibleStart < int32(matchStartLine.Line) && e.VisibleEnd() > int32(matchEndLine.Line)) && !e.Search.MovedAwayFromCurrentMatch {
					// current match is not in view
					// move the view
					e.VisibleStart = int32(matchStartLine.Line) - e.maxLine/2
					if e.VisibleStart < 0 {
						e.VisibleStart = int32(matchStartLine.Line)
					}

				}
			}
			e.highlightBetweenTwoIndexes(textZeroLocation, match[0], match[1], maxH, maxW, e.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), e.cfg.CurrentThemeColors().SelectionForeground.ToColorRGBA())
		}
		e.Search.LastSearchString = e.Search.SearchString
		if len(e.Search.SearchMatches) > 0 {
			e.Cursor.Point = e.Search.SearchMatches[e.Search.CurrentMatch][0]
			e.Cursor.Mark = e.Search.SearchMatches[e.Search.CurrentMatch][0]
		}

		rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts.ToColorRGBA())
		rl.DrawTextEx(e.parent.Font, fmt.Sprintf("Search: %s", e.Search.SearchString), rl.Vector2{
			X: zeroLocation.X,
			Y: zeroLocation.Y,
		}, float32(e.parent.FontSize), 0, rl.White)

	}

	//QueryReplace
	if e.QueryReplace.IsQueryReplace {
		for idx, match := range e.QueryReplace.SearchMatches {
			if idx == e.QueryReplace.CurrentMatch {
				matchStartLine := e.BufferIndexToPosition(match[0])
				matchEndLine := e.BufferIndexToPosition(match[0])
				if !(e.VisibleStart < int32(matchStartLine.Line) && e.VisibleEnd() > int32(matchEndLine.Line)) && !e.Search.MovedAwayFromCurrentMatch {
					// current match is not in view
					// move the view
					e.VisibleStart = int32(matchStartLine.Line) - e.maxLine/2
					if e.VisibleStart < 0 {
						e.VisibleStart = int32(matchStartLine.Line)
					}

				}
			}
			e.highlightBetweenTwoIndexes(textZeroLocation, match[0], match[1], maxH, maxW, e.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), e.cfg.CurrentThemeColors().SelectionForeground.ToColorRGBA())
		}
		if len(e.QueryReplace.SearchMatches) > 0 {
			e.Cursor.Point = e.QueryReplace.SearchMatches[e.QueryReplace.CurrentMatch][0]
			e.Cursor.Mark = e.QueryReplace.SearchMatches[e.QueryReplace.CurrentMatch][0]
		}

		rl.DrawRectangle(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y), e.cfg.CurrentThemeColors().Prompts.ToColorRGBA())
		rl.DrawTextEx(e.parent.Font, fmt.Sprintf("QueryReplace: %s -> %s", e.QueryReplace.SearchString, e.QueryReplace.ReplaceString), rl.Vector2{
			X: zeroLocation.X,
			Y: zeroLocation.Y,
		}, float32(e.parent.FontSize), 0, rl.White)

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

	//TODO: @Perf we should check and re render if view has changed or buffer lines has changed
	for idx, line := range visibleLines {
		if e.cfg.LineNumbers {
			rl.DrawTextEx(e.parent.Font,
				fmt.Sprintf("%d", line.ActualLine),
				rl.Vector2{X: textZeroLocation.X, Y: textZeroLocation.Y + float32(idx)*charSize.Y},
				float32(e.parent.FontSize),
				0,
				e.cfg.CurrentThemeColors().LineNumbersForeground.ToColorRGBA())
		}
		e.renderTextRange(textZeroLocation, line.startIndex, line.endIndex, maxH, maxW, e.cfg.CurrentThemeColors().Foreground.ToColorRGBA())
	}
	if e.cfg.EnableSyntaxHighlighting {
		if len(e.bufferLines) > 0 {
			visibleStartChar := e.bufferLines[e.VisibleStart].startIndex

			var visibleEndChar int
			if len(e.bufferLines) > int(e.VisibleEnd()) {
				visibleEndChar = e.bufferLines[e.VisibleEnd()].endIndex
			} else {
				visibleEndChar = len(e.Buffer.Content)
			}

			for _, h := range e.Buffer.highlights {
				if visibleStartChar <= h.start && visibleEndChar >= h.end {
					e.renderTextRange(textZeroLocation, h.start, h.end, maxH, maxW, h.Color)
				}
			}
		}
	}

	// render cursors
	cursorBlinkDeltaTime := time.Since(e.lastCursorTime)
	if cursorBlinkDeltaTime > time.Millisecond*800 {
		e.showCursors = !e.showCursors
		e.lastCursorTime = time.Now()
	}

	if !e.cfg.CursorBlinking {
		e.showCursors = true
	}

	if e.parent.ActiveDrawableID() == e.ID {
		if e.Cursor.Start() == e.Cursor.End() {
			cursor := e.BufferIndexToPosition(e.Cursor.Start())
			cursorView := Position{
				Line:   cursor.Line - int(e.VisibleStart),
				Column: cursor.Column,
			}
			posX := int32(cursorView.Column)*int32(charSize.X) + int32(textZeroLocation.X)
			if e.cfg.LineNumbers {
				posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
			}
			posY := int32(cursorView.Line)*int32(charSize.Y) + int32(textZeroLocation.Y)

			// if !isVisibleInWindow(float64(posX), float64(posY), zeroLocation, maxH, maxW) {
			//	 continue
			// }
			if e.showCursors {
				switch e.cfg.CursorShape {
				case CURSOR_SHAPE_OUTLINE:
					rl.DrawRectangleLines(posX, posY, int32(charSize.X), int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
				case CURSOR_SHAPE_BLOCK:
					rl.DrawRectangle(posX, posY, int32(charSize.X), int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
					if len(e.Buffer.Content)-1 >= e.Cursor.Point {
						rl.DrawTextEx(e.parent.Font, string(e.Buffer.Content[e.Cursor.Point]), rl.Vector2{X: float32(posX), Y: float32(posY)}, float32(e.parent.FontSize), 0, e.cfg.CurrentThemeColors().Background.ToColorRGBA())
					}

				case CURSOR_SHAPE_LINE:
					rl.DrawRectangleLines(posX, posY, 2, int32(charSize.Y), e.cfg.CurrentThemeColors().Cursor.ToColorRGBA())
				}
			}
			if e.cfg.CursorLineHighlight {
				rl.DrawRectangle(int32(textZeroLocation.X), int32(cursorView.Line)*int32(charSize.Y)+int32(textZeroLocation.Y), e.maxColumn*int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().CursorLineBackground.ToColorRGBA(), 0.15))
			}

			// highlight matching char
			if e.cfg.HighlightMatchingParen {
				matchingIdx := byteutils.FindMatching(e.Buffer.Content, e.Cursor.Point)
				if matchingIdx != -1 {
					idxPosition := e.BufferIndexToPosition(matchingIdx)
					idxPositionView := Position{
						Line:   idxPosition.Line - int(e.VisibleStart),
						Column: idxPosition.Column,
					}
					posX := int32(idxPositionView.Column)*int32(charSize.X) + int32(textZeroLocation.X)
					if e.cfg.LineNumbers {
						posX += int32(e.getLineNumbersMaxLength()) * int32(charSize.X)
					}
					posY := int32(idxPositionView.Line)*int32(charSize.Y) + int32(textZeroLocation.Y)

					rl.DrawRectangle(posX, posY, int32(charSize.X), int32(charSize.Y), rl.Fade(e.cfg.CurrentThemeColors().HighlightMatching.ToColorRGBA(), 0.4))
				}
			}

		} else {
			e.highlightBetweenTwoIndexes(textZeroLocation, e.Cursor.Start(), e.Cursor.End(), maxH, maxW, e.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), e.cfg.CurrentThemeColors().SelectionForeground.ToColorRGBA())
		}
	}

	e.zeroLocation = zeroLocation
	e.textZeroLocation = textZeroLocation
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
	if e.Cursor.Start() == e.Cursor.End() {
		return
	}
	e.AddBufferAction(BufferAction{
		Type: BufferActionType_Delete,
		Idx:  e.Cursor.Start(),
		Data: e.Buffer.Content[e.Cursor.Start() : e.Cursor.End()+1],
	})
	e.Buffer.Content = append(e.Buffer.Content[:e.Cursor.Start()], e.Buffer.Content[e.Cursor.End()+1:]...)
	e.Cursor.Point = e.Cursor.Mark
	e.moveLeft(&e.Cursor, old-1-len(e.Buffer.Content))
}

func (e *BufferView) moveLeft(s *Cursor, n int) {
}

func (e *BufferView) ScrollIfNeeded() {
	pos := e.BufferIndexToPosition(e.Cursor.End())
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

	//replace CRLF with LF
	if bytes.Index(bs, []byte("\r\n")) != -1 {
		bs = bytes.Replace(bs, []byte("\r"), []byte(""), -1)
		e.Buffer.CRLF = true
	}
	e.Buffer.Content = bs
	e.replaceTabsWithSpaces()
	e.SetStateClean()
	return nil
}

// Things that change buffer content

func BufferInsertChar(e *BufferView, char byte) {
	if e.Buffer.Readonly {
		return
	}
	e.deleteSelectionsIfAnySelection()
	e.AddBytesAtIndex([]byte{char}, e.Cursor.Point, true)
	PointRight(e, 1)
	e.SetStateDirty()
	e.ScrollIfNeeded()
	return
}

func DeleteCharBackward(e *BufferView) error {
	if e.Buffer.Readonly {
		return nil
	}
	e.deleteSelectionsIfAnySelection()
	e.RemoveRange(e.Cursor.Point-1, e.Cursor.Point, true)
	e.moveLeft(&e.Cursor, 1)
	e.SetStateDirty()
	return nil
}

func DeleteCharForward(e *BufferView) error {
	if e.Buffer.Readonly {
		return nil
	}
	e.deleteSelectionsIfAnySelection()
	e.RemoveRange(e.Cursor.Point, e.Cursor.Point+1, true)

	e.SetStateDirty()
	return nil
}
func RevertLastBufferAction(e *BufferView) {
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
		e.AddBytesAtIndex(last.Data, last.Idx, false)
	}
	e.SetStateDirty()

	if len(e.ActionStack.data) < 1 {
		e.SetStateClean()
	}
}

func WordAtPoint(e *BufferView) (int, int) {
	currentWordStart := byteutils.SeekPreviousNonLetter(e.Buffer.Content, e.Cursor.Point) + 1
	currentWordEnd := byteutils.SeekNextNonLetter(e.Buffer.Content, e.Cursor.Point)

	return currentWordStart, currentWordEnd
}

func LeftWord(e *BufferView) (int, int) {
	leftWordEnd := byteutils.SeekPreviousNonLetter(e.Buffer.Content, e.Cursor.Point)
	leftWordStart := byteutils.SeekPreviousNonLetter(e.Buffer.Content, leftWordEnd-1) + 1

	return leftWordStart, leftWordEnd
}

func RightWord(e *BufferView) (int, int) {
	rightWordStart := byteutils.SeekNextNonLetter(e.Buffer.Content, e.Cursor.Point) + 1
	rightWordEnd := byteutils.SeekNextNonLetter(e.Buffer.Content, rightWordStart)
	return rightWordStart, rightWordEnd
}

func DeleteWordBackward(e *BufferView) {
	if e.Buffer.Readonly {
		return
	}
	leftWordStart, leftWordEnd := LeftWord(e)
	if leftWordStart == -1 || leftWordEnd == -1 {
		return
	}
	old := len(e.Buffer.Content)
	e.RemoveRange(leftWordStart, e.Cursor.Point, true)
	e.Cursor.SetBoth(e.Cursor.Point + (len(e.Buffer.Content) - old))

	e.SetStateDirty()
}

func Indent(e *BufferView) error {
	e.AddBytesAtIndex([]byte(strings.Repeat(" ", e.Buffer.fileType.TabSize)), e.Cursor.Point, true)
	PointRight(e, e.Buffer.fileType.TabSize)
	e.SetStateDirty()

	return nil
}

func KillLine(e *BufferView) {
	if e.Buffer.Readonly {
		return
	}
	var lastChange int
	old := len(e.Buffer.Content)
	e.moveLeft(&e.Cursor, lastChange)
	line := e.getBufferLineForIndex(e.Cursor.Start())
	e.RemoveRange(e.Cursor.Point, line.endIndex, true)
	lastChange += -1 * (len(e.Buffer.Content) - old)
	e.SetStateDirty()

}
func Cut(e *BufferView) error {
	if e.Buffer.Readonly {
		return nil
	}
	if e.Cursor.Start() != e.Cursor.End() {
		// Copy selection
		WriteToClipboard(e.Buffer.Content[e.Cursor.Start() : e.Cursor.End()+1])
		e.RemoveRange(e.Cursor.Start(), e.Cursor.End()+1, true)
		e.Cursor.Mark = e.Cursor.Point
	} else {
		line := e.getBufferLineForIndex(e.Cursor.Start())
		WriteToClipboard(e.Buffer.Content[line.startIndex : line.endIndex+1])
		e.RemoveRange(line.startIndex, line.endIndex+1, true)
	}
	e.SetStateDirty()

	return nil
}

func Paste(e *BufferView) error {
	if e.Buffer.Readonly {
		return nil
	}
	contentToPaste := GetClipboardContent()
	e.AddBytesAtIndex(contentToPaste, e.Cursor.Start(), true)
	e.SetStateDirty()
	PointRight(e, len(contentToPaste))
	return nil
}

func InteractiveGotoLine(e *BufferView) {
	doneHook := func(userInput string, c *Context) {
		number, err := strconv.Atoi(userInput)
		if err != nil {
			return
		}

		for _, line := range e.bufferLines {
			if line.ActualLine == number {
				e.Cursor.SetBoth(line.startIndex)
				e.ScrollIfNeeded()
			}
		}

		return
	}
	e.parent.SetPrompt("Goto", nil, doneHook, nil, "")

	return
}

// @Scroll

func ScrollUp(e *BufferView, n int) {
	if e.VisibleStart <= 0 {
		return
	}
	e.VisibleStart += int32(-1 * n)

	if e.VisibleStart < 0 {
		e.VisibleStart = 0
	}

	return

}

func ScrollToTop(e *BufferView) {
	e.VisibleStart = 0
	e.Cursor.SetBoth(0)

	return
}

func ScrollToBottom(e *BufferView) {
	e.VisibleStart = int32(len(e.bufferLines) - 1 - int(e.maxLine))
	e.Cursor.SetBoth(e.bufferLines[len(e.bufferLines)-1].startIndex)

	return
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
	e.Cursor.Point -= n
	if e.Cursor.Point < 0 {
		e.Cursor.SetBoth(0)
	}
	e.Cursor.Mark = e.Cursor.Point
	e.ScrollIfNeeded()
	return nil
}

func PointRight(e *BufferView, n int) error {
	e.Cursor.Point += n
	if e.Cursor.Point > len(e.Buffer.Content) {
		e.Cursor.SetBoth(0)
	}
	e.Cursor.Mark = e.Cursor.Point
	e.ScrollIfNeeded()
	return nil
}

func PointUp(e *BufferView) error {
	currentLine := e.getBufferLineForIndex(e.Cursor.Point)
	prevLineIndex := currentLine.Index - 1
	if prevLineIndex < 0 {
		return nil
	}

	prevLine := e.bufferLines[prevLineIndex]
	col := e.Cursor.Point - currentLine.startIndex
	newidx := prevLine.startIndex + col
	if newidx > prevLine.endIndex {
		newidx = prevLine.endIndex
	}
	e.Cursor.SetBoth(newidx)
	e.ScrollIfNeeded()

	return nil
}

func PointDown(e *BufferView) error {
	currentLine := e.getBufferLineForIndex(e.Cursor.Point)
	nextLineIndex := currentLine.Index + 1
	if nextLineIndex >= len(e.bufferLines) {
		return nil
	}

	nextLine := e.bufferLines[nextLineIndex]
	col := e.Cursor.Point - currentLine.startIndex
	newIndex := nextLine.startIndex + col
	if newIndex > nextLine.endIndex {
		newIndex = nextLine.endIndex
	}
	e.Cursor.SetBoth(newIndex)
	e.ScrollIfNeeded()

	return nil
}

func CentralizePoint(e *BufferView) {
	pos := e.convertBufferIndexToLineAndColumn(e.Cursor.Start())
	e.VisibleStart = int32(pos.Line) - (e.maxLine / 2)
	if e.VisibleStart < 0 {
		e.VisibleStart = 0
	}
}

func PointToBeginningOfLine(e *BufferView) error {
	line := e.getBufferLineForIndex(e.Cursor.Start())
	e.Cursor.SetBoth(line.startIndex)
	return nil
}

func PointToEndOfLine(e *BufferView) error {
	line := e.getBufferLineForIndex(e.Cursor.Start())
	e.Cursor.SetBoth(line.endIndex)
	return nil
}

func PointToMatchingChar(e *BufferView) error {
	matching := byteutils.FindMatching(e.Buffer.Content, e.Cursor.Point)
	if matching != -1 {
		e.Cursor.SetBoth(matching)
	}

	return nil

}

// @Mark

func MarkRight(e *BufferView, n int) {
	e.Cursor.Mark += n
	if e.Cursor.Mark >= len(e.Buffer.Content) {
		e.Cursor.Mark = len(e.Buffer.Content)
	}
	e.ScrollIfNeeded()
}

func MarkLeft(e *BufferView, n int) {
	e.Cursor.Mark -= n
	if e.Cursor.Mark < 0 {
		e.Cursor.Mark = 0
	}
	e.ScrollIfNeeded()
}

func MarkUp(e *BufferView, n int) {
	currentLine := e.getBufferLineForIndex(e.Cursor.Mark)
	nextLineIndex := currentLine.Index - n
	if nextLineIndex >= len(e.bufferLines) || nextLineIndex < 0 {
		return
	}

	nextLine := e.bufferLines[nextLineIndex]
	newcol := nextLine.startIndex
	e.Cursor.Mark = newcol
	e.ScrollIfNeeded()
}

func MarkDown(e *BufferView, n int) {
	currentLine := e.getBufferLineForIndex(e.Cursor.Mark)
	nextLineIndex := currentLine.Index + n
	if nextLineIndex >= len(e.bufferLines) {
		return
	}

	nextLine := e.bufferLines[nextLineIndex]
	newcol := nextLine.startIndex
	e.Cursor.Mark = newcol
	e.ScrollIfNeeded()

	return
}

func MarkPreviousWord(e *BufferView) {
	leftWordStart, _ := LeftWord(e)
	if leftWordStart != -1 {
		e.Cursor.Mark = leftWordStart
		e.ScrollIfNeeded()
	}

	return
}

func MarkNextWord(e *BufferView) {
	_, rightWordEnd := RightWord(e)
	if rightWordEnd != -1 {
		e.Cursor.Mark = rightWordEnd
	}
	e.ScrollIfNeeded()

}

func MarkToEndOfLine(e *BufferView) {
	line := e.getBufferLineForIndex(e.Cursor.End())
	e.Cursor.Mark = line.endIndex

	return
}

func MarkToBeginningOfLine(e *BufferView) {
	line := e.getBufferLineForIndex(e.Cursor.End())
	e.Cursor.Mark = line.startIndex
}
func MarkToMatchingChar(e *BufferView) {
	matching := byteutils.FindMatching(e.Buffer.Content, e.Cursor.Point)
	if matching != -1 {
		e.Cursor.Mark = matching
	}

}

// @Cursors

func PointRightWord(e *BufferView) {
	e.Cursor.SetBoth(e.Cursor.Point)
	j := byteutils.SeekNextNonLetter(e.Buffer.Content, e.Cursor.Point)
	if j != -1 {
		e.Cursor.SetBoth(j)
	}
	e.ScrollIfNeeded()

	return
}

func PointLeftWord(e *BufferView) {
	e.Cursor.SetBoth(e.Cursor.Point)
	j := byteutils.SeekPreviousNonLetter(e.Buffer.Content, e.Cursor.Point)
	if j != -1 {
		e.Cursor.SetBoth(j)
	}
	e.ScrollIfNeeded()
}

func Write(e *BufferView) {
	if e.Buffer.Readonly && e.IsSpecial() {
		return
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
		return
	}
	e.SetStateClean()
	e.replaceTabsWithSpaces()
	if e.Buffer.CRLF {
		e.Buffer.Content = bytes.Replace(e.Buffer.Content, []byte("\r\n"), []byte("\n"), -1)
	}
	if e.Buffer.fileType.AfterSave != nil {
		_ = e.Buffer.fileType.AfterSave(e)

	}

	return
}

func Copy(e *BufferView) error {
	if e.Cursor.Start() != e.Cursor.End() {
		// Copy selection
		end := e.Cursor.End() + 1
		if end >= len(e.Buffer.Content) {
			end = len(e.Buffer.Content) - 1
		}
		WriteToClipboard(e.Buffer.Content[e.Cursor.Start():end])
	} else {
		line := e.getBufferLineForIndex(e.Cursor.Start())
		WriteToClipboard(e.Buffer.Content[line.startIndex : line.endIndex+1])
	}

	return nil
}

func CompileAskForCommand(a *BufferView) {
	a.parent.SetPrompt("Compile", nil, func(userInput string, c *Context) {
		a.LastCompileCommand = userInput
		if err := a.parent.OpenCompilationBufferInBuildWindow(userInput); err != nil {
			return
		}

		return
	}, nil, a.LastCompileCommand)

	return
}

func CompileNoAsk(a *BufferView) {
	if a.LastCompileCommand == "" {
		CompileAskForCommand(a)
	}

	if err := a.parent.OpenCompilationBufferInBuildWindow(a.LastCompileCommand); err != nil {
		return
	}

	return
}

func GrepAsk(a *BufferView) {
	a.parent.SetPrompt("Grep", nil, func(userInput string, c *Context) {
		if err := a.parent.OpenGrepBufferInSensibleSplit(fmt.Sprintf("rg --vimgrep %s", userInput)); err != nil {
			return
		}

		return
	}, nil, "")

	return
}

const BIG_FILE_SEARCH_THRESHOLD = 1024 * 1024

func SearchActivate(bufferView *BufferView) {
	if len(bufferView.Buffer.Content) < BIG_FILE_SEARCH_THRESHOLD {
		thisPromptKeymap := PromptKeymap.Clone()
		thisPromptKeymap.BindKey(Key{K: "<esc>"}, func(c *Context) {
			c.ResetPrompt()
			SearchExit(bufferView)
		})
		thisPromptKeymap.BindKey(Key{K: "<enter>"}, func(c *Context) {
			SearchNextMatch(bufferView)
		})
		bufferView.parent.SetPrompt("ISearch", func(query string, c *Context) {
			if !bufferView.Search.IsSearching {
				bufferView.Search.IsSearching = true
				bufferView.keymaps.Push(SearchKeymap)
			}
			bufferView.Search.SearchString = query
			matchPatternAsync(&bufferView.Search.SearchMatches, bufferView.Buffer.Content, []byte(query))
		}, nil, &thisPromptKeymap, "")
		bufferView.parent.Prompt.NoRender = true
	} else {
		bufferView.parent.SetPrompt("Search", nil, func(query string, c *Context) {
			if !bufferView.Search.IsSearching {
				bufferView.Search.IsSearching = true
				bufferView.keymaps.Push(SearchKeymap)
			}
			bufferView.Search.SearchString = query
			matchPatternAsync(&bufferView.Search.SearchMatches, bufferView.Buffer.Content, []byte(query))
		}, nil, "")
	}
}

func SearchExit(editor *BufferView) error {
	if len(editor.Buffer.Content) < BIG_FILE_SEARCH_THRESHOLD {
		editor.keymaps.Pop()
		editor.parent.ResetPrompt()
		editor.Search.IsSearching = false
		editor.Search.SearchMatches = nil
		editor.Search.CurrentMatch = 0
		editor.Search.MovedAwayFromCurrentMatch = false
		editor.parent.Prompt.NoRender = false
		return nil
	} else {
		editor.Search.IsSearching = false
		editor.Search.SearchMatches = nil
		editor.Search.CurrentMatch = 0
		editor.Search.MovedAwayFromCurrentMatch = false
		editor.keymaps.Pop()
	}

	return nil
}

func SearchNextMatch(editor *BufferView) error {
	editor.Search.CurrentMatch++
	if editor.Search.CurrentMatch >= len(editor.Search.SearchMatches) {
		editor.Search.CurrentMatch = 0
	}
	editor.Search.MovedAwayFromCurrentMatch = false
	return nil
}

func SearchPreviousMatch(editor *BufferView) error {
	editor.Search.CurrentMatch--
	if editor.Search.CurrentMatch >= len(editor.Search.SearchMatches) {
		editor.Search.CurrentMatch = 0
	}
	if editor.Search.CurrentMatch < 0 {
		editor.Search.CurrentMatch = len(editor.Search.SearchMatches) - 1
	}
	editor.Search.MovedAwayFromCurrentMatch = false
	return nil
}

func QueryReplaceActivate(bufferView *BufferView) {
	bufferView.parent.SetPrompt("Query", nil, func(query string, c *Context) {
		bufferView.parent.SetPrompt("Replace", nil, func(replace string, c *Context) {
			bufferView.QueryReplace.IsQueryReplace = true
			bufferView.QueryReplace.SearchString = query
			bufferView.QueryReplace.ReplaceString = replace
			bufferView.keymaps.Push(QueryReplaceKeymap)
			matchPatternAsync(&bufferView.QueryReplace.SearchMatches, bufferView.Buffer.Content, []byte(query))
		}, nil, "")
	}, nil, "")
}

func QueryReplaceReplaceThisMatch(bufferView *BufferView) {
	match := bufferView.QueryReplace.SearchMatches[bufferView.QueryReplace.CurrentMatch]
	bufferView.RemoveRange(match[0], match[1]+1, false)
	bufferView.AddBytesAtIndex([]byte(bufferView.QueryReplace.ReplaceString), match[0], false)
	bufferView.QueryReplace.CurrentMatch++
	bufferView.QueryReplace.MovedAwayFromCurrentMatch = false
	if bufferView.QueryReplace.CurrentMatch >= len(bufferView.QueryReplace.SearchMatches) {
		QueryReplaceExit(bufferView)
	}
}
func QueryReplaceIgnoreThisMatch(bufferView *BufferView) {
	bufferView.QueryReplace.CurrentMatch++
	bufferView.QueryReplace.MovedAwayFromCurrentMatch = false
	if bufferView.QueryReplace.CurrentMatch >= len(bufferView.QueryReplace.SearchMatches) {
		QueryReplaceExit(bufferView)
	}
}
func QueryReplaceExit(bufferView *BufferView) {
	bufferView.QueryReplace.IsQueryReplace = false
	bufferView.QueryReplace.SearchString = ""
	bufferView.QueryReplace.ReplaceString = ""
	bufferView.QueryReplace.SearchMatches = nil
	bufferView.QueryReplace.CurrentMatch = 0
	bufferView.QueryReplace.MovedAwayFromCurrentMatch = false
	bufferView.keymaps.Pop()
}

func GetClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func WriteToClipboard(bs []byte) {
	clipboard.Write(clipboard.FmtText, bytes.Clone(bs))
}

func RevertBuffer(bufferView *BufferView) {
	bufferView.readFileFromDisk()
}
