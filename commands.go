package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"slices"

	"github.com/gpigna0/shimmer/util"
	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	var devs []string

	return &cli.Command{
		Name:                   "get",
		Usage:                  "Display info about managed screens",
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "human-readable", Aliases: []string{"H"}, Usage: "Display the brightness as percentage"},
			&cli.BoolFlag{Name: "json", Aliases: []string{"j"}, Usage: "Format output as JSON"},
			&cli.IntFlag{Name: "precision", Aliases: []string{"p"}, Usage: "Number of decimals used to display percentage values. Ignored if -H is not set"},
			&cli.StringSliceFlag{Name: "device", Aliases: []string{"d"}, Usage: "Name of the device to display", Destination: &devs},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			hr := c.Bool("human-readable")

			fullStats := make([]Stats, 0)
			for _, d := range util.Conf.Devices {
				if len(devs) == 0 || slices.Contains(devs, d.Name) {
					stats, err := get(d, hr, c.Int("precision"))
					if err != nil {
						return err
					}
					fullStats = append(fullStats, stats)
				}
			}

			if err := parse(fullStats, hr, c.Bool("json")); err != nil {
				return err
			}

			return nil
		},
	}
}

func cmdSet() *cli.Command {
	var devs []string

	return &cli.Command{
		Name:  "set",
		Usage: "Set the brightness. At least one device must be specified with --device or --all",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "value", UsageText: "VALUE\nAvailable formats for VALUE are:\n\tN -> set brightness in absolute value\n\tN% -> set brightness as percentage\n\t±N% Increment or decrement brightness by N percent"},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "all", Usage: "set brightness for all devices"},
			&cli.StringSliceFlag{Name: "device", Aliases: []string{"d"}, Usage: "Name of the device to control", Destination: &devs},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.StringArg("value") == "" {
				return errors.New("argument error: you must specify a value")
			}

			if !c.Bool("all") && len(devs) == 0 {
				return errors.New("error: you must specify at least one target device")
			}

			connOk := true
			conn, err := net.Dial("unix", util.SOCK)
			if err != nil {
				log.Printf("could not connect to daemon, changes wont be communicated via IPC: %v", err)
				connOk = false
			} else {
				defer conn.Close()
			}

			if connOk && util.CheckAutoWithConn(conn) {
				return errors.New("can't set brightness when auto is active")
			}

			found := 0
			for _, s := range util.Conf.Devices {
				if c.Bool("all") || slices.Contains(devs, s.Name) {
					found++
					if err := set(s, c.StringArg("value")); err != nil {
						if connOk {
							if _, err := fmt.Fprintln(conn, "refresh"); err != nil {
								log.Printf("error while sending to daemon: %v", err)
							}
						}
						return err
					}
				}
			}

			if found != len(devs) {
				fmt.Printf("%d devices were not found: check if the specified names correspond to managed devices\n", len(devs)-found)
			}

			if connOk {
				if _, err := fmt.Fprintln(conn, "refresh"); err != nil {
					log.Printf("error while sending to daemon: %v", err)
				}
			}

			return nil
		},
	}
}

func cmdAuto() *cli.Command {
	return &cli.Command{
		Name:  "auto",
		Usage: "control the state of automatic brightness",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "stop", Aliases: []string{"s"}, Usage: "Stop auto brightness"},
			&cli.BoolFlag{Name: "toggle", Aliases: []string{"t"}, Usage: "Toggle auto brightness"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			conn, err := net.Dial("unix", util.SOCK)
			if err != nil {
				return fmt.Errorf("could not connect to daemon: %w", err)
			}
			defer conn.Close()

			msg := "auto"
			if c.Bool("stop") || c.Bool("toggle") && util.CheckAutoWithConn(conn) {
				msg += "!"
			}

			if _, err := fmt.Fprintln(conn, msg); err != nil {
				return fmt.Errorf("error while sending to daemon: %w", err)
			}

			return nil
		},
	}
}

func cmdDaemon() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "start the daemon",
		Action: func(ctx context.Context, c *cli.Command) error {
			if conn, err := net.Dial("unix", util.SOCK); err == nil {
				conn.Close()
				fmt.Println("A daemon is already running")
				return nil
			}

			if err := daemon(ctx); err != nil {
				return err
			}
			return nil
		},
	}
}

func cmdQuit() *cli.Command {
	return &cli.Command{
		Name:  "quit",
		Usage: "Quit the daemon",
		Action: func(ctx context.Context, c *cli.Command) error {
			conn, err := net.Dial("unix", util.SOCK)
			if err != nil {
				return fmt.Errorf("could not connect to daemon: %w", err)
			}
			defer conn.Close()

			if _, err := fmt.Fprintln(conn, "quit"); err != nil {
				return fmt.Errorf("error while sending to daemon: %w", err)
			}

			return nil
		},
	}
}
