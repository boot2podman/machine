package provisiontest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFakeSSHCommander(t *testing.T) {
	sshCmder := NewFakeSSHCommander(FakeSSHCommanderOptions{FilesystemType: "tmpfs"})
	output, err := sshCmder.SSHCommand("stat -f -c %T /var/lib")
	if err != nil || output != "tmpfs\n" {
		t.Fatal("FakeSSHCommander should have returned tmpfs and no error but returned '", output, "' and error", err)
	}
}

func TestStatSSHCommand(t *testing.T) {
	sshCmder := FakeSSHCommander{
		Responses: map[string]string{"sshcommand": "sshcommandresponse"},
	}

	output, err := sshCmder.SSHCommand("sshcommand")
	assert.NoError(t, err)
	assert.Equal(t, "sshcommandresponse", output)

	output, err = sshCmder.SSHCommand("errorcommand")
	assert.Error(t, err)
}
