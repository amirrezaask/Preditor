package preditor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path/filepath"
)

type LocationItem struct {
	IsDir    bool
	Filename string
}

type FilePickerBuffer struct {
	cfg                *Config
	parent             *Preditor
	keymaps            []Keymap
	maxHeight          int32
	maxWidth           int32
	ZeroLocation       rl.Vector2
	List               ListComponent[LocationItem]
	LastQuery          string
	UserInputComponent *UserInputComponent
}

func (f *FilePickerBuffer) HandleFontChange() {
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

func NewFilePickerBuffer(parent *Preditor,
	cfg *Config,
	root string,
	maxH int32,
	maxW int32,
	zeroLocation rl.Vector2) *FilePickerBuffer {
	if root == "" {
		root, _ = os.Getwd()
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	charSize := measureTextSize(parent.Font, ' ', parent.FontSize, 0)
	startOfListY := int32(zeroLocation.Y) + int32(3*(charSize.Y))

	ofb := &FilePickerBuffer{
		cfg:       cfg,
		parent:    parent,
		keymaps:   []Keymap{FilePickerKeymap},
		maxHeight: maxH,
		maxWidth:  maxW,
		List: ListComponent[LocationItem]{
			MaxLine:      int(parent.MaxHeightToMaxLine(maxH - startOfListY)),
			VisibleStart: 0,
			VisibleEnd:   int(parent.MaxHeightToMaxLine(maxH - startOfListY)),
		},
		ZeroLocation:       zeroLocation,
		UserInputComponent: NewUserInputComponent(parent, cfg, zeroLocation, maxH, maxW),
	}

	ofb.UserInputComponent.setNewUserInput([]byte(absRoot))
	return ofb
}

func (f *FilePickerBuffer) String() string {
	return fmt.Sprintf("File Picker@%s", f.LastQuery)
}

func (f *FilePickerBuffer) calculateLocationItems() {
	if f.LastQuery == string(f.UserInputComponent.UserInput) {
		return
	}

	f.LastQuery = string(f.UserInputComponent.UserInput)
	input := f.UserInputComponent.UserInput
	matches, err := filepath.Glob(string(input) + "*")
	if err != nil {
		return
	}

	f.List.Items = nil

	for _, match := range matches {
		var isDir bool
		stat, err := os.Stat(match)
		if err == nil {
			isDir = stat.IsDir()
		}
		f.List.Items = append(f.List.Items, LocationItem{
			IsDir:    isDir,
			Filename: match,
		})
	}

	if f.List.Selection >= len(f.List.Items) {
		f.List.Selection = len(f.List.Items) - 1
	}

	return
}

func (f *FilePickerBuffer) Render() {
	f.calculateLocationItems()
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

func (f *FilePickerBuffer) SetMaxWidth(w int32) {
	f.maxWidth = w
	f.HandleFontChange()
}

func (f *FilePickerBuffer) SetMaxHeight(h int32) {
	f.maxHeight = h
	f.HandleFontChange()
}

func (f *FilePickerBuffer) GetMaxWidth() int32 {
	return f.maxWidth
}

func (f *FilePickerBuffer) GetMaxHeight() int32 {
	return f.maxHeight
}

func (f *FilePickerBuffer) Keymaps() []Keymap {
	return f.keymaps
}

func (f *FilePickerBuffer) openUserInput() error {
	err := SwitchOrOpenFileInTextBuffer(f.parent, f.cfg, string(f.UserInputComponent.UserInput), f.maxHeight, f.maxWidth, f.ZeroLocation, nil)
	if err != nil {
		panic(err)
	}
	return nil
}

func (f *FilePickerBuffer) openSelection() error {
	err := SwitchOrOpenFileInTextBuffer(f.parent, f.cfg, f.List.Items[f.List.Selection].Filename, f.maxHeight, f.maxWidth, f.ZeroLocation, nil)
	if err != nil {
		panic(err)
	}
	return nil
}

func (f *FilePickerBuffer) tryComplete() error {
	input := f.UserInputComponent.UserInput

	matches, err := filepath.Glob(string(input) + "*")
	if err != nil {
		return nil
	}

	if len(matches) == 1 {
		stat, err := os.Stat(matches[0])
		if err == nil {
			if stat.IsDir() {
				matches[0] += "/"
			}
		}
		f.UserInputComponent.UserInput = []byte(matches[0])
		f.UserInputComponent.CursorRight(len(f.UserInputComponent.UserInput) - len(input))
	}
	return nil
}

func makeFilePickerCommand(f func(e *FilePickerBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*FilePickerBuffer))
	}
}

func init() {
	FilePickerKeymap = Keymap{

		Key{K: "f", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "v", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.paste()
		}),
		Key{K: "c", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.copy()
		}),
		Key{K: "<esc>"}: makeFilePickerCommand(func(p *FilePickerBuffer) error {
			// maybe close ?
			return nil
		}),

		Key{K: "a", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.BeginingOfTheLine()
		}),
		Key{K: "e", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.EndOfTheLine()
		}),

		Key{K: "<right>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.NextWordStart()
		}),
		Key{K: "<left>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.PreviousWord()
		}),
		Key{K: "p", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			e.List.PrevItem()
			return nil
		}),
		Key{K: "n", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "<up>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			e.List.PrevItem()

			return nil
		}),
		Key{K: "<down>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "b", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<home>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputComponent.BeginingOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}:                    makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.openSelection() }),
		Key{K: "<enter>", Control: true}:     makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.openUserInput() }),
		Key{K: "<tab>"}:                      makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.tryComplete() }),
		Key{K: "<space>"}:                    makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputComponent.insertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputComponent.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputComponent.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputComponent.DeleteWordForward() }),
		Key{K: "<delete>"}:                   makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "a"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('a') }),
		Key{K: "b"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('b') }),
		Key{K: "c"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('c') }),
		Key{K: "d"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('d') }),
		Key{K: "e"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('e') }),
		Key{K: "f"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('f') }),
		Key{K: "g"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('g') }),
		Key{K: "h"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('h') }),
		Key{K: "i"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('i') }),
		Key{K: "j"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('j') }),
		Key{K: "k"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('k') }),
		Key{K: "l"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('l') }),
		Key{K: "m"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('m') }),
		Key{K: "n"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('n') }),
		Key{K: "o"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('o') }),
		Key{K: "p"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('p') }),
		Key{K: "q"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('q') }),
		Key{K: "r"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('r') }),
		Key{K: "s"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('s') }),
		Key{K: "t"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('t') }),
		Key{K: "u"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('u') }),
		Key{K: "v"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('v') }),
		Key{K: "w"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('w') }),
		Key{K: "x"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('x') }),
		Key{K: "y"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('y') }),
		Key{K: "z"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('z') }),
		Key{K: "0"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('0') }),
		Key{K: "1"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('1') }),
		Key{K: "2"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('2') }),
		Key{K: "3"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('3') }),
		Key{K: "4"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('4') }),
		Key{K: "5"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('5') }),
		Key{K: "6"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('6') }),
		Key{K: "7"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('7') }),
		Key{K: "8"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('8') }),
		Key{K: "9"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('9') }),
		Key{K: "\\"}:                         makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('Z') }),
		Key{K: "["}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('[') }),
		Key{K: "]"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('}') }),
		Key{K: ";"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(':') }),
		Key{K: "'"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('"') }),
		Key{K: "\""}:                         makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('"') }),
		Key{K: ","}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer(',') }),
		Key{K: "."}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('>') }),
		Key{K: "/"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('?') }),
		Key{K: "-"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('-') }),
		Key{K: "="}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('+') }),
		Key{K: "`"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputComponent.insertCharAtBuffer('~') }),
	}
}

var FilePickerKeymap Keymap
