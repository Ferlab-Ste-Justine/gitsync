package cmd

import (
	"os/exec"
	
	"github.com/Ferlab-Ste-Justine/gitsync/logger"
)

func ExecCommand(command []string, retries uint64, log logger.Logger) error {
	out, err := exec.Command(command[0], command[1:]...).Output()

	if err != nil {
		if retries > 0 {
			log.Warnf("[cmd] Following error occured on post-updated command, will retry: %s", err.Error())
			return ExecCommand(command, retries-1, log)
		}

		return err
	}

	if len(out) > 0 {
		log.Infof("[cmd] Command output: %s", string(out))
	}

	return nil
}
