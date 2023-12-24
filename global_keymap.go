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
}
