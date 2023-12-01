package main

import rl "github.com/gen2brain/raylib-go/raylib"

type UserInputBox struct {
	cfg          *Config
	parent       *Preditor
	maxHeight    int32
	maxWidth     int32
	UserInput    []byte
	ZeroLocation rl.Vector2
	Idx          int
}

func NewuserInputBox(parent *Preditor, cfg *Config, zeroLocation rl.Vector2, maxH int32, maxW int32) *UserInputBox {
	uib := UserInputBox{
		cfg:          cfg,
		parent:       parent,
		maxHeight:    maxH,
		maxWidth:     maxW,
		ZeroLocation: zeroLocation,
	}

	return &uib
}

func (f *UserInputBox) setNewUserInput(bs []byte) {
	f.UserInput = bs
	f.Idx += len(f.UserInput)

	if f.Idx >= len(f.UserInput) {
		f.Idx = len(f.UserInput)
	} else if f.Idx < 0 {
		f.Idx = 0
	}

}
func (f *UserInputBox) insertCharAtBuffer(char byte) error {
	f.setNewUserInput(append(f.UserInput, char))
	return nil
}

func (f *UserInputBox) CursorRight(n int) error {
	if f.Idx >= len(f.UserInput) {
		return nil
	}

	f.Idx += n

	return nil
}

func (f *UserInputBox) paste() error {
	content := getClipboardContent()
	f.UserInput = append(f.UserInput[:f.Idx], append(content, f.UserInput[f.Idx+1:]...)...)

	return nil
}

func (f *UserInputBox) killLine() error {
	f.setNewUserInput(f.UserInput[:f.Idx])
	return nil
}

func (f *UserInputBox) copy() error {
	writeToClipboard(f.UserInput)

	return nil
}

func (f *UserInputBox) BeginingOfTheLine() error {
	f.Idx = 0
	return nil
}

func (f *UserInputBox) EndOfTheLine() error {
	f.Idx = len(f.UserInput)
	return nil
}

func (f *UserInputBox) NextWordStart() error {
	if idx := nextWordInBuffer(f.UserInput, f.Idx); idx != -1 {
		f.Idx = idx
	}

	return nil
}

func (f *UserInputBox) CursorLeft(n int) error {

	if f.Idx <= 0 {
		return nil
	}

	f.Idx -= n

	return nil
}

func (f *UserInputBox) PreviousWord() error {
	if idx := previousWordInBuffer(f.UserInput, f.Idx); idx != -1 {
		f.Idx = idx
	}

	return nil
}

func (f *UserInputBox) DeleteCharBackward() error {
	if f.Idx <= 0 {
		return nil
	}
	if len(f.UserInput) <= f.Idx {
		f.setNewUserInput(f.UserInput[:f.Idx-1])
	} else {
		f.setNewUserInput(append(f.UserInput[:f.Idx-1], f.UserInput[f.Idx:]...))
	}
	return nil
}

func (f *UserInputBox) DeleteWordBackward() error {
	previousWordEndIdx := previousWordInBuffer(f.UserInput, f.Idx)
	if len(f.UserInput) > f.Idx+1 {
		f.setNewUserInput(append(f.UserInput[:previousWordEndIdx+1], f.UserInput[f.Idx+1:]...))
	} else {
		f.setNewUserInput(f.UserInput[:previousWordEndIdx+1])
	}
	return nil
}
func (f *UserInputBox) DeleteWordForward() error {
	nextWordStartIdx := nextWordInBuffer(f.UserInput, f.Idx)
	if len(f.UserInput) > nextWordStartIdx+1 {
		f.setNewUserInput(append(f.UserInput[:f.Idx+1], f.UserInput[nextWordStartIdx+1:]...))
	} else {
		f.setNewUserInput(f.UserInput[:f.Idx])
	}

	return nil
}
func (f *UserInputBox) DeleteCharForward() error {
	if f.Idx < 0 {
		return nil
	}
	f.setNewUserInput(append(f.UserInput[:f.Idx], f.UserInput[f.Idx+1:]...))
	return nil
}
