package deployk8scontrolplane

import (
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type translateOptions struct {
	apiVersion      string
	crName          string
	controllerImage string
	routerImage     string
	natsImage       string
}

func defaultTranslateOptions() translateOptions {
	return translateOptions{
		apiVersion:      util.GetCliApiVersion(),
		crName:          util.GetCliCpCrName(),
		controllerImage: util.GetControllerImage(),
		routerImage:     util.GetRouterImage(),
		natsImage:       util.GetNatsImage(),
	}
}

// TranslateToControlPlaneCR maps CLI KubernetesControlPlane spec to an operator ControlPlane CR.
// CLI-only fields (iofogUser, config, ca, images.operator) are omitted from the result.
func TranslateToControlPlaneCR(cp *rsc.KubernetesControlPlane, namespace string) cpv3.ControlPlane {
	return translateToControlPlaneCR(cp, namespace, defaultTranslateOptions())
}

func translateToControlPlaneCR(cp *rsc.KubernetesControlPlane, namespace string, opts translateOptions) cpv3.ControlPlane {
	replicas := int32(1)
	if cp.Replicas.Controller != 0 {
		replicas = cp.Replicas.Controller
	}
	spec := cpv3.ControlPlaneSpec{
		Auth:       rsc.AuthToCPV3(cp.Auth),
		Database:   databaseToCPV3(cp.Database),
		Events:     eventsToCPV3(cp.Events),
		Controller: rsc.ControllerConfigToCPV3(cp.Controller),
		Replicas: cpv3.Replicas{
			Controller: replicas,
		},
		Images:    imagesToCPV3(cp.Images, opts),
		Services:  servicesToCPV3(cp.Services),
		Ingresses: ingressesToCPV3(cp.Ingresses),
		Nats:      natsSpecToCpv3(cp.Nats),
		Vault:     vaultSpecToCpv3(cp.Vault),
	}
	if cp.Replicas.Nats >= 2 {
		spec.Replicas.Nats = cp.Replicas.Nats
	}
	return cpv3.ControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: opts.apiVersion,
			Kind:       "ControlPlane",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.crName,
			Namespace: namespace,
		},
		Spec: spec,
	}
}

func databaseToCPV3(db rsc.Database) cpv3.Database {
	return cpv3.Database{
		Provider:     db.Provider,
		Host:         db.Host,
		Port:         db.Port,
		User:         db.User,
		Password:     db.Password,
		DatabaseName: db.DatabaseName,
		SSL:          db.SSL,
		CA:           db.CA,
	}
}

func eventsToCPV3(ev rsc.Events) cpv3.Events {
	return cpv3.Events{
		AuditEnabled:     ev.AuditEnabled,
		RetentionDays:    ev.RetentionDays,
		CleanupInterval:  ev.CleanupInterval,
		CaptureIpAddress: ev.CaptureIpAddress,
	}
}

func imagesToCPV3(img rsc.KubeImages, opts translateOptions) cpv3.Images {
	controller := img.Controller
	if controller == "" {
		controller = opts.controllerImage
	}
	router := img.Router
	if router == "" {
		router = opts.routerImage
	}
	nats := img.Nats
	if nats == "" {
		nats = opts.natsImage
	}
	return cpv3.Images{
		PullSecret: img.PullSecret,
		Controller: controller,
		Router:     router,
		Nats:       nats,
	}
}

func serviceToCPV3(s rsc.Service) cpv3.Service {
	return cpv3.Service{
		Type:                  s.Type,
		Address:               s.Address,
		Annotations:           s.Annotations,
		ExternalTrafficPolicy: s.ExternalTrafficPolicy,
	}
}

func servicesToCPV3(s rsc.Services) cpv3.Services {
	return cpv3.Services{
		Controller: serviceToCPV3(s.Controller),
		Router:     serviceToCPV3(s.Router),
		Nats:       serviceToCPV3(s.Nats),
		NatsServer: serviceToCPV3(s.NatsServer),
	}
}

func ingressesToCPV3(in rsc.Ingresses) cpv3.Ingresses {
	return cpv3.Ingresses{
		Controller: cpv3.ControllerIngress{
			Annotations:      in.Controller.Annotations,
			IngressClassName: in.Controller.IngressClassName,
			Host:             in.Controller.Host,
			SecretName:       in.Controller.SecretName,
		},
		Router: cpv3.RouterIngress{
			Address:      in.Router.Address,
			MessagePort:  in.Router.MessagePort,
			InteriorPort: in.Router.InteriorPort,
			EdgePort:     in.Router.EdgePort,
		},
		Nats: cpv3.NatsIngress{
			Address:     in.Nats.Address,
			ServerPort:  in.Nats.ServerPort,
			ClusterPort: in.Nats.ClusterPort,
			LeafPort:    in.Nats.LeafPort,
			MqttPort:    in.Nats.MqttPort,
			HttpPort:    in.Nats.HttpPort,
		},
	}
}
