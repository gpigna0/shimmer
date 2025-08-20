package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gpigna0/shimmer/util"
)

type Stats struct {
	Device     string  `json:"device_name"`
	Type       string  `json:"type"`
	Path       string  `json:"path"`
	Max        int     `json:"max_brigthness"`
	Brightness float64 `json:"brightness"`
	Auto       bool    `json:"auto"`
}

func get(dev util.Device, humanReadable bool, precision int) (Stats, error) {
	brg, err := util.ReadFloat64(dev.Path)
	if err != nil {
		return Stats{}, err
	}
	auto := util.CheckAuto(dev.Name)

	if humanReadable {
		brg = util.ToPercent(brg, dev.Max, precision)
	}

	typ := "screen"
	if dev.Type == 1 {
		typ = "led"
	}

	stats := Stats{
		dev.Name,
		typ,
		dev.Path,
		int(dev.Max),
		brg,
		auto,
	}

	return stats, nil
}

func parse(stats []Stats, humanReadable bool, jsonFmt bool) error {
	if jsonFmt {
		out, err := json.MarshalIndent(stats, "", "\t")
		if err != nil {
			return fmt.Errorf("error while marshaling statistics: %w", err)
		}
		fmt.Println(string(out))
		return nil
	}

	strStats := make([]string, len(stats))
	percent := ""
	if humanReadable {
		percent = "%"
	}
	for i, s := range stats {
		strStats[i] = fmt.Sprintf(
			"Device: %s\nPath: %s\nMax Brightness: %d\nBrightness: %v%s\nAuto brightness: %t\n",
			s.Device,
			s.Path,
			s.Max,
			s.Brightness,
			percent,
			s.Auto,
		)
	}

	fmt.Println(strings.Join(strStats, "\n\n\n"))
	return nil
}
