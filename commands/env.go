package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"

	"github.com/boot2podman/machine/libmachine"
	"github.com/boot2podman/machine/libmachine/log"
	"github.com/boot2podman/machine/libmachine/shell"
)

const (
	envTmpl    = `{{ .Prefix }}PODMAN_USER{{ .Delimiter }}{{ .PodmanUser }}{{ .Suffix }}{{ .Prefix }}PODMAN_HOST{{ .Delimiter }}{{ .PodmanHost }}{{ .Suffix }}{{ .Prefix }}PODMAN_PORT{{ .Delimiter }}{{ .PodmanPort }}{{ .Suffix }}{{ .Prefix }}PODMAN_IDENTITY_FILE{{ .Delimiter }}{{ .IdentityFile }}{{ .Suffix }}{{ if .KnownHosts }}{{ .Prefix }}PODMAN_KNOWN_HOSTS{{ .Delimiter }}{{ .KnownHosts }}{{ .Suffix }}{{else}}{{ .Prefix }}PODMAN_IGNORE_HOSTS{{ .Delimiter }}true{{ .Suffix }}{{end}}{{ .Prefix }}PODMAN_MACHINE_NAME{{ .Delimiter }}{{ .MachineName }}{{ .Suffix }}{{ if .ComposePathsVar }}{{ .Prefix }}COMPOSE_CONVERT_WINDOWS_PATHS{{ .Delimiter }}true{{ .Suffix }}{{end}}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}`
	bridgeTmpl = `{{ .Prefix }}VARLINK_BRIDGE{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}{{ .Prefix }}PODMAN_VARLINK_BRIDGE{{ .Delimiter }}{{ .VarlinkBridge }}{{ .Suffix }}{{ .Prefix }}PODMAN_MACHINE_NAME{{ .Delimiter }}{{ .MachineName }}{{ .Suffix }}{{ .UsageHint }}`
)

var (
	errImproperUnsetEnvArgs = errors.New("Error: Expected no machine name when the -u flag is present")
	defaultUsageHinter      UsageHintGenerator
	runtimeOS               = func() string { return runtime.GOOS }
)

func init() {
	defaultUsageHinter = &EnvUsageHintGenerator{}
}

type ShellConfig struct {
	Prefix          string
	Delimiter       string
	Suffix          string
	PodmanUser      string
	PodmanHost      string
	PodmanPort      int
	IdentityFile    string
	KnownHosts      string
	VarlinkBridge   string
	UsageHint       string
	MachineName     string
	NoProxyVar      string
	NoProxyValue    string
	ComposePathsVar bool
}

func cmdEnv(c CommandLine, api libmachine.API) error {
	var (
		err      error
		shellCfg *ShellConfig
	)

	// Ensure that log messages always go to stderr when this command is
	// being run (it is intended to be run in a subshell)
	log.SetOutWriter(os.Stderr)

	if c.Bool("unset") {
		shellCfg, err = shellCfgUnset(c, api)
		if err != nil {
			return err
		}
	} else {
		shellCfg, err = shellCfgSet(c, api)
		if err != nil {
			return err
		}
	}

	if c.Bool("varlink") {
		return executeTemplateStdout(shellCfg, bridgeTmpl)
	} else {
		return executeTemplateStdout(shellCfg, envTmpl)
	}
}

func shellCfgSet(c CommandLine, api libmachine.API) (*ShellConfig, error) {
	if len(c.Args()) > 1 {
		return nil, ErrExpectedOneMachine
	}

	target, err := targetHost(c, api)
	if err != nil {
		return nil, err
	}

	host, err := api.Load(target)
	if err != nil {
		return nil, err
	}

	userShell, err := getShell(c.String("shell"))
	if err != nil {
		return nil, err
	}

	var shellCfg *ShellConfig
	hint := defaultUsageHinter.GenerateUsageHint(userShell, os.Args)

	if host.Driver != nil && c.Bool("varlink") == false {

		user := host.Driver.GetSSHUsername()
		if err != nil {
			return nil, err
		}

		addr, err := host.Driver.GetSSHHostname()
		if err != nil {
			return nil, err
		}

		port, err := host.Driver.GetSSHPort()
		if err != nil {
			return nil, err
		}

		key := host.Driver.GetSSHKeyPath()

		if addr != "" {
			// always use root@ for socket
			user = "root"
		}

		shellCfg = &ShellConfig{
			PodmanUser:   user,
			PodmanHost:   addr,
			PodmanPort:   port,
			IdentityFile: key,
			UsageHint:    hint,
			MachineName:  host.Name,
		}

	} else if host.Driver != nil {

		client, err := host.CreateExternalRootSSHClient()
		if err != nil {
			return nil, err
		}

		command := []string{client.BinaryPath}
		command = append(command, client.BaseArgs...)
		command = append(command, "varlink", "bridge")
		bridge := strings.Join(command, " ")

		shellCfg = &ShellConfig{
			VarlinkBridge: bridge,
			UsageHint:     hint,
			MachineName:   host.Name,
		}

	} else {
		shellCfg = &ShellConfig{
			UsageHint:   hint,
			MachineName: host.Name,
		}
	}

	if c.Bool("no-proxy") {
		ip, err := host.Driver.GetIP()
		if err != nil {
			return nil, fmt.Errorf("Error getting host IP: %s", err)
		}

		noProxyVar, noProxyValue := findNoProxyFromEnv()

		// add the host to the no_proxy list idempotently
		switch {
		case noProxyValue == "":
			noProxyValue = ip
		case strings.Contains(noProxyValue, ip):
		//ip already in no_proxy list, nothing to do
		default:
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ip)
		}

		shellCfg.NoProxyVar = noProxyVar
		shellCfg.NoProxyValue = noProxyValue
	}

	if runtimeOS() == "windows" {
		shellCfg.ComposePathsVar = true
	}

	switch userShell {
	case "fish":
		shellCfg.Prefix = "set -gx "
		shellCfg.Suffix = "\";\n"
		shellCfg.Delimiter = " \""
	case "powershell":
		shellCfg.Prefix = "$Env:"
		shellCfg.Suffix = "\"\n"
		shellCfg.Delimiter = " = \""
	case "cmd":
		shellCfg.Prefix = "SET "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = "="
	case "tcsh":
		shellCfg.Prefix = "setenv "
		shellCfg.Suffix = "\";\n"
		shellCfg.Delimiter = " \""
	case "emacs":
		shellCfg.Prefix = "(setenv \""
		shellCfg.Suffix = "\")\n"
		shellCfg.Delimiter = "\" \""
	default:
		shellCfg.Prefix = "export "
		shellCfg.Suffix = "\"\n"
		shellCfg.Delimiter = "=\""
	}

	return shellCfg, nil
}

func shellCfgUnset(c CommandLine, api libmachine.API) (*ShellConfig, error) {
	if len(c.Args()) != 0 {
		return nil, errImproperUnsetEnvArgs
	}

	userShell, err := getShell(c.String("shell"))
	if err != nil {
		return nil, err
	}

	shellCfg := &ShellConfig{
		UsageHint: defaultUsageHinter.GenerateUsageHint(userShell, os.Args),
	}

	if c.Bool("no-proxy") {
		shellCfg.NoProxyVar, shellCfg.NoProxyValue = findNoProxyFromEnv()
	}

	switch userShell {
	case "fish":
		shellCfg.Prefix = "set -e "
		shellCfg.Suffix = ";\n"
		shellCfg.Delimiter = ""
	case "powershell":
		shellCfg.Prefix = `Remove-Item Env:\\`
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = ""
	case "cmd":
		shellCfg.Prefix = "SET "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = "="
	case "emacs":
		shellCfg.Prefix = "(setenv \""
		shellCfg.Suffix = ")\n"
		shellCfg.Delimiter = "\" nil"
	case "tcsh":
		shellCfg.Prefix = "unsetenv "
		shellCfg.Suffix = ";\n"
		shellCfg.Delimiter = ""
	default:
		shellCfg.Prefix = "unset "
		shellCfg.Suffix = "\n"
		shellCfg.Delimiter = ""
	}

	return shellCfg, nil
}

func executeTemplateStdout(shellCfg *ShellConfig, strTmpl string) error {
	t := template.New("envConfig")
	tmpl, err := t.Parse(strTmpl)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, shellCfg)
}

func getShell(userShell string) (string, error) {
	if userShell != "" {
		return userShell, nil
	}
	return shell.Detect()
}

func findNoProxyFromEnv() (string, string) {
	// first check for an existing lower case no_proxy var
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv("no_proxy")

	// otherwise default to allcaps HTTP_PROXY
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}

type UsageHintGenerator interface {
	GenerateUsageHint(string, []string) string
}

type EnvUsageHintGenerator struct{}

func (g *EnvUsageHintGenerator) GenerateUsageHint(userShell string, args []string) string {
	cmd := ""
	comment := "#"

	podmanMachinePath := args[0]
	if strings.Contains(podmanMachinePath, " ") || strings.Contains(podmanMachinePath, `\`) {
		args[0] = fmt.Sprintf("\"%s\"", podmanMachinePath)
	}

	commandLine := strings.Join(args, " ")

	switch userShell {
	case "fish":
		cmd = fmt.Sprintf("eval (%s)", commandLine)
	case "powershell":
		cmd = fmt.Sprintf("& %s | Invoke-Expression", commandLine)
	case "cmd":
		cmd = fmt.Sprintf("\t@FOR /f \"tokens=*\" %%i IN ('%s') DO @%%i", commandLine)
		comment = "REM"
	case "emacs":
		cmd = fmt.Sprintf("(with-temp-buffer (shell-command \"%s\" (current-buffer)) (eval-buffer))", commandLine)
		comment = ";;"
	case "tcsh":
		cmd = fmt.Sprintf("eval `%s`", commandLine)
		comment = ":"
	default:
		cmd = fmt.Sprintf("eval $(%s)", commandLine)
	}

	return fmt.Sprintf("%s Run this command to configure your shell: \n%s %s\n", comment, comment, cmd)
}
