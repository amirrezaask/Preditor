package preditor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GrepLocationItem struct {
	Filename string
	Text     string
	Line     int
	Col      int
}

func (g GrepLocationItem) StringWithTruncate(maxLen int) string {
	if len(g.Text) > maxLen {
		return fmt.Sprintf("%s: %s", g.Filename, g.Text[:maxLen-1])
	} else {
		return fmt.Sprintf("%s: %s", g.Filename, g.Text)
	}
}

type GrepBuffer struct {
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	root         string
	maxHeight    int32
	maxWidth     int32
	ZeroLocation rl.Vector2
	Items        []GrepLocationItem
	LastQuery    string
	UserInputBox *UserInputComponent
	Selection    int
	maxColumn    int
}

func NewGrepBuffer(parent *Preditor,
	cfg *Config,
	root string,
	maxH int32,
	maxW int32,
	zeroLocation rl.Vector2) *GrepBuffer {
	if root == "" {
		root, _ = os.Getwd()
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	charSize := measureTextSize(font, ' ', fontSize, 0)
	ofb := &GrepBuffer{
		cfg:          cfg,
		parent:       parent,
		root:         absRoot,
		keymaps:      []Keymap{GrepBufferKeymap},
		maxHeight:    maxH,
		maxWidth:     maxW,
		ZeroLocation: zeroLocation,
		maxColumn:    int(maxW / int32(charSize.X)),
		UserInputBox: NewUserInputComponent(parent, cfg, zeroLocation, maxH, maxW),
	}

	return ofb
}

func (f *GrepBuffer) String() string {
	return fmt.Sprintf("Grep Buffer@%s", f.LastQuery)
}

func (f *GrepBuffer) calculateLocationItems() error {
	c := runRipgrep(string(f.UserInputBox.UserInput))
	go func() {
		lines := <-c
		f.Items = nil

		for _, line := range lines {
			lineS := string(line)
			segs := strings.SplitN(lineS, ":", 4)
			if len(segs) < 4 {
				continue
			}

			line, err := strconv.Atoi(segs[1])
			if err != nil {
				continue
			}
			col, err := strconv.Atoi(segs[2])
			if err != nil {
				continue
			}
			f.Items = append(f.Items, GrepLocationItem{
				Filename: segs[0],
				Line:     line,
				Col:      col,
				Text:     segs[3],
			})
		}
	}()

	return nil
}

func (f *GrepBuffer) Render() {
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
		rl.DrawTextEx(font, item.StringWithTruncate(f.maxColumn), rl.Vector2{
			X: f.ZeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, fontSize, 0, f.cfg.Colors.Foreground)
	}
	if len(f.Items) > 0 {
		rl.DrawRectangle(int32(f.ZeroLocation.X), int32(int(startOfListY)+f.Selection*int(charSize.Y)), f.maxWidth, int32(charSize.Y), rl.Fade(f.cfg.Colors.Selection, 0.2))
	}

}

func (f *GrepBuffer) SetMaxWidth(w int32) {
	f.maxWidth = w
}

func (f *GrepBuffer) SetMaxHeight(h int32) {
	f.maxHeight = h
}

func (f *GrepBuffer) GetMaxWidth() int32 {
	return f.maxWidth
}

func (f *GrepBuffer) GetMaxHeight() int32 {
	return f.maxHeight
}

func (f *GrepBuffer) Keymaps() []Keymap {
	return f.keymaps
}

func (f *GrepBuffer) openUserInput() error {

	return nil
}

func (e *GrepBuffer) NextSelection() error {
	e.Selection++
	if e.Selection >= len(e.Items) {
		e.Selection = len(e.Items) - 1
	}

	return nil
}

func (e *GrepBuffer) PrevSelection() error {
	e.Selection--
	if e.Selection < 0 {
		e.Selection = 0
	}
	return nil
}
func (e *GrepBuffer) OpenSelection() error {
	item := e.Items[e.Selection]

	return SwitchOrOpenFileInTextBuffer(e.parent, e.cfg, item.Filename, e.maxHeight, e.maxWidth, e.ZeroLocation, &Position{Line: item.Line, Column: item.Col})
}

func makeGrepBufferCommand(f func(e *GrepBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*GrepBuffer))
	}
}

func init() {
	GrepBufferKeymap = Keymap{

		Key{K: "f", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.CursorRight(1)
		}),
		Key{K: "v", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.paste()
		}),
		Key{K: "c", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.copy()
		}),
		Key{K: "s", Control: true}: makeGrepBufferCommand(func(a *GrepBuffer) error {
			a.keymaps = append(a.keymaps, SearchTextBufferKeymap)
			return nil
		}),
		Key{K: "<esc>"}: makeGrepBufferCommand(func(p *GrepBuffer) error {
			// maybe close ?
			p.parent.Buffers = p.parent.Buffers[:len(p.parent.Buffers)-1]
			p.parent.ActiveBufferIndex = len(p.parent.Buffers) - 1
			return nil
		}),

		Key{K: "a", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.BeginingOfTheLine()
		}),
		Key{K: "e", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.EndOfTheLine()
		}),

		Key{K: "<right>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.NextWordStart()
		}),
		Key{K: "<left>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.PreviousWord()
		}),

		Key{K: "p", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.PrevSelection()
		}),
		Key{K: "n", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.NextSelection()
		}),
		Key{K: "<up>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.PrevSelection()
		}),
		Key{K: "<down>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.NextSelection()
		}),
		Key{K: "b", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.CursorLeft(1)
		}),
		Key{K: "<home>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			return e.UserInputBox.BeginingOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}:                    makeGrepBufferCommand(func(e *GrepBuffer) error { return e.calculateLocationItems() }),
		Key{K: "<enter>", Control: true}:     makeGrepBufferCommand(func(e *GrepBuffer) error { return e.OpenSelection() }),
		Key{K: "<space>"}:                    makeGrepBufferCommand(func(e *GrepBuffer) error { return e.UserInputBox.insertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                makeGrepBufferCommand(func(e *GrepBuffer) error { return e.UserInputBox.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error { return e.UserInputBox.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           makeGrepBufferCommand(func(e *GrepBuffer) error { return e.UserInputBox.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               makeGrepBufferCommand(func(e *GrepBuffer) error { return e.UserInputBox.DeleteWordForward() }),
		Key{K: "<delete>"}:                   makeGrepBufferCommand(func(e *GrepBuffer) error { return e.UserInputBox.DeleteCharForward() }),
		Key{K: "a"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('a') }),
		Key{K: "b"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('b') }),
		Key{K: "c"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('c') }),
		Key{K: "d"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('d') }),
		Key{K: "e"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('e') }),
		Key{K: "f"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('f') }),
		Key{K: "g"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('g') }),
		Key{K: "h"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('h') }),
		Key{K: "i"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('i') }),
		Key{K: "j"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('j') }),
		Key{K: "k"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('k') }),
		Key{K: "l"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('l') }),
		Key{K: "m"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('m') }),
		Key{K: "n"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('n') }),
		Key{K: "o"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('o') }),
		Key{K: "p"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('p') }),
		Key{K: "q"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('q') }),
		Key{K: "r"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('r') }),
		Key{K: "s"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('s') }),
		Key{K: "t"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('t') }),
		Key{K: "u"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('u') }),
		Key{K: "v"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('v') }),
		Key{K: "w"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('w') }),
		Key{K: "x"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('x') }),
		Key{K: "y"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('y') }),
		Key{K: "z"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('z') }),
		Key{K: "0"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('0') }),
		Key{K: "1"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('1') }),
		Key{K: "2"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('2') }),
		Key{K: "3"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('3') }),
		Key{K: "4"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('4') }),
		Key{K: "5"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('5') }),
		Key{K: "6"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('6') }),
		Key{K: "7"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('7') }),
		Key{K: "8"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('8') }),
		Key{K: "9"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('9') }),
		Key{K: "\\"}:                         makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('Z') }),
		Key{K: "["}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('[') }),
		Key{K: "]"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('}') }),
		Key{K: ";"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer(':') }),
		Key{K: "'"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('"') }),
		Key{K: "\""}:                         makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('"') }),
		Key{K: ","}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer(',') }),
		Key{K: "."}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('>') }),
		Key{K: "/"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('?') }),
		Key{K: "-"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('-') }),
		Key{K: "="}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('+') }),
		Key{K: "`"}:                          makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             makeGrepBufferCommand(func(f *GrepBuffer) error { return f.UserInputBox.insertCharAtBuffer('~') }),
	}
}

var GrepBufferKeymap Keymap
