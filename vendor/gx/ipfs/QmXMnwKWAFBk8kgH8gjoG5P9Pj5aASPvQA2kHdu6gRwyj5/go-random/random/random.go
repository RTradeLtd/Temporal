package main

import (
	"fmt"
	"os"
	"strconv"

	"gx/ipfs/QmQMxG9D52TirZd9eLA37nxiNspnMRkKbyPWrVAa1gvtSy/go-humanize"
	random "gx/ipfs/QmXMnwKWAFBk8kgH8gjoG5P9Pj5aASPvQA2kHdu6gRwyj5/go-random"
)

func main() {
	l := len(os.Args)
	if l != 2 && l != 3 {
		usageError()
	}

	countuint64, err := humanize.ParseBytes(os.Args[1])
	if err != nil {
		usageError()
	}
	count := int64(countuint64)

	if l == 2 {
		err = random.WriteRandomBytes(count, os.Stdout)
	} else {
		seed, err2 := strconv.ParseInt(os.Args[2], 10, 64)
		if err2 != nil {
			usageError()
		}
		err = random.WritePseudoRandomBytes(count, os.Stdout, seed)
	}

	if err != nil {
		die(err)
	}
}

func usageError() {
	fmt.Fprintf(os.Stderr, "Usage: %s <count> [<seed>]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "If <seed> is given, output <count> pseudo random bytes made from <seed> (from Go's math/rand)\n")
	fmt.Fprintf(os.Stderr, "Otherwise, output <count> random bytes (from Go's crypto/rand)\n")
	os.Exit(-1)
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v", err)
	os.Exit(-1)
}
