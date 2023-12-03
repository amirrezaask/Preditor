package preditor

var GlobalKeymap = Keymap{
	Key{K: "q", Alt: true}: func(preditor *Preditor) error {
		preditor.KillBuffer(preditor.ActiveBufferIndex)

		return nil
	},
}
