package configsync

// Config describes the configsync config
type Config struct {
	Files    []string  `json:"files"`
	Commands []Command `json:"commands"`
	Workdir  string    `json:"work_dir"`
	Git      GitConfig `json:"git"`
	Verbose  bool      `json:"verbose"`
}

// Command describes a command to run
type Command struct {
	CommandLine string `json:"command_line"`
	Filepath    string `json:"file_path"`
}

// GitConfig describes git configuration
type GitConfig struct {
	Path          string `json:"path"`
	Author        string `json:"author"`
	RemoteEnabled bool   `json:"remote_enabled"`
	RemoteName    string `json:"remote_name"`
}
