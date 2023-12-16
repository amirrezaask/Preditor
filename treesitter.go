package preditor

import (
	"context"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func TSHighlights(cfg *Config, queryString []byte, prev *sitter.Tree, code []byte) ([]highlight, *sitter.Tree, error) {
	var highlights []highlight
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), prev, code)
	if err != nil {
		return nil, nil, err
	}
	query, err := sitter.NewQuery(queryString, golang.GetLanguage())
	if err != nil {
		return nil, tree, err
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, tree.RootNode())
	for {
		qm, exists := qc.NextMatch()
		if !exists {
			break
		}
		for _, capture := range qm.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			if c, exists := cfg.CurrentThemeColors().SyntaxColors[captureName]; exists {
				highlights = append(highlights, highlight{
					start: int(capture.Node.StartByte()),
					end:   int(capture.Node.EndByte()),
					Color: c.ToColorRGBA(),
				})
			}
		}
	}

	return highlights, tree, nil
}
