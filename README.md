# ConfigSync

ConfigSync is a command line tool to backup configuration files and command-line outputs into a git repo.

# How it Works

ConfigSync takes ownership of the `work_dir` you specified by initializing a git repository there, and checking out a
branch with the hostname of the system.

For each file you specify, ConfigSync will make a 1-to-1 clone of that file in the work directory. If the file already
exists, it will only update it if the files differ by hash. If a glob pattern is specified, it expands and syncs each
file that matched the pattern. Files in the work directory that don't match any file path or glob in the configuration
are removed.

For each command that is specified the `command_line` is executed (within a shell) and the resulting combined output
(both stdout and stderr) is written to `file_path`. If the output of the command matches an existing file in
`file_path`, then the file is not updated. The command is executed every time ConfigSync runs, so it's important that
this command produces consistent output.

Once all files and commands have been synced it will check to see if there have been any changes to the git directory,
and if so it will commit the changes. If git remote is enabled, the changes are pushed to the remote.

# How to Use It

ConfigSync is not a daemon or any kind of background application, it is designed to be run on a set schedule -
such as with a crontab.

For example, you may wish to use:

```
0 */4 * * * /sbin/configsync /etc/configsync/configsync.conf
```

Which will run ConfigSync every 4 hours.

## Requirements

- A Linux, BSD, or Darwin host
- Git 1.8 or newer
- (Optional) Passwordless authentication for a Git Remote (SSH keys or cached HTTP credentials)

## Configuration

ConfigSync uses a config file to define global options, with files and commands being defined within a configuration
directory. This allows you to define a default set of "base" files and additional files that may be specific to the
function of the host.

For example, you may have a base set of files for a Linux install, and an additional set of files for a
PostgreSQL server.

### ConfigSync Options

By default, the configsync binary looks for a file named `configsync.conf` in the current directory. You can specify the
path to the config file by providing it as the one and only argument.

**Example Config:**

```toml
# Required - The directory where file lists and commands are specified. This path is relative to the primary
# configuration file itself.
conf_include = "./conf.d"
# Required - The git working directory where synced files are saved.
workdir = "/root/configuration_files"

[git]
# Optional - Path to the git binary. If omitted will look in $PATH.
path = "/usr/bin/git"
# Optional - If true then will pull/push from the remote. The remote must already be configured in the git working
# directory.
remote_enabled = true
# Optional (required if remote_enabled is true) - The remote name to use for pulling/pushing. Most of the time this is
# 'origin'.
remote_name = "origin"
# Optional - The name of the branch to use for git operations. If omitted the hostname of the system is used.
branch_name = "localhost.localdomain"
```

### File Lists

ConfigSync builds a list of glob patterns based on file lists inside the directory specified in `conf_include`.

File lists must have the extension of ".files" and must only contain one glob pattern per line, separated by
UNIX line terminator. Lines beginning with `#` are ignored, and can be used for comments.

**Example Config:**

```
# Users and Groups
/etc/passwd
/etc/group
/var/spool/cron/*

# Network Interfaces
/etc/sysconfig/network-scripts/ifcfg-*
```

### Commands

Commands are specified by a single configuration file that defines the command to run, and the virtual file path
to save that output to. Command configuration files must inside the directory specified in `conf_include` and have the
extension of ".cmd".

**Example Config:**

```toml
# Required - The pseudo file path where the output of the command will be savedto in the working directory of
# configsync (i.e. the git workdir).
file_path = "/etc/zpool.yml"
# Required - The path to the executable to run. Do not include arguments here!
exe_path = "/usr/sbin/zdb"
# Optional - Array of arguments to pass to the executable
arguments = [ "-C" ]
# Optional - The working directory to start the executable in.
work_dir = "/etc/"
# Optional - Array of strings of key = value pairs for environment variables. By default the executable will inherit
# all variables of configsync itself.
env = [ "key=value" ]
# Optional - User ID number to run the executable as. Will also set ownership of the outputted file.
uid = 1000
# Optional - Group ID number to run the executable as. Will also set ownership of the outputted file.
gid = 1000
```

## Work Directory Setup

If you are not using a remote (`remote_enabled` is set to `false`), then you do not need to prepare the work directory.
ConfigSync will create the directory if needed, and initialize a git project if it has not already.

If you are using a remote, you need to clone the repository first. A remote with the name specified in `remote_name`
must already exist for ConfigSync to work - it will not create it for you.

In either case, ConfigSync will always work in a branch named of the hostname of the system, or `branch_name` if defined
in the configuration file.
