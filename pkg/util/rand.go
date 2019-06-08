package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func RandomString(size int, chars string) string {
	buf := make([]byte, size)
	for idx := range buf {
		buf[idx] = chars[rand.Intn(len(chars))]
	}
	return string(buf)
}

const AlphaNum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const Alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const AlphaLower = "abcdefghijklmnopqrstuvwxyz"
