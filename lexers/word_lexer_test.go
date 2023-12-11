package lexers

import (
	"fmt"
	"testing"
)

func TestWordLexer(t *testing.T) {
	var token Token
	data := "hello world0.3abc"
	lexer := NewWordLexer([]byte(data))
	for token.Type != 3 {
		token = lexer.Next()
		fmt.Printf("%+v'%s'\n", token, data[token.Start:token.End])
	}

}
