package configsync

// CommandType describes a command object
type CommandType struct {
	CommandLine string `toml:"command_line"`
	Filepath    string `toml:"file_path"`
}

// GitOptionsType describes the configuration type for git
type GitOptionsType struct {
	Path          string `toml:"path"`
	Author        string `toml:"author"`
	RemoteEnabled bool   `toml:"remote_enabled"`
	RemoteName    string `toml:"remote_name"`
}
