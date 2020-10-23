package main

type configType struct {
	Files    []string      `json:"files"`
	Commands []commandType `json:"commands"`
	Workdir  string        `json:"work_dir"`
	Git      gitConfigType `json:"git"`
	Verbose  bool          `json:"verbose"`
}

type commandType struct {
	CommandLine string `json:"command_line"`
	Filepath    string `json:"file_path"`
}

type gitConfigType struct {
	Path          string `json:"path"`
	Author        string `json:"author"`
	RemoteEnabled bool   `json:"remote_enabled"`
	RemoteName    string `json:"remote_name"`
}
