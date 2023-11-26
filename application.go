package main

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Colors struct {
	Background            color.RGBA
	Foreground            color.RGBA
	SelectionBackground   color.RGBA
	SelectionForeground   color.RGBA
	StatusBarBackground   color.RGBA
	StatusBarForeground   color.RGBA
	LineNumbersForeground color.RGBA
}

type Application struct {
	Editors           []*TextEditor
	ActiveEditorIndex int
	GlobalKeymaps     []Keymap
	GlobalVariables   Variables
	Commands          Commands
	LineWrapping      bool
	LineNumbers       bool
	Colors            Colors
}

func (e *Application) ActiveEditor() *TextEditor {
	return e.Editors[e.ActiveEditorIndex]
}

type Command func(*Application) error
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

func (e *Application) HandleKeyEvents() {
	key := getKey()
	if !key.IsEmpty() {
		cmd := defaultKeymap[key]
		if cmd != nil {
			cmd(e)
		}
	}

}

func (e *Application) renderStatusBar() {
	t := e.ActiveEditor()
	charSize := measureTextSize(font, ' ', fontSize, 0)

	//render status bar
	rl.DrawRectangle(
		int32(t.ZeroPosition.X),
		t.maxLine*int32(charSize.Y),
		t.MaxWidth,
		int32(charSize.Y),
		t.Colors.StatusBarBackground,
	)
	file := t.File
	if file == "" {
		file = "[scratch]"
	}
	var state string
	if t.State == State_Dirty {
		state = "**"
	} else {
		state = "--"
	}

	rl.DrawTextEx(font,
		fmt.Sprintf("%s %s %d:%d", state, file, t.Cursor.Line, t.Cursor.Column),
		rl.Vector2{X: t.ZeroPosition.X, Y: float32(t.maxLine) * charSize.Y},
		fontSize,
		0,
		t.Colors.StatusBarForeground)
}

func (e *Application) Render() {
	rl.BeginDrawing()
	rl.ClearBackground(e.Colors.Background)

	e.ActiveEditor().Render()
	e.renderStatusBar()

	rl.EndDrawing()

}

func (e *Application) HandleWindowResize() {
	height := rl.GetRenderHeight()
	width := rl.GetRenderWidth()

	// window is resized
	for _, buffer := range e.Editors {
		if buffer.GetMaxWidth() != int32(width) {
			buffer.SetMaxWidth(int32(width))
		}

		if buffer.GetMaxHeight() != int32(height) {
			buffer.SetMaxHeight(int32(height))

		}
	}
}

func (e *Application) HandleMouseEvents() {
	key := getMouseKey()
	if !key.IsEmpty() {
		cmd := defaultKeymap[key]
		if cmd != nil {
			cmd(e)
		}
	}
}
