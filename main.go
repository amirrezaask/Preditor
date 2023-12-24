/*
   move all functionalities to editor struct
   editor.GetCursorBufferIndex()
*/

package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	font     rl.Font
	fontSize float32
)

func main() {
	// basic setup
	rl.InitWindow(1920, 1080, "core editor")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// create editor
	editor := Editor{
		LineWrapping: true,
	}

	fontSize = 20
	rl.SetTextLineSpacing(int(fontSize))
	rl.SetMouseCursor(rl.MouseCursorIBeam)
	editor.Buffers = append(editor.Buffers, Buffer{
		Content: []byte(`lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .
b
lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .

lorem ipsum dolor sit amet . The graphic and typographic operators know this well, in reality all the professions dealing with the universe of communication have a stable relationship with these words, but what is it? Lorem ipsum is a dummy text without any sense.

It is a sequence of Latin words that, as they are positioned, do not form sentences with a complete sense, but give life to a test text useful to fill spaces that will subsequently be occupied from ad hoc texts composed by communication professionals.

It is certainly the most famous placeholder text even if there are different versions distinguishable from the order in which the Latin words are repeated.

Lorem ipsum contains the typefaces more in use, an aspect that allows you to have an overview of the rendering of the text in terms of font choice and font size .
`),
		FilePath: "test.txt",
	})
	editor.Windows = append(editor.Windows, Window{
		BufferIndex: 0,
		zeroLocation: rl.Vector2{
			X: 0, Y: 0,
		},
		Height: rl.GetRenderHeight(),
		Width:  rl.GetRenderWidth(),
		Cursor: Position{},
	})

	font = rl.LoadFontEx("FiraCode.ttf", int32(fontSize), nil)
	for !rl.WindowShouldClose() {
		buffer := &editor.Buffers[editor.Windows[editor.ActiveWindowIndex].BufferIndex]

		// execute any key command that should be executed
		cmd := defaultKeymap[MakeKey(buffer)]
		if cmd != nil {
			if err := cmd(&editor); err != nil {
				panic(err)
			}
		}

		// cmd = defaultKeymap[MakeMouseKey(buffer)]
		// if cmd != nil {
		// 	if err := cmd(editor); cmd != nil {
		// 		panic(err)
		// 	}
		// }

		// Render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		RenderBufferInWindow(&editor, &editor.Buffers[0], &editor.Windows[0])

		rl.EndDrawing()
	}

}

var charSizeCache = map[byte]rl.Vector2{} //TODO: if font size or font changes this is fucked
func measureTextSize(font rl.Font, s byte, size float32, spacing float32) rl.Vector2 {
	if charSize, exists := charSizeCache[s]; exists {
		return charSize
	}
	charSize := rl.MeasureTextEx(font, string(s), size, spacing)
	charSizeCache[s] = charSize
	return charSize
}
