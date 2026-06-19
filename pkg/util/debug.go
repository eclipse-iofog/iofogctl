package util

var debug bool = false

func SetDebug(value bool) {
	debug = value
}

func IsDebug() bool {
	return debug
}
