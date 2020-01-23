# ConfigSync

ConfigSync is a command line tool to sync system configuration files and command outputs in a git repo

## Usage

Git 1.8 or newer must be installed on the client

Save a configuration file such as:

```json
{
    "files": [
        "/etc/passwd",
        "/var/spool/cron/*"
    ],
    "commands": [
        {
            "command_line": "zdb -C",
            "file_path": "/etc/zpool.yml"
        }
    ],
    "work_dir": "/root/configsync",
    "git": {
        "path": "/bin/git",
        "author": "configsync <configsync@localhost>",
        "remote_enabled": true,
        "remote_name": "origin"
    }
}
```

- You can specify single files or glob patterns under `files`
- You can save the output of scripts into files (within the workdir) under `commands`. Commands are executed with a bash shell
- If `remote_enabled` is true, then changes are pushed to the `remote_name`, which must be configured before you run configsync