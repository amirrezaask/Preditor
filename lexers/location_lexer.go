package lexers

type LocationLexer struct {
	data  []byte
	pos   int
	state int
}

func NewLocationLexer(data []byte) Lexer {
	return &LocationLexer{data: data}
}

func (l *LocationLexer) Tokens() []Token {
	return nil
}
