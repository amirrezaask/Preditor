package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path/filepath"
)

type LocationItem struct {
	IsDir    bool
	Filename string
}

type OpenFileBuffer struct {
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	maxHeight    int32
	maxWidth     int32
	UserInput    []byte
	ZeroLocation rl.Vector2
	Idx          int
	Items        []LocationItem
	CursorShape  int
	LastQuery    string
}

func NewOpenFileBuffer(parent *Preditor,
	cfg *Config,
	root string,
	maxH int32,
	maxW int32,
	zeroLocation rl.Vector2) *OpenFileBuffer {
	if root == "" {
		root, _ = os.Getwd()
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	ofb := &OpenFileBuffer{
		cfg:          cfg,
		parent:       parent,
		keymaps:      []Keymap{filePickerKeymap},
		maxHeight:    maxH,
		maxWidth:     maxW,
		CursorShape:  CURSOR_SHAPE_BLOCK,
		ZeroLocation: zeroLocation,
	}

	ofb.setNewUserInput([]byte(absRoot))
	return ofb
}

func (f *OpenFileBuffer) setNewUserInput(bs []byte) {
	f.UserInput = bs
	f.Idx += len(f.UserInput)

	if f.Idx >= len(f.UserInput) {
		f.Idx = len(f.UserInput)
	} else if f.Idx < 0 {
		f.Idx = 0
	}

}

func (f *OpenFileBuffer) calculateLocationItems() {
	if f.LastQuery == string(f.UserInput) {
		return
	}

	f.LastQuery = string(f.UserInput)
	input := f.UserInput
	matches, err := filepath.Glob(string(input) + "*")
	if err != nil {
		return
	}

	f.Items = nil

	for _, match := range matches {
		var isDir bool
		stat, err := os.Stat(match)
		if err == nil {
			isDir = stat.IsDir()
		}
		f.Items = append(f.Items, LocationItem{
			IsDir:    isDir,
			Filename: match,
		})
	}

	return
}

func (f *OpenFileBuffer) Render() {
	f.calculateLocationItems()
	charSize := measureTextSize(font, ' ', fontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(f.ZeroLocation.X), int32(f.ZeroLocation.Y), f.maxWidth, int32(charSize.Y)*2, f.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(font, string(f.UserInput), rl.Vector2{
		X: f.ZeroLocation.X, Y: f.ZeroLocation.Y + charSize.Y/2,
	}, fontSize, 0, f.cfg.Colors.Foreground)

	switch f.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(f.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(f.ZeroLocation.Y) + int32(3*(charSize.Y))
	//draw list of items
	for idx, item := range f.Items {
		rl.DrawTextEx(font, item.Filename, rl.Vector2{
			X: f.ZeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, fontSize, 0, f.cfg.Colors.Foreground)

	}

}

func (f *OpenFileBuffer) SetMaxWidth(w int32) {
	f.maxWidth = w
}

func (f *OpenFileBuffer) SetMaxHeight(h int32) {
	f.maxHeight = h
}

func (f *OpenFileBuffer) GetMaxWidth() int32 {
	return f.maxWidth
}

func (f *OpenFileBuffer) GetMaxHeight() int32 {
	return f.maxHeight
}

func (f *OpenFileBuffer) Keymaps() []Keymap {
	return f.keymaps
}

func (f *OpenFileBuffer) insertCharAtBuffer(char byte) error {
	f.setNewUserInput(append(f.UserInput, char))
	return nil
}

func (f *OpenFileBuffer) CursorRight(n int) error {
	if f.Idx >= len(f.UserInput) {
		return nil
	}

	f.Idx += n

	return nil
}

func (f *OpenFileBuffer) paste() error {
	content := getClipboardContent()
	f.UserInput = append(f.UserInput[:f.Idx], append(content, f.UserInput[f.Idx+1:]...)...)

	return nil
}

func (f *OpenFileBuffer) killLine() error {
	f.setNewUserInput(f.UserInput[:f.Idx])
	return nil
}

func (f *OpenFileBuffer) copy() error {
	writeToClipboard(f.UserInput)

	return nil
}

func (f *OpenFileBuffer) BeginingOfTheLine() error {
	f.Idx = 0
	return nil
}

func (f *OpenFileBuffer) EndOfTheLine() error {
	f.Idx = len(f.UserInput)
	return nil
}

func (f *OpenFileBuffer) NextWordStart() error {
	if idx := nextWordInBuffer(f.UserInput, f.Idx); idx != -1 {
		f.Idx = idx
	}

	return nil
}

func (f *OpenFileBuffer) CursorLeft(n int) error {

	if f.Idx <= 0 {
		return nil
	}

	f.Idx -= n

	return nil
}

func (f *OpenFileBuffer) PreviousWord() error {
	if idx := previousWordInBuffer(f.UserInput, f.Idx); idx != -1 {
		f.Idx = idx
	}

	return nil
}

func (f *OpenFileBuffer) DeleteCharBackward() error {
	if f.Idx <= 0 {
		return nil
	}
	if len(f.UserInput) <= f.Idx {
		f.setNewUserInput(f.UserInput[:f.Idx-1])
	} else {
		f.setNewUserInput(append(f.UserInput[:f.Idx-1], f.UserInput[f.Idx:]...))
	}
	return nil
}

func (f *OpenFileBuffer) DeleteWordBackward() error {
	previousWordEndIdx := previousWordInBuffer(f.UserInput, f.Idx)
	if len(f.UserInput) > f.Idx+1 {
		f.setNewUserInput(append(f.UserInput[:previousWordEndIdx+1], f.UserInput[f.Idx+1:]...))
	} else {
		f.setNewUserInput(f.UserInput[:previousWordEndIdx+1])
	}
	return nil
}
func (f *OpenFileBuffer) DeleteWordForward() error {
	nextWordStartIdx := nextWordInBuffer(f.UserInput, f.Idx)
	if len(f.UserInput) > nextWordStartIdx+1 {
		f.setNewUserInput(append(f.UserInput[:f.Idx+1], f.UserInput[nextWordStartIdx+1:]...))
	} else {
		f.setNewUserInput(f.UserInput[:f.Idx])
	}

	return nil
}
func (f *OpenFileBuffer) DeleteCharForward() error {
	if f.Idx < 0 {
		return nil
	}
	f.setNewUserInput(append(f.UserInput[:f.Idx], f.UserInput[f.Idx+1:]...))
	return nil
}

func (f *OpenFileBuffer) openUserInput() error {
	p := f.parent

	for _, window := range p.Windows {
		switch window.(type) {
		case *TextBuffer:
			tb := window.(*TextBuffer)
			if tb.File == string(f.UserInput) {
				p.Windows[p.ActiveWindowIndex] = window
				return nil
			}
		}
	}

	e, err := NewTextBuffer(f.parent, f.cfg, string(f.UserInput), f.maxHeight, f.maxWidth, f.ZeroLocation)
	if err != nil {
		panic(err)
	}
	p.Windows[p.ActiveWindowIndex] = e
	return nil
}

func (f *OpenFileBuffer) tryComplete() error {
	input := f.UserInput

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
		f.UserInput = []byte(matches[0])
		f.CursorRight(len(f.UserInput) - len(input))
	}
	return nil
}

func makeFilePickerCommand(f func(e *OpenFileBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveWindow().(*OpenFileBuffer))
	}
}

func init() {
	filePickerKeymap = Keymap{

		Key{K: "f", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.CursorRight(1)
		}),
		Key{K: "y", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.paste()
		}),
		Key{K: "k", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.killLine()
		}),
		Key{K: "w", Alt: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.copy()
		}),
		Key{K: "s", Control: true}: makeFilePickerCommand(func(a *OpenFileBuffer) error {
			a.keymaps = append(a.keymaps, searchModeKeymap)
			return nil
		}),
		Key{K: "<esc>"}: makeFilePickerCommand(func(p *OpenFileBuffer) error {
			// maybe close ?
			return nil
		}),

		Key{K: "a", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.BeginingOfTheLine()
		}),
		Key{K: "e", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.EndOfTheLine()
		}),

		Key{K: "<right>"}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.NextWordStart()
		}),
		Key{K: "<left>"}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.PreviousWord()
		}),

		Key{K: "b", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.CursorLeft(1)
		}),
		Key{K: "<home>"}: makeFilePickerCommand(func(e *OpenFileBuffer) error {
			return e.BeginingOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}:                    makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.openUserInput() }),
		Key{K: "<tab>"}:                      makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.tryComplete() }),
		Key{K: "<space>"}:                    makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.insertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.DeleteWordForward() }),
		Key{K: "<delete>"}:                   makeFilePickerCommand(func(e *OpenFileBuffer) error { return e.DeleteCharForward() }),
		Key{K: "a"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('a') }),
		Key{K: "b"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('b') }),
		Key{K: "c"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('c') }),
		Key{K: "d"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('d') }),
		Key{K: "e"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('e') }),
		Key{K: "f"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('f') }),
		Key{K: "g"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('g') }),
		Key{K: "h"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('h') }),
		Key{K: "i"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('i') }),
		Key{K: "j"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('j') }),
		Key{K: "k"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('k') }),
		Key{K: "l"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('l') }),
		Key{K: "m"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('m') }),
		Key{K: "n"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('n') }),
		Key{K: "o"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('o') }),
		Key{K: "p"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('p') }),
		Key{K: "q"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('q') }),
		Key{K: "r"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('r') }),
		Key{K: "s"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('s') }),
		Key{K: "t"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('t') }),
		Key{K: "u"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('u') }),
		Key{K: "v"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('v') }),
		Key{K: "w"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('w') }),
		Key{K: "x"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('x') }),
		Key{K: "y"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('y') }),
		Key{K: "z"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('z') }),
		Key{K: "0"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('0') }),
		Key{K: "1"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('1') }),
		Key{K: "2"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('2') }),
		Key{K: "3"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('3') }),
		Key{K: "4"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('4') }),
		Key{K: "5"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('5') }),
		Key{K: "6"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('6') }),
		Key{K: "7"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('7') }),
		Key{K: "8"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('8') }),
		Key{K: "9"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('9') }),
		Key{K: "\\"}:                         makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('Z') }),
		Key{K: "["}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('[') }),
		Key{K: "]"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('}') }),
		Key{K: ";"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer(':') }),
		Key{K: "'"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('"') }),
		Key{K: "\""}:                         makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('"') }),
		Key{K: ","}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer(',') }),
		Key{K: "."}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('>') }),
		Key{K: "/"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('?') }),
		Key{K: "-"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('-') }),
		Key{K: "="}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('+') }),
		Key{K: "`"}:                          makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             makeFilePickerCommand(func(f *OpenFileBuffer) error { return f.insertCharAtBuffer('~') }),
	}
}

var filePickerKeymap Keymap
