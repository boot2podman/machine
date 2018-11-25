package drivers

import (
	"errors"
	"path/filepath"
)

const (
	DefaultSSHUser          = "root"
	DefaultSSHPort          = 22
	DefaultEngineInstallURL = "https://get.docker.com"
)

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	IPAddress     string
	MachineName   string
	SSHUser       string
	SSHPort       int
	SSHKeyPath    string
	SSHKnownHosts string
	StorePath     string
}

// DriverName returns the name of the driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// GetMachineName returns the machine name
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// GetIP returns the ip
func (d *BaseDriver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", errors.New("IP address is not set")
	}
	return d.IPAddress, nil
}

// GetSSHKeyPath returns the ssh key path
func (d *BaseDriver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

// GetSSHKnownHosts returns the ssh key path
func (d *BaseDriver) GetSSHKnownHosts() string {
	if d.SSHKnownHosts == "" {
		d.SSHKnownHosts = d.ResolveStorePath("known_hosts")
	}
	return d.SSHKeyPath
}

// GetSSHPort returns the ssh port, 22 if not specified
func (d *BaseDriver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = DefaultSSHPort
	}

	return d.SSHPort, nil
}

// GetSSHUsername returns the ssh user name, root if not specified
func (d *BaseDriver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = DefaultSSHUser
	}
	return d.SSHUser
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *BaseDriver) PreCreateCheck() error {
	return nil
}

// ResolveStorePath returns the store path where the machine is
func (d *BaseDriver) ResolveStorePath(file string) string {
	return filepath.Join(d.StorePath, "machines", d.MachineName, file)
}

func EngineInstallURLFlagSet(flags DriverOptions) bool {
	return EngineInstallURLSet(flags.String("engine-install-url"))
}

func EngineInstallURLSet(url string) bool {
	return url != DefaultEngineInstallURL && url != ""
}
