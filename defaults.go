package preditor

import rl "github.com/gen2brain/raylib-go/raylib"

func setupDefaults() {
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
			return InsertChar(e, '\n')
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
