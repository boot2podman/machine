package commands

import (
	"testing"

	"github.com/boot2podman/machine/commands/commandstest"
	"github.com/boot2podman/machine/libmachine/libmachinetest"
	"github.com/stretchr/testify/assert"
)

func TestCmdVersion(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &libmachinetest.FakeAPI{}

	err := cmdVersion(commandLine, api)

	assert.True(t, commandLine.VersionShown)
	assert.NoError(t, err)
}

func TestCmdVersionTooManyNames(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machine1", "machine2"},
	}
	api := &libmachinetest.FakeAPI{}

	err := cmdVersion(commandLine, api)

	assert.EqualError(t, err, "Error: Expected one machine name as an argument")
}

func TestCmdVersionNotFound(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"unknown"},
	}
	api := &libmachinetest.FakeAPI{}

	err := cmdVersion(commandLine, api)

	assert.EqualError(t, err, `Podman machine "unknown" does not exist. Use "podman-machine ls" to list machines. Use "podman-machine create" to add a new one.`)
}
