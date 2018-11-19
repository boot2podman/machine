package commands

import (
	"github.com/boot2podman/machine/libmachine"
	"github.com/boot2podman/machine/libmachine/log"
)

func cmdRestart(c CommandLine, api libmachine.API) error {
	if err := runAction("restart", c, api); err != nil {
		return err
	}

	log.Info("Restarted machines may have new IP addresses. You may need to re-run the `podman-machine env` command.")

	return nil
}
