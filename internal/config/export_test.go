package config

// ExpandPathForTest is a wrapper for testing expandPath.
func ExpandPathForTest(path string) string {
	return expandPath(path)
}
