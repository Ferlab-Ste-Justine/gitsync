# About

This tool will keep the content of a filesystem directory synchronized with the content of a path in a git repository (can be the entiry repo by specifying the empty path).

It is the pure git complement of the following project: https://github.com/Ferlab-Ste-Justine/configurations-auto-updater

Unless the above project, this tool doesn't not stick around in the background to listen to further changes to synchronize (something that etcd supports naturally well, but git servers, not so much).

Rather, this tools runs once and then exists. It dependents on external solutions like systemd timers or kubernetes cron jobs to provide the recurrence.

# Notifications Support

The tool supports notifications whenever a change is applied.

It supports the following cases:
- Running a command with arguments AFTER files are updated with a change (and retry a certain number of time on error if the command returns a non-zero code)
- Push a notification to remote grpc server(s) with the following api contract: https://github.com/Ferlab-Ste-Justine/etcd-sdk/blob/main/keypb/api.proto#L42 . The push occurs BEFORE the files are updated and the files are only updated if the push succeeds. Note that because pushes to later servers (if you push to several servers) or even file update may fail, the same notification may be pushed more than once (and the servers should react to it in an idempotent way). However, assuming that this tool is managed with a recurrence solution that ensures retries, then the servers are guaranteed to eventually receive all file updates.

# Usage

The behavior of the binary is configured with a configuration file (it tries to look for a **config.yml** file in its running directory, but alternatively, you can specify another path for the configuration file with the **GITSYNC_CONFIG_FILE** environment variable).

The **config.yml** file is as follows:

```
filesystem:
  path: "Path on the filesystem that should be synchronized"
  files_permission: "Permission that should be given to generated files in Unix base 8 format"
  directories_permission: "Permission that should be given to generated directories in Unix base 8 format"
git:
  repo: Url of the repo to sync with
  ref: Git reference of the repo to sync with (usually a branch)
  path: Path in the repo to sync with. Can be empty to sync with the entire repo
  auth:
    ssh_key: Path to ssh to use to authentify against the git server
    known_key: Path of known host key file to use to authentify the git server
  accepted_signatures: Path to a directory containing gpg public key files that should be used to authentify the signature of the top git commit to ensure it was commited by a trusted source.
notification_command:
  - "Notification command and its arguments to run whenever there is an update"
notification_command_retries: "Maximum number of time to retry the notification command if it returns a non-zero code"
grpc_notifications:
  - enpoint: "Endpoint to push notifications on a server to in the following format:  <url>:<port>"
    filter: "An optional regexp filter to apply on all file names being pushed. The remote server will be notified only of changes on files that pass the regexp"
    trim_key_path: "If set to true, the path of file names will be trimed out and the remote server will only receives the base of the file names in its notifications"
    max_chunk_size: "Maximum size to send per message in bytes. If the combined size of the updated files is larger, it will be broken down in several messages. Note that this is a best effort 'guarantee' as the message size may still be larger if a single file exceeds this value"
    auth:
      ca_cert: "Path to CA certificate that will validate the server's certificate for mTLS"
      client_cert: "Path to client public certificate that will authentication to the server for mTLS"
      client_key: "Path to client private key that will authentication to the server for mTLS"
  ..
log_level: "Minimum criticality of logs level displayed. Can be: debug, info, warn, error. Defaults to info"
```