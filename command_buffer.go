package preditor

import (
	"bytes"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandLocationItem struct {
	Filename string
	Text     string
	Line     int
	Col      int
}

type CommandBuffer struct {
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	root         string
	maxHeight    int32
	maxWidth     int32
	ZeroLocation rl.Vector2
	LastQuery    string
	UserInputBox *UserInputComponent
	Selection    int
	maxColumn    int
	outputBuffer bytes.Buffer
}

func NewCommandBuffer(parent *Preditor,
	cfg *Config,
	root string,
	maxH int32,
	maxW int32,
	zeroLocation rl.Vector2) *CommandBuffer {
	if root == "" {
		root, _ = os.Getwd()
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}
	charSize := measureTextSize(parent.Font, ' ', parent.FontSize, 0)
	ofb := &CommandBuffer{
		cfg:          cfg,
		parent:       parent,
		root:         absRoot,
		keymaps:      []Keymap{CommandBufferKeymap},
		maxHeight:    maxH,
		maxWidth:     maxW,
		ZeroLocation: zeroLocation,
		outputBuffer: bytes.Buffer{},
		maxColumn:    int(maxW / int32(charSize.X)),
		UserInputBox: NewUserInputComponent(parent, cfg, zeroLocation, maxH, maxW),
	}

	return ofb
}

func (f *CommandBuffer) HandleFontChange() {

}

func (f *CommandBuffer) String() string {
	return fmt.Sprintf("Grep Buffer@%s", f.LastQuery)
}

func (f *CommandBuffer) runCommand() error {
	userCommand := string(f.UserInputBox.UserInput)
	executable := strings.SplitN(userCommand, " ", 2)[0]
	args := strings.SplitN(userCommand, " ", 2)[1]
	cmd := exec.Command(executable, args)
	go func() {
		stdErr := &bytes.Buffer{}
		stdOut := &bytes.Buffer{}
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr

		if err := cmd.Run(); err != nil {
			fmt.Println(err.Error())
			return
		}

		io.Copy(&f.outputBuffer, stdOut)
	}()

	return nil
}

func (f *CommandBuffer) Render() {
	charSize := measureTextSize(f.parent.Font, ' ', f.parent.FontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(f.ZeroLocation.X), int32(f.ZeroLocation.Y), f.maxWidth, int32(charSize.Y)*2, f.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(f.parent.Font, string(f.UserInputBox.UserInput), rl.Vector2{
		X: f.ZeroLocation.X, Y: f.ZeroLocation.Y + charSize.Y/2,
	}, float32(f.parent.FontSize), 0, f.cfg.Colors.Foreground)

	switch f.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(f.UserInputBox.Idx), int32(f.ZeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfOutput := int32(f.ZeroLocation.Y) + int32(3*(charSize.Y))

	rl.DrawTextEx(f.parent.Font, f.outputBuffer.String(), rl.Vector2{
		X: f.ZeroLocation.X,
		Y: float32(startOfOutput),
	}, float32(f.parent.FontSize), 0, rl.White)
}

func (f *CommandBuffer) SetMaxWidth(w int32) {
	f.maxWidth = w
}

func (f *CommandBuffer) SetMaxHeight(h int32) {
	f.maxHeight = h
}

func (f *CommandBuffer) GetMaxWidth() int32 {
	return f.maxWidth
}

func (f *CommandBuffer) GetMaxHeight() int32 {
	return f.maxHeight
}

func (f *CommandBuffer) Keymaps() []Keymap {
	return f.keymaps
}

func (f *CommandBuffer) openUserInput() error {

	return nil
}

func makeCommandBufferCommand(f func(e *CommandBuffer) error) Command {
	return func(preditor *Preditor) error {
		return f(preditor.ActiveBuffer().(*CommandBuffer))
	}
}

func init() {
	CommandBufferKeymap = Keymap{

		Key{K: "f", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.CursorRight(1)
		}),
		Key{K: "v", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.paste()
		}),
		Key{K: "c", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.copy()
		}),
		Key{K: "s", Control: true}: makeCommandBufferCommand(func(a *CommandBuffer) error {
			a.keymaps = append(a.keymaps, SearchTextBufferKeymap)
			return nil
		}),
		Key{K: "<esc>"}: makeCommandBufferCommand(func(p *CommandBuffer) error {
			// maybe close ?
			p.parent.Buffers = p.parent.Buffers[:len(p.parent.Buffers)-1]
			p.parent.ActiveBufferIndex = len(p.parent.Buffers) - 1
			return nil
		}),

		Key{K: "a", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.BeginingOfTheLine()
		}),
		Key{K: "e", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.EndOfTheLine()
		}),

		Key{K: "<right>"}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.NextWordStart()
		}),
		Key{K: "<left>"}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.PreviousWord()
		}),

		Key{K: "b", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.CursorLeft(1)
		}),
		Key{K: "<home>"}: makeCommandBufferCommand(func(e *CommandBuffer) error {
			return e.UserInputBox.BeginingOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}:                    makeCommandBufferCommand(func(e *CommandBuffer) error { return e.runCommand() }),
		Key{K: "<space>"}:                    makeCommandBufferCommand(func(e *CommandBuffer) error { return e.UserInputBox.insertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                makeCommandBufferCommand(func(e *CommandBuffer) error { return e.UserInputBox.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: makeCommandBufferCommand(func(e *CommandBuffer) error { return e.UserInputBox.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           makeCommandBufferCommand(func(e *CommandBuffer) error { return e.UserInputBox.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               makeCommandBufferCommand(func(e *CommandBuffer) error { return e.UserInputBox.DeleteWordForward() }),
		Key{K: "<delete>"}:                   makeCommandBufferCommand(func(e *CommandBuffer) error { return e.UserInputBox.DeleteCharForward() }),
		Key{K: "a"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('a') }),
		Key{K: "b"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('b') }),
		Key{K: "c"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('c') }),
		Key{K: "d"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('d') }),
		Key{K: "e"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('e') }),
		Key{K: "f"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('f') }),
		Key{K: "g"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('g') }),
		Key{K: "h"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('h') }),
		Key{K: "i"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('i') }),
		Key{K: "j"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('j') }),
		Key{K: "k"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('k') }),
		Key{K: "l"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('l') }),
		Key{K: "m"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('m') }),
		Key{K: "n"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('n') }),
		Key{K: "o"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('o') }),
		Key{K: "p"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('p') }),
		Key{K: "q"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('q') }),
		Key{K: "r"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('r') }),
		Key{K: "s"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('s') }),
		Key{K: "t"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('t') }),
		Key{K: "u"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('u') }),
		Key{K: "v"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('v') }),
		Key{K: "w"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('w') }),
		Key{K: "x"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('x') }),
		Key{K: "y"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('y') }),
		Key{K: "z"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('z') }),
		Key{K: "0"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('0') }),
		Key{K: "1"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('1') }),
		Key{K: "2"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('2') }),
		Key{K: "3"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('3') }),
		Key{K: "4"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('4') }),
		Key{K: "5"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('5') }),
		Key{K: "6"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('6') }),
		Key{K: "7"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('7') }),
		Key{K: "8"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('8') }),
		Key{K: "9"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('9') }),
		Key{K: "\\"}:                         makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('Z') }),
		Key{K: "["}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('[') }),
		Key{K: "]"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('}') }),
		Key{K: ";"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer(':') }),
		Key{K: "'"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('"') }),
		Key{K: "\""}:                         makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('"') }),
		Key{K: ","}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer(',') }),
		Key{K: "."}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('>') }),
		Key{K: "/"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('?') }),
		Key{K: "-"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('-') }),
		Key{K: "="}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('+') }),
		Key{K: "`"}:                          makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             makeCommandBufferCommand(func(f *CommandBuffer) error { return f.UserInputBox.insertCharAtBuffer('~') }),
	}
}

var CommandBufferKeymap Keymap
