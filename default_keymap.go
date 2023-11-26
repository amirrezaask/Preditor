package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const SCROLLSPEED = 10

var defaultKeymap = Keymap{

	Key{K: "s", Control: true}: func(e *Application) error {
		return e.ActiveEditor().Write()
	},
	// navigation
	Key{K: "<lmouse>"}: func(e *Application) error {
		return e.ActiveEditor().MoveCursorTo(rl.GetMousePosition())
	},
	Key{K: "<mouse-wheel-up>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollUp(10)

	},
	Key{K: "<mouse-wheel-down>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollDown(10)
	},

	Key{K: "<rmouse>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollDown(10)
	},
	Key{K: "<mmouse>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollUp(10)
	},

	Key{K: "a", Control: true}: func(e *Application) error {
		return e.ActiveEditor().BeginingOfTheLine()
	},
	Key{K: "e", Control: true}: func(e *Application) error {
		return e.ActiveEditor().EndOfTheLine()
	},

	Key{K: "p", Control: true}: func(e *Application) error {
		return e.ActiveEditor().PreviousLine()
	},

	Key{K: "n", Control: true}: func(e *Application) error {
		return e.ActiveEditor().NextLine()
	},

	Key{K: "<up>"}: func(e *Application) error {
		return e.ActiveEditor().CursorUp()
	},
	Key{K: "<down>"}: func(e *Application) error {
		return e.ActiveEditor().CursorDown()
	},
	Key{K: "<right>"}: func(e *Application) error {
		return e.ActiveEditor().CursorRight(1)
	},
	Key{K: "f", Control: true}: func(e *Application) error {
		return e.ActiveEditor().CursorRight(1)
	},
	Key{K: "<left>"}: func(e *Application) error {
		return e.ActiveEditor().CursorLeft()
	},

	Key{K: "b", Control: true}: func(e *Application) error {
		return e.ActiveEditor().CursorLeft()
	},
	Key{K: "<home>"}: func(e *Application) error {
		return e.ActiveEditor().BeginingOfTheLine()
	},
	Key{K: "<pagedown>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollDown(1)
	},
	Key{K: "<pageup>"}: func(e *Application) error {
		return e.ActiveEditor().ScrollUp(1)
	},

	//insertion
	Key{K: "<enter>"}: func(e *Application) error { return insertCharAtCursor(e, '\n') },
	Key{K: "<space>"}: func(e *Application) error { return insertCharAtCursor(e, ' ') },
	Key{K: "<backspace>"}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharBackward()
	},
	Key{K: "d", Control: true}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "d", Control: true}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "<delete>"}: func(e *Application) error {
		return e.ActiveEditor().DeleteCharForward()
	},
	Key{K: "a"}:               func(e *Application) error { return insertCharAtCursor(e, 'a') },
	Key{K: "b"}:               func(e *Application) error { return insertCharAtCursor(e, 'b') },
	Key{K: "c"}:               func(e *Application) error { return insertCharAtCursor(e, 'c') },
	Key{K: "d"}:               func(e *Application) error { return insertCharAtCursor(e, 'd') },
	Key{K: "e"}:               func(e *Application) error { return insertCharAtCursor(e, 'e') },
	Key{K: "f"}:               func(e *Application) error { return insertCharAtCursor(e, 'f') },
	Key{K: "g"}:               func(e *Application) error { return insertCharAtCursor(e, 'g') },
	Key{K: "h"}:               func(e *Application) error { return insertCharAtCursor(e, 'h') },
	Key{K: "i"}:               func(e *Application) error { return insertCharAtCursor(e, 'i') },
	Key{K: "j"}:               func(e *Application) error { return insertCharAtCursor(e, 'j') },
	Key{K: "k"}:               func(e *Application) error { return insertCharAtCursor(e, 'k') },
	Key{K: "l"}:               func(e *Application) error { return insertCharAtCursor(e, 'l') },
	Key{K: "m"}:               func(e *Application) error { return insertCharAtCursor(e, 'm') },
	Key{K: "n"}:               func(e *Application) error { return insertCharAtCursor(e, 'n') },
	Key{K: "o"}:               func(e *Application) error { return insertCharAtCursor(e, 'o') },
	Key{K: "p"}:               func(e *Application) error { return insertCharAtCursor(e, 'p') },
	Key{K: "q"}:               func(e *Application) error { return insertCharAtCursor(e, 'q') },
	Key{K: "r"}:               func(e *Application) error { return insertCharAtCursor(e, 'r') },
	Key{K: "s"}:               func(e *Application) error { return insertCharAtCursor(e, 's') },
	Key{K: "t"}:               func(e *Application) error { return insertCharAtCursor(e, 't') },
	Key{K: "u"}:               func(e *Application) error { return insertCharAtCursor(e, 'u') },
	Key{K: "v"}:               func(e *Application) error { return insertCharAtCursor(e, 'v') },
	Key{K: "w"}:               func(e *Application) error { return insertCharAtCursor(e, 'w') },
	Key{K: "x"}:               func(e *Application) error { return insertCharAtCursor(e, 'x') },
	Key{K: "y"}:               func(e *Application) error { return insertCharAtCursor(e, 'y') },
	Key{K: "z"}:               func(e *Application) error { return insertCharAtCursor(e, 'z') },
	Key{K: "0"}:               func(e *Application) error { return insertCharAtCursor(e, '0') },
	Key{K: "1"}:               func(e *Application) error { return insertCharAtCursor(e, '1') },
	Key{K: "2"}:               func(e *Application) error { return insertCharAtCursor(e, '2') },
	Key{K: "3"}:               func(e *Application) error { return insertCharAtCursor(e, '3') },
	Key{K: "4"}:               func(e *Application) error { return insertCharAtCursor(e, '4') },
	Key{K: "5"}:               func(e *Application) error { return insertCharAtCursor(e, '5') },
	Key{K: "6"}:               func(e *Application) error { return insertCharAtCursor(e, '6') },
	Key{K: "7"}:               func(e *Application) error { return insertCharAtCursor(e, '7') },
	Key{K: "8"}:               func(e *Application) error { return insertCharAtCursor(e, '8') },
	Key{K: "9"}:               func(e *Application) error { return insertCharAtCursor(e, '9') },
	Key{K: "\\"}:              func(e *Application) error { return insertCharAtCursor(e, '\\') },
	Key{K: "\\", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '|') },

	Key{K: "0", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, ')') },
	Key{K: "1", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '!') },
	Key{K: "2", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '@') },
	Key{K: "3", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '#') },
	Key{K: "4", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '$') },
	Key{K: "5", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '%') },
	Key{K: "6", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '^') },
	Key{K: "7", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '&') },
	Key{K: "8", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '*') },
	Key{K: "9", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '(') },
	Key{K: "a", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'A') },
	Key{K: "b", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'B') },
	Key{K: "c", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'C') },
	Key{K: "d", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'D') },
	Key{K: "e", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'E') },
	Key{K: "f", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'F') },
	Key{K: "g", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'G') },
	Key{K: "h", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'H') },
	Key{K: "i", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'I') },
	Key{K: "j", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'J') },
	Key{K: "k", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'K') },
	Key{K: "l", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'L') },
	Key{K: "m", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'M') },
	Key{K: "n", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'N') },
	Key{K: "o", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'O') },
	Key{K: "p", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'P') },
	Key{K: "q", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'Q') },
	Key{K: "r", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'R') },
	Key{K: "s", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'S') },
	Key{K: "t", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'T') },
	Key{K: "u", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'U') },
	Key{K: "v", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'V') },
	Key{K: "w", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'W') },
	Key{K: "x", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'X') },
	Key{K: "y", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'Y') },
	Key{K: "z", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, 'Z') },
	Key{K: "["}:              func(e *Application) error { return insertCharAtCursor(e, '[') },
	Key{K: "]"}:              func(e *Application) error { return insertCharAtCursor(e, ']') },
	Key{K: "{", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '{') },
	Key{K: "}", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '}') },
	Key{K: ";"}:              func(e *Application) error { return insertCharAtCursor(e, ';') },
	Key{K: ";", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, ':') },
	Key{K: "'"}:              func(e *Application) error { return insertCharAtCursor(e, '\'') },
	Key{K: "\""}:             func(e *Application) error { return insertCharAtCursor(e, '"') },
	Key{K: ","}:              func(e *Application) error { return insertCharAtCursor(e, ',') },
	Key{K: "."}:              func(e *Application) error { return insertCharAtCursor(e, '.') },
	Key{K: ",", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '<') },
	Key{K: ".", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '>') },
	Key{K: "/"}:              func(e *Application) error { return insertCharAtCursor(e, '/') },
	Key{K: "/", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '?') },
	Key{K: "-"}:              func(e *Application) error { return insertCharAtCursor(e, '-') },
	Key{K: "="}:              func(e *Application) error { return insertCharAtCursor(e, '=') },
	Key{K: "-", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '_') },
	Key{K: "=", Shift: true}: func(e *Application) error { return insertCharAtCursor(e, '+') },
	Key{K: "<tab>" }:         func(e *Application) error { return e.ActiveEditor().Indent() },
}

func insertCharAtCursor(e *Application, char byte) error {
	return e.ActiveEditor().InsertCharAtCursor(char)
}
