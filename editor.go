package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type BufferOptions struct{
	MaxHeight int32
	MaxWidth int32
	ZeroPosition rl.Vector2
}

type Buffer interface{
	Type() string
	Initialize(BufferOptions) error
	Render()
	Destroy() error
}

type Editor struct {
	Buffers           []Buffer
	ActiveBufferIndex int
	GlobalKeymaps     []Keymap
	GlobalVariables   Variables
	Commands          Commands
	LineWrapping      bool
}

func (e *Editor) ActiveBuffer() Buffer {
	return e.Buffers[e.ActiveBufferIndex]
}

type Command func(*Editor) error
type Variables map[string]any
type Key struct {
	Control  bool
	Alt   bool
	Shift bool
	Super bool
	K     string
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


