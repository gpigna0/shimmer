package util

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Sensor  Sensor
	Devices map[string]Device
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
	Smoothness float64
	Convexity  float64
}

type Device struct {
	Type DevType
	Name string
	Path string
	Max  float64
}

type DevType uint

const (
	SCREEN DevType = iota
	LED
)

func (d DevType) String() string {
	if d == 0 {
		return "screen"
	} else {
		return "led"
	}
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
	Conf.Devices = make(map[string]Device)

	ex, err := PathExists(Conf.Sensor.Path)
	if err != nil {
		return fmt.Errorf("Config error: %w", err)
	} else if !ex {
		Conf.Sensor.Path = ""
	}

	if err := validate(Conf); err != nil {
		return fmt.Errorf("Config error: %w", err)
	}

	if err := findDevs(); err != nil {
		return fmt.Errorf("Config error: %w", err)
	}

	return nil
}

func findDevs() error {
	scr := "/sys/class/backlight"
	led := "/sys/class/leds"
	scrDir, err := os.ReadDir(scr)
	if err != nil {
		return err
	}
	ledDir, err := os.ReadDir(led)
	if err != nil {
		return err
	}

	for _, v := range scrDir {
		maxBrg, err := ReadFloat64(path.Join(scr, v.Name(), "max_brightness"))
		if err != nil {
			log.Printf("error while seaching for devices: %s won't be registered -- %v", v.Name(), err)
			continue
		}
		Conf.Devices[v.Name()] = Device{SCREEN, v.Name(), path.Join(scr, v.Name(), "brightness"), maxBrg}
	}

	for _, v := range ledDir {
		maxBrg, err := ReadFloat64(path.Join(led, v.Name(), "max_brightness"))
		if err != nil {
			log.Printf("error while seaching for devices: %s won't be registered -- %v", v.Name(), err)
			continue
		}
		Conf.Devices[v.Name()] = Device{SCREEN, v.Name(), path.Join(led, v.Name(), "brightness"), maxBrg}
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
	case s.Params.Smoothness < 1:
		msg = "sensor.params.smoothness must be >= 1"
	case s.Params.Convexity <= 0:
		msg = "sensor.params.convexity must be > 0"
	}

	if msg != "" {
		return errors.New("config error: " + msg)
	} else {
		return nil
	}
}
