// Package util contain utility functions used by other components of shimmer
package util

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"path"
	"strconv"
)

var SOCK = path.Join(os.Getenv("XDG_RUNTIME_DIR"), "shimmer.sock")

func CheckAutoWithConn(c net.Conn) bool {
	if _, err := fmt.Fprintln(c, "auto?"); err != nil {
		log.Println("err:", err)
		return false
	}

	s := bufio.NewScanner(c)
	s.Scan()
	return s.Text() == "true"
}

func CheckAuto() bool {
	conn, err := net.Dial("unix", SOCK)
	if err != nil {
		// daemon assumed inactive
		return false
	}
	defer conn.Close()

	return CheckAutoWithConn(conn)
}

func ReadFloat64(pth string) (float64, error) {
	f, err := os.Open(pth)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)
	scanner.Scan()
	res := scanner.Text()

	if res == "" {
		return 0, nil
	}
	result, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func ToPercent(v, max float64, prec int) float64 {
	v = 100 * v / max
	if prec >= 0 {
		decs := math.Pow(10, float64(prec))
		v = math.Round(v*decs) / decs
	}
	return v
}
