package main

import rl "github.com/gen2brain/raylib-go/raylib"

type UserInputBox struct {
	cfg          *Config
	parent       *Preditor
	keymaps      []Keymap
	maxHeight    int32
	maxWidth     int32
	UserInput    []byte
	ZeroLocation rl.Vector2
	Idx          int
	CursorShape  int
}
// move all logics of user input box here so we can re use this
