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
	BaseBuffer
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	root         string
	List         ListComponent[GrepLocationItem]
	LastQuery    string
	UserInputBox *UserInputComponent
	maxColumn    int
}

func NewGrepBuffer(parent *Preditor,
	cfg *Config,
	root string) *GrepBuffer {
	if root == "" {
		root, _ = os.Getwd()
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}

	ofb := &GrepBuffer{
		cfg:          cfg,
		parent:       parent,
		root:         absRoot,
		keymaps:      []Keymap{GrepBufferKeymap},
		List:         ListComponent[GrepLocationItem]{},
		UserInputBox: NewUserInputComponent(parent, cfg),
	}

	return ofb
}

func (f *GrepBuffer) String() string {
	return fmt.Sprintf("Grep Buffer@%s", f.LastQuery)
}

func (f *GrepBuffer) calculateLocationItems() error {
	c := RipgrepAsync(string(f.UserInputBox.UserInput))
	go func() {
		lines := <-c
		f.List.Items = nil

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
			f.List.Items = append(f.List.Items, GrepLocationItem{
				Filename: segs[0],
				Line:     line,
				Col:      col,
				Text:     segs[3],
			})
		}
	}()

	return nil
}

func (f *GrepBuffer) Render(zeroLocation rl.Vector2, maxH int32, maxW int32) {
	charSize := measureTextSize(f.parent.Font, ' ', f.parent.FontSize, 0)
	f.maxColumn = int(maxW / int32(charSize.X))

	//draw input box
	rl.DrawRectangleLines(int32(zeroLocation.X), int32(zeroLocation.Y), maxW, int32(charSize.Y)*2, f.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(f.parent.Font, string(f.UserInputBox.UserInput), rl.Vector2{
		X: zeroLocation.X, Y: zeroLocation.Y + charSize.Y/2,
	}, float32(f.parent.FontSize), 0, f.cfg.Colors.Foreground)

	switch f.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(zeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(zeroLocation.Y) + int32(3*(charSize.Y))
	maxLine := int((maxH - startOfListY) / int32(charSize.Y))

	//draw list of items
	for idx, item := range f.List.VisibleView(maxLine) {
		rl.DrawTextEx(f.parent.Font, item.StringWithTruncate(f.maxColumn), rl.Vector2{
			X: zeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, float32(f.parent.FontSize), 0, f.cfg.Colors.Foreground)
	}
	if len(f.List.Items) > 0 {
		rl.DrawRectangle(int32(zeroLocation.X), int32(int(startOfListY)+(f.List.Selection-f.List.VisibleStart)*int(charSize.Y)), maxW, int32(charSize.Y), rl.Fade(f.cfg.Colors.Selection, 0.2))
	}

}

func (f *GrepBuffer) Keymaps() []Keymap {
	return f.keymaps
}

func (e *GrepBuffer) OpenSelection() error {
	item := e.List.Items[e.List.Selection]

	return SwitchOrOpenFileInTextBuffer(e.parent, e.cfg, item.Filename, &Position{Line: item.Line, Column: item.Col})
}

func makeGrepBufferCommand(f func(e *GrepBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*GrepBuffer))
	}
}

func init() {
	GrepBufferKeymap = Keymap{
		Key{K: "=", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			e.parent.IncreaseFontSize(5)

			return nil
		}),
		Key{K: "-", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			e.parent.DecreaseFontSize(5)

			return nil
		}),
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
			e.List.PrevItem()
			return nil
		}),
		Key{K: "n", Control: true}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "<up>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			e.List.PrevItem()

			return nil
		}),
		Key{K: "<down>"}: makeGrepBufferCommand(func(e *GrepBuffer) error {
			e.List.NextItem()
			return nil
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