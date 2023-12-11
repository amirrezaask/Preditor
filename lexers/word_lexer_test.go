package lexers

import (
	"fmt"
	"testing"
)

func TestWordLexer(t *testing.T) {
	data := "\n\nhello world0.3abc"
	lexer := NewWordLexer([]byte(data))
	tokens := lexer.Tokens()
	for _, token := range tokens {
		fmt.Printf("%+v'%s'\n", token, data[token.Start:token.End])
	}

}
