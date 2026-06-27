package install

import (
	"errors"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

// GlobalCertificates holds optional router/NATS CA blocks deployed once via Controller API.
type GlobalCertificates struct {
	RouterSiteCA  *SiteCertificate
	RouterLocalCA *SiteCertificate
	NatsSiteCA    *SiteCertificate
	NatsLocalCA   *SiteCertificate
}

func (g GlobalCertificates) empty() bool {
	return g.RouterSiteCA == nil && g.RouterLocalCA == nil && g.NatsSiteCA == nil && g.NatsLocalCA == nil
}

// DeployGlobalCertificates uploads router and NATS site/local CA secrets once (authenticated client).
func DeployGlobalCertificates(clt *client.Client, certs GlobalCertificates) error {
	if clt == nil || certs.empty() {
		return nil
	}
	pairs := []struct {
		name string
		cert *SiteCertificate
	}{
		{"router-site-ca", certs.RouterSiteCA},
		{"default-router-local-ca", certs.RouterLocalCA},
		{"nats-site-ca", certs.NatsSiteCA},
		{"default-nats-local-ca", certs.NatsLocalCA},
	}
	for _, pair := range pairs {
		if err := deployGlobalCertificate(clt, pair.name, pair.cert); err != nil {
			return err
		}
	}
	return nil
}

type globalCertDeployer interface {
	GetSecret(name string) (*client.SecretInfo, error)
	GetCA(name string) (*client.CAInfo, error)
	CreateSecret(request *client.SecretCreateRequest) error
	CreateCA(request *client.CACreateRequest) error
}

func isControllerNotFound(err error) bool {
	var notFound *client.NotFoundError
	return errors.As(err, &notFound)
}

func isControllerConflict(err error) bool {
	var conflict *client.ConflictError
	return errors.As(err, &conflict)
}

func controllerSecretExists(clt globalCertDeployer, name string) (bool, error) {
	_, err := clt.GetSecret(name)
	if err == nil {
		return true, nil
	}
	if isControllerNotFound(err) {
		return false, nil
	}
	return false, err
}

func controllerCAExists(clt globalCertDeployer, name string) (bool, error) {
	_, err := clt.GetCA(name)
	if err == nil {
		return true, nil
	}
	if isControllerNotFound(err) {
		return false, nil
	}
	return false, err
}

func deployGlobalCertificate(clt globalCertDeployer, secretName string, cert *SiteCertificate) error {
	if cert == nil {
		return nil
	}

	secretExists, err := controllerSecretExists(clt, secretName)
	if err != nil {
		return err
	}
	caExists, err := controllerCAExists(clt, secretName)
	if err != nil {
		return err
	}
	if secretExists && caExists {
		return nil
	}

	if !secretExists {
		secretRequest := client.SecretCreateRequest{
			Name: secretName,
			Type: "tls",
			Data: map[string]string{
				"ca.crt":  cert.TLSCert,
				"tls.crt": cert.TLSCert,
				"tls.key": cert.TLSKey,
			},
		}
		if err := clt.CreateSecret(&secretRequest); err != nil && !isControllerConflict(err) {
			return err
		}
	}

	if !caExists {
		caRequest := client.CACreateRequest{
			Name:       secretName,
			Type:       "direct",
			SecretName: secretName,
		}
		if err := clt.CreateCA(&caRequest); err != nil && !isControllerConflict(err) {
			return err
		}
	}
	return nil
}
