package connectk8scontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	connectcontrolplane "github.com/eclipse-iofog/iofogctl/internal/connect/controlplane"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type kubernetesExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
}

func newKubernetesExecutor(controlPlane *rsc.KubernetesControlPlane, namespace string) *kubernetesExecutor {
	return &kubernetesExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return "Kubernetes Control Plane"
}

func NewManualExecutor(namespace, endpoint, kubeConfig, email, password string) (execute.Executor, error) {
	controlPlane := &rsc.KubernetesControlPlane{
		IofogUser:  rsc.IofogUser{Email: email, Password: password},
		KubeConfig: kubeConfig,
		Endpoint:   formatEndpoint(endpoint),
	}
	if err := controlPlane.Sanitize(); err != nil {
		return nil, err
	}
	return newKubernetesExecutor(controlPlane, namespace), nil
}

func NewExecutor(namespace, name string, yaml []byte, kind config.Kind) (execute.Executor, error) {
	// Read the input file
	controlPlane, err := rsc.UnmarshallKubernetesControlPlane(yaml)
	if err != nil {
		return nil, err
	}

	if err := validate(&controlPlane); err != nil {
		return nil, err
	}

	return newKubernetesExecutor(&controlPlane, namespace), nil
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Instantiate Kubernetes cluster object
	k8s, err := install.NewKubernetes(exe.controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Set HTTPS configuration if present in the control plane
	if exe.controlPlane.Controller.Https != nil {
		k8s.SetHttpsEnabled(exe.controlPlane.Controller.Https)
	}

	if exe.controlPlane.Controller.EcnViewerURL != "" {
		viewerDns := true
		k8s.SetIsViewerDns(&viewerDns)
	}

	// Check the resources exist in K8s namespace
	if err = k8s.ExistsInNamespace(exe.namespace); err != nil {
		return
	}

	// Get Controller endpoint
	endpoint, err := k8s.GetControllerEndpoint()
	if err != nil {
		return
	}

	// Establish connection
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return
	}
	err = connectcontrolplane.Connect(exe.controlPlane, endpoint, ns)
	if err != nil {
		return
	}

	pods, err := k8s.GetControllerPods()
	if err != nil {
		return
	}
	for idx := range pods {
		k8sPod := rsc.KubernetesController{
			Endpoint: endpoint,
			PodName:  pods[idx].Name,
		}
		if err := exe.controlPlane.AddController(&k8sPod); err != nil {
			return err
		}
	}
	exe.controlPlane.Endpoint = endpoint

	// Save changes
	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}

func formatEndpoint(endpoint string) string {
	before := util.Before(endpoint, ":")
	after := util.After(endpoint, ":")
	if after == "" {
		after = iofog.ControllerPortString
	}
	return before + ":" + after
}

func validate(controlPlane rsc.ControlPlane) (err error) {
	// Validate user
	user := controlPlane.GetUser()
	if user.Email == "" {
		return util.NewInputError("To connect, Control Plane Iofog User must contain non-empty value in email field")
	}

	return
}
