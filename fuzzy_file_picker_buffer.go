package preditor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type LocationItemScored struct {
	Filename string
	Score    int
}

type FuzzyFilePickerBuffer struct {
	Root                           string
	cfg                            *Config
	parent                         *Preditor
	keymaps                        []Keymap
	maxHeight                      int32
	maxWidth                       int32
	ZeroLocation                   rl.Vector2
	List                           ListComponent[LocationItemScored]
	UserInputComponent             *UserInputComponent
	LastInputWeCalculatedScoresFor string
}

func (f *FuzzyFilePickerBuffer) HandleFontChange() {
	charSize := measureTextSize(f.parent.Font, ' ', f.parent.FontSize, 0)
	startOfListY := int32(f.ZeroLocation.Y) + int32(3*(charSize.Y))
	oldEnd := f.List.VisibleEnd
	oldStart := f.List.VisibleStart
	f.List.MaxLine = int(f.parent.MaxHeightToMaxLine(f.maxHeight - startOfListY))
	f.List.VisibleEnd = int(f.parent.MaxHeightToMaxLine(f.maxHeight - startOfListY))
	f.List.VisibleStart += (f.List.VisibleEnd - oldEnd)

	if int(f.List.VisibleEnd) >= len(f.List.Items) {
		f.List.VisibleEnd = len(f.List.Items) - 1
		f.List.VisibleStart = f.List.VisibleEnd - f.List.MaxLine
	}

	if f.List.VisibleStart < 0 {
		f.List.VisibleStart = 0
		f.List.VisibleEnd = f.List.MaxLine
	}
	if f.List.VisibleEnd < 0 {
		f.List.VisibleStart = 0
		f.List.VisibleEnd = f.List.MaxLine
	}

	diff := f.List.VisibleStart - oldStart
	f.List.Selection += diff
}

func NewFuzzyFilePickerBuffer(parent *Preditor,
	cfg *Config,
	root string,
	maxH int32,
	maxW int32,
	zeroLocation rl.Vector2) *FuzzyFilePickerBuffer {
	charSize := measureTextSize(parent.Font, ' ', parent.FontSize, 0)
	startOfListY := int32(zeroLocation.Y) + int32(3*(charSize.Y))

	ofb := &FuzzyFilePickerBuffer{
		cfg:       cfg,
		parent:    parent,
		Root:      root,
		keymaps:   []Keymap{FilePickerKeymap},
		maxHeight: maxH,
		maxWidth:  maxW,
		List: ListComponent[LocationItemScored]{
			MaxLine:      int(parent.MaxHeightToMaxLine(maxH - startOfListY)),
			VisibleStart: 0,
			VisibleEnd:   int(parent.MaxHeightToMaxLine(maxH - startOfListY)),
		},
		ZeroLocation:       zeroLocation,
		UserInputComponent: NewUserInputComponent(parent, cfg, zeroLocation, maxH, maxW),
	}

	files := RipgrepFiles()
	for _, file := range files {
		ofb.List.Items = append(ofb.List.Items, LocationItemScored{
			Filename: file,
		})
	}
	return ofb
}

func (f *FuzzyFilePickerBuffer) String() string {
	return fmt.Sprintf("Fuzzy File Picker@%s", f.UserInputComponent.LastInput)
}

func (f *FuzzyFilePickerBuffer) SortItems() {
	for idx, item := range f.List.Items {
		f.List.Items[idx].Score = fuzzy.RankMatch(string(f.UserInputComponent.UserInput), item.Filename)
	}

	sortme(f.List.Items, func(t1 LocationItemScored, t2 LocationItemScored) bool {
		return t1.Score > t2.Score
	})

}

func (f *FuzzyFilePickerBuffer) Render() {
	if f.LastInputWeCalculatedScoresFor != string(f.UserInputComponent.UserInput) {
		f.LastInputWeCalculatedScoresFor = string(f.UserInputComponent.UserInput)
		f.SortItems()
	}
	charSize := measureTextSize(f.parent.Font, ' ', f.parent.FontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(f.ZeroLocation.X), int32(f.ZeroLocation.Y), f.maxWidth, int32(charSize.Y)*2, f.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(f.parent.Font, string(f.UserInputComponent.UserInput), rl.Vector2{
		X: f.ZeroLocation.X, Y: f.ZeroLocation.Y + charSize.Y/2,
	}, float32(f.parent.FontSize), 0, f.cfg.Colors.Foreground)

	switch f.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputComponent.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(f.UserInputComponent.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputComponent.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(f.ZeroLocation.Y) + int32(3*(charSize.Y))
	//draw list of items
	for idx, item := range f.List.Items {
		rl.DrawTextEx(f.parent.Font, item.Filename, rl.Vector2{
			X: f.ZeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, float32(f.parent.FontSize), 0, f.cfg.Colors.Foreground)
	}

	if len(f.List.Items) > 0 {
		rl.DrawRectangle(int32(f.ZeroLocation.X), int32(int(startOfListY)+(f.List.Selection-f.List.VisibleStart)*int(charSize.Y)), f.maxWidth, int32(charSize.Y), rl.Fade(f.cfg.Colors.Selection, 0.2))
	}

}

func (f *FuzzyFilePickerBuffer) SetMaxWidth(w int32) {
	f.maxWidth = w
	f.HandleFontChange()
}

func (f *FuzzyFilePickerBuffer) SetMaxHeight(h int32) {
	f.maxHeight = h
	f.HandleFontChange()
}

func (f *FuzzyFilePickerBuffer) GetMaxWidth() int32 {
	return f.maxWidth
}

func (f *FuzzyFilePickerBuffer) GetMaxHeight() int32 {
	return f.maxHeight
}

func (f *FuzzyFilePickerBuffer) Keymaps() []Keymap {
	return f.keymaps
}

func (f *FuzzyFilePickerBuffer) openSelection() error {
	err := SwitchOrOpenFileInTextBuffer(f.parent, f.cfg, f.List.Items[f.List.Selection].Filename, f.maxHeight, f.maxWidth, f.ZeroLocation, nil)
	if err != nil {
		panic(err)
	}
	return nil
}

func makeFuzzyFilePickerCommand(f func(e *FuzzyFilePickerBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*FuzzyFilePickerBuffer))
	}
}

func init() {
	FuzzyFilePickerKeymap = Keymap{

		Key{K: "f", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "v", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.paste()
		}),
		Key{K: "c", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.copy()
		}),
		Key{K: "<esc>"}: makeFuzzyFilePickerCommand(func(p *FuzzyFilePickerBuffer) error {
			// maybe close ?
			return nil
		}),

		Key{K: "a", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.BeginingOfTheLine()
		}),
		Key{K: "e", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.EndOfTheLine()
		}),

		Key{K: "<right>"}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.NextWordStart()
		}),
		Key{K: "<left>"}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.PreviousWord()
		}),
		Key{K: "p", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			e.List.PrevItem()
			return nil
		}),
		Key{K: "n", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "<up>"}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			e.List.PrevItem()

			return nil
		}),
		Key{K: "<down>"}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "b", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<home>"}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error {
			return e.UserInputComponent.BeginingOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}:                    makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.openSelection() }),
		Key{K: "<space>"}:                    makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.UserInputComponent.insertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.UserInputComponent.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.UserInputComponent.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.UserInputComponent.DeleteWordForward() }),
		Key{K: "<delete>"}:                   makeFuzzyFilePickerCommand(func(e *FuzzyFilePickerBuffer) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "a"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('a') }),
		Key{K: "b"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('b') }),
		Key{K: "c"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('c') }),
		Key{K: "d"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('d') }),
		Key{K: "e"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('e') }),
		Key{K: "f"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('f') }),
		Key{K: "g"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('g') }),
		Key{K: "h"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('h') }),
		Key{K: "i"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('i') }),
		Key{K: "j"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('j') }),
		Key{K: "k"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('k') }),
		Key{K: "l"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('l') }),
		Key{K: "m"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('m') }),
		Key{K: "n"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('n') }),
		Key{K: "o"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('o') }),
		Key{K: "p"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('p') }),
		Key{K: "q"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('q') }),
		Key{K: "r"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('r') }),
		Key{K: "s"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('s') }),
		Key{K: "t"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('t') }),
		Key{K: "u"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('u') }),
		Key{K: "v"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('v') }),
		Key{K: "w"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('w') }),
		Key{K: "x"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('x') }),
		Key{K: "y"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('y') }),
		Key{K: "z"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('z') }),
		Key{K: "0"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('0') }),
		Key{K: "1"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('1') }),
		Key{K: "2"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('2') }),
		Key{K: "3"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('3') }),
		Key{K: "4"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('4') }),
		Key{K: "5"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('5') }),
		Key{K: "6"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('6') }),
		Key{K: "7"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('7') }),
		Key{K: "8"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('8') }),
		Key{K: "9"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('9') }),
		Key{K: "\\"}:                         makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Z') }),
		Key{K: "["}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('[') }),
		Key{K: "]"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('}') }),
		Key{K: ";"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(':') }),
		Key{K: "'"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('"') }),
		Key{K: "\""}:                         makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('"') }),
		Key{K: ","}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(',') }),
		Key{K: "."}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('>') }),
		Key{K: "/"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('?') }),
		Key{K: "-"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('-') }),
		Key{K: "="}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('+') }),
		Key{K: "`"}:                          makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             makeFuzzyFilePickerCommand(func(f *FuzzyFilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('~') }),
	}
}

var FuzzyFilePickerKeymap Keymap
