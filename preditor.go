package preditor

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/flopp/go-findfont"

	"github.com/davecgh/go-spew/spew"
	"golang.design/x/clipboard"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets/logo.png
var logoBytes []byte

type RGBA color.RGBA

func (r RGBA) String() string {
	colorAsHex := fmt.Sprintf("#%02x%02x%02x%02x", r.R, r.G, r.B, r.A)
	return colorAsHex
}

func (r RGBA) ToColorRGBA() color.RGBA {
	return color.RGBA(r)
}

type SyntaxColors map[string]RGBA

type Colors struct {
	Background                RGBA
	Foreground                RGBA
	SelectionBackground       RGBA
	SelectionForeground       RGBA
	Prompts                   RGBA
	StatusBarBackground       RGBA
	StatusBarForeground       RGBA
	ActiveStatusBarBackground RGBA
	ActiveStatusBarForeground RGBA
	LineNumbersForeground     RGBA
	ActiveWindowBorder        RGBA
	Cursor                    RGBA
	CursorLineBackground      RGBA
	HighlightMatching         RGBA
	SyntaxColors              SyntaxColors
}

type Drawable interface {
	GetID() int
	SetID(int)
	Render(zeroLocation rl.Vector2, maxHeight float64, maxWidth float64)
	//TODO: instead of returning a keymap slice we should have smth like
	// Binding(Key) Command
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
	DrawableID    int
	ZeroLocationX float64
	ZeroLocationY float64
	Width         float64
	Height        float64
}

func (w *Window) Render(c *Context, zeroLocation rl.Vector2, maxHeight float64, maxWidth float64) {
	if buf := c.GetDrawable(w.DrawableID); buf != nil {
		buf.Render(zeroLocation, maxHeight, maxWidth)
	}
}

type Prompt struct {
	IsActive   bool
	Text       string
	UserInput  string
	Keymap     Keymap
	DoneHook   func(userInput string, c *Context)
	ChangeHook func(userInput string, c *Context)
	NoRender   bool
}

const (
	BuildWindowState_Hide = iota
	BuildWindowState_Normal
	BuildWindowState_Maximized
	BuildWindowStateStateCount
)

type BuildWindow struct {
	Window
	State int
}

var PromptKeymap = Keymap{}

type Context struct {
	CWD               string
	Cfg               *Config
	ScratchBufferID   int
	MessageDrawableID int
	Buffers           map[string]*Buffer
	GlobalNoStatusbar bool
	Drawables         []Drawable
	GlobalKeymap      Keymap
	GlobalVariables   Variables
	Commands          Commands
	FontData          []byte
	Font              rl.Font
	FontSize          int32
	OSWindowHeight    float64
	OSWindowWidth     float64
	Windows           [][]*Window
	BuildWindow       BuildWindow
	Prompt            Prompt
	ActiveWindowIndex int
}

var GlobalKeymap = Keymap{}

func (c *Context) GetBufferByFilename(filename string) *Buffer {
	return c.Buffers[filename]
}

func (c *Context) OpenFileAsBuffer(filename string) *Buffer {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("ERROR: cannot read file", err.Error())
	}

	buf := Buffer{
		File:  filename,
		State: State_Clean,
	}

	//replace CRLF with LF
	if bytes.Index(content, []byte("\r\n")) != -1 {
		content = bytes.Replace(content, []byte("\r"), []byte(""), -1)
		buf.CRLF = true
	}

	fileType, exists := FileTypes[path.Ext(buf.File)]
	if exists {
		buf.fileType = fileType
		buf.needParsing = true
	}
	buf.Content = content
	c.Buffers[filename] = &buf

	return &buf
}

func (c *Context) BuildWindowIsVisible() bool {
	return c.BuildWindow.State != BuildWindowState_Hide && c.GetDrawable(c.BuildWindow.DrawableID) != nil
}

func (c *Context) BuildWindowMaximized() {
	c.BuildWindow.State = BuildWindowState_Maximized
}

func (c *Context) BuildWindowNormal() {
	c.BuildWindow.State = BuildWindowState_Normal
}
func (c *Context) BuildWindowHide() {
	c.BuildWindow.State = BuildWindowState_Hide
}
func (c *Context) BuildWindowToggleState() {
	c.BuildWindow.State++
	if c.BuildWindow.State >= BuildWindowStateStateCount {
		c.BuildWindow.State = 0
	}
}
func (c *Context) ResetPrompt() {
	c.Prompt.IsActive = false
	c.Prompt.UserInput = ""
	c.Prompt.DoneHook = nil
	c.Prompt.ChangeHook = nil

}
func (c *Context) SetPrompt(text string,
	changeHook func(userInput string, c *Context),
	doneHook func(userInput string, c *Context), keymap *Keymap, defaultValue string) {
	c.Prompt.IsActive = true
	c.Prompt.Text = text
	c.Prompt.DoneHook = doneHook
	c.Prompt.UserInput = defaultValue
	c.Prompt.ChangeHook = changeHook
	if keymap != nil {
		c.Prompt.Keymap = *keymap
	} else {
		c.Prompt.Keymap = PromptKeymap
	}
}

func (c *Context) ActiveWindow() *Window {
	w := c.GetWindow(c.ActiveWindowIndex)
	return w
}

func (c *Context) ActiveDrawable() Drawable {
	if win := c.GetWindow(c.ActiveWindowIndex); win != nil {
		return c.GetDrawable(win.DrawableID)
	}

	return nil
}
func (c *Context) ActiveDrawableID() int {
	if win := c.GetWindow(c.ActiveWindowIndex); win != nil {
		drawableID := win.DrawableID
		return drawableID
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
	c.GetDrawable(c.MessageDrawableID).(*BufferView).Buffer.Content = append(c.GetDrawable(c.MessageDrawableID).(*BufferView).Buffer.Content, []byte(fmt.Sprintln(msg))...)
}

func (c *Context) getCWD() string {
	if tb, isTextBuffer := c.ActiveDrawable().(*BufferView); isTextBuffer {
		if strings.Contains(tb.Buffer.File, "*Grep") || strings.Contains(tb.Buffer.File, "*Compilation") {
			segs := strings.Split(tb.Buffer.File, "@")
			if len(segs) > 1 {
				return segs[1]
			}
		} else {
			wd, _ := filepath.Abs(tb.Buffer.File)
			wd = filepath.Dir(wd)
			return wd
		}
	}
	return c.CWD
}
func (c *Context) AddDrawable(b Drawable) {
	id := rand.Intn(10000)
	b.SetID(id)
	c.Drawables = append(c.Drawables, b)
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

func (c *Context) MarkDrawableAsActive(id int) {
	c.ActiveWindow().DrawableID = id
}

func (c *Context) GetDrawable(id int) Drawable {
	for _, d := range c.Drawables {
		if d != nil && d.GetID() == id {
			return d
		}
	}

	return nil
}

func (c *Context) KillDrawable(id int) {
	for i, drawable := range c.Drawables {
		if drawable != nil && drawable.GetID() == id {
			c.Drawables[i] = nil
			break
		}
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

type Command func(*Context)
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

func (k Keymap) BindKey(key Key, command Command) {
	k[key] = command
}
func (k Keymap) SetKeys(k2 Keymap) {
	for b, f := range k2 {
		k[b] = f
	}
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

		keymaps := []Keymap{c.GlobalKeymap}
		if c.ActiveDrawable() != nil {
			keymaps = append(keymaps, c.ActiveDrawable().Keymaps()...)
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
	if id == -10 {
		return &c.BuildWindow.Window
	}
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
	var buildWindowHeightRatio float64
	if c.BuildWindow.State == BuildWindowState_Normal {
		buildWindowHeightRatio = c.Cfg.BuildWindowNormalHeight
	} else if c.BuildWindow.State == BuildWindowState_Maximized {
		buildWindowHeightRatio = c.Cfg.BuildWindowMaximizedHeight
	} else if c.BuildWindow.State == BuildWindowState_Hide {
		buildWindowHeightRatio = 0
	}
	charsize := measureTextSize(c.Font, ' ', c.FontSize, 0)
	if c.Prompt.IsActive && !c.Prompt.NoRender {
		height -= float64(charsize.Y)
	}
	if c.BuildWindowIsVisible() {
		height -= float64(buildWindowHeightRatio * c.OSWindowHeight)
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

	c.BuildWindow.ZeroLocationX = 0
	c.BuildWindow.Width = c.OSWindowWidth

	c.BuildWindow.ZeroLocationY = c.OSWindowHeight - (c.OSWindowHeight * buildWindowHeightRatio)
	c.BuildWindow.Height = c.OSWindowHeight * buildWindowHeightRatio

	if c.BuildWindowIsVisible() {
		buf := c.GetDrawable(c.BuildWindow.DrawableID)
		buf.Render(rl.Vector2{X: float32(c.BuildWindow.ZeroLocationX), Y: float32(c.BuildWindow.ZeroLocationY)}, c.BuildWindow.Height, c.BuildWindow.Width)
	}

	if c.Prompt.IsActive && !c.Prompt.NoRender {
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
		// handle build window
		if c.BuildWindowIsVisible() && float64(pos.Y) >= c.BuildWindow.ZeroLocationY {
			c.ActiveWindowIndex = c.BuildWindow.ID
		}

		keymaps := []Keymap{c.GlobalKeymap}
		if c.ActiveDrawable() != nil {
			keymaps = append(keymaps, c.ActiveDrawable().Keymaps()...)
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
	rl.InitWindow(800, 600, "Preditor")
	rl.SetTargetFPS(60)
	rl.SetTextLineSpacing(cfg.FontSize)
	img, _, err := image.Decode(bytes.NewReader(logoBytes))
	if err != nil {
		panic(err)
	}

	rlImage := rl.NewImageFromImage(img)
	rl.SetWindowIcon(*rlImage)
	rl.SetExitKey(0)

}

func New() (*Context, error) {
	var configPath string
	var startTheme string
	flag.StringVar(&configPath, "cfg", path.Join(os.Getenv("HOME"), ".preditor"), "path to config file, defaults to: ~/.preditor")
	flag.StringVar(&startTheme, "theme", "", "Start theme to use overrides the config and editor defaults.")
	flag.Parse()

	// read config file
	cfg, err := ReadConfig(configPath, startTheme)
	if err != nil {
		panic(err)
	}

	// create editor
	setupRaylib(cfg)

	if err := clipboard.Init(); err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	wd, err = filepath.Abs(wd)
	p := &Context{
		Cfg:            cfg,
		CWD:            wd,
		Drawables:      []Drawable{},
		OSWindowHeight: float64(rl.GetRenderHeight()),
		OSWindowWidth:  float64(rl.GetRenderWidth()),
		Windows:        [][]*Window{},
		Buffers:        map[string]*Buffer{},
	}

	setupDefaults()
	err = p.LoadFont(cfg.FontName, int32(cfg.FontSize))
	if err != nil {
		return nil, err
	}
	scratch := NewBufferViewFromFilename(p, p.Cfg, "*Scratch*")
	message := NewBufferViewFromFilename(p, p.Cfg, "*Messages*")
	message.Buffer.Readonly = true
	message.Buffer.Content = append(message.Buffer.Content, []byte(fmt.Sprintf("Loaded Configuration from '%s':\n%s\n", configPath, cfg))...)

	p.AddDrawable(scratch)
	p.AddDrawable(message)

	p.MessageDrawableID = message.ID
	p.ScratchBufferID = scratch.ID

	mainWindow := Window{}
	p.AddWindowInANewColumn(&mainWindow)

	p.MarkWindowAsActive(mainWindow.ID)

	p.MarkDrawableAsActive(scratch.ID)

	p.GlobalKeymap = GlobalKeymap

	p.BuildWindow = BuildWindow{
		Window: Window{ID: -10},
		State:  BuildWindowState_Normal,
	}
	// handle command line argument
	filename := ""
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
		if filename == "-" {
			//stdin
			tb := NewBufferViewFromFilename(p, cfg, "stdin")
			p.AddDrawable(tb)
			p.MarkDrawableAsActive(tb.ID)
			tb.Buffer.Readonly = true
			go func() {
				r := bufio.NewReader(os.Stdin)
				for {
					b, err := r.ReadByte()
					if err != nil {
						p.WriteMessage(err.Error())
						break
					}
					tb.Buffer.Content = append(tb.Buffer.Content, b)
				}
			}()
		} else {
			err = SwitchOrOpenFileInCurrentWindow(p, cfg, filename, nil)
			if err != nil {
				panic(err)
			}
		}
	}

	return p, nil
}

func Exit(c *Context) {
	rl.CloseWindow()
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

			fmt.Printf("%v\n%s\n", r, string(debug.Stack()))
		}
	}()
	//TODO: Check for drag and dropped files rl.IsFileDropped()
	for !rl.WindowShouldClose() {
		if rl.IsFileDropped() {
			files := rl.LoadDroppedFiles()
			if len(files) > 0 {
				for _, file := range files {
					SwitchOrOpenFileInCurrentWindow(c, c.Cfg, file, nil)
				}
			}
		}
		c.HandleWindowResize()
		c.HandleMouseEvents()
		c.HandleKeyEvents()
		c.Render()
	}
}

func MakeCommand[T Drawable](f func(t T)) Command {
	return func(c *Context) {
		f(c.ActiveDrawable().(T))

	}

}

func (c *Context) MaxHeightToMaxLine(maxH int32) int32 {
	return maxH / int32(measureTextSize(c.Font, ' ', c.FontSize, 0).Y)
}
func (c *Context) MaxWidthToMaxColumn(maxW int32) int32 {
	return maxW / int32(measureTextSize(c.Font, ' ', c.FontSize, 0).X)
}

func (c *Context) OpenFileList() {
	ofb := NewFileList(c, c.Cfg, c.getCWD())
	c.AddDrawable(ofb)
	c.MarkDrawableAsActive(ofb.ID)
}

func (c *Context) OpenFuzzyFileList() {
	ofb := NewFuzzyFileList(c, c.Cfg, c.getCWD())
	c.AddDrawable(ofb)
	c.MarkDrawableAsActive(ofb.ID)

}
func (c *Context) OpenBufferList() {
	ofb := NewBufferList(c, c.Cfg)
	c.AddDrawable(ofb)
	c.MarkDrawableAsActive(ofb.ID)
}

func (c *Context) OpenThemesList() {
	ofb := NewThemeList(c, c.Cfg)
	c.AddDrawable(ofb)
	c.MarkDrawableAsActive(ofb.ID)
}

func SwitchOrOpenFileInWindow(parent *Context, cfg *Config, filename string, startingPos *Position, window *Window) error {
	bufferView := NewBufferViewFromFilename(parent, cfg, filename)
	parent.AddDrawable(bufferView)
	window.DrawableID = bufferView.ID
	bufferView.MoveToPositionInNextRender = startingPos
	return nil
}

func SwitchOrOpenFileInCurrentWindow(parent *Context, cfg *Config, filename string, startingPos *Position) error {
	return SwitchOrOpenFileInWindow(parent, cfg, filename, startingPos, parent.ActiveWindow())
}

func (c *Context) openCompilationBuffer(command string) error {
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddDrawable(cb)
	c.MarkDrawableAsActive(cb.ID)

	return nil
}

func (c *Context) OpenCompilationBufferInAVSplit(command string) error {
	win := VSplit(c)
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddDrawable(cb)
	win.DrawableID = cb.ID
	return nil
}

func (c *Context) OpenCompilationBufferInAHSplit(command string) error {
	win := HSplit(c)
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddDrawable(cb)
	win.DrawableID = cb.ID
	return nil
}

func (c *Context) OpenCompilationBufferInSensibleSplit(command string) error {
	var window *Window
	for _, col := range c.Windows {
		for _, win := range col {
			if buf := c.GetDrawable(win.DrawableID); buf != nil {
				if b, is := buf.(*BufferView); is {
					if b.Buffer.File == "*Compilation*" {
						window = win
					}
				}
			}
		}
	}

	if window == nil {
		window = HSplit(c)
	}
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddDrawable(cb)
	window.DrawableID = cb.ID

	return nil
}

func Compile(c *Context) {
	c.SetPrompt("Compile", nil, func(userInput string, c *Context) {
		if err := c.openCompilationBuffer(userInput); err != nil {
			return
		}

		return
	}, nil, "")

}

func (c *Context) OpenGrepBufferInSensibleSplit(command string) error {
	var window *Window
	for _, col := range c.Windows {
		for _, win := range col {
			if buf := c.GetDrawable(win.DrawableID); buf != nil {
				if b, is := buf.(*BufferView); is {
					if b.Buffer.File == "*Grep*" {
						window = win
					}
				}
			}
		}
	}

	if window == nil {
		window = VSplit(c)
	}
	cb, err := NewGrepBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}
	c.AddDrawable(cb)
	window.DrawableID = cb.ID

	return nil
}

func (c *Context) OpenCompilationBufferInBuildWindow(command string) error {
	cb, err := NewCompilationBuffer(c, c.Cfg, command)
	if err != nil {
		return err
	}

	c.AddDrawable(cb)

	c.BuildWindow.DrawableID = cb.ID

	return nil
}

func VSplit(c *Context) *Window {
	win := &Window{}
	c.AddWindowInANewColumn(win)
	return win
}

func HSplit(c *Context) *Window {
	win := &Window{}
	c.AddWindowInCurrentColumn(win)
	return win
}

func (c *Context) OtherWindow() {
	for i, col := range c.Windows {
		for j, win := range col {
			if c.ActiveWindowIndex == -10 {
				c.ActiveWindowIndex = win.ID
				break
			}
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

func ToggleGlobalNoStatusbar(c *Context) {
	c.GlobalNoStatusbar = !c.GlobalNoStatusbar
}
