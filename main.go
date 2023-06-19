package main

import (
	"os"

	"github.com/Ferlab-Ste-Justine/gitsync/cmd"
	"github.com/Ferlab-Ste-Justine/gitsync/config"
	"github.com/Ferlab-Ste-Justine/gitsync/filesystem"
	"github.com/Ferlab-Ste-Justine/gitsync/git"
	"github.com/Ferlab-Ste-Justine/gitsync/grpc"
	"github.com/Ferlab-Ste-Justine/gitsync/logger"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
)

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	log := logger.Logger{LogLevel: logger.ERROR}

	conf, err := config.GetConfig(getEnv("GITSYNC_CONFIG_FILE", "config.yml"))
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}

	log.LogLevel = conf.GetLogLevel()

	fsErr := filesystem.EnsureFilesystemDir(conf.Filesystem.Path, filesystem.ConvertFileMode(conf.Filesystem.DirectoriesPermission))
	if fsErr != nil {
		log.Errorf(fsErr.Error())
		os.Exit(1)
	}

	store, cloneErr := git.Clone(conf.Git)
	if cloneErr != nil {
		log.Errorf(cloneErr.Error())
		os.Exit(1)
	}

	gitKeys, gitKeysErr := store.GetKeyVals(conf.Git.Path)
	if gitKeysErr != nil {
		log.Errorf(gitKeysErr.Error())
		os.Exit(1)
	}

	fsKeys, fsKeysErr := filesystem.GetDirectoryContent(conf.Filesystem.Path)
	if fsKeysErr != nil {
		log.Errorf(fsKeysErr.Error())
		os.Exit(1)
	}

	diff := client.GetKeyDiff(gitKeys, fsKeys)

	if diff.IsEmpty() {
		log.Infof("[main] No update to apply")
		return
	}
	
	log.Infof(
		"[main] Applying update: %d inserts, %d updates and %d deletions",
		len(diff.Inserts),
		len(diff.Updates),
		len(diff.Deletions),
	)

	if len(conf.GrpcNotifications) > 0 {
		err = func() error {
			notifCli, err := grpc.ConnectToNotifEndpoints(conf.GrpcNotifications)
			if err != nil {
				return err
			}

			defer notifCli.Close()
	
			sendErr := notifCli.Send(diff)
			return sendErr
		}()
		if err != nil {
			log.Errorf(err.Error())
			os.Exit(1)
		}
	}

	applyErr := filesystem.ApplyDiffToDirectory(
		conf.Filesystem.Path,
		diff,
		filesystem.ConvertFileMode(conf.Filesystem.FilesPermission),
		filesystem.ConvertFileMode(conf.Filesystem.DirectoriesPermission),
	)
	if applyErr != nil {
		log.Errorf(applyErr.Error())
		os.Exit(1)
	}

	if len(conf.NotificationCommand) > 0 {
		cmdErr := cmd.ExecCommand(conf.NotificationCommand, conf.NotificationCommandRetries, log)
		if cmdErr != nil {
			log.Errorf(cmdErr.Error())
			os.Exit(1)
		}
	}
}