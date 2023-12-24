package preditor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BufferSwitcherBuffer struct {
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	maxHeight    int32
	maxWidth     int32
	ZeroLocation rl.Vector2
	Items        []Buffer
	Selection    int
}

func NewBufferSwitcherBuffer(parent *Preditor,
	cfg *Config,
	maxH int32,
	maxW int32,
	zeroLocation rl.Vector2) *BufferSwitcherBuffer {
	bufferSwitcher := &BufferSwitcherBuffer{
		cfg:          cfg,
		parent:       parent,
		keymaps:      []Keymap{bufferSwitcherKeymap},
		maxHeight:    maxH,
		maxWidth:     maxW,
		ZeroLocation: zeroLocation,
	}
	bufferSwitcher.Items = parent.Buffers
	return bufferSwitcher
}

func (b *BufferSwitcherBuffer) Render() {
	charSize := measureTextSize(font, ' ', FontSize, 0)

	//draw list of items
	for idx, item := range b.Items {
		rl.DrawTextEx(font, item.String(), rl.Vector2{
			X: b.ZeroLocation.X, Y: b.ZeroLocation.Y + float32(idx)*charSize.Y,
		}, FontSize, 0, b.cfg.Colors.Foreground)
	}
	//draw selection
	rl.DrawRectangle(int32(b.ZeroLocation.X), int32(b.ZeroLocation.Y+float32(float32(b.Selection)*charSize.Y)), b.maxWidth, int32(charSize.Y), rl.Fade(b.cfg.Colors.Selection, 0.3))
}

func (b *BufferSwitcherBuffer) String() string {
	return fmt.Sprintf("BufferSwitcher")
}

func (b *BufferSwitcherBuffer) SetMaxWidth(w int32) {
	b.maxWidth = w
}

func (b *BufferSwitcherBuffer) SetMaxHeight(h int32) {
	b.maxHeight = h
}

func (b *BufferSwitcherBuffer) GetMaxWidth() int32 {
	return b.maxWidth
}

func (b *BufferSwitcherBuffer) GetMaxHeight() int32 {
	return b.maxHeight
}

func (b *BufferSwitcherBuffer) Keymaps() []Keymap {
	return b.keymaps
}

var bufferSwitcherKeymap = Keymap{
	Key{K: "<enter>"}: func(preditor *Preditor) error {
		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)
		preditor.ActiveBufferIndex = buffer.Selection

		return nil
	},
	Key{K: "<up>"}: func(preditor *Preditor) error {
		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)

		buffer.Selection--
		if buffer.Selection < 0 {
			buffer.Selection = 0
		}

		return nil
	},
	Key{K: "<down>"}: func(preditor *Preditor) error {
		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)

		buffer.Selection++
		if buffer.Selection >= len(buffer.Items) {
			buffer.Selection = len(buffer.Items) - 1
		}

		return nil
	},
}
