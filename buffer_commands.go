package preditor

import (
	"bytes"
	"fmt"
	"golang.design/x/clipboard"
	"os"
	"strconv"
	"strings"
)

func InsertChar(e *Buffer, char byte) error {
	if e.Readonly {
		return nil
	}
	e.removeDuplicateSelectionsAndSort()
	e.deleteSelectionsIfAnySelection()
	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], i*1)

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
		if len(e.Content) > e.Cursors[i].Start()+1 {
			e.moveLeft(&e.Cursors[i], i*1)
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

func DeleteWordBackward(e *Buffer) {
	if e.Readonly || len(e.Cursors) > 1 {
		return
	}
	e.deleteSelectionsIfAnySelection()

	for i := range e.Cursors {
		cur := &e.Cursors[i]
		tokenPos := e.findIndexPositionInTokens(cur.Point)
		if tokenPos == -1 {
			continue
		}
		start := e.Tokens[tokenPos].Start
		if start == cur.Point && tokenPos-1 >= 0 {
			start = e.Tokens[tokenPos-1].Start
		}
		old := len(e.Content)
		if len(e.Content) > cur.Start()+1 {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  start,
				Data: e.Content[start:cur.Start()],
			})
			e.Content = append(e.Content[:start], e.Content[cur.Start():]...)
		} else {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Delete,
				Idx:  start,
				Data: e.Content[start:],
			})
			e.Content = e.Content[:start]
		}
		cur.SetBoth(cur.Point + (len(e.Content) - old))
	}

	e.SetStateDirty()
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
	e.calculateVisualLines()
	if e.fileType.AfterSave != nil {
		_ = e.fileType.AfterSave(e)

	}

	return nil
}

func Indent(e *Buffer) error {
	e.removeDuplicateSelectionsAndSort()

	for i := range e.Cursors {
		e.moveRight(&e.Cursors[i], i*e.fileType.TabSize)
		if e.Cursors[i].Start() >= len(e.Content) { // end of file, appending
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Cursors[i].Start(),
				Data: []byte(strings.Repeat(" ", e.fileType.TabSize)),
			})
			e.Content = append(e.Content, []byte(strings.Repeat(" ", e.fileType.TabSize))...)
		} else {
			e.AddUndoAction(EditorAction{
				Type: EditorActionType_Insert,
				Idx:  e.Cursors[i].Start(),
				Data: []byte(strings.Repeat(" ", e.fileType.TabSize)),
			})
			e.Content = append(e.Content[:e.Cursors[i].Start()], append([]byte(strings.Repeat(" ", e.fileType.TabSize)), e.Content[e.Cursors[i].Start():]...)...)
		}
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
		WriteToClipboard(e.Content[cur.Start():line.endIndex])
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

func Copy(e *Buffer) error {
	if len(e.Cursors) > 1 {
		return nil
	}
	cur := e.Cursors[0]
	if cur.Start() != cur.End() {
		// Copy selection
		WriteToClipboard(e.Content[cur.Start():cur.End()])
	} else {
		line := e.getBufferLineForIndex(cur.Start())
		WriteToClipboard(e.Content[line.startIndex : line.endIndex+1])
	}

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
		e.AddUndoAction(EditorAction{
			Type: EditorActionType_Delete,
			Idx:  cur.Start(),
			Data: e.Content[cur.Start():cur.End()],
		})
		e.Content = append(e.Content[:cur.Start()], e.Content[cur.End()+1:]...)
	} else {
		line := e.getBufferLineForIndex(cur.Start())
		WriteToClipboard(e.Content[line.startIndex : line.endIndex+1])
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

func Paste(e *Buffer) error {
	if e.Readonly || len(e.Cursors) > 1 {
		return nil
	}
	e.deleteSelectionsIfAnySelection()
	contentToPaste := GetClipboardContent()
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

	PointRight(e, len(contentToPaste))
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
