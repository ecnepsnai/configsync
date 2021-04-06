package configsync

// ConfigType describes the configuration type for configsync
type ConfigType struct {
	Files    []string      `json:"files"`
	Commands []CommandType `json:"commands"`
	Workdir  string        `json:"work_dir"`
	Git      GitConfigType `json:"git"`
	Verbose  bool          `json:"verbose"`
}

// CommandType describes a command object
type CommandType struct {
	CommandLine string `json:"command_line"`
	Filepath    string `json:"file_path"`
}

// GitConfigType describes the configuration type for git
type GitConfigType struct {
	Path          string `json:"path"`
	Author        string `json:"author"`
	RemoteEnabled bool   `json:"remote_enabled"`
	RemoteName    string `json:"remote_name"`
}
