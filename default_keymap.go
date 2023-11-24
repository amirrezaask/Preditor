package main

import "fmt"

var defaultKeymap = Keymap{

	// navigation
	Key{K: "a", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).BeginingOfTheLine()
		}

		return nil
	},

	Key{K: "<up>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).CursorUp()
		}

		return nil
	},
	Key{K: "<down>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).CursorDown()
		}

		return nil
	},
	Key{K: "<right>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).CursorRight()
		}

		return nil

	},
	Key{K: "<left>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).CursorLeft()
		}

		return nil
	},
	Key{K: "<home>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).BeginingOfTheLine()
		}

		return nil
	},
	Key{K: "<pagedown>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).ScrollDown(1)
		}

		return nil
	},
	Key{K: "<pageup>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).ScrollUp(1)
		}

		return nil
	},

	//insertion
	Key{K: "<enter>"}: func(e *Editor) error { return insertCharAtCursor(e, '\n') },
	Key{K: "<space>"}: func(e *Editor) error { return insertCharAtCursor(e, ' ') },
	Key{K: "<backspace>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).DeleteCharBackward()
		}

		return nil
	},
	Key{K: "d", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).DeleteCharForeward()
		}

		return nil
	},
	Key{K: "d", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).DeleteCharForeward()
		}

		return nil
	},
	Key{K: "<delete>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).DeleteCharForeward()
		}

		return nil
	},
	Key{K: "a"}:  func(e *Editor) error { return insertCharAtCursor(e, 'a') },
	Key{K: "b"}:  func(e *Editor) error { return insertCharAtCursor(e, 'b') },
	Key{K: "c"}:  func(e *Editor) error { return insertCharAtCursor(e, 'c') },
	Key{K: "d"}:  func(e *Editor) error { return insertCharAtCursor(e, 'd') },
	Key{K: "e"}:  func(e *Editor) error { return insertCharAtCursor(e, 'e') },
	Key{K: "f"}:  func(e *Editor) error { return insertCharAtCursor(e, 'f') },
	Key{K: "g"}:  func(e *Editor) error { return insertCharAtCursor(e, 'g') },
	Key{K: "h"}:  func(e *Editor) error { return insertCharAtCursor(e, 'h') },
	Key{K: "i"}:  func(e *Editor) error { return insertCharAtCursor(e, 'i') },
	Key{K: "j"}:  func(e *Editor) error { return insertCharAtCursor(e, 'j') },
	Key{K: "k"}:  func(e *Editor) error { return insertCharAtCursor(e, 'k') },
	Key{K: "l"}:  func(e *Editor) error { return insertCharAtCursor(e, 'l') },
	Key{K: "m"}:  func(e *Editor) error { return insertCharAtCursor(e, 'm') },
	Key{K: "n"}:  func(e *Editor) error { return insertCharAtCursor(e, 'n') },
	Key{K: "o"}:  func(e *Editor) error { return insertCharAtCursor(e, 'o') },
	Key{K: "p"}:  func(e *Editor) error { return insertCharAtCursor(e, 'p') },
	Key{K: "q"}:  func(e *Editor) error { return insertCharAtCursor(e, 'q') },
	Key{K: "r"}:  func(e *Editor) error { return insertCharAtCursor(e, 'r') },
	Key{K: "s"}:  func(e *Editor) error { return insertCharAtCursor(e, 's') },
	Key{K: "t"}:  func(e *Editor) error { return insertCharAtCursor(e, 't') },
	Key{K: "u"}:  func(e *Editor) error { return insertCharAtCursor(e, 'u') },
	Key{K: "v"}:  func(e *Editor) error { return insertCharAtCursor(e, 'v') },
	Key{K: "w"}:  func(e *Editor) error { return insertCharAtCursor(e, 'w') },
	Key{K: "x"}:  func(e *Editor) error { return insertCharAtCursor(e, 'x') },
	Key{K: "y"}:  func(e *Editor) error { return insertCharAtCursor(e, 'y') },
	Key{K: "z"}:  func(e *Editor) error { return insertCharAtCursor(e, 'z') },
	Key{K: "0"}:  func(e *Editor) error { return insertCharAtCursor(e, '0') },
	Key{K: "1"}:  func(e *Editor) error { return insertCharAtCursor(e, '1') },
	Key{K: "2"}:  func(e *Editor) error { return insertCharAtCursor(e, '2') },
	Key{K: "3"}:  func(e *Editor) error { return insertCharAtCursor(e, '3') },
	Key{K: "4"}:  func(e *Editor) error { return insertCharAtCursor(e, '4') },
	Key{K: "5"}:  func(e *Editor) error { return insertCharAtCursor(e, '5') },
	Key{K: "6"}:  func(e *Editor) error { return insertCharAtCursor(e, '6') },
	Key{K: "7"}:  func(e *Editor) error { return insertCharAtCursor(e, '7') },
	Key{K: "8"}:  func(e *Editor) error { return insertCharAtCursor(e, '8') },
	Key{K: "9"}:  func(e *Editor) error { return insertCharAtCursor(e, '9') },
	Key{K: "\\"}: func(e *Editor) error { return insertCharAtCursor(e, '\\') },
}

func insertCharAtCursor(e *Editor, char byte) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).InsertCharAtCursor(char)
	default:
		return fmt.Errorf("InsertChartAtCursor is not implemented for %s", t)
	}
}
