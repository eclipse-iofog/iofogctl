package deployremotecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

func runExecutors(executors []execute.Executor) error {
	if errs, _ := execute.ForParallel(executors); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
	return nil
}

func globalCertificatesFromCP(cp *rsc.RemoteControlPlane) install.GlobalCertificates {
	if cp == nil {
		return install.GlobalCertificates{}
	}
	out := install.GlobalCertificates{}
	if cp.RouterSiteCA != nil {
		out.RouterSiteCA = &install.SiteCertificate{
			TLSCert: cp.RouterSiteCA.TLSCert,
			TLSKey:  cp.RouterSiteCA.TLSKey,
		}
	}
	if cp.RouterLocalCA != nil {
		out.RouterLocalCA = &install.SiteCertificate{
			TLSCert: cp.RouterLocalCA.TLSCert,
			TLSKey:  cp.RouterLocalCA.TLSKey,
		}
	}
	if cp.NatsSiteCA != nil {
		out.NatsSiteCA = &install.SiteCertificate{
			TLSCert: cp.NatsSiteCA.TLSCert,
			TLSKey:  cp.NatsSiteCA.TLSKey,
		}
	}
	if cp.NatsLocalCA != nil {
		out.NatsLocalCA = &install.SiteCertificate{
			TLSCert: cp.NatsLocalCA.TLSCert,
			TLSKey:  cp.NatsLocalCA.TLSKey,
		}
	}
	return out
}
