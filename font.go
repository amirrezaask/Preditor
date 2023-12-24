package main

import (
	"github.com/flopp/go-findfont"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var charSizeCache = map[byte]rl.Vector2{} //TODO: if font size or font changes this is fucked
func measureTextSize(font rl.Font, s byte, size float32, spacing float32) rl.Vector2 {
	if charSize, exists := charSizeCache[s]; exists {
		return charSize
	}
	charSize := rl.MeasureTextEx(font, string(s), size, spacing)
	charSizeCache[s] = charSize
	return charSize
}


var (
	fontPath string
	font     rl.Font
	fontSize float32
)

func loadFont(name string, size float32) error {
	var err error
	fontPath, err = findfont.Find(name + ".ttf")
	if err != nil {
		return err
	}

	fontSize = size
	font = rl.LoadFontEx(fontPath, int32(fontSize), nil)

	return nil
}

func increaseFontSize(n int) {
	fontSize += float32(n)
	font = rl.LoadFontEx(fontPath, int32(fontSize), nil)
	charSizeCache = map[byte]rl.Vector2{}
}

func decreaseFontSize(n int) {
	fontSize -= float32(n)
	font = rl.LoadFontEx(fontPath, int32(fontSize), nil)
	charSizeCache = map[byte]rl.Vector2{}

}
