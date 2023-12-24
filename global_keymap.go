package preditor

var GlobalKeymap = Keymap{
	Key{K: "q", Alt: true}: func(preditor *Preditor) error {
		preditor.KillBuffer(preditor.ActiveBufferIndex)

		return nil
	},
	Key{K: "o", Alt: true}: func(a *Preditor) error {
		a.openFileBuffer()
		return nil
	},
	Key{K: "o", Alt: true, Shift: true}: func(a *Preditor) error {
		a.openFuzzyFilePicker()

		return nil
	},
	Key{K: "b", Alt: true}: func(a *Preditor) error {
		a.openBufferSwitcher()

		return nil
	},
	Key{K: "s", Alt: true}: func(a *Preditor) error {
		a.openGrepBuffer()

		return nil
	},
}
