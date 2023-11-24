package main

import (
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
	font     rl.Font
	fontSize float32
)
