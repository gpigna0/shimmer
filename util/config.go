package util

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Sensor Sensor
	Screen []Screen
}

type Sensor struct {
	Path   string
	Bounds Bounds
	Params Params
}

type Bounds struct {
	Min float64
	Max float64
}

type Params struct {
	Evolution  float64
	Smoothness float64
	Convexity  float64
}

type Screen struct {
	Name string
	Path string
	Max  float64
}

var Conf Config

func InitConfig() error {
	// WARN: Must be called at main()
	configPath := os.Getenv("HOME")
	if configPath == "" {
		return errors.New("Config error: HOME environment variable is not set")
	}

	configPath = path.Join(configPath, ".config", "shimmer", "config.toml")

	if _, err := toml.DecodeFile(configPath, &Conf); err != nil {
		return fmt.Errorf("Config error: %w", err)
	}

	ex, err := PathExists(Conf.Sensor.Path)
	if err != nil {
		return fmt.Errorf("Config error: %w", err)
	} else if !ex {
		Conf.Sensor.Path = ""
	}

	for i, s := range Conf.Screen {
		m, err := ReadFloat64(path.Join(s.Path, "max_brightness"))
		if err != nil {
			return fmt.Errorf("Config error: %w", err)
		}
		Conf.Screen[i].Max = m
		Conf.Screen[i].Path = path.Join(Conf.Screen[i].Path, "brightness")
	}

	if err := validate(Conf); err != nil {
		return fmt.Errorf("Config error: %w", err)
	}

	return nil
}

func validate(c Config) error {
	s := c.Sensor
	msg := ""
	switch {
	case s.Bounds.Max < s.Bounds.Min:
		msg = "sensor.bounds.max is less than sensor.bounds.min"
	case s.Bounds.Min < 0 || s.Bounds.Max < 0:
		msg = "both sensor.bounds must be >= 0"
	case s.Params.Evolution <= 0 || s.Params.Evolution > 1:
		msg = "sensor.params.evolution must be between 0 and 1, 0 excluded"
	case s.Params.Smoothness <= 0:
		msg = "sensor.params.smoothness must be > 0"
	case s.Params.Convexity <= 0:
		msg = "sensor.params.convexity must be > 0"
	}

	if msg != "" {
		return errors.New("config error: " + msg)
	} else {
		return nil
	}
}
