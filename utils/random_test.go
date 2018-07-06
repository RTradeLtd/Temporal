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
	ru.ReSeed()
	random1 := ru.GenerateString(5, letterBytesLower)
	random2 := ru.GenerateString(5, letterBytesLower)
	if random1 == random2 {
		t.Fatal("generated two random strings that were the same")
	}

}
