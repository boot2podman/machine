package commands

import (
	"testing"

	"github.com/boot2podman/machine/commands/commandstest"
	"github.com/boot2podman/machine/libmachine"
	"github.com/boot2podman/machine/libmachine/libmachinetest"
	"github.com/stretchr/testify/assert"
)

func TestCmdInspect(t *testing.T) {
	testCases := []struct {
		commandLine CommandLine
		api         libmachine.API
		expectedErr error
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"foo", "bar"},
			},
			api:         &libmachinetest.FakeAPI{},
			expectedErr: ErrExpectedOneMachine,
		},
	}

	for _, tc := range testCases {
		err := cmdInspect(tc.commandLine, tc.api)
		assert.Equal(t, tc.expectedErr, err)
	}
}
