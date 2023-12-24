package main

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type BufferOptions struct {
	MaxHeight    int32
	MaxWidth     int32
	ZeroPosition rl.Vector2
	Colors       Colors
}

type Buffer interface {
	Type() string
	Initialize(BufferOptions) error
	SetMaxWidth(w int32)
	SetMaxHeight(h int32)
	Render()
	Destroy() error
	MoveCursorTo(rl.Vector2) error
}

type Colors struct {
	Background          color.RGBA
	Foreground          color.RGBA
	SelectionBackground color.RGBA
	SelectionForeground color.RGBA
}

type Editor struct {
	Buffers           []Buffer
	ActiveBufferIndex int
	GlobalKeymaps     []Keymap
	GlobalVariables   Variables
	Commands          Commands
	LineWrapping      bool
	Colors            Colors
}

func (e *Editor) ActiveBuffer() Buffer {
	return e.Buffers[e.ActiveBufferIndex]
}

type Command func(*Editor) error
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


func (e *Editor) HandleKeyEvents() {
	key := getKey()
	if !key.IsEmpty() {
		cmd := defaultKeymap[key]
		if cmd != nil {
			cmd(e)
		}
	}

}

func (e *Editor) Render() {
	rl.BeginDrawing()
	rl.ClearBackground(e.Colors.Background)

	e.ActiveBuffer().Render()

	rl.EndDrawing()

}


func (e *Editor) HandleWindowResize() {
	if !(rl.IsWindowResized() || rl.IsWindowMaximized()) {
		return
	}
	height := rl.GetRenderHeight()
	width := rl.GetRenderWidth()

	// window is resized
	for _, buffer := range e.Buffers {
		buffer.SetMaxWidth(int32(width))
		buffer.SetMaxHeight(int32(height))
	}
}

func (e *Editor) HandleMouseEvents() {
	key := getMouseKey()
	if !key.IsEmpty() {
		cmd := defaultKeymap[key]
		if cmd != nil {
			cmd(e)
		}
	}
}
