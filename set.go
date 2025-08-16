package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	"github.com/gpigna0/shimmer/util"
)

var rgx = regexp.MustCompile(`^([+-]?)(\d+(?:\.\d+)?)%$|^(\d+)$`)

func set(screen util.Device, value string) error {
	matches := rgx.FindStringSubmatch(value)
	if len(matches) == 0 {
		return errors.New(
			"argument parsing error (" +
				value +
				"):\n\tthe value to set must be in the form\n\t[+ | -]N.n% or N")
	}

	val := ""
	switch {
	case matches[3] != "":
		val = matches[3]
	case matches[1] != "":
		v, err := percentIncrement(screen, matches[1]+matches[2])
		if err != nil {
			return err
		}
		val = v
	default:
		v, err := percentBrightness(screen, matches[2])
		if err != nil {
			return err
		}
		val = v
	}

	if err := os.WriteFile(screen.Path, []byte(val), 0644); err != nil {
		return fmt.Errorf("error writing new brightness %s: %w", val, err)
	}

	return nil
}

func percentBrightness(screen util.Device, val string) (string, error) {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return "", err
	}

	switch {
	case v > 100:
		v = 100
	case v < 0:
		v = 0
	}
	v = math.Round(screen.Max * v / 100)

	return strconv.Itoa(int(v)), nil
}

func percentIncrement(screen util.Device, amt string) (string, error) {
	curr, err := util.ReadFloat64(screen.Path)
	if err != nil {
		return "", err
	}
	v, err := strconv.ParseFloat(amt, 64)
	if err != nil {
		return "", err
	}

	new := math.Round(curr + (screen.Max * v / 100))
	switch {
	case new < 0:
		return "0", nil
	case new > screen.Max:
		return strconv.Itoa(int(screen.Max)), nil
	default:
		return strconv.Itoa(int(new)), nil
	}
}
