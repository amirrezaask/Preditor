package lexers

// Token is the basic data type of lexer
// buffer[Start:End)
type Token struct {
	Start int
	End   int
	Type  int
}

// Lexer produces stream tokens.
// We are already doing a single sweep on the buffer to split it
// into visual lines that user see, we can do better by paying the
// same price but we get a language aware
// tokens instead of visual lines and we can operate on tokens.
type Lexer interface {
	Tokens() []Token
}
