# ConfigSync

ConfigSync is a command line tool to backup configuration files and command-line outputs into a git repo.

# How it Works

ConfigSync takes ownership of the `work_dir` you specified by initializing a git repository there, and checking out a branch with the hostname of the system.

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
    "work_dir": "/root/configsync",
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
    "git": {
        "path": "/bin/git",
        "author": "configsync <configsync@localhost>",
        "remote_enabled": true,
        "remote_name": "origin"
    }
}
```

|Object|Type|Description|
|-|-|-|
|`files`|[]string|A list of file paths or glob patterns to include|
|`commands`|[]object|A list of command objects (defined below) of which their value will be includes|
|`work_dir`|string|The absolute path to the directory where the git repository will exist|
|`git`|object|Properties to control git configuration. Defined below.|
|`verbose`|boolean|Optional, set to true to enable verbose logging.|

### Commands

|Object|Type|Description|
|-|-|-|
|`command_line`|string|The command line to execute. Shell options are available.|
|`file_path`|string|The pseudo file path where the output of `command_line` will be saved in the work dir.|

### Git

|Object|Type|Description|
|-|-|-|
|`path`|string|Optional, specify the absolute path to the git binary. Defaults to $PATH.|
|`author`|string|Optional, specify the git author in `$NAME <$EMAIL>` format.|
|`remote_enabled`|boolean|Should changes to the work dir be synced to a git remote.|
|`remote_name`|string|Optional, if remote is enabled optionally specify the remote name. Required if `remote_enabled` is true.|

## Work Directory Setup

If you are not using a remote, you don't need to prepare the work directory.

If you are using a remote, you should clone the repository first. A remote with the name specified in `remote_name` must already exist for ConfigSync to work - it will not create it for you.
