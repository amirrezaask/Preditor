package preditor

var GlobalKeymap = Keymap{
	Key{K: "q", Alt: true}: func(preditor *Context) error {
		preditor.KillBuffer(preditor.ActiveBufferID)

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
}
