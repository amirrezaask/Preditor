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

type FilePickerBuffer struct {
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	maxHeight    int32
	maxWidth     int32
	ZeroLocation rl.Vector2
	Items        []LocationItem
	LastQuery    string
	UserInputBox *UserInputBox
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
	ofb := &FilePickerBuffer{
		cfg:          cfg,
		parent:       parent,
		keymaps:      []Keymap{filePickerKeymap},
		maxHeight:    maxH,
		maxWidth:     maxW,
		ZeroLocation: zeroLocation,
		UserInputBox: NewuserInputBox(parent, cfg, zeroLocation, maxH, maxW),
	}

	ofb.UserInputBox.setNewUserInput([]byte(absRoot))
	return ofb
}

func (f *FilePickerBuffer) calculateLocationItems() {
	if f.LastQuery == string(f.UserInputBox.UserInput) {
		return
	}

	f.LastQuery = string(f.UserInputBox.UserInput)
	input := f.UserInputBox.UserInput
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

func (f *FilePickerBuffer) Render() {
	f.calculateLocationItems()
	charSize := measureTextSize(font, ' ', fontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(f.ZeroLocation.X), int32(f.ZeroLocation.Y), f.maxWidth, int32(charSize.Y)*2, f.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(font, string(f.UserInputBox.UserInput), rl.Vector2{
		X: f.ZeroLocation.X, Y: f.ZeroLocation.Y + charSize.Y/2,
	}, fontSize, 0, f.cfg.Colors.Foreground)

	switch f.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(f.ZeroLocation.Y) + int32(3*(charSize.Y))
	//draw list of items
	for idx, item := range f.Items {
		rl.DrawTextEx(font, item.Filename, rl.Vector2{
			X: f.ZeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, fontSize, 0, f.cfg.Colors.Foreground)

	}

}

func (f *FilePickerBuffer) SetMaxWidth(w int32) {
	f.maxWidth = w
}

func (f *FilePickerBuffer) SetMaxHeight(h int32) {
	f.maxHeight = h
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
	p := f.parent

	for _, window := range p.Buffers {
		switch window.(type) {
		case *TextBuffer:
			tb := window.(*TextBuffer)
			if tb.File == string(f.UserInputBox.UserInput) {
				p.Buffers[p.ActiveWindowIndex] = window
				return nil
			}
		}
	}

	e, err := NewTextBuffer(f.parent, f.cfg, string(f.UserInputBox.UserInput), f.maxHeight, f.maxWidth, f.ZeroLocation)
	if err != nil {
		panic(err)
	}
	p.Buffers[p.ActiveWindowIndex] = e
	return nil
}

func (f *FilePickerBuffer) tryComplete() error {
	input := f.UserInputBox.UserInput

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
		f.UserInputBox.UserInput = []byte(matches[0])
		f.UserInputBox.CursorRight(len(f.UserInputBox.UserInput) - len(input))
	}
	return nil
}

func makeFilePickerCommand(f func(e *FilePickerBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*FilePickerBuffer))
	}
}

func init() {
	filePickerKeymap = Keymap{

		Key{K: "f", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.CursorRight(1)
		}),
		Key{K: "y", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.paste()
		}),
		Key{K: "k", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.killLine()
		}),
		Key{K: "w", Alt: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.copy()
		}),
		Key{K: "s", Control: true}: makeFilePickerCommand(func(a *FilePickerBuffer) error {
			a.keymaps = append(a.keymaps, searchModeKeymap)
			return nil
		}),
		Key{K: "<esc>"}: makeFilePickerCommand(func(p *FilePickerBuffer) error {
			// maybe close ?
			return nil
		}),

		Key{K: "a", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.BeginingOfTheLine()
		}),
		Key{K: "e", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.EndOfTheLine()
		}),

		Key{K: "<right>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.NextWordStart()
		}),
		Key{K: "<left>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.PreviousWord()
		}),

		Key{K: "b", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.CursorLeft(1)
		}),
		Key{K: "<home>"}: makeFilePickerCommand(func(e *FilePickerBuffer) error {
			return e.UserInputBox.BeginingOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}:                    makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.openUserInput() }),
		Key{K: "<tab>"}:                      makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.tryComplete() }),
		Key{K: "<space>"}:                    makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputBox.insertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputBox.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputBox.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputBox.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputBox.DeleteWordForward() }),
		Key{K: "<delete>"}:                   makeFilePickerCommand(func(e *FilePickerBuffer) error { return e.UserInputBox.DeleteCharForward() }),
		Key{K: "a"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('a') }),
		Key{K: "b"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('b') }),
		Key{K: "c"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('c') }),
		Key{K: "d"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('d') }),
		Key{K: "e"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('e') }),
		Key{K: "f"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('f') }),
		Key{K: "g"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('g') }),
		Key{K: "h"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('h') }),
		Key{K: "i"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('i') }),
		Key{K: "j"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('j') }),
		Key{K: "k"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('k') }),
		Key{K: "l"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('l') }),
		Key{K: "m"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('m') }),
		Key{K: "n"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('n') }),
		Key{K: "o"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('o') }),
		Key{K: "p"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('p') }),
		Key{K: "q"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('q') }),
		Key{K: "r"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('r') }),
		Key{K: "s"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('s') }),
		Key{K: "t"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('t') }),
		Key{K: "u"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('u') }),
		Key{K: "v"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('v') }),
		Key{K: "w"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('w') }),
		Key{K: "x"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('x') }),
		Key{K: "y"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('y') }),
		Key{K: "z"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('z') }),
		Key{K: "0"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('0') }),
		Key{K: "1"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('1') }),
		Key{K: "2"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('2') }),
		Key{K: "3"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('3') }),
		Key{K: "4"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('4') }),
		Key{K: "5"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('5') }),
		Key{K: "6"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('6') }),
		Key{K: "7"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('7') }),
		Key{K: "8"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('8') }),
		Key{K: "9"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('9') }),
		Key{K: "\\"}:                         makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('Z') }),
		Key{K: "["}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('[') }),
		Key{K: "]"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('}') }),
		Key{K: ";"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer(':') }),
		Key{K: "'"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('"') }),
		Key{K: "\""}:                         makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('"') }),
		Key{K: ","}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer(',') }),
		Key{K: "."}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('>') }),
		Key{K: "/"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('?') }),
		Key{K: "-"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('-') }),
		Key{K: "="}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('+') }),
		Key{K: "`"}:                          makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             makeFilePickerCommand(func(f *FilePickerBuffer) error { return f.UserInputBox.insertCharAtBuffer('~') }),
	}
}

var filePickerKeymap Keymap
