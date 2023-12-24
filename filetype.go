package main

import (
	"go/format"
	"image/color"
	"regexp"
)

type FileType struct {
	BeforeSave  func(*Editor) error
	ColorGroups map[*regexp.Regexp]color.RGBA
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
			ColorGroups: map[*regexp.Regexp]color.RGBA{
				regexp.MustCompile("(func|if|package|import)"): cfg.SyntaxKeyword,
			},
		},
	}
}

var fileTypeMappings map[string]FileType
