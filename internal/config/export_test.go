package config

// ExpandPathForTest is a wrapper for testing expandPath.
func ExpandPathForTest(path string) string {
	return expandPath(path)
}

// SetLoadedPathForTest allows setting the internal loadedPath for testing Save.
func SetLoadedPathForTest(path string) {
	loadedPath = path
}
