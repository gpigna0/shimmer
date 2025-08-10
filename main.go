package main

import (
	"context"
	"log"
	"os"

	"github.com/gpigna0/shimmer/util"
	"github.com/urfave/cli/v3"
)

func main() {
	if err := util.InitConfig(); err != nil {
		log.Fatalf("shimmer failed with errors:\n\n%v\n", err)
	}

	get := cmdGet()
	set := cmdSet()
	auto := cmdAuto()
	daemon := cmdDaemon()
	quit := cmdQuit()

	cmd := cli.Command{
		Name:     "shimmer",
		Usage:    "Control your screen brightness",
		Commands: []*cli.Command{get, set, auto, daemon, quit},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("shimmer failed with errors:\n\n%v\n", err)
	}
}
