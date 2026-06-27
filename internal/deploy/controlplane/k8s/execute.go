package deployk8scontrolplane

import (
	"context"
	"fmt"

	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	"github.com/eclipse-iofog/iofogctl/internal/auth"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	inputvalidate "github.com/eclipse-iofog/iofogctl/internal/validate"
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

	// Configure operator deploy (CLI-only image fields)
	installer.SetOperatorImage(exe.controlPlane.Images.Operator)
	installer.SetPullSecret(exe.controlPlane.Images.PullSecret)

	desired := TranslateToControlPlaneCR(exe.controlPlane, exe.namespace)
	if ca := rsc.GetTrustCA(exe.controlPlane); ca != "" {
		if err := trust.StoreCA(exe.namespace, ca); err != nil {
			return err
		}
	}

	endpoint, err := installer.CreateControlPlane(desired)
	if err != nil {
		return
	}

	if err := trust.WaitForControllerAPI(context.Background(), exe.namespace, endpoint); err != nil {
		return err
	}

	if err := auth.EnsureIofogUserEmbedded(context.Background(), exe.namespace, endpoint, auth.EmbeddedAuthSpec{
		Mode:      exe.controlPlane.Auth.Mode,
		Bootstrap: exe.controlPlane.Auth.Bootstrap,
		User:      &exe.controlPlane.IofogUser,
	}); err != nil {
		return err
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
	if exe.controlPlane.Controller.PublicUrl == "" {
		exe.controlPlane.Controller.PublicUrl = endpoint
	}
	if err := rsc.BackfillConsoleURL(exe.controlPlane); err != nil {
		return err
	}

	return err
}

const (
	clusterIP        = "ClusterIP"
	authModeEmbedded = "embedded"
	authModeExternal = "external"
)

func validateControlPlaneUser(controlPlane *rsc.KubernetesControlPlane) error {
	user := controlPlane.GetUser()
	if user.Email == "" {
		return util.NewInputError("Control Plane Iofog User must contain non-empty value in email field")
	}
	if rawPassword := user.GetRawPassword(); rawPassword != "" {
		if err := inputvalidate.ValidatePasswordComplexity(rawPassword); err != nil {
			return err
		}
	}
	return nil
}

func validateControlPlaneAuth(controlPlane *rsc.KubernetesControlPlane) error {
	auth := controlPlane.Auth
	switch auth.Mode {
	case authModeEmbedded:
		return validateEmbeddedAuth(auth)
	case authModeExternal:
		return validateExternalAuth(auth)
	case "":
		return util.NewInputError("Control Plane auth.mode is required (embedded or external)")
	default:
		return util.NewInputError(fmt.Sprintf("Control Plane auth.mode %q is invalid (embedded or external)", auth.Mode))
	}
}

func validateEmbeddedAuth(auth rsc.Auth) error {
	if auth.Bootstrap == nil {
		return util.NewInputError("Control Plane auth.bootstrap is required when auth.mode is embedded")
	}
	if auth.Bootstrap.Username == "" {
		return util.NewInputError("Control Plane auth.bootstrap.username is required when auth.mode is embedded")
	}
	if auth.Bootstrap.Password == "" {
		return util.NewInputError("Control Plane auth.bootstrap.password is required in YAML when auth.mode is embedded")
	}
	if err := inputvalidate.ValidatePasswordComplexity(auth.Bootstrap.Password); err != nil {
		return err
	}
	return nil
}

func validateExternalAuth(auth rsc.Auth) error {
	if auth.IssuerUrl == "" {
		return util.NewInputError("Control Plane auth.issuerUrl is required when auth.mode is external")
	}
	if auth.Client == nil || auth.Client.ID == "" {
		return util.NewInputError("Control Plane auth.client.id is required when auth.mode is external")
	}
	if auth.Client.Secret == "" {
		return util.NewInputError("Control Plane auth.client.secret is required when auth.mode is external")
	}
	return nil
}

func natsEnabled(controlPlane *rsc.KubernetesControlPlane) bool {
	if controlPlane.Nats == nil {
		return true
	}
	if controlPlane.Nats.Enabled == nil {
		return true
	}
	return *controlPlane.Nats.Enabled
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
	if !natsEnabled(controlPlane) {
		return nil
	}
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
