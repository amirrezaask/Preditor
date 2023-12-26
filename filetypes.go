package preditor

import (
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/php"
	"go/format"
)

/*
   Treesitter captures:
   - type
   - string
   - ident
   - function_name
*/

var GoFileType = FileType{
	Name:       "Go",
	TabSize:    4,
	TSLanguage: golang.GetLanguage(),
	BeforeSave: func(e *BufferView) error {
		newBytes, err := format.Source(e.Buffer.Content)
		if err != nil {
			return err
		}

		e.Buffer.Content = newBytes
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
(function_declaration name: (_) @function_name)
(call_expression function: (_) @function_name)
`),
	DefaultCompileCommand: "go build -v ./...",
}

var PHPFileType = FileType{
	Name:             "PHP",
	TabSize:          4,
	TSLanguage:       php.GetLanguage(),
	TSHighlightQuery: []byte(``),
}
