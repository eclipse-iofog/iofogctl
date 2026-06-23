package install

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/stretchr/testify/require"
)

type fakeGlobalCertClient struct {
	secrets           map[string]struct{}
	cas               map[string]struct{}
	createSecretCalls int
	createCACalls     int
}

func (f *fakeGlobalCertClient) GetSecret(name string) (*client.SecretInfo, error) {
	if _, ok := f.secrets[name]; ok {
		return &client.SecretInfo{Name: name}, nil
	}
	return nil, client.NewNotFoundError(name)
}

func (f *fakeGlobalCertClient) GetCA(name string) (*client.CAInfo, error) {
	if _, ok := f.cas[name]; ok {
		return &client.CAInfo{Name: name}, nil
	}
	return nil, client.NewNotFoundError(name)
}

func (f *fakeGlobalCertClient) CreateSecret(request *client.SecretCreateRequest) error {
	f.createSecretCalls++
	if f.secrets == nil {
		f.secrets = map[string]struct{}{}
	}
	f.secrets[request.Name] = struct{}{}
	return nil
}

func (f *fakeGlobalCertClient) CreateCA(request *client.CACreateRequest) error {
	f.createCACalls++
	if f.cas == nil {
		f.cas = map[string]struct{}{}
	}
	f.cas[request.Name] = struct{}{}
	return nil
}

func testSiteCertificate() *SiteCertificate {
	return &SiteCertificate{
		TLSCert: "cert-pem",
		TLSKey:  "key-pem",
	}
}

func TestDeployGlobalCertificateSkipsWhenBothExist(t *testing.T) {
	clt := &fakeGlobalCertClient{
		secrets: map[string]struct{}{"router-site-ca": {}},
		cas:     map[string]struct{}{"router-site-ca": {}},
	}

	require.NoError(t, deployGlobalCertificate(clt, "router-site-ca", testSiteCertificate()))
	require.Equal(t, 0, clt.createSecretCalls)
	require.Equal(t, 0, clt.createCACalls)
}

func TestDeployGlobalCertificateCreatesBothWhenMissing(t *testing.T) {
	clt := &fakeGlobalCertClient{}

	require.NoError(t, deployGlobalCertificate(clt, "router-site-ca", testSiteCertificate()))
	require.Equal(t, 1, clt.createSecretCalls)
	require.Equal(t, 1, clt.createCACalls)
}

func TestDeployGlobalCertificateCreatesCAWhenSecretExists(t *testing.T) {
	clt := &fakeGlobalCertClient{
		secrets: map[string]struct{}{"router-site-ca": {}},
	}

	require.NoError(t, deployGlobalCertificate(clt, "router-site-ca", testSiteCertificate()))
	require.Equal(t, 0, clt.createSecretCalls)
	require.Equal(t, 1, clt.createCACalls)
}

func TestDeployGlobalCertificateCreatesSecretWhenCAExists(t *testing.T) {
	clt := &fakeGlobalCertClient{
		cas: map[string]struct{}{"router-site-ca": {}},
	}

	require.NoError(t, deployGlobalCertificate(clt, "router-site-ca", testSiteCertificate()))
	require.Equal(t, 1, clt.createSecretCalls)
	require.Equal(t, 0, clt.createCACalls)
}

func TestDeployGlobalCertificateNilCertNoOp(t *testing.T) {
	clt := &fakeGlobalCertClient{}
	require.NoError(t, deployGlobalCertificate(clt, "router-site-ca", nil))
	require.Equal(t, 0, clt.createSecretCalls)
	require.Equal(t, 0, clt.createCACalls)
}

func TestDeployGlobalCertificateTreatsCreateConflictAsExists(t *testing.T) {
	clt := &conflictOnCreateGlobalCertClient{}
	require.NoError(t, deployGlobalCertificate(clt, "router-site-ca", testSiteCertificate()))
}

type conflictOnCreateGlobalCertClient struct {
	fakeGlobalCertClient
}

func (f *conflictOnCreateGlobalCertClient) CreateSecret(request *client.SecretCreateRequest) error {
	return client.NewConflictError(request.Name)
}

func (f *conflictOnCreateGlobalCertClient) CreateCA(request *client.CACreateRequest) error {
	return client.NewConflictError(request.Name)
}
