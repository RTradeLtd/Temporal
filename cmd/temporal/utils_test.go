package main

import (
	"testing"
)

func Test_printNoOp(t *testing.T) {
	printNoOp([]string{"temporal", "help"})
}

func Test_printHelp(t *testing.T) {
	printHelp(map[string]cmd{})
}
