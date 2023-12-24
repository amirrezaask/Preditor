package main

import (
	"errors"
	"image/color"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Colors   Colors
	FontName string
	FontSize int
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
		Background:            mustParseHexColor("#333333"),
		Foreground:            mustParseHexColor("#F2F2F2"),
		SelectionBackground:   mustParseHexColor("#48B9C7"),
		SelectionForeground:   mustParseHexColor("#FFFFFF"),
		StatusBarBackground:   mustParseHexColor("#ffffff"),
		StatusBarForeground:   mustParseHexColor("#000000"),
		LineNumbersForeground: mustParseHexColor("#F2F2F2"),
	},
	FontName: "Consolas",
	FontSize: 20,
}

func addToConfig(cfg *Config, key string, value string) error {
	var err error
	switch key {
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
		cfg.Colors.SelectionBackground, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "selection_foreground":
		cfg.Colors.SelectionForeground, err = parseHexColor(value)
		if err != nil {
			return err
		}
	case "line_numbers_foreground":
		cfg.Colors.LineNumbersForeground, err = parseHexColor(value)
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
		splitted := strings.Split(line, " ")
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
