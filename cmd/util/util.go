package util

// Check export
func Check(err error) {
	if err != nil {
		panic(err)
	}
}