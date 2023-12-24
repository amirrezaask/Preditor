package main

import (
	"fmt"
	"go/format"
	"image/color"
	"regexp"
	"strings"
)

type FileType struct {
	TabSize          int
	BeforeSave       func(*Editor) error
	SyntaxHighlights *SyntaxHighlights
}

type SyntaxHighlights struct {
	Keywords SyntaxHighlight
	Types    SyntaxHighlight
}

type SyntaxHighlight struct {
	Regex *regexp.Regexp
	Color color.RGBA
}

func keywordPat(word string) string {
	return fmt.Sprintf("\\b%s\\b", word)
}
func keywordsPat(words ...string) string {
	var pats []string
	for _, word := range words {
		pats = append(pats, keywordPat(word))
	}

	return fmt.Sprintf("(%s)", strings.Join(pats, "|"))
}

func initFileTypes(cfg Colors) {
	fileTypeMappings = map[string]FileType{
		".go": {
			TabSize: 4,
			BeforeSave: func(e *Editor) error {
				newBytes, err := format.Source(e.Content)
				if err != nil {
					return err
				}

				e.Content = newBytes
				return nil
			},

			SyntaxHighlights: &SyntaxHighlights{
				Keywords: SyntaxHighlight{
					Regex: regexp.MustCompile(keywordsPat("break", "case", "const",
						"continue", "default", "defer", "else", "fallthrough", "for", "func", "go", "goto", "if",
						"import", "interface", "package", "range", "return", "select", "struct", "switch", "type", "var", "len", "nil", "iota", "append", "cap", "clear", "close", "complex",
						"copy", "delete", "imag", "len", "make",
						"max", "min", "new", "panic", "print",
						"println", "real", "recover")),
					Color: cfg.SyntaxKeywords,
				},
				Types: SyntaxHighlight{
					Regex: regexp.MustCompile(keywordsPat("u*int8", "u*int16", "u*int32", "u*int64", "u*int", "float(32|64)", "bool", "true", "false", "chan", "byte", "map")),
					Color: cfg.SyntaxTypes,
				},
			},
		},
	}
}

var fileTypeMappings map[string]FileType
