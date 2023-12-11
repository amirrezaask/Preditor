package lexers

const (
	WORD_LEXER_TOKEN_TYPE_WORD       = 1
	WORD_LEXER_TOKEN_TYPE_SYMBOL     = 2
	WORD_LEXER_TOKEN_TYPE_WHITESPACE = 3
)

const (
	wordLexer_insideWord        = 1
	wordLexer_insideWhitespaces = 2
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

func isSymbol(c byte) bool {
	return (c >= 33 && c <= 47) || (c >= 58 && c <= 64) || (c >= 91 && c <= 96) || (c >= 123 && c <= 126)
}

func (w *WordLexer) Tokens() []Token {
	var tokens []Token
	for i := 0; i < len(w.data); i++ {
		c := w.data[i]
		switch {
		case isLetterOrDigit(c):
			switch w.state {
			case wordLexer_insideWord, 0:
				w.state = wordLexer_insideWord
				continue
			case wordLexer_insideWhitespaces:
				tokens = append(tokens, Token{
					Start: w.point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WHITESPACE,
				})
				w.state = wordLexer_insideWord
				w.point = i

			}
		case isWhitespace(c):
			switch w.state {
			case wordLexer_insideWhitespaces, 0:
				continue
			case wordLexer_insideWord:
				tokens = append(tokens, Token{
					Start: w.point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WORD,
				})
				w.state = wordLexer_insideWhitespaces
				w.point = i
			}
		default:
			switch w.state {
			case wordLexer_insideWord:
				tokens = append(tokens, Token{
					Start: w.point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WORD,
				})
				w.point = i
				w.state = 0

			case wordLexer_insideWhitespaces:
				tokens = append(tokens, Token{
					Start: w.point,
					End:   i,
					Type:  WORD_LEXER_TOKEN_TYPE_WHITESPACE,
				})
				w.point = i
				w.state = 0
			}

			tokens = append(tokens, Token{
				Start: w.point,
				End:   i + 1,
				Type:  WORD_LEXER_TOKEN_TYPE_SYMBOL,
			})
			w.point = i + 1
		}
	}
	var typ int
	if w.state == wordLexer_insideWord {
		typ = WORD_LEXER_TOKEN_TYPE_WORD
	} else if w.state == wordLexer_insideWhitespaces {
		typ = WORD_LEXER_TOKEN_TYPE_WHITESPACE
	} else {
		typ = WORD_LEXER_TOKEN_TYPE_SYMBOL
	}

	tokens = append(tokens, Token{
		Start: w.point,
		End:   len(w.data) - 1,
		Type:  typ,
	})

	return tokens
}

func NewWordLexer(data []byte) Lexer {
	return &WordLexer{data: data}
}
