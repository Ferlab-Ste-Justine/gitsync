package main

import (
	"os"

	"github.com/Ferlab-Ste-Justine/gitsync/config"
	"github.com/Ferlab-Ste-Justine/gitsync/filesystem"
	"github.com/Ferlab-Ste-Justine/gitsync/git"
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

	gitKeys, gitKeysErr := store.GetKeyVals()
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
		log.Infof("No update to apply")
	} else {
		log.Infof(
			"Applying update: %d inserts, %d updates and %d deletions",
			len(diff.Inserts),
			len(diff.Updates),
			len(diff.Deletions),
		)	
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
}