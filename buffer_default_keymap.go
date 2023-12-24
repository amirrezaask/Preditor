package preditor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var SearchTextBufferKeymap = Keymap{
	Key{K: "<backspace>"}: MakeCommand(func(e *Buffer) error {
		return ISearchDeleteBackward(e)
	}),
	Key{K: "<enter>"}: MakeCommand(func(editor *Buffer) error {
		return ISearchNextMatch(editor)
	}),
	Key{K: "s", Control: true}: MakeCommand(func(editor *Buffer) error {
		return ISearchNextMatch(editor)
	}),
	Key{K: "r", Control: true}: MakeCommand(func(editor *Buffer) error {
		return ISearchPreviousMatch(editor)
	}),

	Key{K: "<enter>", Control: true}: MakeCommand(func(editor *Buffer) error {
		return ISearchPreviousMatch(editor)
	}),
	Key{K: "<esc>"}: MakeCommand(func(editor *Buffer) error {
		ISearchExit(editor)
		return nil
	}),
	Key{K: "<lmouse>-click"}: MakeCommand(func(e *Buffer) error {
		return e.moveCursorTo(rl.GetMousePosition())
	}),
	Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *Buffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true
		return ScrollUp(e, 20)

	}),
	Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *Buffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true

		return ScrollDown(e, 20)
	}),

	Key{K: "<rmouse>-click"}: MakeCommand(func(editor *Buffer) error {
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
	Key{K: "<mmouse>-click"}: MakeCommand(func(editor *Buffer) error {
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
	Key{K: "<pagedown>"}: MakeCommand(func(e *Buffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true
		return ScrollDown(e, 1)
	}),
	Key{K: "<pageup>"}: MakeCommand(func(e *Buffer) error {
		e.ISearch.MovedAwayFromCurrentMatch = true

		return ScrollUp(e, 1)
	}),
}
var EditorKeymap Keymap

func init() {

	EditorKeymap = Keymap{

		Key{K: ".", Control: true}: MakeCommand(func(e *Buffer) error {
			return e.AnotherSelectionOnMatch()
		}),
		Key{K: ",", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			ScrollToTop(e)

			return nil
		}),
		Key{K: "l", Control: true}: MakeCommand(func(e *Buffer) error {
			CentralizePoint(e)

			return nil
		}),
		Key{K: "g", Alt: true}: MakeCommand(func(t *Buffer) error {
			return GrepAsk(t)
		}),
		Key{K: ".", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			ScrollToBottom(e)

			return nil
		}),
		Key{K: "<right>", Shift: true}: MakeCommand(func(e *Buffer) error {
			MarkRight(e, 1)

			return nil
		}),
		Key{K: "<right>", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkNextWord(e)

			return nil
		}),
		Key{K: "<left>", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkPreviousWord(e)

			return nil
		}),
		Key{K: "<left>", Shift: true}: MakeCommand(func(e *Buffer) error {
			MarkLeft(e, 1)

			return nil
		}),
		Key{K: "<up>", Shift: true}: MakeCommand(func(e *Buffer) error {
			MarkUp(e, 1)

			return nil
		}),
		Key{K: "<down>", Shift: true}: MakeCommand(func(e *Buffer) error {
			MarkDown(e, 1)

			return nil
		}),
		Key{K: "a", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkToBeginningOfLine(e)

			return nil
		}),
		Key{K: "e", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkToEndOfLine(e)

			return nil
		}),
		Key{K: "n", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkDown(e, 1)

			return nil
		}),
		Key{K: "p", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkUp(e, 1)

			return nil
		}),
		Key{K: "f", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkRight(e, 1)

			return nil
		}),
		Key{K: "b", Shift: true, Control: true}: MakeCommand(func(e *Buffer) error {
			MarkLeft(e, 1)

			return nil
		}),
		Key{K: "<lmouse>-click", Control: true}: MakeCommand(func(e *Buffer) error {
			return e.addAnotherCursorAt(rl.GetMousePosition())
		}),
		Key{K: "<lmouse>-hold", Control: true}: MakeCommand(func(e *Buffer) error {
			return e.addAnotherCursorAt(rl.GetMousePosition())
		}),
		Key{K: "<up>", Control: true}: MakeCommand(func(e *Buffer) error {
			return AddCursorPreviousLine(e)
		}),

		Key{K: "<down>", Control: true}: MakeCommand(func(e *Buffer) error {
			return AddCursorNextLine(e)
		}),
		Key{K: "r", Alt: true}: MakeCommand(func(e *Buffer) error {
			return e.readFileFromDisk()
		}),
		Key{K: "z", Control: true}: MakeCommand(func(e *Buffer) error {
			e.PopAndReverseLastAction()
			return nil
		}),
		Key{K: "f", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointRight(e, 1)
		}),
		Key{K: "x", Control: true}: MakeCommand(func(e *Buffer) error {
			return Cut(e)
		}),
		Key{K: "v", Control: true}: MakeCommand(func(e *Buffer) error {
			return Paste(e)
		}),
		Key{K: "k", Control: true}: MakeCommand(func(e *Buffer) error {
			return KillLine(e)
		}),
		Key{K: "g", Control: true}: MakeCommand(func(e *Buffer) error {
			return InteractiveGotoLine(e)
		}),
		Key{K: "c", Control: true}: MakeCommand(func(e *Buffer) error {
			return Copy(e)
		}),

		Key{K: "c", Alt: true}: MakeCommand(func(a *Buffer) error {
			return CompileAskForCommand(a)
		}),
		Key{K: "s", Control: true}: MakeCommand(func(a *Buffer) error {
			return ISearchActivate(a)
		}),
		Key{K: "w", Control: true}: MakeCommand(func(a *Buffer) error {
			return Write(a)
		}),
		Key{K: "<esc>"}: MakeCommand(func(p *Buffer) error {
			return RemoveAllCursorsButOne(p)
		}),

		// navigation
		Key{K: "<lmouse>-click"}: MakeCommand(func(e *Buffer) error {
			return e.moveCursorTo(rl.GetMousePosition())
		}),

		Key{K: "<mouse-wheel-down>"}: MakeCommand(func(e *Buffer) error {
			return ScrollDown(e, 5)
		}),

		Key{K: "<mouse-wheel-up>"}: MakeCommand(func(e *Buffer) error {
			return ScrollUp(e, 5)
		}),

		Key{K: "<lmouse>-hold"}: MakeCommand(func(e *Buffer) error {
			return e.moveCursorTo(rl.GetMousePosition())
		}),

		Key{K: "a", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointToBeginningOfLine(e)
		}),
		Key{K: "e", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointToEndOfLine(e)
		}),

		Key{K: "p", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointUp(e)
		}),

		Key{K: "n", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointDown(e)
		}),

		Key{K: "<up>"}: MakeCommand(func(e *Buffer) error {
			return PointUp(e)
		}),
		Key{K: "<down>"}: MakeCommand(func(e *Buffer) error {
			return PointDown(e)
		}),
		Key{K: "<right>"}: MakeCommand(func(e *Buffer) error {
			return PointRight(e, 1)
		}),
		Key{K: "<right>", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointForwardWord(e, 1)
		}),
		Key{K: "<left>"}: MakeCommand(func(e *Buffer) error {
			return PointLeft(e, 1)
		}),
		Key{K: "<left>", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointBackwardWord(e, 1)
		}),

		Key{K: "b", Control: true}: MakeCommand(func(e *Buffer) error {
			return PointLeft(e, 1)
		}),
		Key{K: "<home>"}: MakeCommand(func(e *Buffer) error {
			return PointToBeginningOfLine(e)
		}),
		Key{K: "<pagedown>"}: MakeCommand(func(e *Buffer) error {
			return ScrollDown(e, 1)
		}),
		Key{K: "<pageup>"}: MakeCommand(func(e *Buffer) error {
			return ScrollUp(e, 1)
		}),

		//insertion
		Key{K: "<enter>"}: MakeCommand(func(e *Buffer) error {
			return InsertChar(e, '\n')
		}),
		Key{K: "<backspace>", Control: true}: MakeCommand(func(e *Buffer) error {
			DeleteWordBackward(e)
			return nil
		}),
		Key{K: "<backspace>"}: MakeCommand(func(e *Buffer) error {
			return DeleteCharBackward(e)
		}),
		Key{K: "<backspace>", Shift: true}: MakeCommand(func(e *Buffer) error {
			return DeleteCharBackward(e)
		}),

		Key{K: "d", Control: true}: MakeCommand(func(e *Buffer) error {
			return DeleteCharForward(e)
		}),
		Key{K: "<delete>"}: MakeCommand(func(e *Buffer) error {
			return DeleteCharForward(e)
		}),

		Key{K: "<tab>"}: MakeCommand(func(e *Buffer) error { return Indent(e) }),
	}
}
