package main

import (
	"os"
)

type Config struct {
	Colors   Colors
	FontName string
	FontSize string
}

func readConfig(cfgPath string) (*Config, error) {
	var cfg Config
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
	}

	return &cfg, nil
}
