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
	Key{K: "<esc>"}: func(preditor *Context) error {
		preditor.CloseBottomOverlay()
		return nil
	},
	Key{K: "p", Alt: true}: func(preditor *Context) error {
		preditor.OpenBottomOverlay()
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
	Key{K: "t", Alt: true}: func(preditor *Context) error {
		preditor.openThemeSwitcher()
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

func init() {
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

}
