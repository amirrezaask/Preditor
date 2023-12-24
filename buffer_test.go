package preditor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearch(t *testing.T) {
	matched := matchPatternCaseInsensitive([]byte("Hello World"), []byte("Hell"))
	assert.Equal(t, [][]int{{0, 3}}, matched)
}

// func Test_Gotoline(t *testing.T) {}

// ///////////// These are the buggy ones
func Test_BufferInsertChar(t *testing.T) {
	bufferView := BufferView{
		Buffer: &Buffer{
			File:    "",
			Content: []byte("12"),
			CRLF:    false,
		},
		Cursors: []Cursor{
			{
				Point: 0,
				Mark:  0,
			},
		},
		ActionStack: NewStack[BufferAction](10),
	}

	BufferInsertChar(&bufferView, '0')
	assert.Equal(t, []byte("012"), bufferView.Buffer.Content)
	bufferView.Cursors[0].SetBoth(3)
	BufferInsertChar(&bufferView, '3')
	assert.Equal(t, []byte("0123"), bufferView.Buffer.Content)

	RevertLastBufferAction(&bufferView)
	assert.Equal(t, []byte("012"), bufferView.Buffer.Content)
	RevertLastBufferAction(&bufferView)
	assert.Equal(t, []byte("12"), bufferView.Buffer.Content)
}
func Test_RemoveRange(t *testing.T) {
	bufferView := BufferView{
		Buffer: &Buffer{
			File:    "",
			Content: []byte("012345678\n012345678"),
			CRLF:    false,
		},
		Cursors: []Cursor{
			{
				Point: 2,
				Mark:  2,
			},
		},
		ActionStack: NewStack[BufferAction](10),
	}

	bufferView.RemoveRange(3, 8, true)
	assert.Equal(t, "0128\n012345678", string(bufferView.Buffer.Content))
	RevertLastBufferAction(&bufferView)
	assert.Equal(t, "012345678\n012345678", string(bufferView.Buffer.Content))
}

func Test_KillLine(t *testing.T) {
	bufferView := BufferView{
		Buffer: &Buffer{
			File:    "",
			Content: []byte("012345678\n012345678"),
			CRLF:    false,
		},
		Cursors: []Cursor{
			{
				Point: 2,
				Mark:  2,
			},
		},
		ActionStack: NewStack[BufferAction](10),
	}
	bufferView.calcRenderState()
	KillLine(&bufferView)
	assert.Equal(t, "01\n012345678", string(bufferView.Buffer.Content))
	RevertLastBufferAction(&bufferView)
	assert.Equal(t, "012345678\n012345678", string(bufferView.Buffer.Content))

}

//func Test_Cut(t *testing.T)                {}
//func Test_Paste(t *testing.T)              {}
//func Test_DeleteCharBackword(t *testing.T) {}
//func Test_DeleteCharForeward(t *testing.T) {}
//
//////////////////////////////////////////////////
//func Test_WordAtPoint(t *testing.T)             {}
//func Test_LeftWord(t *testing.T)                {}
//func Test_RightWord(t *testing.T)               {}
//func Test_DeleteWordBackward(t *testing.T)      {}
//func Test_Indent(t *testing.T)                  {}
//func Test_ScrollUp(t *testing.T)                {}
//func Test_ScrollToTop(t *testing.T)             {}
//func Test_ScrollToBottom(t *testing.T)          {}
//func Test_ScrollDown(t *testing.T)              {}
//func Test_PointLeft(t *testing.T)               {}
//func Test_PointRight(t *testing.T)              {}
//func Test_PointUp(t *testing.T)                 {}
//func Test_PointDown(t *testing.T)               {}
//func Test_CentralizePoint(t *testing.T)         {}
//func Test_PointToBeginningOfLine(t *testing.T)  {}
//func Test_PointToEndOfLine(t *testing.T)        {}
//func Test_PointToMatchingChar(t *testing.T)     {}
//func Test_MarkRight(t *testing.T)               {}
//func Test_MarkLeft(t *testing.T)                {}
//func Test_MarkUp(t *testing.T)                  {}
//func Test_MarkDown(t *testing.T)                {}
//func Test_MarkPreviousWord(t *testing.T)        {}
//func Test_MarkNextWord(t *testing.T)            {}
//func Test_MarkToEndOfLine(t *testing.T)         {}
//func Test_MarkToBeginningOfLine(t *testing.T)   {}
//func Test_MarkToMatchingChar(t *testing.T)      {}
//func Test_RemoveAllCursorsButOne(t *testing.T)  {}
//func Test_AddCursorNextLine(t *testing.T)       {}
//func Test_AddCursorPreviousLine(t *testing.T)   {}
//func Test_PointForwardWord(t *testing.T)        {}
//func Test_PointBackwardWord(t *testing.T)       {}
//func Test_Copy(t *testing.T)                    {}
//func Test_Search(t *testing.T)                  {}
//func Test_QueryReplace(t *testing.T)            {}
//func Test_IsvalidCursorPosition(t *testing.T)   {}
//func Test_AnotherSelectionOnMatch(t *testing.T) {}
