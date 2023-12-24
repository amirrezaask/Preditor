package main

import "go/format"

type FileType struct {
	BeforeSave func(*Editor) error
}

var GoFileType = FileType{
	BeforeSave: func(e *Editor) error {
		newBytes, err := format.Source(e.Content)
		if err != nil {
			return err
		}

		e.Content = newBytes
		return nil
	},
}

var fileTypeMapping = map[string]FileType{
	".go": GoFileType,
}
