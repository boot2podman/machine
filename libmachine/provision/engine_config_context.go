package provision

import (
	"github.com/boot2podman/machine/libmachine/auth"
	"github.com/boot2podman/machine/libmachine/engine"
)

type EngineConfigContext struct {
	AuthOptions      auth.Options
	EngineOptions    engine.Options
	EngineOptionsDir string
}
