package components

import (
	"bytes"
	"github.com/amirrezaask/preditor/byteutils"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.design/x/clipboard"
)

type UserInputComponent struct {
	maxHeight    int32
	maxWidth     int32
	UserInput    []byte
	ZeroLocation rl.Vector2
	Idx          int
	LastInput    string
}

func NewUserInputComponent() *UserInputComponent {
	uib := UserInputComponent{}

	return &uib
}

func (f *UserInputComponent) SetNewUserInput(bs []byte) {
	f.LastInput = string(f.UserInput)
	f.UserInput = bs
	f.Idx += len(f.UserInput)

	if f.Idx >= len(f.UserInput) {
		f.Idx = len(f.UserInput)
	} else if f.Idx < 0 {
		f.Idx = 0
	}

}
func (f *UserInputComponent) InsertCharAtBuffer(char byte) error {
	f.SetNewUserInput(append(f.UserInput, char))
	return nil
}

func (f *UserInputComponent) CursorRight(n int) error {
	if f.Idx >= len(f.UserInput) {
		return nil
	}

	f.Idx += n

	return nil
}

func (f *UserInputComponent) Paste() error {
	content := getClipboardContent()
	f.UserInput = append(f.UserInput[:f.Idx], append(content, f.UserInput[f.Idx+1:]...)...)

	return nil
}

func (f *UserInputComponent) KillLine() error {
	f.SetNewUserInput(f.UserInput[:f.Idx])
	return nil
}

func (f *UserInputComponent) Copy() error {
	writeToClipboard(f.UserInput)

	return nil
}

func (f *UserInputComponent) BeginningOfTheLine() error {
	f.Idx = 0
	return nil
}

func (f *UserInputComponent) EndOfTheLine() error {
	f.Idx = len(f.UserInput)
	return nil
}

func (f *UserInputComponent) NextWordStart() error {
	if idx := byteutils.NextWordInBuffer(f.UserInput, f.Idx); idx != -1 {
		f.Idx = idx
	}

	return nil
}

func (f *UserInputComponent) CursorLeft(n int) error {

	if f.Idx <= 0 {
		return nil
	}

	f.Idx -= n

	return nil
}

func (f *UserInputComponent) PreviousWord() error {
	if idx := byteutils.PreviousWordInBuffer(f.UserInput, f.Idx); idx != -1 {
		f.Idx = idx
	}

	return nil
}

func (f *UserInputComponent) DeleteCharBackward() error {
	if f.Idx <= 0 {
		return nil
	}
	if len(f.UserInput) <= f.Idx {
		f.SetNewUserInput(f.UserInput[:f.Idx-1])
	} else {
		f.SetNewUserInput(append(f.UserInput[:f.Idx-1], f.UserInput[f.Idx:]...))
	}
	return nil
}

func (f *UserInputComponent) DeleteWordBackward() error {
	previousWordEndIdx := byteutils.PreviousWordInBuffer(f.UserInput, f.Idx)
	if len(f.UserInput) > f.Idx+1 {
		f.SetNewUserInput(append(f.UserInput[:previousWordEndIdx+1], f.UserInput[f.Idx+1:]...))
	} else {
		f.SetNewUserInput(f.UserInput[:previousWordEndIdx+1])
	}
	return nil
}
func (f *UserInputComponent) DeleteWordForward() error {
	nextWordStartIdx := byteutils.NextWordInBuffer(f.UserInput, f.Idx)
	if len(f.UserInput) > nextWordStartIdx+1 {
		f.SetNewUserInput(append(f.UserInput[:f.Idx+1], f.UserInput[nextWordStartIdx+1:]...))
	} else {
		f.SetNewUserInput(f.UserInput[:f.Idx])
	}

	return nil
}
func (f *UserInputComponent) DeleteCharForward() error {
	if f.Idx < 0 {
		return nil
	}
	f.SetNewUserInput(append(f.UserInput[:f.Idx], f.UserInput[f.Idx+1:]...))
	return nil
}

func getClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func writeToClipboard(bs []byte) {
	clipboard.Write(clipboard.FmtText, bytes.Clone(bs))
}
