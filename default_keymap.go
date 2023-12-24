package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var defaultKeymap = Keymap{

	Key{K: "s", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).Write()
		}

		return nil
	},
	// navigation
	Key{K: "<lmouse>"}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).MoveCursorTo(rl.GetMousePosition())
		}

		return nil

	},
	Key{K: "a", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).BeginingOfTheLine()
		}

		return nil
	},
	Key{K: "e", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).EndOfTheLine()
		}

		return nil
	},

	Key{K: "p", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).PreviousLine()
		}

		return nil

	},

	Key{K: "n", Control: true}: func(e *Editor) error {
		if e.ActiveBuffer().Type() == "text_editor_buffer" {
			return e.ActiveBuffer().(*TextEditorBuffer).NextLine()
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
	Key{K: "f", Control: true}: func(e *Editor) error {
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

	Key{K: "b", Control: true}: func(e *Editor) error {
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

	Key{K: "0", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, ')') },
	Key{K: "1", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '!') },
	Key{K: "2", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '@') },
	Key{K: "3", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '#') },
	Key{K: "4", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '$') },
	Key{K: "5", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '%') },
	Key{K: "6", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '^') },
	Key{K: "7", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '&') },
	Key{K: "8", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '*') },
	Key{K: "9", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '(') },
	Key{K: "a", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'A') },
	Key{K: "b", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'B') },
	Key{K: "c", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'C') },
	Key{K: "d", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'D') },
	Key{K: "e", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'E') },
	Key{K: "f", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'F') },
	Key{K: "g", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'G') },
	Key{K: "h", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'H') },
	Key{K: "i", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'I') },
	Key{K: "j", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'J') },
	Key{K: "k", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'K') },
	Key{K: "l", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'L') },
	Key{K: "m", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'M') },
	Key{K: "n", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'N') },
	Key{K: "o", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'O') },
	Key{K: "p", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'P') },
	Key{K: "q", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'Q') },
	Key{K: "r", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'R') },
	Key{K: "s", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'S') },
	Key{K: "t", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'T') },
	Key{K: "u", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'U') },
	Key{K: "v", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'V') },
	Key{K: "w", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'W') },
	Key{K: "x", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'X') },
	Key{K: "y", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'Y') },
	Key{K: "z", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, 'Z') },
	Key{K: "["}:              func(e *Editor) error { return insertCharAtCursor(e, '[') },
	Key{K: "]"}:              func(e *Editor) error { return insertCharAtCursor(e, ']') },
	Key{K: "{", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '{') },
	Key{K: "}", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '}') },
	Key{K: ";"}:              func(e *Editor) error { return insertCharAtCursor(e, ';') },
	Key{K: ";", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, ':') },
	Key{K: "'"}:              func(e *Editor) error { return insertCharAtCursor(e, '\'') },
	Key{K: "\""}:             func(e *Editor) error { return insertCharAtCursor(e, '"') },
	Key{K: ","}:              func(e *Editor) error { return insertCharAtCursor(e, ',') },
	Key{K: "."}:              func(e *Editor) error { return insertCharAtCursor(e, '.') },
	Key{K: ",", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '<') },
	Key{K: ".", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '>') },
	Key{K: "/"}:              func(e *Editor) error { return insertCharAtCursor(e, '/') },
	Key{K: "/", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '?') },
	Key{K: "-"}:              func(e *Editor) error { return insertCharAtCursor(e, '-') },
	Key{K: "="}:              func(e *Editor) error { return insertCharAtCursor(e, '=') },
	Key{K: "-", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '_') },
	Key{K: "=", Shift: true}: func(e *Editor) error { return insertCharAtCursor(e, '+') },
}

func insertCharAtCursor(e *Editor, char byte) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).InsertCharAtCursor(char)
	default:
		return fmt.Errorf("InsertChartAtCursor is not implemented for %s", t)
	}
}
