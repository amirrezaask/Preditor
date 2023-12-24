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
				Background:            mustParseHexColor("#000000"),
				Foreground:            mustParseHexColor("#a9a9a9"),
				SelectionBackground:   mustParseHexColor("#0000cd"),
				SelectionForeground:   mustParseHexColor("#ffffff"),
				Prompts:               mustParseHexColor("#333333"),
				StatusBarBackground:   mustParseHexColor("#696969"),
				StatusBarForeground:   mustParseHexColor("#000000"),
				LineNumbersForeground: mustParseHexColor("#F2F2F2"),
				ActiveWindowBorder:    mustParseHexColor("#292929"),
				Cursor:                mustParseHexColor("#00ff00"),
				CursorLineBackground:  mustParseHexColor("#52534E"),
				SyntaxColors: SyntaxColors{
					"ident":   mustParseHexColor("#eedd82"),
					"type":    mustParseHexColor("#8cde94"),
					"keyword": mustParseHexColor("#cd950c"),
					"string":  mustParseHexColor("#118a1a"),
					"comment": mustParseHexColor("#118a1a"),
				},
			},
		},
	},
	CursorLineHighlight:        true,
	TabSize:                    4,
	LineNumbers:                true,
	EnableSyntaxHighlighting:   true,
	CursorShape:                CURSOR_SHAPE_BLOCK,
	CursorBlinking:             false,
	FontName:                   "LiberationMono-Regular",
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
	case "font_size":
		var err error
		cfg.FontSize, err = strconv.Atoi(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadConfig(cfgPath string) (*Config, error) {
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
