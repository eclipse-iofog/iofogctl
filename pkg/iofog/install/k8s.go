package install

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	ioclient "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	opk8s "github.com/eclipse-iofog/iofog-operator/v3/pkg/k8s"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	corev1 "k8s.io/api/core/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // GCP auth
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	opclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	config        *restclient.Config
	opClient      opclient.Client
	clientset     *kubernetes.Clientset
	extsClientset *extsclientset.Clientset
	ns            string
	operator      *microservice
	services      cpv3.Services
	images        cpv3.Images
	ingresses     cpv3.Ingresses
	httpsEnabled  *bool
	cpSpec        cpv3.ControlPlaneSpec
}

// NewKubernetes constructs an object to manage cluster
func NewKubernetes(configFilename, namespace string) (*Kubernetes, error) {
	// Get the kubernetes config from the filepath.
	config, err := clientcmd.BuildConfigFromFlags("", configFilename)
	if err != nil {
		return nil, err
	}

	// Instantiate Kubernetes clients
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	extsClientset, err := extsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		config:        config,
		clientset:     clientset,
		extsClientset: extsClientset,
		ns:            namespace,
		operator:      newOperatorMicroservice(),
	}, nil
}

func (k8s *Kubernetes) SetOperatorImage(image string) {
	if image != "" {
		k8s.operator.containers[0].image = image
	} else {
		k8s.operator.containers[0].image = util.GetOperatorImage()
	}
}

func (k8s *Kubernetes) SetRouterImage(image string) {
	if image != "" {
		k8s.images.Router = image
	} else {
		k8s.images.Router = util.GetRouterImage()
	}
}

func (k8s *Kubernetes) SetControllerImage(image string) {
	if image != "" {
		k8s.images.Controller = image
	} else {
		k8s.images.Controller = util.GetControllerImage()
	}
}

func (k8s *Kubernetes) SetNatsImage(image string) {
	if image != "" {
		k8s.images.Nats = image
	} else {
		k8s.images.Nats = util.GetNatsImage()
	}
}

func (k8s *Kubernetes) SetPullSecret(pullSecret string) {
	if pullSecret != "" {
		k8s.images.PullSecret = pullSecret
		k8s.operator.imagePullSecret = pullSecret
	}
}

func (k8s *Kubernetes) SetHttpsEnabled(enabled *bool) {
	k8s.httpsEnabled = enabled
}

const controlPlaneReadyTimeout = 600 * time.Second
const controlPlanePollInterval = 3 * time.Second

func (k8s *Kubernetes) enableCustomResources() error {
	ctx := context.Background()
	for _, crd := range controlPlaneCRDsToInstall() {
		// Try create new
		if _, err := k8s.extsClientset.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{}); err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				return err
			}
			// Update
			existingCRD, err := k8s.extsClientset.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// Always update the CRD if:
			// 1. The CRD is not supported (major version mismatch)
			// 2. The versions array is different (new features/types added)
			shouldUpdate := !isSupportedControlPlaneCRD(existingCRD, crd) ||
				!reflect.DeepEqual(existingCRD.Spec.Versions, crd.Spec.Versions)

			if shouldUpdate {
				// Preserve the existing status
				existingCRD.Spec.Versions = crd.Spec.Versions

				// Update the CRD
				if _, err := k8s.extsClientset.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, existingCRD, metav1.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
	}

	// Deploy operator again
	if err := k8s.createOperator(); err != nil {
		return err
	}

	// Enable client
	if err := k8s.enableOperatorClient(); err != nil {
		return err
	}

	return nil
}

func (k8s *Kubernetes) enableOperatorClient() (err error) {
	scheme := initOperatorClientScheme()
	k8s.opClient, err = opclient.New(k8s.config, opclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	return nil
}

// CreateControlPlane applies or updates the operator ControlPlane CR in the cluster.
func (k8s *Kubernetes) CreateControlPlane(desired cpv3.ControlPlane) (endpoint string, err error) {
	// Create namespace if required
	Verbose("Creating namespace " + k8s.ns)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ns,
		},
	}
	if _, err = k8s.clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Set up CRDs if required
	Verbose("Enabling CRDs")
	if err = k8s.enableCustomResources(); err != nil {
		return
	}

	// Check if Control Plane exists
	Verbose("Finding existing Control Plane")
	crName := util.GetCliCpCrName()
	cpKey := opclient.ObjectKey{
		Name:      crName,
		Namespace: k8s.ns,
	}
	var cp cpv3.ControlPlane
	found := true
	if err = k8s.opClient.Get(context.Background(), cpKey, &cp); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
		// Not found, set basic info
		found = false
		cp = cpv3.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      crName,
				Namespace: k8s.ns,
			},
		}
	}

	cp.Spec = desired.Spec
	k8s.cpSpec = desired.Spec

	// Store HTTPS configuration for endpoint generation
	k8s.SetHttpsEnabled(desired.Spec.Controller.Https)

	// Create or update Control Plane
	if found {
		Verbose("Updating existing Control Plane")
		if err = k8s.opClient.Update(context.Background(), &cp); err != nil {
			return
		}
	} else {
		cp.SetConditionDeploying(nil)
		Verbose("Deploying new Control Plane")
		if err = k8s.opClient.Create(context.Background(), &cp); err != nil {
			return
		}
	}

	if err = k8s.waitControlPlaneReady(context.Background(), controlPlaneReadyTimeout); err != nil {
		return
	}

	endpoint, err = k8s.GetControllerEndpoint()
	return endpoint, err
}

func (k8s *Kubernetes) waitControlPlaneReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	cpKey := opclient.ObjectKey{
		Name:      util.GetCliCpCrName(),
		Namespace: k8s.ns,
	}
	ticker := time.NewTicker(controlPlanePollInterval)
	defer ticker.Stop()

	for {
		var cp cpv3.ControlPlane
		if err := k8s.opClient.Get(ctx, cpKey, &cp); err != nil {
			return util.NewInternalError("Error reading Control Plane resource: " + err.Error())
		}
		if cp.IsReady() {
			return nil
		}
		if time.Now().After(deadline) {
			return util.NewInternalError("Timed out waiting for Control Plane to become ready")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (k8s *Kubernetes) deleteOperator() (err error) {
	// Resource name for deletions
	name := k8s.operator.name

	ctx := context.Background()

	// Service Account
	if err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Role
	if err = k8s.clientset.RbacV1().Roles(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Role Binding
	if err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Deployment
	if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	return nil
}

func (k8s *Kubernetes) createOperator() (err error) {
	ctx := context.Background()

	// Service Account
	opSvcAcc := newServiceAccount(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(ctx, opSvcAcc, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Role
	role := newRole(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().Roles(k8s.ns).Create(ctx, role, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Role Binding
	rb := newRoleBinding(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Create(ctx, rb, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Deployment
	opDep := newDeployment(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(ctx, opDep, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
		// Redeploy the operator
		if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(ctx, k8s.operator.name, metav1.DeleteOptions{}); err != nil {
			return
		}
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(ctx, opDep, metav1.CreateOptions{}); err != nil {
			return
		}
	}
	return nil
}

func (k8s *Kubernetes) DeleteControlPlane(deleteNamespace bool) error {
	// Prepare Control Plane client
	if err := k8s.enableOperatorClient(); err != nil {
		return err
	}

	// Delete Control Plane
	cp := &cpv3.ControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GetCliCpCrName(),
			Namespace: k8s.ns,
		},
	}
	if err := k8s.opClient.Delete(context.Background(), cp); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// Delete Operator
	if err := k8s.deleteOperator(); err != nil {
		return err
	}

	if deleteNamespace && k8s.ns != "default" {
		if err := k8s.clientset.CoreV1().Namespaces().Delete(context.Background(), k8s.ns, metav1.DeleteOptions{}); err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (k8s *Kubernetes) waitForService(name string, targetPort int32) (addr string, nodePort int32, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), controlPlaneReadyTimeout)
	defer cancel()

	svc, err := k8s.waitForServiceExists(ctx, name)
	if err != nil {
		return "", 0, err
	}

	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		getter := opk8s.CoreV1ServiceGetter{Kube: k8s.clientset}
		remaining := time.Until(deadlineFromContext(ctx))
		addr, err = opk8s.WaitLoadBalancerAddress(ctx, getter, k8s.ns, name, remaining)
		if err != nil {
			return "", 0, err
		}
		return addr, targetPort, nil

	case corev1.ServiceTypeNodePort:
		addr, err = k8s.getNodePortAddress(name)
		if err != nil {
			util.PrintNotify("Could not get an external IP address of any Kubernetes nodes for NodePort service " + name + "\nTrying to reach the cluster IP of the service")
			addr, err = k8s.getClusterIPAddress(name)
			if err != nil {
				return "", 0, err
			}
		}
		nodePort, err = k8s.getPort(svc, name, targetPort)
		return addr, nodePort, err

	case corev1.ServiceTypeClusterIP:
		addr, err = k8s.waitForIngress(ctx, controller)
		if err != nil {
			return "", 0, err
		}
		return addr, targetPort, nil

	default:
		return "", 0, util.NewError("Found Service was not of supported type")
	}
}

func deadlineFromContext(ctx context.Context) time.Time {
	if d, ok := ctx.Deadline(); ok {
		return d
	}
	return time.Now().Add(controlPlaneReadyTimeout)
}

func (k8s *Kubernetes) waitForServiceExists(ctx context.Context, name string) (*corev1.Service, error) {
	ticker := time.NewTicker(controlPlanePollInterval)
	defer ticker.Stop()
	for {
		svc, err := k8s.clientset.CoreV1().Services(k8s.ns).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			return svc, nil
		}
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, util.NewInternalError("Timed out waiting for service " + name)
		case <-ticker.C:
		}
	}
}

func (k8s *Kubernetes) waitForIngress(ctx context.Context, name string) (addr string, err error) {
	scheme := "http"
	if k8s.ingressUsesHTTPS() {
		scheme = "https"
	}

	ticker := time.NewTicker(controlPlanePollInterval)
	defer ticker.Stop()
	for {
		ingress, err := k8s.clientset.NetworkingV1().Ingresses(k8s.ns).Get(ctx, name, metav1.GetOptions{})
		if err == nil && len(ingress.Spec.Rules) > 0 && ingress.Spec.Rules[0].Host != "" {
			return scheme + "://" + ingress.Spec.Rules[0].Host, nil
		}
		if err != nil && !k8serrors.IsNotFound(err) {
			return "", err
		}
		select {
		case <-ctx.Done():
			return "", util.NewInternalError("Timed out waiting for Ingress " + name)
		case <-ticker.C:
		}
	}
}

func (k8s *Kubernetes) ingressUsesHTTPS() bool {
	if k8s.httpsEnabled != nil && *k8s.httpsEnabled {
		return true
	}
	if k8s.cpSpec.Ingresses.Controller.SecretName != "" {
		return true
	}
	return false
}

func (k8s *Kubernetes) getCRPublicURL(ctx context.Context) string {
	var cp cpv3.ControlPlane
	if err := k8s.opClient.Get(ctx, opclient.ObjectKey{
		Name:      util.GetCliCpCrName(),
		Namespace: k8s.ns,
	}, &cp); err != nil {
		return ""
	}
	return cp.Spec.Controller.PublicUrl
}

func (k8s *Kubernetes) getPort(svc *corev1.Service, name string, targetPort int32) (nodePort int32, err error) {
	// Get the port allocated on the node
	for _, port := range svc.Spec.Ports {
		if port.TargetPort.IntVal == targetPort {
			nodePort = port.NodePort
			break
		}
	}
	if nodePort == 0 {
		err = util.NewError("Could not get node port for Kubernetes service " + name)
		return
	}
	return
}

func (k8s *Kubernetes) getClusterIPAddress(name string) (addr string, err error) {
	// Get a list of K8s nodes and return one of their external IPs
	var nodeList *corev1.NodeList
	nodeList, err = k8s.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for idx := range nodeList.Items {
		node := &nodeList.Items[idx]
		for _, addrs := range node.Status.Addresses {
			if addrs.Type == corev1.NodeInternalIP {
				addr = addrs.Address
				break
			}
		}
	}
	if addr == "" {
		err = util.NewError("Could not get address for ClusterIP " + name)
	}
	return
}

func (k8s *Kubernetes) getNodePortAddress(name string) (addr string, err error) {
	// Get a list of K8s nodes and return one of their external IPs
	var nodeList *corev1.NodeList
	nodeList, err = k8s.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	if len(nodeList.Items) == 0 {
		err = util.NewError("Could not find Kubernetes nodes when waiting for NodePort service " + name)
		return
	}
	// Return external IP of any of the nodes in the cluster
	for idx := range nodeList.Items {
		node := &nodeList.Items[idx]
		for _, addrs := range node.Status.Addresses {
			if addrs.Type == corev1.NodeExternalIP {
				addr = addrs.Address
				break
			}
		}
	}
	if addr == "" {
		err = util.NewError("Could not find address in Node Port service " + name)
	}
	return
}

func (k8s *Kubernetes) SetControllerService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.Controller.Type = svcType
	} else {
		k8s.services.Controller.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Controller.Address = address
	k8s.services.Controller.Annotations = annotations
	k8s.services.Controller.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetRouterService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.Router.Type = svcType
	} else {
		k8s.services.Router.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Router.Address = address
	k8s.services.Router.Annotations = annotations
	k8s.services.Router.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetControllerIngress(annotations map[string]string, ingressClassName string, host string, secretName string) {
	k8s.ingresses.Controller.Annotations = annotations
	k8s.ingresses.Controller.IngressClassName = ingressClassName
	k8s.ingresses.Controller.Host = host
	k8s.ingresses.Controller.SecretName = secretName
}

func (k8s *Kubernetes) SetRouterIngress(address string, messagePort int, interiorPort int, edgePort int) {
	k8s.ingresses.Router.Address = address
	k8s.ingresses.Router.MessagePort = messagePort
	k8s.ingresses.Router.InteriorPort = interiorPort
	k8s.ingresses.Router.EdgePort = edgePort
}

func (k8s *Kubernetes) SetNatsService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.Nats.Type = svcType
	} else {
		k8s.services.Nats.Type = string(corev1.ServiceTypeClusterIP)
	}
	k8s.services.Nats.Address = address
	k8s.services.Nats.Annotations = annotations
	k8s.services.Nats.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetNatsServerService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.NatsServer.Type = svcType
	} else {
		k8s.services.NatsServer.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.NatsServer.Address = address
	k8s.services.NatsServer.Annotations = annotations
	k8s.services.NatsServer.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetNatsIngress(address string, serverPort, clusterPort, leafPort, mqttPort, httpPort int) {
	k8s.ingresses.Nats.Address = address
	k8s.ingresses.Nats.ServerPort = serverPort
	k8s.ingresses.Nats.ClusterPort = clusterPort
	k8s.ingresses.Nats.LeafPort = leafPort
	k8s.ingresses.Nats.MqttPort = mqttPort
	k8s.ingresses.Nats.HttpPort = httpPort
}

// func (k8s *Kubernetes) SetRouterConfig(HA *bool) {
// 	k8s.router.HA = HA

// }

func (k8s *Kubernetes) ExistsInNamespace(namespace string) error {
	ctx := context.Background()
	// Check namespace exists
	if _, err := k8s.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return util.NewError("Could not find Namespace " + namespace + " on Kubernetes cluster")
		}
		return err
	}

	// Check services exist
	svcList, err := k8s.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for idx := range svcList.Items {
		svc := &svcList.Items[idx]
		if svc.Name == controller {
			return nil
		}
	}
	return util.NewError("Could not find Controller Service in Kubernetes namespace " + namespace)
}

func (k8s *Kubernetes) formatEndpoint(endpoint string, port int32) (*url.URL, error) {
	// Ensure protocol
	if !strings.Contains(endpoint, "://") {
		// Check if HTTPS should be used
		if k8s.httpsEnabled != nil && *k8s.httpsEnabled {
			endpoint = fmt.Sprintf("https://%s", endpoint)
		} else {
			endpoint = fmt.Sprintf("http://%s", endpoint)
		}
	}
	URL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	// Ensure port on non-HTTPS endpoints when omitted.
	if !strings.Contains(URL.Host, ":") && URL.Scheme != "https" {
		URL.Host += fmt.Sprintf(":%d", port)
	}
	return URL, nil

}

func (k8s *Kubernetes) GetControllerEndpoint() (endpoint string, err error) {
	ip, port, err := k8s.waitForService(controller, ioclient.ControllerPort)
	if err != nil {
		if publicURL := k8s.getCRPublicURL(context.Background()); publicURL != "" {
			return publicURL, nil
		}
		return "", err
	}
	if ip == "" {
		if publicURL := k8s.getCRPublicURL(context.Background()); publicURL != "" {
			return publicURL, nil
		}
		return "", util.NewInternalError("Could not resolve Controller endpoint")
	}

	formattedURL, err := k8s.formatEndpoint(ip, port)
	if err != nil {
		return "", err
	}
	endpoint = formattedURL.String()

	useHTTPS := k8s.httpsEnabled != nil && *k8s.httpsEnabled
	return util.GetControllerEndpoint(endpoint, useHTTPS)
}

func (k8s *Kubernetes) GetControllerPods() (podNames []Pod, err error) {
	podNames = []Pod{}
	// List pods
	pods, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	// Find Controller pods
	for idx := range pods.Items {
		if pods.Items[idx].Labels["iofog.org/component"] == controller {
			podNames = append(podNames, Pod{
				Name:   pods.Items[idx].Name,
				Status: string(pods.Items[idx].Status.Phase),
			})
		}
	}
	return
}
