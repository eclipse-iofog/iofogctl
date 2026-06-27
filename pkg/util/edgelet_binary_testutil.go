package util

// SetEdgeletReleaseBaseForTest overrides the ldflag release base in unit tests.
func SetEdgeletReleaseBaseForTest(base string) {
	edgeletReleaseBase = base
}

// SetEdgeletBinaryVersionForTest overrides the ldflag binary version in unit tests.
func SetEdgeletBinaryVersionForTest(version string) {
	edgeletBinaryVersion = version
}

// ResetEdgeletReleaseBaseForTest restores the ldflag release base after tests.
func ResetEdgeletReleaseBaseForTest() {
	edgeletReleaseBase = "undefined"
}

// ResetEdgeletBinaryVersionForTest restores the ldflag binary version after tests.
func ResetEdgeletBinaryVersionForTest() {
	edgeletBinaryVersion = "undefined"
}
