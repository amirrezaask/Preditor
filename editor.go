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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Public API
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func InsertCharAtCursor(e *Editor, char byte) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).InsertCharAtCursor(char)
	default:
		return fmt.Errorf("InsertChartAtCursor is not implemented for %s", t)
	}
}

func ScrollUp(e *Editor) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).ScrollUp(1)
	default:
		return fmt.Errorf("ScrollUp is not implemented for %s", t)
	}
	
}

func ScrollDown(e *Editor) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).ScrollDown(1)
	default:
		return fmt.Errorf("ScrollDown is not implemented for %s", t)
	}
	
}

func CursorLeft(e *Editor) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).CursorLeft()
	default:
		return fmt.Errorf("CursorLeft is not implemented for %s", t)
	}
	
}
func CursorRight(e *Editor) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).CursorRight()
	default:
		return fmt.Errorf("CursorRight is not implemented for %s", t)
	}
	
}

func CursorUp(e *Editor) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).CursorUp()
	default:
		return fmt.Errorf("CursorUp is not implemented for %s", t)
	}

}
func CursorDown(e *Editor) error {
	switch t := e.ActiveBuffer().Type(); t {
	case "text_editor_buffer":
		return e.ActiveBuffer().(*TextEditorBuffer).CursorDown()
	default:
		return fmt.Errorf("CursorDown is not implemented for %s", t)
	}

	
}
