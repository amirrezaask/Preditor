package preditor

import rl "github.com/gen2brain/raylib-go/raylib"

var GlobalKeymap = Keymap{
	Key{K: "\\", Alt: true}: func(c *Context) error {
		c.VSplit()

		return nil
	},
	Key{K: "=", Alt: true}: func(c *Context) error {
		c.HSplit()

		return nil
	},
	Key{K: ";", Control: true}: func(context *Context) error {
		Compile(context)
		return nil
	},
	Key{K: "q", Alt: true}: func(preditor *Context) error {
		preditor.CloseWindow(preditor.ActiveWindowIndex)

		return nil
	},
	Key{K: "q", Alt: true, Shift: true}: func(preditor *Context) error {
		preditor.Exit()
		return nil
	},
	Key{K: "0", Control: true}: func(preditor *Context) error {
		preditor.CloseWindow(preditor.ActiveWindowIndex)

		return nil
	},
	Key{K: "1", Control: true}: func(preditor *Context) error {
		preditor.BuildWindowToggleState()

		return nil
	},
	Key{K: "k", Alt: true}: func(preditor *Context) error {
		preditor.KillDrawable(preditor.ActiveDrawableID())

		return nil
	},
	Key{K: "t", Alt: true}: func(preditor *Context) error {
		preditor.openThemeSwitcher()
		return nil
	},

	Key{K: "o", Control: true}: func(a *Context) error {
		a.openFileBuffer()
		return nil
	},
	Key{K: "b", Alt: true}: func(a *Context) error {
		a.openBufferSwitcher()

		return nil
	},

	Key{K: "<mouse-wheel-down>", Control: true}: func(c *Context) error {
		c.DecreaseFontSize(2)

		return nil
	},

	Key{K: "<mouse-wheel-up>", Control: true}: func(c *Context) error {
		c.IncreaseFontSize(2)
		return nil

	},
	Key{K: "=", Control: true}: func(e *Context) error {
		e.IncreaseFontSize(2)

		return nil
	},
	Key{K: "-", Control: true}: func(e *Context) error {
		e.DecreaseFontSize(2)

		return nil
	},
	Key{K: "w", Alt: true}: func(c *Context) error {
		c.OtherWindow()

		return nil
	},
	Key{K: "<left>", Alt: true}: func(c *Context) error {
		//c.SwitchPrevBuffer()
		return nil
	},
	Key{K: "<right>", Alt: true}: func(c *Context) error {
		//c.SwitchNextBuffer()
		return nil
	},
}

func MakeInsertionKeys(insertor func(c *Context, b byte) error) Keymap {
	return Keymap{
		Key{K: "a"}:                    func(c *Context) error { return insertor(c, 'a') },
		Key{K: "b"}:                    func(c *Context) error { return insertor(c, 'b') },
		Key{K: "c"}:                    func(c *Context) error { return insertor(c, 'c') },
		Key{K: "d"}:                    func(c *Context) error { return insertor(c, 'd') },
		Key{K: "e"}:                    func(c *Context) error { return insertor(c, 'e') },
		Key{K: "f"}:                    func(c *Context) error { return insertor(c, 'f') },
		Key{K: "g"}:                    func(c *Context) error { return insertor(c, 'g') },
		Key{K: "h"}:                    func(c *Context) error { return insertor(c, 'h') },
		Key{K: "i"}:                    func(c *Context) error { return insertor(c, 'i') },
		Key{K: "j"}:                    func(c *Context) error { return insertor(c, 'j') },
		Key{K: "k"}:                    func(c *Context) error { return insertor(c, 'k') },
		Key{K: "l"}:                    func(c *Context) error { return insertor(c, 'l') },
		Key{K: "m"}:                    func(c *Context) error { return insertor(c, 'm') },
		Key{K: "n"}:                    func(c *Context) error { return insertor(c, 'n') },
		Key{K: "o"}:                    func(c *Context) error { return insertor(c, 'o') },
		Key{K: "p"}:                    func(c *Context) error { return insertor(c, 'p') },
		Key{K: "q"}:                    func(c *Context) error { return insertor(c, 'q') },
		Key{K: "r"}:                    func(c *Context) error { return insertor(c, 'r') },
		Key{K: "s"}:                    func(c *Context) error { return insertor(c, 's') },
		Key{K: "t"}:                    func(c *Context) error { return insertor(c, 't') },
		Key{K: "u"}:                    func(c *Context) error { return insertor(c, 'u') },
		Key{K: "v"}:                    func(c *Context) error { return insertor(c, 'v') },
		Key{K: "w"}:                    func(c *Context) error { return insertor(c, 'w') },
		Key{K: "x"}:                    func(c *Context) error { return insertor(c, 'x') },
		Key{K: "y"}:                    func(c *Context) error { return insertor(c, 'y') },
		Key{K: "z"}:                    func(c *Context) error { return insertor(c, 'z') },
		Key{K: "0"}:                    func(c *Context) error { return insertor(c, '0') },
		Key{K: "1"}:                    func(c *Context) error { return insertor(c, '1') },
		Key{K: "2"}:                    func(c *Context) error { return insertor(c, '2') },
		Key{K: "3"}:                    func(c *Context) error { return insertor(c, '3') },
		Key{K: "4"}:                    func(c *Context) error { return insertor(c, '4') },
		Key{K: "5"}:                    func(c *Context) error { return insertor(c, '5') },
		Key{K: "6"}:                    func(c *Context) error { return insertor(c, '6') },
		Key{K: "7"}:                    func(c *Context) error { return insertor(c, '7') },
		Key{K: "8"}:                    func(c *Context) error { return insertor(c, '8') },
		Key{K: "9"}:                    func(c *Context) error { return insertor(c, '9') },
		Key{K: "\\"}:                   func(c *Context) error { return insertor(c, '\\') },
		Key{K: "\\", Shift: true}:      func(c *Context) error { return insertor(c, '|') },
		Key{K: "0", Shift: true}:       func(c *Context) error { return insertor(c, ')') },
		Key{K: "1", Shift: true}:       func(c *Context) error { return insertor(c, '!') },
		Key{K: "2", Shift: true}:       func(c *Context) error { return insertor(c, '@') },
		Key{K: "3", Shift: true}:       func(c *Context) error { return insertor(c, '#') },
		Key{K: "4", Shift: true}:       func(c *Context) error { return insertor(c, '$') },
		Key{K: "5", Shift: true}:       func(c *Context) error { return insertor(c, '%') },
		Key{K: "6", Shift: true}:       func(c *Context) error { return insertor(c, '^') },
		Key{K: "7", Shift: true}:       func(c *Context) error { return insertor(c, '&') },
		Key{K: "8", Shift: true}:       func(c *Context) error { return insertor(c, '*') },
		Key{K: "9", Shift: true}:       func(c *Context) error { return insertor(c, '(') },
		Key{K: "a", Shift: true}:       func(c *Context) error { return insertor(c, 'A') },
		Key{K: "b", Shift: true}:       func(c *Context) error { return insertor(c, 'B') },
		Key{K: "c", Shift: true}:       func(c *Context) error { return insertor(c, 'C') },
		Key{K: "d", Shift: true}:       func(c *Context) error { return insertor(c, 'D') },
		Key{K: "e", Shift: true}:       func(c *Context) error { return insertor(c, 'E') },
		Key{K: "f", Shift: true}:       func(c *Context) error { return insertor(c, 'F') },
		Key{K: "g", Shift: true}:       func(c *Context) error { return insertor(c, 'G') },
		Key{K: "h", Shift: true}:       func(c *Context) error { return insertor(c, 'H') },
		Key{K: "i", Shift: true}:       func(c *Context) error { return insertor(c, 'I') },
		Key{K: "j", Shift: true}:       func(c *Context) error { return insertor(c, 'J') },
		Key{K: "k", Shift: true}:       func(c *Context) error { return insertor(c, 'K') },
		Key{K: "l", Shift: true}:       func(c *Context) error { return insertor(c, 'L') },
		Key{K: "m", Shift: true}:       func(c *Context) error { return insertor(c, 'M') },
		Key{K: "n", Shift: true}:       func(c *Context) error { return insertor(c, 'N') },
		Key{K: "o", Shift: true}:       func(c *Context) error { return insertor(c, 'O') },
		Key{K: "p", Shift: true}:       func(c *Context) error { return insertor(c, 'P') },
		Key{K: "q", Shift: true}:       func(c *Context) error { return insertor(c, 'Q') },
		Key{K: "r", Shift: true}:       func(c *Context) error { return insertor(c, 'R') },
		Key{K: "s", Shift: true}:       func(c *Context) error { return insertor(c, 'S') },
		Key{K: "t", Shift: true}:       func(c *Context) error { return insertor(c, 'T') },
		Key{K: "u", Shift: true}:       func(c *Context) error { return insertor(c, 'U') },
		Key{K: "v", Shift: true}:       func(c *Context) error { return insertor(c, 'V') },
		Key{K: "w", Shift: true}:       func(c *Context) error { return insertor(c, 'W') },
		Key{K: "x", Shift: true}:       func(c *Context) error { return insertor(c, 'X') },
		Key{K: "y", Shift: true}:       func(c *Context) error { return insertor(c, 'Y') },
		Key{K: "z", Shift: true}:       func(c *Context) error { return insertor(c, 'Z') },
		Key{K: "["}:                    func(c *Context) error { return insertor(c, '[') },
		Key{K: "]"}:                    func(c *Context) error { return insertor(c, ']') },
		Key{K: "[", Shift: true}:       func(c *Context) error { return insertor(c, '{') },
		Key{K: "]", Shift: true}:       func(c *Context) error { return insertor(c, '}') },
		Key{K: ";"}:                    func(c *Context) error { return insertor(c, ';') },
		Key{K: ";", Shift: true}:       func(c *Context) error { return insertor(c, ':') },
		Key{K: "'"}:                    func(c *Context) error { return insertor(c, '\'') },
		Key{K: "'", Shift: true}:       func(c *Context) error { return insertor(c, '"') },
		Key{K: "\""}:                   func(c *Context) error { return insertor(c, '"') },
		Key{K: ","}:                    func(c *Context) error { return insertor(c, ',') },
		Key{K: "."}:                    func(c *Context) error { return insertor(c, '.') },
		Key{K: ",", Shift: true}:       func(c *Context) error { return insertor(c, '<') },
		Key{K: ".", Shift: true}:       func(c *Context) error { return insertor(c, '>') },
		Key{K: "/"}:                    func(c *Context) error { return insertor(c, '/') },
		Key{K: "/", Shift: true}:       func(c *Context) error { return insertor(c, '?') },
		Key{K: "-"}:                    func(c *Context) error { return insertor(c, '-') },
		Key{K: "="}:                    func(c *Context) error { return insertor(c, '=') },
		Key{K: "-", Shift: true}:       func(c *Context) error { return insertor(c, '_') },
		Key{K: "=", Shift: true}:       func(c *Context) error { return insertor(c, '+') },
		Key{K: "`"}:                    func(c *Context) error { return insertor(c, '`') },
		Key{K: "`", Shift: true}:       func(c *Context) error { return insertor(c, '~') },
		Key{K: "<space>", Shift: true}: func(c *Context) error { return insertor(c, ' ') },
		Key{K: "<space>"}:              func(c *Context) error { return insertor(c, ' ') },
	}
}

func MergeKeymaps(k1 Keymap, k2 Keymap) Keymap {
	for k, v := range k2 {
		k1[k] = v
	}
	return k1
}

var DefaultPromptKeymap Keymap

func setupDefaults() {
	DefaultPromptKeymap = MergeKeymaps(Keymap{
		Key{K: "<enter>"}: func(c *Context) error {
			c.Prompt.IsActive = false
			userInput := c.Prompt.UserInput
			c.Prompt.UserInput = ""
			c.Prompt.DoneHook(userInput, c)
			c.Prompt.DoneHook = nil
			c.Prompt.ChangeHook = nil
			return nil
		},

		Key{K: "<backspace>"}: func(c *Context) error {
			c.Prompt.UserInput = c.Prompt.UserInput[:len(c.Prompt.UserInput)-1]

			return nil
		},
		Key{K: "<esc>"}: func(c *Context) error {
			c.Prompt.IsActive = false
			c.Prompt.UserInput = ""
			c.Prompt.DoneHook = nil
			c.Prompt.ChangeHook = nil

			return nil
		},
	}, MakeInsertionKeys(func(c *Context, b byte) error {
		c.Prompt.UserInput += string(b)
		return nil
	}))
	EditorKeymap = Keymap{
		Key{K: ".", Control: true}: MakeCommand(func(e *BufferView) error {
			return e.AnotherSelectionOnMatch()
		}),
		Key{K: ",", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			ScrollToTop(e)

			return nil
		}),
		Key{K: "l", Control: true}: MakeCommand(func(e *BufferView) error {
			CentralizePoint(e)

			return nil
		}),
		Key{K: ";", Control: true}: MakeCommand(func(editor *BufferView) error {
			return CompileNoAsk(editor)
		}),
		Key{K: ";", Control: true, Shift: true}: MakeCommand(func(editor *BufferView) error {
			return CompileAskForCommand(editor)
		}),

		Key{K: "g", Alt: true}: MakeCommand(func(t *BufferView) error {
			return GrepAsk(t)
		}),
		Key{K: ".", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			ScrollToBottom(e)

			return nil
		}),
		Key{K: "<right>", Shift: true}: MakeCommand(func(e *BufferView) error {
			MarkRight(e, 1)

			return nil
		}),
		Key{K: "<right>", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkNextWord(e)

			return nil
		}),
		Key{K: "<left>", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkPreviousWord(e)

			return nil
		}),
		Key{K: "<left>", Shift: true}: MakeCommand(func(e *BufferView) error {
			MarkLeft(e, 1)

			return nil
		}),
		Key{K: "<up>", Shift: true}: MakeCommand(func(e *BufferView) error {
			MarkUp(e, 1)

			return nil
		}),
		Key{K: "<down>", Shift: true}: MakeCommand(func(e *BufferView) error {
			MarkDown(e, 1)

			return nil
		}),
		Key{K: "a", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkToBeginningOfLine(e)

			return nil
		}),
		Key{K: "e", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkToEndOfLine(e)

			return nil
		}),
		Key{K: "n", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkDown(e, 1)

			return nil
		}),
		Key{K: "p", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkUp(e, 1)

			return nil
		}),
		Key{K: "f", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkRight(e, 1)

			return nil
		}),
		Key{K: "b", Shift: true, Control: true}: MakeCommand(func(e *BufferView) error {
			MarkLeft(e, 1)

			return nil
		}),
		Key{K: "5", Control: true, Shift: true}: MakeCommand(func(e *BufferView) error {
			return MarkToMatchingChar(e)
		}),
		Key{K: "m", Control: true, Shift: true}: MakeCommand(func(e *BufferView) error {
			return MarkToMatchingChar(e)
		}),
		Key{K: "<lmouse>-click", Control: true}: MakeCommand(func(e *BufferView) error {
			return e.addAnotherCursorAt(rl.GetMousePosition())
		}),
		Key{K: "<lmouse>-hold", Control: true}: MakeCommand(func(e *BufferView) error {
			return e.addAnotherCursorAt(rl.GetMousePosition())
		}),
		Key{K: "<up>", Control: true}: MakeCommand(func(e *BufferView) error {
			return AddCursorPreviousLine(e)
		}),

		Key{K: "<down>", Control: true}: MakeCommand(func(e *BufferView) error {
			return AddCursorNextLine(e)
		}),
		Key{K: "r", Alt: true}: MakeCommand(func(e *BufferView) error {
			return e.readFileFromDisk()
		}),
		Key{K: "z", Control: true}: MakeCommand(func(e *BufferView) error {
			e.RevertLastBufferAction()
			return nil
		}),
		Key{K: "f", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointRight(e, 1)
		}),
		Key{K: "x", Control: true}: MakeCommand(func(e *BufferView) error {
			return Cut(e)
		}),
		Key{K: "v", Control: true}: MakeCommand(func(e *BufferView) error {
			return Paste(e)
		}),
		Key{K: "k", Control: true}: MakeCommand(func(e *BufferView) error {
			return KillLine(e)
		}),
		Key{K: "g", Control: true}: MakeCommand(func(e *BufferView) error {
			return InteractiveGotoLine(e)
		}),
		Key{K: "c", Control: true}: MakeCommand(func(e *BufferView) error {
			return Copy(e)
		}),

		Key{K: "c", Alt: true}: MakeCommand(func(a *BufferView) error {
			return CompileAskForCommand(a)
		}),
		Key{K: "s", Control: true}: MakeCommand(func(a *BufferView) error {
			return ISearchActivate(a)
		}),
		Key{K: "w", Control: true}: MakeCommand(func(a *BufferView) error {
			return Write(a)
		}),
		Key{K: "<esc>"}: MakeCommand(func(p *BufferView) error {
			return RemoveAllCursorsButOne(p)
		}),

		// navigation
		Key{K: "<lmouse>-click"}: MakeCommand(func(e *BufferView) error {
			return e.moveCursorTo(rl.GetMousePosition())
		}),

		Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *BufferView) error {
			return ScrollDown(e, 5)
		}),

		Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *BufferView) error {
			return ScrollUp(e, 5)
		}),

		Key{K: "<lmouse>-hold"}: MakeCommand(func(e *BufferView) error {
			return e.moveCursorTo(rl.GetMousePosition())
		}),

		Key{K: "a", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointToBeginningOfLine(e)
		}),
		Key{K: "e", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointToEndOfLine(e)
		}),
		Key{K: "5", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointToMatchingChar(e)
		}),
		Key{K: "m", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointToMatchingChar(e)
		}),
		Key{K: "p", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointUp(e)
		}),

		Key{K: "n", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointDown(e)
		}),

		Key{K: "<up>"}: MakeCommand(func(e *BufferView) error {
			return PointUp(e)
		}),
		Key{K: "<down>"}: MakeCommand(func(e *BufferView) error {
			return PointDown(e)
		}),
		Key{K: "<right>"}: MakeCommand(func(e *BufferView) error {
			return PointRight(e, 1)
		}),
		Key{K: "<right>", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointForwardWord(e, 1)
		}),
		Key{K: "<left>"}: MakeCommand(func(e *BufferView) error {
			return PointLeft(e, 1)
		}),
		Key{K: "<left>", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointBackwardWord(e, 1)
		}),

		Key{K: "b", Control: true}: MakeCommand(func(e *BufferView) error {
			return PointLeft(e, 1)
		}),
		Key{K: "<home>"}: MakeCommand(func(e *BufferView) error {
			return PointToBeginningOfLine(e)
		}),
		Key{K: "<pagedown>"}: MakeCommand(func(e *BufferView) error {
			return ScrollDown(e, 1)
		}),
		Key{K: "<pageup>"}: MakeCommand(func(e *BufferView) error {
			return ScrollUp(e, 1)
		}),

		//insertion
		Key{K: "<enter>"}: MakeCommand(func(e *BufferView) error {
			return BufferInsertChar(e, '\n')
		}),
		Key{K: "<backspace>", Control: true}: MakeCommand(func(e *BufferView) error {
			DeleteWordBackward(e)
			return nil
		}),
		Key{K: "<backspace>"}: MakeCommand(func(e *BufferView) error {
			return DeleteCharBackward(e)
		}),
		Key{K: "<backspace>", Shift: true}: MakeCommand(func(e *BufferView) error {
			return DeleteCharBackward(e)
		}),

		Key{K: "d", Control: true}: MakeCommand(func(e *BufferView) error {
			return DeleteCharForward(e)
		}),
		Key{K: "<delete>"}: MakeCommand(func(e *BufferView) error {
			return DeleteCharForward(e)
		}),

		Key{K: "<tab>"}: MakeCommand(func(e *BufferView) error { return Indent(e) }),
	}

	SearchTextBufferKeymap = Keymap{
		Key{K: "<backspace>"}: MakeCommand(func(e *BufferView) error {
			return ISearchDeleteBackward(e)
		}),
		Key{K: "<enter>"}: MakeCommand(func(editor *BufferView) error {
			return ISearchNextMatch(editor)
		}),
		Key{K: "s", Control: true}: MakeCommand(func(editor *BufferView) error {
			return ISearchNextMatch(editor)
		}),
		Key{K: "r", Control: true}: MakeCommand(func(editor *BufferView) error {
			return ISearchPreviousMatch(editor)
		}),
		Key{K: "<enter>", Control: true}: MakeCommand(func(editor *BufferView) error {
			return ISearchPreviousMatch(editor)
		}),
		Key{K: "<esc>"}: MakeCommand(func(editor *BufferView) error {
			ISearchExit(editor)
			return nil
		}),
		Key{K: "<lmouse>-click"}: MakeCommand(func(e *BufferView) error {
			return e.moveCursorTo(rl.GetMousePosition())
		}),
		Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *BufferView) error {
			e.ISearch.MovedAwayFromCurrentMatch = true
			return ScrollUp(e, 30)

		}),
		Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *BufferView) error {
			e.ISearch.MovedAwayFromCurrentMatch = true

			return ScrollDown(e, 30)
		}),

		Key{K: "<rmouse>-click"}: MakeCommand(func(editor *BufferView) error {
			ISearchNextMatch(editor)

			return nil
		}),
		Key{K: "<mmouse>-click"}: MakeCommand(func(editor *BufferView) error {
			ISearchPreviousMatch(editor)

			return nil
		}),
		Key{K: "<pagedown>"}: MakeCommand(func(e *BufferView) error {
			e.ISearch.MovedAwayFromCurrentMatch = true
			return ScrollDown(e, 1)
		}),
		Key{K: "<pageup>"}: MakeCommand(func(e *BufferView) error {
			e.ISearch.MovedAwayFromCurrentMatch = true

			return ScrollUp(e, 1)
		}),
	}

}
