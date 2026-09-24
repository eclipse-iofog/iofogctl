package deployregistry

import (
	"strings"
	"testing"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func strPtr(v string) *string { return &v }
func boolPtr(v bool) *bool    { return &v }

func TestValidatePrivateOCIRequiresCredentials(t *testing.T) {
	err := validate(rsc.Registry{
		URL:      strPtr("registry.example.com"),
		Private:  boolPtr(true),
		Type:     strPtr(registryTypeOCI),
		Username: strPtr("user"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Username and password")

	err = validate(rsc.Registry{
		URL:      strPtr("registry.example.com"),
		Private:  boolPtr(true),
		Type:     strPtr(registryTypeOCI),
		Username: strPtr("user"),
		Password: strPtr("secret"),
	})
	require.NoError(t, err)
}

func TestValidatePrivateHFRequiresPasswordOnly(t *testing.T) {
	err := validate(rsc.Registry{
		URL:     strPtr("https://huggingface.co"),
		Private: boolPtr(true),
		Type:    strPtr(registryTypeHF),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Password")

	err = validate(rsc.Registry{
		URL:      strPtr("https://huggingface.co"),
		Private:  boolPtr(true),
		Type:     strPtr(registryTypeHF),
		Password: strPtr("hf_xxx"),
	})
	require.NoError(t, err)
}

func TestValidateEmailOptional(t *testing.T) {
	err := validate(rsc.Registry{
		URL:      strPtr("registry.example.com"),
		Private:  boolPtr(true),
		Type:     strPtr(registryTypeOCI),
		Username: strPtr("user"),
		Password: strPtr("secret"),
	})
	require.NoError(t, err)
}

func TestValidateTypeMustBeOciOrHf(t *testing.T) {
	err := validate(rsc.Registry{
		URL:  strPtr("registry.example.com"),
		Type: strPtr("docker"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "oci or hf")
}

func TestUnmarshalRegistryTypeCAInsecure(t *testing.T) {
	raw := []byte(`
url: registry.example.com
private: true
type: oci
insecure: false
ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t
username: user
password: secret
`)
	var registry rsc.Registry
	require.NoError(t, yaml.UnmarshalStrict(raw, &registry))
	require.Equal(t, "oci", *registry.Type)
	require.NotNil(t, registry.Insecure)
	require.False(t, *registry.Insecure)
	require.Equal(t, "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t", *registry.CA)
}

func TestUnmarshalRegistryRejectsLegacyFlags(t *testing.T) {
	raw := []byte(`
url: registry.example.com
requiresCert: true
`)
	var registry rsc.Registry
	err := yaml.UnmarshalStrict(raw, &registry)
	require.Error(t, err)
}

func TestToCreateRequestMapsTypeCAInsecure(t *testing.T) {
	req := toCreateRequest(rsc.Registry{
		URL:      strPtr("registry.example.com"),
		Private:  boolPtr(true),
		Type:     strPtr(registryTypeOCI),
		CA:       strPtr("pem"),
		Insecure: boolPtr(true),
		Username: strPtr("user"),
		Password: strPtr("secret"),
	})
	require.Equal(t, "registry.example.com", req.URL)
	require.False(t, req.IsPublic)
	require.Equal(t, "oci", req.Type)
	require.Equal(t, "pem", req.CA)
	require.True(t, req.Insecure)
	require.Equal(t, "user", req.Username)
}

func TestDefaultTypeIsOci(t *testing.T) {
	raw := []byte("url: registry.example.com\nprivate: false\n")
	var registry rsc.Registry
	require.NoError(t, yaml.UnmarshalStrict(raw, &registry))
	if registry.Type == nil || strings.TrimSpace(*registry.Type) == "" {
		t := registryTypeOCI
		registry.Type = &t
	}
	require.NoError(t, validate(registry))
	require.Equal(t, registryTypeOCI, *registry.Type)
}
