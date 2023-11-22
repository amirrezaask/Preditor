package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var defaultKeymap = Keymap{
	Key{K: "<up>"}: func(e Editor) error {
		window := e.CurrentWindow()
		if window.Cursor.Line-1 >= 0 {
			window.Cursor.Line = window.Cursor.Line - 1
		}

		return nil
	},
	Key{K: "<down>"}: func(e Editor) error {
		window := e.CurrentWindow()
		if window.Cursor.Line+1 < len(window.VisualLines) {
			window.Cursor.Line = window.Cursor.Line + 1
		}
		return nil

	},
	Key{K: "<right>"}: func(e Editor) error {
		window := e.CurrentWindow()
		charSize := measureTextSize(font, ' ', fontSize, 0)

		if window.Cursor.Column+1 < (window.Width / int(charSize.X)) {
			window.Cursor.Column = window.Cursor.Column + 1
		}
		return nil

	},
	Key{K: "<left>"}: func(e Editor) error {
		window := e.CurrentWindow()
		if window.Cursor.Column-1 >= 0 {
			window.Cursor.Column = window.Cursor.Column - 1
		}

		return nil
	},
	Key{K: "a"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('a') },
	Key{K: "b"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('b') },
	Key{K: "c"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('c') },
	Key{K: "d"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('d') },
	Key{K: "e"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('e') },
	Key{K: "f"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('f') },
	Key{K: "g"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('g') },
	Key{K: "h"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('h') },
	Key{K: "i"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('i') },
	Key{K: "j"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('j') },
	Key{K: "k"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('k') },
	Key{K: "l"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('l') },
	Key{K: "m"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('m') },
	Key{K: "n"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('n') },
	Key{K: "o"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('o') },
	Key{K: "p"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('p') },
	Key{K: "q"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('q') },
	Key{K: "r"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('r') },
	Key{K: "s"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('s') },
	Key{K: "t"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('t') },
	Key{K: "u"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('u') },
	Key{K: "v"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('v') },
	Key{K: "w"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('w') },
	Key{K: "x"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('x') },
	Key{K: "y"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('y') },
	Key{K: "z"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('z') },
	Key{K: "0"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('0') },
	Key{K: "1"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('1') },
	Key{K: "2"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('2') },
	Key{K: "3"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('3') },
	Key{K: "4"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('4') },
	Key{K: "5"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('5') },
	Key{K: "6"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('6') },
	Key{K: "7"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('7') },
	Key{K: "8"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('8') },
	Key{K: "9"}: func(e Editor) error { return e.CurrentBuffer().InsertCharAtCursor('9') },
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
	}


	return k
}
