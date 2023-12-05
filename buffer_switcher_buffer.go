package preditor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BufferSwitcherBuffer struct {
	BaseBuffer
	cfg     *Config
	parent  *Preditor
	keymaps []Keymap
	List    ListComponent[Buffer]
}

func NewBufferSwitcherBuffer(parent *Preditor,
	cfg *Config) *BufferSwitcherBuffer {
	var buffers []Buffer
	for _, v := range parent.Buffers {
		buffers = append(buffers, v)
	}
	bufferSwitcher := &BufferSwitcherBuffer{
		cfg:     cfg,
		parent:  parent,
		keymaps: []Keymap{bufferSwitcherKeymap},
		List: ListComponent[Buffer]{
			Items:        buffers,
			VisibleStart: 0,
		},
	}
	return bufferSwitcher
}

func (b *BufferSwitcherBuffer) Render(zeroPosition rl.Vector2, maxH int32, maxW int32) {
	charSize := measureTextSize(b.parent.Font, ' ', b.parent.FontSize, 0)
	maxLine := maxH / int32(charSize.Y)
	//draw list of items
	for idx, item := range b.List.VisibleView(int(maxLine)) {
		rl.DrawTextEx(b.parent.Font, item.String(), rl.Vector2{
			X: zeroPosition.X, Y: zeroPosition.Y + float32(idx)*charSize.Y,
		}, float32(b.parent.FontSize), 0, b.cfg.Colors.Foreground)
	}
	//draw selection
	rl.DrawRectangle(int32(zeroPosition.X), int32(zeroPosition.Y+float32(float32(b.List.Selection)*charSize.Y)), maxW, int32(charSize.Y), rl.Fade(b.cfg.Colors.Selection, 0.3))
}

func (b *BufferSwitcherBuffer) String() string {
	return fmt.Sprintf("BufferSwitcher")
}

func (b *BufferSwitcherBuffer) Keymaps() []Keymap {
	return b.keymaps
}

var bufferSwitcherKeymap = Keymap{
	Key{K: "<enter>"}: func(preditor *Preditor) error {
		defer handlePanicAndWriteMessage(preditor)

		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)
		preditor.KillBuffer(buffer.ID)
		preditor.MarkBufferAsActive(buffer.List.Items[buffer.List.Selection].GetID())
		return nil
	},
	Key{K: "<up>"}: func(preditor *Preditor) error {
		defer handlePanicAndWriteMessage(preditor)

		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)

		buffer.List.Selection--
		if buffer.List.Selection < 0 {
			buffer.List.Selection = 0
		}

		return nil
	},
	Key{K: "<down>"}: func(preditor *Preditor) error {
		defer handlePanicAndWriteMessage(preditor)

		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)

		buffer.List.Selection++
		if buffer.List.Selection >= len(buffer.List.Items) {
			buffer.List.Selection = len(buffer.List.Items) - 1
		}

		return nil
	},
}
