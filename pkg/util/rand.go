package util

import (
	"math/rand"
)

func RandomString(size int) string {
	buf := make([]byte, size)
	for idx := range buf {
		buf[idx] = chars[rand.Intn(len(chars))]
	}
	return string(buf)
}

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
