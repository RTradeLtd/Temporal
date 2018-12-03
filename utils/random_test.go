package utils_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/utils"
)

const (
	letterBytesLower      = "abcdefghijklmnopqrstuvwxyz"
	letterBytesLowerUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterBytesMixed      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPRQRSTUVWXYZ"
)

func TestRandomUtils(t *testing.T) {
	ru := utils.GenerateRandomUtils()
	oldSeed := ru.Seed
	ru.ReSeed()
	newSeed := ru.Seed
	if oldSeed == newSeed {
		t.Fatal("failed to generate new seed")
	}
	random1 := ru.GenerateString(5, letterBytesLower)
	if len(random1) != 5 {
		t.Fatal("failed to contruct random string of valid length")
	}
	random2 := ru.GenerateString(5, letterBytesLowerUpper)
	if len(random2) != 5 {
		t.Fatal("failed to construct random string of valid length")
	}
	random3 := ru.GenerateString(5, letterBytesMixed)
	if len(random3) != 5 {
		t.Fatal("failed to construct random string of valid length")
	}
	if random1 == random2 {
		t.Fatal("generated two random strings that were the same")
	}

	if random2 == random3 {
		t.Fatal("generated two random strings that were the same")
	}
}
