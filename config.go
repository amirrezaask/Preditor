package main

import (
	"errors"
	"image/color"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Colors                   Colors
	LineNumbers              bool
	FontName                 string
	FontSize                 int
	CursorShape              int
	CursorBlinking           bool
	EnableSyntaxHighlighting bool
}

func mustParseHexColor(hex string) color.RGBA {
	c, err := parseHexColor(hex)
	if err != nil {
		panic(err)
	}
	return c
}

var defaultConfig = Config{
	Colors: Colors{
		Background:            mustParseHexColor("#141414"),
		Foreground:            mustParseHexColor("#999999"),
		Selection:             mustParseHexColor("#0000cd"),
		StatusBarBackground:   mustParseHexColor("#ffffff"),
		StatusBarForeground:   mustParseHexColor("#000000"),
		LineNumbersForeground: mustParseHexColor("#F2F2F2"),
		Cursor:                mustParseHexColor("#00ff00"),
		CursorLineBackground:  mustParseHexColor("#52534E"),
		SyntaxKeywords:        mustParseHexColor("#cd950c"),
		SyntaxTypes:           mustParseHexColor("#8cde94"),
	},
	LineNumbers:              false,
	EnableSyntaxHighlighting: true,
	CursorShape:              CURSOR_SHAPE_BLOCK,
	CursorBlinking:           false,
	FontName:                 "Consolas",
	FontSize:                 30,
}

func addToConfig(cfg *Config, key string, value string) error {
	var err error
	switch key {
	case "syntax":
		cfg.EnableSyntaxHighlighting = value == "true"
	case "cursor_shape":
		switch value {
		case "block":
			cfg.CursorShape = CURSOR_SHAPE_BLOCK

		case "bar":
			cfg.CursorShape = CURSOR_SHAPE_LINE

		case "outline":
			cfg.CursorShape = CURSOR_SHAPE_OUTLINE
		}

	case "cursor_blinking":
		cfg.CursorBlinking = value == "true"
	case "font":
		cfg.FontName = value
	case "font_size":
		var err error
		cfg.FontSize, err = strconv.Atoi(value)
		if err != nil {
			return err
		}
	case "background":
		cfg.Colors.Background, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "foreground":
		cfg.Colors.Foreground, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "statusbar_background":
		cfg.Colors.StatusBarBackground, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "statusbar_foreground":
		cfg.Colors.StatusBarForeground, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "selection_background":
		cfg.Colors.Selection, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "line_numbers_foreground":
		cfg.Colors.LineNumbersForeground, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "cursor_background":
		cfg.Colors.Cursor, err = parseHexColor(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func readConfig(cfgPath string) (*Config, error) {
	cfg := defaultConfig
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		return &cfg, nil
	}
	bs, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bs), "\n")

	for _, line := range lines {
		splitted := strings.SplitN(line, " ", 2)
		if len(splitted) != 2 {
			continue
		}
		key := splitted[0]
		value := splitted[1]
		key = strings.Trim(key, " \t\r")
		value = strings.Trim(value, " \t\r")
		addToConfig(&cfg, key, value)
	}

	return &cfg, nil
}
