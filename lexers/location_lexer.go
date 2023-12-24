package lexers

const (
	locationLexer_InFileName = 0
	locationLexer_InLine     = 1
	locationLexer_InColumn   = 2
)

type LocationLexer struct {
	data  []byte
	pos   int
	state int
}

func NewLocationLexer(data []byte) Lexer {
	return &LocationLexer{data: data}
}

func (l *LocationLexer) Tokens() []Token {
	var tokens []Token
	//	 currentTokenFilenameStart := -1
	//	 for i := 0; i < len(l.data); i++ {
	//		 c := l.data[i]
	//		 switch l.state {
	//		 case locationlexer_InFileName:
	//			 switch {
	//			 case IsLetterOrDigit(c):
	//				 if currentTokenFilenameStart == -1 {
	//					 currentTokenFilenameStart = i
	//				 }
	//			 case c == ':':
	//
	//			 }
	//		 case locationlexer_InLine:
	//		 case locationlexer_InColumn:
	//		 }
	//
	//	 }
	return tokens
}
