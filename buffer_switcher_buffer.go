package preditor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type BufferSwitcherBuffer struct {
	BaseBuffer
	cfg                            *Config
	parent                         *Preditor
	keymaps                        []Keymap
	List                           ListComponent[BufferItem]
	UserInputComponent             *UserInputComponent
	LastInputWeCalculatedScoresFor string
}
type BufferItem struct {
	buffer Buffer
	Score  int
}

func NewBufferSwitcherBuffer(parent *Preditor,
	cfg *Config) *BufferSwitcherBuffer {
	var buffers []BufferItem
	for _, v := range parent.Buffers {
		buffers = append(buffers, BufferItem{buffer: v})
	}
	bufferSwitcher := &BufferSwitcherBuffer{
		cfg:     cfg,
		parent:  parent,
		keymaps: []Keymap{bufferSwitcherKeymap},
		List: ListComponent[BufferItem]{
			Items:        buffers,
			VisibleStart: 0,
		},
		UserInputComponent: NewUserInputComponent(parent, cfg),
	}
	return bufferSwitcher
}
func (f *BufferSwitcherBuffer) SortItems() {
	for idx, item := range f.List.Items {
		f.List.Items[idx].Score = fuzzy.RankMatchNormalizedFold(string(f.UserInputComponent.UserInput), item.buffer.String())
	}

	sortme(f.List.Items, func(t1 BufferItem, t2 BufferItem) bool {
		return t1.Score > t2.Score
	})

}

func (b *BufferSwitcherBuffer) Render(zeroPosition rl.Vector2, maxH int32, maxW int32) {
	if b.LastInputWeCalculatedScoresFor != string(b.UserInputComponent.UserInput) {
		b.LastInputWeCalculatedScoresFor = string(b.UserInputComponent.UserInput)
		b.SortItems()
	}
	charSize := measureTextSize(b.parent.Font, ' ', b.parent.FontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(zeroPosition.X), int32(zeroPosition.Y), maxW, int32(charSize.Y)*2, b.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(b.parent.Font, string(b.UserInputComponent.UserInput), rl.Vector2{
		X: zeroPosition.X, Y: zeroPosition.Y + charSize.Y/2,
	}, float32(b.parent.FontSize), 0, b.cfg.Colors.Foreground)

	switch b.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(b.UserInputComponent.Idx), int32(zeroPosition.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(b.UserInputComponent.Idx), int32(zeroPosition.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(b.UserInputComponent.Idx), int32(zeroPosition.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}
	startOfListY := int32(zeroPosition.Y) + int32(3*(charSize.Y))
	//draw list of items
	for idx, item := range b.List.Items {
		rl.DrawTextEx(b.parent.Font, item.buffer.String(), rl.Vector2{
			X: zeroPosition.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, float32(b.parent.FontSize), 0, b.cfg.Colors.Foreground)
	}

	if len(b.List.Items) > 0 {
		rl.DrawRectangle(int32(zeroPosition.X), int32(int(startOfListY)+(b.List.Selection-b.List.VisibleStart)*int(charSize.Y)), maxW, int32(charSize.Y), rl.Fade(b.cfg.Colors.Selection, 0.2))
	}
}

func (b *BufferSwitcherBuffer) String() string {
	return fmt.Sprintf("BufferSwitcher")
}

func (b *BufferSwitcherBuffer) Keymaps() []Keymap {
	return b.keymaps
}
func makeBufferSwitcherCommand(f func(e *BufferSwitcherBuffer) error) Command {
	return func(preditor *Preditor) error {
		defer handlePanicAndWriteMessage(preditor)
		return f(preditor.ActiveBuffer().(*BufferSwitcherBuffer))
	}
}

var bufferSwitcherKeymap = Keymap{
	Key{K: "f", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.CursorRight(1)
	}),
	Key{K: "v", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.paste()
	}),
	Key{K: "c", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.copy()
	}),
	Key{K: "<esc>"}: makeBufferSwitcherCommand(func(p *BufferSwitcherBuffer) error {
		// maybe close ?
		return nil
	}),

	Key{K: "a", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.BeginingOfTheLine()
	}),
	Key{K: "e", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.EndOfTheLine()
	}),

	Key{K: "<right>"}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.CursorRight(1)
	}),
	Key{K: "<right>", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.NextWordStart()
	}),
	Key{K: "<left>"}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.CursorLeft(1)
	}),
	Key{K: "<left>", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.PreviousWord()
	}),
	Key{K: "p", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		e.List.PrevItem()
		return nil
	}),
	Key{K: "n", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		e.List.NextItem()
		return nil
	}),
	Key{K: "<up>"}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		e.List.PrevItem()

		return nil
	}),
	Key{K: "<down>"}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		e.List.NextItem()
		return nil
	}),
	Key{K: "b", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.CursorLeft(1)
	}),
	Key{K: "<home>"}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		return e.UserInputComponent.BeginingOfTheLine()
	}),

	//insertion
	Key{K: "<enter>"}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error {
		defer handlePanicAndWriteMessage(e.parent)
		e.parent.KillBuffer(e.ID)
		e.parent.MarkBufferAsActive(e.List.Items[e.List.Selection].buffer.GetID())
		return nil
	}),
	Key{K: "<space>"}:                    makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error { return e.UserInputComponent.insertCharAtBuffer(' ') }),
	Key{K: "<backspace>"}:                makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error { return e.UserInputComponent.DeleteCharBackward() }),
	Key{K: "<backspace>", Control: true}: makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error { return e.UserInputComponent.DeleteWordBackward() }),
	Key{K: "d", Control: true}:           makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error { return e.UserInputComponent.DeleteCharForward() }),
	Key{K: "d", Alt: true}:               makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error { return e.UserInputComponent.DeleteWordForward() }),
	Key{K: "<delete>"}:                   makeBufferSwitcherCommand(func(e *BufferSwitcherBuffer) error { return e.UserInputComponent.DeleteCharForward() }),
	Key{K: "a"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('a') }),
	Key{K: "b"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('b') }),
	Key{K: "c"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('c') }),
	Key{K: "d"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('d') }),
	Key{K: "e"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('e') }),
	Key{K: "f"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('f') }),
	Key{K: "g"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('g') }),
	Key{K: "h"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('h') }),
	Key{K: "i"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('i') }),
	Key{K: "j"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('j') }),
	Key{K: "k"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('k') }),
	Key{K: "l"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('l') }),
	Key{K: "m"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('m') }),
	Key{K: "n"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('n') }),
	Key{K: "o"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('o') }),
	Key{K: "p"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('p') }),
	Key{K: "q"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('q') }),
	Key{K: "r"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('r') }),
	Key{K: "s"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('s') }),
	Key{K: "t"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('t') }),
	Key{K: "u"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('u') }),
	Key{K: "v"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('v') }),
	Key{K: "w"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('w') }),
	Key{K: "x"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('x') }),
	Key{K: "y"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('y') }),
	Key{K: "z"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('z') }),
	Key{K: "0"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('0') }),
	Key{K: "1"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('1') }),
	Key{K: "2"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('2') }),
	Key{K: "3"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('3') }),
	Key{K: "4"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('4') }),
	Key{K: "5"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('5') }),
	Key{K: "6"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('6') }),
	Key{K: "7"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('7') }),
	Key{K: "8"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('8') }),
	Key{K: "9"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('9') }),
	Key{K: "\\"}:                         makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('\\') }),
	Key{K: "\\", Shift: true}:            makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('|') }),
	Key{K: "0", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer(')') }),
	Key{K: "1", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('!') }),
	Key{K: "2", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('@') }),
	Key{K: "3", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('#') }),
	Key{K: "4", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('$') }),
	Key{K: "5", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('%') }),
	Key{K: "6", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('^') }),
	Key{K: "7", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('&') }),
	Key{K: "8", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('*') }),
	Key{K: "9", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('(') }),
	Key{K: "a", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('A') }),
	Key{K: "b", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('B') }),
	Key{K: "c", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('C') }),
	Key{K: "d", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('D') }),
	Key{K: "e", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('E') }),
	Key{K: "f", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('F') }),
	Key{K: "g", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('G') }),
	Key{K: "h", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('H') }),
	Key{K: "i", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('I') }),
	Key{K: "j", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('J') }),
	Key{K: "k", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('K') }),
	Key{K: "l", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('L') }),
	Key{K: "m", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('M') }),
	Key{K: "n", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('N') }),
	Key{K: "o", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('O') }),
	Key{K: "p", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('P') }),
	Key{K: "q", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Q') }),
	Key{K: "r", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('R') }),
	Key{K: "s", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('S') }),
	Key{K: "t", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('T') }),
	Key{K: "u", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('U') }),
	Key{K: "v", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('V') }),
	Key{K: "w", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('W') }),
	Key{K: "x", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('X') }),
	Key{K: "y", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Y') }),
	Key{K: "z", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Z') }),
	Key{K: "["}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('[') }),
	Key{K: "]"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer(']') }),
	Key{K: "[", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('{') }),
	Key{K: "]", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('}') }),
	Key{K: ";"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer(';') }),
	Key{K: ";", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer(':') }),
	Key{K: "'"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('\'') }),
	Key{K: "'", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('"') }),
	Key{K: "\""}:                         makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('"') }),
	Key{K: ","}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer(',') }),
	Key{K: "."}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('.') }),
	Key{K: ",", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('<') }),
	Key{K: ".", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('>') }),
	Key{K: "/"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('/') }),
	Key{K: "/", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('?') }),
	Key{K: "-"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('-') }),
	Key{K: "="}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('=') }),
	Key{K: "-", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('_') }),
	Key{K: "=", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('+') }),
	Key{K: "`"}:                          makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('`') }),
	Key{K: "`", Shift: true}:             makeBufferSwitcherCommand(func(f *BufferSwitcherBuffer) error { return f.UserInputComponent.insertCharAtBuffer('~') }),
}
