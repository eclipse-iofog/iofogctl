package deployk8scontrolplane

import (
	"fmt"

	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type kubernetesControlPlaneExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
	name         string
}

func (exe kubernetesControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))
	if err := exe.executeInstall(); err != nil {
		return err
	}

	// Update config
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return
	}
	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}

func (exe kubernetesControlPlaneExecutor) GetName() string {
	return exe.name
}

func newControlPlaneExecutor(namespace, name string, controlPlane *rsc.KubernetesControlPlane) execute.Executor {
	return kubernetesControlPlaneExecutor{
		namespace:    namespace,
		controlPlane: controlPlane,
		name:         name,
	}
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	_, err = config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}

	// Read the input file
	controlPlane, err := rsc.UnmarshallKubernetesControlPlane(opt.Yaml)
	if err != nil {
		return
	}
	if err := validate(&controlPlane); err != nil {
		return nil, err
	}

	return newControlPlaneExecutor(opt.Namespace, opt.Name, &controlPlane), nil
}

func (exe *kubernetesControlPlaneExecutor) executeInstall() (err error) {

	// Get Kubernetes deployer
	installer, err := install.NewKubernetes(exe.controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	installer.SetOperatorImage(exe.controlPlane.Images.Operator)
	installer.SetPullSecret(exe.controlPlane.Images.PullSecret)
	installer.SetRouterImage(exe.controlPlane.Images.Router)
	installer.SetControllerImage(exe.controlPlane.Images.Controller)
	installer.SetNatsImage(exe.controlPlane.Images.Nats)
	installer.SetControllerService(exe.controlPlane.Services.Controller.Type, exe.controlPlane.Services.Controller.Address, exe.controlPlane.Services.Controller.Annotations, exe.controlPlane.Services.Controller.ExternalTrafficPolicy)
	installer.SetRouterService(exe.controlPlane.Services.Router.Type, exe.controlPlane.Services.Router.Address, exe.controlPlane.Services.Router.Annotations, exe.controlPlane.Services.Router.ExternalTrafficPolicy)
	installer.SetNatsService(exe.controlPlane.Services.Nats.Type, exe.controlPlane.Services.Nats.Address, exe.controlPlane.Services.Nats.Annotations, exe.controlPlane.Services.Nats.ExternalTrafficPolicy)
	installer.SetNatsServerService(exe.controlPlane.Services.NatsServer.Type, exe.controlPlane.Services.NatsServer.Address, exe.controlPlane.Services.NatsServer.Annotations, exe.controlPlane.Services.NatsServer.ExternalTrafficPolicy)
	installer.SetControllerIngress(exe.controlPlane.Ingresses.Controller.Annotations, exe.controlPlane.Ingresses.Controller.IngressClassName, exe.controlPlane.Ingresses.Controller.Host, exe.controlPlane.Ingresses.Controller.SecretName)
	installer.SetRouterIngress(exe.controlPlane.Ingresses.Router.Address, exe.controlPlane.Ingresses.Router.MessagePort, exe.controlPlane.Ingresses.Router.InteriorPort, exe.controlPlane.Ingresses.Router.EdgePort)
	installer.SetNatsIngress(exe.controlPlane.Ingresses.Nats.Address, exe.controlPlane.Ingresses.Nats.ServerPort, exe.controlPlane.Ingresses.Nats.ClusterPort, exe.controlPlane.Ingresses.Nats.LeafPort, exe.controlPlane.Ingresses.Nats.MqttPort, exe.controlPlane.Ingresses.Nats.HttpPort)
	// installer.SetRouterConfig(exe.controlPlane.Router.HA)

	// Set isViewerDns based on EcnViewerURL presence
	if exe.controlPlane.Controller.EcnViewerURL != "" {
		viewerDns := true
		installer.SetIsViewerDns(&viewerDns)
	}

	replicas := int32(1)
	if exe.controlPlane.Replicas.Controller != 0 {
		replicas = exe.controlPlane.Replicas.Controller
	}
	replicasNats := exe.controlPlane.Replicas.Nats
	natsSpec := natsSpecToCpv3(exe.controlPlane.Nats)
	vaultSpec := vaultSpecToCpv3(exe.controlPlane.Vault)
	// Create controller on cluster
	// user := install.IofogUser(exe.controlPlane.IofogUser)
	conf := install.K8SControllerConfig{
		// User:          user,
		Replicas:      replicas,
		ReplicasNats:  replicasNats,
		Auth:          install.Auth(exe.controlPlane.Auth),
		Database:      install.Database(exe.controlPlane.Database),
		Events:        install.Events(exe.controlPlane.Events),
		PidBaseDir:    exe.controlPlane.Controller.PidBaseDir,
		EcnViewerPort: exe.controlPlane.Controller.EcnViewerPort,
		EcnViewerURL:  exe.controlPlane.Controller.EcnViewerURL,
		LogLevel:      exe.controlPlane.Controller.LogLevel,
		Https:         exe.controlPlane.Controller.Https,
		SecretName:    exe.controlPlane.Controller.SecretName,
		Nats:          natsSpec,
		Vault:         vaultSpec,
	}
	endpoint, err := installer.CreateControlPlane(&conf)
	if err != nil {
		return
	}

	// Create controller pods for config
	pods, err := installer.GetControllerPods()
	if err != nil {
		return
	}
	for idx := range pods {
		k8sPod := rsc.KubernetesController{
			Endpoint: endpoint,
			PodName:  pods[idx].Name,
			Created:  util.NowUTC(),
		}
		if err := exe.controlPlane.AddController(&k8sPod); err != nil {
			return err
		}
	}

	// Assign control plane endpoint
	exe.controlPlane.Endpoint = endpoint

	return err
}

const clusterIP = "ClusterIP"

func validateControlPlaneUser(controlPlane *rsc.KubernetesControlPlane) error {
	user := controlPlane.GetUser()
	if user.Email == "" {
		return util.NewInputError("Control Plane Iofog User must contain non-empty value in email field")
	}
	return nil
}

func validateControlPlaneAuth(controlPlane *rsc.KubernetesControlPlane) error {
	auth := controlPlane.Auth
	if auth.URL == "" || auth.Realm == "" || auth.SSL == "" || auth.RealmKey == "" || auth.ControllerClient == "" || auth.ControllerSecret == "" || auth.ViewerClient == "" {
		return util.NewInputError("Control Plane Auth Config must contain non-empty values in all fields")
	}
	return nil
}

func validateControlPlaneDatabase(controlPlane *rsc.KubernetesControlPlane) error {
	db := controlPlane.Database
	replicas := controlPlane.Replicas.Controller
	if replicas > 1 {
		if db.Provider == "" || db.Host == "" || db.DatabaseName == "" || db.Password == "" || db.Port == 0 || db.User == "" {
			msg := `When you would like to deploy controller with replicas you must specify an external database for the Control Plane, and you must provide non-empty values in host, databasename, user, password, and port fields.`
			return util.NewInputError(msg)
		}
	}
	return nil
}

func validateControllerServiceAndIngress(controlPlane *rsc.KubernetesControlPlane) error {
	controllerService := controlPlane.Services.Controller
	controllerIngress := controlPlane.Ingresses.Controller
	if controllerService.Type == clusterIP {
		if controllerIngress.Host == "" || controllerIngress.SecretName == "" {
			return util.NewInputError("When Controller service type is ClusterIP, You must provide Ingress configuration for Controller")
		}
	}
	return nil
}

func validateRouterServiceAndIngress(controlPlane *rsc.KubernetesControlPlane) error {
	routerService := controlPlane.Services.Router
	routerIngress := controlPlane.Ingresses.Router
	if routerService.Type == clusterIP {
		if routerIngress.Address == "" || routerIngress.MessagePort == 0 || routerIngress.InteriorPort == 0 || routerIngress.EdgePort == 0 {
			return util.NewInputError("When Router service type is ClusterIP, You must provide Ingress configuration for Default-Router")
		}
	}
	return nil
}

func validateNatsReplicas(controlPlane *rsc.KubernetesControlPlane) error {
	if controlPlane.Replicas.Nats > 0 && controlPlane.Replicas.Nats < 2 {
		return util.NewInputError("When NATS is enabled, replicas.nats must be at least 2")
	}
	return nil
}

func validateControlPlaneVault(controlPlane *rsc.KubernetesControlPlane) error {
	if controlPlane.Vault == nil {
		return nil
	}
	if controlPlane.Vault.Provider == "" {
		return nil
	}
	switch controlPlane.Vault.Provider {
	case "hashicorp", "openbao", "vault":
		if controlPlane.Vault.Hashicorp == nil || (controlPlane.Vault.Hashicorp.Address == "" && controlPlane.Vault.Hashicorp.Token == "") {
			return util.NewInputError("Vault provider " + controlPlane.Vault.Provider + " requires hashicorp block with address and token")
		}
	case "aws", "aws-secrets-manager":
		if controlPlane.Vault.Aws == nil {
			return util.NewInputError("Vault provider " + controlPlane.Vault.Provider + " requires aws block")
		}
	case "azure", "azure-key-vault":
		if controlPlane.Vault.Azure == nil {
			return util.NewInputError("Vault provider " + controlPlane.Vault.Provider + " requires azure block")
		}
	case "google", "google-secret-manager":
		if controlPlane.Vault.Google == nil {
			return util.NewInputError("Vault provider " + controlPlane.Vault.Provider + " requires google block")
		}
	}
	return nil
}

func validate(controlPlane *rsc.KubernetesControlPlane) (err error) {
	if err := validateControlPlaneUser(controlPlane); err != nil {
		return err
	}
	if err := validateControlPlaneAuth(controlPlane); err != nil {
		return err
	}
	if err := validateControlPlaneDatabase(controlPlane); err != nil {
		return err
	}
	if err := validateControllerServiceAndIngress(controlPlane); err != nil {
		return err
	}
	if err := validateRouterServiceAndIngress(controlPlane); err != nil {
		return err
	}
	if err := validateNatsReplicas(controlPlane); err != nil {
		return err
	}
	if err := validateControlPlaneVault(controlPlane); err != nil {
		return err
	}
	return nil
}

func natsSpecToCpv3(n *rsc.NatsSpec) *cpv3.Nats {
	if n == nil {
		return nil
	}
	out := &cpv3.Nats{
		Enabled: n.Enabled,
	}
	if n.JetStream.StorageSize != "" || n.JetStream.MemoryStoreSize != "" || n.JetStream.StorageClassName != "" {
		out.JetStream = cpv3.NatsJetStream{
			StorageSize:      n.JetStream.StorageSize,
			MemoryStoreSize:  n.JetStream.MemoryStoreSize,
			StorageClassName: n.JetStream.StorageClassName,
		}
	}
	return out
}

func vaultSpecToCpv3(v *rsc.VaultSpec) *cpv3.Vault {
	if v == nil {
		return nil
	}
	out := &cpv3.Vault{
		Enabled:  v.Enabled,
		Provider: v.Provider,
		BasePath: v.BasePath,
	}
	if v.Hashicorp != nil {
		out.Hashicorp = &cpv3.VaultHashicorp{
			Address: v.Hashicorp.Address,
			Token:   v.Hashicorp.Token,
			Mount:   v.Hashicorp.Mount,
		}
	}
	if v.Aws != nil {
		out.Aws = &cpv3.VaultAws{
			Region:      v.Aws.Region,
			AccessKeyId: v.Aws.AccessKeyId,
			AccessKey:   v.Aws.AccessKey,
		}
	}
	if v.Azure != nil {
		out.Azure = &cpv3.VaultAzure{
			URL:          v.Azure.URL,
			TenantId:     v.Azure.TenantId,
			ClientId:     v.Azure.ClientId,
			ClientSecret: v.Azure.ClientSecret,
		}
	}
	if v.Google != nil {
		out.Google = &cpv3.VaultGoogle{
			ProjectId:   v.Google.ProjectId,
			Credentials: v.Google.Credentials,
		}
	}
	return out
}
