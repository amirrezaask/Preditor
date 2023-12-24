package main

import (
	"fmt"
	"go/format"
	"image/color"
	"regexp"
	"strings"
)

type FileType struct {
	BeforeSave func(*Editor) error
	SyntaxHighlights
}

type SyntaxHighlights struct {
	Keywords    SyntaxHighlight
	Types       SyntaxHighlight
	Identifiers SyntaxHighlight
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
			BeforeSave: func(e *Editor) error {
				newBytes, err := format.Source(e.Content)
				if err != nil {
					return err
				}

				e.Content = newBytes
				return nil
			},

			SyntaxHighlights: SyntaxHighlights{
				Keywords: SyntaxHighlight{
					Regex: regexp.MustCompile(keywordsPat("if", "struct", "type", "interface", "else", "func", "package", "import")),
					Color: cfg.SyntaxKeywords,
				},
				Types: SyntaxHighlight{
					Regex: regexp.MustCompile(keywordsPat("int8", "int16", "int32", "int64", "int")),
					Color: cfg.SyntaxTypes,
				},
				Identifiers: SyntaxHighlight{
					Regex: regexp.MustCompile("\\b[a-zA-Z_][a-zA-Z0-9_]*\\b"),
					Color: cfg.SyntaxIdentifiers,
				},
			},
		},
	}
}

var fileTypeMappings map[string]FileType
