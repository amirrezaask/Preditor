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
	List         ListComponent[Buffer]
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
		List: ListComponent[Buffer]{
			Items:        parent.Buffers,
			MaxLine:      int(parent.MaxHeightToMaxLine(maxH)),
			VisibleStart: 0,
			VisibleEnd:   int(parent.MaxHeightToMaxLine(maxH) - 1),
		},
	}
	return bufferSwitcher
}

func (b *BufferSwitcherBuffer) HandleFontChange() {
	charSize := measureTextSize(b.parent.Font, ' ', b.parent.FontSize, 0)
	startOfListY := int32(b.ZeroLocation.Y) + int32(3*(charSize.Y))
	oldEnd := b.List.VisibleEnd
	oldStart := b.List.VisibleStart
	b.List.MaxLine = int(b.parent.MaxHeightToMaxLine(b.maxHeight - startOfListY))
	b.List.VisibleEnd = int(b.parent.MaxHeightToMaxLine(b.maxHeight - startOfListY))
	b.List.VisibleStart += (b.List.VisibleEnd - oldEnd)

	if int(b.List.VisibleEnd) >= len(b.List.Items) {
		b.List.VisibleEnd = len(b.List.Items) - 1
		b.List.VisibleStart = b.List.VisibleEnd - b.List.MaxLine
	}

	if b.List.VisibleStart < 0 {
		b.List.VisibleStart = 0
		b.List.VisibleEnd = b.List.MaxLine
	}
	if b.List.VisibleEnd < 0 {
		b.List.VisibleStart = 0
		b.List.VisibleEnd = b.List.MaxLine
	}

	diff := b.List.VisibleStart - oldStart
	b.List.Selection += diff
}

func (b *BufferSwitcherBuffer) Render() {
	charSize := measureTextSize(b.parent.Font, ' ', b.parent.FontSize, 0)

	//draw list of items
	for idx, item := range b.List.VisibleView() {
		rl.DrawTextEx(b.parent.Font, item.String(), rl.Vector2{
			X: b.ZeroLocation.X, Y: b.ZeroLocation.Y + float32(idx)*charSize.Y,
		}, float32(b.parent.FontSize), 0, b.cfg.Colors.Foreground)
	}
	//draw selection
	rl.DrawRectangle(int32(b.ZeroLocation.X), int32(b.ZeroLocation.Y+float32(float32(b.List.Selection)*charSize.Y)), b.maxWidth, int32(charSize.Y), rl.Fade(b.cfg.Colors.Selection, 0.3))
}

func (b *BufferSwitcherBuffer) String() string {
	return fmt.Sprintf("BufferSwitcher")
}

func (b *BufferSwitcherBuffer) SetMaxWidth(w int32) {
	b.maxWidth = w
	b.HandleFontChange()
}

func (b *BufferSwitcherBuffer) SetMaxHeight(h int32) {
	b.maxHeight = h
	b.HandleFontChange()
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
		preditor.ActiveBufferIndex = buffer.List.Selection

		return nil
	},
	Key{K: "<up>"}: func(preditor *Preditor) error {
		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)

		buffer.List.Selection--
		if buffer.List.Selection < 0 {
			buffer.List.Selection = 0
		}

		return nil
	},
	Key{K: "<down>"}: func(preditor *Preditor) error {
		buffer := preditor.ActiveBuffer().(*BufferSwitcherBuffer)

		buffer.List.Selection++
		if buffer.List.Selection >= len(buffer.List.Items) {
			buffer.List.Selection = len(buffer.List.Items) - 1
		}

		return nil
	},
}
