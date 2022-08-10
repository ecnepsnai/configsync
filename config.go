package configsync

// CommandType describes a command object
type CommandType struct {
	FilePath  string   `toml:"file_path"`
	ExePath   string   `toml:"exe_path"`
	Arguments []string `toml:"arguments"`
	WorkDir   string   `toml:"work_dir"`
	Env       []string `toml:"env"`
	User      uint32   `toml:"uid"`
	Group     uint32   `toml:"gid"`
}

// GitOptionsType describes the configuration type for git
type GitOptionsType struct {
	Path          string `toml:"path"`
	Author        string `toml:"author"`
	RemoteEnabled bool   `toml:"remote_enabled"`
	RemoteName    string `toml:"remote_name"`
}
