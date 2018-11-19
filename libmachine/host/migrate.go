package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/boot2podman/machine/drivers/none"
	"github.com/boot2podman/machine/libmachine/log"
	"github.com/boot2podman/machine/libmachine/version"
)

var (
	errConfigFromFuture = errors.New("config version is from the future -- you should upgrade your Podman Machine client")
)

type RawDataDriver struct {
	*none.Driver
	Data []byte // passed directly back when invoking json.Marshal on this type
}

func (r *RawDataDriver) MarshalJSON() ([]byte, error) {
	return r.Data, nil
}

func (r *RawDataDriver) UnmarshalJSON(data []byte) error {
	r.Data = data
	return nil
}

func getMigratedHostMetadata(data []byte) (*Metadata, error) {
	var (
		hostMetadata *Metadata
	)

	if err := json.Unmarshal(data, &hostMetadata); err != nil {
		return &Metadata{}, err
	}

	return hostMetadata, nil
}

func MigrateHost(h *Host, data []byte) (*Host, bool, error) {
	var (
		migrationNeeded    = false
		migrationPerformed = false
	)

	migratedHostMetadata, err := getMigratedHostMetadata(data)
	if err != nil {
		return nil, false, err
	}

	globalStorePath := filepath.Dir(filepath.Dir(migratedHostMetadata.HostOptions.AuthOptions.StorePath))

	driver := &RawDataDriver{none.NewDriver(h.Name, globalStorePath), nil}

	if migratedHostMetadata.ConfigVersion > version.ConfigVersion {
		return nil, false, errConfigFromFuture
	}

	if migratedHostMetadata.ConfigVersion == version.ConfigVersion {
		h.Driver = driver
		if err := json.Unmarshal(data, &h); err != nil {
			return nil, migrationPerformed, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
		}
	} else {
		migrationNeeded = true
	}

	if migrationNeeded {
		migrationPerformed = true
		for h.ConfigVersion = migratedHostMetadata.ConfigVersion; h.ConfigVersion < version.ConfigVersion; h.ConfigVersion++ {
			log.Debugf("Migrating to config v%d", h.ConfigVersion)
			switch h.ConfigVersion {
			case 3:
			}
		}
	}

	h.RawDriver = driver.Data

	return h, migrationPerformed, nil
}
