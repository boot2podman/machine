package check

import (
	"fmt"
	"net/url"

	"github.com/boot2podman/machine/libmachine/auth"
	"github.com/boot2podman/machine/libmachine/cert"
	"github.com/boot2podman/machine/libmachine/host"
)

var (
	DefaultConnChecker ConnChecker
)

func init() {
	DefaultConnChecker = &MachineConnChecker{}
}

// ErrCertInvalid for when the cert is computed to be invalid.
type ErrCertInvalid struct {
	wrappedErr error
	hostURL    string
}

func (e ErrCertInvalid) Error() string {
	return fmt.Sprintf(`There was an error validating certificates for host %q: %s
You can attempt to regenerate them using 'podman-machine regenerate-certs [name]'.
`, e.hostURL, e.wrappedErr)
}

type ConnChecker interface {
	Check(*host.Host) (podmanHost string, authOptions *auth.Options, err error)
}

type MachineConnChecker struct{}

func (mcc *MachineConnChecker) Check(h *host.Host) (string, *auth.Options, error) {
	podmanHost, err := h.Driver.GetURL()
	if err != nil {
		return "", &auth.Options{}, err
	}

	podmanURL := podmanHost

	u, err := url.Parse(podmanURL)
	if err != nil {
		return "", &auth.Options{}, fmt.Errorf("Error parsing URL: %s", err)
	}

	authOptions := h.AuthOptions()

	if err := checkCert(u.Host, authOptions); err != nil {
		return "", &auth.Options{}, fmt.Errorf("Error checking and/or regenerating the certs: %s", err)
	}

	return podmanURL, authOptions, nil
}

func checkCert(hostURL string, authOptions *auth.Options) error {
	valid, err := cert.ValidateCertificate(hostURL, authOptions)
	if !valid || err != nil {
		return ErrCertInvalid{
			wrappedErr: err,
			hostURL:    hostURL,
		}
	}

	return nil
}
