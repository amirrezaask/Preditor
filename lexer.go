package preditor

// Token is the basic data type of lexer
// buffer[Start:End)
type Token struct {
	Data  []byte
	Start int
	End   int
}

// Lexer produces tokens.
// We are already doing a single sweep on the buffer to split it
// into visual lines that user see
// we can do better by paying the same price but we get a language aware
// tokens instead of visual lines and we can operate on tokens.
type Lexer interface {
	Tokens(bs []byte) []Token
}
