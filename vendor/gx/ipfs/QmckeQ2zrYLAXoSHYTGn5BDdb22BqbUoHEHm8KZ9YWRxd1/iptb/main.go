package main

import (
	"fmt"
	"os"

	"github.com/ipfs/iptb/cli"
)

func main() {
	cli := cli.NewCli()

	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(cli.ErrWriter, "%s\n", err)
		os.Exit(1)
	}
}
