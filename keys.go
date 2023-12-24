package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var defaultKeymap = Keymap{
	Key{K: "<up>"}: func(e Editor) error {
		window := e.CurrentWindow()
		buffer := e.CurrentBuffer()
		newPosition := window.Cursor
		newPosition.Line--
		if e.isValidCursorPosition(window, buffer, newPosition) {
			window.Cursor.Line = window.Cursor.Line - 1
		}

		return nil
	},
	Key{K: "<down>"}: func(e Editor) error {
		window := e.CurrentWindow()
		buffer := e.CurrentBuffer()
		newPosition := window.Cursor
		newPosition.Line++
		if e.isValidCursorPosition(window, buffer, newPosition) {
			window.Cursor.Line = window.Cursor.Line + 1
		}
		return nil

	},
	Key{K: "<right>"}: func(e Editor) error {
		window := e.CurrentWindow()
		buffer := e.CurrentBuffer()
		newPosition := window.Cursor
		newPosition.Column++

		if e.isValidCursorPosition(window, buffer, newPosition) {
			window.Cursor.Column = window.Cursor.Column + 1
		}
		return nil

	},
	Key{K: "<left>"}: func(e Editor) error {
		window := e.CurrentWindow()
		buffer := e.CurrentBuffer()
		newPosition := window.Cursor
		newPosition.Column--
		if e.isValidCursorPosition(window, buffer, newPosition) {
			window.Cursor.Column = window.Cursor.Column - 1
		}

		return nil
	},
	Key{K: "<enter>"}: func(e Editor) error { return e.InsertCharAtCursor('\n') },
	Key{K: " "}:       func(e Editor) error { return e.InsertCharAtCursor(' ') },
	Key{K: "a"}:       func(e Editor) error { return e.InsertCharAtCursor('a') },
	Key{K: "b"}:       func(e Editor) error { return e.InsertCharAtCursor('b') },
	Key{K: "c"}:       func(e Editor) error { return e.InsertCharAtCursor('c') },
	Key{K: "d"}:       func(e Editor) error { return e.InsertCharAtCursor('d') },
	Key{K: "e"}:       func(e Editor) error { return e.InsertCharAtCursor('e') },
	Key{K: "f"}:       func(e Editor) error { return e.InsertCharAtCursor('f') },
	Key{K: "g"}:       func(e Editor) error { return e.InsertCharAtCursor('g') },
	Key{K: "h"}:       func(e Editor) error { return e.InsertCharAtCursor('h') },
	Key{K: "i"}:       func(e Editor) error { return e.InsertCharAtCursor('i') },
	Key{K: "j"}:       func(e Editor) error { return e.InsertCharAtCursor('j') },
	Key{K: "k"}:       func(e Editor) error { return e.InsertCharAtCursor('k') },
	Key{K: "l"}:       func(e Editor) error { return e.InsertCharAtCursor('l') },
	Key{K: "m"}:       func(e Editor) error { return e.InsertCharAtCursor('m') },
	Key{K: "n"}:       func(e Editor) error { return e.InsertCharAtCursor('n') },
	Key{K: "o"}:       func(e Editor) error { return e.InsertCharAtCursor('o') },
	Key{K: "p"}:       func(e Editor) error { return e.InsertCharAtCursor('p') },
	Key{K: "q"}:       func(e Editor) error { return e.InsertCharAtCursor('q') },
	Key{K: "r"}:       func(e Editor) error { return e.InsertCharAtCursor('r') },
	Key{K: "s"}:       func(e Editor) error { return e.InsertCharAtCursor('s') },
	Key{K: "t"}:       func(e Editor) error { return e.InsertCharAtCursor('t') },
	Key{K: "u"}:       func(e Editor) error { return e.InsertCharAtCursor('u') },
	Key{K: "v"}:       func(e Editor) error { return e.InsertCharAtCursor('v') },
	Key{K: "w"}:       func(e Editor) error { return e.InsertCharAtCursor('w') },
	Key{K: "x"}:       func(e Editor) error { return e.InsertCharAtCursor('x') },
	Key{K: "y"}:       func(e Editor) error { return e.InsertCharAtCursor('y') },
	Key{K: "z"}:       func(e Editor) error { return e.InsertCharAtCursor('z') },
	Key{K: "0"}:       func(e Editor) error { return e.InsertCharAtCursor('0') },
	Key{K: "1"}:       func(e Editor) error { return e.InsertCharAtCursor('1') },
	Key{K: "2"}:       func(e Editor) error { return e.InsertCharAtCursor('2') },
	Key{K: "3"}:       func(e Editor) error { return e.InsertCharAtCursor('3') },
	Key{K: "4"}:       func(e Editor) error { return e.InsertCharAtCursor('4') },
	Key{K: "5"}:       func(e Editor) error { return e.InsertCharAtCursor('5') },
	Key{K: "6"}:       func(e Editor) error { return e.InsertCharAtCursor('6') },
	Key{K: "7"}:       func(e Editor) error { return e.InsertCharAtCursor('7') },
	Key{K: "8"}:       func(e Editor) error { return e.InsertCharAtCursor('8') },
	Key{K: "9"}:       func(e Editor) error { return e.InsertCharAtCursor('9') },
}

func MakeKey(buffer *Buffer) Key {
	keyPressed := rl.GetKeyPressed()
	var k Key
	switch {
	case rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl):
		k.Ctrl = true
	case rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt):
		k.Alt = true
	case rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift):
		k.Shift = true
	case keyPressed == rl.KeyUp:
		k.K = "<up>"
	case keyPressed == rl.KeyDown:
		k.K = "<down>"
	case keyPressed == rl.KeyRight:
		k.K = "<right>"
	case keyPressed == rl.KeyLeft:
		k.K = "<left>"
	case keyPressed == rl.KeyKp0:
		k.K = "0"
	case keyPressed == rl.KeyKp1:
		k.K = "1"
	case keyPressed == rl.KeyKp2:
		k.K = "2"
	case keyPressed == rl.KeyKp3:
		k.K = "3"
	case keyPressed == rl.KeyKp4:
		k.K = "4"
	case keyPressed == rl.KeyKp5:
		k.K = "5"
	case keyPressed == rl.KeyKp6:
		k.K = "6"
	case keyPressed == rl.KeyKp7:
		k.K = "7"
	case keyPressed == rl.KeyKp8:
		k.K = "8"
	case keyPressed == rl.KeyKp9:
		k.K = "9"
	// case keyPressed == rl.KeyKpDecimal:
	// case keyPressed == rl.KeyKpDivide:
	// case keyPressed == rl.KeyKpMultiply:
	// case keyPressed == rl.KeyKpSubtract:
	// case keyPressed == rl.KeyKpAdd:
	// case keyPressed == rl.KeyKpEnter:
	// case keyPressed == rl.KeyKpEqual:
	case keyPressed == rl.KeyApostrophe:
		k.K = "'"
	case keyPressed == rl.KeyComma:
		k.K = ","
	case keyPressed == rl.KeyMinus:
		k.K = "-"
	case keyPressed == rl.KeyPeriod:
		k.K = "."
	case keyPressed == rl.KeySlash:
		k.K = "/"
	case keyPressed == rl.KeyZero:
		k.K = "0"
	case keyPressed == rl.KeyOne:
		k.K = "1"
	case keyPressed == rl.KeyTwo:
		k.K = "2"
	case keyPressed == rl.KeyThree:
		k.K = "3"
	case keyPressed == rl.KeyFour:
		k.K = "4"
	case keyPressed == rl.KeyFive:
		k.K = "5"
	case keyPressed == rl.KeySix:
		k.K = "6"
	case keyPressed == rl.KeySeven:
		k.K = "7"
	case keyPressed == rl.KeyEight:
		k.K = "8"
	case keyPressed == rl.KeyNine:
		k.K = "9"
	case keyPressed == rl.KeySemicolon:
		k.K = ";"
	case keyPressed == rl.KeyEqual:
		k.K = "="
	case keyPressed == rl.KeyA:
		k.K = "a"
	case keyPressed == rl.KeyB:
		k.K = "b"
	case keyPressed == rl.KeyC:
		k.K = "c"
	case keyPressed == rl.KeyD:
		k.K = "d"
	case keyPressed == rl.KeyE:
		k.K = "e"
	case keyPressed == rl.KeyF:
		k.K = "f"
	case keyPressed == rl.KeyG:
		k.K = "g"
	case keyPressed == rl.KeyH:
		k.K = "h"
	case keyPressed == rl.KeyI:
		k.K = "i"
	case keyPressed == rl.KeyJ:
		k.K = "j"
	case keyPressed == rl.KeyK:
		k.K = "k"
	case keyPressed == rl.KeyL:
		k.K = "l"
	case keyPressed == rl.KeyM:
		k.K = "m"
	case keyPressed == rl.KeyN:
		k.K = "n"
	case keyPressed == rl.KeyO:
		k.K = "o"
	case keyPressed == rl.KeyP:
		k.K = "p"
	case keyPressed == rl.KeyQ:
		k.K = "q"
	case keyPressed == rl.KeyR:
		k.K = "r"
	case keyPressed == rl.KeyS:
		k.K = "s"
	case keyPressed == rl.KeyT:
		k.K = "t"
	case keyPressed == rl.KeyU:
		k.K = "u"
	case keyPressed == rl.KeyV:
		k.K = "v"
	case keyPressed == rl.KeyW:
		k.K = "w"
	case keyPressed == rl.KeyX:
		k.K = "x"
	case keyPressed == rl.KeyY:
		k.K = "y"
	case keyPressed == rl.KeyZ:
		k.K = "z"

	case keyPressed == rl.KeySpace:
		k.K = " "
	case keyPressed == rl.KeyEscape:
	case keyPressed == rl.KeyEnter:
		k.K = "<enter>"
	case keyPressed == rl.KeyTab:
		k.K = "<tab>"
	case keyPressed == rl.KeyBackspace:
		k.K = "<backspace>"
	case keyPressed == rl.KeyInsert:
		k.K = "<insert>"
	case keyPressed == rl.KeyDelete:
		k.K = "<delete>"
	case keyPressed == rl.KeyPageUp:
		k.K = "<pageup>"
	case keyPressed == rl.KeyPageDown:
		k.K = "<pagedown>"
	case keyPressed == rl.KeyHome:
		k.K = "<home>"
	case keyPressed == rl.KeyEnd:
		k.K = "<end>"
	case keyPressed == rl.KeyCapsLock:
		k.K = "<capslock>"
	case keyPressed == rl.KeyScrollLock:
		k.K = "<scorlllock>"
	case keyPressed == rl.KeyNumLock:
		k.K = "<numlock>"
	case keyPressed == rl.KeyPrintScreen:
		k.K = "<printscreen>"
	case keyPressed == rl.KeyPause:
		k.K = "<pause>"
	case keyPressed == rl.KeyF1:
		k.K = "<f1>"
	case keyPressed == rl.KeyF2:
		k.K = "<f2>"
	case keyPressed == rl.KeyF3:
		k.K = "<f3>"
	case keyPressed == rl.KeyF4:
		k.K = "<f4>"
	case keyPressed == rl.KeyF5:
		k.K = "<f5>"
	case keyPressed == rl.KeyF6:
		k.K = "<f6>"
	case keyPressed == rl.KeyF7:
		k.K = "<f7>"
	case keyPressed == rl.KeyF8:
		k.K = "<f8>"
	case keyPressed == rl.KeyF9:
		k.K = "<f9>"
	case keyPressed == rl.KeyF10:
		k.K = "<f10>"
	case keyPressed == rl.KeyF11:
		k.K = "<f11>"
	case keyPressed == rl.KeyF12:
		k.K = "<f12>"
	case keyPressed == rl.KeyLeftBracket:
		k.K = "["
	case keyPressed == rl.KeyBackSlash:
		k.K = "\\"
	case keyPressed == rl.KeyRightBracket:
		k.K = "]"
	}

	return k
}
