package main

import (
	"log"
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gpigna0/shimmer/util"
)

const TIME = time.Millisecond * 100

type devInfo struct {
	util.Device
	old    float64
	oldBrg int
}

func autoHandler(pubChan chan<- string, action chan string) {
	dir := path.Join(util.Conf.Sensor.Path, "in_illuminance_raw")
	devs := make(map[string]devInfo)
	tick := time.NewTicker(TIME)

	for {
		select {
		case v, ok := <-action:
			if !ok {
				tick.Stop()
				return
			}
			args := strings.SplitN(v, " ", 2)
			switch args[0] {
			case "stop":
				delete(devs, args[1])
			case "start":
				devs[args[1]] = createDev(args[1])
			case "toggle":
				if _, ok := devs[args[1]]; !ok {
					devs[args[1]] = createDev(args[1])
					action <- "true"
				} else {
					delete(devs, args[1])
					action <- "false"
				}
			case "get":
				_, ok := devs[args[1]]
				action <- strconv.FormatBool(ok)
			}
		case <-tick.C:
			curr, err := util.ReadFloat64(dir)
			if err != nil {
				log.Println("err:", err)
				break
			}

			for k, d := range devs {
				s := virtualSensor(curr, d.old, util.Conf.Sensor.Params)
				v := brightness(s, util.Conf.Sensor.Params.Convexity, d.Max, util.Conf.Sensor.Bounds)
				val := strconv.Itoa(v)

				if err := set(d.Device, val); err != nil {
					log.Println("error setting brightness:", err)
				}

				if v != d.oldBrg {
					pubBrightness(k, pubChan)
				}
				d.old = s
				d.oldBrg = v
				devs[k] = d
			}
		}
	}
}

func createDev(devName string) devInfo {
	d := util.Conf.Devices[devName] // existence of the device should already be checked
	init, err := util.ReadFloat64(path.Join(util.Conf.Sensor.Path, "in_illuminance_raw"))
	if err != nil {
		log.Println("err:", err)
	}
	return devInfo{d, init, -1}
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
