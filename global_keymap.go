package preditor

var GlobalKeymap = Keymap{
	Key{K: "\\", Alt: true}: func(c *Context) error {
		c.VSplit()

		return nil
	},
	Key{K: "=", Alt: true}: func(c *Context) error {
		c.HSplit()

		return nil
	},
	Key{K: "q", Alt: true}: func(preditor *Context) error {
		preditor.CloseWindow(preditor.ActiveWindowIndex)

		return nil
	},
	Key{K: "0", Alt: true}: func(preditor *Context) error {
		preditor.CloseWindow(preditor.ActiveWindowIndex)

		return nil
	},
	Key{K: "k", Alt: true}: func(preditor *Context) error {
		preditor.KillBuffer(preditor.ActiveBufferID())

		return nil
	},
	Key{K: "o", Alt: true}: func(a *Context) error {
		a.openFileBuffer()
		return nil
	},
	Key{K: "o", Alt: true, Shift: true}: func(a *Context) error {
		a.openFuzzyFilePicker()

		return nil
	},
	Key{K: "b", Alt: true}: func(a *Context) error {
		a.openBufferSwitcher()

		return nil
	},
	Key{K: "s", Alt: true}: func(a *Context) error {
		a.openGrepBuffer()

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
}

func MakeInsertionKeys[T Buffer](insertor func(b byte) error) Keymap {
	return Keymap{
		Key{K: "a"}:                    MakeCommand(func(e T) error { return insertor('a') }),
		Key{K: "b"}:                    MakeCommand(func(e T) error { return insertor('b') }),
		Key{K: "c"}:                    MakeCommand(func(e T) error { return insertor('c') }),
		Key{K: "d"}:                    MakeCommand(func(e T) error { return insertor('d') }),
		Key{K: "e"}:                    MakeCommand(func(e T) error { return insertor('e') }),
		Key{K: "f"}:                    MakeCommand(func(e T) error { return insertor('f') }),
		Key{K: "g"}:                    MakeCommand(func(e T) error { return insertor('g') }),
		Key{K: "h"}:                    MakeCommand(func(e T) error { return insertor('h') }),
		Key{K: "i"}:                    MakeCommand(func(e T) error { return insertor('i') }),
		Key{K: "j"}:                    MakeCommand(func(e T) error { return insertor('j') }),
		Key{K: "k"}:                    MakeCommand(func(e T) error { return insertor('k') }),
		Key{K: "l"}:                    MakeCommand(func(e T) error { return insertor('l') }),
		Key{K: "m"}:                    MakeCommand(func(e T) error { return insertor('m') }),
		Key{K: "n"}:                    MakeCommand(func(e T) error { return insertor('n') }),
		Key{K: "o"}:                    MakeCommand(func(e T) error { return insertor('o') }),
		Key{K: "p"}:                    MakeCommand(func(e T) error { return insertor('p') }),
		Key{K: "q"}:                    MakeCommand(func(e T) error { return insertor('q') }),
		Key{K: "r"}:                    MakeCommand(func(e T) error { return insertor('r') }),
		Key{K: "s"}:                    MakeCommand(func(e T) error { return insertor('s') }),
		Key{K: "t"}:                    MakeCommand(func(e T) error { return insertor('t') }),
		Key{K: "u"}:                    MakeCommand(func(e T) error { return insertor('u') }),
		Key{K: "v"}:                    MakeCommand(func(e T) error { return insertor('v') }),
		Key{K: "w"}:                    MakeCommand(func(e T) error { return insertor('w') }),
		Key{K: "x"}:                    MakeCommand(func(e T) error { return insertor('x') }),
		Key{K: "y"}:                    MakeCommand(func(e T) error { return insertor('y') }),
		Key{K: "z"}:                    MakeCommand(func(e T) error { return insertor('z') }),
		Key{K: "0"}:                    MakeCommand(func(e T) error { return insertor('0') }),
		Key{K: "1"}:                    MakeCommand(func(e T) error { return insertor('1') }),
		Key{K: "2"}:                    MakeCommand(func(e T) error { return insertor('2') }),
		Key{K: "3"}:                    MakeCommand(func(e T) error { return insertor('3') }),
		Key{K: "4"}:                    MakeCommand(func(e T) error { return insertor('4') }),
		Key{K: "5"}:                    MakeCommand(func(e T) error { return insertor('5') }),
		Key{K: "6"}:                    MakeCommand(func(e T) error { return insertor('6') }),
		Key{K: "7"}:                    MakeCommand(func(e T) error { return insertor('7') }),
		Key{K: "8"}:                    MakeCommand(func(e T) error { return insertor('8') }),
		Key{K: "9"}:                    MakeCommand(func(e T) error { return insertor('9') }),
		Key{K: "\\"}:                   MakeCommand(func(e T) error { return insertor('\\') }),
		Key{K: "\\", Shift: true}:      MakeCommand(func(e T) error { return insertor('|') }),
		Key{K: "0", Shift: true}:       MakeCommand(func(e T) error { return insertor(')') }),
		Key{K: "1", Shift: true}:       MakeCommand(func(e T) error { return insertor('!') }),
		Key{K: "2", Shift: true}:       MakeCommand(func(e T) error { return insertor('@') }),
		Key{K: "3", Shift: true}:       MakeCommand(func(e T) error { return insertor('#') }),
		Key{K: "4", Shift: true}:       MakeCommand(func(e T) error { return insertor('$') }),
		Key{K: "5", Shift: true}:       MakeCommand(func(e T) error { return insertor('%') }),
		Key{K: "6", Shift: true}:       MakeCommand(func(e T) error { return insertor('^') }),
		Key{K: "7", Shift: true}:       MakeCommand(func(e T) error { return insertor('&') }),
		Key{K: "8", Shift: true}:       MakeCommand(func(e T) error { return insertor('*') }),
		Key{K: "9", Shift: true}:       MakeCommand(func(e T) error { return insertor('(') }),
		Key{K: "a", Shift: true}:       MakeCommand(func(e T) error { return insertor('A') }),
		Key{K: "b", Shift: true}:       MakeCommand(func(e T) error { return insertor('B') }),
		Key{K: "c", Shift: true}:       MakeCommand(func(e T) error { return insertor('C') }),
		Key{K: "d", Shift: true}:       MakeCommand(func(e T) error { return insertor('D') }),
		Key{K: "e", Shift: true}:       MakeCommand(func(e T) error { return insertor('E') }),
		Key{K: "f", Shift: true}:       MakeCommand(func(e T) error { return insertor('F') }),
		Key{K: "g", Shift: true}:       MakeCommand(func(e T) error { return insertor('G') }),
		Key{K: "h", Shift: true}:       MakeCommand(func(e T) error { return insertor('H') }),
		Key{K: "i", Shift: true}:       MakeCommand(func(e T) error { return insertor('I') }),
		Key{K: "j", Shift: true}:       MakeCommand(func(e T) error { return insertor('J') }),
		Key{K: "k", Shift: true}:       MakeCommand(func(e T) error { return insertor('K') }),
		Key{K: "l", Shift: true}:       MakeCommand(func(e T) error { return insertor('L') }),
		Key{K: "m", Shift: true}:       MakeCommand(func(e T) error { return insertor('M') }),
		Key{K: "n", Shift: true}:       MakeCommand(func(e T) error { return insertor('N') }),
		Key{K: "o", Shift: true}:       MakeCommand(func(e T) error { return insertor('O') }),
		Key{K: "p", Shift: true}:       MakeCommand(func(e T) error { return insertor('P') }),
		Key{K: "q", Shift: true}:       MakeCommand(func(e T) error { return insertor('Q') }),
		Key{K: "r", Shift: true}:       MakeCommand(func(e T) error { return insertor('R') }),
		Key{K: "s", Shift: true}:       MakeCommand(func(e T) error { return insertor('S') }),
		Key{K: "t", Shift: true}:       MakeCommand(func(e T) error { return insertor('T') }),
		Key{K: "u", Shift: true}:       MakeCommand(func(e T) error { return insertor('U') }),
		Key{K: "v", Shift: true}:       MakeCommand(func(e T) error { return insertor('V') }),
		Key{K: "w", Shift: true}:       MakeCommand(func(e T) error { return insertor('W') }),
		Key{K: "x", Shift: true}:       MakeCommand(func(e T) error { return insertor('X') }),
		Key{K: "y", Shift: true}:       MakeCommand(func(e T) error { return insertor('Y') }),
		Key{K: "z", Shift: true}:       MakeCommand(func(e T) error { return insertor('Z') }),
		Key{K: "["}:                    MakeCommand(func(e T) error { return insertor('[') }),
		Key{K: "]"}:                    MakeCommand(func(e T) error { return insertor(']') }),
		Key{K: "[", Shift: true}:       MakeCommand(func(e T) error { return insertor('{') }),
		Key{K: "]", Shift: true}:       MakeCommand(func(e T) error { return insertor('}') }),
		Key{K: ";"}:                    MakeCommand(func(e T) error { return insertor(';') }),
		Key{K: ";", Shift: true}:       MakeCommand(func(e T) error { return insertor(':') }),
		Key{K: "'"}:                    MakeCommand(func(e T) error { return insertor('\'') }),
		Key{K: "'", Shift: true}:       MakeCommand(func(e T) error { return insertor('"') }),
		Key{K: "\""}:                   MakeCommand(func(e T) error { return insertor('"') }),
		Key{K: ","}:                    MakeCommand(func(e T) error { return insertor(',') }),
		Key{K: "."}:                    MakeCommand(func(e T) error { return insertor('.') }),
		Key{K: ",", Shift: true}:       MakeCommand(func(e T) error { return insertor('<') }),
		Key{K: ".", Shift: true}:       MakeCommand(func(e T) error { return insertor('>') }),
		Key{K: "/"}:                    MakeCommand(func(e T) error { return insertor('/') }),
		Key{K: "/", Shift: true}:       MakeCommand(func(e T) error { return insertor('?') }),
		Key{K: "-"}:                    MakeCommand(func(e T) error { return insertor('-') }),
		Key{K: "="}:                    MakeCommand(func(e T) error { return insertor('=') }),
		Key{K: "-", Shift: true}:       MakeCommand(func(e T) error { return insertor('_') }),
		Key{K: "=", Shift: true}:       MakeCommand(func(e T) error { return insertor('+') }),
		Key{K: "`"}:                    MakeCommand(func(e T) error { return insertor('`') }),
		Key{K: "`", Shift: true}:       MakeCommand(func(e T) error { return insertor('~') }),
		Key{K: "<space>", Shift: true}: MakeCommand(func(e T) error { return insertor(' ') }),
	}
}

func MergeKeymaps(k1 Keymap, k2 Keymap) Keymap {
	for k, v := range k2 {
		k1[k] = v
	}
	return k1
}
