package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)


type modifierKeyState struct {
	control bool
	alt bool
	shift bool
	super bool

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
	modifierState:= getModifierKeyState()
	key := getKeyPressedString()

	k := Key{
		Control: modifierState.control,
		Alt: modifierState.alt,
		Super: modifierState.super,
		Shift: modifierState.shift,
		K: key,
	}
	if !k.IsEmpty() {
		fmt.Println("=================================")
		fmt.Printf("key: %+v\n", k)
		fmt.Println("=================================")
		
	}

	return k
}



func getKeyPressedString() string {
	switch {
	case rl.IsKeyPressed(rl.KeySpace):
		return "<space>"
	case rl.IsKeyPressed(rl.KeyEscape):
		return "<esc>"
	case rl.IsKeyPressed(rl.KeyEnter):
		return "<enter>"
	case rl.IsKeyPressed(rl.KeyTab):
		return "<tab>"
	case rl.IsKeyPressed(rl.KeyBackspace):
		return "<backspace>"
	case rl.IsKeyPressed(rl.KeyInsert):
		return "<insert>"
	case rl.IsKeyPressed(rl.KeyDelete):
		return "<delete>"
	case rl.IsKeyPressed(rl.KeyRight):
		return "<right>"
	case rl.IsKeyPressed(rl.KeyLeft):
		return "<left>"
	case rl.IsKeyPressed(rl.KeyDown):
		return "<down>"
	case rl.IsKeyPressed(rl.KeyUp):
		return "<up>"
	case rl.IsKeyPressed(rl.KeyPageUp):
		return "<pageup>"
	case rl.IsKeyPressed(rl.KeyPageDown):
		return "<pagedown>"
	case rl.IsKeyPressed(rl.KeyHome):
		return "<home>"
	case rl.IsKeyPressed(rl.KeyEnd):
		return "<end>"
	case rl.IsKeyPressed(rl.KeyCapsLock):
		return "<capslock>"
	case rl.IsKeyPressed(rl.KeyScrollLock):
		return "<scrolllock>"
	case rl.IsKeyPressed(rl.KeyNumLock):
		return "<numlock>"
	case rl.IsKeyPressed(rl.KeyPrintScreen):
		return "<printscreen>"
	case rl.IsKeyPressed(rl.KeyPause):
		return "<pause>"
	case rl.IsKeyPressed(rl.KeyF1):
		return "<f1>"
	case rl.IsKeyPressed(rl.KeyF2):
		return "<f2>"
	case rl.IsKeyPressed(rl.KeyF3):
		return "<f3>"
	case rl.IsKeyPressed(rl.KeyF4):
		return "<f4>"
	case rl.IsKeyPressed(rl.KeyF5):
		return "<f5>"
	case rl.IsKeyPressed(rl.KeyF6):
		return "<f6>"
	case rl.IsKeyPressed(rl.KeyF7):
		return "<f7>"
	case rl.IsKeyPressed(rl.KeyF8):
		return "<f8>"
	case rl.IsKeyPressed(rl.KeyF9):
		return "<f9>"
	case rl.IsKeyPressed(rl.KeyF10):
		return "<f10>"
	case rl.IsKeyPressed(rl.KeyF11):
		return "<f11>"
	case rl.IsKeyPressed(rl.KeyF12):
		return "<f12>"
	case rl.IsKeyPressed(rl.KeyLeftBracket):
		return "["
	case rl.IsKeyPressed(rl.KeyBackSlash):
		return "\\"
	case rl.IsKeyPressed(rl.KeyRightBracket):
		return "]"
	case rl.IsKeyPressed(rl.KeyKp0):
		return "0"
	case rl.IsKeyPressed(rl.KeyKp1):
		return "1"
	case rl.IsKeyPressed(rl.KeyKp2):
		return "2"
	case rl.IsKeyPressed(rl.KeyKp3):
		return "3"
	case rl.IsKeyPressed(rl.KeyKp4):
		return "4"
	case rl.IsKeyPressed(rl.KeyKp5):
		return "5"
	case rl.IsKeyPressed(rl.KeyKp6):
		return "6"
	case rl.IsKeyPressed(rl.KeyKp7):
		return "7"
	case rl.IsKeyPressed(rl.KeyKp8):
		return "8"
	case rl.IsKeyPressed(rl.KeyKp9):
		return "9"
	case rl.IsKeyPressed(rl.KeyKpDecimal):
		return "."
	case rl.IsKeyPressed(rl.KeyKpDivide):
		return "/"
	case rl.IsKeyPressed(rl.KeyKpMultiply):
		return "*"
	case rl.IsKeyPressed(rl.KeyKpSubtract):
		return "-"
	case rl.IsKeyPressed(rl.KeyKpAdd):
		return "+"
	case rl.IsKeyPressed(rl.KeyKpEnter):
		return "<enter>"
	case rl.IsKeyPressed(rl.KeyKpEqual):
		return "="
	case rl.IsKeyPressed(rl.KeyApostrophe):
		return "'"
	case rl.IsKeyPressed(rl.KeyComma):
		return ","
	case rl.IsKeyPressed(rl.KeyMinus):
		return "-"
	case rl.IsKeyPressed(rl.KeyPeriod):
		return "."
	case rl.IsKeyPressed(rl.KeySlash):
		return "/"
	case rl.IsKeyPressed(rl.KeyZero):
		return "0"
	case rl.IsKeyPressed(rl.KeyOne):
		return "1"
	case rl.IsKeyPressed(rl.KeyTwo):
		return "2"
	case rl.IsKeyPressed(rl.KeyThree):
		return "3"
	case rl.IsKeyPressed(rl.KeyFour):
		return "4"
	case rl.IsKeyPressed(rl.KeyFive):
		return "5"
	case rl.IsKeyPressed(rl.KeySix):
		return "6"
	case rl.IsKeyPressed(rl.KeySeven):
		return "7"
	case rl.IsKeyPressed(rl.KeyEight):
		return "8"
	case rl.IsKeyPressed(rl.KeyNine):
		return "9"
	case rl.IsKeyPressed(rl.KeySemicolon):
		return ";"
	case rl.IsKeyPressed(rl.KeyEqual):
		return "="
	case rl.IsKeyPressed(rl.KeyA):
		return "a"
	case rl.IsKeyPressed(rl.KeyB):
		return "b"
	case rl.IsKeyPressed(rl.KeyC):
		return "c"
	case rl.IsKeyPressed(rl.KeyD):
		return "d"
	case rl.IsKeyPressed(rl.KeyE):
		return "e"
	case rl.IsKeyPressed(rl.KeyF):
		return "f"
	case rl.IsKeyPressed(rl.KeyG):
		return "g"
	case rl.IsKeyPressed(rl.KeyH):
		return "h"
	case rl.IsKeyPressed(rl.KeyI):
		return "i"
	case rl.IsKeyPressed(rl.KeyJ):
		return "j"
	case rl.IsKeyPressed(rl.KeyK):
		return "k"
	case rl.IsKeyPressed(rl.KeyL):
		return "l"
	case rl.IsKeyPressed(rl.KeyM):
		return "m"
	case rl.IsKeyPressed(rl.KeyN):
		return "n"
	case rl.IsKeyPressed(rl.KeyO):
		return "o"
	case rl.IsKeyPressed(rl.KeyP):
		return "p"
	case rl.IsKeyPressed(rl.KeyQ):
		return "q"
	case rl.IsKeyPressed(rl.KeyR):
		return "r"
	case rl.IsKeyPressed(rl.KeyS):
		return "s"
	case rl.IsKeyPressed(rl.KeyT):
		return "t"
	case rl.IsKeyPressed(rl.KeyU):
		return "u"
	case rl.IsKeyPressed(rl.KeyV):
		return "v"
	case rl.IsKeyPressed(rl.KeyW):
		return "w"
	case rl.IsKeyPressed(rl.KeyX):
		return "x"
	case rl.IsKeyPressed(rl.KeyY):
		return "y"
	case rl.IsKeyPressed(rl.KeyZ):
		return "z"
	default:
		return ""
	}


	return ""
	
}

