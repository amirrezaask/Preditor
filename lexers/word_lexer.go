package lexers

const (
	WORD_LEXER_TOKEN_TYPE_WORD     = 1
	WORD_LEXER_TOKEN_TYPE_NON_WORD = 2
	WORD_LEXER_TOKEN_TYPE_EOF      = 3
)

const (
	wordLexerState_InsideWord = 1
)

type WordLexer struct {
	data  []byte
	point int
	state int
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isWhitespace(b byte) bool {
	return b == '\n' || b == '\r' || b == ' '
}

func isLetterOrDigit(b byte) bool {
	return isLetter(b) || isDigit(b)
}

func (w *WordLexer) Next() Token {
	if w.point == len(w.data)-1 {
		return Token{Type: 3}
	}
	for i := w.point; i < len(w.data); i++ {
		switch {
		case w.state == wordLexerState_InsideWord && isLetterOrDigit(w.data[i]):
			continue
		case w.state == 0 && isLetterOrDigit(w.data[i]):
			w.state = wordLexerState_InsideWord
			continue
		case w.state == wordLexerState_InsideWord && !isLetterOrDigit(w.data[i]):
			w.state = 0
			token := Token{
				Start: w.point,
				End:   i,
				Type:  1,
			}

			w.point = i + 1
			if w.point >= len(w.data) {
				w.point = len(w.data) - 1
			}
			return token
		case w.state == 0 && !isLetterOrDigit(w.data[i]):
			if w.point+1 < len(w.data) {
				w.point++
			}
			continue
		}
	}
	var typ int
	if isLetterOrDigit(w.data[w.point]) {
		typ = WORD_LEXER_TOKEN_TYPE_WORD
	}
	token := Token{
		Start: w.point,
		End:   len(w.data) - 1,
		Type:  typ,
	}

	w.point = len(w.data) - 1

	return token

}

func NewWordLexer(data []byte) Lexer {
	return &WordLexer{data: data}
}
