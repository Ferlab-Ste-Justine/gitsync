package git

import (
	"errors"
	"path/filepath"
	"fmt"
	"io/fs"
	"os"

	"github.com/Ferlab-Ste-Justine/gitsync/config"

	git "github.com/Ferlab-Ste-Justine/git-sdk"
)

func verifyRepoSignatures(repo *git.GitRepository, signaturesPath string) error {
	keys := []string{}
	err := filepath.Walk(signaturesPath, func(fPath string, fInfo fs.FileInfo, fErr error) error {
		if fErr != nil {
			return fErr
		}

		if fInfo.IsDir() {
			return nil
		}

		key, keyErr := os.ReadFile(fPath)
		if keyErr != nil {
			return errors.New(fmt.Sprintf("Error reading accepted signature: %s", keyErr.Error()))
		}

		keys = append(keys, string(key))

		return nil
	})

	if err != nil {
		return err
	}

	return git.VerifyTopCommit(repo, keys)
}

func Clone(conf config.ConfigGit) (*git.MemoryStore, error) {
	sshCreds, sshCredsErr := git.GetSshCredentials(conf.Auth.SshKey, conf.Auth.KnownKey, conf.Auth.User)
	if sshCredsErr != nil {
		return nil, sshCredsErr
	}

	repo, store, repErr := git.MemCloneGitRepo(conf.Repo, conf.Ref, 1, sshCreds)
	if repErr != nil {
		return nil, repErr
	}
	
	if conf.AcceptedSignatures != "" {
		verifyErr := verifyRepoSignatures(repo, conf.AcceptedSignatures)
		if verifyErr != nil {
			return nil, verifyErr
		}
	}

	return store, nil
}
