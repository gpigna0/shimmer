package main

import (
	"log"
	"math"
	"path"
	"strconv"
	"time"

	"github.com/gpigna0/shimmer/util"
)

const TIME = time.Millisecond * 100

func auto(pubChan chan<- string, action chan string) {
	// WARN: For now auto works only with one screen
	dir := path.Join(util.Conf.Sensor.Path, "in_illuminance_raw")
	maxBrg := util.Conf.Devices[0].Max
	bounds := util.Conf.Sensor.Bounds
	par := util.Conf.Sensor.Params
	c := par.Convexity

	tick := time.NewTicker(TIME)
	tick.Stop()
	working := false
	old, err := util.ReadFloat64(dir)
	if err != nil {
		log.Println("err:", err)
	}
	oldBrg := -1 // with this publish BRIGHTNESS only when it changes

	for {
		select {
		case v, ok := <-action:
			if !ok {
				pubChan <- "AUTO::false"
				tick.Stop()
				return
			}
			switch v {
			case "stop":
				tick.Stop()
				working = false
				pubChan <- "AUTO::false"
			case "start":
				oldBrg = -1
				working = true
				pubChan <- "AUTO::true"
				tick.Reset(TIME)
			case "publish":
				pubChan <- "AUTO::" + strconv.FormatBool(working)
			default:
				action <- strconv.FormatBool(working)
			}
		case <-tick.C:
			curr, err := util.ReadFloat64(dir)
			if err != nil {
				log.Println("err:", err)
				break
			}

			s := virtualSensor(curr, old, par)
			v := brightness(s, c, maxBrg, bounds)
			val := strconv.Itoa(v)

			if err := set(util.Conf.Devices[0], val); err != nil {
				log.Println("error setting brightness:", err)
			}

			if v != oldBrg {
				pubBrightness(pubChan)
			}
			old = s
			oldBrg = v
		}
	}
}

func virtualSensor(curr, old float64, params util.Params) float64 {
	delta := math.Abs(curr - old)
	weight := params.Evolution / (1 + delta/params.Smoothness)
	return weight*curr + (1-weight)*old
}

func brightness(sensor, convexity, maxBrg float64, bounds util.Bounds) int {
	xM := bounds.Max / (bounds.Max + convexity)
	xm := bounds.Min / (bounds.Min + convexity)

	a := (maxBrg + 1) / (xM - xm)
	b := 1 - a*xm
	curve := a * sensor / (sensor + convexity)

	return int(math.Round(curve + b))
}
