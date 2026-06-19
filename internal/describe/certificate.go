package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type certificateExecutor struct {
	namespace string
	name      string
	filename  string
}

func newCertificateExecutor(namespace, name, filename string) *certificateExecutor {
	return &certificateExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *certificateExecutor) GetName() string {
	return exe.name
}

func (exe *certificateExecutor) Execute() error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get secret from Controller
	certificate, err := clt.GetCertificate(exe.name)
	if err != nil {
		return err
	}

	var headerKind string
	var headerSpec interface{}
	var headerData interface{}

	if certificate.IsCA {
		headerKind = string(config.CertificateAuthorityKind)
		headerData = rsc.CAInfo{
			Subject:      certificate.Subject,
			IsCA:         certificate.IsCA,
			ValidFrom:    certificate.ValidFrom,
			ValidTo:      certificate.ValidTo,
			SerialNumber: certificate.SerialNumber,
			Certificate:  certificate.Data.Certificate,
			PrivateKey:   certificate.Data.PrivateKey,
		}
	} else {
		headerKind = string(config.CertificateKind)
		// Convert certificate chain to our new structure
		certChain := make([]rsc.CertificateChainItem, len(certificate.CertificateChain))
		for i, cert := range certificate.CertificateChain {
			certChain[i] = rsc.CertificateChainItem{
				Name:    cert.Name,
				Subject: cert.Subject,
			}
		}

		headerData = rsc.CertificateInfo{
			Subject:       certificate.Subject,
			Hosts:         certificate.Hosts,
			IsCA:          certificate.IsCA,
			ValidFrom:     certificate.ValidFrom,
			ValidTo:       certificate.ValidTo,
			SerialNumber:  certificate.SerialNumber,
			CAName:        certificate.CAName,
			DaysRemaining: certificate.DaysRemaining,
			IsExpired:     certificate.IsExpired,
			Certificate:   certificate.Data.Certificate,
			PrivateKey:    certificate.Data.PrivateKey,
		}
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.Kind(headerKind),
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: headerSpec,
		Data: headerData,
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
