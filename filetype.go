package preditor

import (
	"go/format"
)

type FileType struct {
	TabSize                  int
	BeforeSave               func(*Buffer) error
	AfterSave                func(*Buffer) error
	DefaultCompileCommand    string
	CommentLineBeginingChars []byte
	FindRootOfProject        func(currentFilePath string) (string, error)
	TSHighlightQuery         []byte
}

var FileTypes map[string]FileType

func init() {
	FileTypes = map[string]FileType{
		".go": {
			TabSize: 4,
			BeforeSave: func(e *Buffer) error {
				newBytes, err := format.Source(e.Content)
				if err != nil {
					return err
				}

				e.Content = newBytes
				return nil
			},
			TSHighlightQuery: []byte(`
[
  "break"
  "case"
  "chan"
  "const"
  "continue"
  "default"
  "defer"
  "else"
  "fallthrough"
  "for"
  "func"
  "go"
  "goto"
  "if"
  "import"
  "interface"
  "map"
  "package"
  "range"
  "return"
  "select"
  "struct"
  "switch"
  "type"
  "var"
] @keyword

(type_identifier) @type
(comment) @comment
[(interpreted_string_literal) (raw_string_literal)] @string
[(identifier)] @ident
(selector_expression operand: (_) @selector field: (_) @field)
(if_statement condition: (_) @if_condition)
`),
			AfterSave: func(buffer *Buffer) error {
				return CompileNoAsk(buffer)
			},
			DefaultCompileCommand: "go build -v ./...",
		},
	}
}
