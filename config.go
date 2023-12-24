package preditor

import (
	_ "embed"
	"errors"
	"fmt"
	"reflect"

	"image/color"
	"os"
	"strconv"
	"strings"
)

//go:embed fonts/liberationmono-regular.ttf
var liberationMonoRegularTTF []byte

//go:embed fonts/jetbrainsmono.ttf
var jetbrainsMonoTTF []byte

type CursorShape int

const (
	CURSOR_SHAPE_BLOCK   CursorShape = 1
	CURSOR_SHAPE_OUTLINE CursorShape = 2
	CURSOR_SHAPE_LINE    CursorShape = 3
)

func (c CursorShape) String() string {
	switch c {
	case CURSOR_SHAPE_BLOCK:
		return "block"
	case CURSOR_SHAPE_OUTLINE:
		return "outline"
	case CURSOR_SHAPE_LINE:
		return "bar"
	default:
		return ""
	}
}

type Theme struct {
	Name   string
	Colors Colors
}

func (t Theme) String() string {
	var colors []string
	v := reflect.ValueOf(t.Colors)
	typ := reflect.TypeOf(t.Colors)
	for i := 0; i < v.NumField(); i++ {
		colors = append(colors, typ.Field(i).Name, v.Field(i).String())
	}
	//return fmt.Sprintf("Theme: %s\n%s", t.Name, strings.Join(colors, "\n"))
	return t.Name
}

type Config struct {
	Themes                     []Theme
	CurrentTheme               string
	TabSize                    int
	LineNumbers                bool
	FontName                   string
	FontSize                   int
	CursorShape                CursorShape
	CursorBlinking             bool
	EnableSyntaxHighlighting   bool
	HighlightMatchingParen     bool
	CursorLineHighlight        bool
	BuildWindowNormalHeight    float64
	BuildWindowMaximizedHeight float64
}

func (c *Config) String() string {
	var output []string
	v := reflect.ValueOf(c).Elem()
	t := reflect.TypeOf(c).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typ := t.Field(i)
		if typ.Type.String() == "color.RGBA" {
			rgbaColor := v.Interface().(color.RGBA)
			colorAsHex := fmt.Sprintf("#%02x%02x%02x%02x", rgbaColor.R, rgbaColor.G, rgbaColor.B, rgbaColor.A)
			output = append(output, fmt.Sprintf("%s = %s", typ.Name, colorAsHex))

		} else {
			output = append(output, fmt.Sprintf("%s = %v", typ.Name, field.Interface()))
		}
	}

	return strings.Join(output, "\n")
}

func mustParseHexColor(hex string) RGBA {
	c, err := parseHexColor(hex)
	if err != nil {
		panic(err)
	}
	return RGBA(c)
}

var defaultConfig = Config{
	CurrentTheme: "Default",
	Themes: []Theme{
		{
			Name: "Default",
			Colors: Colors{
				Background:                mustParseHexColor("#0c0c0c"),
				Foreground:                mustParseHexColor("#90B090"),
				SelectionBackground:       mustParseHexColor("#FF44DD"),
				SelectionForeground:       mustParseHexColor("#ffffff"),
				Prompts:                   mustParseHexColor("#333333"),
				StatusBarBackground:       mustParseHexColor("#888888"),
				StatusBarForeground:       mustParseHexColor("#000000"),
				ActiveStatusBarBackground: mustParseHexColor("#BBBBBB"),
				ActiveStatusBarForeground: mustParseHexColor("#000000"),
				LineNumbersForeground:     mustParseHexColor("#F2F2F2"),
				ActiveWindowBorder:        mustParseHexColor("#292929"),
				Cursor:                    mustParseHexColor("#00ff00"),
				CursorLineBackground:      mustParseHexColor("#52534E"),
				HighlightMatching:         mustParseHexColor("#00ff00"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#90B090"),
					"type":    mustParseHexColor("#90B090"),
					"keyword": mustParseHexColor("#D08F20"),
					"string":  mustParseHexColor("#50FF30"),
					"comment": mustParseHexColor("#2090F0"),
				},
			},
		},
		{
			Name: "Light",
			Colors: Colors{
				Background:            mustParseHexColor("#ffffff"),
				Foreground:            mustParseHexColor("#000000"),
				SelectionBackground:   mustParseHexColor("#ADD6FF"),
				SelectionForeground:   mustParseHexColor("#000000"),
				Prompts:               mustParseHexColor("#333333"),
				StatusBarBackground:   mustParseHexColor("#696969"),
				StatusBarForeground:   mustParseHexColor("#000000"),
				LineNumbersForeground: mustParseHexColor("#010101"),
				ActiveWindowBorder:    mustParseHexColor("#8cde94"),
				Cursor:                mustParseHexColor("#171717"),
				CursorLineBackground:  mustParseHexColor("#52534E"),
				HighlightMatching:     mustParseHexColor("#171717"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#000000"),
					"type":    mustParseHexColor("#0000ff"),
					"keyword": mustParseHexColor("#0000ff"),
					"string":  mustParseHexColor("#a31515"),
					"comment": mustParseHexColor("#008000"),
				},
			},
		},
		{
			Name: "4Coder_Fleury",
			Colors: Colors{
				Background:                mustParseHexColor("#020202"),
				Foreground:                mustParseHexColor("#b99468"),
				SelectionBackground:       mustParseHexColor("#FF44DD"),
				SelectionForeground:       mustParseHexColor("#ffffff"),
				Prompts:                   mustParseHexColor("#333333"),
				StatusBarBackground:       mustParseHexColor("#000000"),
				StatusBarForeground:       mustParseHexColor("#ffa900"),
				ActiveStatusBarBackground: mustParseHexColor("#000000"),
				ActiveStatusBarForeground: mustParseHexColor("#ffa933"),
				LineNumbersForeground:     mustParseHexColor("#010101"),
				ActiveWindowBorder:        mustParseHexColor("#8cde94"),
				Cursor:                    mustParseHexColor("#e0741b"),
				CursorLineBackground:      mustParseHexColor("#52534E"),
				HighlightMatching:         mustParseHexColor("#e0741b"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#90B090"),
					"type":    mustParseHexColor("#f0c674"),
					"keyword": mustParseHexColor("#f0c674"),
					"string":  mustParseHexColor("#ffa900"),
					"comment": mustParseHexColor("#666666"),
				},
			},
		},
		{
			Name: "Naysayer",
			Colors: Colors{
				Background:                mustParseHexColor("#072626"),
				Foreground:                mustParseHexColor("#d3b58d"),
				SelectionBackground:       mustParseHexColor("#0000ff"),
				SelectionForeground:       mustParseHexColor("#d3b58d"),
				Prompts:                   mustParseHexColor("#333333"),
				StatusBarBackground:       mustParseHexColor("#ffffff"),
				StatusBarForeground:       mustParseHexColor("#000000"),
				ActiveStatusBarBackground: mustParseHexColor("#d3b58d"),
				ActiveStatusBarForeground: mustParseHexColor("#000000"),
				LineNumbersForeground:     mustParseHexColor("#d3b58d"),
				ActiveWindowBorder:        mustParseHexColor("#8cde94"),
				Cursor:                    mustParseHexColor("#90ee90"),
				CursorLineBackground:      mustParseHexColor("#52534E"),
				HighlightMatching:         mustParseHexColor("#90ee90"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#c8d4ec"),
					"type":    mustParseHexColor("#8cde94"),
					"keyword": mustParseHexColor("#d4d4d4"),
					"string":  mustParseHexColor("#0fdfaf"),
					"comment": mustParseHexColor("#3fdf1f"),
				},
			},
		},
		{
			Name: "Solarized_Dark",
			Colors: Colors{
				Background:                mustParseHexColor("#002B36"),
				Foreground:                mustParseHexColor("#839496"),
				SelectionBackground:       mustParseHexColor("#274642"),
				SelectionForeground:       mustParseHexColor("#d3b58d"),
				Prompts:                   mustParseHexColor("#333333"),
				StatusBarBackground:       mustParseHexColor("#00212B"),
				StatusBarForeground:       mustParseHexColor("#93A1A1"),
				ActiveStatusBarBackground: mustParseHexColor("#00212B"),
				ActiveStatusBarForeground: mustParseHexColor("#93A1A1"),
				LineNumbersForeground:     mustParseHexColor("#d3b58d"),
				ActiveWindowBorder:        mustParseHexColor("#8cde94"),
				Cursor:                    mustParseHexColor("#D30102"),
				CursorLineBackground:      mustParseHexColor("#073642"),
				HighlightMatching:         mustParseHexColor("#cdcdcd"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#268BD2"),
					"type":    mustParseHexColor("#CB4B16"),
					"keyword": mustParseHexColor("#859900"),
					"string":  mustParseHexColor("#2AA198"),
					"comment": mustParseHexColor("#586E75"),
				},
			},
		},
		{
			Name: "Solarized_Light",
			Colors: Colors{
				Background:                mustParseHexColor("#FDF6E3"),
				Foreground:                mustParseHexColor("#657B83"),
				SelectionBackground:       mustParseHexColor("#EEE8D5"),
				SelectionForeground:       mustParseHexColor("#d3b58d"),
				Prompts:                   mustParseHexColor("#333333"),
				StatusBarBackground:       mustParseHexColor("#EEE8D5"),
				StatusBarForeground:       mustParseHexColor("#586E75"),
				ActiveStatusBarBackground: mustParseHexColor("#EEE8D5"),
				ActiveStatusBarForeground: mustParseHexColor("#586E75"),
				LineNumbersForeground:     mustParseHexColor("#d3b58d"),
				ActiveWindowBorder:        mustParseHexColor("#8cde94"),
				Cursor:                    mustParseHexColor("#657B83"),
				CursorLineBackground:      mustParseHexColor("#EEE8D5"),
				HighlightMatching:         mustParseHexColor("#cdcdcd"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#268BD2"),
					"type":    mustParseHexColor("#CB4B16"),
					"keyword": mustParseHexColor("#859900"),
					"string":  mustParseHexColor("#2AA198"),
					"comment": mustParseHexColor("#93A1A1"),
				},
			},
		},
	},
	CursorLineHighlight:        false,
	TabSize:                    4,
	LineNumbers:                true,
	EnableSyntaxHighlighting:   true,
	CursorShape:                CURSOR_SHAPE_BLOCK,
	CursorBlinking:             false,
	FontName:                   "LiberationMono-Regular",
	HighlightMatchingParen:     true,
	FontSize:                   17,
	BuildWindowNormalHeight:    0.2,
	BuildWindowMaximizedHeight: 0.5,
}

func (c *Config) CurrentThemeColors() *Colors {
	for _, theme := range c.Themes {
		if theme.Name == c.CurrentTheme {
			return &theme.Colors
		}
	}
	return &c.Themes[0].Colors
}

func addToConfig(cfg *Config, key string, value string) error {
	switch key {
	case "syntax":
		cfg.EnableSyntaxHighlighting = value == "true"
	case "theme":
		cfg.CurrentTheme = value
	case "cursor_shape":
		switch value {
		case "block":
			cfg.CursorShape = CURSOR_SHAPE_BLOCK

		case "bar":
			cfg.CursorShape = CURSOR_SHAPE_LINE

		case "outline":
			cfg.CursorShape = CURSOR_SHAPE_OUTLINE
		}
	case "line_numbers":
		cfg.LineNumbers = value == "true"
	case "cursor_blinking":
		cfg.CursorBlinking = value == "true"
	case "font":
		cfg.FontName = value
	case "cursor_line_highlight":
		cfg.CursorLineHighlight = value == "true"
	case "hl_matching_char":
		cfg.HighlightMatchingParen = value == "true"
	case "font_size":

		var err error
		cfg.FontSize, err = strconv.Atoi(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadConfig(cfgPath string, startTheme string) (*Config, error) {
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

	if startTheme != "" {
		cfg.CurrentTheme = startTheme
	}

	return &cfg, nil
}
