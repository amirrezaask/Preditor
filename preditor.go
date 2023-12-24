package preditor

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/flopp/go-findfont"
	"golang.design/x/clipboard"
	"image/color"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Colors struct {
	Background            color.RGBA
	Foreground            color.RGBA
	Selection             color.RGBA
	StatusBarBackground   color.RGBA
	StatusBarForeground   color.RGBA
	LineNumbersForeground color.RGBA
	Cursor                color.RGBA
	CursorLineBackground  color.RGBA
	SyntaxKeywords        color.RGBA
	SyntaxTypes           color.RGBA
}

type Buffer interface {
	GetID() int
	SetID(int)
	Render(zeroLocation rl.Vector2, maxHeight int32, maxWidth int32)
	Keymaps() []Keymap
	fmt.Stringer
}
type BaseBuffer struct {
	ID int
}

func (b BaseBuffer) GetID() int {
	return b.ID
}
func (b *BaseBuffer) SetID(i int) {
	b.ID = i
}

type Preditor struct {
	CWD             string
	Cfg             *Config
	ScratchBufferID int
	MessageBufferID int
	Buffers         map[int]Buffer
	ActiveBufferID  int
	GlobalKeymap    Keymap
	GlobalVariables Variables
	Commands        Commands
	FontPath        string
	Font            rl.Font
	FontSize        int32
	OSWindowHeight  int32
	OSWindowWidth   int32
}

var charSizeCache = map[byte]rl.Vector2{} //TODO: if font size or font changes this is fucked
func measureTextSize(font rl.Font, s byte, size int32, spacing float32) rl.Vector2 {
	if charSize, exists := charSizeCache[s]; exists {
		return charSize
	}
	charSize := rl.MeasureTextEx(font, string(s), float32(size), spacing)
	charSizeCache[s] = charSize
	return charSize
}
func (p *Preditor) WriteMessage(msg string) {
	p.GetBuffer(p.MessageBufferID).(*TextBuffer).Content = append(p.GetBuffer(p.MessageBufferID).(*TextBuffer).Content, []byte(fmt.Sprintln(msg))...)
}
func (p *Preditor) getCWD() string {
	if tb, isTextBuffer := p.ActiveBuffer().(*TextBuffer); isTextBuffer && !tb.IsSpecial() {
		return path.Dir(tb.File)
	} else {
		return p.CWD
	}
}
func (p *Preditor) AddBuffer(b Buffer) {
	id := len(p.Buffers) + 1
	p.Buffers[id] = b
	b.SetID(id)
}
func (p *Preditor) MarkBufferAsActive(id int) {
	p.ActiveBufferID = id
}

func (p *Preditor) GetBuffer(id int) Buffer {
	return p.Buffers[id]
}

func (e *Preditor) KillBuffer(id int) {
	delete(e.Buffers, id)

	for k := range e.Buffers {
		e.ActiveBufferID = k
	}
}

func (p *Preditor) LoadFont(name string, size int32) error {
	var err error
	p.FontPath, err = findfont.Find(name + ".ttf")
	if err != nil {
		return err
	}

	p.FontSize = size
	p.Font = rl.LoadFontEx(p.FontPath, p.FontSize, nil)
	return nil
}

func (p *Preditor) IncreaseFontSize(n int) {
	p.FontSize += int32(n)
	p.Font = rl.LoadFontEx(p.FontPath, p.FontSize, nil)
	charSizeCache = map[byte]rl.Vector2{}
}

func (p *Preditor) DecreaseFontSize(n int) {
	p.FontSize -= int32(n)
	p.Font = rl.LoadFontEx(p.FontPath, p.FontSize, nil)
	charSizeCache = map[byte]rl.Vector2{}

}

func (e *Preditor) ActiveBuffer() Buffer {
	return e.Buffers[e.ActiveBufferID]
}

type Command func(*Preditor) error
type Variables map[string]any
type Key struct {
	Control bool
	Alt     bool
	Shift   bool
	Super   bool
	K       string
}

func (k Key) IsEmpty() bool {
	return k.K == ""
}

type Keymap map[Key]Command
type Commands map[string]Command
type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("Line: %d Column:%d\n", p.Line, p.Column)
}

func parseHexColor(v string) (out color.RGBA, err error) {
	if len(v) != 7 {
		return out, errors.New("hex color must be 7 characters")
	}
	if v[0] != '#' {
		return out, errors.New("hex color must start with '#'")
	}
	var red, redError = strconv.ParseUint(v[1:3], 16, 8)
	if redError != nil {
		return out, errors.New("red component invalid")
	}
	out.R = uint8(red)
	var green, greenError = strconv.ParseUint(v[3:5], 16, 8)
	if greenError != nil {
		return out, errors.New("green component invalid")
	}
	out.G = uint8(green)
	var blue, blueError = strconv.ParseUint(v[5:7], 16, 8)
	if blueError != nil {
		return out, errors.New("blue component invalid")
	}
	out.B = uint8(blue)
	out.A = 255
	return
}

func (e *Preditor) HandleKeyEvents() {
	key := getKey()
	if !key.IsEmpty() {
		keymaps := append([]Keymap{e.GlobalKeymap}, e.ActiveBuffer().Keymaps()...)
		for i := len(keymaps) - 1; i >= 0; i-- {
			cmd := keymaps[i][key]
			if cmd != nil {
				cmd(e)
				break
			}
		}
	}

}

func (e *Preditor) Render() {
	rl.BeginDrawing()
	rl.ClearBackground(e.Cfg.Colors.Background)
	e.ActiveBuffer().Render(rl.Vector2{}, e.OSWindowHeight, e.OSWindowWidth)
	rl.EndDrawing()
}

func (e *Preditor) HandleWindowResize() {
	e.OSWindowHeight = int32(rl.GetRenderHeight())
	e.OSWindowWidth = int32(rl.GetRenderWidth())
}

func (e *Preditor) HandleMouseEvents() {
	key := getMouseKey()
	if !key.IsEmpty() {
		keymaps := append([]Keymap{e.GlobalKeymap}, e.ActiveBuffer().Keymaps()...)
		for i := len(keymaps) - 1; i >= 0; i-- {
			cmd := keymaps[i][key]
			if cmd != nil {
				cmd(e)
				break
			}
		}
	}
}

type modifierKeyState struct {
	control bool
	alt     bool
	shift   bool
	super   bool
}

func getModifierKeyState() modifierKeyState {
	state := modifierKeyState{}
	if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
		state.control = true
	}
	if rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt) {
		state.alt = true
	}
	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		state.shift = true
	}
	if rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper) {
		state.super = true
	}

	return state
}

func getKey() Key {
	modifierState := getModifierKeyState()
	key := getKeyPressedString()

	k := Key{
		Control: modifierState.control,
		Alt:     modifierState.alt,
		Super:   modifierState.super,
		Shift:   modifierState.shift,
		K:       key,
	}
	// if !k.IsEmpty() {
	// 	fmt.Println("=================================")
	// 	fmt.Printf("key: %+v\n", k)
	// 	fmt.Println("=================================")
	// }

	return k
}

func getMouseKey() Key {
	modifierState := getModifierKeyState()
	var key string
	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		key = "<lmouse>-click"
	case rl.IsMouseButtonPressed(rl.MouseButtonMiddle):
		key = "<mmouse>-click"
	case rl.IsMouseButtonPressed(rl.MouseButtonRight):
		key = "<rmouse>-click"
	case rl.IsMouseButtonDown(rl.MouseButtonLeft):
		key = "<lmouse>-hold"

	case rl.IsMouseButtonDown(rl.MouseButtonMiddle):
		key = "<mmouse>-hold"

	case rl.IsMouseButtonDown(rl.MouseButtonRight):
		key = "<rmouse>-hold"
	}

	if wheel := rl.GetMouseWheelMoveV(); wheel.X != 0 || wheel.Y != 0 {
		if wheel.Y != 0 {
			if wheel.Y < 0 {
				key = "<mouse-wheel-down>"
			}
			if wheel.Y > 0 {
				key = "<mouse-wheel-up>"
			}

		}
	}

	if key == "" {
		return Key{}
	}

	k := Key{
		Control: modifierState.control,
		Alt:     modifierState.alt,
		Super:   modifierState.super,
		Shift:   modifierState.shift,
		K:       key,
	}
	// if !k.IsEmpty() {
	// 	fmt.Println("=================================")
	// 	fmt.Printf("key: %+v\n", k)
	// 	fmt.Println("=================================")
	// }

	return k

}

func isPressed(key int32) bool {
	return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
}

func getKeyPressedString() string {
	switch {
	case isPressed(rl.KeyApostrophe):
		return "'"
	case isPressed(rl.KeySpace):
		return "<space>"
	case isPressed(rl.KeyEscape):
		return "<esc>"
	case isPressed(rl.KeyEnter):
		return "<enter>"
	case isPressed(rl.KeyTab):
		return "<tab>"
	case isPressed(rl.KeyBackspace):
		return "<backspace>"
	case isPressed(rl.KeyInsert):
		return "<insert>"
	case isPressed(rl.KeyDelete):
		return "<delete>"
	case isPressed(rl.KeyRight):
		return "<right>"
	case isPressed(rl.KeyLeft):
		return "<left>"
	case isPressed(rl.KeyDown):
		return "<down>"
	case isPressed(rl.KeyUp):
		return "<up>"
	case isPressed(rl.KeyPageUp):
		return "<pageup>"
	case isPressed(rl.KeyPageDown):
		return "<pagedown>"
	case isPressed(rl.KeyHome):
		return "<home>"
	case isPressed(rl.KeyEnd):
		return "<end>"
	case isPressed(rl.KeyCapsLock):
		return "<capslock>"
	case isPressed(rl.KeyScrollLock):
		return "<scrolllock>"
	case isPressed(rl.KeyNumLock):
		return "<numlock>"
	case isPressed(rl.KeyPrintScreen):
		return "<printscreen>"
	case isPressed(rl.KeyPause):
		return "<pause>"
	case isPressed(rl.KeyF1):
		return "<f1>"
	case isPressed(rl.KeyF2):
		return "<f2>"
	case isPressed(rl.KeyF3):
		return "<f3>"
	case isPressed(rl.KeyF4):
		return "<f4>"
	case isPressed(rl.KeyF5):
		return "<f5>"
	case isPressed(rl.KeyF6):
		return "<f6>"
	case isPressed(rl.KeyF7):
		return "<f7>"
	case isPressed(rl.KeyF8):
		return "<f8>"
	case isPressed(rl.KeyF9):
		return "<f9>"
	case isPressed(rl.KeyF10):
		return "<f10>"
	case isPressed(rl.KeyF11):
		return "<f11>"
	case isPressed(rl.KeyF12):
		return "<f12>"
	case isPressed(rl.KeyLeftBracket):
		return "["
	case isPressed(rl.KeyBackSlash):
		return "\\"
	case isPressed(rl.KeyRightBracket):
		return "]"
	case isPressed(rl.KeyKp0):
		return "0"
	case isPressed(rl.KeyKp1):
		return "1"
	case isPressed(rl.KeyKp2):
		return "2"
	case isPressed(rl.KeyKp3):
		return "3"
	case isPressed(rl.KeyKp4):
		return "4"
	case isPressed(rl.KeyKp5):
		return "5"
	case isPressed(rl.KeyKp6):
		return "6"
	case isPressed(rl.KeyKp7):
		return "7"
	case isPressed(rl.KeyKp8):
		return "8"
	case isPressed(rl.KeyKp9):
		return "9"
	case isPressed(rl.KeyKpDecimal):
		return "."
	case isPressed(rl.KeyKpDivide):
		return "/"
	case isPressed(rl.KeyKpMultiply):
		return "*"
	case isPressed(rl.KeyKpSubtract):
		return "-"
	case isPressed(rl.KeyKpAdd):
		return "+"
	case isPressed(rl.KeyKpEnter):
		return "<enter>"
	case isPressed(rl.KeyKpEqual):
		return "="
	case isPressed(rl.KeyApostrophe):
		return "'"
	case isPressed(rl.KeyComma):
		return ","
	case isPressed(rl.KeyMinus):
		return "-"
	case isPressed(rl.KeyPeriod):
		return "."
	case isPressed(rl.KeySlash):
		return "/"
	case isPressed(rl.KeyZero):
		return "0"
	case isPressed(rl.KeyOne):
		return "1"
	case isPressed(rl.KeyTwo):
		return "2"
	case isPressed(rl.KeyThree):
		return "3"
	case isPressed(rl.KeyFour):
		return "4"
	case isPressed(rl.KeyFive):
		return "5"
	case isPressed(rl.KeySix):
		return "6"
	case isPressed(rl.KeySeven):
		return "7"
	case isPressed(rl.KeyEight):
		return "8"
	case isPressed(rl.KeyNine):
		return "9"
	case isPressed(rl.KeySemicolon):
		return ";"
	case isPressed(rl.KeyEqual):
		return "="
	case isPressed(rl.KeyA):
		return "a"
	case isPressed(rl.KeyB):
		return "b"
	case isPressed(rl.KeyC):
		return "c"
	case isPressed(rl.KeyD):
		return "d"
	case isPressed(rl.KeyE):
		return "e"
	case isPressed(rl.KeyF):
		return "f"
	case isPressed(rl.KeyG):
		return "g"
	case isPressed(rl.KeyH):
		return "h"
	case isPressed(rl.KeyI):
		return "i"
	case isPressed(rl.KeyJ):
		return "j"
	case isPressed(rl.KeyK):
		return "k"
	case isPressed(rl.KeyL):
		return "l"
	case isPressed(rl.KeyM):
		return "m"
	case isPressed(rl.KeyN):
		return "n"
	case isPressed(rl.KeyO):
		return "o"
	case isPressed(rl.KeyP):
		return "p"
	case isPressed(rl.KeyQ):
		return "q"
	case isPressed(rl.KeyR):
		return "r"
	case isPressed(rl.KeyS):
		return "s"
	case isPressed(rl.KeyT):
		return "t"
	case isPressed(rl.KeyU):
		return "u"
	case isPressed(rl.KeyV):
		return "v"
	case isPressed(rl.KeyW):
		return "w"
	case isPressed(rl.KeyX):
		return "x"
	case isPressed(rl.KeyY):
		return "y"
	case isPressed(rl.KeyZ):
		return "z"
	default:
		return ""
	}
}

func setupRaylib(cfg *Config) {
	// basic setup
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(1920, 1080, "Preditor")
	rl.SetTargetFPS(120)
	rl.SetTextLineSpacing(cfg.FontSize)
	rl.SetExitKey(0)
}

func New() (*Preditor, error) {
	var configPath string
	flag.StringVar(&configPath, "cfg", path.Join(os.Getenv("HOME"), ".preditor"), "path to config file, defaults to: ~/.preditor")
	flag.Parse()

	// read config file
	cfg, err := ReadConfig(configPath)
	if err != nil {
		panic(err)
	}

	// create editor
	setupRaylib(cfg)
	initFileTypes(cfg.Colors)

	if err := clipboard.Init(); err != nil {
		panic(err)
	}
	p := &Preditor{
		Cfg:            cfg,
		Buffers:        map[int]Buffer{},
		OSWindowHeight: int32(rl.GetRenderHeight()),
		OSWindowWidth:  int32(rl.GetRenderWidth()),
	}

	err = p.LoadFont(cfg.FontName, int32(cfg.FontSize))
	if err != nil {
		return nil, err
	}
	scratch, err := NewTextBuffer(p, p.Cfg, "*Scratch*")
	if err != nil {
		return nil, err
	}
	message, err := NewTextBuffer(p, p.Cfg, "*Messages*")
	if err != nil {
		return nil, err
	}

	p.AddBuffer(scratch)
	p.AddBuffer(message)

	p.MessageBufferID = message.ID
	p.ScratchBufferID = scratch.ID

	p.MarkBufferAsActive(scratch.ID)
	p.GlobalKeymap = GlobalKeymap

	// handle command line argument
	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
		if filename == "-" {
			//stdin
			tb, err := NewTextBuffer(p, cfg, "stdin")
			if err != nil {
				panic(err)
			}

			p.AddBuffer(tb)
			p.MarkBufferAsActive(tb.ID)
			tb.Readonly = true
			go func() {
				r := bufio.NewReader(os.Stdin)
				for {
					b, err := r.ReadByte()
					if err != nil {
						p.WriteMessage(err.Error())
						break
					}
					tb.Content = append(tb.Content, b)
				}
			}()
		} else {
			err = SwitchOrOpenFileInTextBuffer(p, cfg, filename, nil)
			if err != nil {
				panic(err)
			}
		}
	}

	return p, nil
}

func (p *Preditor) StartMainLoop() {
	defer func() {

		if r := recover(); r != nil {
			err := os.WriteFile(path.Join(os.Getenv("HOME"),
				fmt.Sprintf("preditor-crashlog-%d", time.Now().Unix())),
				[]byte(fmt.Sprintf("%v\n%s\n%s", r, string(debug.Stack()), spew.Sdump(p))), 0644)
			if err != nil {
				fmt.Println("we are doomed")
				fmt.Println(err)
			}

			fmt.Printf("%v\n%s\n%s\n", r, string(debug.Stack()), spew.Sdump(p))
		}

	}()

	for !rl.WindowShouldClose() {
		p.HandleWindowResize()
		p.HandleMouseEvents()
		p.HandleKeyEvents()
		p.Render()
	}
}

func (p *Preditor) MaxHeightToMaxLine(maxH int32) int32 {
	return maxH / int32(measureTextSize(p.Font, ' ', p.FontSize, 0).Y)
}
func (p *Preditor) MaxWidthToMaxColumn(maxW int32) int32 {
	return maxW / int32(measureTextSize(p.Font, ' ', p.FontSize, 0).X)
}

func (e *Preditor) openFileBuffer() {
	ofb := NewFilePickerBuffer(e, e.Cfg, e.getCWD())
	e.AddBuffer(ofb)
	e.MarkBufferAsActive(ofb.ID)
}

func (e *Preditor) openFuzzyFilePicker() {
	ofb := NewFuzzyFilePickerBuffer(e, e.Cfg, e.getCWD())
	e.AddBuffer(ofb)
	e.MarkBufferAsActive(ofb.ID)

}
func (e *Preditor) openBufferSwitcher() {
	ofb := NewBufferSwitcherBuffer(e, e.Cfg)
	e.AddBuffer(ofb)
	e.MarkBufferAsActive(ofb.ID)

}

func (e *Preditor) openGrepBuffer() {
	ofb := NewGrepBuffer(e, e.Cfg, e.getCWD())
	e.AddBuffer(ofb)
	e.MarkBufferAsActive(ofb.ID)
}

func handlePanicAndWriteMessage(p *Preditor) {
	r := recover()
	if r != nil {
		msg := fmt.Sprintf("%v\n%s", r, string(debug.Stack()))
		fmt.Println(msg)
		p.WriteMessage(msg)
	}
}
