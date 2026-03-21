package install

import (
	"github.com/knadh/koanf/v2"
)

// FileProvider is a wrapper for fileProvider.
func FileProvider(path string) koanf.Provider {
	return fileProvider(path)
}

// InjectConfig is a wrapper for injectConfig.
func InjectConfig(info AgentConfigInfo, def MCPServerConfig) error {
	return injectConfig(info, def)
}

// ExpandPath is a wrapper for expandPath.
func ExpandPath(path string) string {
	return expandPath(path)
}

// InjectJSON is a wrapper for injectJSON.
func InjectJSON(path string, def MCPServerConfig) error {
	return injectJSON(path, def)
}

// InjectTOML is a wrapper for injectTOML.
func InjectTOML(path string, def MCPServerConfig) error {
	return injectTOML(path, def)
}

// ProjectFiles is a wrapper for projectFiles.
func ProjectFiles() []ProjectFile {
	return projectFiles()
}
