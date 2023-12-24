package main

import (
	"errors"
	"image/color"
	"os"
	"strconv"
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

func readConfig(cfgPath string) (*Config, error) {
	cfg := defaultConfig
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		return &cfg, nil
	}
	bs, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var currentConfigKey string
	var currentConfigValue string
	for _, b := range bs {
		switch {
		case b != ' ' && currentConfigKey == "":
			currentConfigKey += string(b)
		case b != ' ' && currentConfigKey != "":
			currentConfigValue += string(b)
		}
	}

	switch currentConfigKey {
	case "font":
		cfg.FontName = currentConfigValue
	case "font_size":
		var err error
		cfg.FontSize, err = strconv.Atoi(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "background":
		cfg.Colors.Background, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "foreground":
		cfg.Colors.Foreground, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "statusbar_background":
		cfg.Colors.StatusBarBackground, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "statusbar_foreground":
		cfg.Colors.StatusBarForeground, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "selection_background":
		cfg.Colors.SelectionBackground, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "selection_foreground":
		cfg.Colors.SelectionForeground, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	case "line_numbers_foreground":
		cfg.Colors.LineNumbersForeground, err = parseHexColor(currentConfigValue)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
