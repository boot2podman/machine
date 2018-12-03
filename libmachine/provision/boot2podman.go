package provision

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"net"
	"path"
	"text/template"
	//"time"

	"github.com/boot2podman/machine/commands/mcndirs"
	"github.com/boot2podman/machine/libmachine/auth"
	"github.com/boot2podman/machine/libmachine/drivers"
	"github.com/boot2podman/machine/libmachine/engine"
	"github.com/boot2podman/machine/libmachine/log"
	"github.com/boot2podman/machine/libmachine/mcnutils"
	"github.com/boot2podman/machine/libmachine/provision/pkgaction"
	"github.com/boot2podman/machine/libmachine/provision/serviceaction"
	"github.com/boot2podman/machine/libmachine/state"
)

func init() {
	Register("boot2podman", &RegisteredProvisioner{
		New: NewBoot2PodmanProvisioner,
	})
}

func NewBoot2PodmanProvisioner(d drivers.Driver) Provisioner {
	return &Boot2PodmanProvisioner{
		Driver: d,
	}
}

type Boot2PodmanProvisioner struct {
	OsReleaseInfo *OsRelease
	Driver        drivers.Driver
	AuthOptions   auth.Options
	EngineOptions engine.Options
}

func (provisioner *Boot2PodmanProvisioner) String() string {
	return "boot2podman"
}

func (provisioner *Boot2PodmanProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	_, err := provisioner.SSHCommand(fmt.Sprintf("sudo /etc/init.d/%s %s", name, action.String()))
	return err
}

func (provisioner *Boot2PodmanProvisioner) upgradeIso() error {
	// TODO: Ideally, we should not read from mcndirs directory at all.
	// The driver should be able to communicate how and where to place the
	// relevant files.
	b2putils := mcnutils.NewB2pUtils(mcndirs.GetBaseDir())

	// Check if the driver has specified a custom b2p url
	jsonDriver, err := json.Marshal(provisioner.GetDriver())
	if err != nil {
		return err
	}
	var d struct {
		Boot2PodmanURL string
	}
	json.Unmarshal(jsonDriver, &d)

	log.Info("Stopping machine to do the upgrade...")

	if err := provisioner.Driver.Stop(); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Stopped)); err != nil {
		return err
	}

	machineName := provisioner.GetDriver().GetMachineName()

	log.Infof("Upgrading machine %q...", machineName)

	// Either download the latest version of the b2p url that was explicitly
	// specified when creating the VM or copy the (updated) default ISO
	if err := b2putils.CopyIsoToMachineDir(d.Boot2PodmanURL, machineName); err != nil {
		return err
	}

	log.Infof("Starting machine back up...")

	if err := provisioner.Driver.Start(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Running))
}

func (provisioner *Boot2PodmanProvisioner) Package(name string, action pkgaction.PackageAction) error {
	if name == "podman" && action == pkgaction.Upgrade {
		if err := provisioner.upgradeIso(); err != nil {
			return err
		}
	}
	return nil
}

func (provisioner *Boot2PodmanProvisioner) Hostname() (string, error) {
	return provisioner.SSHCommand("hostname")
}

func (provisioner *Boot2PodmanProvisioner) SetHostname(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && sudo mkdir -p /var/lib/boot2podman/etc && echo %q | sudo tee /var/lib/boot2podman/etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2PodmanProvisioner) GetEngineOptionsDir() string {
	return "/var/lib/boot2podman"
}

func (provisioner *Boot2PodmanProvisioner) GetAuthOptions() auth.Options {
	return provisioner.AuthOptions
}

func (provisioner *Boot2PodmanProvisioner) GenerateEngineOptions() (*EngineOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `
EXTRA_ARGS='
{{ range .EngineOptions.Labels }}--label {{.}}
{{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}}
{{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}}
{{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}}
{{ end }}
'
CACERT={{.AuthOptions.CaCertRemotePath}}
SERVERKEY={{.AuthOptions.ServerKeyRemotePath}}
SERVERCERT={{.AuthOptions.ServerCertRemotePath}}

{{range .EngineOptions.Env}}export \"{{ printf "%q" . }}\"
{{end}}
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		AuthOptions:   provisioner.AuthOptions,
		EngineOptions: provisioner.EngineOptions,
	}

	t.Execute(&engineCfg, engineConfigContext)

	daemonOptsDir := path.Join(provisioner.GetEngineOptionsDir(), "profile")
	return &EngineOptions{
		EngineOptionsString: engineCfg.String(),
		EngineOptionsPath:   daemonOptsDir,
	}, nil
}

func (provisioner *Boot2PodmanProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.ID == "boot2podman" || provisioner.OsReleaseInfo.ID == "tinycore"
}

func (provisioner *Boot2PodmanProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *Boot2PodmanProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return provisioner.OsReleaseInfo, nil
}

func (provisioner *Boot2PodmanProvisioner) Provision(authOptions auth.Options, engineOptions engine.Options) error {
	var (
		err error
	)

	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	if err = provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err = makeEngineOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err = ConfigureAuth(provisioner); err != nil {
		return err
	}

	return err
}

func (provisioner *Boot2PodmanProvisioner) SSHCommand(args string) (string, error) {
	return drivers.RunSSHCommandFromDriver(provisioner.Driver, args)
}

func (provisioner *Boot2PodmanProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
