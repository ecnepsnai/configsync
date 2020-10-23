# ConfigSync

ConfigSync is a command line tool to backup configuration files and command-line outputs into a git repo.

# How it Works

ConfigSync takes ownership of the `work_dir` you specified by initalizing a git repository there, and checking out a branch with the hostname of the system.

For each file you specify, ConfigSync will make a 1-to-1 clone of that file in the work directory. If the file already exists, it will only update it if the files differ by hash. If a glob pattern is specified, it expands and syncs each file that matched the pattern.

For each command that is specified the `command_line` is executed (within a shell) and the resulting combined output (both stdout and stderr) is written to `file_path`. If the output of the command matches an existing file in `file_path`, then the file is not updated. The command is executed every time ConfigSync runs.

Once all files and commands have been synced it will check to see if there have been any changes to the git directory, and if so it will commit the changes. If git remote is enabled, the changes are pushed to the remote.

# How to Use It

ConfigSync is not a daemon or any kind of background application, it is designed to be run on a set schedule - such as with a crontab.

For example, you may wish to use:

```
0 */4 * * * /root/configsync /root/configsync.json
```

Which will run ConfigSync every 4 hours.

## Requirements

- A Linux, BSD, or Darwin host
- Git 1.8 or newer
- (Optional) Passwordless authentication for a Git Remote (SSH keys or cached HTTP credentials)

## Configuration

Configuration is defined in a JSON file, such as:

```json
{
    // must specify a workdir
    "work_dir": "/root/configsync",

    // at least file or command is required
    "files": [
        "/etc/passwd", // Single File
        "/var/spool/cron/*" // Glob to expand to multiple files
    ],
    "commands": [
        {
            // the output of this command will be saved to `/etc/zpool.yml` in the workdir
            "command_line": "zdb -C",
            "file_path": "/etc/zpool.yml"
        }
    ],

    "git": {
        // Optional path to the git binary (will use $PATH otherwise)
        "path": "/bin/git",
        // Optional
        "author": "configsync <configsync@localhost>",
        // If true, `remote_name` is required
        "remote_enabled": true,
        "remote_name": "origin"
    }
}
```

## Work Directory Setup

If you are not using a remote, you don't need to prepare the work directory.

If you are using a remote, you should clone the repository first. A remote with the name specified in `remote_name` must already exist for ConfigSync to work - it will not create it for you.