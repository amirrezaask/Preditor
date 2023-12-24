package preditor

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/flopp/go-findfont"
	"golang.design/x/clipboard"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RGBA color.RGBA

func (r RGBA) String() string {
	colorAsHex := fmt.Sprintf("#%02x%02x%02x%02x", r.R, r.G, r.B, r.A)
	return colorAsHex
}

func (r RGBA) ToColorRGBA() color.RGBA {
	return color.RGBA(r)
}

type Colors struct {
	Background            RGBA
	Foreground            RGBA
	Selection             RGBA
	Prompts               RGBA
	StatusBarBackground   RGBA
	StatusBarForeground   RGBA
	LineNumbersForeground RGBA
	ActiveWindowBorder    RGBA
	Cursor                RGBA
	CursorLineBackground  RGBA
	SyntaxKeywords        RGBA
	SyntaxTypes           RGBA
	SyntaxComments        RGBA
	SyntaxStrings         RGBA
}

type Drawable interface {
	GetID() int
	SetID(int)
	Render(zeroLocation rl.Vector2, maxHeight float64, maxWidth float64)
	Keymaps() []Keymap
	fmt.Stringer
}
type BaseDrawable struct {
	ID int
}

func (b BaseDrawable) GetID() int {
	return b.ID
}
func (b *BaseDrawable) SetID(i int) {
	b.ID = i
}

type Window struct {
	ID            int
	BufferID      int
	ZeroLocationX float64
	ZeroLocationY float64
	Width         float64
	Height        float64
}

type Overlay struct {
	Active   bool
	BufferID int
}

func (w *Window) Render(c *Context, zeroLocation rl.Vector2, maxHeight float64, maxWidth float64) {
	if buf := c.Buffers[w.BufferID]; buf != nil {
		buf.Render(zeroLocation, maxHeight, maxWidth)
	}
}

type Prompt struct {
	IsActive   bool
	Text       string
	UserInput  string
	Keymap     Keymap
	ChangeHook func(userInput string, c *Context) error
	DoneHook   func(userInput string, c *Context) error
}

type Context struct {
	CWD               string
	Cfg               *Config
	ScratchBufferID   int
	MessageBufferID   int
	Buffers           map[int]Drawable
	GlobalKeymap      Keymap
	GlobalVariables   Variables
	Commands          Commands
	FontData          []byte
	Font              rl.Font
	FontSize          int32
	OSWindowHeight    float64
	OSWindowWidth     float64
	Windows           [][]*Window
	Prompt            Prompt
	BottomOverlay     Overlay
	ActiveWindowIndex int
}

func (c *Context) CloseBottomOverlay() {
	c.BottomOverlay.Active = false
}
func (c *Context) OpenBottomOverlay() {
	c.BottomOverlay.Active = true
}
func (c *Context) SetPrompt(text string,
	changeHook func(userInput string, c *Context) error,
	doneHook func(userInput string, c *Context) error, keymap *Keymap) {
	c.Prompt.IsActive = true
	c.Prompt.Text = text
	c.Prompt.ChangeHook = changeHook
	c.Prompt.DoneHook = doneHook
	if keymap != nil {
		c.Prompt.Keymap = *keymap
	} else {
		c.Prompt.Keymap = DefaultPromptKeymap
	}
}

func (c *Context) ActiveWindow() *Window {
	w := c.GetWindow(c.ActiveWindowIndex)
	return w
}

func (c *Context) ActiveBuffer() Drawable {
	if win := c.GetWindow(c.ActiveWindowIndex); win != nil {
		bufferid := win.BufferID
		return c.Buffers[bufferid]
	}

	return nil
}
func (c *Context) ActiveBufferID() int {
	if win := c.GetWindow(c.ActiveWindowIndex); win != nil {
		bufferid := win.BufferID
		return bufferid
	}

	return -1
}

var charSizeCache = map[byte]rl.Vector2{}

func measureTextSize(font rl.Font, s byte, size int32, spacing float32) rl.Vector2 {
	if charSize, exists := charSizeCache[s]; exists {
		return charSize
	}
	charSize := rl.MeasureTextEx(font, string(s), float32(size), spacing)
	charSizeCache[s] = charSize
	return charSize
}

func (c *Context) WriteMessage(msg string) {
	c.GetBuffer(c.MessageBufferID).(*Buffer).Content = append(c.GetBuffer(c.MessageBufferID).(*Buffer).Content, []byte(fmt.Sprintln(msg))...)
}

func (c *Context) getCWD() string {
	if tb, isTextBuffer := c.ActiveBuffer().(*Buffer); isTextBuffer && !tb.IsSpecial() {
		return path.Dir(tb.File)
	} else {
		return c.CWD
	}
}
func (c *Context) AddBuffer(b Drawable) {
	id := len(c.Buffers) + 1
	c.Buffers[id] = b
	b.SetID(id)
}

func (c *Context) windowCount() int {
	var count int
	for _, col := range c.Windows {
		for range col {
			count++
		}
	}

	return count
}

func (c *Context) AddWindowInANewColumn(w *Window) {
	c.Windows = append(c.Windows, []*Window{w})
	w.ID = c.windowCount()
}

func (c *Context) AddWindowInANewColumnAndSwitchToIt(w *Window) {
	c.Windows = append(c.Windows, []*Window{w})
	w.ID = c.windowCount()
	c.ActiveWindowIndex = w.ID
}

func (c *Context) AddWindowInCurrentColumn(w *Window) {
	currentColIndex := -1
HERE:
	for i, col := range c.Windows {
		for _, win := range col {
			if win.ID == c.ActiveWindowIndex {
				currentColIndex = i
				break HERE
			}
		}
	}

	if currentColIndex != -1 {
		c.Windows[currentColIndex] = append(c.Windows[currentColIndex], w)
		w.ID = c.windowCount()
	}
}

func (c *Context) MarkWindowAsActive(id int) {
	c.ActiveWindowIndex = id
}

func (c *Context) MarkBufferAsActive(id int) {
	c.ActiveWindow().BufferID = id
}

func (c *Context) GetBuffer(id int) Drawable {
	return c.Buffers[id]
}

func (c *Context) KillBuffer(id int) {
	delete(c.Buffers, id)
	for _, buf := range c.Buffers {
		bufid := buf.GetID()
		c.ActiveWindow().BufferID = bufid
	}
}

func (c *Context) LoadFont(name string, size int32) error {
	switch strings.ToLower(name) {
	case "liberationmono-regular":
		c.FontSize = size
		c.FontData = liberationMonoRegularTTF
	case "jetbrainsmono":
		c.FontSize = size
		c.FontData = jetbrainsMonoTTF
	default:
		var err error
		path, err := findfont.Find(name + ".ttf")
		if err != nil {
			return err
		}
		c.FontData, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}

	c.FontSize = size
	c.Font = rl.LoadFontFromMemory(".ttf", c.FontData, int32(len(c.FontData)), c.FontSize, nil, 0)
	return nil
}

func (c *Context) IncreaseFontSize(n int) {
	c.FontSize += int32(n)
	c.Font = rl.LoadFontFromMemory(".ttf", c.FontData, int32(len(c.FontData)), c.FontSize, nil, 0)
	charSizeCache = map[byte]rl.Vector2{}
}

func (c *Context) DecreaseFontSize(n int) {
	c.FontSize -= int32(n)
	c.Font = rl.LoadFontFromMemory(".ttf", c.FontData, int32(len(c.FontData)), c.FontSize, nil, 0)
	charSizeCache = map[byte]rl.Vector2{}

}

type Command func(*Context) error
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

func (k Keymap) Clone() Keymap {
	cloned := Keymap{}
	for i, v := range k {
		cloned[i] = v
	}

	return cloned
}

func (k Keymap) SetKeyCommand(key Key, command Command) {
	k[key] = command
}

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

func (c *Context) HandleKeyEvents() {
	defer handlePanicAndWriteMessage(c)
	key := getKey()
	if !key.IsEmpty() {
		if c.BottomOverlay.Active {
			keymaps := []Keymap{c.GlobalKeymap}
			if buf := c.GetBuffer(c.BottomOverlay.BufferID); buf != nil {
				keymaps = append(keymaps, buf.Keymaps()...)
			}
			for i := len(keymaps) - 1; i >= 0; i-- {
				cmd := keymaps[i][key]
				if cmd != nil {
					if err := cmd(c); err != nil {
						c.WriteMessage(err.Error())
					}
					break
				}
			}
			return
		}
		keymaps := []Keymap{c.GlobalKeymap}
		if c.ActiveBuffer() != nil {
			keymaps = append(keymaps, c.ActiveBuffer().Keymaps()...)
		}
		if c.Prompt.IsActive {
			keymaps = append(keymaps, c.Prompt.Keymap)
		}
		for i := len(keymaps) - 1; i >= 0; i-- {
			cmd := keymaps[i][key]
			if cmd != nil {
				cmd(c)
				break
			}
		}
	}

}

func (c *Context) GetWindow(id int) *Window {
	for _, col := range c.Windows {
		for _, win := range col {
			if win.ID == id {
				return win
			}
		}
	}

	return nil
}

func isVisibleInWindow(posX float64, posY float64, zeroLocation rl.Vector2, maxH float64, maxW float64) bool {
	return float32(posX) >= zeroLocation.X &&
		float64(posX) <= (float64(zeroLocation.X)+maxW) &&
		float64(posY) >= float64(zeroLocation.Y) &&
		float64(posY) <= float64(zeroLocation.Y)+maxH
}

func (c *Context) Render() {
	rl.BeginDrawing()
	rl.ClearBackground(c.Cfg.CurrentThemeColors().Background.ToColorRGBA())
	height := c.OSWindowHeight
	charsize := measureTextSize(c.Font, ' ', c.FontSize, 0)
	if c.Prompt.IsActive {
		height -= float64(charsize.Y)
	}
	for i, column := range c.Windows {
		columnWidth := c.OSWindowWidth / float64(len(c.Windows))
		columnZeroX := float64(i) * float64(columnWidth)
		for j, win := range column {
			if win == nil {
				continue
			}
			winHeight := height / float64(len(column))
			winZeroY := float64(j) * winHeight
			zeroLocation := rl.Vector2{X: float32(columnZeroX), Y: float32(winZeroY)}
			win.Width = columnWidth
			win.Height = winHeight
			win.ZeroLocationX = float64(zeroLocation.X)
			win.ZeroLocationY = float64(zeroLocation.Y)
			win.Render(c, zeroLocation, winHeight, columnWidth)
			if c.ActiveWindowIndex == win.ID {
				rl.DrawRectangleLines(int32(columnZeroX), int32(winZeroY), int32(columnWidth), int32(winHeight), c.Cfg.CurrentThemeColors().ActiveWindowBorder.ToColorRGBA())
			}

		}

	}
	if c.BottomOverlay.Active {
		bottomOverlayZerolocationY := c.OSWindowHeight - (c.OSWindowHeight * c.Cfg.BottomOverlayHeight)
		rl.DrawRectangle(0, int32(bottomOverlayZerolocationY), int32(c.OSWindowWidth), int32(c.OSWindowHeight), c.Cfg.CurrentThemeColors().Background.ToColorRGBA())
		rl.DrawRectangleLines(0, int32(bottomOverlayZerolocationY), int32(c.OSWindowWidth), int32(c.OSWindowHeight), rl.Red)
		if buf := c.GetBuffer(c.BottomOverlay.BufferID); buf != nil {
			buf.Render(rl.Vector2{X: 0, Y: float32(bottomOverlayZerolocationY)}, c.OSWindowHeight, c.OSWindowWidth)
		}
	}

	if c.Prompt.IsActive {
		rl.DrawRectangle(0, int32(height), int32(c.OSWindowWidth), int32(charsize.Y), c.Cfg.CurrentThemeColors().Prompts.ToColorRGBA())
		rl.DrawTextEx(c.Font, fmt.Sprintf("%s: %s", c.Prompt.Text, c.Prompt.UserInput), rl.Vector2{
			X: 0,
			Y: float32(height),
		}, float32(c.FontSize), 0, rl.White)
	}
	rl.DrawTextEx(c.Font, fmt.Sprint(rl.GetFPS()), rl.Vector2{
		X: float32(c.OSWindowWidth - float64(charsize.X*3)),
		Y: float32(c.OSWindowHeight - float64(charsize.Y)),
	}, float32(c.FontSize), 0, rl.Red)
	rl.EndDrawing()
}

func (c *Context) HandleWindowResize() {
	c.OSWindowHeight = float64(rl.GetRenderHeight())
	c.OSWindowWidth = float64(rl.GetRenderWidth())
}

func (c *Context) HandleMouseEvents() {
	defer handlePanicAndWriteMessage(c)

	key := getMouseKey()
	if !key.IsEmpty() {
		if c.BottomOverlay.Active {
			keymaps := []Keymap{c.GlobalKeymap}
			if buf := c.GetBuffer(c.BottomOverlay.BufferID); buf != nil {
				keymaps = append(keymaps, buf.Keymaps()...)
				for i := len(keymaps) - 1; i >= 0; i-- {
					cmd := keymaps[i][key]
					if cmd != nil {
						if err := cmd(c); err != nil {
							c.WriteMessage(err.Error())
						}
						break
					}
				}
			}
			return
		}

		//first check if mouse position is in the window context otherwise switch
		pos := rl.GetMousePosition()
		win := c.GetWindow(c.ActiveWindowIndex)
		if float64(pos.X) < win.ZeroLocationX ||
			float64(pos.Y) < win.ZeroLocationY ||
			float64(pos.X) > win.ZeroLocationX+win.Width ||
			float64(pos.Y) > win.ZeroLocationY+win.Height {
			for _, col := range c.Windows {
				for _, win := range col {
					if float64(pos.X) >= win.ZeroLocationX &&
						float64(pos.Y) >= win.ZeroLocationY &&
						float64(pos.X) <= win.ZeroLocationX+win.Width &&
						float64(pos.Y) <= win.ZeroLocationY+win.Height {
						c.ActiveWindowIndex = win.ID
						break
					}
				}
			}
		}

		keymaps := []Keymap{c.GlobalKeymap}
		if c.ActiveBuffer() != nil {
			keymaps = append(keymaps, c.ActiveBuffer().Keymaps()...)
		}
		for i := len(keymaps) - 1; i >= 0; i-- {
			cmd := keymaps[i][key]
			if cmd != nil {
				if err := cmd(c); err != nil {
					c.WriteMessage(err.Error())
				}
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

	return k

}

func isPressed(key int32) bool {
	return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
}

func getKeyPressedString() string {
	switch {
	case isPressed(rl.KeyGrave):
		return "`"
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
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagWindowMaximized | rl.FlagVsyncHint)
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(1920, 1080, "Preditor")
	rl.SetTargetFPS(60)
	rl.SetTextLineSpacing(cfg.FontSize)
	rl.SetExitKey(0)
}

func New() (*Context, error) {
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
	initFileTypes(*cfg.CurrentThemeColors())

	if err := clipboard.Init(); err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	p := &Context{
		Cfg:            cfg,
		CWD:            wd,
		Buffers:        map[int]Drawable{},
		OSWindowHeight: float64(rl.GetRenderHeight()),
		OSWindowWidth:  float64(rl.GetRenderWidth()),
		Windows:        [][]*Window{},
	}

	err = p.LoadFont(cfg.FontName, int32(cfg.FontSize))
	if err != nil {
		return nil, err
	}
	scratch, err := NewBuffer(p, p.Cfg, "*Scratch*")
	if err != nil {
		return nil, err
	}
	message, err := NewBuffer(p, p.Cfg, "*Messages*")
	if err != nil {
		return nil, err
	}
	message.Readonly = true
	message.Content = append(message.Content, []byte(fmt.Sprintf("Loaded Configuration:\n%s\n", cfg))...)

	p.AddBuffer(scratch)
	p.AddBuffer(message)

	p.MessageBufferID = message.ID
	p.ScratchBufferID = scratch.ID

	mainWindow := Window{}
	p.AddWindowInANewColumn(&mainWindow)

	p.MarkWindowAsActive(mainWindow.ID)

	p.MarkBufferAsActive(scratch.ID)

	p.GlobalKeymap = GlobalKeymap

	// handle command line argument
	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
		if filename == "-" {
			//stdin
			tb, err := NewBuffer(p, cfg, "stdin")
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

func (c *Context) StartMainLoop() {
	defer func() {
		if r := recover(); r != nil {
			err := os.WriteFile(path.Join(os.Getenv("HOME"),
				fmt.Sprintf("preditor-crashlog-%d", time.Now().Unix())),
				[]byte(fmt.Sprintf("%v\n%s\n%s", r, string(debug.Stack()), spew.Sdump(c))), 0644)
			if err != nil {
				fmt.Println("we are doomed")
				fmt.Println(err)
			}

			fmt.Printf("%v\n%s\n%s\n", r, string(debug.Stack()), spew.Sdump(c))
		}
	}()

	for !rl.WindowShouldClose() {
		c.HandleWindowResize()
		c.HandleMouseEvents()
		c.HandleKeyEvents()
		c.Render()
	}
}

func MakeCommand[T Drawable](f func(t T) error) Command {
	return func(c *Context) error {
		if c.BottomOverlay.Active {
			if buf := c.GetBuffer(c.BottomOverlay.BufferID); buf != nil {
				return f(buf.(T))
			}
		}
		return f(c.ActiveBuffer().(T))

	}

}

func (c *Context) MaxHeightToMaxLine(maxH int32) int32 {
	return maxH / int32(measureTextSize(c.Font, ' ', c.FontSize, 0).Y)
}
func (c *Context) MaxWidthToMaxColumn(maxW int32) int32 {
	return maxW / int32(measureTextSize(c.Font, ' ', c.FontSize, 0).X)
}

func (c *Context) openFileBuffer() {
	ofb := NewInteractiveFilePicker(c, c.Cfg, c.getCWD())
	c.AddBuffer(ofb)
	c.MarkBufferAsActive(ofb.ID)
}

func (c *Context) openFuzzyFilePicker() {
	ofb := NewInteractiveFuzzyFile(c, c.Cfg, c.getCWD())
	c.AddBuffer(ofb)
	c.MarkBufferAsActive(ofb.ID)

}
func (c *Context) openBufferSwitcher() {
	ofb := NewBufferSwitcher(c, c.Cfg)
	c.AddBuffer(ofb)
	c.MarkBufferAsActive(ofb.ID)
}

func (c *Context) openThemeSwitcher() {
	ofb := NewThemeSwitcher(c, c.Cfg)
	c.AddBuffer(ofb)
	c.MarkBufferAsActive(ofb.ID)
}

func (c *Context) openGrepBuffer() {

	ofb := NewInteractiveGrep(c, c.Cfg, c.getCWD())
	c.AddBuffer(ofb)
	c.MarkBufferAsActive(ofb.ID)
}

func (c *Context) openCompilationBuffer(command string) error {
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddBuffer(cb)
	c.MarkBufferAsActive(cb.ID)

	return nil
}

func (c *Context) openCompilationBufferInAVSplit(command string) error {
	c.VSplit()
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddBuffer(cb)

	return nil
}

func (c *Context) openCompilationBufferInCompilationPanel(command string) error {
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	cb.keymaps[0].SetKeyCommand(Key{K: "<esc>"}, func(context *Context) error {
		context.CloseBottomOverlay()
		return nil
	})
	c.AddBuffer(cb)
	c.BottomOverlay.BufferID = cb.ID
	c.BottomOverlay.Active = true

	return nil
}

func (c *Context) VSplit() {
	win := &Window{}
	c.AddWindowInANewColumn(win)
}

func (c *Context) HSplit() {
	win := &Window{}
	c.AddWindowInCurrentColumn(win)
}

func (c *Context) OtherWindow() {
	for i, col := range c.Windows {
		for j, win := range col {
			if win.ID == c.ActiveWindowIndex {
				if j+1 < len(col) {
					c.ActiveWindowIndex = col[j+1].ID
					return
				} else {
					if i+1 < len(c.Windows) {
						c.ActiveWindowIndex = c.Windows[i+1][0].ID
						return

					} else {
						c.ActiveWindowIndex = c.Windows[0][0].ID
						return

					}
				}
			}
		}
	}
}

func removeSliceIndex[T any](s []T, i int) []T {
	if i == len(s)-1 {
		s = s[:i]
	} else {
		s = append(s[:i], s[i+1:]...)
	}

	return s
}

func (c *Context) CloseWindow(id int) {
	if c.windowCount() < 2 {
		return
	}
	for i := 0; i < len(c.Windows); i++ {
		checkCol := -1
		for j := 0; j < len(c.Windows[i]); j++ {
			if c.Windows[i][j].ID == id {
				c.Windows[i] = removeSliceIndex(c.Windows[i], j)
				checkCol = i
				break
			}
		}

		if checkCol != -1 && len(c.Windows[checkCol]) == 0 {
			c.Windows = removeSliceIndex(c.Windows, checkCol)
		}

	}

	for _, col := range c.Windows {
		for _, win := range col {
			c.ActiveWindowIndex = win.ID
			break
		}
	}
}

func handlePanicAndWriteMessage(p *Context) {
	r := recover()
	if r != nil {
		msg := fmt.Sprintf("%v\n%s", r, string(debug.Stack()))
		fmt.Println(msg)
		p.WriteMessage(msg)
	}
}
