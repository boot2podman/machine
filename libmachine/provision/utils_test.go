package provision

import (
	"fmt"
	"strings"
	"testing"

	"github.com/boot2podman/machine/drivers/fakedriver"
	"github.com/boot2podman/machine/libmachine/auth"
	"github.com/boot2podman/machine/libmachine/engine"
	"github.com/boot2podman/machine/libmachine/provision/pkgaction"
	"github.com/boot2podman/machine/libmachine/provision/provisiontest"
	"github.com/boot2podman/machine/libmachine/provision/serviceaction"
	"github.com/stretchr/testify/assert"
)

func TestGenerateDockerOptionsBoot2Podman(t *testing.T) {
	p := &Boot2PodmanProvisioner{
		Driver: &fakedriver.Driver{},
	}
	dockerPort := 1234
	p.AuthOptions = auth.Options{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}
	engineConfigPath := "/var/lib/boot2podman/profile"

	dockerCfg, err := p.GenerateDockerOptions(dockerPort)
	if err != nil {
		t.Fatal(err)
	}

	if dockerCfg.EngineOptionsPath != engineConfigPath {
		t.Fatalf("expected engine path %s; received %s", engineConfigPath, dockerCfg.EngineOptionsPath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("CACERT=%s", p.AuthOptions.CaCertRemotePath)) == -1 {
		t.Fatalf("CACERT option invalid; expected %s", p.AuthOptions.CaCertRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("SERVERKEY=%s", p.AuthOptions.ServerKeyRemotePath)) == -1 {
		t.Fatalf("SERVERKEY option invalid; expected %s", p.AuthOptions.ServerKeyRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("SERVERCERT=%s", p.AuthOptions.ServerCertRemotePath)) == -1 {
		t.Fatalf("SERVERCERT option invalid; expected %s", p.AuthOptions.ServerCertRemotePath)
	}
}

type fakeProvisioner struct {
	GenericProvisioner
}

func (provisioner *fakeProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

func (provisioner *fakeProvisioner) Provision(authOptions auth.Options, engineOptions engine.Options) error {
	return nil
}

func (provisioner *fakeProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	return nil
}

func (provisioner *fakeProvisioner) String() string {
	return "fake"
}

func TestDecideStorageDriver(t *testing.T) {
	var tests = []struct {
		suppliedDriver       string
		defaultDriver        string
		remoteFilesystemType string
		expectedDriver       string
	}{
		{"", "aufs", "ext4", "aufs"},
		{"", "aufs", "btrfs", "btrfs"},
		{"", "overlay", "btrfs", "overlay"},
		{"devicemapper", "aufs", "ext4", "devicemapper"},
		{"devicemapper", "aufs", "btrfs", "devicemapper"},
	}

	p := &fakeProvisioner{GenericProvisioner{
		Driver: &fakedriver.Driver{},
	}}
	for _, test := range tests {
		p.SSHCommander = provisiontest.NewFakeSSHCommander(
			provisiontest.FakeSSHCommanderOptions{
				FilesystemType: test.remoteFilesystemType,
			},
		)
		storageDriver, err := decideStorageDriver(p, test.defaultDriver, test.suppliedDriver)
		assert.NoError(t, err)
		assert.Equal(t, test.expectedDriver, storageDriver)
	}
}

func TestGetFilesystemType(t *testing.T) {
	p := &fakeProvisioner{GenericProvisioner{
		Driver: &fakedriver.Driver{},
	}}
	p.SSHCommander = &provisiontest.FakeSSHCommander{
		Responses: map[string]string{
			"stat -f -c %T /var/lib": "btrfs\n",
		},
	}
	fsType, err := getFilesystemType(p, "/var/lib")
	assert.NoError(t, err)
	assert.Equal(t, "btrfs", fsType)
}

func TestDockerClientVersion(t *testing.T) {
	cases := []struct {
		output, want string
	}{
		{"Docker version 1.9.1, build a34a1d5\n", "1.9.1"},
		{"Docker version 1.9.1\n", "1.9.1"},
		{"Docker version 1.13.0-rc1, build deadbeef\n", "1.13.0-rc1"},
		{"Docker version 1.13.0-dev, build deadbeef\n", "1.13.0-dev"},
	}

	sshCmder := &provisiontest.FakeSSHCommander{
		Responses: make(map[string]string),
	}

	for _, tc := range cases {
		sshCmder.Responses["docker --version"] = tc.output
		got, err := DockerClientVersion(sshCmder)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("Unexpected version string from %q; got %q, want %q", tc.output, tc.want, got)
		}
	}
}
