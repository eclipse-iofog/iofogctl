package nodeversion

// SemverPtr returns a pointer to semver when non-empty, otherwise nil.
func SemverPtr(semver string) *string {
	if semver == "" {
		return nil
	}
	return &semver
}
