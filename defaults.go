package preditor

import rl "github.com/gen2brain/raylib-go/raylib"

func MakeInsertionKeys(insertor func(c *Context, b byte)) Keymap {
	return Keymap{
		Key{K: "a"}:                    func(c *Context) { insertor(c, 'a') },
		Key{K: "b"}:                    func(c *Context) { insertor(c, 'b') },
		Key{K: "c"}:                    func(c *Context) { insertor(c, 'c') },
		Key{K: "d"}:                    func(c *Context) { insertor(c, 'd') },
		Key{K: "e"}:                    func(c *Context) { insertor(c, 'e') },
		Key{K: "f"}:                    func(c *Context) { insertor(c, 'f') },
		Key{K: "g"}:                    func(c *Context) { insertor(c, 'g') },
		Key{K: "h"}:                    func(c *Context) { insertor(c, 'h') },
		Key{K: "i"}:                    func(c *Context) { insertor(c, 'i') },
		Key{K: "j"}:                    func(c *Context) { insertor(c, 'j') },
		Key{K: "k"}:                    func(c *Context) { insertor(c, 'k') },
		Key{K: "l"}:                    func(c *Context) { insertor(c, 'l') },
		Key{K: "m"}:                    func(c *Context) { insertor(c, 'm') },
		Key{K: "n"}:                    func(c *Context) { insertor(c, 'n') },
		Key{K: "o"}:                    func(c *Context) { insertor(c, 'o') },
		Key{K: "p"}:                    func(c *Context) { insertor(c, 'p') },
		Key{K: "q"}:                    func(c *Context) { insertor(c, 'q') },
		Key{K: "r"}:                    func(c *Context) { insertor(c, 'r') },
		Key{K: "s"}:                    func(c *Context) { insertor(c, 's') },
		Key{K: "t"}:                    func(c *Context) { insertor(c, 't') },
		Key{K: "u"}:                    func(c *Context) { insertor(c, 'u') },
		Key{K: "v"}:                    func(c *Context) { insertor(c, 'v') },
		Key{K: "w"}:                    func(c *Context) { insertor(c, 'w') },
		Key{K: "x"}:                    func(c *Context) { insertor(c, 'x') },
		Key{K: "y"}:                    func(c *Context) { insertor(c, 'y') },
		Key{K: "z"}:                    func(c *Context) { insertor(c, 'z') },
		Key{K: "0"}:                    func(c *Context) { insertor(c, '0') },
		Key{K: "1"}:                    func(c *Context) { insertor(c, '1') },
		Key{K: "2"}:                    func(c *Context) { insertor(c, '2') },
		Key{K: "3"}:                    func(c *Context) { insertor(c, '3') },
		Key{K: "4"}:                    func(c *Context) { insertor(c, '4') },
		Key{K: "5"}:                    func(c *Context) { insertor(c, '5') },
		Key{K: "6"}:                    func(c *Context) { insertor(c, '6') },
		Key{K: "7"}:                    func(c *Context) { insertor(c, '7') },
		Key{K: "8"}:                    func(c *Context) { insertor(c, '8') },
		Key{K: "9"}:                    func(c *Context) { insertor(c, '9') },
		Key{K: "\\"}:                   func(c *Context) { insertor(c, '\\') },
		Key{K: "\\", Shift: true}:      func(c *Context) { insertor(c, '|') },
		Key{K: "0", Shift: true}:       func(c *Context) { insertor(c, ')') },
		Key{K: "1", Shift: true}:       func(c *Context) { insertor(c, '!') },
		Key{K: "2", Shift: true}:       func(c *Context) { insertor(c, '@') },
		Key{K: "3", Shift: true}:       func(c *Context) { insertor(c, '#') },
		Key{K: "4", Shift: true}:       func(c *Context) { insertor(c, '$') },
		Key{K: "5", Shift: true}:       func(c *Context) { insertor(c, '%') },
		Key{K: "6", Shift: true}:       func(c *Context) { insertor(c, '^') },
		Key{K: "7", Shift: true}:       func(c *Context) { insertor(c, '&') },
		Key{K: "8", Shift: true}:       func(c *Context) { insertor(c, '*') },
		Key{K: "9", Shift: true}:       func(c *Context) { insertor(c, '(') },
		Key{K: "a", Shift: true}:       func(c *Context) { insertor(c, 'A') },
		Key{K: "b", Shift: true}:       func(c *Context) { insertor(c, 'B') },
		Key{K: "c", Shift: true}:       func(c *Context) { insertor(c, 'C') },
		Key{K: "d", Shift: true}:       func(c *Context) { insertor(c, 'D') },
		Key{K: "e", Shift: true}:       func(c *Context) { insertor(c, 'E') },
		Key{K: "f", Shift: true}:       func(c *Context) { insertor(c, 'F') },
		Key{K: "g", Shift: true}:       func(c *Context) { insertor(c, 'G') },
		Key{K: "h", Shift: true}:       func(c *Context) { insertor(c, 'H') },
		Key{K: "i", Shift: true}:       func(c *Context) { insertor(c, 'I') },
		Key{K: "j", Shift: true}:       func(c *Context) { insertor(c, 'J') },
		Key{K: "k", Shift: true}:       func(c *Context) { insertor(c, 'K') },
		Key{K: "l", Shift: true}:       func(c *Context) { insertor(c, 'L') },
		Key{K: "m", Shift: true}:       func(c *Context) { insertor(c, 'M') },
		Key{K: "n", Shift: true}:       func(c *Context) { insertor(c, 'N') },
		Key{K: "o", Shift: true}:       func(c *Context) { insertor(c, 'O') },
		Key{K: "p", Shift: true}:       func(c *Context) { insertor(c, 'P') },
		Key{K: "q", Shift: true}:       func(c *Context) { insertor(c, 'Q') },
		Key{K: "r", Shift: true}:       func(c *Context) { insertor(c, 'R') },
		Key{K: "s", Shift: true}:       func(c *Context) { insertor(c, 'S') },
		Key{K: "t", Shift: true}:       func(c *Context) { insertor(c, 'T') },
		Key{K: "u", Shift: true}:       func(c *Context) { insertor(c, 'U') },
		Key{K: "v", Shift: true}:       func(c *Context) { insertor(c, 'V') },
		Key{K: "w", Shift: true}:       func(c *Context) { insertor(c, 'W') },
		Key{K: "x", Shift: true}:       func(c *Context) { insertor(c, 'X') },
		Key{K: "y", Shift: true}:       func(c *Context) { insertor(c, 'Y') },
		Key{K: "z", Shift: true}:       func(c *Context) { insertor(c, 'Z') },
		Key{K: "["}:                    func(c *Context) { insertor(c, '[') },
		Key{K: "]"}:                    func(c *Context) { insertor(c, ']') },
		Key{K: "[", Shift: true}:       func(c *Context) { insertor(c, '{') },
		Key{K: "]", Shift: true}:       func(c *Context) { insertor(c, '}') },
		Key{K: ";"}:                    func(c *Context) { insertor(c, ';') },
		Key{K: ";", Shift: true}:       func(c *Context) { insertor(c, ':') },
		Key{K: "'"}:                    func(c *Context) { insertor(c, '\'') },
		Key{K: "'", Shift: true}:       func(c *Context) { insertor(c, '"') },
		Key{K: "\""}:                   func(c *Context) { insertor(c, '"') },
		Key{K: ","}:                    func(c *Context) { insertor(c, ',') },
		Key{K: "."}:                    func(c *Context) { insertor(c, '.') },
		Key{K: ",", Shift: true}:       func(c *Context) { insertor(c, '<') },
		Key{K: ".", Shift: true}:       func(c *Context) { insertor(c, '>') },
		Key{K: "/"}:                    func(c *Context) { insertor(c, '/') },
		Key{K: "/", Shift: true}:       func(c *Context) { insertor(c, '?') },
		Key{K: "-"}:                    func(c *Context) { insertor(c, '-') },
		Key{K: "="}:                    func(c *Context) { insertor(c, '=') },
		Key{K: "-", Shift: true}:       func(c *Context) { insertor(c, '_') },
		Key{K: "=", Shift: true}:       func(c *Context) { insertor(c, '+') },
		Key{K: "`"}:                    func(c *Context) { insertor(c, '`') },
		Key{K: "`", Shift: true}:       func(c *Context) { insertor(c, '~') },
		Key{K: "<space>", Shift: true}: func(c *Context) { insertor(c, ' ') },
		Key{K: "<space>"}:              func(c *Context) { insertor(c, ' ') },
	}
}

func setupDefaults() {
	PromptKeymap.SetKeys(MakeInsertionKeys(func(c *Context, b byte) {
		c.Prompt.UserInput += string(b)
	}))
	PromptKeymap.BindKey(Key{K: "<enter>"}, func(c *Context) {
		c.Prompt.IsActive = false
		userInput := c.Prompt.UserInput
		c.Prompt.UserInput = ""
		doneHook := c.Prompt.DoneHook
		c.Prompt.DoneHook = nil
		doneHook(userInput, c)
	})

	PromptKeymap.BindKey(Key{K: "<backspace>"}, func(c *Context) {
		c.Prompt.UserInput = c.Prompt.UserInput[:len(c.Prompt.UserInput)-1]
	})

	PromptKeymap.BindKey(Key{K: "<esc>"}, func(c *Context) {
		c.Prompt.IsActive = false
		c.Prompt.UserInput = ""
		c.Prompt.DoneHook = nil
	})

	BufferKeymap.SetKeys(MakeInsertionKeys(func(c *Context, b byte) {
		BufferInsertChar(c.ActiveDrawable().(*BufferView), b)
	}))

	BufferKeymap.BindKey(Key{K: ".", Control: true}, MakeCommand(AnotherSelectionOnMatch))
	BufferKeymap.BindKey(Key{K: ",", Shift: true, Control: true}, MakeCommand(ScrollToTop))
	BufferKeymap.BindKey(Key{K: "l", Control: true}, MakeCommand(CentralizePoint))
	BufferKeymap.BindKey(Key{K: ";", Control: true}, MakeCommand(CompileNoAsk))
	BufferKeymap.BindKey(Key{K: ";", Control: true, Shift: true}, MakeCommand(CompileAskForCommand))
	BufferKeymap.BindKey(Key{K: "g", Alt: true}, MakeCommand(GrepAsk))
	BufferKeymap.BindKey(Key{K: ".", Shift: true, Control: true}, MakeCommand(ScrollToBottom))
	BufferKeymap.BindKey(Key{K: "<right>", Shift: true}, MakeCommand(func(e *BufferView) { MarkRight(e, 1) }))
	BufferKeymap.BindKey(Key{K: "<right>", Shift: true, Control: true}, MakeCommand(MarkNextWord))
	BufferKeymap.BindKey(Key{K: "<left>", Shift: true, Control: true}, MakeCommand(MarkPreviousWord))
	BufferKeymap.BindKey(Key{K: "<left>", Shift: true}, MakeCommand(func(e *BufferView) { MarkLeft(e, 1) }))
	BufferKeymap.BindKey(Key{K: "<up>", Shift: true}, MakeCommand(func(e *BufferView) { MarkUp(e, 1) }))
	BufferKeymap.BindKey(Key{K: "<down>", Shift: true}, MakeCommand(func(e *BufferView) { MarkDown(e, 1) }))
	BufferKeymap.BindKey(Key{K: "n", Shift: true, Control: true}, MakeCommand(func(e *BufferView) { MarkDown(e, 1) }))
	BufferKeymap.BindKey(Key{K: "p", Shift: true, Control: true}, MakeCommand(func(e *BufferView) { MarkUp(e, 1) }))
	BufferKeymap.BindKey(Key{K: "f", Shift: true, Control: true}, MakeCommand(func(e *BufferView) { MarkRight(e, 1) }))
	BufferKeymap.BindKey(Key{K: "b", Shift: true, Control: true}, MakeCommand(func(e *BufferView) { MarkLeft(e, 1) }))
	BufferKeymap.BindKey(Key{K: "a", Shift: true, Control: true}, MakeCommand(MarkToBeginningOfLine))
	BufferKeymap.BindKey(Key{K: "e", Shift: true, Control: true}, MakeCommand(MarkToEndOfLine))
	BufferKeymap.BindKey(Key{K: "5", Shift: true, Control: true}, MakeCommand(MarkToMatchingChar))
	BufferKeymap.BindKey(Key{K: "m", Shift: true, Control: true}, MakeCommand(MarkToMatchingChar))
	BufferKeymap.BindKey(Key{K: "r", Control: true}, MakeCommand(QueryReplaceActivate))
	BufferKeymap.BindKey(Key{K: "<lmouse>-click", Control: true}, MakeCommand(func(e *BufferView) {
		e.addAnotherCursorAt(rl.GetMousePosition())
	}))
	BufferKeymap.BindKey(Key{K: "<lmouse>-hold", Control: true}, MakeCommand(func(e *BufferView) {
		e.addAnotherCursorAt(rl.GetMousePosition())
	}))
	BufferKeymap.BindKey(Key{K: "<lmouse>-hold", Control: true}, MakeCommand(func(e *BufferView) {
		e.addAnotherCursorAt(rl.GetMousePosition())
	}))
	BufferKeymap.BindKey(Key{K: "<up>", Control: true}, MakeCommand(func(e *BufferView) {
		AddCursorPreviousLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "<down>", Control: true}, MakeCommand(func(e *BufferView) {
		AddCursorNextLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "r", Alt: true}, MakeCommand(func(e *BufferView) {
		e.readFileFromDisk()
	}))
	BufferKeymap.BindKey(Key{K: "z", Control: true}, MakeCommand(func(e *BufferView) {
		e.RevertLastBufferAction()
	}))
	BufferKeymap.BindKey(Key{K: "f", Control: true}, MakeCommand(func(e *BufferView) {
		PointRight(e, 1)
	}))
	BufferKeymap.BindKey(Key{K: "x", Control: true}, MakeCommand(func(e *BufferView) {
		Cut(e)
	}))
	BufferKeymap.BindKey(Key{K: "v", Control: true}, MakeCommand(func(e *BufferView) {
		Paste(e)
	}))
	BufferKeymap.BindKey(Key{K: "k", Control: true}, MakeCommand(func(e *BufferView) {
		KillLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "g", Control: true}, MakeCommand(func(e *BufferView) {
		InteractiveGotoLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "c", Control: true}, MakeCommand(func(e *BufferView) {
		Copy(e)
	}))
	BufferKeymap.BindKey(Key{K: "c", Alt: true}, MakeCommand(func(a *BufferView) {
		CompileAskForCommand(a)
	}))
	BufferKeymap.BindKey(Key{K: "s", Control: true}, MakeCommand(func(a *BufferView) {
		SearchActivate(a)
	}))
	BufferKeymap.BindKey(Key{K: "w", Control: true}, MakeCommand(func(a *BufferView) {
		Write(a)
	}))
	BufferKeymap.BindKey(Key{K: "<esc>"}, MakeCommand(func(p *BufferView) {
		RemoveAllCursorsButOne(p)
	}))

	BufferKeymap.BindKey(Key{K: "<lmouse>-click"}, MakeCommand(func(e *BufferView) {
		e.moveCursorTo(rl.GetMousePosition())
	}))

	BufferKeymap.BindKey(Key{K: "<mouse-wheel-down>"}, MakeCommand(func(e *BufferView) {
		ScrollDown(e, 5)
	}))
	BufferKeymap.BindKey(Key{K: "<mouse-wheel-up>"}, MakeCommand(func(e *BufferView) {
		ScrollUp(e, 5)
	}))
	BufferKeymap.BindKey(Key{K: "<lmouse>-hold"}, MakeCommand(func(e *BufferView) {
		e.moveCursorTo(rl.GetMousePosition())
	}))

	BufferKeymap.BindKey(Key{K: "a", Control: true}, MakeCommand(func(e *BufferView) {
		PointToBeginningOfLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "e", Control: true}, MakeCommand(func(e *BufferView) {
		PointToEndOfLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "5", Control: true}, MakeCommand(func(e *BufferView) {
		PointToMatchingChar(e)
	}))
	BufferKeymap.BindKey(Key{K: "m", Control: true}, MakeCommand(func(e *BufferView) {
		PointToMatchingChar(e)
	}))
	BufferKeymap.BindKey(Key{K: "p", Control: true}, MakeCommand(func(e *BufferView) {
		PointUp(e)
	}))
	BufferKeymap.BindKey(Key{K: "n", Control: true}, MakeCommand(func(e *BufferView) {
		PointDown(e)
	}))

	BufferKeymap.BindKey(Key{K: "<up>"}, MakeCommand(func(e *BufferView) {
		PointUp(e)
	}))
	BufferKeymap.BindKey(Key{K: "<down>"}, MakeCommand(func(e *BufferView) {
		PointDown(e)
	}))
	BufferKeymap.BindKey(Key{K: "<right>"}, MakeCommand(func(e *BufferView) {
		PointRight(e, 1)
	}))
	BufferKeymap.BindKey(Key{K: "<right>", Control: true}, MakeCommand(func(e *BufferView) {
		PointForwardWord(e)
	}))
	BufferKeymap.BindKey(Key{K: "<left>"}, MakeCommand(func(e *BufferView) {
		PointLeft(e, 1)
	}))
	BufferKeymap.BindKey(Key{K: "<left>", Control: true}, MakeCommand(func(e *BufferView) {
		PointBackwardWord(e)
	}))

	BufferKeymap.BindKey(Key{K: "b", Control: true}, MakeCommand(func(e *BufferView) {
		PointLeft(e, 1)
	}))
	BufferKeymap.BindKey(Key{K: "<home>"}, MakeCommand(func(e *BufferView) {
		PointToBeginningOfLine(e)
	}))
	BufferKeymap.BindKey(Key{K: "<pagedown>"}, MakeCommand(func(e *BufferView) {
		ScrollDown(e, 1)
	}))
	BufferKeymap.BindKey(Key{K: "<pageup>"}, MakeCommand(func(e *BufferView) {
		ScrollUp(e, 1)
	}))
	BufferKeymap.BindKey(Key{K: "<enter>"}, MakeCommand(func(e *BufferView) {
		BufferInsertChar(e, '\n')
	}))
	BufferKeymap.BindKey(Key{K: "<backspace>", Control: true}, MakeCommand(func(e *BufferView) {
		DeleteWordBackward(e)
	}))
	BufferKeymap.BindKey(Key{K: "<backspace>"}, MakeCommand(func(e *BufferView) {
		DeleteCharBackward(e)
	}))
	BufferKeymap.BindKey(Key{K: "<backspace>", Shift: true}, MakeCommand(func(e *BufferView) {
		DeleteCharBackward(e)
	}))
	BufferKeymap.BindKey(Key{K: "d", Control: true}, MakeCommand(func(e *BufferView) {
		DeleteCharForward(e)
	}))
	BufferKeymap.BindKey(Key{K: "<delete>"}, MakeCommand(func(e *BufferView) {
		DeleteCharForward(e)
	}))
	BufferKeymap.BindKey(Key{K: "<tab>"}, MakeCommand(func(e *BufferView) { Indent(e) }))


	CompileKeymap.BindKey(Key{K: "<enter>"}, BufferOpenLocationInCurrentLine)

	GlobalKeymap.BindKey(Key{K: "\\", Alt: true}, func(c *Context) { VSplit(c) })
	GlobalKeymap.BindKey(Key{K: "=", Alt: true}, func(c *Context) { HSplit(c) })
	GlobalKeymap.BindKey(Key{K: ";", Control: true}, func(c *Context) { Compile(c) })
	GlobalKeymap.BindKey(Key{K: "q", Alt: true}, func(c *Context) { c.CloseWindow(c.ActiveWindowIndex) })
	GlobalKeymap.BindKey(Key{K: "q", Alt: true, Shift: true}, Exit)
	GlobalKeymap.BindKey(Key{K: "0", Control: true}, func(c *Context) { c.CloseWindow(c.ActiveWindowIndex) })
	GlobalKeymap.BindKey(Key{K: "1", Control: true}, func(c *Context) { c.BuildWindowToggleState() })
	GlobalKeymap.BindKey(Key{K: "k", Alt: true}, func(c *Context) { c.KillDrawable(c.ActiveDrawableID()) })
	GlobalKeymap.BindKey(Key{K: "t", Alt: true}, func(c *Context) { c.OpenThemesList() })
	GlobalKeymap.BindKey(Key{K: "o", Control: true}, func(c *Context) { c.OpenFileList() })
	GlobalKeymap.BindKey(Key{K: "b", Alt: true}, func(c *Context) { c.OpenBufferList() })
	GlobalKeymap.BindKey(Key{K: "<mouse-wheel-down>", Control: true}, func(c *Context) { c.DecreaseFontSize(2) })
	GlobalKeymap.BindKey(Key{K: "<mouse-wheel-up>", Control: true}, func(c *Context) { c.IncreaseFontSize(2) })
	GlobalKeymap.BindKey(Key{K: "=", Control: true}, func(c *Context) { c.IncreaseFontSize(2) })
	GlobalKeymap.BindKey(Key{K: "-", Control: true}, func(c *Context) { c.DecreaseFontSize(2) })
	GlobalKeymap.BindKey(Key{K: "w", Alt: true}, func(c *Context) { c.OtherWindow() })

	// Search
	SearchKeymap.BindKey(Key{K: "<enter>"}, MakeCommand(func(editor *BufferView) {
		SearchNextMatch(editor)
	}))
	SearchKeymap.BindKey(Key{K: "s", Control: true}, MakeCommand(func(editor *BufferView) {
		SearchNextMatch(editor)
	}))
	SearchKeymap.BindKey(Key{K: "r", Control: true}, MakeCommand(func(editor *BufferView) {
		SearchPreviousMatch(editor)
	}))
	SearchKeymap.BindKey(Key{K: "<enter>", Control: true}, MakeCommand(func(editor *BufferView) {
		SearchPreviousMatch(editor)
	}))
	SearchKeymap.BindKey(Key{K: "<esc>"}, MakeCommand(func(editor *BufferView) {
		SearchExit(editor)
	}))
	SearchKeymap.BindKey(Key{K: "<mouse-wheel-up>"}, MakeCommand(func(e *BufferView) {
		e.Search.MovedAwayFromCurrentMatch = true
		ScrollUp(e, 30)
	}))
	SearchKeymap.BindKey(Key{K: "<mouse-wheel-down>"}, MakeCommand(func(e *BufferView) {
		e.Search.MovedAwayFromCurrentMatch = true
		ScrollDown(e, 30)
	}))
	SearchKeymap.BindKey(Key{K: "<rmouse>-click"}, MakeCommand(func(editor *BufferView) {
		SearchNextMatch(editor)
	}))
	SearchKeymap.BindKey(Key{K: "<mmouse>-click"}, MakeCommand(func(editor *BufferView) {
		SearchPreviousMatch(editor)
	}))
	SearchKeymap.BindKey(Key{K: "<pagedown>"}, MakeCommand(func(e *BufferView) {
		e.Search.MovedAwayFromCurrentMatch = true
		ScrollDown(e, 1)
	}))
	SearchKeymap.BindKey(Key{K: "<pageup>"}, MakeCommand(func(e *BufferView) {
		e.Search.MovedAwayFromCurrentMatch = true
		ScrollUp(e, 1)
	}))


	// Query replace
	QueryReplaceKeymap.BindKey(Key{K: "r", Control: true}, MakeCommand(func(editor *BufferView) {
		QueryReplaceReplaceThisMatch(editor)
	}))

	QueryReplaceKeymap.BindKey(Key{K: "<enter>"}, MakeCommand(func(editor *BufferView) {
		QueryReplaceReplaceThisMatch(editor)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "y"}, MakeCommand(func(editor *BufferView) {
		QueryReplaceReplaceThisMatch(editor)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<enter>", Control: true}, MakeCommand(func(editor *BufferView) {
		QueryReplaceIgnoreThisMatch(editor)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<esc>"}, MakeCommand(func(editor *BufferView) {
		QueryReplaceExit(editor)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<lmouse>-click"}, MakeCommand(func(e *BufferView) {
		e.moveCursorTo(rl.GetMousePosition())
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<mouse-wheel-up>"}, MakeCommand(func(e *BufferView) {
		e.QueryReplace.MovedAwayFromCurrentMatch = true
		ScrollUp(e, 30)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<mouse-wheel-down>"}, MakeCommand(func(e *BufferView) {
		e.QueryReplace.MovedAwayFromCurrentMatch = true
		ScrollDown(e, 30)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<pagedown>"}, MakeCommand(func(e *BufferView) {
		e.QueryReplace.MovedAwayFromCurrentMatch = true
		ScrollDown(e, 1)
	}))
	QueryReplaceKeymap.BindKey(Key{K: "<pageup>"}, MakeCommand(func(e *BufferView) {
		e.QueryReplace.MovedAwayFromCurrentMatch = true
		ScrollUp(e, 1)
	}))

}
