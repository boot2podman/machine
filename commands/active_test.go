package commands

import (
	"testing"

	"github.com/boot2podman/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestCmdActiveNone(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:       "host1",
			ActiveHost: false,
			State:      state.Running,
		},
		{
			Name:       "host2",
			ActiveHost: false,
			State:      state.Running,
		},
		{
			Name:       "host3",
			ActiveHost: false,
			State:      state.Running,
		},
	}
	_, err := activeHost(hostListItems)
	assert.Equal(t, err, errNoActiveHost)
}

func TestCmdActiveHost(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:       "host1",
			ActiveHost: false,
			State:      state.Timeout,
		},
		{
			Name:       "host2",
			ActiveHost: true,
			State:      state.Running,
		},
		{
			Name:       "host3",
			ActiveHost: false,
			State:      state.Running,
		},
	}
	active, err := activeHost(hostListItems)
	assert.Equal(t, err, nil)
	assert.Equal(t, active.Name, "host2")
}

func TestCmdActiveTimeout(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:       "host1",
			ActiveHost: false,
			State:      state.Running,
		},
		{
			Name:       "host2",
			ActiveHost: false,
			State:      state.Running,
		},
		{
			Name:       "host3",
			ActiveHost: false,
			State:      state.Timeout,
		},
	}
	_, err := activeHost(hostListItems)
	assert.Equal(t, err, errActiveTimeout)
}
