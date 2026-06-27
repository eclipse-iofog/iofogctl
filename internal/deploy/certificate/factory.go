package deploycertificate

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
	Kind      config.Kind
}

type executor struct {
	certificate rsc.CertificateCreateRequest
	caCert      rsc.CACreateRequest
	namespace   string
	name        string
	isCA        bool
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) updateCertificate(clt *client.Client) error {
	_, err := clt.GetCertificate(exe.name)
	if err != nil {
		return err
	}
	return util.NewInputError("Updating Certificates are not allowed If you want you can delete it first then re create it")
}

func (exe *executor) createCertificate(clt *client.Client) error {
	if exe.isCA {
		if (exe.caCert.Type == "k8s-secret" || exe.caCert.Type == "direct") && (exe.name != exe.caCert.SecretName) {
			return util.NewInputError("While importing CA certificate from Kubernetes secret or direct PoT Secret, the name of the CA Certificate must be the same as the Secret name")
		}
		var secretName string
		if exe.caCert.Type == "k8s-secret" && exe.caCert.SecretName == "" {
			secretName = exe.name
		} else if exe.caCert.Type == "direct" && exe.caCert.SecretName == "" {
			secretName = exe.name
		}
		// Create CA certificate
		request := client.CACreateRequest{
			Name:       exe.name,
			Subject:    exe.caCert.Subject,
			Type:       exe.caCert.Type,
			Expiration: exe.caCert.Expiration,
			SecretName: secretName,
		}
		err := clt.CreateCA(&request)
		return err
	} else {
		// Create regular certificate
		request := client.CertificateCreateRequest{
			Name:       exe.name,
			Subject:    exe.certificate.Subject,
			Hosts:      exe.certificate.Hosts,
			Expiration: exe.certificate.Expiration,
			CA: client.CertificateCreateCA{
				Type:       exe.certificate.CA.Type,
				SecretName: exe.certificate.CA.SecretName,
			},
		}
		err := clt.CreateCertificate(&request)
		return err
	}
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying certificate %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if _, err = clt.GetCertificate(exe.name); err != nil {
		return exe.createCertificate(clt)
	}
	return exe.updateCertificate(clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Certificate")
	}

	isCA := opt.Kind == config.CertificateAuthorityKind
	// Create appropriate request based on kind
	if isCA {
		caCert := rsc.CACreateRequest{
			Name: opt.Name,
		}
		if err = yaml.UnmarshalStrict(opt.Yaml, &caCert); err != nil {
			return nil, util.NewUnmarshalError(err.Error())
		}
		return &executor{
			namespace: opt.Namespace,
			caCert:    caCert,
			name:      opt.Name,
			isCA:      true,
		}, nil
	} else {
		certificate := rsc.CertificateCreateRequest{
			Name: opt.Name,
		}
		if err = yaml.UnmarshalStrict(opt.Yaml, &certificate); err != nil {
			return nil, util.NewUnmarshalError(err.Error())
		}
		return &executor{
			namespace:   opt.Namespace,
			certificate: certificate,
			name:        opt.Name,
			isCA:        false,
		}, nil
	}
}
