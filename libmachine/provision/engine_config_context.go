package provision

import (
	"github.com/boot2podman/machine/libmachine/auth"
	"github.com/boot2podman/machine/libmachine/engine"
)

type EngineConfigContext struct {
	DockerPort       int
	AuthOptions      auth.Options
	EngineOptions    engine.Options
	DockerOptionsDir string
}
