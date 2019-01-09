package commands

import (
	"fmt"
	"os"

	"github.com/boot2podman/machine/libmachine"
	"github.com/boot2podman/machine/libmachine/log"
)

func cmdConfig(c CommandLine, api libmachine.API) error {
	// Ensure that log messages always go to stderr when this command is
	// being run (it is intended to be run in a subshell)
	log.SetOutWriter(os.Stderr)

	target, err := targetHost(c, api)
	if err != nil {
		return err
	}

	host, err := api.Load(target)
	if err != nil {
		return err
	}

	if host.Driver == nil {
		return err
	}

	user := host.Driver.GetSSHUsername()

	addr, err := host.Driver.GetSSHHostname()
	if err != nil {
		return err
	}

	port, err := host.Driver.GetSSHPort()
	if err != nil {
		return err
	}

	key := host.Driver.GetSSHKeyPath()

	if addr != "" {
		// always use root@ for socket
		user = "root"
	}

	fmt.Printf("--username=%s\n--host=%s\n--port=%d\n--identity-file=%s\n",
		user, addr, port, key)

	return nil
}
